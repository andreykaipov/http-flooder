#!/bin/sh

cleanup() {
	echo "Ensuring clean environment"
	rm -f report.json
	docker rm -f flooder || true
	docker rm -f api || true
	docker network rm wow || true
}

main() {
	echo "Creating test network"
	docker network create wow

	echo "Starting dummy API"
	docker run --detach --network=wow --name=api dummy-api \
		-failure-rate 0.27 \
		-delay-interval 0,50

	echo "Starting flooder"
	docker run --network=wow --name=flooder flooder \
		-endpoint http://api:8080/time \
		-duration 3 \
		-requests-per-second 100 \
		-report /report.json 2>/dev/null

	echo "Fetching report"
	docker cp flooder:/report.json .
	jq . report.json
}

echo "Running integration test"
trap cleanup EXIT INT TERM
cleanup 2>/dev/null
main
