// **Die gewählte Karte steht in der Menüleiste** —
// #2482,
// Befund 11 des Zweitblicks über Phase 7. Rechtsklick und ⇧F10 teilten ihre
// Liste, die Menüleiste bot die Tischhandlungen nicht an. Jetzt meldet der
// Tisch dieselbe Liste an — mit `schreibt`, damit ein Leser sie gesperrt
// sieht, und nur, solange der Tisch bedienbar ist.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import type { Objekthandlungen } from '../objekthandlungen'
import { Objekthandlungsraum } from '../objekthandlungsraum'
import { Partietisch } from './Partietisch'
import { Ruecknahmeraum } from '../ruecknahme'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { einLauf, einObjekt, einePartie } from './partieprobe'

afterEach(() => {
  raeumeApi()
})

it('meldet die Handlungen der gewählten Karte an der Menüleiste an', async () => {
  const melden = vi.fn<
    (meldung: Objekthandlungen) => Promise<{ fehler: null }>
  >(() => Promise.resolve({ fehler: null }))
  belegeApi({ ObjekthandlungenMelden: melden })
  const lauf = einLauf({
    partie: einePartie({
      objekte: [
        einObjekt({
          kennung: 'sitz1/wald/0',
          zone: 'tisch',
          name: 'Wald',
          geheim: false,
        }),
      ],
    }),
    handlungen: [
      {
        objekt: 'sitz1/wald/0',
        handlungen: ['objekt.tappen'],
        primaer: 'objekt.tappen',
        zweit: '',
      },
    ],
  })

  render(
    <Sprachraum sprache="de">
      <Objekthandlungsraum>
        <Ruecknahmeraum>
          <Partietisch lauf={lauf} partie={lauf.partie ?? einePartie()} />
        </Ruecknahmeraum>
      </Objekthandlungsraum>
    </Sprachraum>,
  )

  // **⌘-Klick wählt** (P4) und stellt nichts.
  fireEvent.click(screen.getByRole('button', { name: /Wald/ }), {
    detail: 1,
    metaKey: true,
  })

  await waitFor(() => {
    const karte = melden.mock.calls
      .map(([meldung]) => meldung)
      .filter((meldung) => meldung.art === 'spielobjekt')
      .at(-1)
    expect(karte?.kennung).toBe('sitz1/wald/0')
    expect(karte?.eintraege.map((eintrag) => eintrag.name)).toContain(
      kataloge.de.seiten.handlungen.handlungsname['objekt.tappen'],
    )
  })
})
