// **Die Anwendung duzt** — Werkstück C von Phase 4.10
// (#1325).
//
// # Warum das ein Wächter ist und keine Verabredung
//
// Der deutsche Katalog duzt an fast jeder Stelle: „Was in **deinem** Bestand
// liegt", „Was **du** wirklich besitzt", „**Tippe** etwas, um Karten zu
// finden". Auf der Symbollasche stand am 2026-09-03 „**Wählen Sie** ein Teil"
// — eine einzelne Stelle, die niemandem auffällt, bis sie jemand liest.
//
// Eine Ansprache ist eine Entscheidung über den Ton der ganzen Anwendung. Wo
// sie an einer Stelle kippt, wirkt sie überall unentschieden.
//
// # Was er nicht hält
//
// **Ob der Satz gut ist.** Er sucht die Höflichkeitsform in ihren häufigen
// Gestalten — Imperativ mit „Sie", und das großgeschriebene Possessiv mitten
// im Satz. Ein Siezen ohne diese Wörter käme durch.
//
// **Und nichts über Englisch.** Dort gibt es die Unterscheidung nicht.
import { expect, it } from 'vitest'

import { kataloge } from './kataloge'

// hoeflich sind die Muster, an denen ein Siezen erkennbar ist.
//
// **`Sie` nur nach einem Verb**: „wie Sie" wäre ebenfalls ein Siezen, aber
// „Sie" am Satzanfang ist meist ein „sie" — und ein Wächter, der bei jedem
// Satzanfang anschlägt, wird abgeschaltet.
//
// **Und `Ihre` nur mitten im Satz.** Am Satzanfang ist es das gewöhnliche
// „ihre": *„Diese Felder kennt das neue Spiel nicht. Ihre Werte bleiben im
// Bestand"* meint die Werte der Felder. Der erste Anlauf dieses Wächters hat
// beide Sätze angestrichen — ein Wächter, der recht zu haben scheint und
// unrecht hat, kostet mehr, als er einbringt.
const hoeflich = [
  /\b[A-ZÄÖÜ][a-zäöüß]+en Sie\b/,
  /[a-zäöüß,;)] Ihre[nmrs]? /,
  /\bIhnen\b/,
]

// alleTexte sammelt jede Zeichenkette eines Katalogs.
function alleTexte(wert: unknown, wohin: string[]): string[] {
  if (typeof wert === 'string') {
    wohin.push(wert)
  } else if (Array.isArray(wert)) {
    for (const eines of wert) {
      alleTexte(eines, wohin)
    }
  } else if (wert !== null && typeof wert === 'object') {
    for (const eines of Object.values(wert)) {
      alleTexte(eines, wohin)
    }
  }

  return wohin
}

it('siezt an keiner Stelle', () => {
  const texte = alleTexte(kataloge.de, [])

  // **Ein leerer Leser wäre grün** — deshalb steht die Zahl hier.
  expect(texte.length).toBeGreaterThan(500)

  const gesiezt = texte.filter((satz) =>
    hoeflich.some((muster) => muster.test(satz)),
  )

  expect(gesiezt).toEqual([])
})
