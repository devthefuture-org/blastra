import { Link } from 'wouter'

import Error404 from '@/pages/_404'

export default function Error4xx({ statusCode }) {
  if (statusCode === 404) {
    return <Error404 />
  }
  return (
    <div>
      <h1>{statusCode || 'Client Error'}</h1>
      <p>Sorry, something went wrong with your request.</p>
      <Link href="/">Go Home</Link>
    </div>
  )
}
