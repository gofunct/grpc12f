.PHONY: build
build: ## bash ./build.sh
	bash ./build.sh

.PHONY: test
test: ## bash ./all/test.sh
	bash ./all/test.sh

.PHONY: push
push: build test ## bash ./push.sh
	bash ./push.sh

.PHONY: push-latest
push-latest: ## bash ./push.sh true
	bash ./push.sh true

.PHONY: tag-latest
tag-latest: ## bash ./build.sh true
	bash ./build.sh true

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort