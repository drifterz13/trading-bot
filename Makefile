.PHONY: build-image
build-image:
	docker build --tag mixz-robo .

.PHONY: run-dev
run-dev:
	export APP_ENV=dev && \
	rm -rf ./data/dev.db && touch ./data/dev.db && go run .
