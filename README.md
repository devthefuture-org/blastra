# ⚡️ BLASTRA ✴️

<p align="center">
  <img src="docs/assets/blastra_logo.png" width="200" alt="Blastra Logo">
</p>

Fast. Reliable. Stellar.

Blastra is an **opinionated SSR (Server-Side Rendering) React framework** built on top of **Vite.js**. It combines file-based routing, isomorphic data loading, and optional high-performance Go server deployments to provide a straightforward and scalable solution for building SEO-friendly React web applications.

- [1. Key Features](#1-key-features)
- [2. Getting Started](#2-getting-started)
  - [Project Creation](#project-creation)
  - [Project Scripts](#project-scripts)
- [3. Directory Structure \& Routing](#3-directory-structure--routing)
  - [File-Based Routing](#file-based-routing)
  - [Dynamic Routes](#dynamic-routes)
  - [Custom 404 and Global Head](#custom-404-and-global-head)
- [4. Data Loading \& SSR](#4-data-loading--ssr)
  - [data.js and the `loader` function](#datajs-and-the-loader-function)
  - [fieldsConfig for Hydration Optimization](#fieldsconfig-for-hydration-optimization)
  - [meta.js for SEO Metadata](#metajs-for-seo-metadata)
  - [Isomorphic Fetch](#isomorphic-fetch)
  - [No Query Param SSR Logic](#no-query-param-ssr-logic)
- [5. Hydration \& Avoiding Duplicate Payloads](#5-hydration--avoiding-duplicate-payloads)
- [6. The Optional Go Server](#6-the-optional-go-server)
  - [Docker Deployment](#docker-deployment)
  - [Caching \& Scalability Strategies](#caching--scalability-strategies)
- [7. Under the Hood](#7-under-the-hood)
  - [Core Packages](#core-packages)
  - [Vite Plugins](#vite-plugins)
  - [File Server \& SSR Execution](#file-server--ssr-execution)
  - [Caching Architecture](#caching-architecture)
  - [Worker Pool / SSR Workers](#worker-pool--ssr-workers)
- [8. FAQ](#8-faq)
- [9. Further Reading \& Contributing](#9-further-reading--contributing)


* * *

## 1. Key Features

* **File-Based Routing**: Organize routes by placing React components in `src/pages`. Folder names reflect URL paths (including dynamic parameters).
* **Isomorphic Data Loading**: Each page can provide a `loader()` function in `data.js` that runs on both server and client, ensuring consistent data fetching with `fetch`.
* **SEO-First SSR**: Blastra focuses on SSR primarily for SEO. Personalized user logic remains on the client side after hydration.
* **Hydration Optimization**: Use `fieldsConfig` to avoid duplicating large data (e.g., HTML strings) in both SSR HTML and JSON inline scripts.
* **Optional High-Performance Go Server**: Build and run your Blastra app on a Go-based SSR server for improved performance.
* **Docker-Ready**: Built-in Dockerfile and docker-compose integration.
* **Vite-Based**: Enjoy the speed and simplicity of Vite dev server and the flexible plugin ecosystem.

* * *

## 2. Getting Started

### Project Creation

The easiest way to create a new Blastra project is with npx:

```bash
npx create-blastra
```

This command scaffolds a new project with sensible defaults. You'll get a directory structure like:

```kotlin
my-blastra-app
├── docker-compose.yml
├── Dockerfile
├── .env.example
├── public/
├── src/
│   ├── config.js
│   ├── index.css
│   └── pages/
│       ├── _404.jsx
│       ├── _head.jsx
│       ├── about/
│       │   ├── data.js
│       │   ├── index.jsx
│       │   └── meta.js
│       ├── data.js
│       ├── index.jsx
│       ├── meta.js
│       └── project/
│           └── [projectId]/
│               ├── data.js
│               ├── index.jsx
│               └── meta.js
├── package.json
└── yarn.lock
```

### Project Scripts

Within your newly created project, the following scripts are available:

```bash
yarn dev       # Start development server with SSR + HMR (Vite in middlewareMode)
yarn build     # Build the project for production
yarn preview   # Quickly preview the production build (Node.js-based SSR)
yarn start     # Start the SSR server in production mode (Node.js)
```

For Docker-based deployment using the **Go server**:

```bash
docker compose build app
# Builds a production container image using the Go server for SSR
```

* * *

## 3. Directory Structure & Routing

### File-Based Routing

Blastra automatically creates routes based on the folder and file structure under `src/pages`. Here's how it works:

* **`src/pages/index.jsx`** → Renders at the root path `/`.
* **`src/pages/about/index.jsx`** → Renders at `/about`.
* **`src/pages/project/[projectId]/index.jsx`** → Dynamic route `/project/:projectId`.

For each route folder (like `about`, `project/[projectId]`), there are typically three files:

1. **`index.jsx`** — Exports the React component for that page.
2. **`data.js`** — Exports the `loader()` function and an optional `fieldsConfig` object.
3. **`meta.js`** — Exports the `meta(data)` function for SEO metadata.

### Dynamic Routes

Dynamic routes are folder names wrapped in brackets, such as `[projectId]`. These become URL parameters. For example:  
`src/pages/project/[projectId]/index.jsx` => The route `/project/:projectId`

Inside `data.js`:

```js
export async function loader({ params }) {
  // params.projectId contains the dynamic segment
}
```

### Custom 404 and Global Head

* **Custom 404**: Create `_404.jsx` in `src/pages` for a custom Not Found page.
* **Global Head**: `_head.jsx` in `src/pages` overrides or extends the default `<head>` logic. You can inject global meta tags, favicons, etc.

* * *

## 4. Data Loading & SSR

### data.js and the `loader` function

Each route folder can contain a `data.js` file that exports:

* A **`loader()`** function: Runs on the server for the initial SSR, and again on the client for client-side transitions in SPA mode.
* An optional **`fieldsConfig`**: Configures advanced hydration strategies.

```js
// Example: src/pages/about/data.js
import { API_URL } from "../../config.js"

export const fieldsConfig = {
  teamInfo: {
    hydrateId: "team-info",
  },
}

export async function loader() {
  const response = await fetch(`${API_URL}/about`) // isomorphic fetch
  const data = await response.json()
  return data
}
```

**Key Points**:

* **Isomorphic**: Use the same `fetch` code for both SSR and client side.
* The return value of `loader()` is passed to the corresponding page component's `props.data` on SSR. On the client, the framework calls `loader()` again during navigation to fetch data.

### fieldsConfig for Hydration Optimization

`fieldsConfig` is an object describing how certain fields in your data should be hydrated on the client. This helps **avoid sending large payloads** twice (in HTML and inline JSON). Example:

```js
export const fieldsConfig = {
  content: {
    // Instead of putting the entire content in inline JSON, 
    // retrieve it from an element in the DOM.
    hydrateFunction: () => document.getElementById("home-content").innerHTML,
  },
}
```

When the client hydrates, Blastra uses either:

* `hydrateFunction` – a JS function string that runs client-side to fetch the DOM content.
* `hydrateId` – an HTML element ID where the data can be extracted (innerHTML or attribute).

### meta.js for SEO Metadata

Each route folder can also have a `meta.js` file exporting a `meta(data)` function that receives the result of `loader()` and returns page metadata:

```js
// Example: src/pages/about/meta.js
export function meta(data) {
  return {
    title: data.title,
    meta: [
      { name: "description", content: data.description },
      { name: "og:title", content: data.title },
      { name: "og:description", content: data.description },
    ],
  }
}
```

Blastra injects these tags server-side for SEO. At runtime (client-side), a `<MetaUpdater>` component updates the `<head>` dynamically.

### Isomorphic Fetch

Your **backend API** should be separate from Blastra. All data loading uses isomorphic `fetch`. For example:

```js
// src/pages/project/[projectId]/data.js
export async function loader({ params }) {
  const response = await fetch(`http://localhost:3001/api/project/${params.projectId}`)
  return await response.json()
}
```

### No Query Param SSR Logic

Blastra **does not** handle query parameters on the server side. Only path-based routing is SSR'd (use dynamic routes if needed). Query params logic runs on the **client side**.

* * *

## 5. Hydration & Avoiding Duplicate Payloads

By default, SSR includes an inline JSON hydration script containing the page's data. If some fields are large (e.g., big HTML blocks), you can instruct Blastra to **omit** them from the JSON blob and pick them up from the DOM directly.

```js
export const fieldsConfig = {
  content: {
    hydrateFunction: () => document.getElementById("home-content").innerHTML,
  },
}
```

This keeps the HTML payload lighter since that large content is not duplicated in the inline script.

* * *

## 6. The Optional Go Server

Blastra comes with a **Go-based SSR server** that can serve your production build. Key benefits include:

* **Performance**: Often faster SSR than Node-based servers.
* **Caching & Scalability**: Built-in in-memory, Redis, or filesystem cache.
* **Load Balancing**: Integrate into microservices or container orchestration easily.

### Docker Deployment

A `Dockerfile` and `docker-compose.yml` are included by default:

```bash
docker compose build app
docker compose up
```

Once running, the container will serve your SSR app on the configured port (commonly `8080` or `3000`).

### Caching & Scalability Strategies

The Go server supports multiple caching layers:

1. **In-Memory** SSR Cache: Speeds up repeated SSR requests for the same route/data.
2. **Redis** or **Filesystem** External Cache: For distributed or persistent caching.
3. **NotFound Cache**: Specifically caches 404 pages to quickly serve repeated "not found" routes.

All caches have configurable TTL and max size, and can be layered for robust performance under heavy load.

* * *

## 7. Under the Hood

### Core Packages

* **`@blastra/core`**: The main SSR logic, Vite configuration, file-based router, hydration mechanisms, and error boundaries.
* **`create-blastra`**: A CLI tool that scaffolds new Blastra projects.
* **Go Server** (in `pkg/`): High-performance server with SSR caching, rate limiting, and health checks.

### Vite Plugins

In `@blastra/core`, you'll find custom Vite plugins:

* **`virtual-blastra`** plugin: Provides virtual modules that tie together client and server entry points (`entry-client.jsx` / `entry-server.jsx`) for SSR.
* **`fallback`** plugin: Fallbacks certain special pages (`_head`, `_404`, etc.) to core defaults if not defined in your project.

### File Server & SSR Execution

When running Node-based SSR (via `yarn start` or `yarn preview`), Blastra:

* Uses an Express server in dev/prod mode,
* Integrates Vite either in middleware mode (development) or static mode (production).

`yarn build` outputs two folders:

* `dist/client` — client bundle
* `dist/server` — SSR server bundle.

The SSR rendering is orchestrated by `render.js`, which imports `entry-server.js`, calls `render`, and returns a full HTML string. That string is sent to the client with `<head>` metadata, hydration scripts, etc.

### Caching Architecture

By default, caching is optional. You can enable SSR caching in the Go or Node server via environment variables:

* **Memory Cache**: Stores rendered HTML keyed by route path.
* **External Cache**: Redis or filesystem for distributed/persistent caching.
* **NotFoundCache**: Specifically caches 404 pages.

### Worker Pool / SSR Workers

Blastra can run SSR in a **worker pool**. By default, SSR is done in the main Node or Go process, but you can configure worker pools for concurrency. The pool is round-robin load-balanced among workers on different ports.

* * *

## 8. FAQ

**Q: What if I want SSR for user-specific pages?**  
A: Blastra focuses on SSR for SEO-friendly, non-personalized pages. For user-specific or authenticated pages, do them on the client side after hydration or create a separate backend approach.

**Q: How do I deploy without the Go server?**  
A: Just run `yarn build && yarn start` on Node. The Go server is optional. For maximum performance or unique caching needs, switch to the Go version with Docker.

* * *

## 9. Further Reading & Contributing

Check out the **Blastra** repository or your new project's codebase for deeper exploration. The `packages/core` folder shows the SSR system, router, and build scripts. The `pkg/` folder contains the optional Go server, caching logic, and advanced SSR server infrastructure.

**Contributions** are welcome. Feel free to open issues or pull requests if you want to extend or improve Blastra's capabilities.

* * *

**Blastra** – Enjoy building your next SEO-friendly React app with isomorphic data loading, straightforward routing, and optional high-performance server options!  
_© 2024 Blastra / DevTheFuture. Released under MIT License._
