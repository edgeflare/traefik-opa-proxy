// Requests to this server on port 8080 return information about the request
// Useful for HTTP request debugging and testing reverse proxies.
package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(LogHeaders)
	e.Use(middleware.Logger())

	// Add CORS middleware
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Listen to all routes and return request data
	e.Any("/*", func(c echo.Context) error {
		// Read request body
		bodyBytes, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		// Get form parameters and handle potential error
		formParams, err := c.FormParams()
		if err != nil {
			return err
		}

		// Build data to send in response
		request := map[string]interface{}{
			"headers":          c.Request().Header,
			"query_params":     c.QueryParams(),
			"form_params":      formParams,
			"request_body":     string(bodyBytes),
			"request_method":   c.Request().Method,
			"request_uri":      c.Request().RequestURI,
			"request_path":     c.Request().URL.Path,
			"request_host":     c.Request().Host,
			"request_protocol": c.Request().Proto,
		}

		return c.JSON(http.StatusOK, request)
	})

	e.Start(":8080")
}

func LogHeaders(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		for name, headers := range req.Header {
			for _, h := range headers {
				fmt.Printf("%v: %v\n", name, h)
			}
		}
		return next(c)
	}
}
