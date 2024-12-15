import { Router } from "wouter"
import PageWrapper from "./PageWrapper"

export default function App({ url, initialData, router, statusCode }) {
  return (
    <Router ssrPath={url}>
      <PageWrapper initialData={initialData} router={router} statusCode={statusCode} />
    </Router>
  )
}
