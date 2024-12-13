export function meta(data) {
  return {
    title: data.title,
    meta: [
      {
        name: "description",
        content: data.description,
      },
      {
        name: "og:title",
        content: data.title,
      },
      {
        name: "og:description",
        content: data.description,
      },
    ],
  }
}
