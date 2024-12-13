import { API_URL } from "@/config"

export const fieldsConfig = {
  content: {
    hydrateId: "project-content",
  },
}

export async function loader({ params }) {
  const { projectId } = params
  const response = await fetch(`${API_URL}/project/${projectId}`)
  const data = await response.json()
  return data
}
