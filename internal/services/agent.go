package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ChatEventType 标识一条流式事件的类型。
type ChatEventType string

const (
	ChatEventDelta ChatEventType = "delta" // 模型文本增量
	ChatEventTool  ChatEventType = "tool"  // 工具调用状态变化
	ChatEventDone  ChatEventType = "done"  // 本次回答结束
	ChatEventError ChatEventType = "error" // 本次回答出错
)

type AgentServices struct {
	repo ModelRepo
}

func NewAgentServices(repo ModelRepo) *AgentServices {
	return &AgentServices{
		repo: repo,
	}
}

type ChatRequest struct {
	UserInput string `json:"user_input"` // 用户输入的问题
}

type ChatEvent struct {
	Type       ChatEventType `json:"type"`                  // 事件类型
	Delta      string        `json:"delta,omitempty"`       // type=delta：增量文本
	Tool       string        `json:"tool,omitempty"`        // type=tool：工具名
	ToolStatus string        `json:"tool_status,omitempty"` // type=tool：执行状态
	Content    string        `json:"content,omitempty"`     // type=done：完整回答
	Error      string        `json:"error,omitempty"`       // type=error：错误信息
}

func (a *AgentServices) Ask(ctx context.Context, input string, onEvent func(ChatEvent) error) error {
	//todo 规范常量变化
	config, err := a.repo.GetByType(ctx, 1)
	if err != nil {
		return err
	}
	// 后面有多个配置，需要做队列，负载，并发控制
	cfg := config[0]
	// 构建一个agent
	chat, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: cfg.BaseURL,
		Model:   cfg.Name,
		APIKey:  cfg.Key,
	})
	if err != nil {
		return nil
	}
	// 构建工具
	search, err := NewWebSerach(cfg.Key)
	if err != nil {
		return err
	}
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name: "智能小助手",
		Description: `你是AI个人助手，负责解答用户的各种问题。你的主要职责是：
             1. **信息准确性守护者**：确保提供的信息准确无误。
             2. **搜索成本优化师**：在信息准确性和搜索成本之间找到最佳平衡。
             # 任务说明
             ## 1. 联网意图判断
             当用户提出的问题涉及以下情况时，需使用 "web_search" 进行联网搜索：
             - **时效性**：问题需要最新或实时的信息。
             - **知识盲区**：问题超出当前知识范围，无法准确解答。
             - **信息不足**：现有知识库无法提供完整或详细的解答。
             ## 2. 联网后回答
             - 在回答中，优先使用已搜索到的资料。
             - 回复结构应清晰，使用序号、分段等方式帮助用户理解。
             ## 3. 引用已搜索资料
             - 当使用联网搜索的资料时，在正文中明确引用来源，引用格式为：
             "[1]  (URL地址)"。
             ## 4. 总结与参考资料
             - 在回复的最后，列出所有已参考的资料。格式为：
             1. [资料标题](URL地址1)
             2. [资料标题](URL地址2)`,
		Model: chat,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{
					search,
				},
			},
		},
	})
	if err != nil {
		log.Fatalf(err.Error())
		return err
	}
	// 运行agent
	iter := agent.Run(ctx, &adk.AgentInput{
		Messages: []adk.Message{schema.UserMessage(input)},
	})
	// 累积完整回答，结束时随 done 事件兜底下发。
	var full strings.Builder
	emit := func(ev ChatEvent) error {
		return onEvent(ev)
	}
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			if ctx.Err() != nil {
				// 前端断了而已
				fmt.Print("前端断开")
				return ctx.Err()
			}
			log.Fatal(event.Err.Error())
			_ = emit(ChatEvent{Type: ChatEventError, Error: event.Err.Error()})
			return event.Err
		}
		mo := event.Output.MessageOutput
		if mo == nil {
			// 继续等
			continue
		}
		// ReAct 循环中工具执行完成的消息：Role=Tool 且带 ToolName。
		if mo.Role == schema.Tool {
			if err := emit(ChatEvent{Type: ChatEventTool, Tool: mo.ToolName, ToolStatus: "done"}); err != nil {
				log.Fatal(err.Error())
				return err
			}
			continue
		}
		// 模型输出：流式模式下逐帧转发增量文本。
		if mo.IsStreaming {
			for {
				// Recv 逐帧读取；io.EOF 表示本段流结束。
				chunk, err := mo.MessageStream.Recv()
				if err == io.EOF {
					log.Fatal(err.Error())
					break
				}
				if err != nil {
					if ctx.Err() != nil {
						return ctx.Err()
					}
					log.Fatal(err.Error())
					_ = emit(ChatEvent{Type: ChatEventError, Error: err.Error()})
					return err
				}
				// 决定调用工具的帧 Content 为空（ToolCalls 分片），跳过。
				if chunk.Content == "" {
					continue
				}
				full.WriteString(chunk.Content)
				if err := emit(ChatEvent{Type: ChatEventDelta, Delta: chunk.Content}); err != nil {
					log.Fatal(err.Error())
					return err
				}
			}
			continue
		}

		// 非流式兜底：某些模型/配置下整段返回。
		if mo.Message != nil && mo.Message.Content != "" {
			full.WriteString(mo.Message.Content)
			if err := emit(ChatEvent{Type: ChatEventDelta, Delta: mo.Message.Content}); err != nil {
				return err
			}
		}
	}
	return nil
}

