export function createFileBasedRouter({ pageModules, dataModules, metaModules }) {

  const routes = generateRoutes()
  
  function extractParams(pattern, path) {
    if (!pattern || !path) return null
    
    const paramNames = (pattern.match(/:[^/]+/g) || []).map(p => p.slice(1))
    const regexPattern = pattern.replace(/:[^/]+/g, '([^/]+)')
    const regex = new RegExp(`^${regexPattern}$`)
    const matches = path.match(regex)
    
    if (!matches) return null
    
    const params = {}
    paramNames.forEach((name, index) => {
      params[name] = matches[index + 1]
    })
    return params
  }
  
  function generateRoutes() {
    const routes = []
    
    for (const [path, component] of Object.entries(pageModules)) {
      // Convert file path to route path
      // remove /src/pages prefix using regex
      let routePath = path
        .replace(/^\/src\/pages/, '')
        .replace(/\/index\.jsx$/, '')
        .replace(/\[([^\]]+)\]/g, ':$1') // Convert [param] to :param
        || '/'
      
      // Get corresponding data and meta modules
      const dataPath = path.replace('index.jsx', 'data.js')
      const metaPath = path.replace('index.jsx', 'meta.js')
      const data = dataModules[dataPath] || {}
      const meta = metaModules[metaPath]?.meta
      
      routes.push({
        path: routePath,
        component: component.default,
        loader: data.loader,
        meta,
        fieldsConfig: data.fieldsConfig
      })
    }
    
    return routes.sort((a, b) => {
      // Sort routes to ensure more specific routes come first
      const aSegments = a.path.split('/').length
      const bSegments = b.path.split('/').length
      if (aSegments !== bSegments) return bSegments - aSegments
      return a.path.includes(':') ? 1 : -1
    })
  }
  
  function findRoute(path) {
    if (!path) return null
    
    return routes.find(route => {
      if (route.path.includes(':')) {
        const params = extractParams(route.path, path)
        return params !== null
      }
      return route.path === path // Case sensitive matching
    })
  }
  
  function getParams(path) {
    const route = findRoute(path)
    if (!route || !route.path.includes(':')) return {}
    return extractParams(route.path, path) || {}
  }
  
  function getLoader(path) {
    const route = findRoute(path)
    if (!route?.loader) return async () => ({})
    
    const params = getParams(path)
    return () => route.loader({ params })
  }
  
  function getMeta(path) {
    const route = findRoute(path)
    const routeMeta = route?.meta || (() => ({}))

    return (...args) => {
      // Get meta from route
      const meta = routeMeta(...args)

      // Get default canonical URL if not explicitly set
      if (meta.canonical === undefined) {
        let canonicalUrl
        // Remove query parameters by only using pathname
        const cleanPath = path.split('?')[0]
        if (typeof window !== 'undefined') {
          // Client-side: use window.location.origin + pathname
          canonicalUrl = `${window.location.origin}${cleanPath}`
        } else {
          // Server-side: construct from host
          const host = process.env.SITE_URL || `http://localhost:${process.env.PORT || 5173}`
          canonicalUrl = `${host}${cleanPath}`
        }
        meta.canonical = canonicalUrl
      }

      return meta
    }
  }
  
  function getFieldsConfig(path) {
    const route = findRoute(path)
    return route?.fieldsConfig || {}
  }
  
  function getComponent(path) {
    const route = findRoute(path)
    return route?.component
  }

  function getRoutes() {
    return routes
  }

  return {
    getLoader,
    getMeta,
    getFieldsConfig,
    getComponent,
    getRoutes,
    findRoute,
  }
}
