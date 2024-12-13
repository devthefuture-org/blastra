# Server Package

This package implements the HTTP server functionality for the Blastra application. It has been organized into several focused modules for better maintainability and clarity.

## Package Structure

### Core Components

- `server.go`: Contains core server types and the main SSR handler orchestration
  - Defines the `Config` type for server configuration
  - Implements the main `SSRHandler` that coordinates between caching, worker-based SSR, and direct SSR

### Server Setup and Configuration

- `mime.go`: Handles MIME type initialization
  - Registers additional MIME types for various file extensions
  - Ensures proper content type handling for different file types

- `routes.go`: Implements route setup and request handling
  - Sets up main HTTP routes
  - Handles static file serving and SSR fallback
  - Configures health check endpoints

### Static File Handling

- `static.go`: Implements static file serving functionality
  - Provides the `staticFileHandler` for serving static files
  - Handles proper content type setting
  - Implements caching and security headers

### Server-Side Rendering (SSR)

- `ssr_worker.go`: Implements worker-based SSR handling
  - Manages SSR requests through the worker pool
  - Handles worker endpoint communication
  - Implements response caching for worker results

- `ssr_direct.go`: Implements direct command execution SSR
  - Handles SSR through direct command execution
  - Manages command input/output
  - Implements response parsing and error handling

## Key Features

- Modular design with clear separation of concerns
- Flexible SSR implementation with worker pool support
- Comprehensive static file serving with proper MIME types
- Built-in caching for SSR responses
- Health check endpoints for monitoring
- Security headers and proper content type handling

## Usage

The server package is typically initialized through the main application, which:

1. Creates a server configuration
2. Initializes the SSR cache and worker pool
3. Sets up routes and handlers
4. Starts the HTTP/HTTPS server

Example:

```go
config := &server.Config{
    BlastraCWD:    cwd,
    StaticDir:     staticDir,
    SSRHandler:    ssrHandler,
    HealthChecker: healthChecker,
}

serverInitConfig := &server.ServerInitConfig{
    HTTPPort:     httpPort,
    GzipEnabled:  true,
    ServerConfig: config,
}

srv := server.InitializeServer(serverInitConfig)
