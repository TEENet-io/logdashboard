package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	sls20201230 "github.com/alibabacloud-go/sls-20201230/v6/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials"
)

// Config holds the configuration for Aliyun SLS client
type Config struct {
	AccessKeyID     string
	AccessKeySecret string
	Endpoint        string
	Project         string
	Logstore        string
}

// LogQuery represents a log query request
type LogQuery struct {
	From  int32
	To    int32
	Query string
}

// AliyunLogClient wraps the SLS client with additional functionality
type AliyunLogClient struct {
	client  *sls20201230.Client
	config  *Config
	runtime *util.RuntimeOptions
	headers map[string]*string
}

// NewAliyunLogClient creates a new Aliyun SLS client with the given configuration
// It uses environment variables for security instead of hardcoded credentials
func NewAliyunLogClient() (*AliyunLogClient, error) {
	config := &Config{
		AccessKeyID:     getEnvOrDefault("ALIYUN_ACCESS_KEY_ID", ""),
		AccessKeySecret: getEnvOrDefault("ALIYUN_ACCESS_KEY_SECRET", ""),
		Endpoint:        getEnvOrDefault("ALIYUN_SLS_ENDPOINT", "ap-southeast-1.log.aliyuncs.com"),
		Project:         getEnvOrDefault("ALIYUN_SLS_PROJECT", "userlogsystem"),
		Logstore:        getEnvOrDefault("ALIYUN_SLS_LOGSTORE", "userlog"),
	}

	// Validate required configuration
	if config.AccessKeyID == "" || config.AccessKeySecret == "" {
		return nil, fmt.Errorf("ALIYUN_ACCESS_KEY_ID and ALIYUN_ACCESS_KEY_SECRET environment variables are required")
	}

	client, err := createSLSClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create SLS client: %w", err)
	}

	return &AliyunLogClient{
		client:  client,
		config:  config,
		runtime: &util.RuntimeOptions{},
		headers: make(map[string]*string),
	}, nil
}

// createSLSClient initializes the Aliyun SLS client with credentials
func createSLSClient(config *Config) (*sls20201230.Client, error) {
	// Create credential configuration
	credConfig := new(credentials.Config).
		SetType("access_key").
		SetAccessKeyId(config.AccessKeyID).
		SetAccessKeySecret(config.AccessKeySecret)

	// Initialize credential
	cred, err := credentials.NewCredential(credConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	// Create OpenAPI configuration
	clientConfig := &openapi.Config{
		Credential: cred,
		Endpoint:   tea.String(config.Endpoint),
	}

	// Create and return SLS client
	client, err := sls20201230.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create SLS client: %w", err)
	}

	return client, nil
}

// QueryLogs executes a log query and returns the results
func (c *AliyunLogClient) QueryLogs(query *LogQuery) (*sls20201230.GetLogsResponse, error) {
	if query == nil {
		return nil, fmt.Errorf("query cannot be nil")
	}

	// Create log query request
	request := &sls20201230.GetLogsRequest{
		From:  tea.Int32(query.From),
		To:    tea.Int32(query.To),
		Query: tea.String(query.Query),
	}

	// Execute query with error handling
	response, err := c.executeQuery(request)
	if err != nil {
		return nil, fmt.Errorf("failed to execute log query: %w", err)
	}

	return response, nil
}

// executeQuery performs the actual API call with proper error handling
func (c *AliyunLogClient) executeQuery(request *sls20201230.GetLogsRequest) (*sls20201230.GetLogsResponse, error) {
	var response *sls20201230.GetLogsResponse
	var err error

	// Execute the query with recovery mechanism
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic occurred during query execution: %v", r)
		}
	}()

	response, err = c.client.GetLogsWithOptions(
		tea.String(c.config.Project),
		tea.String(c.config.Logstore),
		request,
		c.headers,
		c.runtime,
	)

	if err != nil {
		return nil, c.handleAPIError(err)
	}

	return response, nil
}

// handleAPIError processes and formats API errors
func (c *AliyunLogClient) handleAPIError(err error) error {
	if sdkError, ok := err.(*tea.SDKError); ok {
		// Extract error details
		errorMsg := tea.StringValue(sdkError.Message)

		// Try to extract recommendation from error data
		if sdkError.Data != nil {
			var data interface{}
			decoder := json.NewDecoder(strings.NewReader(tea.StringValue(sdkError.Data)))
			if decoder.Decode(&data) == nil {
				if dataMap, ok := data.(map[string]interface{}); ok {
					if recommend, exists := dataMap["Recommend"]; exists {
						return fmt.Errorf("API error: %s, recommendation: %v", errorMsg, recommend)
					}
				}
			}
		}

		return fmt.Errorf("API error: %s", errorMsg)
	}

	return fmt.Errorf("unknown error: %w", err)
}

