default:
	@echo "Please specify a task:"
	@awk -F: '/^[^.].+?:$$/ {print "-",$$1}' Makefile | tail -n+2

images: api-image flooder-image

.PHONY: api
api:
	go build -o ./bin/$@ ./$@

api-image:
	docker build -t dummy-api api/

.PHONY: flooder
flooder:
	go build -o ./bin/$@ ./$@

flooder-image:
	docker build -t flooder flooder/

clean:
	rm -rf bin
