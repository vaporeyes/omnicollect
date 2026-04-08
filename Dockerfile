# ABOUTME: Multi-stage Docker build for OmniCollect cloud deployment.
# ABOUTME: Builds Go binary + Vue frontend, packages into minimal alpine image.

# Stage 1: Build Go binary
FROM golang:1.25-alpine AS go-builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY storage/ ./storage/
RUN CGO_ENABLED=0 GOOS=linux go build -o omnicollect .

# Stage 2: Build Vue frontend
FROM node:18-alpine AS node-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Stage 3: Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=go-builder /app/omnicollect .
COPY --from=node-builder /app/frontend/dist ./frontend/dist

EXPOSE 8080

ENTRYPOINT ["./omnicollect", "--serve"]
