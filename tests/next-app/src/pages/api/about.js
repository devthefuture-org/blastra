export default async function handler(req, res) {
  try {
    const response = await fetch('https://jsonplaceholder.typicode.com/users?_limit=3')
    const users = await response.json()
    
    const teamInfo = users.map(user => (
      `${user.name} - Expert in ${user.company.bs}`
    )).join('\n\n')
    
    res.status(200).json({
      title: 'About Us',
      description: 'Learn more about our team',
      teamInfo: "Our amazing team is dedicated to building great software:\n\n" + teamInfo
    })
  } catch (error) {
    res.status(500).json({ error: 'Failed to fetch team data' })
  }
}
