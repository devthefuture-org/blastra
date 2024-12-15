export default function Meta({ meta }) {
  if (!meta) return null

  const { title, canonical, meta: metaTags } = meta

  return (
    <>
      {title && <title>{title}</title>}
      {canonical && <link rel="canonical" href={canonical} data-dynamic="true" />}
      {metaTags?.map((metaItem, index) => (
        <meta key={index} {...metaItem} data-dynamic="true" />
      ))}
    </>
  )
}
