.DEFAULT: err-pub
.PHONY: fmt test gen clean run help sql

targets := err-pub

sys := systemctl --user

VERSION ?= v0.0.0

$(targets): ## Build a target server binary
	go build -o bin/$@ ./cmd/$@

deploy: $(targets) ## Deploy all binaries
	mv bin/* $(HOME)/.local/bin
	cp -r systemd/* $(HOME)/.config/systemd/user
	$(sys) stop $(services)
	$(sys) disable $(services)
	$(sys) daemon-reload
	$(sys) enable $(services)
	$(sys) start $(services)	

help: ## Print help
	@printf "\033[36m%-30s\033[0m %s\n" "(target)" "Build a target binary in current arch for running locally: $(targets)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
