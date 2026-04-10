FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o yangobot ./cmd

# ---

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/yangobot /yangobot

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/yangobot"]
