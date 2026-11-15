// **Löschen und Entfernen meinen nicht dasselbe** — UX-10.AK-04
// (#1682):
// *„Löschen und Entfernen aus einer Ansicht werden nicht sprachlich
// verwechselt."*
//
// # Die Regel, und sie ist gemessen und nicht gesetzt
//
// Nachgezählt am 2026-09-21 über alle fünf Kataloge: **55 Texte** nennen
// eines der beiden Wörter, und sie trennen sauber:
//
//   - **Löschen** — die Sache ist weg, und keine andere Stelle hat sie noch.
//     *Deck löschen*, *Karte löschen*, *Turnier löschen*, *Nutzer löschen*,
//     *Verlauf löschen*.
//   - **Entfernen** — die Sache verlässt eine Ansicht, eine Liste, einen
//     Entwurf. Sie gibt es weiter. *Bild entfernen*, *Abzeichen entfernen*,
//     *Zeile entfernen*, *Eigenes Bild entfernen*.
//
// # Was dieser Wächter hält
//
// **Dass Schlüssel und Wort einander nicht widersprechen.** Ein Eintrag
// namens `zugangLoeschen`, der „Entfernen" anzeigt, ist genau die
// Verwechslung, die AK-04 nennt — nur dass sie zwischen Quelltext und
// Bildschirm liegt: wer den Code liest, sieht ein Löschen, wer die Anwendung
// bedient, ein Entfernen.
//
// **Und er verlangt nicht das Wort selbst.** Ein erster Anlauf tat das und
// meldete „Verlauf löschen" als *Clear history*, „Token löschen" als
// *Zurückziehen* und „Marke entfernen" als *abnehmen* — drei richtige
// Wendungen, dreimal rot. Eine Regel, die Synonyme verbietet, verbietet die
// Sprache; gemeint ist der **Widerspruch**.
//
// # Und was er ausdrücklich **nicht** hält
//
// **Ob das Wort zur Handlung passt.** Ein Knopf, der Dateien von der Platte
// nimmt und „Entfernen" heißt, hat einen Schlüssel, der ebenfalls
// „entfernen" sagt — dieser Lauf wäre grün. Genau so ein Fall stand bis zum
// 2026-09-21 bei den Paketen (`Paketverwalter.Entfernen` räumt das
// Verzeichnis, der Go-Kommentar daneben sagt „vergleicht und **löscht**").
// Das sieht nur ein Mensch, und deshalb steht es hier (§10).
import { describe, expect, it } from 'vitest'

import { kataloge, sprachen } from './sprache/kataloge'

// woerter sagt, woran das **andere** Wort zu erkennen ist — je Sprache.
//
// **Stämme und keine ganzen Wörter**: „entfernen", „entfernt", „Entfernung"
// sind dieselbe Sache, und „löschen" steht auch als „gelöscht" da.
const woerter = {
  de: { loesch: 'lösch', entfern: 'entfern' },
  en: { loesch: 'delet', entfern: 'remov' },
} as const

// texte sammelt jeden Katalogeintrag mit seinem Pfad.
function texte(knoten: unknown, pfad: string[] = []): [string, string][] {
  if (typeof knoten === 'string') {
    return [[pfad.join('.'), knoten]]
  }

  if (typeof knoten !== 'object' || knoten === null) {
    return []
  }

  return Object.entries(knoten).flatMap(([schluessel, wert]) =>
    texte(wert, [...pfad, schluessel]),
  )
}

// **Nur, wo die Wortstämme gepflegt sind** — Deutsch und Englisch. Die sechs
// maschinell übersetzten Sprachen (LOC-01, #1687) prüft dieser Wächter
// nicht; ihre Stämme zu erraten wäre eine Liste, die niemand geprüft hat.
const gepruefte = sprachen.filter(
  (sprache): sprache is keyof typeof woerter => sprache in woerter,
)

describe('Löschen und Entfernen', () => {
  it.each(gepruefte)('sagen in %s dasselbe wie ihr Schlüssel', (sprache) => {
    const erwartet = woerter[sprache]
    const befunde: string[] = []

    for (const [pfad, text] of texte(kataloge[sprache])) {
      const schluessel = pfad.split('.').at(-1)?.toLowerCase() ?? ''
      const klein = text.toLowerCase()

      // **Der Schlüssel muss auf das Wort enden**, und zwar genau.
      //
      // # Warum nicht „enthält"
      //
      // **Weil ein Schlüssel auch den Dialog benennen kann und nicht den
      // Knopf.** `loeschenNein` ist der Abbrechen-Knopf **im** Löschdialog,
      // `loeschenFrage` die Rückfrage, `formatLoeschenHinweis` der Satz
      // darunter. Ein erster Anlauf prüfte auf „enthält" und meldete
      // **zwölf** Befunde, von denen neun keine waren — ein Wächter, der so
      // meldet, wird abgeschaltet und nicht gelesen.
      //
      // **Wer auf das Wort endet, benennt die Handlung selbst**:
      // `zugangLoeschen`, `bildEntfernen`, `verlaufLoeschen`,
      // `positionEntfernen`.
      if (schluessel.endsWith('loeschen') && klein.includes(erwartet.entfern)) {
        befunde.push(
          `${pfad}: „${text}" — der Schlüssel sagt löschen, das Wort entfernen`,
        )
      }

      if (schluessel.endsWith('entfernen') && klein.includes(erwartet.loesch)) {
        befunde.push(
          `${pfad}: „${text}" — der Schlüssel sagt entfernen, das Wort löschen`,
        )
      }
    }

    expect(befunde).toEqual([])
  })
})
