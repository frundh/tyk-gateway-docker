ifeq (, $(shell which docker))
    $(error "Docker is not installed!")
endif

.PHONY: help dotnet-test go-test

help:
	@echo "Available targets:"
	@echo "  dotnet-test       Run dotnet tests"
	@echo "  go-test           Run go tests"
	@echo "  help              Show this help message"

dotnet-test:
	@docker run --rm -v /var/run/docker.sock.raw:/var/run/docker.sock -v $(PWD):/src -w /src/tests/dotnet --add-host=host.docker.internal:host-gateway -e TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal mcr.microsoft.com/dotnet/sdk:6.0 dotnet test

go-test:
	@docker run --rm -v /var/run/docker.sock.raw:/var/run/docker.sock -v $(PWD):/src -w /src/tests/go --add-host=host.docker.internal:host-gateway -e TC_HOST=host.docker.internal golang:1.21-bullseye go test -v
