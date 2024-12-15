export default function sanitizeId(id) {
  // Replace invalid characters with hyphens and ensure valid HTML5 id
  return id
    .replace(/[^a-zA-Z0-9-_:]/g, "-") // Replace invalid chars with hyphen
    .replace(/^[^a-zA-Z]+/, "id-") // Ensure starts with letter
    .replace(/^-+|-+$/g, "") // Remove leading/trailing hyphens
    .toLowerCase() // Convert to lowercase for consistency
}
