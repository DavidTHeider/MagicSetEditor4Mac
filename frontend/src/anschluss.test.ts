// **Was ein Anschluss zusagt** — UX-03.AK-01
// (#1675).
import { readdirSync, readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

import { anschlussarten, anschlussseite, setgebunden } from './anschluss'
import { anwendungsschluessel, istSetseite } from './navigation'

// quellen sind alle Dateien unter `src` ohne Tests — `Vitest` läuft in
// `frontend/`, wie `bausteineohneaufrufer.test.ts` es auch annimmt.
function quellen(ort = 'src'): string[] {
  return readdirSync(ort, { withFileTypes: true }).flatMap((eintrag) => {
    const pfad = join(ort, eintrag.name)

    if (eintrag.isDirectory()) return quellen(pfad)

    return /\.tsx?$/u.test(eintrag.name) &&
      !/\.test\.tsx?$/u.test(eintrag.name) &&
      !pfad.includes('/test/')
      ? [pfad]
      : []
  })
}

describe('wohin ein Anschluss führt', () => {
  // **Der Übersetzer hält nur, dass es eine Seite ist** — nicht, dass sie
  // ohne offenes Set erreichbar wäre. Eine Setlasche als Ziel landete ins
  // Leere, sobald kein Set offen ist, und das ist der Normalfall: wer in der
  // Sammlung steht, hat womöglich nie eines geöffnet.
  //
  // **Außer für eine Art, die ihr Set mitbringt** (UX-02.AK-02, #1674): die
  // Befehlspalette bietet „Karte erstellen" nur an, wenn ein Set offen ist,
  // und die Kartenseite prüft beim Aufnehmen, ob es noch dasselbe ist.
  it('führt jede Art auf eine Anwendungsseite — eine setgebundene auf eine Setlasche', () => {
    for (const art of anschlussarten) {
      const ziel = anschlussseite[art]

      if (setgebunden.includes(art)) {
        expect(istSetseite(ziel)).toBe(true)
      } else {
        expect(istSetseite(ziel)).toBe(false)
        expect(anwendungsschluessel).toContain(ziel)
      }
    }
  })

  // **Jede Art wird aufgenommen, und zwar auf ihrer Fläche.** Hier stand bis
  // zum 2026-09-24 „gibt jeder Art eine eigene Fläche": zwei Arten auf
  // dieselbe Seite, hieß es, ließen die Fläche nicht wissen, welche Felder
  // gefüllt sind. **Das hält der unterscheidende Verbund**, nicht die Seite
  // — die Fläche verengt an `art`, bevor sie liest. Seit UX-02.AK-02 führen
  // `deck` (ein Druck will hinein) und `deckansicht` (ein Deck soll auf)
  // beide zu den Decks. Was die alte Zusage eigentlich schützte, prüft
  // dieser Lauf statt ihrer: **eine Art, die niemand aufnimmt, ist ein
  // Anschluss ins Leere.**
  it('nimmt jede Art irgendwo auf', () => {
    const texte = quellen().map((pfad) => readFileSync(pfad, 'utf8'))

    const ohneAufnahme = anschlussarten.filter(
      (art) =>
        !texte.some((text) =>
          text.includes(`anschluss.offen?.art === '${art}'`),
        ),
    )

    expect(ohneAufnahme).toEqual([])
  })
})
