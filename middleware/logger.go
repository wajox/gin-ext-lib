package middleware

import (
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LoggerConfig struct {
	Output    io.Writer
	SkipPaths []string
}

func Logger() gin.HandlerFunc {
	return LoggerWithConfig(LoggerConfig{})
}

func LoggerWithWriter(out io.Writer, notlogged ...string) gin.HandlerFunc {
	return LoggerWithConfig(LoggerConfig{
		Output:    out,
		SkipPaths: notlogged,
	})
}

func LoggerWithConfig(conf LoggerConfig) gin.HandlerFunc {
	if conf.Output == nil {
		conf.Output = gin.DefaultWriter
	}

	return newLoggerMiddleware(conf)
}

func newLoggerMiddleware(conf LoggerConfig) gin.HandlerFunc {
	skip := computeSkip(conf)
	log.Logger = zerolog.New(conf.Output)

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log only when path is not being skipped
		if _, ok := skip[path]; ok {
			return
		}

		log.Info().
			Str("StartTimestamp", fmt.Sprintf("%d", start.Unix())).
			Str("ClientIP", c.ClientIP()).
			Str("Method", c.Request.Method).
			Str("Status", fmt.Sprintf("%d", c.Writer.Status())).
			Str("BodySize", fmt.Sprintf("%d", c.Writer.Size())).
			Str("ErrorMessage", c.Errors.ByType(gin.ErrorTypePrivate).String()).
			Str("Path", path).
			Str("Query", raw).
			Msg(path)
	}
}

func computeSkip(conf LoggerConfig) map[string]struct{} {
	notlogged := conf.SkipPaths

	var skip map[string]struct{}

	if length := len(notlogged); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range notlogged {
			skip[path] = struct{}{}
		}
	}

	return skip
}
