import Link from 'next/link'
import { API_URL } from '../../../config.js'

export default function Project({ data }) {
  return (
    <div>
      <h1>{data.title}</h1>
      <p>{data.description}</p>
      <div id="project-content">{data.content}</div>
      <Link href="/">Home</Link>
    </div>
  )
}

export async function getServerSideProps({ params }) {
  const { projectId } = params
  const response = await fetch(`${API_URL}/project/${projectId}`)
  const data = await response.json()
  
  return {
    props: {
      data
    }
  }
}
