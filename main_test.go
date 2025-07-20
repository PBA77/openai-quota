package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// Test setup
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Endpoint health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Grupa v1 (bez prefiksu /api)
	v1 := r.Group("/v1")
	{
		v1.POST("/chat/completions", chatCompletionsProxy)
		v1.GET("/chat/completions", info)
	}

	// Grupa api/v1 (z prefiksem /api)
	apiV1 := r.Group("/api/v1")
	{
		apiV1.POST("/chat/completions", chatCompletionsProxy)
		apiV1.GET("/chat/completions", info)
	}

	// Endpoint cennika
	r.GET("/pricing", pricing)
	r.GET("/api/pricing", pricing)

	return r
}

func resetGlobalState() {
	totalCost = 0.0
	costLimitUSD = 2.0
	modelPricing = make(map[string]ModelPricing)

	// Dodaj domyślne ceny testowe
	modelPricing["gpt-4o"] = ModelPricing{
		Model:  "gpt-4o",
		Input:  2.5,  // $2.5 za 1M tokenów
		Output: 10.0, // $10 za 1M tokenów
	}
	modelPricing["gpt-4o-mini"] = ModelPricing{
		Model:  "gpt-4o-mini",
		Input:  0.15, // $0.15 za 1M tokenów
		Output: 0.6,  // $0.6 za 1M tokenów
	}
	modelPricing["gpt-4.1"] = ModelPricing{
		Model:  "gpt-4.1",
		Input:  2.0,
		Output: 8.0,
	}
	modelPricing["o3"] = ModelPricing{
		Model:  "o3",
		Input:  2.0,
		Output: 8.0,
	}
	modelPricing["o4-mini"] = ModelPricing{
		Model:  "o4-mini",
		Input:  1.1,
		Output: 4.4,
	}
	modelPricing["gpt-3.5-turbo"] = ModelPricing{
		Model:  "gpt-3.5-turbo",
		Input:  0.5,
		Output: 1.5,
	}
	
	// Generate allowed prefixes from test models
	generateAllowedPrefixes()
}

func createTestCSV(filename string, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

// Test CSV loading
func TestLoadModelPricing(t *testing.T) {
	tests := []struct {
		name        string
		csvContent  string
		expectError bool
		expectedLen int
	}{
		{
			name: "Valid CSV",
			csvContent: `model,version,input,cached_input,output
gpt-4o,gpt-4o-2024-08-06,2.5,1.25,10.0
gpt-4o-mini,gpt-4o-mini-2024-07-18,0.15,0.075,0.6`,
			expectError: false,
			expectedLen: 4, // 2 models + 2 versions = 4 entries
		},
		{
			name: "CSV with empty cached_input",
			csvContent: `model,version,input,cached_input,output
gpt-4o,gpt-4o-2024-08-06,2.5,,10.0`,
			expectError: false,
			expectedLen: 2, // 1 model + 1 version = 2 entries
		},
		{
			name: "Invalid input price",
			csvContent: `model,version,input,cached_input,output
gpt-4o,gpt-4o-2024-08-06,invalid,1.25,10.0`,
			expectError: false,
			expectedLen: 0,
		},
		{
			name: "Incomplete row",
			csvContent: `model,version,input,cached_input,output
gpt-4o,gpt-4o-2024-08-06`,
			expectError: true, // Should expect error for malformed CSV
			expectedLen: 0,
		},
		{
			name:        "Empty file",
			csvContent:  "",
			expectError: true,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := fmt.Sprintf("test_%s.csv", strings.ReplaceAll(tt.name, " ", "_"))
			defer os.Remove(filename)

			err := createTestCSV(filename, tt.csvContent)
			if err != nil {
				t.Fatalf("Failed to create test CSV: %v", err)
			}

			// Reset pricing before test
			modelPricing = make(map[string]ModelPricing)

			err = loadModelPricing(filename)
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if len(modelPricing) != tt.expectedLen {
				t.Errorf("Expected %d models, got %d", tt.expectedLen, len(modelPricing))
			}
		})
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"2.5", 2.5, false},
		{"0", 0.0, false},
		{"", 0.0, false},
		{"invalid", 0.0, true},
		{"10.123456", 10.123456, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseFloat(tt.input)
			if tt.hasError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.hasError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestGetPricingForModel(t *testing.T) {
	resetGlobalState()

	tests := []struct {
		model         string
		expectedFound bool
		expectedModel string
	}{
		{"gpt-4o", true, "gpt-4o"},
		{"gpt-4o-2024-08-06", true, "gpt-4o"}, // prefix match
		{"gpt-4o-mini", true, "gpt-4o-mini"},
		{"unknown-model", false, "default"},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			pricing, found := getPricingForModel(tt.model)
			if found != tt.expectedFound {
				t.Errorf("Expected found=%v, got %v", tt.expectedFound, found)
			}
			if found && pricing.Model != tt.expectedModel {
				t.Errorf("Expected model=%s, got %s", tt.expectedModel, pricing.Model)
			}
		})
	}
}

