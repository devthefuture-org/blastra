import { Link } from "wouter"

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
