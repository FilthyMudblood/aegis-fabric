.PHONY: build docker run clean test demo-up demo-down demo-logs demo-snapshots demo-report sdk-proto sdk-test kind-quickstart demo-agent-docker proto

APP_NAME = afp-sidecar
DEMO_AGENT_IMAGE = ghcr.io/filthymudblood/afp-demo-agent:latest
VERSION = latest

# --- paths (grouped cmd layout) ---
CMD_DP = ./cmd/dataplane
CMD_CP = ./cmd/controlplane
CMD_DEMO = ./cmd/demo

proto:
	buf generate

build: proto
	go build -o bin/sidecar $(CMD_DP)/sidecar
	go build -o bin/egressclient $(CMD_DP)/egressclient
	go build -o bin/testclient $(CMD_DP)/testclient
	go build -o bin/preflightclient $(CMD_DP)/preflightclient
	go build -o bin/operator $(CMD_CP)/operator
	go build -o bin/policy-controller $(CMD_CP)/policy-controller
	go build -o bin/policyctl $(CMD_CP)/policyctl

docker:
	docker build -t local/$(APP_NAME):$(VERSION) .

demo-agent-docker:
	docker build -f Dockerfile.demo-agent -t $(DEMO_AGENT_IMAGE) .

operator-docker:
	docker build -f Dockerfile.operator -t ghcr.io/filthymudblood/aegis-fabric-operator:latest .

run: build
	./bin/sidecar

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
