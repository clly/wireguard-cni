

help: ## displays this help message
	@perl -n -e'(/(^[a-zA-Z_\/-]+):.*##(.*)/ && printf "\033[34m%-12s\033[0m %s\n", $$1, $$2) || /(^[a-zA-Z_\/-]+):/ && printf "\033[34m%-12s\033[0m %s\n", $$1' $(MAKEFILE_LIST) | \
		sort | \
		grep -v '#' 