func TestCalculateCost(t *testing.T) {
	resetGlobalState()

	tests := []struct {
		promptTokens     int
		completionTokens int
		model            string
		expectedCost     float64
	}{
		{1000, 500, "gpt-4o", 0.0075},       // (1000 * 2.5 + 500 * 10.0) / 1000000
		{500, 200, "gpt-4o-mini", 0.000195}, // (500 * 0.15 + 200 * 0.6) / 1000000
		{0, 0, "gpt-4o", 0.0},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%d_%d", tt.model, tt.promptTokens, tt.completionTokens), func(t *testing.T) {
			cost := calculateCost(tt.promptTokens, tt.completionTokens, tt.model)
			// Use approximate comparison due to floating point precision
			if cost < tt.expectedCost-0.000001 || cost > tt.expectedCost+0.000001 {
				t.Errorf("Expected cost %f, got %f", tt.expectedCost, cost)
			}
		})
	}
}

func TestIsModelAllowed(t *testing.T) {
	tests := []struct {
		model    string
		expected bool
	}{
		{"gpt-4o", true},
		{"gpt-4o-mini", true},
		{"gpt-4.1", true},
		{"o3", true},
		{"o4", true},
		{"gpt-3.5", true},
		{"claude-3", false},
		{"llama-2", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			result := isModelAllowed(tt.model)
			if result != tt.expected {
				t.Errorf("Expected %v for model %s, got %v", tt.expected, tt.model, result)
			}
		})
	}
}

func TestCalculateTokensFromMessages(t *testing.T) {
	messages := []ChatMessage{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there!"},
	}

	tokens := calculateTokensFromMessages(messages, "gpt-4o")
	if tokens <= 0 {
		t.Error("Expected positive token count")
	}

	// Test empty messages
	emptyMessages := []ChatMessage{}
	emptyTokens := calculateTokensFromMessages(emptyMessages, "gpt-4o")
	if emptyTokens != 3 { // base tokens
		t.Errorf("Expected 3 tokens for empty messages, got %d", emptyTokens)
	}
}

func TestHealthEndpoint(t *testing.T) {
	router := setupTestRouter()
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response["status"])
	}
}

func TestInfoEndpoint(t *testing.T) {
	resetGlobalState()
	router := setupTestRouter()

	endpoints := []string{"/v1/chat/completions", "/api/v1/chat/completions"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", endpoint, nil)
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to parse JSON response: %v", err)
			}

			if response["cost_limit"] != 2.0 {
				t.Errorf("Expected cost_limit 2.0, got %v", response["cost_limit"])
			}
		})
	}
}

func TestPricingEndpoint(t *testing.T) {
	resetGlobalState()
	router := setupTestRouter()

	endpoints := []string{"/pricing", "/api/pricing"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", endpoint, nil)
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to parse JSON response: %v", err)
			}

			if response["pricing"] == nil {
				t.Error("Expected pricing data in response")
			}
		})
	}
}

func TestChatCompletionsProxy_MissingAuth(t *testing.T) {
	resetGlobalState()
	router := setupTestRouter()

	reqBody := ChatRequest{
		Model: "gpt-4o",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	endpoints := []string{"/v1/chat/completions", "/api/v1/chat/completions"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected status 401, got %d", w.Code)
			}

			var response ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Errorf("Failed to parse JSON response: %v", err)
			}

			if !strings.Contains(response.Error, "Authorization") {
				t.Errorf("Expected authorization error, got: %s", response.Error)
			}
		})
	}
}

