import { useState, useEffect, useRef } from "react"
import { useLocation } from "wouter"

import { MetaUpdater } from "./MetaUpdater"
import log from "../utils/clientLogger"

import Error4xx from "@/pages/_4xx"
import Error5xx from "@/pages/_5xx"
import Error404 from "@/pages/_404"

export default function PageWrapper({ initialData, router, statusCode: initalStatusCode }) {
  const { getComponent, getMeta, getLoader } = router

  const [location] = useLocation()
  const [data, setData] = useState(initialData)
  const [meta, setMeta] = useState(null)
  const [loading, setLoading] = useState(false)
  const initialRenderDone = useRef(false)
  const lastLocation = useRef(location)
  const [statusCode, setStatusCode] = useState(initalStatusCode)

  useEffect(() => {
    // Skip data fetch only on the very first render with SSR data
    if (!initialRenderDone.current) {
      log("Initial render with SSR data:", initialData)
      // Get initial meta
      const metaFn = getMeta(location)
      const initialMeta = metaFn(initialData)
      setMeta(initialMeta)
      initialRenderDone.current = true
      lastLocation.current = location
      return
    }

    // If location changed, fetch new data
    if (location !== lastLocation.current) {
      setLoading(true)
      const path = location.startsWith("/") ? location : `/${location}`
      log("Fetching data for path:", path)

      const loader = getLoader(path)
      loader()
        .then((newData) => {
          log("Received new data:", newData)
          setData(newData)
          // Update meta with new data
          const metaFn = getMeta(location)
          const newMeta = metaFn(newData)
          setMeta(newMeta)
          setLoading(false)
          setStatusCode(200)
          lastLocation.current = location
        })
        .catch((err) => {
          // try to get status code from fetch error
          const code = err?.response?.status || 500
          console.error("Error loading data:", err)
          setData({ error: "Failed to load data" })
          setMeta({ title: "Error" })
          setLoading(false)
          setStatusCode(code)
          lastLocation.current = location
        })
    }
  }, [location, initialData, getLoader])

  if (loading) {
    return <div>Loading...</div>
  }

  let Component
  if (statusCode >= 500) {
    Component = Error5xx
  } else if (statusCode >= 400) {
    Component = Error4xx
  } else {
    Component = getComponent(location) || Error404
  }

  return (
    <>
      <MetaUpdater meta={meta} />
      <Component data={data} statusCode={statusCode} />
    </>
  )
}
