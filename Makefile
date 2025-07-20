BINARY_NAME=openai-quota
PORT=5000
QUOTA=2.0
PRICING_FILE=config/model_pricing.csv

.PHONY: build run clean test deps help

# Budowanie aplikacji
build:
	go build -o $(BINARY_NAME) .

# Uruchomienie z automatyczną kompilacją (domyślne parametry)
run:
	go run main.go

# Uruchomienie z custom parametrami
run-quota:
	go run main.go -quota $(QUOTA) -port $(PORT) -pricing $(PRICING_FILE)

# Uruchomienie skompilowanej wersji
run-binary: build
	./$(BINARY_NAME)

# Uruchomienie skompilowanej wersji z parametrami
run-binary-quota: build
	./$(BINARY_NAME) -quota $(QUOTA) -port $(PORT) -pricing $(PRICING_FILE)

# Pokazanie pomocy aplikacji
app-help: build
	./$(BINARY_NAME) -help

# Test endpointów
test-endpoints: build
	@echo "Testing info endpoint..."
	@curl -s http://localhost:$(PORT)/v1/chat/completions | head -c 200
	@echo "\n\nTesting pricing endpoint..."
	@curl -s http://localhost:$(PORT)/pricing | head -c 200
	@echo "\n"

# Instalacja zależności
deps:
	go mod tidy
	go mod download

# Czyszczenie
clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -f test_*.csv

# Testowanie
test:
	go test ./...

# Pełny pakiet testów
test-full:
	./scripts/run_tests.sh

# Szybkie testy
test-quick:
	./scripts/quick_test.sh

# Testy z szczegółowym wyjściem
test-verbose:
	go test -v ./...

# Testy z pokryciem kodu
test-coverage:
	go test -cover ./...

# Raport pokrycia w HTML
test-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Raport pokrycia wygenerowany: coverage.html"

# Testy wydajnościowe
test-bench:
	go test -bench=. -benchmem ./...

# Liczba testów
test-count:
	@echo "Liczba testów:"
	@go test -v 2>&1 | grep "=== RUN" | wc -l

# Pokrycie per funkcja
test-coverage-func:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

# Uruchamianie konkretnego testu
test-run:
	@read -p "Podaj wzorzec testu: " pattern; \
	go test -run "$$pattern" -v ./...

# Pełny raport testowy
test-full-report:
	./scripts/run_tests.sh

# Kompilacja dla różnych platform
build-all:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows-amd64.exe .

# Formatowanie kodu
fmt:
	go fmt ./...

# Sprawdzanie kodu
vet:
	go vet ./...

# Pomoc dla Makefile
help:
	@echo "Dostępne komendy:"
	@echo "  build                - Kompiluje aplikację"
	@echo "  run                  - Uruchamia aplikację (quota=2.0, port=5000)"
	@echo "  run-quota            - Uruchamia z custom parametrami"
	@echo "  run-binary           - Uruchamia skompilowaną wersję"
	@echo "  run-binary-quota     - Uruchamia skompilowaną wersję z parametrami"
	@echo "  app-help             - Pokazuje pomoc aplikacji"
	@echo "  test-endpoints       - Testuje endpointy API"
	@echo "  clean                - Usuwa pliki binarne i testy"
	@echo "  deps                 - Instaluje zależności"
	@echo ""
	@echo "Testowanie:"
	@echo "  test                 - Uruchamia wszystkie testy"
	@echo "  test-verbose         - Testy z szczegółowym wyjściem"
	@echo "  test-coverage        - Testy z pokryciem kodu"
	@echo "  test-html            - Generuje raport HTML pokrycia"
	@echo "  test-bench           - Testy wydajnościowe"
	@echo "  test-count           - Liczba testów"
	@echo "  test-coverage-func   - Pokrycie per funkcja"
	@echo "  test-run             - Uruchamia konkretny test"
	@echo "  test-full-report     - Pełny raport testowy (./run_tests.sh)"
	@echo "  test-quick           - Szybki test (./quick_test.sh)"
	@echo ""
	@echo "Jakość kodu:"
	@echo "  fmt                  - Formatuje kod"
	@echo "  vet                  - Sprawdza kod"
	@echo ""
	@echo "Parametry:"
	@echo "  QUOTA              - Limit kosztów (domyślnie: $(QUOTA))"
	@echo "  PORT               - Port serwera (domyślnie: $(PORT))"
	@echo "  PRICING_FILE       - Plik cennika (domyślnie: $(PRICING_FILE))"
	@echo ""
	@echo "Przykłady:"
	@echo "  make run-quota QUOTA=5.0 PORT=8080"
	@echo "  make run-binary-quota QUOTA=10.0 PRICING_FILE=custom.csv"
