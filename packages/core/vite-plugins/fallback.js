import fs from "fs"
import path from "path"
import { createRequire } from "module"
const require = createRequire(import.meta.url)

export default function fallbackPlugin({ alias, fallbacks }) {
  return {
    name: "fallback",
    async resolveId(source, _importer) {
      const cwd = process.cwd()
      let sourcePath = source
      if (source.startsWith(cwd)) {
        sourcePath = source.slice(cwd.length)
      }
      for (const [target, fallback] of Object.entries(fallbacks)) {
        let targetPath = target
        for (const [aliasSrc, aliasDest] of Object.entries(alias)) {
          if (targetPath.startsWith(aliasSrc)) {
            targetPath = path.join(aliasDest, targetPath.slice(aliasSrc.length))
            break
          }
        }
        if (targetPath.startsWith(cwd)) {
          targetPath = targetPath.slice(cwd.length)
        }
        if (targetPath !== sourcePath) {
          continue
        }
        const exists = fs.existsSync(sourcePath)
        if (exists) {
          return null
        }

        // console.log(
        //   `[vite-plugin-fallback] module "${target}" not found. Falling back to "${fallback}".`
        // );
        return require.resolve(fallback)
      }
    },
  }
}
