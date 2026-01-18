FROM golang:1.25-alpine AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -buildvcs=false -ldflags='-w -s' \
    -o /smtp ./cmd/smtp

FROM gcr.io/distroless/static-debian12 AS build-release-stage

WORKDIR /

COPY --from=build-stage /smtp /smtp

EXPOSE 2525

USER nonroot:nonroot

ENTRYPOINT ["/smtp"]
