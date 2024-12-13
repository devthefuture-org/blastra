import { API_URL } from "../config.js"

export const fieldsConfig = {
  content: {
    hydrateFunction: () => document.getElementById("home-content").innerHTML,
  },
}

export async function loader() {
  const response = await fetch(`${API_URL}/`)
  const data = await response.json()
  return data
}
