export default function handler(req, res) {
  const { projectId } = req.query

  res.status(200).json({
    title: `Project ${projectId}`,
    description: `This is project ${projectId}`,
    content: `Detailed content for project ${projectId}. This project demonstrates our capabilities in building scalable solutions.`
  })
}