func TestChatCompletionsProxy_InvalidAuth(t *testing.T) {
	resetGlobalState()
	router := setupTestRouter()

	reqBody := ChatRequest{
		Model: "gpt-4o",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	tests := []struct {
		name       string
		authHeader string
	}{
		{"Invalid format", "InvalidFormat"},
		{"Empty Bearer", "Bearer "},
		{"No Bearer prefix", "sk-test-key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", tt.authHeader)
			router.ServeHTTP(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Expected status 401, got %d", w.Code)
			}
		})
	}
}

func TestChatCompletionsProxy_InvalidJSON(t *testing.T) {
	resetGlobalState()
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-test-key")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestChatCompletionsProxy_DisallowedModel(t *testing.T) {
	resetGlobalState()
	router := setupTestRouter()

	reqBody := ChatRequest{
		Model: "claude-3",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-test-key")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if !strings.Contains(response.Error, "not in the allowed list") {
		t.Errorf("Expected model restriction error, got: %s", response.Error)
	}
}

func TestChatCompletionsProxy_QuotaExceeded(t *testing.T) {
	resetGlobalState()
	costLimitUSD = 0.0 // Set quota to 0
	router := setupTestRouter()

	reqBody := ChatRequest{
		Model: "gpt-4o",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-test-key")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if !strings.Contains(response.Error, "cost limit") {
		t.Errorf("Expected cost limit error, got: %s", response.Error)
	}
}

