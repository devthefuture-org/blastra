import Link from 'next/link'
import { API_URL } from '../../config.js'

export default function About({ data }) {
  return (
    <div>
      <h1>{data.title}</h1>
      <p>{data.description}</p>
      <div id="team-info">{data.teamInfo}</div>
      <Link href="/">Home</Link>
    </div>
  )
}

export async function getServerSideProps() {
  const response = await fetch(`${API_URL}/about`)
  const data = await response.json()
  
  return {
    props: {
      data
    }
  }
}
