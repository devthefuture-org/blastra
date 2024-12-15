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
RUN yarn workspaces foreach -At run postinstall
RUN yarn build
RUN yarn workspaces focus --production && yarn cache clean

# Stage 2: Build Go server
FROM golang:1.22-alpine AS go-builder

# Build arguments for Go module
ARG BLASTRA_GOSERVER_URL=github.com/devthefuture-org/blastra
ARG BLASTRA_GOSERVER_VERSION=main
ARG GO_BUILD_FLAGS=""

WORKDIR /app

# Install and build the Go server
ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64
RUN go install ${BLASTRA_GOSERVER_URL}/pkg@${BLASTRA_GOSERVER_VERSION} ${GO_BUILD_FLAGS} \
  && mv /go/bin/pkg /go/bin/blastra-server

# Stage 3: Final image
FROM alpine:latest

USER 1000

WORKDIR /app

# Copy built assets from Node.js build
COPY --from=node-builder /app/dist ./dist
COPY --from=node-builder /app/node_modules ./node_modules

# Copy Go binary from the Go binary path and ensure it's in PATH
COPY --from=go-builder /go/bin/blastra-server /usr/local/bin/blastra-server
ENV PATH="/usr/local/bin:${PATH}"

EXPOSE 8080

CMD ["blastra-server"]
