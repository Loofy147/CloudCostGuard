# build stage
FROM golang:1.22-alpine AS build

WORKDIR /app

# Copy the entire project context
COPY . .

# Build the CLI binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /cloudcostguard ./cmd/cloudcostguard

# final stage
FROM alpine:latest
COPY --from=build /cloudcostguard /cloudcostguard
COPY testdata /testdata
ENTRYPOINT ["/cloudcostguard"]
