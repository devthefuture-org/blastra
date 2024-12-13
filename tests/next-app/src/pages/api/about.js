export default function handler(req, res) {
  res.status(200).json({
    title: 'About Us',
    description: 'Learn more about our team',
    teamInfo: 'We are a dedicated team working on amazing projects.'
  })
}
