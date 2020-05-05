default:
	@echo "Please specify a task:"
	@awk -F: '/^[^\.\t].+$$/ {print "-",$$1}' Makefile | tail -n+2

images: dummy-api-image flooder-image

.PHONY: dummy-api
dummy-api:
	go build -o ./bin/$@ ./$@

dummy-api-image:
	docker build -t dummy-api dummy-api/

.PHONY: flooder
flooder:
	go build -o ./bin/$@ ./$@

flooder-image:
	docker build -t flooder flooder/

clean:
	rm -rf bin
