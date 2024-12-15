import express from "express"
import cors from "cors"

const app = express()
const port = 3001

app.use(cors())
app.use(express.json())

// Home page data
app.get("/api/", (req, res) => {
  res.json({
    title: "Welcome to Blastra Template",
    description: "This is the home page",
    content: '"When you let go of what you are, you become what you might be." Lao Tzu',
  })
})

// About page data
app.get("/api/about", (req, res) => {
  res.json({
    title: "About Us",
    description: "Learn more about our team",
    teamInfo: "Our amazing team is dedicated to building great software.",
  })
})

// Project page data
app.get("/api/project/:projectId", (req, res) => {
  const { projectId } = req.params
  res.json({
    title: `Project ${projectId}`,
    description: `Details about project ${projectId}`,
    content: `This is the detailed content for project ${projectId}`,
  })
})

app.listen(port, () => {
  console.log(`API server running at http://localhost:${port}`)
})
