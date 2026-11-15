// **⌘↩ legt auch im leeren Set eine Karte an** — Befund 2 des Zweitblicks
// über Phase 6 von 0.9.0 (AP-7, #2364): das Kürzel stand nur an den
// Handlungen einer offenen Karte, und ein leeres Set hat keine. Anlegen ist
// eine Handlung am Set; der Kopf der Lasche meldet sie, solange keine Karte
// offen ist — ist eine offen, hat deren Liste die Leiste.
import { act, fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Objekthandlungsraum } from '../objekthandlungsraum'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { ausKarten } from '../test/kartenproben'
import { Karten } from './Karten'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.karten

describe('das leere Set', () => {
  it('legt mit ⌘↩ eine Karte an und meldet »Karte anlegen« an die Leiste', async () => {
    const anlegen = vi.fn(() => Promise.resolve({ karteID: '1', fehler: null }))
    const gemeldet = vi.fn<
      (meldung: {
        eintraege: { kuerzel?: string }[]
      }) => Promise<{ fehler: null }>
    >(() => Promise.resolve({ fehler: null }))
    belegeApi({
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      ...ausKarten(() => Promise.resolve({ karten: [] })),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Verlauf: () => Promise.resolve({ schritte: [] }),
      KarteAnlegen: anlegen,
      ObjekthandlungenMelden: gemeldet,
    })

    render(
      <Sprachraum sprache="de">
        <Objekthandlungsraum>
          <Karten
            karteGewaehlt={vi.fn()}
            setID="1"
            karteID={undefined}
            verlaufsstand={0}
            verlaufsfehler={undefined}
            bestandGeaendert={() => undefined}
          />
        </Objekthandlungsraum>
      </Sprachraum>,
    )
    await screen.findByText(eigene.leerTitel)
    await act(async () => {
      await Promise.resolve()
    })

    expect(
      gemeldet.mock.calls.some(([meldung]) =>
        meldung.eintraege.some(
          (eintrag) => eintrag.kuerzel === 'befehl+return',
        ),
      ),
    ).toBe(true)

    fireEvent.keyDown(document.body, { key: 'Enter', metaKey: true })

    expect(anlegen).toHaveBeenCalledWith('1')
  })
})
