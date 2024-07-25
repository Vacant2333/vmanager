OUTPUT_DIR=_output
IMAGE_PREFIX=vacantsh
TAG=1.0

webhook-manager:
	CC=gcc CGO_ENABLED=0 go build -o ${OUTPUT_DIR}/webhook-manager ./cmd/webhook-manager

images:
	docker buildx build -t "${IMAGE_PREFIX}/webhook-manager:$(TAG)" . -f ./dockerfile/webhook-manager/Dockerfile --output=type=docker
