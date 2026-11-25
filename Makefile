.PHONY: init
init:
	go run ./cmd/repo-scripts init

.PHONY: pre-commit
pre-commit:
	go run ./cmd/repo-scripts pre-commit

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  init        Initialize git hooks"
	@echo "  pre-commit  Run pre-commit hook"
	@echo "  pre-push    Run pre-push hook"
	@echo "  help        Show this help message"