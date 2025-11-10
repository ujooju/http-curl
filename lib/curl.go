package httpcurl

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// CurlValue represents a curl option's value, supporting a single string or multiple strings.
type CurlValue []string

var printArgs = true

// SetPrintArgs sets whether to print curl arguments
func SetPrintArgs(print bool) {
	printArgs = print
}

// UnmarshalJSON customizes the unmarshalling of CurlValue to handle string or []string.
func (cv *CurlValue) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a single string
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*cv = CurlValue{single}
		return nil
	}

	// Try to unmarshal as a slice of strings
	var multiple []string
	if err := json.Unmarshal(data, &multiple); err == nil {
		*cv = CurlValue(multiple)
		return nil
	}

	// Return an error if neither works
	return fmt.Errorf("invalid value for CurlValue, must be string or []string")
}

// CurlOption represents the JSON structure for curl parameters.
type CurlOption map[string]CurlValue

// AllowedCurlOptions is a predefined list of safe curl options
var AllowedCurlOptions = map[string]bool{
	"-k":         true, // Skip SSL verification
	"-x":         true, // HTTP Proxy
	"-X":         true, // HTTP method
	"-d":         true, // Data payload
	"--data":     true, // Data payload (alternative)
	"--location": true, // Follow redirects
	"-H":         true, // HTTP headers
	"--tls-max":  true, // Set max tls version
}

// sanitizeInput validates and restricts the curl options
func sanitizeInput(input CurlOption) ([]string, error) {
	var args []string

	for key, values := range input {
		// Validate that the option is allowed
		if !AllowedCurlOptions[key] {
			return nil, fmt.Errorf("unauthorized curl option: %s", key)
		}

		// Add each value associated with the key to the arguments
		for _, value := range values {
			// Handle standalone options
			if value == "" || value == "true" {
				args = append(args, key)
			} else {
				args = append(args, key, value)
			}
		}
	}

	return args, nil
}

func HttpCurl(options CurlOption, timeout time.Duration) (output []byte, err error) {
	// Sanitize and validate input
	curlArgs, err := sanitizeInput(options)
	if err != nil {
		return
	}

	// Add the silent flag to suppress the progress bar
	curlArgs = append([]string{"-s"}, curlArgs...)

	if printArgs {
		fmt.Println("curlArgs :", curlArgs)
	}

	// Create a context with a timeout to prevent long-running commands
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Execute the curl command
	cmd := exec.CommandContext(ctx, "curl", curlArgs...)
	output, err = cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		err = fmt.Errorf("request timed out")
		return
	}

	return
}
