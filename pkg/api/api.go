package api

import (
	"context"
	"net/http"
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
)

var config *ServerConfiguration

func (c *ServerConfiguration) Start(ctx context.Context, listeningAddr string, cacheInSeconds int) error {
	r := gin.Default()

	memoryStore := persist.NewMemoryStore(time.Duration(cacheInSeconds) * time.Second)

	r.GET("", cache.CacheByRequestURI(memoryStore, time.Duration(cacheInSeconds)*time.Second), func(c *gin.Context) {
		Home(ctx, c)
	})

	r.GET("/healthz", Health)

	apiV1 := r.Group("/api/v1")
	apiV1.Use()
	{

	}

	server := &http.Server{
		Addr:              listeningAddr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           r,
	}

	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func NewServerConfiguration() *ServerConfiguration {
	return &ServerConfiguration{}
}

func FromCtx(ctx context.Context) *ServerConfiguration {
	if l, ok := ctx.Value(ctxConfigKey{}).(*ServerConfiguration); ok {
		return l
	} else if l := config; l != nil {
		return l
	}
	return NewServerConfiguration()
}

func WithCtx(ctx context.Context, l *ServerConfiguration) context.Context {
	if lp, ok := ctx.Value(ctxConfigKey{}).(*ServerConfiguration); ok {
		if lp == l {
			return ctx
		}
	}
	return context.WithValue(ctx, ctxConfigKey{}, l)
}
