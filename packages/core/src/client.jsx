import { hydrateRoot } from "react-dom/client"
import App from "./components/App"

// Disable HMR in production
if (import.meta.hot && process.env.NODE_ENV !== "production") {
  import.meta.hot.accept()
}

export function hydrate({ router }) {
  const { data: initialData, fieldHydrators, statusCode } = window.__BLASTRA_HYDRATION__

  // Process field hydrators
  for (const [key, hydrator] of Object.entries(fieldHydrators)) {
    if (hydrator.func) {
      // Execute hydration function
      const fn = new Function(`return ${hydrator.func}`)()
      initialData[key] = fn()
    } else {
      const { id, attr } = hydrator
      const element = document.getElementById(id)
      if (element) {
        initialData[key] = attr ? element.getAttribute(attr) : element.innerHTML
      }
    }
  }

  hydrateRoot(
    document.getElementById("app"),
    <App
      router={router}
      url={window.location.pathname}
      initialData={initialData}
      statusCode={statusCode}
    />
  )
}
