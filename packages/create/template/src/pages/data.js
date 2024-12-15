export const fieldsConfig = {
  content: {
    hydrateFunction: () => document.getElementById("home-content").innerHTML,
  },
}

export async function loader() {
  return {
    title: "Welcome to Blastra Template",
    description: "This is the home page",
    content: `"When you let go of what you are, you become what you might be." Lao Tzu`,
  }
}
