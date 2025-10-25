FROM golang:1.22-alpine

WORKDIR /app
COPY test/mock-github/main.go .
RUN go build -o /mock-github .

EXPOSE 8081
ENTRYPOINT ["/mock-github"]
