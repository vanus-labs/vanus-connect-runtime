VANUS_ROOT=$(shell pwd)
VSPROTO_ROOT=$(VANUS_ROOT)/proto
GIT_COMMIT=$(shell git log -1 --format='%h' | awk '{print $0}')
DATE=$(shell date +%Y-%m-%d_%H:%M:%S%z)
GO_VERSION=$(shell go version)

export VANUS_LOG_LEVEL=info

DOCKER_REGISTRY ?= public.ecr.aws
DOCKER_REPO ?= ${DOCKER_REGISTRY}/vanus
IMAGE_TAG ?= ${GIT_COMMIT}
#os linux or darwin
GOOS ?= linux
#arch amd64 or arm64
GOARCH ?= amd64

VERSION ?= ${IMAGE_TAG}

GO_BUILD= GO111MODULE=on CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath
DOCKER_BUILD_ARG= --build-arg TARGETARCH=$(GOARCH) --build-arg TARGETOS=$(GOOS)
DOCKER_PLATFORM ?= linux/amd64,linux/arm64

clean :
	rm -rf bin

docker-push:
	docker buildx build --platform ${DOCKER_PLATFORM} -t ${DOCKER_REPO}/connect-runtime:${IMAGE_TAG} -f build/images/timer/Dockerfile . --push
docker-build:
	docker build -t ${DOCKER_REPO}/timer:${IMAGE_TAG} $(DOCKER_BUILD_ARG) -f Dockerfile .
build:
	$(GO_BUILD)  -o bin/runtime cmd/main.go

# install swagger binary file
.PHONY: install-swagger
install-swagger:
	@go install github.com/go-swagger/go-swagger/cmd/swagger@v0.27.0

# vaild swagger
.PHONY: valid-swagger
valid-swagger:
	swagger validate ./api/swagger.yaml

# gen swagger 
.PHONY: gen-swagger
gen-swagger:
	swagger validate ./api/swagger.yaml
	rm -rf ./api/restapi
	rm -rf ./api/models
	swagger generate server --target ./api --spec ./api/swagger.yaml --exclude-main  --name vanus-connect-runtime