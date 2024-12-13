import render from "./render.js"

async function main() {
  const url = process.argv[2] || '/'
  
  let html = null
  let error = null
  
  try {
    const output = await render({ url })
    if (output && output.html) {
      html = output.html
    } else {
      error = new Error('No HTML output generated')
    }
  } catch (e) {
    error = e
  }
  
  process.stdout.write(JSON.stringify({ html, error: error ? error.message : null }))
}

main()
