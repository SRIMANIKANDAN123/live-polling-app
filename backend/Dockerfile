FROM golang:1.22-alpine AS build
WORKDIR /app
RUN apk add --no-cache git
ENV GOTOOLCHAIN=auto
COPY go.mod ./
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /server ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /server /app/server
EXPOSE 8080
ENTRYPOINT ["/app/server"]