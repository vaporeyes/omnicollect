# ABOUTME: Multi-stage Docker build for OmniCollect cloud deployment.
# ABOUTME: Builds Vue frontend first (for go:embed), then Go binary, packages into alpine.

# Stage 1: Build Vue frontend
FROM node:18-alpine AS node-builder
ARG VITE_AUTH0_DOMAIN=""
ARG VITE_AUTH0_CLIENT_ID=""
ARG VITE_AUTH0_AUDIENCE=""
ENV VITE_AUTH0_DOMAIN=$VITE_AUTH0_DOMAIN
ENV VITE_AUTH0_CLIENT_ID=$VITE_AUTH0_CLIENT_ID
ENV VITE_AUTH0_AUDIENCE=$VITE_AUTH0_AUDIENCE
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Stage 2: Build Go binary (needs frontend/dist for go:embed)
FROM golang:latest AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
COPY auth/ ./auth/
COPY ai/ ./ai/
COPY storage/ ./storage/
COPY showcase/ ./showcase/
COPY --from=node-builder /app/frontend/dist ./frontend/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o omnicollect .

# Stage 3: Runtime
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=go-builder /app/omnicollect .
COPY --from=node-builder /app/frontend/dist ./frontend/dist

EXPOSE 8080

ENTRYPOINT ["./omnicollect", "--serve"]
