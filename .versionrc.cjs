module.exports = {
  bumpFiles: [
    { filename:"package.json" },
    { filename:"packages/core/package.json" },
    { filename:"packages/create/package.json" },
    { filename:"packages/create/template/package.json", updater: "scripts/update-template-version.cjs" },
  ],
}