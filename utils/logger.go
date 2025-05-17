package utils

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type (
	KVAny struct {
		Key string
		Any interface{}
	}

	Skipper func(c echo.Context) bool

	Logger struct {
		logger  zerolog.Logger
		skipper Skipper
	}
)

func NewLogger(logger zerolog.Logger, skipper Skipper) *Logger {
	return &Logger{
		logger:  logger,
		skipper: skipper,
	}
}

func DefaultSkipper(_ echo.Context) bool {
	return false
}

func (l *Logger) Info(kvAny *[]KVAny, msg string, fields ...interface{}) {
	info := l.logger.Info().Stack()
	if kvAny != nil {
		for _, kv := range *kvAny {
			info = info.Any(kv.Key, kv.Any)
		}
	}
	info.Msgf(msg, fields...)
}

func (l *Logger) Warn(kvAny *[]KVAny, msg string, fields ...interface{}) {
	warn := l.logger.Warn().Stack()
	if kvAny != nil {
		for _, kv := range *kvAny {
			warn = warn.Any(kv.Key, kv.Any)
		}
	}
	warn.Msgf(msg, fields...)
}

func (l *Logger) Error(kvAny *[]KVAny, msg error) {
	//nolint:zerologlint
	logErr := l.logger.Error().Stack()
	if kvAny != nil {
		for _, kv := range *kvAny {
			logErr = logErr.Any(kv.Key, kv.Any)
		}
	}
	logErr.Err(msg).Msg("")
}

func (l *Logger) Debug(kvAny *[]KVAny, msg string, fields ...interface{}) {
	//nolint:zerologlint
	debug := l.logger.Debug().Stack()
	if kvAny != nil {
		for _, kv := range *kvAny {
			debug = debug.Any(kv.Key, kv.Any)
		}
	}
	debug.Msgf(msg, fields...)
}

func (l *Logger) Trace(kvAny *[]KVAny, msg string, fields ...interface{}) {
	trace := l.logger.Trace().Stack()
	if kvAny != nil {
		for _, kv := range *kvAny {
			trace = trace.Any(kv.Key, kv.Any)
		}
	}
	trace.Msgf(msg, fields...)
}

func (l *Logger) Fatal(kvAny *[]KVAny, msg error) {
	//nolint:zerologlint
	fatal := l.logger.Fatal().Stack()
	if kvAny != nil {
		for _, kv := range *kvAny {
			fatal = fatal.Any(kv.Key, kv.Any)
		}
	}
	fatal.Err(msg).Msg("")
}

func (l *Logger) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if l.skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()
			start := time.Now()
			if err = next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			xRequestID := req.Header.Get(echo.HeaderXRequestID)
			if xRequestID == "" {
				xRequestID = res.Header().Get(echo.HeaderXRequestID)
			}

			p := req.URL.Path
			if p == "" {
				p = "/"
			}

			kvAny := []KVAny{
				{
					Key: "x-request-id",
					Any: xRequestID,
				},
				{
					Key: "remote_ip",
					Any: c.RealIP(),
				},
				{
					Key: "host",
					Any: req.Host,
				},
				{
					Key: "uri",
					Any: req.RequestURI,
				},
				{
					Key: "method",
					Any: req.Method,
				},
				{
					Key: "path",
					Any: p,
				},
				{
					Key: "route",
					Any: c.Path(),
				},
				{
					Key: "protocol",
					Any: req.Proto,
				},
				{
					Key: "referer",
					Any: req.Referer(),
				},
				{
					Key: "user_agent",
					Any: req.UserAgent(),
				},
				{
					Key: "status",
					Any: res.Status,
				},
				{
					Key: "error",
					Any: err,
				},
				{
					Key: "latency",
					Any: stop.Sub(start),
				},
				{
					Key: "latency_human",
					Any: stop.Sub(start).String(),
				},
				{
					Key: "bytes_in",
					Any: req.Header.Get(echo.HeaderContentLength),
				},
				{
					Key: "bytes_out",
					Any: res.Size,
				},
				{
					Key: "query",
					Any: req.URL.Query(),
				},
			}

			l.Trace(&kvAny, "http request: %s", req.Method)
			return
		}
	}
}
