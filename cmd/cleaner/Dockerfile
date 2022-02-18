FROM golang:latest AS builder

WORKDIR /cleaner/
ADD go.mod .
ADD go.sum .
RUN go mod download -x
RUN go mod verify
ADD . .
ARG GOARCH
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o /gnomock-cleaner .

FROM scratch

COPY --from=builder /gnomock-cleaner /gnomock-cleaner
ENTRYPOINT ["/gnomock-cleaner"]
