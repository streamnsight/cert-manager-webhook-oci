ARG ORACLE_LINUX_VERSION=7.9
FROM oraclelinux:${ORACLE_LINUX_VERSION} AS oracle-linux-go-dev

ARG GOLANG_VERSION="1.19-1.0.1.el7"
RUN yum update -y \
    && yum install -y oracle-golang-release-el7 git \
    && yum install -y golang-${GOLANG_VERSION}

########################################################
FROM oracle-linux-go-dev AS build_deps

WORKDIR /workspace
ENV GO111MODULE=on
COPY go.mod .
COPY go.sum .
RUN go mod download

########################################################
FROM build_deps AS build

WORKDIR /workspace
ENV GO111MODULE=on
COPY . .
RUN CGO_ENABLED=0 go build -o webhook -ldflags '-w -extldflags "-static"' .

########################################################
FROM oraclelinux:${ORACLE_LINUX_VERSION}

COPY --from=build /workspace/webhook /usr/local/bin/webhook
ENTRYPOINT ["webhook"]
