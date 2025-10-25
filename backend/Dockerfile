# build stage
FROM golang:1.22-alpine AS build

WORKDIR /app

# Copy the entire project context
COPY . .

# Build the backend binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /cloudcostguard-backend ./backend

# final stage
FROM alpine:latest
COPY --from=build /cloudcostguard-backend /cloudcostguard-backend
EXPOSE 8080
ENTRYPOINT ["/cloudcostguard-backend"]
