FROM golang:1.27-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY Database ./Database
COPY models ./models
COPY handlers ./handlers
COPY routes ./routes
COPY middleware ./middleware
RUN CGO_ENABLED=0 go build -o /api .

FROM alpine:3.23

RUN apk add --no-cache ca-certificates

COPY --from=build /api /api

USER 10001:10001
EXPOSE 8080

ENTRYPOINT ["/api"]
