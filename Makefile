# --- CONFIGURATION ---
BINARY_NAME := ikelos
PKG := .
GO := go

# --- COULEURS ---
RESET := \033[0m
BOLD := \033[1m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
CYAN := \033[36m
RED := \033[31m

.PHONY: all build run test vet clean deps install help

all: check-deps vet test build
	@echo "$(GREEN)✨ System Ready.$(RESET)"
	@echo "Run $(BOLD)make run$(RESET) to access the console."

check-deps:
	@echo "$(YELLOW)🔍 Scanning dependencies...$(RESET)"
	@$(GO) mod tidy
	@echo "$(GREEN)✔ Dependencies locked.$(RESET)"

test:
	@echo "$(CYAN)🧪 Running test suite...$(RESET)"
	@$(GO) test ./... || (echo "$(RED)❌ Tests Failed$(RESET)"; exit 1)
	@echo "$(GREEN)✔ All tests passed.$(RESET)"

vet:
	@echo "$(CYAN)🔬 Vetting...$(RESET)"
	@$(GO) vet ./...

build:
	@echo "$(CYAN)🔨 Compiling THE SURGEON...$(RESET)"
	@$(GO) build -ldflags="-s -w" -o $(BINARY_NAME) $(PKG) || (echo "$(RED)❌ Compilation Failed$(RESET)"; exit 1)
	@echo "$(GREEN)✔ Binary generated: ./$(BINARY_NAME)$(RESET)"

run: build
	@./$(BINARY_NAME)

clean:
	@echo "$(YELLOW)🧹 Purging workspace...$(RESET)"
	@rm -f $(BINARY_NAME)
	@rm -rf cloned_site
	@echo "$(GREEN)✔ System clean.$(RESET)"

install: build
	@echo "$(BLUE)📦 Installing to /usr/local/bin...$(RESET)"
	@sudo mv $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✅ IKELOS is now global. Type 'ikelos' anywhere.$(RESET)"

help:
	@echo "$(BOLD)IKELOS COMMAND CENTER$(RESET)"
	@echo "  $(CYAN)make run$(RESET)      : Start Interactive Mode"
	@echo "  $(CYAN)make install$(RESET)  : Install globally"
	@echo "  $(CYAN)ikelos -url ...$(RESET): Run direct attack"