import { Link } from "wouter"

export default function About({ data }) {
  return (
    <div>
      <h1>{data.title}</h1>
      <p>{data.description}</p>
      <pre id="team-info">{data.teamInfo}</pre>
      <Link href="/">Home</Link>
    </div>
  )
}
