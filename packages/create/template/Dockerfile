# Stage 1: Build Node.js app
FROM node:20-alpine AS node-builder

WORKDIR /app

# Copy yarn configuration first
COPY yarn.lock .yarnrc.yml ./
COPY .yarn .yarn

# Use yarn fetch for optimized dependency installation
RUN yarn fetch

# Copy the rest of the source files
COPY . .

# Build the application
RUN yarn postinstall
RUN yarn build
RUN yarn workspaces focus --production && yarn cache clean

# Stage 2: Build Go server
FROM golang:1.22-alpine AS go-builder

# Build arguments for Go module
ARG MODULE_PATH=github.com/devthefuture-org/blastra
ARG MODULE_VERSION=main
ARG GO_BUILD_FLAGS=""

WORKDIR /app

# Install and build the Go server
ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64
RUN go install ${MODULE_PATH}@${MODULE_VERSION} ${GO_BUILD_FLAGS} && \
  cp $(which blastra) /app/blastra-server

# Stage 3: Final image
FROM alpine:latest

USER 1000

WORKDIR /app

# Copy built assets from Node.js build
COPY --from=node-builder /app/dist ./dist
COPY --from=node-builder /app/node_modules ./node_modules


# Copy Go binary
COPY --from=go-builder /app/blastra-server .

EXPOSE 8080

CMD ["./blastra-server"]
