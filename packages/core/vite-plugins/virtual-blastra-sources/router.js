import { createFileBasedRouter } from "@blastra/core/router"
export default createFileBasedRouter({
  pageModules: import.meta.glob(
    [
      "@/pages/**/index.jsx",
      "@/pages/index.jsx", // Include root index
    ],
    { eager: true }
  ),
  dataModules: import.meta.glob(
    [
      "@/pages/**/data.js",
      "@/pages/data.js", // Include root data
    ],
    { eager: true }
  ),
  metaModules: import.meta.glob(
    [
      "@/pages/**/meta.js",
      "@/pages/meta.js", // Include root meta
    ],
    { eager: true }
  ),
})
