// **Die Vorwahl springt nicht, während Daten nachkommen** — #2529, gefunden
// in der Gesamtabnahme (QA01, #2382) am Drehbuch `g07palette`: wer „Sets"
// tippte und gleich ⏎ drückte, nahm eine andere Zeile als wer einen
// Augenblick wartete — die Sets kamen nach, und mit ihnen rückte die Vorwahl.
import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { eineListe } from '../seiten/peerkontaktprobe'
import { Palette } from './Palette'

import type { SetDTO } from '../api'

afterEach(raeumeApi)

const werkstatt: SetDTO = {
  id: '7',
  name: 'Werkstatt der Dämmerung',
  spiel: 'magic',
  stil: 'm15',
  herkunft: '',
  karten: 41,
  herkunftFehlt: false,
}

function vorwahl(): string {
  return document.querySelector('[data-vorwahl]')?.textContent ?? ''
}

it('lässt die Vorwahl stehen, wenn die Sets nach dem Tippen kommen', async () => {
  let antworte: (sets: SetDTO[]) => void = () => undefined
  belegeApi({
    Befehlskatalog: () => Promise.resolve({ befehle: [], fehler: null }),
    Sets: () =>
      new Promise((fertig) => {
        antworte = (sets) => {
          fertig({ sets, fehler: null })
        }
      }),
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Peerkontakte: () => Promise.resolve(eineListe([])),
    Suchen: () =>
      Promise.resolve({ treffer: [], gekuerzt: false, fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Anschlussraum aktuell="start" wechseln={vi.fn()}>
        <Palette
          schliessen={vi.fn()}
          waehlen={vi.fn()}
          waehleBefehl={vi.fn()}
          wege={{ setID: '7', waehleSet: vi.fn(), oeffneKarte: vi.fn() }}
        />
      </Anschlussraum>
    </Sprachraum>,
  )

  await userEvent.type(screen.getByRole('textbox'), 'Sets')
  const vorher = vorwahl()
  expect(vorher).not.toBe('')

  // **Unter „Sets" zeigt sich die Ankunft nicht** — gewartet wird deshalb die
  // Makroaufgabe, mit der die Vorrichtung antwortet (`test/api.ts`), und
  // belegt wird sie danach unter „Karte" (§3: vor eine Verneinung gehört ein
  // Beleg, dass jemand zuhört).
  antworte([werkstatt])
  await act(async () => {
    await new Promise((weiter) => {
      setTimeout(weiter, 20)
    })
  })

  expect(vorwahl()).toBe(vorher)

  const feld = screen.getByRole('textbox')
  await userEvent.clear(feld)
  await userEvent.type(feld, 'Karte')
  const tun = await screen.findByRole('list', {
    name: kataloge.de.allgemein.palette.tun,
  })
  expect(
    within(tun).getByText('Gestalten · Werkstatt der Dämmerung'),
  ).toBeInTheDocument()
})
