.PHONY: build docker run clean test demo-up demo-down demo-logs demo-snapshots demo-report sdk-proto sdk-test kind-quickstart demo-agent-docker proto

APP_NAME = afp-sidecar
DEMO_AGENT_IMAGE = ghcr.io/filthymudblood/afp-demo-agent:latest
VERSION = latest

proto:
	buf generate

build: proto
	go build -o bin/$(APP_NAME) ./cmd/sidecar/main.go
	go build -o bin/egressclient ./cmd/egressclient/main.go
	go build -o bin/testclient ./cmd/testclient/main.go
	go build -o bin/preflightclient ./cmd/preflightclient/main.go
	go build -o bin/operator ./cmd/operator/main.go
	go build -o bin/policy-controller ./cmd/policy-controller/main.go
	go build -o bin/policyctl ./cmd/policyctl/main.go

docker:
	docker build -t local/$(APP_NAME):$(VERSION) .

demo-agent-docker:
	docker build -f Dockerfile.demo-agent -t $(DEMO_AGENT_IMAGE) .

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

kind-quickstart:
	./scripts/kind-quickstart.sh
