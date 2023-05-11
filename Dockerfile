FROM golang:1.20.4-alpine AS builder
ARG BUILD_VERSION
ARG BUILD_DATE
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o steamquery -trimpath -ldflags="-s -w -X main.buildVersion=$BUILD_VERSION -X main.buildDate=$BUILD_DATE" ./...

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/files ./files
COPY --from=builder /app/steamquery ./
CMD ["./steamquery"]