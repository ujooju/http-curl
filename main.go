package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	httpcurl "github.com/ujooju/http-curl/lib"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	hideCurlOptions := os.Getenv("HIDE_CURL_OPTIONS")
	if hideCurlOptions == "true" {
		httpcurl.SetPrintArgs(false)
	}

	e := echo.New()
	e.HideBanner = true

	// Add middlewares
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	// Dynamic route for proxying
	e.POST("/curl", handleCurl)

	e.Any("/waiting/:milli", func(c echo.Context) error {
		milliStr := c.Param("milli")
		milli, err := strconv.Atoi(milliStr)
		if err != nil || milli < 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid milliseconds"})
		}

		// Convert milliseconds to duration
		time.Sleep(time.Duration(milli) * time.Millisecond)

		return c.String(http.StatusOK, "Ok")
	})

	e.Logger.Fatal(e.Start(":" + port))
}

func handleCurl(c echo.Context) error {
	if c.Request().Header.Get("Content-Type") != "application/json" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Content-Type must be application/json"})
	}

	timeout := 10 * time.Second
	// Check if the timeout query parameter is provided
	timeoutStr := c.QueryParam("timeout")
	if timeoutStr != "" {
		// Parse the duration string
		timeoutDuration, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Error parsing timeout duration"})
		}
		timeout = timeoutDuration
	}

	var reqData httpcurl.CurlOption
	if err := c.Bind(&reqData); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON input"})
	}

	output, err := httpcurl.HttpCurl(reqData, timeout)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error":   fmt.Sprintf("Error executing curl command: %v", err),
			"details": string(output),
		})
	}

	formattedOutput := string(output)

	base64format := c.QueryParam("base64")
	if base64format == "true" {
		formattedOutput = base64.StdEncoding.EncodeToString(output)
	}

	plain := c.QueryParam("plain")
	if plain == "true" {
		accept := c.Request().Header.Get("Accept")
		// return the output as is, it could be anything
		return c.Blob(http.StatusOK, accept, []byte(formattedOutput))
	}

	// Return the output of the curl command
	return c.JSON(http.StatusOK, map[string]string{"result": formattedOutput})
}
