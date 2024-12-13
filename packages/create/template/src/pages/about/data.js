import { API_URL } from "../../config.js"

export const fieldsConfig = {
  teamInfo: {
    hydrateId: "team-info",
  },
}

export async function loader() {
  const response = await fetch(`${API_URL}/about`)
  const data = await response.json()
  return data
}
