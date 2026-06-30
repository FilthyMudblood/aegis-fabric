# ==========================================
# Stage 1: Build the binaries
# ==========================================
FROM golang:1.25-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git tzdata

# 设置工作目录
WORKDIR /app

# 预取依赖 (利用 Docker 缓存加速构建)
COPY go.mod go.sum ./
RUN go mod download

# 拷贝全量源码
COPY . .

# 静态编译三大核心组件 (禁用 CGO 以确保可移植性)
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-w -s" -o /bin/sidecar ./cmd/sidecar
RUN go build -ldflags="-w -s" -o /bin/http_gateway ./cmd/http_gateway
RUN go build -ldflags="-w -s" -o /bin/simulator ./cmd/simulator
RUN go build -ldflags="-w -s" -o /bin/preflightclient ./cmd/preflightclient
RUN go build -ldflags="-w -s" -o /bin/policy-controller ./cmd/policy-controller
RUN go build -ldflags="-w -s" -o /bin/policyctl ./cmd/policyctl

# ==========================================
# Stage 2: Create the minimal production image
# ==========================================
FROM alpine:3.19

# 安装基础网络排障工具和证书
RUN apk add --no-cache ca-certificates curl net-tools tzdata

WORKDIR /app

# 从 builder 阶段拷贝编译好的二进制文件
COPY --from=builder /bin/sidecar /usr/local/bin/sidecar
COPY --from=builder /bin/http_gateway /usr/local/bin/http_gateway
COPY --from=builder /bin/simulator /usr/local/bin/simulator
COPY --from=builder /bin/preflightclient /usr/local/bin/preflightclient
COPY --from=builder /bin/policy-controller /usr/local/bin/policy-controller
COPY --from=builder /bin/policyctl /usr/local/bin/policyctl

# 声明 HTTP 网关的默认端口
EXPOSE 8082

# 默认启动 HTTP Gateway
CMD ["http_gateway"]
