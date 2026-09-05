# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ghp ./src/cmd

FROM gcr.io/distroless/static-debian12:nonroot AS runner

LABEL org.opencontainers.image.title="ghp" \
      org.opencontainers.image.description="GHP toolchain — PHP-style templates with real embedded Go" \
      org.opencontainers.image.source="https://github.com/GHP-GoLang-Framework/GHP"

COPY --from=builder /out/ghp /usr/local/bin/ghp

WORKDIR /app
USER nonroot:nonroot

ENTRYPOINT ["ghp"]
CMD ["help"]
