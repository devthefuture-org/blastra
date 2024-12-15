export default async function handler(req, res) {
  const { projectId } = req.query

  try {
    const response = await fetch(`https://jsonplaceholder.typicode.com/posts/${projectId}`)
    const post = await response.json()

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

    res.status(200).json({
      title: `Project ${projectId}`,
      description: `Details about project ${projectId}`,
      content: projectContent
    })
  } catch (error) {
    res.status(500).json({ error: 'Failed to fetch project data' })
  }
}
