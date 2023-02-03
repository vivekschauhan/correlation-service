.PHONY: all build
PROJECT_NAME := correlation-service
WORKSPACE ?= $$(pwd)
GO_PKG_LIST := $(shell go list ./...)
export GOFLAGS := -mod=mod

all: clean package
	@echo "Done"

clean:
	@rm -f ./correlation-service
	@echo "Clean complete"

dep:
	@echo "Resolving go package dependencies"
	@go mod tidy
	@echo "Package dependencies completed"

${WORKSPACE}/$(PROJECT_NAME):
	@CGO_ENABLED=0 GOARCH=amd64 go build -o ./$(PROJECT_NAME) main.go

build: dep ${WORKSPACE}/$(PROJECT_NAME)
	@echo "Build complete"

docker:
	docker build -t $(PROJECT_NAME):latest -f ${WORKSPACE}/Dockerfile .
	@echo "Docker build complete"
