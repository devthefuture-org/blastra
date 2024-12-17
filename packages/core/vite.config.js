import { defineConfig } from "vite"
import react from "@vitejs/plugin-react"
import { resolve } from "path"
import virtualBlastra from "@blastra/core/vite-plugins/virtual-blastra"
import fallback from "@blastra/core/vite-plugins/fallback"

const alias = {
  "@": resolve(process.cwd(), "src"),
  "@public": resolve(process.cwd(), "public"),
}

export default defineConfig(({ mode }) => ({
  plugins: [
    virtualBlastra(),
    fallback({
      alias,
      fallbacks: {
        "@/pages/_head": "@blastra/core/components/Head",
        "@/pages/_loader": "@blastra/core/components/Loader",

        "@/pages/_4xx": "@blastra/core/components/error/4xx",
        "@/pages/_5xx": "@blastra/core/components/error/5xx",
        "@/pages/_404": "@blastra/core/components/error/404",
        "@/pages/_500": "@blastra/core/components/error/500",
        "@/pages/_503": "@blastra/core/components/error/503",
      },
    }),
    react(),
  ],
  publicDir: "public",
  base: "/",
  optimizeDeps: {
    include: ["@blastra/core", "react-dom/client", "react-dom"],
  },
  build: {
    rollupOptions: {
      input:
        mode === "production"
          ? "virtual:blastra/entry-client.jsx"
          : {
              client: "virtual:blastra/entry-client.jsx",
              server: "virtual:blastra/entry-server.jsx",
            },
    },
    manifest: true,
    outDir: mode === "production" ? "dist/client" : "dist",
    ssrManifest: true,
  },
  resolve: {
    alias,
  },
  ssr: {
    noExternal: ["wouter", "@blastra/core/router"],
    target: "node",
    format: "esm",
  },
  server: {
    hmr: mode !== "production",
    middlewareMode: true,
  },
}))
