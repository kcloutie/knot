package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/kcloutie/knot/pkg/config"
	knothttp "github.com/kcloutie/knot/pkg/http"
	"github.com/kcloutie/knot/pkg/listener"
	uuid "github.com/satori/go.uuid"
)

func CreateRouter(ctx context.Context, cacheInSeconds int) *gin.Engine {
	router := gin.Default()

	router.Use(RequestIdMiddleware())
	router.Use(TraceLogsMiddleware())
	memoryStore := persist.NewMemoryStore(time.Duration(cacheInSeconds) * time.Second)

	router.GET("", cache.CacheByRequestURI(memoryStore, time.Duration(cacheInSeconds)*time.Second), func(c *gin.Context) {
		Home(ctx, c)
	})

	router.GET("/healthz", Health)

	apiV1 := router.Group("/api/v1")
	apiV1.Use()
	{
		for _, l := range listener.GetListeners() {
			apiV1.POST(fmt.Sprintf("/%s", l.GetApiPath()), func(c *gin.Context) {
				ExecuteListener(ctx, c, l)
			})
		}
	}
	return router
}

func Start(ctx context.Context, router *gin.Engine, cfg *config.ServerConfiguration, listeningAddr string) error {
	knothttp.TraceHeaderKey = cfg.TraceHeaderKey

	server := &http.Server{
		Addr:              listeningAddr,
		ReadHeaderTimeout: 3 * time.Second,
		Handler:           router,
	}

	err := server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func RequestIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := uuid.NewV4().String()
		c.Set(knothttp.RequestHeaderKey, rid)
		c.Writer.Header().Set(knothttp.RequestHeaderKey, rid)
		c.Next()
	}
}

func TraceLogsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Request.Header[knothttp.TraceHeaderKey]
		if exists {
			if len(val) == 1 {
				c.Set(knothttp.TraceHeaderKey, val[0])
			}
		}
		c.Next()
	}
}
