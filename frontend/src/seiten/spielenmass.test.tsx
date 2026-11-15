// **Die Spielenseite mit echten Mengen, nach dem Prüfmaß** — UXR-13
// (#1921).
//
// # Wozu
//
// Die Ausgangslage in `docs/UIREVIEW.md` ist am Probebestand gemessen — ohne
// Deck, ohne Partie, ohne Gegenüber. Dort zählte *Spielen* 17
// Bedienelemente und drei Felder. Mit dem Bestand aus `test/spielbestand.ts`
// (acht Decks, neunzehn Partien, drei Gegenüber, zwei Profile) und vor
// UXR-13 zeichnete dieselbe Seite **142 Bedienelemente und 19 Felder**, mit
// gewähltem Deck 21, und mit „Wir wählen selbst" **34 Felder** — gezählt nach
// den Wählern der Hüllenprobe `uirevue` (`test/bedienzaehlung.ts`).
//
// # Was es hält
//
//   - **R3**: höchstens zwölf Felder auf der Seite — ohne Deck, mit Deck und
//     im freien Modus.
//   - **R1**: höchstens ein Hauptknopf je Region, und der Anfang führt.
//   - **R2 als Schranke für das Gezeichnete**: jsdom rechnet kein Layout,
//     „im Bild" sieht es nicht; was gezeichnet ist, ist eine obere Schranke
//     dafür. Nach UXR-13, gemessen am 2026-10-02: 20 Bedienelemente und neun
//     Felder, mit gewähltem Deck 22 und elf (die Regeltafel kommt dazu und
//     schiebt den Anfang nach unten), im freien Modus 24 und elf.
//
// # Was es nicht hält
//
// **Die Rolltiefe und das waagerechte Rollen** (R4, R5) — die misst die
// Hüllenprobe `uxr13` am nativen Fenster.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { Spielen } from './Spielen'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { mitDecks } from './spielenprobe'
import { spielbestand } from '../test/spielbestand'
import { zaehle } from '../test/bedienzaehlung'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.spielen

afterEach(() => {
  raeumeApi()
})

// zeige zeichnet die Seite mit dem Bestand und wartet, bis Liste, Sitze und
// Decks dastehen.
async function zeige(): Promise<HTMLElement> {
  mitDecks(spielbestand())

  const { container } = render(
    <Sprachraum sprache="de">
      <Spielen wechseln={() => undefined} />
    </Sprachraum>,
  )

  await screen.findAllByText(/Ada · Bo/)
  await screen.findByLabelText(`${eigene.sitzName} 1`)
  await screen.findAllByRole('option', { name: 'Glutsturm' })

  return container
}

// **Seit P6 geht die Seite mit vorbelegten Sitzen auf** (#2374) — der
// Zustand ohne Deck kommt nur noch ohne eigenes Deck vor, und dort steht der
// Leerzustand. Die Schranke ist die mit gewähltem Deck, gemessen am
// 2026-10-02 (oben): 22.
it('hält R1 bis R3 mit echten Mengen, wie die Seite aufgeht', async () => {
  const seite = await zeige()
  await screen.findByText(eigene.regelnTafel)

  const zahl = zaehle(seite)

  expect(zahl.felder).toBeLessThanOrEqual(12)
  expect(zahl.hauptJeRegion).toBeLessThanOrEqual(1)
  expect(zahl.bedien).toBeLessThanOrEqual(22)
})

it('hält R3 mit gewähltem Deck und im freien Modus', async () => {
  const seite = await zeige()
  const hand = userEvent.setup()

  for (const feld of screen.getAllByLabelText(eigene.sitzDeck)) {
    await hand.selectOptions(feld, 'aschenrat')
  }
  await screen.findByText(eigene.regelnTafel)

  expect(zaehle(seite).felder).toBeLessThanOrEqual(12)

  // **Der Anfang führt**: sein Knopf ist der Hauptknopf der Seite.
  expect(screen.getByRole('button', { name: eigene.anfangen })).toHaveClass(
    'bg-marke',
  )
  expect(zaehle(seite).haupt).toBe(1)

  await hand.click(screen.getByRole('radio', { name: eigene.modusFrei }))

  expect(zaehle(seite).felder).toBeLessThanOrEqual(12)
})
