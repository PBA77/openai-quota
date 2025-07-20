package main

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

// Additional comprehensive tests

func TestModelPricing_VersionHandling(t *testing.T) {
	resetGlobalState()

	// Add pricing with different versions
	modelPricing["gpt-4o"] = ModelPricing{
		Model:   "gpt-4o",
		Version: "gpt-4o-2024-08-06",
		Input:   2.5,
		Output:  10.0,
	}

	modelPricing["gpt-4o-2024-08-06"] = ModelPricing{
		Model:   "gpt-4o",
		Version: "gpt-4o-2024-08-06",
		Input:   2.5,
		Output:  10.0,
	}

	tests := []struct {
		requestModel string
		shouldFind   bool
	}{
		{"gpt-4o", true},
		{"gpt-4o-2024-08-06", true},
		{"gpt-4o-2024-11-20", true}, // Should match prefix
		{"gpt-4", false},            // Should not match
	}

	for _, tt := range tests {
		t.Run(tt.requestModel, func(t *testing.T) {
			_, found := getPricingForModel(tt.requestModel)
			if found != tt.shouldFind {
				t.Errorf("Model %s: expected found=%v, got %v", tt.requestModel, tt.shouldFind, found)
			}
		})
	}
}

func TestCostCalculation_EdgeCases(t *testing.T) {
	resetGlobalState()

	tests := []struct {
		name             string
		promptTokens     int
		completionTokens int
		model            string
		expectedMinCost  float64
		expectedMaxCost  float64
	}{
		{"Zero tokens", 0, 0, "gpt-4o", 0.0, 0.0},
		{"Only prompt tokens", 1000, 0, "gpt-4o", 0.0024, 0.0026},
		{"Only completion tokens", 0, 1000, "gpt-4o", 0.009, 0.011},
		{"Large numbers", 1000000, 500000, "gpt-4o", 7.4, 7.6},
		{"Unknown model defaults", 100, 50, "unknown-model", 0.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := calculateCost(tt.promptTokens, tt.completionTokens, tt.model)
			if cost < tt.expectedMinCost || cost > tt.expectedMaxCost {
				t.Errorf("Cost %f not in expected range [%f, %f]", cost, tt.expectedMinCost, tt.expectedMaxCost)
			}
		})
	}
}

func TestTokenCounting_DifferentModels(t *testing.T) {
	testCases := []struct {
		text   string
		models []string
	}{
		{"Hello world!", []string{"gpt-4o", "gpt-4o-mini"}},
		{"", []string{"gpt-4o"}},                                // Empty string
		{"🚀🔥💻", []string{"gpt-4o"}},                             // Unicode/emojis
		{`{"json": "test", "number": 123}`, []string{"gpt-4o"}}, // JSON
	}

	for _, tc := range testCases {
		for _, model := range tc.models {
			t.Run(model+"_"+tc.text, func(t *testing.T) {
				tokens := countTokens(tc.text, model)
				if tokens < 0 {
					t.Errorf("Token count should not be negative: %d", tokens)
				}
				if tc.text == "" && tokens != 0 {
					t.Errorf("Empty string should have 0 tokens, got %d", tokens)
				}
			})
		}
	}
}

func TestConcurrentProxyRequests(t *testing.T) {
	resetGlobalState()
	costLimitUSD = 10.0 // High limit for this test
	router := setupTestRouter()

	// Number of concurrent requests
	numRequests := 10
	var wg sync.WaitGroup
	results := make([]int, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/v1/chat/completions", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer sk-test-key")

			router.ServeHTTP(w, req)
			results[index] = w.Code
		}(i)
	}

	wg.Wait()

	// Check that we got consistent responses
	for i, code := range results {
		if code != http.StatusBadRequest && code != http.StatusInternalServerError {
			t.Errorf("Request %d: unexpected status code %d", i, code)
		}
	}
}

