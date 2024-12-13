import Link from 'next/link'

export default function Home({ data }) {
  return (
    <div>
      <h1>{data.title}</h1>
      <p>{data.description}</p>
      <div>
        <h2>Projects</h2>
        <ul>
          <li>
            <Link href="/project/1">Project 1</Link>
          </li>
          <li>
            <Link href="/project/2">Project 2</Link>
          </li>
          <li>
            <Link href="/project/3">Project 3</Link>
          </li>
        </ul>
      </div>
      <div id="home-content">{data.content}</div>
      <Link href="/about">About</Link>
    </div>
  )
}

// Match Blastra's data fetching pattern
export async function getStaticProps() {
  // Import data the same way Blastra does
  const data = {
    title: "Welcome to Next.js",
    description: "This is a Next.js app for benchmarking against Blastra",
    content: "Main content here"
  };

  return {
    props: {
      data
    }
  };
}
