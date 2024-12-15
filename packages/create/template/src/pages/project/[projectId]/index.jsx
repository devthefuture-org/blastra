import { Link } from "wouter"

export default function Project({ data }) {
  return (
    <div>
      <h1>{data.title}</h1>
      <p>{data.description}</p>
      <pre id="project-content">{data.content}</pre>
      <Link href="/">Home</Link>
    </div>
  )
}
