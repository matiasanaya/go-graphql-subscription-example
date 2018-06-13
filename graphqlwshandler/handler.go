package graphqlwshandler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/matiasanaya/graphql-transport-ws/graphqlws"
	"github.com/matiasanaya/graphql-transport-ws/graphqlws/event"
)

// NewHandler returns a new Handler that supports both websocket and http transports
func NewHandler(s *graphql.Schema, httpHandler http.Handler) http.HandlerFunc {
	wsHandler := NewDefaultHandler(s)
	return func(w http.ResponseWriter, r *http.Request) {
		for _, subprotocol := range websocket.Subprotocols(r) {
			if subprotocol == "graphql-ws" {
				wsHandler.ServeHTTP(w, r)
				return
			}
		}
		httpHandler.ServeHTTP(w, r)
	}
}

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
	b, err := json.Marshal(args.StartMessage.Variables)
	if err != nil {
		return nil, nil, err
	}

	variables := map[string]interface{}{}
	err = json.Unmarshal(b, &variables)
	if err != nil {
		return nil, nil, err
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
