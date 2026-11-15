// **Besitze ich diese Karte?** — Werkstück 5 von Phase 4.5
// (#1292,
// Vermerk 0057).
import { render, screen } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { Besitzstand } from './Besitzstand'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'

import type { BesitzDTO } from '../antwortensammlung'

afterEach(() => {
  raeumeApi()
})

// zeige stellt den Kasten mit einem festen Stand auf.
function zeige(besitz: BesitzDTO | null) {
  belegeApi({
    SammlungBesitz: () => Promise.resolve({ besitz, fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Besitzstand printing="probe-druck" />
    </Sprachraum>,
  )
}

// **Wer die Karte hat, liest wie viele.**
it('sagt, wie viele Stücke da sind', async () => {
  zeige({
    printing: 'probe-druck',
    stueck: 4,
    gewuenscht: 0,
    besitzt: true,
    zeilen: [],
  })

  expect(await screen.findByText('Du besitzt 4 davon.')).toBeInTheDocument()
})

// **Ein Wunsch ist kein Besitz** (0057, Punkt 2).
//
// Der Fall, für den es diese Zeile gibt: die Karte steht in der Sammlung —
// aber auf der Wunschliste.
it('sagt Nein, wenn es die Karte nur als Wunsch gibt', async () => {
  zeige({
    printing: 'probe-druck',
    stueck: 0,
    gewuenscht: 3,
    besitzt: false,
    zeilen: [],
  })

  expect(
    await screen.findByText('Du besitzt diesen Druck nicht.'),
  ).toBeInTheDocument()

  // **Und der Wunsch steht trotzdem dabei** — er ist die Antwort auf dieselbe
  // Frage mit dem anderen Vorzeichen.
  expect(screen.getByText('(3 auf der Wunschliste)')).toBeInTheDocument()
})

// **Ohne Antwort steht nichts da** — und nichts stürzt ab.
it('zeigt nichts, solange keine Antwort da ist', () => {
  zeige(null)

  expect(screen.queryByText(/besitzt/)).not.toBeInTheDocument()
})
