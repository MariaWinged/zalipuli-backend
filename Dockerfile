FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=./api/gen-server.yaml ./api/openapi.yaml
RUN go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=./api/gen-client.yaml ./api/openapi.yaml
RUN go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=./api/gen-types.yaml ./api/openapi.yaml

RUN CGO_ENABLED=0 GOOS=linux go build -o server .

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/server .

ARG PORT=8080
ENV PORT=${PORT}

EXPOSE ${PORT}

CMD ["./server"]
