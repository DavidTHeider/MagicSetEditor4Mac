// **Der eine Handel, an dem du dran bist, steht offen** — P13
// (#2381),
// H4 ≤ 2: wer ⌘5 drückt, weil ein Angebot kam, soll gleich zustimmen
// können. Bei mehreren wählt ein Mensch, und was er schon aufgeschlagen
// hat, bleibt.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { Handeln } from './Handeln'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { eineUebersicht, eineZeile } from './handelslisteprobe'
import { kataloge } from '../sprache/kataloge'

import type { Handelsvorgangzeile } from '../apihandelsverzeichnis'

const eigene = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
})

const turfer = eineZeile()
const wartend = eineZeile(
  {
    lage: 'wartet',
    naechster: '',
    gegenueber: [{ kennung: 'bzweiter', name: 'Zweiter' }],
  },
  { vorgang: 'h-2', lage: 'wartet', naechster: '' },
)

function zeige(zeilen: Handelsvorgangzeile[]) {
  belegeApi({
    HandelUebersicht: () => eineUebersicht(zeilen),
    HandelVorgaenge: (kennung: string) =>
      Promise.resolve({
        vorgaenge: zeilen
          .filter((zeile) => zeile.gegenueber[0]?.kennung === kennung)
          .map((zeile) => zeile.vorgang),
        fehler: null,
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )
}

describe('Handeln beim Öffnen', () => {
  it('schlägt den einen Handel auf, an dem du dran bist', async () => {
    zeige([turfer, wartend])

    expect(
      await screen.findByRole('button', { name: /^Turfer/, current: true }),
    ).toBeVisible()
    expect(
      await screen.findByRole('button', { name: eigene.schrittZustimmen }),
    ).toBeVisible()
  })

  it('wählt nicht, wenn du an mehreren dran bist', async () => {
    zeige([
      turfer,
      eineZeile(
        { gegenueber: [{ kennung: 'bdritter', name: 'Dritter' }] },
        { vorgang: 'h-3' },
      ),
    ])

    // **Erst die Zeilen, dann die Verneinung** (§3).
    await screen.findByRole('button', { name: /^Dritter/ })
    expect(screen.queryByRole('button', { current: true })).toBeNull()
    expect(screen.getByText(eigene.partnerKeiner)).toBeVisible()
  })

  it('schlägt einen Handel ohne Gegenüber nicht auf', async () => {
    // **Nach der Matrix gibt es ihn nicht**, in einem von Hand geänderten
    // Bestand schon — und ohne Gegenüber gibt es keinen Raum.
    zeige([eineZeile({ gegenueber: [] }), wartend])

    await screen.findByRole('button', { name: /^Zweiter/ })
    expect(screen.queryByRole('button', { current: true })).toBeNull()
    expect(screen.getByText(eigene.partnerKeiner)).toBeVisible()
  })

  it('wählt nichts von selbst, wo keiner auf dich wartet', async () => {
    zeige([wartend])
    const hand = userEvent.setup()

    await screen.findByRole('button', { name: /^Zweiter/ })
    expect(screen.queryByRole('button', { current: true })).toBeNull()

    // **Was ein Mensch aufschlägt, schlägt er auf** — und danach wählt die
    // Seite nichts mehr von selbst (`vonSelbst` fällt).
    await hand.click(screen.getByRole('button', { name: /^Zweiter/ }))
    expect(
      await screen.findByRole('button', { name: /^Zweiter/, current: true }),
    ).toBeVisible()
  })
})
