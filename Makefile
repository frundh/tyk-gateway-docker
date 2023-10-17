ifeq (, $(shell which docker))
    $(error "Docker is not installed!")
endif

.PHONY: help test

help:
	@echo "Available targets:"
	@echo "  test: Run tests"
	@echo "  help: Show this help message"

test:
	@docker run --rm -v /var/run/docker.sock.raw:/var/run/docker.sock -v $(PWD):/src -w /src/tests --add-host=host.docker.internal:host-gateway -e TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal mcr.microsoft.com/dotnet/sdk:6.0 dotnet test
