.DEFAULT_GOAL = build

VERSION ?= `git describe --abbrev=0 --tags $(git rev-list --tags --max-count=1)`
VERSION_FLAG := -X `go list ./cmd`.Version=$(VERSION)

FISH_DIR = ~/.config/fish/completions
KS_BINARY = /usr/local/bin/kubeswitch

ZSH_EXISTS := $(shell test -f ~/.zshrc && grep kubeswitch.zsh ~/.zshrc)
BASH_EXISTS := $(shell test -f ~/.bashrc && grep kubeswitch.bash ~/.bashrc)

build:
	@echo -n Building kubeswitch...
	@go build \
		-ldflags "-w -s $(VERSION_FLAG)" \
		-o bin/kubeswitch .
	@echo done

install:
	@echo -n Installing kubeswitch...
	@cp -f bin/kubeswitch $(KS_BINARY)
	@chown root:root $(KS_BINARY)
	@chmod 755 $(KS_BINARY)
	@echo done

clean:
	@rm -rf bin/

install-config:
	@cp -f kubeswitch.yaml $(HOME)/.kubeswitch.yaml

bash-completion:
	@echo -n Installing Bash completion...
ifeq ($(BASH_EXISTS), )
	@cp -f completion/kubeswitch $(HOME)/.kubeswitch.bash
	@echo "source $$HOME/.kubeswitch.bash" >> $(HOME)/.bashrc
endif
	@echo done

zsh-completion:
	@echo -n Installing ZSH completion...
ifeq ($(ZSH_EXISTS), )
	@cp -f completion/kubeswitch $(HOME)/.kubeswitch.zsh
	@echo "source $$HOME/.kubeswitch.zsh" >> $(HOME)/.zshrc
endif
	@echo done

fish-completion: $(FISH_DIR)
	@echo -n Installing Fish completion...
	@cp -f completion/kubeswitch.fish $(FISH_DIR)
	@echo done

$(FISH_DIR):
	@mkdir -p $(FISH_DIR)

