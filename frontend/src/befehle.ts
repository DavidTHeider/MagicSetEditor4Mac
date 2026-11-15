// **Parameter aus dem Schema, Argumente aus den Werten, Verlauf ohne
// Geheimnisse** — Werkstück U12 von Phase 5.1 (#1423). Ohne DOM prüfbar.
import type { Schema } from './apibefehle'

export type Parameterart = 'text' | 'zahl' | 'wahrheit' | 'wahl' | 'json'

export interface Parameter {
  name: string
  art: Parameterart
  pflicht: boolean
  erklaerung: string
  werte: string[]
}

interface Eigenschaft {
  type?: string | string[]
  description?: string
  enum?: unknown[]
}

function artVon(eigenschaft: Eigenschaft): Parameterart {
  if (Array.isArray(eigenschaft.enum)) {
    return 'wahl'
  }

  const typ = Array.isArray(eigenschaft.type)
    ? eigenschaft.type[0]
    : eigenschaft.type

  switch (typ) {
    case 'number':
    case 'integer':
      return 'zahl'
    case 'boolean':
      return 'wahrheit'
    case 'string':
    case undefined:
      return 'text'
    default:
      return 'json'
  }
}

// parameterAus liest die Eigenschaften eines Objektschemas — in der
// Reihenfolge, in der das Werkzeug sie nennt.
export function parameterAus(schema: Schema): Parameter[] {
  const eigenschaften = schema.properties
  if (typeof eigenschaften !== 'object' || eigenschaften === null) {
    return []
  }

  const pflicht = new Set(
    Array.isArray(schema.required)
      ? schema.required.filter((n): n is string => typeof n === 'string')
      : [],
  )

  return Object.entries(eigenschaften as Record<string, Eigenschaft>).map(
    ([name, eigenschaft]) => ({
      name,
      art: artVon(eigenschaft),
      pflicht: pflicht.has(name),
      erklaerung: eigenschaft.description ?? '',
      werte: Array.isArray(eigenschaft.enum)
        ? eigenschaft.enum.map((wert) => String(wert))
        : [],
    }),
  )
}

export type Werte = Record<string, string | boolean>

export interface Argumentbefund {
  name: string
  grund: 'pflicht' | 'zahl' | 'json'
}

// argumenteAus baut das JSON-Objekt, das an das Werkzeug geht — oder nennt,
// welches Feld noch nicht taugt. **Kein Feld wird erraten**: ein leeres
// Textfeld fehlt, es ist nicht die leere Zeichenkette.
export function argumenteAus(
  parameter: Parameter[],
  werte: Werte,
): { argumente: Record<string, unknown>; befunde: Argumentbefund[] } {
  const argumente: Record<string, unknown> = {}
  const befunde: Argumentbefund[] = []

  for (const einer of parameter) {
    const wert = werte[einer.name]

    if (einer.art === 'wahrheit') {
      if (wert === true) {
        argumente[einer.name] = true
      } else if (einer.pflicht) {
        argumente[einer.name] = false
      }
      continue
    }

    const text = typeof wert === 'string' ? wert.trim() : ''
    if (text === '') {
      if (einer.pflicht) {
        befunde.push({ name: einer.name, grund: 'pflicht' })
      }
      continue
    }

    if (einer.art === 'zahl') {
      const zahl = Number(text)
      if (Number.isNaN(zahl)) {
        befunde.push({ name: einer.name, grund: 'zahl' })
      } else {
        argumente[einer.name] = zahl
      }
      continue
    }

    if (einer.art === 'json') {
      try {
        argumente[einer.name] = JSON.parse(text) as unknown
      } catch {
        befunde.push({ name: einer.name, grund: 'json' })
      }
      continue
    }

    argumente[einer.name] = text
  }

  return { argumente, befunde }
}

// geheim sagt, ob ein Parametername ein Geheimnis trägt — der Verlauf zeigt
// seinen Wert dann nicht.
export function geheim(name: string): boolean {
  return /token|kennwort|passwort|geheimnis|secret|password/i.test(name)
}

// bereinigt ersetzt jeden geheimen Wert im Argumentobjekt — für den Verlauf.
export function bereinigt(
  argumente: Record<string, unknown>,
): Record<string, unknown> {
  const kopie: Record<string, unknown> = {}
  for (const [name, wert] of Object.entries(argumente)) {
    kopie[name] = geheim(name) ? '•••' : wert
  }

  return kopie
}

// alsBaumtext richtet einen Antworttext als JSON aus, wenn er eines ist.
export function huebsch(text: string): { json: boolean; text: string } {
  try {
    return { json: true, text: JSON.stringify(JSON.parse(text), null, 2) }
  } catch {
    return { json: false, text }
  }
}
