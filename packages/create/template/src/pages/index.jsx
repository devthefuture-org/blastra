import { Link } from "wouter"

export default function Home({ data }) {
  return (
    <div style={{ maxWidth: "1200px", margin: "0 auto", padding: "2rem" }}>
      <h1 style={{ textAlign: "center", marginBottom: "2rem" }}>{data.title}</h1>
      <div
        style={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          margin: "2rem 0",
          padding: "1rem",
        }}
      >
        <img
          src="/home.png"
          alt="Home"
          style={{
            maxWidth: "80%",
            height: "auto",
            borderRadius: "4px",
          }}
        />
      </div>
      <p
        style={{
          textAlign: "center",
          fontSize: "1.1rem",
          lineHeight: "1.6",
          marginBottom: "2rem",
        }}
      >
        {data.description}
      </p>
      <div style={{ marginBottom: "2rem" }}>
        <h2 style={{ textAlign: "center", marginBottom: "1rem" }}>Projects</h2>
        <ul
          style={{
            listStyle: "none",
            padding: 0,
            display: "flex",
            justifyContent: "center",
            gap: "1rem",
          }}
        >
          <li>
            <Link
              href="/project/1"
              style={{
                textDecoration: "none",
                color: "#007bff",
                padding: "0.5rem 1rem",
                border: "1px solid #007bff",
                borderRadius: "4px",
                transition: "all 0.3s ease",
              }}
            >
              Project 1
            </Link>
          </li>
          <li>
            <Link
              href="/project/2"
              style={{
                textDecoration: "none",
                color: "#007bff",
                padding: "0.5rem 1rem",
                border: "1px solid #007bff",
                borderRadius: "4px",
                transition: "all 0.3s ease",
              }}
            >
              Project 2
            </Link>
          </li>
          <li>
            <Link
              href="/project/3"
              style={{
                textDecoration: "none",
                color: "#007bff",
                padding: "0.5rem 1rem",
                border: "1px solid #007bff",
                borderRadius: "4px",
                transition: "all 0.3s ease",
              }}
            >
              Project 3
            </Link>
          </li>
        </ul>
      </div>
      <div
        id="home-content"
        style={{
          textAlign: "center",
          marginBottom: "2rem",
        }}
      >
        {data.content}
      </div>
      <div style={{ textAlign: "center" }}>
        <Link
          href="/fail"
          style={{
            textDecoration: "none",
            color: "#6c757d",
            padding: "0.5rem 1rem",
            border: "1px solid #6c757d",
            borderRadius: "4px",
            transition: "all 0.3s ease",
          }}
        >
          About
        </Link>
      </div>
    </div>
  )
}
