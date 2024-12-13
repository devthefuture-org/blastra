module.exports = {
  readVersion: (contents) => {
    const pkg = JSON.parse(contents)
    return pkg.dependencies["@blastra/core"]
  },
  writeVersion: (contents, version) => {
    const pkg = JSON.parse(contents)
    pkg.dependencies["@blastra/core"] = `^${version}`
    return JSON.stringify(pkg, null, 2)
  },
}
