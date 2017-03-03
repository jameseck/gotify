DOCKER_CI_IMAGE = gotify-ci
build:
	go build -v
fmt:
	go fmt ./...
test:
	go test -v ./...

test-docker: build .build-docker
	docker run --rm -it -v $(CURDIR):/go $(DOCKER_CI_IMAGE)

.build-docker:
	docker build -t $(DOCKER_CI_IMAGE) .
	@docker inspect -f '{{.Id}}' $(DOCKER_CI_IMAGE) > .build-docker
