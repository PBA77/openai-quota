package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/pkoukk/tiktoken-go"
)

var (
	allowedModelPrefixes = []string{"gpt-4o", "gpt-4-1106-preview", "gpt-4.1", "o3", "o4", "gpt-3.5"}
	costLimitUSD         float64
	modelPricing         = make(map[string]ModelPricing)
	totalCost            = 0.0
	mu                   sync.Mutex
)

type ModelPricing struct {
	Model       string  `json:"model"`
	Version     string  `json:"version"`
	Input       float64 `json:"input"`        // cena za 1M tokenów input
	CachedInput float64 `json:"cached_input"` // cena za 1M tokenów cached input
	Output      float64 `json:"output"`       // cena za 1M tokenów output
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type ChatRequest struct {
	Model            string        `json:"model"`
	Messages         []ChatMessage `json:"messages"`
	Temperature      *float64      `json:"temperature,omitempty"`
	MaxTokens        *int          `json:"max_tokens,omitempty"`
	N                *int          `json:"n,omitempty"`
	Stop             interface{}   `json:"stop,omitempty"`
	PresencePenalty  *float64      `json:"presence_penalty,omitempty"`
	FrequencyPenalty *float64      `json:"frequency_penalty,omitempty"`
	Functions        interface{}   `json:"functions,omitempty"`
	FunctionCall     interface{}   `json:"function_call,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Choice struct {
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
	Index        int         `json:"index"`
}

type ChatResponse struct {
	ID         string      `json:"id"`
	Object     string      `json:"object"`
	Created    int64       `json:"created"`
	Model      string      `json:"model"`
	Choices    []Choice    `json:"choices"`
	Usage      Usage       `json:"usage"`
	ProxyUsage *ProxyUsage `json:"proxy_usage,omitempty"`
}

type ProxyUsage struct {
	PromptTokens     int     `json:"prompt_tokens"`
	CompletionTokens int     `json:"completion_tokens"`
	CostUSD          float64 `json:"cost_usd"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func loadModelPricing(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("cannot open pricing file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("error reading CSV: %w", err)
	}

	if len(records) < 2 {
		return fmt.Errorf("CSV file must contain at least header and one data row")
	}

	// Pomijamy nagłówek (pierwszy wiersz)
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) < 5 {
			log.Printf("Skipping incomplete row: %v", record)
			continue
		}

		model := record[0]
		version := record[1]

		input, err := parseFloat(record[2])
		if err != nil {
			log.Printf("Invalid input price for model %s: %v", model, err)
			continue
		}

		cachedInput, _ := parseFloat(record[3]) // może być puste

		output, err := parseFloat(record[4])
		if err != nil {
			log.Printf("Invalid output price for model %s: %v", model, err)
			continue
		}

		pricing := ModelPricing{
			Model:       model,
			Version:     version,
			Input:       input,
			CachedInput: cachedInput,
			Output:      output,
		}

		modelPricing[model] = pricing
		// Również dodaj pod pełną nazwą wersji, jeśli się różni
		if version != "" && version != model {
			modelPricing[version] = pricing
		}
	}

	log.Printf("Loaded pricing for %d models", len(modelPricing))
	return nil
}

func parseFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

func getPricingForModel(model string) (ModelPricing, bool) {
	// Sprawdź bezpośrednie dopasowanie
	if pricing, exists := modelPricing[model]; exists {
		return pricing, true
	}

	// Sprawdź prefiksy (np. gpt-4o-2024-08-06 -> gpt-4o)
	for modelKey, pricing := range modelPricing {
		if strings.HasPrefix(model, modelKey) {
			return pricing, true
		}
	}

	// Fallback - domyślne ceny GPT-4
	return ModelPricing{
		Model:  "default",
		Input:  30.0, // $30 za 1M tokenów
		Output: 60.0, // $60 za 1M tokenów
	}, false
}

func calculateCost(promptTokens, completionTokens int, model string) float64 {
	pricing, found := getPricingForModel(model)
	if !found {
		log.Printf("Pricing not found for model %s, using defaults", model)
	}

	// Ceny w CSV są za 1M tokenów, więc dzielimy przez 1,000,000
	costPrompt := float64(promptTokens) * (pricing.Input / 1000000.0)
	costCompletion := float64(completionTokens) * (pricing.Output / 1000000.0)

	return costPrompt + costCompletion
}

func getAvailableModels() []string {
	models := make([]string, 0, len(modelPricing))
	for model := range modelPricing {
		models = append(models, model)
	}
	return models
}

func isModelAllowed(model string) bool {
	for _, prefix := range allowedModelPrefixes {
		if strings.HasPrefix(model, prefix) {
			return true
		}
	}
	return false
}

func countTokens(text, model string) int {
	enc, err := tiktoken.EncodingForModel(model)
	if err != nil {
		enc, _ = tiktoken.GetEncoding("cl100k_base")
	}

	tokens := enc.Encode(text, nil, nil)
	return len(tokens)
}

func calculateTokensFromMessages(messages []ChatMessage, model string) int {
	totalTokens := 0
	for _, msg := range messages {
		text := msg.Role + msg.Name + msg.Content
		totalTokens += countTokens(text, model)
	}
	return totalTokens + 3*len(messages) + 3
}

func callOpenAI(reqData ChatRequest, apiKey string) (*ChatResponse, error) {
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, err
	}

	return &chatResp, nil
}

