package toolkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"taie/internal/po"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type WebSearchToolConf struct {
	ApiKey string
	Url    string
}

type DoubaoSearchInput struct {
	Query string `json:"query" jsonschema:"required" jsonschema_description:"要搜索的关键词/问题"`
	Count int    `json:"count,omitempty" jsonschema_description:"期望返回的结果条数，不填默认 10,最大50"`
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

func buildWebSearch(data []po.CommonJson) (tool.InvokableTool, error) {
	var conf WebSearchToolConf
	for _, item := range data {
		if item.Label == "apiKey" {
			conf.ApiKey = item.Value
		}
		if item.Label == "url" {
			conf.Url = item.Value
		}
	}
	if conf.ApiKey == "" || conf.Url == "" {
		return nil, fmt.Errorf("解析websearch工具错误")
	}
	return utils.InferTool(
		"web_search",
		"联网搜索工具，用于查询互联网最新信息、新闻、实时事实。当问题存在时效性、模型知识截止后内容时调用。参数query填写用户原始问题。",
		func(ctx context.Context, in DoubaoSearchInput) (string, error) {
			if in.Query == "" {
				return "", fmt.Errorf("query 不能为空")
			}
			if in.Count <= 0 {
				// 做个简单的保护
				in.Count = 10
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
			// https://open.feedcoopapi.com/search_api/web_search
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, conf.Url, bytes.NewReader(reqBody))
			if err != nil {
				return "", fmt.Errorf("new request failed: %w", err)
			}
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", conf.ApiKey))
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
