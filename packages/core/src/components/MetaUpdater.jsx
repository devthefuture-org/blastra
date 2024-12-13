import { useEffect } from 'react'

export function MetaUpdater({ meta }) {
  useEffect(() => {
    if (!meta) return

    // Update title
    if (meta.title) {
      document.title = meta.title
    }

    // Update meta tags
    if (meta.meta) {
      // Remove existing meta tags we might have added
      const existingMeta = document.head.querySelectorAll('meta[data-dynamic="true"]')
      existingMeta.forEach(tag => tag.remove())

      // Add new meta tags
      meta.meta.forEach(({ name, content }) => {
        const metaTag = document.createElement('meta')
        metaTag.setAttribute('name', name)
        metaTag.setAttribute('content', content)
        metaTag.setAttribute('data-dynamic', 'true')
        document.head.appendChild(metaTag)
      })
    }

    // Update canonical link
    if (meta.canonical !== undefined) {
      // Remove existing canonical link if present
      const existingCanonical = document.head.querySelector('link[rel="canonical"][data-dynamic="true"]')
      if (existingCanonical) {
        existingCanonical.remove()
      }

      // Add new canonical link if value provided
      if (meta.canonical) {
        const canonicalLink = document.createElement('link')
        canonicalLink.setAttribute('rel', 'canonical')
        canonicalLink.setAttribute('href', meta.canonical)
        canonicalLink.setAttribute('data-dynamic', 'true')
        document.head.appendChild(canonicalLink)
      }
    }

    return () => {
      // Cleanup dynamic meta tags and canonical on unmount
      const dynamicElements = document.head.querySelectorAll('[data-dynamic="true"]')
      dynamicElements.forEach(el => el.remove())
    }
  }, [meta])

  return null
}
