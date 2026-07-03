# ==========================================
# Stage 1: Build the binaries
# ==========================================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-w -s" -o /bin/sidecar ./cmd/dataplane/sidecar
RUN go build -ldflags="-w -s" -o /bin/http_gateway ./cmd/demo/http_gateway
RUN go build -ldflags="-w -s" -o /bin/simulator ./cmd/demo/simulator
RUN go build -ldflags="-w -s" -o /bin/preflightclient ./cmd/dataplane/preflightclient
RUN go build -ldflags="-w -s" -o /bin/policy-controller ./cmd/controlplane/policy-controller
RUN go build -ldflags="-w -s" -o /bin/policyctl ./cmd/controlplane/policyctl

# ==========================================
# Stage 2: Create the minimal production image
# ==========================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates curl net-tools tzdata

WORKDIR /app

COPY --from=builder /bin/sidecar /usr/local/bin/sidecar
COPY --from=builder /bin/http_gateway /usr/local/bin/http_gateway
COPY --from=builder /bin/simulator /usr/local/bin/simulator
COPY --from=builder /bin/preflightclient /usr/local/bin/preflightclient
COPY --from=builder /bin/policy-controller /usr/local/bin/policy-controller
COPY --from=builder /bin/policyctl /usr/local/bin/policyctl

EXPOSE 8082

CMD ["http_gateway"]
