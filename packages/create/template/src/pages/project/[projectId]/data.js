import { API_URL } from "../../../config.js"

export const fieldsConfig = {
  content: {
    hydrateId: "project-content",
  },
}

export async function loader({ params }) {
  const { projectId } = params
  const response = await fetch(`${API_URL}/posts/${projectId}`)
  const post = await response.json()
  
  // Transform the post data into a project-like format
  const projectContent = `
Project Overview:
${post.title}

Details:
${post.body}

Key Features:
- Innovative solution design
- Scalable architecture
- User-centric approach
  `.trim()
  
  return {
    title: `Project ${projectId}`,
    description: `Details about project ${projectId}`,
    content: projectContent,
  }
}
