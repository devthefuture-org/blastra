export default function Html({ lang='en', children }) {
  return (
    <html lang={lang}>
      {children}
    </html>
  )
}
