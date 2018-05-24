package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/functionalfoundry/graphqlws"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

const schema = `
	schema {
		subscription: Subscription
		mutation: Mutation
		query: Query
	}

	type Query {
		hello: String!
	}

	type Subscription {
		helloSaid(): HelloSaidEvent!
	}

	type Mutation {
		sayHello(msg: String!): HelloSaidEvent!
	}

	type HelloSaidEvent {
		id: String!
		msg: String!
	}
`

func main() {
	// graphiql handler
	http.HandleFunc("/", http.HandlerFunc(graphiql))

	// init graphQL schema
	s, err := graphql.ParseSchema(schema, newResolver())
	if err != nil {
		panic(err)
	}

	// graphQL query & mutation handler
	queryHandler := &relay.Handler{Schema: s}
	http.HandleFunc("/graphql", queryHandler.ServeHTTP)

	// graphQL subscription handler
	m := newSubscriptionsManager(s)
	subscriptionHandler := graphqlws.NewHandler(graphqlws.HandlerConfig{SubscriptionManager: m})
	http.HandleFunc("/subscriptions", subscriptionHandler.ServeHTTP)

	// start HTTP server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

type subscriptionsManager struct {
	graphqlws.SubscriptionManager
	s *graphql.Schema
}

func newSubscriptionsManager(s *graphql.Schema) *subscriptionsManager {
	c := make(chan *graphqlws.Subscription)
	m := subscriptionsManager{
		s: s,
		SubscriptionManager: graphqlws.NewSubscriptionManager(
			func(subscription *graphqlws.Subscription) {
				c <- subscription
			},
		),
	}
	go m.initSubscriptions(c)

	return &m
}

func (m *subscriptionsManager) initSubscriptions(subscriptions <-chan *graphqlws.Subscription) {
	for subscription := range subscriptions {
		ctx := context.Background()
		c, _ := m.s.Subscribe(ctx, subscription.Query, subscription.OperationName, subscription.Variables)

		localSubscription := subscription
		go func() {
			for {
				select {
				case <-localSubscription.StopCh():
					fmt.Println("Should shutdown upstream sub:", localSubscription.ID)
					return
				case resp := <-c:
					data := graphqlws.DataMessagePayload{
						Data: resp.Data,
						// TODO: send errors
						// Errors: resp.Errors,
					}
					localSubscription.SendData(&data)
				}
			}
		}()
	}
}

type resolver struct {
	helloSaidEvents chan *helloSaidEvent
}

func newResolver() *resolver {
	return &resolver{helloSaidEvents: make(chan *helloSaidEvent)}
}

func (r *resolver) Hello() string {
	return "Hello world!"
}

func (r *resolver) SayHello(args struct{ Msg string }) *helloSaidEvent {
	e := &helloSaidEvent{msg: args.Msg, id: randomID()}
	go func() {
		select {
		case r.helloSaidEvents <- e:
		case <-time.After(1 * time.Second):
		}
	}()
	return e
}

func (r *resolver) HelloSaid() <-chan *helloSaidEvent {
	return r.helloSaidEvents
}

type helloSaidEvent struct {
	id  string
	msg string
}

func (r *helloSaidEvent) Msg() string {
	return r.msg
}

func (r *helloSaidEvent) ID() string {
	return r.id
}

func randomID() string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, 16)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

var graphiql = func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`
  <!DOCTYPE html>
  <html>
       <head>
               <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.10/graphiql.css" />
               <script src="https://cdnjs.cloudflare.com/ajax/libs/fetch/1.1.0/fetch.min.js"></script>
               <script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react.min.js"></script>
               <script src="https://cdnjs.cloudflare.com/ajax/libs/react/15.5.4/react-dom.min.js"></script>
               <script src="https://cdnjs.cloudflare.com/ajax/libs/graphiql/0.11.10/graphiql.js"></script>
               <script src="//unpkg.com/subscriptions-transport-ws@0.8.3/browser/client.js"></script>
               <script src="//unpkg.com/graphiql-subscriptions-fetcher@0.0.2/browser/client.js"></script>
       </head>
       <body style="width: 100%; height: 100%; margin: 0; overflow: hidden;">
               <div id="graphiql" style="height: 100vh;">Loading...</div>
               <script>
                       function graphQLFetcher(graphQLParams) {
                               return fetch("/graphql", {
                                       method: "post",
                                       body: JSON.stringify(graphQLParams),
                                       credentials: "include",
                               }).then(function (response) {
                                       return response.text();
                               }).then(function (responseBody) {
                                       try {
                                               return JSON.parse(responseBody);
                                       } catch (error) {
                                               return responseBody;
                                       }
                               });
                       }

                       var subscriptionsClient = new window.SubscriptionsTransportWs.SubscriptionClient('ws://localhost:8080/subscriptions', { reconnect: true });
                       var subscriptionsFetcher = window.GraphiQLSubscriptionsFetcher.graphQLFetcher(subscriptionsClient, graphQLFetcher);

                       ReactDOM.render(
                               React.createElement(GraphiQL, {fetcher: subscriptionsFetcher}),
                               document.getElementById("graphiql")
                       );
               </script>
       </body>
  </html>
  `))
}
