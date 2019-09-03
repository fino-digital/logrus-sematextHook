package sematextHook

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/neko-neko/echo-logrus/v2/log"
)

// middleware that produces accessLog messages with additional fields
//noinspection GoUnusedExportedFunction
func AccessLog() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			var err error
			if err = next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}
			reqSize := req.Header.Get(echo.HeaderContentLength)
			if reqSize == "" {
				reqSize = "0"
			}

			requestDuration := stop.Sub(start)
			responseLength := strconv.FormatInt(res.Size, 10)
			log.Logger().
				WithFields(map[string]interface{}{
					"RequestId":     id,
					"remoteAddress": c.RealIP(),
					"timestamp":     stop.UnixNano(),

					"httpMethod":      req.Method,
					"protocol":        req.Proto,
					"requestLength":   reqSize,
					"requestUri":      req.RequestURI,
					"virtualhost":     req.Host,
					"userAgent":       req.UserAgent(),
					"requestEncoding": extractEncoding(req.Header),

					"responseContentType": res.Header().Get(echo.HeaderContentType),
					"responseEncoding":    extractEncoding(res.Header()),
					"responseLength":      responseLength,
					"responseStatus":      res.Status,
					"responseTimeNanos":   requestDuration.Nanoseconds(),
				}).
				Infof("%s %s [%v] %s %-7s %s %3d %s %s %13v %s %s",
					id,
					c.RealIP(),
					stop.Format(time.RFC3339),
					req.Host,
					req.Method,
					req.RequestURI,
					res.Status,
					reqSize,
					responseLength,
					requestDuration.String(),
					req.Referer(),
					req.UserAgent(),
				)
			return err
		}
	}
}

func extractEncoding(header http.Header) string {
	encoding := header.Get(echo.HeaderContentEncoding)
	if encoding != "" {
		return strings.ToLower(encoding)
	}
	contentType := header.Get(echo.HeaderContentType)
	if contentType != "" {
		split := strings.Split(contentType, ";")
		if len(split) > 1 {
			for _, v := range split {
				if strings.Contains(v, "charset=") {
					after := strings.SplitAfter(v, "charset=")
					return strings.ToLower(after[len(after)-1])
				}
			}
		}
	}
	return ""
}