// 豆包搜索（问搜文） todo // 后续需要动态注入
// DoubaoSearchInput 工具入参，tag 用于自动生成 JSON Schema
type DoubaoSearchInput struct {
	Query string `json:"query" jsonschema:"required" jsonschema_description:"要搜索的关键词/问题"`
	Count int    `json:"count,omitempty" jsonschema_description:"期望返回的结果条数，不填默认 8"`
}

// webSearchFilter 控制返回内容：不返回正文、返回链接
type webSearchFilter struct {
	NeedContent bool `json:"NeedContent"`
	NeedUrl     bool `json:"NeedUrl"`
}

// webSearchQueryControl 查询改写控制
type webSearchQueryControl struct {
	QueryRewrite bool `json:"QueryRewrite"`
}

// webSearchReq 问搜文 API 请求体，字段名大写开头
type webSearchReq struct {
	Query        string                `json:"Query"`
	SearchType   string                `json:"SearchType"`
	Count        int                   `json:"Count"`
	Filter       webSearchFilter       `json:"Filter"`
	QueryControl webSearchQueryControl `json:"QueryControl"`
}

type searchItem struct {
	Title       string `json:"Title"`
	URL         string `json:"Url"`
	SiteName    string `json:"SiteName"`
	Summary     string `json:"Summary"`
	Snippet     string `json:"Snippet"`
	PublishTime string `json:"PublishTime"`
}

// API返回结构
type searchResp struct {
	ResponseMetadata struct {
		RequestId string `json:"RequestId"`
		Error     *struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error"`
	} `json:"ResponseMetadata"`
	Result struct {
		ResultCount int          `json:"ResultCount"`
		WebResults  []searchItem `json:"WebResults"`
	} `json:"Result"`
}

func NewWebSerach(apikey string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"web_search",
		"联网搜索工具，用于查询互联网最新信息、新闻、实时事实。当问题存在时效性、模型知识截止后内容时调用。参数query填写用户原始问题。",
		func(ctx context.Context, in DoubaoSearchInput) (string, error) {
			if in.Query == "" {
				return "", fmt.Errorf("query 不能为空")
			}
			if in.Count <= 0 {
				// 做个简单的保护
				in.Count = 8
			}
			reqBody, err := json.Marshal(webSearchReq{
				Query:        in.Query,
				SearchType:   "web",
				Count:        in.Count,
				Filter:       webSearchFilter{NeedContent: false, NeedUrl: true},
				QueryControl: webSearchQueryControl{QueryRewrite: false},
			})
			if err != nil {
				return "", fmt.Errorf("marshal req failed: %w", err)
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://open.feedcoopapi.com/search_api/web_search", bytes.NewReader(reqBody))
			if err != nil {
				return "", fmt.Errorf("new request failed: %w", err)
			}
			req.Header.Set("Authorization", "Bearer "+apikey)
			req.Header.Set("Content-Type", "application/json")
			client := http.Client{
				Timeout: 12 * time.Second,
			}
			resp, err := client.Do(req)
			if err != nil {
				return "", fmt.Errorf("do request failed: %w", err)
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", fmt.Errorf("read resp body failed: %w", err)
			}
			var res searchResp
			if err := json.Unmarshal(body, &res); err != nil {
				return "", fmt.Errorf("unmarshal resp failed: %w, raw=%s", err, string(body))
			}
			if errMeta := res.ResponseMetadata.Error; errMeta != nil {
				return "", fmt.Errorf("search api err: code=%s, msg=%s, request_id=%s", errMeta.Code, errMeta.Message, res.ResponseMetadata.RequestId)
			}

			var output strings.Builder
			for idx, item := range res.Result.WebResults {
				summary := item.Summary
				if summary == "" {
					summary = item.Snippet
				}
				fmt.Fprintf(&output, "[%d]标题：%s\n来源：%s\n发布时间：%s\n链接：%s\n摘要：%s\n\n",
					idx+1, item.Title, item.SiteName, item.PublishTime, item.URL, summary)
			}
			if output.Len() == 0 {
				return "未检索到相关网页结果", nil
			}

			return output.String(), nil
		},
	)
}
