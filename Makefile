
clean:
	rm -f redzilla

build:
	CGO_ENABLED=0 go build -a -ldflags '-s' -o redzilla

docker/build:
	docker build . -t raptorbox/redzilla:latest

docker/push: docker/build
	docker push raptorbox/redzilla:latest
