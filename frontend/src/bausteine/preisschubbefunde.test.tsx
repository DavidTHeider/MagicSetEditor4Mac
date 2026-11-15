// **Die Befunde eines Preisschubs** — Werkstück 6 von Phase 4.5
// (#1292),
// ausgelagert aus `Massenvorgaenge.test.tsx`, als der Massenbericht seinen
// Stand vorher bekam (DOM-03.AK-03) und die Datei über die Zeilengrenze ging.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Massenvorgaenge } from './Massenvorgaenge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { Api } from '../api'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.sammlung

const zwei = [
  { id: 1, fassung: 3 },
  { id: 2, fassung: 1 },
]

// zeige stellt den Kasten auf — nur der Preisschub ist belegt.
function zeige(
  _gewaehlt: typeof zwei,
  zusatz: { MassePreise: Api['MassePreise'] },
) {
  belegeApi(zusatz)

  render(
    <Sprachraum sprache="de">
      <Massenvorgaenge gewaehlt={zwei} gesamt={5} aufGeaendert={vi.fn()} />
    </Sprachraum>,
  )
}

// **Und die des Preisschubs ebenso.**
it('zeigt die Befunde eines Preisschubs', async () => {
  zeige(zwei, {
    MassePreise: () =>
      Promise.resolve({
        schub: {
          geholt: 1,
          offen: 2,
          fehlgeschlagen: 1,
          fertig: false,
          steht: false,
          befunde: ['probe-druck: kein Netz'],
        },
        fehler: null,
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: texte.massePreise }))

  expect(await screen.findByText('probe-druck: kein Netz')).toBeInTheDocument()
})

// **Ein Schub, der nichts bewegt hat, sagt es — und sagt, dass er steht.**
//
// # Der Befund, der diesen Lauf verlangt
//
// `offen` sank bei jedem Fehlschlag, obwohl der Druck offen blieb; ein ganzer
// Schub voller Fehlschläge meldete »Alle Preise sind da.«. Jetzt steht die
// Wahrheit da — und daneben der Satz, an dem ein Aufrufer aufhört zu rufen.
//
// Nachlauf über Phase 4.5, Befund 7.
it('nennt Fehlschläge und sagt, wenn der Schub steht', async () => {
  zeige(zwei, {
    MassePreise: () =>
      Promise.resolve({
        schub: {
          geholt: 0,
          offen: 20,
          fehlgeschlagen: 20,
          fertig: false,
          steht: true,
          befunde: [],
        },
        fehler: null,
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: texte.massePreise }))

  expect(await screen.findByText(/hat nichts bewegt/)).toBeInTheDocument()

  // **Und die Zahl steht dabei** — ohne sie wäre »steht« eine Behauptung
  // ohne Beleg.
  expect(screen.getByText(/20 fehlgeschlagen/)).toBeInTheDocument()
})
