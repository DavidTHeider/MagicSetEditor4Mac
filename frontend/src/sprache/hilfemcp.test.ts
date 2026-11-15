// **Was die Hilfeseite über `mse-mcp` sagt, stimmt** —
// #912.
//
// # Der Anlass
//
// Sie sagte bis zum 2026-08-23, ein Sprachmodell könne „Karten lesen und
// **einzelne Felder ändern**". Seit
// #621 kann es
// mehr: anlegen, kopieren, löschen, tauschen.
//
// Derselbe Fehler wie in der README
// (#900) — nur
// steht er hier **in der ausgelieferten Anwendung**, auf der Seite, die gerade
// sagen soll, was nach außen möglich ist.
//
// # Was dieser Wächter hält, und was nicht
//
// Er verbietet die **Untertreibung**, die dagewesen ist, und verlangt, dass der
// Lebenszyklus vorkommt. Er hält **nicht**, dass der Satz im Übrigen stimmt:
// eine Prosa-Zeile lässt sich nicht gegen `internal/mcp` rechnen, und die
// Hilfeseite soll auch keine Werkzeugliste sein — sie erklärt, sie zählt nicht
// auf.
//
// **Die Liste der Werkzeuge hält dagegen ein echter Wächter**:
// `internal/werkzeuge/readmewerkzeuge_test.go` gegen `mcp.Werkzeuge`.
import { describe, expect, it } from 'vitest'

import { kataloge } from './kataloge'

// untertreibungen sind die Wendungen, mit denen die Seite zu wenig sagte.
const untertreibungen: Record<string, string[]> = {
  de: ['einzelne Felder ändern'],
  en: ['change single fields'],
}

// lebenszyklus ist das, was seit 0040 dazugehört und vorkommen muss.
const lebenszyklus: Record<string, string[]> = {
  de: ['anlegen', 'löschen'],
  en: ['create', 'delete'],
}

describe('der Hilfetext über den MCP-Server', () => {
  for (const sprache of ['de', 'en'] as const) {
    describe(sprache, () => {
      const satz = kataloge[sprache].hilfe.schnittstellen.mcpSatz

      it('untertreibt nicht mehr', () => {
        for (const wendung of untertreibungen[sprache] ?? []) {
          expect(satz).not.toContain(wendung)
        }
      })

      it('nennt den Lebenszyklus einer Karte', () => {
        for (const wort of lebenszyklus[sprache] ?? []) {
          expect(satz.toLowerCase()).toContain(wort)
        }
      })
    })
  }
})
