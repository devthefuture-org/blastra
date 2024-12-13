import { Link } from "wouter"
// import NotFound from '@blastra/core/components/NotFound'
// export default NotFound
export default function NotFound() {
  return (
    <div>
      <h1>😺 Optional Custom 404 - Page Not Found</h1>
      <p>The page you're looking for doesn't exist.</p>
      <Link href="/">Go Home</Link>
    </div>
  )
}