func chatCompletionsProxy(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	if totalCost >= costLimitUSD {
		log.Printf("Request blocked: quota limit exceeded, current_cost=$%.6f, limit=$%.6f", totalCost, costLimitUSD)
		c.JSON(http.StatusTooManyRequests, ErrorResponse{
			Error: "Global cost limit exceeded.",
		})
		return
	}

	// Pobierz klucz API z nagłówka Authorization
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Missing Authorization header. Use: Authorization: Bearer your-api-key",
		})
		return
	}

	// Sprawdź format nagłówka Authorization
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Invalid Authorization header format. Use: Authorization: Bearer your-api-key",
		})
		return
	}

	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Empty API key. Use: Authorization: Bearer your-api-key",
		})
		return
	}

	var reqData ChatRequest
	if err := c.ShouldBindJSON(&reqData); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Missing JSON data in request.",
		})
		return
	}

	if !isModelAllowed(reqData.Model) {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: fmt.Sprintf("Model %s is not in the allowed list.", reqData.Model),
		})
		return
	}

	// Oblicz tokeny promptu przed wywołaniem API
	promptTokens := calculateTokensFromMessages(reqData.Messages, reqData.Model)

	// Sprawdź czy sam prompt nie przekroczy limitu kosztów
	promptCost := calculateCost(promptTokens, 0, reqData.Model)
	if totalCost+promptCost >= costLimitUSD {
		log.Printf("Request blocked: prompt would exceed quota, prompt_tokens=%d, prompt_cost=$%.6f, current_cost=$%.6f, limit=$%.6f",
			promptTokens, promptCost, totalCost, costLimitUSD)
		c.JSON(http.StatusTooManyRequests, ErrorResponse{
			Error: "Request would exceed global cost limit.",
		})
		return
	}

	response, err := callOpenAI(reqData, apiKey)
	if err != nil {
		// Nawet jeśli request do OpenAI się nie powiódł, policz tokeny dla logowania
		costTotalRequest := calculateCost(promptTokens, 0, reqData.Model) // brak completion tokenów

		log.Printf("Failed request: model=%s, prompt_tokens=%d, completion_tokens=0, estimated_cost=$%.6f, total_cost=$%.6f, remaining=$%.6f, error=%v",
			reqData.Model, promptTokens, costTotalRequest, totalCost, costLimitUSD-totalCost, err)

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: fmt.Sprintf("OpenAI API call error: %s", err.Error()),
		})
		return
	}

	// Aktualizuj tokeny z odpowiedzi API (mogą być dokładniejsze)
	if response.Usage.PromptTokens > 0 {
		promptTokens = response.Usage.PromptTokens
	}
	completionTokens := response.Usage.CompletionTokens

	if promptTokens == 0 || completionTokens == 0 {
		promptTokens = calculateTokensFromMessages(reqData.Messages, reqData.Model)

		completionText := ""
		for _, choice := range response.Choices {
			completionText += choice.Message.Content
		}
		completionTokens = countTokens(completionText, reqData.Model)
	}

	costTotalRequest := calculateCost(promptTokens, completionTokens, reqData.Model)

	totalCost += costTotalRequest

	// Logowanie szczegółowych informacji o zużyciu
	log.Printf("Request: model=%s, prompt_tokens=%d, completion_tokens=%d, cost=$%.6f, total_cost=$%.6f, remaining=$%.6f",
		reqData.Model, promptTokens, completionTokens, costTotalRequest, totalCost, costLimitUSD-totalCost)

	response.ProxyUsage = &ProxyUsage{
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		CostUSD:          float64(int(costTotalRequest*1000000)) / 1000000, // round to 6 decimal places
	}

	c.JSON(http.StatusOK, response)
}

func info(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"info":             "Local OpenAI proxy. Available method: POST.",
		"cost_limit":       costLimitUSD,
		"current_cost":     totalCost,
		"remaining":        costLimitUSD - totalCost,
		"available_models": getAvailableModels(),
		"models_count":     len(modelPricing),
	})
}

func pricing(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"pricing": modelPricing,
	})
}

func main() {
	// Parsowanie argumentów wiersza poleceń
	var (
		quota       = flag.Float64("quota", 2.0, "Global cost limit in USD")
		port        = flag.String("port", "5000", "Port to run server on")
		pricingFile = flag.String("pricing", "config/model_pricing.csv", "Path to CSV file with model pricing")
		help        = flag.Bool("help", false, "Show help")
		h           = flag.Bool("h", false, "Show help (short)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "OpenAI Quota Proxy - proxy server with cost control\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nAuthorization:\n")
		fmt.Fprintf(os.Stderr, "  OpenAI API key must be passed in Authorization header of each request:\n")
		fmt.Fprintf(os.Stderr, "  Authorization: Bearer your-openai-api-key\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -quota 5.0 -port 8080\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -quota 10.0 -pricing custom_pricing.csv\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nExample request:\n")
		fmt.Fprintf(os.Stderr, "  curl -X POST http://localhost:5000/v1/chat/completions \\\n")
		fmt.Fprintf(os.Stderr, "       -H \"Authorization: Bearer your-api-key\" \\\n")
		fmt.Fprintf(os.Stderr, "       -H \"Content-Type: application/json\" \\\n")
		fmt.Fprintf(os.Stderr, "       -d '{\"model\":\"gpt-4o\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello\"}]}'\n")
	}

	flag.Parse()

	if *help || *h {
		flag.Usage()
		os.Exit(0)
	}

	// Wczytaj cennik modeli
	if err := loadModelPricing(*pricingFile); err != nil {
		log.Printf("Warning: Cannot load pricing file (%s): %v", *pricingFile, err)
		log.Printf("Using default pricing for models")
	}

	// Ustawienie globalnych zmiennych
	costLimitUSD = *quota

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

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

	log.Printf("Starting server on port %s with quota limit: $%.2f", *port, costLimitUSD)
	log.Printf("Loaded pricing for models: %v", getAvailableModels())

	if err := r.Run("127.0.0.1:" + *port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
