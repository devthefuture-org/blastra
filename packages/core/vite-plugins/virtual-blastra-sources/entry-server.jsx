import router from "virtual:blastra/router.js"
import { createRender } from "@blastra/core/server"
export { default as router } from "virtual:blastra/router.js"
export const render = createRender({ router })
