FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/auth ./cmd/sso && \
    CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/migrator ./cmd/migrator

FROM alpine:3.22
RUN addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/auth /app/auth
COPY --from=build /out/migrator /app/migrator
COPY migrations /app/migrations
RUN mkdir -p /app/out/logs && chown -R app:app /app/out
USER app
EXPOSE 50052
CMD ["/app/auth"]
