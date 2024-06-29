FROM golang:1.23rc1-alpine3.20 AS Build
WORKDIR /app
# COPY ./templates ./templates/
COPY ../main.go ./
COPY ./go.* ./

ENV CGO_ENABLED=1
# Build
RUN apk add --no-cache gcc musl-dev
RUN go mod download
# RUN go build main.go
RUN go build -ldflags='-s -w -extldflags "-static"' main.go

# FROM golang:1.23rc1-alpine3.20 as Run
# WORKDIR /app
# COPY --from=Build /app/main /app/main
# COPY ./templates /app/templates

# EXPOSE 8080

# # Run
# CMD ["/app/main"]

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=Build /app/main /app/main
COPY ./templates /app/templates

EXPOSE 8080

# Run
CMD ["/app/main"]