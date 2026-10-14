# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder
WORKDIR /src

# Only the module files plus the two package trees the binary imports are
# copied in; the explicit paths replace .dockerignore, which is no longer
# needed.

COPY go.mod go.sum* ./
RUN go mod download

COPY src/cmd ./src/cmd
COPY src/internal ./src/internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ghp ./src/cmd

FROM gcr.io/distroless/static-debian12:nonroot AS runner

LABEL org.opencontainers.image.title="ghp" \
      org.opencontainers.image.description="GHP toolchain — PHP-style templates with real embedded Go" \
      org.opencontainers.image.source="https://github.com/GHP-GoLang-Framework/ghp"

COPY --from=builder /out/ghp /usr/local/bin/ghp

WORKDIR /app
USER nonroot:nonroot

ENTRYPOINT ["ghp"]
CMD ["help"]
