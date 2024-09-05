.DEFAULT: natsify
.PHONY: help deploy

targets := natsify

sys := systemctl --user

$(targets): ## Build a target server binary
	go build -o bin/$@ ./cmd/$@

deploy: $(targets) ## Deploy all binaries
	mv bin/* $(HOME)/.local/bin
	cp -r systemd/* $(HOME)/.config/systemd/user
	$(sys) stop $(targets)
	$(sys) disable $(targets)
	$(sys) daemon-reload
	$(sys) enable $(targets)
	$(sys) start $(targets)	

help: ## Print help
	@printf "\033[36m%-30s\033[0m %s\n" "(target)" "Build a target binary in current arch for running locally: $(targets)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
