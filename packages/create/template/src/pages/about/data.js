import { API_URL } from "../../config.js"

export const fieldsConfig = {
  teamInfo: {
    hydrateId: "team-info",
  },
}

export async function loader() {
  const response = await fetch(`${API_URL}/users?_limit=3`)
  const users = await response.json()
  
  const teamInfo = users.map(user => (
    `${user.name} - Expert in ${user.company.bs}`
  )).join('\n\n')
  
  return {
    title: "About Us",
    description: "Learn more about our team",
    teamInfo: "Our amazing team is dedicated to building great software:\n\n" + teamInfo,
  }
}
