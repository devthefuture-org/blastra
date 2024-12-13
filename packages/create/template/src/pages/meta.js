export function meta(data) {
  return {
    title: data.title,
    meta: [
      {
        name: "description",
        content: data.description,
      },
    ],
  }
}
