// **Der Verweis auf einen Druck ohne Alias** — Werkstück A2 von Phase 5.3
// (#1441).
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KartenverweisDTO } from '../apiscryfall'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Kartenverweis } from './Kartenverweis'

function verweis(teile: Partial<KartenverweisDTO> = {}): KartenverweisDTO {
  return {
    karteID: '7',
    printing: '516e137a-d9d9-4ba7-89fb-e680ddb2c2d1',
    druckID: '',
    vorlaeufig: false,
    name: 'Lightning Bolt',
    setcode: 'lea',
    nummer: '161',
    adresse: 'https://scryfall.com/card/lea/161',
    gesetzt: '2026-08-27T12:00:00Z',
    ...teile,
  }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Kartenverweis
        karteID="7"
        vorschlag={undefined}
        vorgang={{ starte: (_, was) => was(), laeuft: () => false }}
        aufFehler={vi.fn()}
      />
    </Sprachraum>,
  )
}

afterEach(() => {
  raeumeApi()
})

const marke = kataloge.de.bausteine.druck.vorlaeufig

// **Ein Druck ohne Alias ist vorläufig** — Werkstück A2 von Phase 5.3 (#1441).
describe('der Verweis und die vorläufige Identität', () => {
  it('zeigt die Marke, wenn der bezogene Druck keinen Alias hat', async () => {
    belegeApi({
      KarteVerweis: vi.fn().mockResolvedValue({
        verweis: verweis({ printing: '', druckID: '9', vorlaeufig: true }),
        befunde: [],
        fehler: null,
      }),
      SammlungBesitz: vi.fn().mockResolvedValue({
        besitz: { printing: '', stueck: 0, gewuenscht: 0, besitzt: false },
        fehler: null,
      }),
    })

    zeige()

    expect(await screen.findByText(marke)).toBeInTheDocument()
  })
})
