# build stage
FROM golang:1.22-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /cloudcostguard ./cmd/cloudcostguard

# final stage
FROM alpine:latest
COPY --from=build /cloudcostguard /cloudcostguard
ENTRYPOINT ["/cloudcostguard"]
