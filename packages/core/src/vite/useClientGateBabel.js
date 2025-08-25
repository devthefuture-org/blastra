import { transformSync } from "@babel/core"
import * as t from "@babel/types"

const VIRTUAL_ID = "virtual:use-client-gate"
const VIRTUAL_RESOLVED = "\0" + VIRTUAL_ID

const VIRTUAL_SOURCE = `
  import React, { useEffect, useState } from 'react'
  export function __withClientGate(Comp, Fallback){
    function Wrapped(props){
      const [m, setM] = useState(false)
      useEffect(() => setM(true), [])
      return m
        ? React.createElement(Comp, props)
        : React.createElement('span', { 'data-clientonly': '', suppressHydrationWarning: true },
            Fallback ? React.createElement(Fallback, null) : null
          )
    }
    Wrapped.displayName = (Comp && (Comp.displayName || Comp.name))
      ? 'ClientGate(' + (Comp.displayName || Comp.name) + ')'
      : 'ClientGate'
    return Wrapped
  }
`

function useClientBabelPlugin() {
  return {
    name: "babel-use-client-gate",
    visitor: {
      Program(programPath) {
        const hasUseClient = (programPath.node.directives || []).some(
          (d) => d.value && d.value.value === "use client"
        )
        if (!hasUseClient) return

        // 1) Ensure helper import
        let hasHelperImport = false
        for (const node of programPath.node.body) {
          if (t.isImportDeclaration(node) && node.source.value === VIRTUAL_ID) {
            hasHelperImport = node.specifiers.some(
              (s) =>
                t.isImportSpecifier(s) && t.isIdentifier(s.imported, { name: "__withClientGate" })
            )
          }
        }
        if (!hasHelperImport) {
          const importDecl = t.importDeclaration(
            [t.importSpecifier(t.identifier("__WCG"), t.identifier("__withClientGate"))],
            t.stringLiteral(VIRTUAL_ID)
          )
          programPath.node.body.unshift(importDecl)
        }

        // 2) Wrap default export
        let replaced = false
        programPath.get("body").forEach((stmt) => {
          if (replaced || !stmt.isExportDefaultDeclaration()) return
          const def = stmt.node.declaration

          if (t.isIdentifier(def) || t.isCallExpression(def) || t.isArrowFunctionExpression(def)) {
            stmt.replaceWith(
              t.exportDefaultDeclaration(t.callExpression(t.identifier("__WCG"), [def]))
            )
            replaced = true
            return
          }

          if (t.isFunctionDeclaration(def) || t.isClassDeclaration(def)) {
            const name = def.id ? def.id.name : "__UseClientDefault"
            const id = t.identifier(name)
            const decl = def.id
              ? def
              : t.isFunctionDeclaration(def)
                ? t.functionDeclaration(id, def.params, def.body, def.generator, def.async)
                : t.classDeclaration(id, def.superClass, def.body, def.decorators || [])
            stmt.replaceWithMultiple([
              decl,
              t.exportDefaultDeclaration(t.callExpression(t.identifier("__WCG"), [id])),
            ])
            replaced = true
          }
        })
      },
    },
  }
}

export function useClientGateBabel() {
  return {
    name: "blastra-use-client-gate-babel",
    enforce: "pre",

    resolveId(id) {
      if (id === VIRTUAL_ID) return VIRTUAL_RESOLVED
    },

    load(id) {
      if (id === VIRTUAL_RESOLVED) return VIRTUAL_SOURCE
    },

    transform(code, id) {
      if (!/\.(t|j)sx?$/.test(id)) return
      if (id.includes("node_modules")) return

      const result = transformSync(code, {
        filename: id,
        sourceMaps: true,
        inputSourceMap: this.getCombinedSourcemap?.(),
        plugins: [useClientBabelPlugin()],
        parserOpts: {
          sourceType: "module",
          plugins: [
            "jsx",
            "typescript",
            "importAssertions",
            "classProperties",
            "classPrivateProperties",
            "classPrivateMethods",
            "decorators-legacy",
            "topLevelAwait",
          ],
        },
        generatorOpts: { sourceMaps: true, decoratorsBeforeExport: true },
      })

      if (!result || result.code === code) return null
      return { code: result.code, map: result.map }
    },
  }
}

// (optional) for tests
export function __rawUseClientBabel() {
  return useClientBabelPlugin()
}
