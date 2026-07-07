package logging

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const traceIDHeader = "X-Trace-ID"

func RequestLogger(baseLogger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader(traceIDHeader)
		if traceID == "" {
			traceID = fmt.Sprintf("%d", time.Now().UnixNano())
		}

		c.Writer.Header().Set(traceIDHeader, traceID)

		requestLogger := baseLogger.With(
			zap.String("trace_id", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.FullPath()),
			zap.String("client_ip", c.ClientIP()),
		)

		c.Request = c.Request.WithContext(WithLogger(c.Request.Context(), requestLogger))
		c.Set("trace_id", traceID)

		start := time.Now()
		c.Next()

		requestLogger.Info("http_request",
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.Int("response_size", c.Writer.Size()),
		)
	}
}
