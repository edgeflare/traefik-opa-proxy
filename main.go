package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/urfave/cli/v2"
)

const (
	defaultOPAURL      = "http://opa:8181/v1/data/httpapi/authz"
	defaultServicePort = 8182
)

type Config struct {
	OPA_URL string
	Port    int
}

type OPAResponse struct {
	Result struct {
		Allow bool `json:"allow"`
	} `json:"result"`
}

func main() {
	app := &cli.App{
		Name:    "traefik-opa-proxy",
		Usage:   "Translates OPA's decisions (allow: true or false) into HTTP status codes (200 or 403)",
		Version: "v0.0.1",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "opa-url",
				Value:   defaultOPAURL,
				Usage:   "URL for the Open Policy Agent",
				EnvVars: []string{"OPA_URL"},
			},
			&cli.IntFlag{
				Name:    "port",
				Value:   defaultServicePort,
				Usage:   "Port to run the service on",
				EnvVars: []string{"SERVICE_PORT"},
			},
		},
		Action: startProxy,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Failed to start the app: %v", err)
	}
}

func startProxy(c *cli.Context) error {
	config := Config{
		OPA_URL: c.String("opa-url"),
		Port:    c.Int("port"),
	}

	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())
	e.Use(middleware.Recover())
	e.Any("/*", getAuthorizationFromOPA(config.OPA_URL))

	return e.Start(fmt.Sprintf(":%d", config.Port))
}

func getAuthorizationFromOPA(OPAURL string) echo.HandlerFunc {

	return func(c echo.Context) error {
		payload, err := buildPayload(c)
		if err != nil {
			return echoResponse(c, http.StatusInternalServerError, "Failed to build payload")
		}

		respBody, err := sendRequestToOPA(payload, OPAURL)
		if err != nil {
			return echoResponse(c, http.StatusInternalServerError, "Error interacting with OPA")
		}

		return determineResponse(c, respBody)
	}
}

func buildPayload(c echo.Context) (map[string]interface{}, error) {
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return nil, err
	}

	formParams, err := c.FormParams()
	if err != nil {
		return nil, err
	}

	method := c.Request().Header.Get("X-Forwarded-Method")
	if method == "" {
		method = c.Request().Method
	}

	host := c.Request().Header.Get("X-Forwarded-Host")
	if host == "" {
		host = c.Request().Host
	}

	path := c.Request().Header.Get("X-Forwarded-Uri")
	if path == "" {
		path = c.Request().URL.Path
	}

	protocol := c.Request().Header.Get("X-Forwarded-Proto")
	if path == "" {
		protocol = c.Request().Proto
	}

	// Build the payload to match OPA-Envoy's expected input
	payload := map[string]interface{}{
		"attributes": map[string]interface{}{
			"request": map[string]interface{}{
				"http": map[string]interface{}{
					"method": method,
					"scheme": c.Request().URL.Scheme, // TODO: Is this correct?
					"host":   host,
					"path":   path,
					"headers": map[string]interface{}{
						"Authorization": c.Request().Header.Get("Authorization"),
						// Add other headers if necessary
					},
					"body":         string(bodyBytes),
					"query_params": c.QueryParams(),
					"form_params":  formParams,
					"protocol":     protocol,
				},
			},
		},
	}

	return payload, nil
}

func sendRequestToOPA(payload map[string]interface{}, OpaURL string) ([]byte, error) {
	// Convert the map to a JSON string
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// URL-escape the JSON string
	escapedString := url.QueryEscape(string(jsonBytes))

	// Construct the full URL with the encoded query parameter
	fullURL := fmt.Sprintf("%s?input=%s", OpaURL, escapedString)

	// Make the GET request
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and return the response body
	return io.ReadAll(resp.Body)
}

func determineResponse(c echo.Context, body []byte) error {
	var opaResp OPAResponse
	if err := json.Unmarshal(body, &opaResp); err != nil {
		return echoResponse(c, http.StatusInternalServerError, "Unknown OPA response format")
	}

	if opaResp.Result.Allow {
		return echoResponse(c, http.StatusOK, "ok")
	}

	fmt.Println(opaResp)

	return echoResponse(c, http.StatusForbidden, "forbidden")
}

func echoResponse(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, map[string]string{"message": message})
}
