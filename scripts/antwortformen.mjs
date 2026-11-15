// **Die Antwortformen aller Bindungen, aus den Typen des Fensters** — C01
// (#2226).
//
//     node scripts/antwortformen.mjs > formen.json
//
// Liest `Api` aus `frontend/src/api.ts` mit dem TypeScript-Compiler des
// Projekts (`frontend/node_modules/typescript`, gepinnt in der Sperrdatei)
// und gibt je Methode die Form ihres aufgelösten Rückgabewerts
// (`Promise<T>` → `T`) — die Formen, die `scripts/antwortform.mjs` prüft.
//
// **Aus den Typen des Fensters und nicht aus den Go-Strukturen**: geprüft
// werden soll, ob das, was über die Brücke kommt, das ist, was das Fenster
// erwartet — die Naht zwischen zwei Prüfebenen (§3).
import { createRequire } from 'node:module'
import { join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const wurzel = resolve(fileURLToPath(import.meta.url), '../..')
const ts = createRequire(import.meta.url)(join(wurzel, 'frontend/node_modules/typescript'))

// formAus wandelt einen Typ in eine Form — rekursiv, mit einer Grenze gegen
// Zyklen (ein Typ, der sich selbst enthält, wird an der zweiten Stelle
// `unbekannt`).
export function formAus(pruefer, typ, offen = new Set()) {
  if (typ.flags & ts.TypeFlags.Any || typ.flags & ts.TypeFlags.Unknown) return { art: 'unbekannt' }
  if (typ.flags & ts.TypeFlags.Null) return { art: 'null' }
  if (typ.flags & ts.TypeFlags.Undefined || typ.flags & ts.TypeFlags.Void) return { art: 'null' }
  if (typ.isStringLiteral()) return { art: 'literal', wert: typ.value }
  if (typ.isNumberLiteral()) return { art: 'literal', wert: typ.value }
  if (typ.flags & ts.TypeFlags.BooleanLiteral) return { art: 'literal', wert: pruefer.typeToString(typ) === 'true' }
  if (typ.flags & ts.TypeFlags.StringLike) return { art: 'string' }
  if (typ.flags & ts.TypeFlags.NumberLike) return { art: 'number' }
  if (typ.flags & ts.TypeFlags.BooleanLike) return { art: 'boolean' }
  if (typ.isUnion()) {
    const teile = typ.types
    // **boolean ist true | false** — als eine Art, nicht als zwei Literale.
    const literalBool = teile.filter((t) => t.flags & ts.TypeFlags.BooleanLiteral)
    const rest = teile.filter((t) => !(t.flags & ts.TypeFlags.BooleanLiteral))
    const von = [...(literalBool.length === 2 ? [{ art: 'boolean' }] : literalBool.map((t) => formAus(pruefer, t, offen))), ...rest.map((t) => formAus(pruefer, t, offen))]
    return von.length === 1 ? von[0] : { art: 'oder', von }
  }
  if (pruefer.isArrayType(typ) || pruefer.isTupleType(typ)) {
    const [element] = pruefer.getTypeArguments(typ)
    return { art: 'liste', von: element ? formAus(pruefer, element, offen) : { art: 'unbekannt' } }
  }
  if (typ.flags & ts.TypeFlags.Object) {
    if (offen.has(typ)) return { art: 'unbekannt' }
    const naechste = new Set(offen).add(typ)
    const index = pruefer.getIndexInfosOfType(typ)
    const eigenschaften = pruefer.getPropertiesOfType(typ)
    if (index.length > 0 && eigenschaften.length === 0) return { art: 'abbildung', von: formAus(pruefer, index[0].type, naechste) }
    const felder = {}
    for (const eigenschaft of eigenschaften) {
      const feldtyp = pruefer.getTypeOfSymbol(eigenschaft)
      if (feldtyp.getCallSignatures().length > 0) continue
      felder[eigenschaft.getName()] = {
        form: formAus(pruefer, feldtyp, naechste),
        optional: (eigenschaft.flags & ts.SymbolFlags.Optional) !== 0,
      }
    }
    return { art: 'objekt', felder }
  }

  return { art: 'unbekannt' }
}

// antwortformen gibt je Methode von `Api` die Form ihrer Antwort.
export function antwortformen() {
  const konfiguration = ts.getParsedCommandLineOfConfigFile(join(wurzel, 'frontend/tsconfig.json'), {}, { ...ts.sys, onUnRecoverableConfigFileDiagnostic: () => {} })
    ?? { options: { strict: true, target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler, jsx: ts.JsxEmit.ReactJSX } }
  const programm = ts.createProgram([join(wurzel, 'frontend/src/api.ts')], { ...konfiguration.options, noEmit: true })
  const pruefer = programm.getTypeChecker()
  const quelle = programm.getSourceFile(join(wurzel, 'frontend/src/api.ts'))
  let api
  ts.forEachChild(quelle, (knoten) => {
    if (ts.isInterfaceDeclaration(knoten) && knoten.name.text === 'Api') api = pruefer.getTypeAtLocation(knoten)
  })
  if (!api) throw new Error('keine Schnittstelle Api in frontend/src/api.ts')
  const formen = {}
  for (const methode of pruefer.getPropertiesOfType(api)) {
    const signatur = pruefer.getTypeOfSymbol(methode).getCallSignatures()[0]
    if (!signatur) continue
    let rueck = signatur.getReturnType()
    const promise = rueck.getSymbol()?.getName() === 'Promise' ? pruefer.getTypeArguments(rueck)[0] : undefined
    if (promise) rueck = promise
    formen[methode.getName()] = formAus(pruefer, rueck)
  }

  return formen
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  process.stdout.write(`${JSON.stringify(antwortformen())}\n`)
}

