package agentkit

import (
	"taie/internal/agentkit/hooks"
	"taie/internal/agentkit/trace"
	"taie/internal/agentkit/workspace"

	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(workspace.NewCheckpointStore, NewChatAgent, trace.NewLogCollectHandler, hooks.BuildHooks)
