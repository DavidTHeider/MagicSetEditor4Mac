// **Das aufgeschlagene Gegenüber behält den Fokus** — UX-10.AK-02
// (#1682).
//
// Gemessen am 2026-09-25 im Drehbuch `ux10-tastatur`: wer „Bo aufschlagen"
// mit Enter drückte, fand den Fokus danach auf `body` — aus dem Knopf wurde
// ein Text, und der Knopf verschwand unter dem Fokus. Sichtbar stand er
// nirgends, bis zum nächsten Tab.
import { act, fireEvent, render, screen } from '@testing-library/react'
import { useState } from 'react'
import { afterEach, describe, expect, it } from 'vitest'

import { Partnerwahl } from './Partnerwahl'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { Peerkontakt } from '../apipeerkontakt'

const eigene = kataloge.de.seiten.handeln

afterEach(() => {
  raeumeApi()
})

const bo: Peerkontakt = {
  kennung: 'k-bo',
  fingerabdruck: 'ab12',
  selbstgewaehlterName: 'Bo',
  eigenerName: '',
  stand: 'bestaetigt',
  gesehen: '',
  rechte: [],
  adresse: '',
  erstmals: '',
  notiz: '',
  lage: 'bestaetigt',
  vergleichscode: '',
}

// Probe hält, welches Gegenüber aufgeschlagen ist — wie die Handelsseite.
function Probe() {
  const [offen, setzeOffen] = useState<string>()

  return <Partnerwahl offen={offen} waehlen={setzeOffen} />
}

describe('die Wahl des Gegenübers', () => {
  it('lässt den Knopf stehen, gedrückt und mit dem Fokus', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve({ kontakte: [bo], zwecke: [] }),
    })
    render(
      <Sprachraum sprache="de">
        <Probe />
      </Sprachraum>,
    )

    const knopf = await screen.findByRole('button', {
      name: eigene.partnerWaehlenName.replace('{name}', 'Bo'),
    })
    knopf.focus()

    act(() => {
      fireEvent.click(knopf)
    })

    const gedrueckt = screen.getByRole('button', {
      name: eigene.partnerOffenName.replace('{name}', 'Bo'),
    })

    expect(gedrueckt).toBe(knopf)
    expect(gedrueckt).toHaveFocus()
    expect(gedrueckt).toHaveAttribute('aria-pressed', 'true')
    expect(gedrueckt).toHaveTextContent(eigene.partnerOffen)
  })
})