// PrintLogResults prints the log query results in a formatted way
func (c *AliyunLogClient) PrintLogResults(response *sls20201230.GetLogsResponse) {
	if response == nil || response.Body == nil {
		fmt.Println("No log data returned")
		return
	}

	fmt.Printf("Query executed successfully\n")
	fmt.Printf("Project: %s\n", c.config.Project)
	fmt.Printf("Logstore: %s\n", c.config.Logstore)

	// Extract and print only content fields
	c.extractAndPrintContent(response)
}

// extractAndPrintContent extracts and prints only the content field from log entries
func (c *AliyunLogClient) extractAndPrintContent(response *sls20201230.GetLogsResponse) {
	if response.Body == nil {
		fmt.Println("No response body found")
		return
	}

	// First try to print the raw response structure for debugging
	bodyBytes, err := json.Marshal(response.Body)
	if err != nil {
		fmt.Printf("Failed to marshal response body: %v\n", err)
		fmt.Printf("Response Body Type: %T\n", response.Body)
		fmt.Printf("Response Body: %+v\n", response.Body)
		return
	}

	// Parse the response body as JSON to extract log entries
	var responseData interface{}
	if err := json.Unmarshal(bodyBytes, &responseData); err != nil {
		fmt.Printf("Failed to unmarshal response: %v\n", err)
		return
	}

	// Handle different possible response structures
	var logs []map[string]interface{}

	// Case 1: Response body is directly an array of logs
	if logArray, ok := responseData.([]interface{}); ok {
		logs = make([]map[string]interface{}, len(logArray))
		for i, logItem := range logArray {
			if logMap, ok := logItem.(map[string]interface{}); ok {
				logs[i] = logMap
			}
		}
	} else if responseMap, ok := responseData.(map[string]interface{}); ok {
		// Case 2: Response body contains logs field
		if logsData, exists := responseMap["logs"]; exists {
			if logArray, ok := logsData.([]interface{}); ok {
				logs = make([]map[string]interface{}, len(logArray))
				for i, logItem := range logArray {
					if logMap, ok := logItem.(map[string]interface{}); ok {
						logs[i] = logMap
					}
				}
			}
		} else {
			// Case 3: Response body itself might be a single log entry
			logs = []map[string]interface{}{responseMap}
		}
	}

	if len(logs) == 0 {
		fmt.Println("No log entries found")
		// Print debug information
		fmt.Printf("Response structure: %s\n", string(bodyBytes))
		return
	}

	fmt.Printf("\n=== Log Contents (Total: %d entries) ===\n", len(logs))

	for i, logData := range logs {
		// Extract content field
		if content, exists := logData["content"]; exists {
			// Extract timestamp for better readability
			timestamp := ""
			if timeStr, ok := logData["_time_"]; ok {
				timestamp = fmt.Sprintf("[%v] ", timeStr)
			} else if timeField, ok := logData["__time__"]; ok {
				timestamp = fmt.Sprintf("[%v] ", timeField)
			}

			fmt.Printf("%s%s: %v\n", timestamp, fmt.Sprintf("Log[%d]", i+1), content)
		} else {
			fmt.Printf("Log[%d]: No 'content' field found\n", i+1)
			// Print available fields for debugging
			fmt.Printf("  Available fields: ")
			for key := range logData {
				fmt.Printf("%s ", key)
			}
			fmt.Printf("\n")
		}
	}

	fmt.Printf("=== End of Log Contents ===\n")
}

// PrintRawResponse prints the complete raw response (for debugging)
func (c *AliyunLogClient) PrintRawResponse(response *sls20201230.GetLogsResponse) {
	if response == nil || response.Body == nil {
		fmt.Println("No log data returned")
		return
	}

	fmt.Printf("=== Raw Response ===\n")
	fmt.Printf("Response: %+v\n", response.Body)
	fmt.Printf("=== End Raw Response ===\n")
}

