FROM golang:1.25 AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-s -w -X github.com/zetaoss/zavatar/cmd/zavatar.Version=${VERSION}" \
    -o /out/zavatar ./cmd/zavatar

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/zavatar /app/zavatar

EXPOSE 8080
ENTRYPOINT ["/app/zavatar"]