func TestChatCompletionsProxy_PromptCostExceedsQuota(t *testing.T) {
	resetGlobalState()
	costLimitUSD = 0.000001 // Very low quota
	router := setupTestRouter()

	// Long message that will have high token count
	longMessage := strings.Repeat("This is a very long message that will consume many tokens. ", 50)

	reqBody := ChatRequest{
		Model: "gpt-4o",
		Messages: []ChatMessage{
			{Role: "user", Content: longMessage},
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-test-key")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if !strings.Contains(response.Error, "exceed") {
		t.Errorf("Expected exceed error, got: %s", response.Error)
	}
}

// Test helper functions
func TestGetAvailableModels(t *testing.T) {
	resetGlobalState()

	models := getAvailableModels()
	if len(models) != len(modelPricing) {
		t.Errorf("Expected %d models, got %d", len(modelPricing), len(models))
	}

	// Check if all models from pricing are included
	for expectedModel := range modelPricing {
		found := false
		for _, model := range models {
			if model == expectedModel {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Model %s not found in available models", expectedModel)
		}
	}
}

func TestCountTokens(t *testing.T) {
	// Test basic token counting
	tokens := countTokens("Hello world", "gpt-4o")
	if tokens <= 0 {
		t.Error("Expected positive token count")
	}

	// Test empty string
	emptyTokens := countTokens("", "gpt-4o")
	if emptyTokens != 0 {
		t.Errorf("Expected 0 tokens for empty string, got %d", emptyTokens)
	}

	// Test longer text has more tokens
	shortTokens := countTokens("Hi", "gpt-4o")
	longTokens := countTokens("This is a much longer text that should have more tokens", "gpt-4o")
	if longTokens <= shortTokens {
		t.Error("Expected longer text to have more tokens")
	}
}

// Integration tests
func TestFullWorkflow_ValidRequest(t *testing.T) {
	resetGlobalState()

	// Mock OpenAI API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("Authorization") != "Bearer sk-test-key" {
			t.Errorf("Expected Bearer token in request to OpenAI")
		}

		// Return mock response
		response := ChatResponse{
			ID:      "test-id",
			Object:  "chat.completion",
			Created: 1234567890,
			Model:   "gpt-4o",
			Choices: []Choice{
				{
					Message: ChatMessage{
						Role:    "assistant",
						Content: "Hello! How can I help you?",
					},
					FinishReason: "stop",
					Index:        0,
				},
			},
			Usage: Usage{
				PromptTokens:     5,
				CompletionTokens: 8,
				TotalTokens:      13,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	// Note: This test would need to mock the actual OpenAI API call
	// For now, we'll test the validation logic

	router := setupTestRouter()

	reqBody := ChatRequest{
		Model: "gpt-4o",
		Messages: []ChatMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-test-key")
	router.ServeHTTP(w, req)

	// Should fail with OpenAI API error (since we're not mocking the actual call)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 (due to mock API), got %d", w.Code)
	}
}

// Benchmark tests
func BenchmarkCalculateTokensFromMessages(b *testing.B) {
	messages := []ChatMessage{
		{Role: "user", Content: "Hello, how are you today?"},
		{Role: "assistant", Content: "I'm doing well, thank you for asking!"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateTokensFromMessages(messages, "gpt-4o")
	}
}

func BenchmarkCalculateCost(b *testing.B) {
	resetGlobalState()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateCost(1000, 500, "gpt-4o")
	}
}

func BenchmarkGetPricingForModel(b *testing.B) {
	resetGlobalState()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getPricingForModel("gpt-4o")
	}
}

// Edge case tests
func TestChatRequest_EmptyMessages(t *testing.T) {
	resetGlobalState()
	router := setupTestRouter()

	reqBody := ChatRequest{
		Model:    "gpt-4o",
		Messages: []ChatMessage{},
	}

	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-test-key")
	router.ServeHTTP(w, req)

	// Should pass validation but fail at OpenAI API
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestChatRequest_VeryLongContent(t *testing.T) {
	resetGlobalState()
	costLimitUSD = 1.0 // Set reasonable limit
	router := setupTestRouter()

	// Create very long content
	veryLongContent := strings.Repeat("A", 10000)

	reqBody := ChatRequest{
		Model: "gpt-4o",
		Messages: []ChatMessage{
			{Role: "user", Content: veryLongContent},
		},
	}

	jsonData, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/v1/chat/completions", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-test-key")
	router.ServeHTTP(w, req)

	// Should likely be blocked due to high token cost
	if w.Code != http.StatusTooManyRequests && w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 429 or 500, got %d", w.Code)
	}
}

func TestModelPricing_DeepCopy(t *testing.T) {
	original := ModelPricing{
		Model:       "test-model",
		Version:     "v1",
		Input:       1.0,
		CachedInput: 0.5,
		Output:      2.0,
	}

	// Test that struct values are copied correctly
	copy := original
	copy.Input = 999.0

	if original.Input == copy.Input {
		t.Error("Struct should be copied by value, not reference")
	}
}

func TestConcurrentAccess(t *testing.T) {
	resetGlobalState()

	// Test concurrent access to global variables
	// This is a basic test - in production you'd want more sophisticated concurrency testing
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			calculateCost(100, 50, "gpt-4o")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			getPricingForModel("gpt-4o")
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}

// Test error handling
func TestErrorHandling_InvalidModelInPricing(t *testing.T) {
	// Test what happens when we try to get pricing for a model that doesn't exist
	pricing, found := getPricingForModel("nonexistent-model-12345")

	if found {
		t.Error("Should not find pricing for nonexistent model")
	}

	if pricing.Model != "default" {
		t.Errorf("Expected default pricing, got model: %s", pricing.Model)
	}
}

func TestValidateStructFields(t *testing.T) {
	// Test that our structs have the expected fields
	chatReq := ChatRequest{}
	chatResp := ChatResponse{}

	// Use reflection to check struct fields exist
	reqType := reflect.TypeOf(chatReq)
	respType := reflect.TypeOf(chatResp)

	// Check some key fields exist
	requiredReqFields := []string{"Model", "Messages", "Temperature"}
	for _, field := range requiredReqFields {
		if _, found := reqType.FieldByName(field); !found {
			t.Errorf("ChatRequest missing required field: %s", field)
		}
	}

	requiredRespFields := []string{"ID", "Object", "Choices", "Usage"}
	for _, field := range requiredRespFields {
		if _, found := respType.FieldByName(field); !found {
			t.Errorf("ChatResponse missing required field: %s", field)
		}
	}
}
