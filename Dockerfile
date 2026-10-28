# syntax=docker/dockerfile:1
# Ubuntu 20.04 / 通用 Linux：静态链接单文件，无运行时依赖
# 构建请开启 BuildKit（Docker 23+ 默认开启）：DOCKER_BUILDKIT=1 docker build ...
# 模块代理默认 goproxy.cn（面向国内构建机）；海外构建请覆盖，例如：
#   --build-arg GOPROXY=https://proxy.golang.org,direct
FROM golang:1.22-bookworm AS build
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY internal ./internal
COPY cmd ./cmd
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/igozhang ./cmd/igozhang

# scratch：静态二进制无需运行时镜像（避免依赖 gcr.io 等外网 registry）
# 拷入 CA 包，保证 /api/fetch 等 HTTPS 出站校验证书可用
# 默认 root：便于写入 hostPath 挂载的 /igozhang.cn（若需 nonroot 请同步调整宿主机目录权限）
FROM scratch
WORKDIR /
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /out/igozhang /igozhang
EXPOSE 3003
ENV LISTEN=:3003
ENV DATA_ROOT=/igozhang.cn
ENV TRANSFER_SUBDIR=fs
ENTRYPOINT ["/igozhang"]
