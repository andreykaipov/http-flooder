default:
	@echo "Please specify a task" 
	@awk -F: '/^[^.].+?:$$/ {print "-",$$1}' Makefile | tail -n+2

.PHONY: api
api:
	go build -o ./bin/$@ ./$@

.PHONY: flooder
flooder:
	go build -o ./bin/$@ ./$@

clean:
	rm -rf bin
