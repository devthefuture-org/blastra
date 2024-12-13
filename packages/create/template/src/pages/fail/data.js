export const fieldsConfig = {
  content: {
    hydrateId: "project-content",
  },
}

export async function loader({ params }) {
  const { projectId } = params
  const response = await fetch(`http://localhost/project/${projectId}`)
  const data = await response.json()
  return data
}
