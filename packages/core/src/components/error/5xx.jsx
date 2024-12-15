import { Link } from "wouter"

import Error500 from "@/pages/_500"
import Error503 from "@/pages/_503"

export default function Error5xx({ statusCode }) {
  if (statusCode === 500) {
    return <Error500 />
  }
  if (statusCode === 503) {
    return <Error503 />
  }
  return (
    <div>
      <h1>{statusCode || "Server Error"}</h1>
      <p>Oops! Something went wrong on our end. Please try again later.</p>
      <Link href="/">Go Home</Link>
    </div>
  )
}
