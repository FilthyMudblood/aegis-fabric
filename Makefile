.PHONY: build docker run clean test demo-up demo-down demo-logs demo-snapshots demo-report sdk-proto sdk-test

APP_NAME = afp-sidecar
VERSION = latest

build:
	go build -o bin/$(APP_NAME) ./cmd/sidecar/main.go
	go build -o bin/egressclient ./cmd/egressclient/main.go
	go build -o bin/testclient ./cmd/testclient/main.go
	go build -o bin/preflightclient ./cmd/preflightclient/main.go
	go build -o bin/operator ./cmd/operator/main.go

docker:
	docker build -t local/$(APP_NAME):$(VERSION) .

run: build
	./bin/$(APP_NAME)

test:
	go test ./... -v

clean:
	rm -rf bin/

demo-up:
	docker compose up --build -d

demo-down:
	docker compose down --remove-orphans

demo-logs:
	docker compose logs -f --tail=200

demo-snapshots:
	./scripts/export_demo_snapshots.sh

demo-report:
	DEMO_AUTO_UP=1 WAIT_TIMEOUT_SECONDS=180 ./scripts/generate_demo_report.sh

sdk-proto:
	./sdk/python/scripts/gen_proto.sh

sdk-test: sdk-proto
	cd sdk/python && PYTHONPATH=. pytest tests -v
