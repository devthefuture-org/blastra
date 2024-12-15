import { resolve } from "path"
import fs from "fs"
import { logger } from "./src/utils/logger.js"

export default async ({ url }) => {
  try {
    const serverEntry = resolve(process.cwd(), "dist/server/entry-server.js")

    const render = (await import(serverEntry)).render

    const manifestPath = resolve(process.cwd(), "dist/client/.vite/manifest.json")

    const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf-8"))

    const { html, statusCode } = await render({ url, manifest })

    return {
      html: "<!DOCTYPE html>" + html,
      statusCode,
    }
  } catch (error) {
    logger.error("SSR failed for URL:", url, "\nError:", error)
    throw error
  }
}
