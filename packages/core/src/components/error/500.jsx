import { Link } from 'wouter'

export default function Error500() {
  return (
    <div>
      <h1>500 - Internal Server Error</h1>
      <p>Something went wrong on our end. Please try again later.</p>
      <Link href="/">Go Home</Link>
    </div>
  )
}