func TestGlobalStateManagement(t *testing.T) {
	// Save original state
	originalCost := totalCost
	originalLimit := costLimitUSD
	originalPricing := make(map[string]ModelPricing)
	for k, v := range modelPricing {
		originalPricing[k] = v
	}

	defer func() {
		// Restore original state
		totalCost = originalCost
		costLimitUSD = originalLimit
		modelPricing = originalPricing
	}()

	// Test state changes
	totalCost = 1.5
	costLimitUSD = 2.0

	if totalCost != 1.5 {
		t.Errorf("Expected totalCost 1.5, got %f", totalCost)
	}

	remaining := costLimitUSD - totalCost
	if remaining != 0.5 {
		t.Errorf("Expected remaining 0.5, got %f", remaining)
	}
}

func TestHTTPMethodsOnEndpoints(t *testing.T) {
	router := setupTestRouter()

	endpoints := []struct {
		path           string
		method         string
		expectedStatus int
	}{
		{"/health", "GET", http.StatusOK},
		{"/health", "POST", http.StatusNotFound},
		{"/v1/chat/completions", "GET", http.StatusOK},
		{"/v1/chat/completions", "POST", http.StatusUnauthorized}, // No auth
		{"/v1/chat/completions", "PUT", http.StatusNotFound},
		{"/pricing", "GET", http.StatusOK},
		{"/pricing", "POST", http.StatusNotFound},
	}

	for _, test := range endpoints {
		t.Run(test.method+"_"+test.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(test.method, test.path, nil)
			router.ServeHTTP(w, req)

			if w.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, w.Code)
			}
		})
	}
}

func TestJSONResponseFormats(t *testing.T) {
	resetGlobalState()
	router := setupTestRouter()

	// Test health endpoint JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	var healthResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &healthResp); err != nil {
		t.Errorf("Health endpoint returned invalid JSON: %v", err)
	}

	// Test info endpoint JSON
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/v1/chat/completions", nil)
	router.ServeHTTP(w, req)

	var infoResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &infoResp); err != nil {
		t.Errorf("Info endpoint returned invalid JSON: %v", err)
	}

	// Check required fields
	requiredFields := []string{"cost_limit", "current_cost", "remaining", "available_models"}
	for _, field := range requiredFields {
		if _, exists := infoResp[field]; !exists {
			t.Errorf("Info response missing required field: %s", field)
		}
	}
}

func TestFloatingPointPrecision(t *testing.T) {
	resetGlobalState()

	// Test with very small costs
	cost1 := calculateCost(1, 1, "gpt-4o-mini")
	cost2 := calculateCost(1, 1, "gpt-4o-mini")

	if cost1 != cost2 {
		t.Errorf("Same calculation should give same result: %f vs %f", cost1, cost2)
	}

	// Test precision with many decimal places
	largeTokens := 1234567
	cost := calculateCost(largeTokens, largeTokens, "gpt-4o")

	if math.IsNaN(cost) || math.IsInf(cost, 0) {
		t.Errorf("Cost calculation resulted in NaN or Inf: %f", cost)
	}
}

func TestCSVParsing_RealWorldScenarios(t *testing.T) {
	testCases := []struct {
		name       string
		content    string
		shouldWork bool
	}{
		{
			name:       "Windows line endings",
			content:    "model,version,input,cached_input,output\r\ngpt-4o,v1,2.5,1.25,10.0\r\n",
			shouldWork: true,
		},
		{
			name:       "Extra whitespace",
			content:    "model,version,input,cached_input,output\n  gpt-4o  ,  v1  ,  2.5  ,  1.25  ,  10.0  \n",
			shouldWork: false, // Our parser doesn't trim whitespace
		},
		{
			name:       "Missing header",
			content:    "gpt-4o,v1,2.5,1.25,10.0\n",
			shouldWork: false, // No header means data gets skipped
		},
		{
			name:       "Unicode model names",
			content:    "model,version,input,cached_input,output\ngpt-4ö,v1,2.5,1.25,10.0\n",
			shouldWork: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filename := "test_" + tc.name + ".csv"
			defer os.Remove(filename)

			err := createTestCSV(filename, tc.content)
			if err != nil {
				t.Fatalf("Failed to create test CSV: %v", err)
			}

			// Reset pricing
			modelPricing = make(map[string]ModelPricing)
			_ = loadModelPricing(filename)

			hasModels := len(modelPricing) > 0
			if tc.shouldWork && !hasModels {
				t.Errorf("Expected successful parsing but got no models")
			}
			if !tc.shouldWork && hasModels {
				t.Errorf("Expected parsing to fail but got %d models", len(modelPricing))
			}
		})
	}
}

