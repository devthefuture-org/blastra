import fs from "fs"
import { resolve } from "path"

export default () => {
  const prefix = "virtual:blastra/"
  return {
    name: "virtual-blastra",
    resolveId(source) {
      const cwd = process.cwd()
      if (source.startsWith(cwd)) {
        source = source.slice(cwd.length + 1)
      }
      if (source.startsWith(prefix)) {
        return "\0" + source
      }
      return null
    },
    load(id) {
      if (id.startsWith("\0" + prefix)) {
        return fs.readFileSync(
          resolve(
            "node_modules/@blastra/core/vite-plugins/virtual-blastra-sources",
            id.slice(1 + prefix.length)
          ),
          "utf-8"
        )
      }
      return null
    },
  }
}
