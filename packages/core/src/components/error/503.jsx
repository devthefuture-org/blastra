import { Link } from 'wouter'

export default function Error503() {
  return (
    <div>
      <h1>503 - Service Unavailable</h1>
      <p>Sorry, we're currently undergoing maintenance. Please try again later.</p>
      <Link href="/">Go Home</Link>
    </div>
  )
}
