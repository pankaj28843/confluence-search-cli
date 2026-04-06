.PHONY: build install uninstall test clean

BINARY := confluence-search

INSTALL_DIR := $(shell \
	if [ -d "$$HOME/.local/bin" ] && echo "$$PATH" | grep -q "$$HOME/.local/bin"; then \
		echo "$$HOME/.local/bin"; \
	elif [ -w /usr/local/bin ]; then \
		echo "/usr/local/bin"; \
	else \
		echo "$$HOME/.local/bin"; \
	fi)

build:
	go build -o $(BINARY) ./cmd/confluence-search/

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
