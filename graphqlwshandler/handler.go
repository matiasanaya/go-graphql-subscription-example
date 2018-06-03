package graphqlwshandler

import (
	"context"
	"encoding/json"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/matiasanaya/graphql-transport-ws/graphqlws"
	"github.com/matiasanaya/graphql-transport-ws/graphqlws/event"
)

// NewDefaultHandler returns a new Handler with default callbacks
func NewDefaultHandler(s *graphql.Schema) *graphqlws.Handler {
	return graphqlws.NewHandler(newDefaultCallback(s))
}

type defaultCallback struct {
	schema *graphql.Schema
}

func newDefaultCallback(schema *graphql.Schema) *defaultCallback {
	return &defaultCallback{schema: schema}
}

func (h *defaultCallback) OnOperation(ctx context.Context, args *event.OnOperationArgs) (json.RawMessage, func(), error) {
	variables := map[string]interface{}{}
	for k, v := range args.StartMessage.Variables {
		variables[k] = v
	}

	ctx, cancel := context.WithCancel(ctx)
	c, err := h.schema.Subscribe(ctx, args.StartMessage.Query, args.StartMessage.OperationName, variables)
	if err != nil {
		cancel()
		return nil, nil, err
	}

	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case response, more := <-c:
				if !more {
					return
				}
				responseJSON, err := json.Marshal(response)
				if err != nil {
					args.Send(json.RawMessage(`{"errors":["internal error: can't marshal response into json"]}`))
					continue
				}
				args.Send(responseJSON)
			}
		}
	}()

	return nil, cancel, nil
}
