import ReactDOMServer from 'react-dom/server'
import App from './components/App'
import Layout from './components/Layout'
import { logger } from './utils/logger.js'

export function createRender({ router }){
  return async function render({ url, manifest }) {
    const { getLoader, getMeta, getFieldsConfig, findRoute } = router

    try {
      // Get the route and its loader and meta functions
      const loader = getLoader(url)
      const fieldsConfig = getFieldsConfig(url)
      const metaVar = getMeta(url)
      
      // Load data
      const data = await loader()
      logger.debug('SSR data loaded:', data)
      
      // Get metadata
      const meta = typeof metaVar === 'function' ? metaVar(data) : metaVar
      logger.debug('Meta generated:', meta)
      
      const statusCode = findRoute(url) ? 200 : 404
      
      logger.debug('App rendering with url:', url, 'initialData:', data)
      
      const html = ReactDOMServer.renderToString(
        <Layout data={data} manifest={manifest} meta={meta} fieldsConfig={fieldsConfig}>
          <App router={router} url={url} initialData={data} statusCode={statusCode} />
        </Layout>
      )
      
      logger.success('SSR completed for:', url)
      return { html, data, meta, statusCode }
    } catch (e) {
      logger.error('SSR error:', e)
      const data = { error: 'Failed to load data' }
      const meta = { title: 'Error' }
      const statusCode = 500
      return {
        html: ReactDOMServer.renderToString(
          <Layout data={data} manifest={manifest} meta={meta} statusCode={statusCode}>
            <App router={router} url={url} initialData={data} statusCode={statusCode} />
          </Layout>
        ),
        data,
        meta,
        statusCode,
      }
    }
  }
}
