import express from "express"
import { createServer as createViteServer } from "vite"
import { resolve } from "path"

import render from "./render.js"

/**
 * Global process-level error handlers to ensure crashes produce clear logs.
 */
process.on("uncaughtException", (err) => {
  console.error("uncaughtException:", err)
  // Exit to let supervisor (Go) restart and capture exit code + stderr tail
  process.exit(1)
})
process.on("unhandledRejection", (reason) => {
  console.error("unhandledRejection:", reason)
  process.exit(1)
})

async function createServer() {
  const app = express()
  const isProd = process.env.NODE_ENV === "production"

  let vite
  if (!isProd) {
    vite = await createViteServer({
      server: { middlewareMode: true },
      appType: "custom",
    })
    app.use(vite.middlewares)
  } else {
    app.use(
      express.static(resolve(process.cwd(), "dist/client"), {
        index: false,
      })
    )
  }

  app.use("*", async (req, res, next) => {
    const url = req.originalUrl
    try {
      let finalHtml, statusCode
      if (isProd) {
        const result = await render({ url })
        finalHtml = result.html
        statusCode = result.statusCode
      } else {
        const { render: devRender } = await vite.ssrLoadModule("virtual:blastra/entry-server.jsx")
        const result = await devRender({ url })
        finalHtml = await vite.transformIndexHtml(url, "<!DOCTYPE html>" + result.html)
        statusCode = result.statusCode
      }

      res
        .status(statusCode || 200)
        .set({ "Content-Type": "text/html" })
        .end(finalHtml)
    } catch (e) {
      if (!isProd && vite) {
        vite.ssrFixStacktrace(e)
      }
      console.error("SSR error:", e)
      next(e)
    }
  })

  const port = process.env.PORT || 5173

  // Store active connections
  const connections = new Set()

  // Create HTTP server instance
  const server = app.listen(port, () => {
    const mode = isProd ? "production" : "development"
    console.log(`Server running at http://localhost:${port} (${mode} mode)`)
    // Machine-parsable readiness token for the Go worker supervisor
    // Keep it on a single line for easy detection.
    console.log(`BLASTRA_READY port=${port} pid=${process.pid} mode=${mode}`)
  })

  // Track connections
  server.on("connection", (connection) => {
    connections.add(connection)
    connection.on("close", () => {
      connections.delete(connection)
    })
  })

  // Graceful shutdown function
  const gracefulShutdown = async (signal) => {
    console.log(`\n${signal} received. Starting graceful shutdown...`)

    // Stop accepting new connections
    server.close(async (err) => {
      if (err) {
        console.error("Error during server shutdown:", err)
        process.exit(1)
      }

      // Close Vite server in development mode
      if (!isProd && vite) {
        try {
          await vite.close()
          console.log("Vite dev server closed")
        } catch (error) {
          console.error("Error closing Vite server:", error)
        }
      }

      // Close all existing connections
      for (const connection of connections) {
        connection.destroy()
      }

      console.log("Graceful shutdown completed")
      process.exit(0)
    })

    // Force shutdown after 10 seconds
    setTimeout(() => {
      console.error("Forced shutdown after timeout")
      process.exit(1)
    }, 10000)
  }

  // Handle shutdown signals
  process.on("SIGTERM", () => gracefulShutdown("SIGTERM"))
  process.on("SIGINT", () => gracefulShutdown("SIGINT"))

  return { server, vite }
}

createServer()
