package graphqlgin

import (
	"context"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/ianidi/exchange-server/graph"
	"github.com/ianidi/exchange-server/graph/generated"
	"github.com/ianidi/exchange-server/internal/jwt"
)

// // Defining the Graphql handler
// func GraphqlHandler() gin.HandlerFunc {
// 	// NewExecutableSchema and Config are in the generated.go file
// 	// Resolver is in the resolver.go file
// 	h := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

// 	return func(c *gin.Context) {
// 		h.ServeHTTP(c.Writer, c.Request)
// 	}
// }

func New() generated.Config {
	return generated.Config{
		Resolvers: &graph.Resolver{
			// Rooms: map[string]*Chatroom{},
		},
		// Directives: &graph.DirectiveRoot{
		// 	User: func(ctx context.Context, obj interface{}, next graphql.Resolver, username string) (res interface{}, err error) {
		// 		return next(context.WithValue(ctx, "username", username))
		// 	},
		// },
	}
}

func GraphqlHandler(h *jwt.ProfileHandler, private bool) gin.HandlerFunc {
	c := generated.Config{Resolvers: &graph.Resolver{
		ProfileHandler: h,
		// DB:    DB,
		// Redis: Redis,

	}}
	// c.Directives.Meta = func(ctx context.Context, obj interface{}, next graphql.Resolver, json *string, gorm *string, validate *string) (res interface{}, err error) {
	// 	return next(ctx)
	// }
	// c.Directives.Permission = func(ctx context.Context, obj interface{}, next graphql.Resolver, name *string) (res interface{}, err error) {
	// 	user := auth.UserFromContext(ctx)
	// 	if name != nil {
	// 		if !UserHasPermission(*name, user) {
	// 			return nil, errors.New("access denied")
	// 		}
	// 	}
	// 	return next(ctx)
	// }
	// h := handler.NewDefaultServer(generated.NewExecutableSchema(c))

	srv := handler.New(generated.NewExecutableSchema(c))
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Websocket{
		KeepAlivePingInterval: 10 * time.Second,
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})
	srv.Use(extension.Introspection{})

	// WebsocketInitFunc(func(ctx context.Context, initPayload InitPayload) (context.Context, error) {
	// 	userId, err := validateAndGetUserID(payload["token"])
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	// get the user from the database
	// 	user := getUserByID(db, userId)

	// 	// put it in context
	// 	userCtx := context.WithValue(r.Context(), userCtxKey, user)

	// 	// and return it so the resolvers can see it
	// 	return userCtx, nil
	// }))

	// h.AddTransport(&transport.Websocket{
	// 	Upgrader: websocket.Upgrader{
	// 		CheckOrigin: func(r *http.Request) bool {
	// 			return true
	// 		},
	// 		ReadBufferSize:  1024,
	// 		WriteBufferSize: 1024,
	// 	},
	// 	KeepAlivePingInterval: 10 * time.Second,
	// })

	//handler.WebsocketKeepAliveDuration(10 * time.Second)
	return func(c *gin.Context) {

		// wsauth := defaultHandler.GetInitPayload(c).GetString("Authorization")
		// fmt.Println("wsauth", wsauth)

		//If the route is protected, check auth token
		if private {
			err := jwt.TokenValid(c.Request)
			if err != nil {
				c.JSON(http.StatusUnauthorized, "unauthorized")
				c.Abort()
				return
			}

			metadata, err := h.TK.ExtractTokenMetadata(c.Request)
			if err != nil {
				c.JSON(http.StatusUnauthorized, "unauthorized")
				return
			}
			// userId, err := h.RD.FetchAuth(metadata.TokenUuid)
			// if err != nil {
			// 	c.JSON(http.StatusUnauthorized, "unauthorized")
			// 	return
			// }

			// fmt.Println(c.FullPath(), userId)

			c.Set(jwt.MemberKey, metadata.UserId)
		}

		srv.ServeHTTP(c.Writer, c.Request)
	}
}

// Defining the Playground handler
func PlaygroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), "GinContextKey", c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
