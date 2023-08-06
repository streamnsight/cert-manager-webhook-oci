ARG ORACLE_LINUX_VERSION=7.9
FROM golang:1.20.7-alpine AS build_deps
ARG TARGETARCH

RUN apk add --no-cache git


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
RUN CGO_ENABLED=0 GOARCH=$TARGETARCH go build -o webhook -ldflags '-w -extldflags "-static"' .

########################################################
FROM oraclelinux:${ORACLE_LINUX_VERSION}

COPY --from=build /workspace/webhook /usr/local/bin/webhook
ENTRYPOINT ["webhook"]
