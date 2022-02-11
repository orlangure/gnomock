FROM golang:latest AS builder

WORKDIR /gnomock/
ADD go.mod .
ADD go.sum .
RUN go mod download -x
RUN go mod verify
ADD . .
ARG GOARCH
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /gnomockd ./cmd/server

FROM scratch

COPY --from=builder /gnomockd /gnomockd
ENV GNOMOCK_ENV=gnomockd
ENTRYPOINT ["/gnomockd"]
