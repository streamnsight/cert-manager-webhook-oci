ARG ORACLE_LINUX_VERSION=10
ARG GOLANG_VERSION=1.25
ARG GOLANG_PACKAGE=go-toolset

FROM oraclelinux:${ORACLE_LINUX_VERSION} AS oracle-linux-go-dev

ARG GOLANG_VERSION
ARG GOLANG_PACKAGE
RUN dnf -y update \
    && GOLANG_NEVRA="$(dnf repoquery --latest-limit 1 --qf '%{name}-%{version}-%{release}.%{arch}' "${GOLANG_PACKAGE}" | grep "^${GOLANG_PACKAGE}-${GOLANG_VERSION}\\.")" \
    && test -n "${GOLANG_NEVRA}" \
    && dnf -y install "git" "${GOLANG_NEVRA}" \
    && dnf clean all \
    && rm -rf /var/cache/dnf

FROM oracle-linux-go-dev AS build_deps

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

FROM build_deps AS build

WORKDIR /workspace
COPY . .
ARG TARGETOS=linux
ARG TARGETARCH
RUN --mount=type=cache,target=/workspace/.cache/go-build \
    --mount=type=cache,target=/workspace/.cache/go-tmp \
    mkdir -p /workspace/.cache/go-build /workspace/.cache/go-tmp \
    && GOCACHE=/workspace/.cache/go-build GOTMPDIR=/workspace/.cache/go-tmp GOFLAGS=-p=1 CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o webhook .

FROM oraclelinux:${ORACLE_LINUX_VERSION}

COPY --from=build /workspace/webhook /usr/local/bin/webhook
ENTRYPOINT ["webhook"]