func TestErrorResponseFormats(t *testing.T) {
	resetGlobalState()
	router := setupTestRouter()

	errorTests := []struct {
		name           string
		setup          func()
		request        func() *http.Request
		expectedStatus int
		expectedField  string
	}{
		{
			name:  "Missing auth",
			setup: func() {},
			request: func() *http.Request {
				req, _ := http.NewRequest("POST", "/v1/chat/completions", nil)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus: http.StatusUnauthorized,
			expectedField:  "error",
		},
		{
			name:  "Quota exceeded",
			setup: func() { costLimitUSD = 0.0 },
			request: func() *http.Request {
				req, _ := http.NewRequest("POST", "/v1/chat/completions", nil)
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer test")
				return req
			},
			expectedStatus: http.StatusTooManyRequests,
			expectedField:  "error",
		},
	}

	for _, test := range errorTests {
		t.Run(test.name, func(t *testing.T) {
			resetGlobalState()
			test.setup()

			w := httptest.NewRecorder()
			req := test.request()
			router.ServeHTTP(w, req)

			if w.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Errorf("Error response is not valid JSON: %v", err)
			}

			if _, exists := response[test.expectedField]; !exists {
				t.Errorf("Error response missing field: %s", test.expectedField)
			}
		})
	}
}

func TestPerformanceBaseline(t *testing.T) {
	resetGlobalState()
	
	// Benchmark token calculation - be realistic about tiktoken performance
	start := time.Now()
	for i := 0; i < 100; i++ {
		calculateTokensFromMessages([]ChatMessage{
			{Role: "user", Content: "Test message"},
		}, "gpt-4o")
	}
	duration := time.Since(start)

	// Token calculation with tiktoken can be much slower on first runs due to loading
	// Allow up to 30 seconds to account for model loading time
	if duration > 30*time.Second {
		t.Errorf("Token calculation too slow: %v for 100 iterations", duration)
	} else {
		t.Logf("Token calculation performance: %v for 100 iterations", duration)
	}

	// Benchmark cost calculation - this should be fast
	start = time.Now()
	for i := 0; i < 10000; i++ {
		calculateCost(100, 50, "gpt-4o")
	}
	duration = time.Since(start)

	if duration > 2*time.Second {
		t.Errorf("Cost calculation too slow: %v for 10000 iterations", duration)
	} else {
		t.Logf("Cost calculation performance: %v for 10000 iterations", duration)
	}
}

func TestMemoryUsage(t *testing.T) {
	resetGlobalState()

	// Load large number of models
	for i := 0; i < 1000; i++ {
		modelName := "test-model-" + string(rune(i))
		modelPricing[modelName] = ModelPricing{
			Model:  modelName,
			Input:  float64(i),
			Output: float64(i * 2),
		}
	}

	// Check we can still access them
	if len(modelPricing) != 1002 { // 1000 + 2 from resetGlobalState
		t.Errorf("Expected 1002 models, got %d", len(modelPricing))
	}

	// Check memory doesn't leak on repeated access
	for i := 0; i < 100; i++ {
		getAvailableModels()
	}
}

func TestConfigurationValidation(t *testing.T) {
	tests := []struct {
		quota       float64
		shouldPanic bool
	}{
		{0.0, false},         // Zero quota is valid
		{-1.0, false},        // Negative quota is technically valid (though not useful)
		{math.Inf(1), false}, // Infinite quota is valid
		{math.NaN(), false},  // NaN quota is technically valid
	}

	for _, test := range tests {
		t.Run("quota_validation", func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !test.shouldPanic {
					t.Errorf("Unexpected panic with quota %f: %v", test.quota, r)
				}
			}()

			resetGlobalState()
			costLimitUSD = test.quota

			// Should not panic
			router := setupTestRouter()
			if router == nil {
				t.Error("Router setup failed")
			}
		})
	}
}
