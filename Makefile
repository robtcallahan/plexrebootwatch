BINARY_NAME := plexrebootwatch
INSTALL_BIN := /usr/local/bin/$(BINARY_NAME)
CONFIG_FILE := plexrebootwatch.json
INSTALL_CONFIG := /usr/local/etc/$(CONFIG_FILE)
PLIST_FILE := com.rob.plexrebootwatch.plist
INSTALL_PLIST := $(HOME)/Library/LaunchAgents/$(PLIST_FILE)

.PHONY: all build install load clean

all: build install

build:
	go build -o $(BINARY_NAME) .

install: build
	@echo "Installing binary to $(INSTALL_BIN)"
	install -m 755 $(BINARY_NAME) $(INSTALL_BIN)

	@echo "Installing config to $(INSTALL_CONFIG)"
	install -m 644 $(CONFIG_FILE) $(INSTALL_CONFIG)

	@echo "Installing plist to $(INSTALL_PLIST)"
	install -m 644 $(PLIST_FILE) $(INSTALL_PLIST)

load:
	@echo "Loading LaunchAgent..."
	launchctl unload $(INSTALL_PLIST) 2>/dev/null || true
	launchctl load $(INSTALL_PLIST)

clean:
	rm -f $(BINARY_NAME)
