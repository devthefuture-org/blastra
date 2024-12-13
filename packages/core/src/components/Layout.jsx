import sanitizeId from "../utils/sanitizeId"
import Html from "./Html"
import Meta from "./Meta"
import Head from "@/pages/_head"

export default function Layout({ children, data, manifest, meta, statusCode, fieldsConfig = {} }) {
  // Check if we're in production based on manifest presence
  const isProd = !!manifest
  
  // In production, use manifest paths
  // In development, use Vite's dev server paths
  const clientScript = isProd 
    ? `/${manifest['virtual:blastra/entry-client.jsx'].file}`
    : '/@id/virtual:blastra/entry-client.jsx'
  const clientCss = isProd && manifest['virtual:blastra/entry-client.jsx'].css?.[0]
    ? `/${manifest['virtual:blastra/entry-client.jsx'].css[0]}`
    : null

  // Get hydration configuration from fieldsConfig
  const hydrateData = {}
  const fieldHydrators = {}

  const defaultHydrateId = (key) => `hydrate-${key}`

  // Process data based on fieldsConfig
  Object.entries(data).forEach(([key, value]) => {
    const config = fieldsConfig[key] || {}
    const  { hydrateFunction } = config
    if (hydrateFunction){
      // serialize the function
      fieldHydrators[key] = {func: hydrateFunction.toString()}
      return 
    }

    let { hydrateSource, hydrateId, hydrateAttr } = config
    if(!hydrateSource){
      hydrateSource = hydrateId || hydrateAttr ? (hydrateAttr ? 'attr' : 'innerHTML') : 'JSON'
    }
    if(!hydrateId){
      hydrateId = defaultHydrateId
    }
    switch(hydrateSource){
      case 'JSON':
        hydrateData[key] = value
        break
      case 'innerHTML':
      case 'attr':
        if(!hydrateId){
          hydrateId = defaultHydrateId
        }
        const id = sanitizeId(typeof hydrateId === 'function' ? hydrateId(key) : hydrateId)
        fieldHydrators[key] = {id}
        if(hydrateSource === 'attr'){
          fieldHydrators[key].attr = hydrateAttr
        }
        break
    }
  })

  return (
    <Html>
      <Head>
        <meta charSet="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <Meta meta={meta} />
        {clientCss && <link rel="stylesheet" href={clientCss} />}
      </Head>
      <body>
        <div id="app">
          {children}
        </div>
        <script
          dangerouslySetInnerHTML={{
            __html: `window.__BLASTRA_HYDRATION__ = ${JSON.stringify({data: hydrateData, fieldHydrators, statusCode})};`
          }}
        />
        {!isProd && <script type="module" src="/@vite/client" />}
        <script type="module" src={clientScript}></script>
      </body>
    </Html>
  )
}