// getEnvOrDefault retrieves an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// parseArguments parses command line arguments and returns configuration
func parseArguments(args []string) (from, to int32, outputFormat string, showRaw bool, err error) {
	// Default values
	now := time.Now().Unix()
	oneHourAgo := now - 3600
	from = int32(oneHourAgo)
	to = int32(now)
	outputFormat = "content"
	showRaw = false

	// Parse arguments
	i := 0
	for i < len(args) {
		switch args[i] {
		case "--format", "-f":
			if i+1 < len(args) {
				outputFormat = args[i+1]
				i += 2
			} else {
				return 0, 0, "", false, fmt.Errorf("--format requires a value (content|raw|both)")
			}
		case "--raw", "-r":
			showRaw = true
			i++
		case "--help", "-h":
			printUsage()
			os.Exit(0)
		default:
			// Try to parse as time range
			if i < len(args) && i+1 < len(args) {
				if fromInt, err := strconv.ParseInt(args[i], 10, 32); err == nil {
					from = int32(fromInt)
				}
				if toInt, err := strconv.ParseInt(args[i+1], 10, 32); err == nil {
					to = int32(toInt)
				}
				i += 2
			} else {
				i++
			}
		}
	}

	if from >= to {
		return 0, 0, "", false, fmt.Errorf("invalid time range: from (%d) must be less than to (%d)", from, to)
	}

	return from, to, outputFormat, showRaw, nil
}

// printUsage prints usage information
func printUsage() {
	fmt.Println("Aliyun SLS Log Query Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  go run aliyungetlog.go [options] [from_timestamp] [to_timestamp]")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -f, --format <format>   Output format: content (default), raw, both")
	fmt.Println("  -r, --raw              Show raw response (same as --format raw)")
	fmt.Println("  -h, --help             Show this help message")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run aliyungetlog.go                           # Query last hour")
	fmt.Println("  go run aliyungetlog.go 1691552442 1754710856     # Query specific time range")
	fmt.Println("  go run aliyungetlog.go --format content          # Show only content fields")
	fmt.Println("  go run aliyungetlog.go --format raw              # Show raw response")
	fmt.Println("  go run aliyungetlog.go --format both             # Show both content and raw")
	fmt.Println("")
	fmt.Println("Environment Variables:")
	fmt.Println("  ALIYUN_ACCESS_KEY_ID      - Required: Aliyun Access Key ID")
	fmt.Println("  ALIYUN_ACCESS_KEY_SECRET  - Required: Aliyun Access Key Secret")
	fmt.Println("  ALIYUN_SLS_ENDPOINT       - Optional: SLS endpoint (default: ap-southeast-1.log.aliyuncs.com)")
	fmt.Println("  ALIYUN_SLS_PROJECT        - Optional: SLS project (default: userlogsystem)")
	fmt.Println("  ALIYUN_SLS_LOGSTORE       - Optional: SLS logstore (default: userlog)")
	fmt.Println("  ALIYUN_LOG_QUERY          - Optional: Custom log query")
}

// main function with improved structure and error handling
func main() {
	// Parse command line arguments
	args := os.Args[1:]

	// Parse arguments
	from, to, outputFormat, showRaw, err := parseArguments(args)
	if err != nil {
		log.Fatalf("Failed to parse arguments: %v", err)
	}

	// Create log client
	client, err := NewAliyunLogClient()
	if err != nil {
		log.Fatalf("Failed to create Aliyun log client: %v", err)
	}

	// Default query - can be customized via environment variable
	queryString := getEnvOrDefault(
		"ALIYUN_LOG_QUERY",
		"* and __tag__:_image_name_: \"user-app:51b7a751e2a754051bde7d4d718b19f8\"",
	)

	// Create log query
	query := &LogQuery{
		From:  from,
		To:    to,
		Query: queryString,
	}

	fmt.Printf("Executing log query...\n")
	fmt.Printf("Time range: %d to %d\n", from, to)
	fmt.Printf("Query: %s\n", queryString)
	fmt.Printf("Output format: %s\n", outputFormat)

	// Execute query
	response, err := client.QueryLogs(query)
	if err != nil {
		log.Fatalf("Failed to query logs: %v", err)
	}

	// Print results based on format
	switch outputFormat {
	case "content":
		client.PrintLogResults(response)
	case "raw":
		client.PrintRawResponse(response)
	case "both":
		client.PrintLogResults(response)
		fmt.Println("\n" + strings.Repeat("=", 80))
		client.PrintRawResponse(response)
	default:
		fmt.Printf("Unknown output format: %s. Using 'content' format.\n", outputFormat)
		client.PrintLogResults(response)
	}

	// Handle legacy --raw flag
	if showRaw && outputFormat == "content" {
		fmt.Println("\n" + strings.Repeat("=", 80))
		client.PrintRawResponse(response)
	}

	fmt.Println("Log query completed successfully")
}
