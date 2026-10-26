FROM golang:1.24-alpine AS build
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/orenkod-api .

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/orenkod-api /app/orenkod-api
EXPOSE 8080
USER 10001:10001
ENTRYPOINT ["/app/orenkod-api"]
