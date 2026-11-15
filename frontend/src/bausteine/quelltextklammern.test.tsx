// **Die Klammermarken im Feld** — Werkstück B von Phase 4.9
// (#1324).
//
// # Warum eine eigene Datei
//
// `quelltextfeld.test.tsx` ist bei 354 Zeilen, und die Zeilengrenze aus §2
// gilt für Tests ausdrücklich mit. Der Schnitt folgt der Frage: dort steht,
// ob Farbe, Nummer und Fehlermarke am richtigen Ort liegen; hier, ob das
// Klammerpaar unter dem Schreibzeiger erscheint.
//
// # Was `klammerpaare.test.ts` nicht prüft
//
// Die Rechnung ist dort geprüft. **Hier geht es um den Weg dorthin**: dass
// der Schreibzeiger überhaupt ankommt, dass die Marke gezeichnet wird und
// dass sie wieder verschwindet. Zwei grüne Seiten einer Grenze sind kein
// Nachweis für die Grenze.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Quelltextfeld } from './Quelltextfeld'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import type { Merkmalstelle } from '../antworten'

// klammermerkmale belegt die Farbbindung so, wie der Lexer eine einzeilige
// Quelle zerlegte — nur die Klammern, denn nur die zählen hier.
function klammermerkmale(quelle: string) {
  const stellen: Merkmalstelle[] = []

  for (let bei = 0; bei < quelle.length; bei += 1) {
    if ('()[]{}'.includes(quelle.charAt(bei))) {
      stellen.push({ art: 'Operator', zeile: 1, spalte: bei + 1, breite: 1 })
    }
  }

  belegeApi({ Skriptfarben: () => Promise.resolve({ merkmale: stellen }) })
}

// zeigeFeld stellt das Feld mit einer Quelle dar und gibt die Eingabe.
async function zeigeFeld(quelle: string) {
  klammermerkmale(quelle)

  render(
    <Sprachraum sprache="de">
      <Quelltextfeld bezeichnung="Ausdruck" wert={quelle} setzeWert={vi.fn()} />
    </Sprachraum>,
  )

  const feld = await screen.findByLabelText('Ausdruck')

  // **Auf die Farben warten, bevor der Zeiger gesetzt wird.** Ohne Merkmale
  // gibt es keine Klammern zu finden, und der Test prüfte die leere Liste.
  await waitFor(() => {
    expect(screen.getByTestId('quelltextanzeige').textContent).toContain(quelle)
  })

  return feld
}

// zeigeAuf setzt den Schreibzeiger auf eine Stelle.
//
// **Ohne Klick.** Der erste Anlauf setzte die Auswahl und klickte danach —
// und `userEvent.click` setzt den Zeiger unter jsdom **ans Ende**, weil es
// dort kein Layout gibt und niemand weiß, welches Zeichen unter der Maus
// liegt. Alle drei Fälle waren daraufhin grün beziehungsweise rot aus dem
// falschen Grund: gemessen wurde immer dieselbe Stelle 7.
//
// `fireEvent.select` meldet genau das, was das Feld hört (`onSelect`), und
// verschiebt dabei nichts.
function zeigeAuf(feld: HTMLElement, stelle: number) {
  const eingabe = feld as HTMLTextAreaElement

  eingabe.setSelectionRange(stelle, stelle)

  fireEvent.select(eingabe)
}

describe('Die Klammermarken', () => {
  afterEach(raeumeApi)

  it('hebt beide Klammern eines Paares hervor', async () => {
    const feld = await zeigeFeld('f(a, b)')

    zeigeAuf(feld, 2)

    const marken = await screen.findAllByTestId('klammermarke')

    expect(marken).toHaveLength(2)
    expect(marken.map((marke) => marke.textContent)).toEqual(['(', ')'])
  })

  // **Der Fall, für den es die Anzeige gibt.** Eine Klammer ohne Partner ist
  // derselbe Befund, den der Zerteiler später meldet — nur sieht man ihn hier
  // beim Tippen und nicht erst beim Auswerten.
  it('streicht eine Klammer ohne Partner anders an', async () => {
    const feld = await zeigeFeld('f(a, b')

    zeigeAuf(feld, 2)

    const marke = await screen.findByTestId('klammerohnepartner')

    expect(marke.textContent).toBe('(')
    expect(screen.queryByTestId('klammermarke')).toBeNull()
  })

  it('zeigt nichts, wo keine Klammer steht', async () => {
    const feld = await zeigeFeld('f(a, b)')

    // **Erst eine Klammer und dann keine.** Eine Verneinung allein träfe auch
    // zu, wenn der Zeiger nie angekommen wäre (§3) — der erste Griff belegt,
    // dass er ankommt.
    zeigeAuf(feld, 2)
    await screen.findAllByTestId('klammermarke')

    zeigeAuf(feld, 4)

    await waitFor(() => {
      expect(screen.queryAllByTestId('klammermarke')).toHaveLength(0)
    })
  })
})
