.PHONY: build install uninstall test clean

BINARY := confluence-search
VERSION := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILD_TIME := $(VERSION)
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.commit=$(COMMIT)

INSTALL_DIR := $(shell \
	if [ -d "$$HOME/.local/bin" ] && echo "$$PATH" | grep -q "$$HOME/.local/bin"; then \
		echo "$$HOME/.local/bin"; \
	elif [ -w /usr/local/bin ]; then \
		echo "/usr/local/bin"; \
	else \
		echo "$$HOME/.local/bin"; \
	fi)

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/confluence-search/

install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "Installed $(INSTALL_DIR)/$(BINARY)"

uninstall:
	rm -f $(INSTALL_DIR)/$(BINARY)
	@echo "Removed $(INSTALL_DIR)/$(BINARY)"

test:
	go test ./... -v

clean:
	rm -f $(BINARY)
