// **Das Menü einer Karte trägt ihre Kürzel** — AP-7
// (#2364),
// Befunde B-02 und B-05, Entwurf `Kontext09`: ⌘↩ legt an, ⌘D verdoppelt,
// ⌘⌫ löscht mit Rückfrage, ⇞ und ⇟ blättern — aus **derselben**
// Handlungsliste, die »Diese Karte«, der Rechtsklick und die Leiste zeigen.
//
// Was die Tasten am Fenster tun — auch mit dem Fokus in einem Textfeld —,
// misst `ap5kontext09`; hier steht, dass jede Zeile ihr Kürzel trägt und ihre
// Handlung tut.
import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, KarteDTO } from '../api'
import { Objekthandlungsraum } from '../objekthandlungsraum'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { vorschau } from '../test/bildflaeche'
import { ausKarten } from '../test/kartenproben'
import { oeffneKartenhandlungen } from '../test/kartenlasche'
import { Karten } from './Karten'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.karten

const karten: KarteDTO[] = [
  { id: '1', nummer: 0, werte: [{ feld: 'name', wert: 'Feuerball' }] },
  { id: '2', nummer: 1, werte: [{ feld: 'name', wert: 'Mirri' }] },
]

function zeige(teile: Partial<Api> = {}, karteID = '1') {
  const gewaehlt = vi.fn()
  belegeApi({
    Stile: () => Promise.resolve({ stile: [] }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),
    Kartenfelder: () => Promise.resolve({ felder: [] }),
    ...ausKarten(() => Promise.resolve({ karten })),
    Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
    Verlauf: () => Promise.resolve({ schritte: [] }),
    Vorschau: () => Promise.resolve(vorschau()),
    Probevorschau: () => Promise.resolve(vorschau()),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Karten
        karteGewaehlt={gewaehlt}
        setID="1"
        karteID={karteID}
        verlaufsstand={0}
        verlaufsfehler={undefined}
        bestandGeaendert={() => undefined}
      />
    </Sprachraum>,
  )

  return gewaehlt
}

describe('das Menü »Diese Karte«', () => {
  it('trägt Anlegen, Verdoppeln, Löschen und Blättern mit ihren Kürzeln', async () => {
    zeige()
    await oeffneKartenhandlungen()

    for (const [name, tasten] of [
      [eigene.anlegen, 'Meta+Enter'],
      [eigene.duplizieren, 'Meta+D'],
      [eigene.blaettern.vorige, 'PageUp'],
      [eigene.blaettern.naechste, 'PageDown'],
    ] as const) {
      expect(
        screen.getByRole('menuitem', { name: new RegExp(`^${name}`) }),
      ).toHaveAttribute('aria-keyshortcuts', tasten)
    }
    expect(
      screen.getByRole('menuitem', { name: new RegExp(`^${eigene.loeschen}`) }),
    ).toHaveAttribute('aria-keyshortcuts', 'Meta+Backspace')
  })

  it('legt über »Karte anlegen« eine Karte an', async () => {
    const anlegen = vi.fn(() => Promise.resolve({ karteID: '3', fehler: null }))
    zeige({ KarteAnlegen: anlegen })
    await oeffneKartenhandlungen()

    await userEvent.click(
      screen.getByRole('menuitem', { name: new RegExp(`^${eigene.anlegen}`) }),
    )

    expect(anlegen).toHaveBeenCalledWith('1')
  })

  it('blättert zur nächsten Karte und ruht vor der ersten', async () => {
    const gewaehlt = zeige()
    await oeffneKartenhandlungen()

    expect(
      screen.getByRole('menuitem', {
        name: new RegExp(`^${eigene.blaettern.vorige}`),
      }),
    ).toHaveAttribute('aria-disabled', 'true')

    await userEvent.click(
      screen.getByRole('menuitem', {
        name: new RegExp(`^${eigene.blaettern.naechste}`),
      }),
    )

    expect(gewaehlt).toHaveBeenLastCalledWith('2')
  })

  it('blättert von der zweiten zur vorigen Karte', async () => {
    const gewaehlt = zeige({}, '2')
    await oeffneKartenhandlungen()

    await userEvent.click(
      screen.getByRole('menuitem', {
        name: new RegExp(`^${eigene.blaettern.vorige}`),
      }),
    )

    expect(gewaehlt).toHaveBeenLastCalledWith('1')
  })
})

// **Rechtsklick in der Kartenliste** — AP-5
// (#2362),
// Befund B-01: auf der offenen Karte öffnet er ihr Menü; auf einer anderen
// wählt er sie zuerst und zeigt dann **ihres** — nie das der vorigen.
describe('der Rechtsklick in der Kartenliste', () => {
  // Buehne hält die gewählte Karte wie die Anwendung — sonst bliebe die
  // vorige offen, und das Menü gehörte ihr.
  function Buehne({ gewaehlt }: { gewaehlt: (karteID: string) => void }) {
    const [karte, setzeKarte] = useState('1')

    return (
      <Sprachraum sprache="de">
        <Objekthandlungsraum>
          <Karten
            karteGewaehlt={(karteID) => {
              gewaehlt(karteID)
              setzeKarte(karteID)
            }}
            setID="1"
            karteID={karte}
            verlaufsstand={0}
            verlaufsfehler={undefined}
            bestandGeaendert={() => undefined}
          />
        </Objekthandlungsraum>
      </Sprachraum>
    )
  }

  function zeigeBuehne() {
    const gewaehlt = vi.fn()
    belegeApi({
      Stile: () => Promise.resolve({ stile: [] }),
      Stiloptionen: () => Promise.resolve({ optionen: [] }),
      Kartenfelder: () => Promise.resolve({ felder: [] }),
      ...ausKarten(() => Promise.resolve({ karten })),
      Vorbehalte: () => Promise.resolve({ vorbehalte: [] }),
      Verlauf: () => Promise.resolve({ schritte: [] }),
      Vorschau: () => Promise.resolve(vorschau()),
      Probevorschau: () => Promise.resolve(vorschau()),
      ObjekthandlungenMelden: () => Promise.resolve({ fehler: null }),
    })
    render(<Buehne gewaehlt={gewaehlt} />)

    return gewaehlt
  }

  it('öffnet auf der offenen Karte ihr Menü', async () => {
    zeigeBuehne()
    const liste = await screen.findByRole('grid', { name: eigene.titel })
    // **Erst, wenn die Karte ihre Handlungen gemeldet hat**, gehört der
    // Rechtsklick ihr — bis dahin bleibt er bei macOS.
    await waitFor(() => {
      expect(liste.closest('[data-objekt]')).toHaveAttribute(
        'data-objekt',
        'karte/1',
      )
    })

    const ereignis = fireEvent.contextMenu(
      screen.getByRole('row', { name: 'Feuerball' }),
      { clientX: 20, clientY: 40 },
    )

    expect(ereignis).toBe(false)
    expect(
      screen.getByRole('menu', { name: eigene.kartenmenue }),
    ).toBeInTheDocument()
  })

  it('wählt eine andere Karte und zeigt dann ihr Menü — nie das der vorigen', async () => {
    const gewaehlt = zeigeBuehne()
    const mirri = await screen.findByRole('row', { name: 'Mirri' })

    const ereignis = fireEvent.contextMenu(mirri, { clientX: 20, clientY: 40 })

    expect(ereignis).toBe(false)
    expect(gewaehlt).toHaveBeenLastCalledWith('2')
    const menue = await screen.findByRole('menu', { name: eigene.kartenmenue })
    expect(menue.closest('[data-objekt]')).toHaveAttribute(
      'data-objekt',
      'karte/2',
    )
    expect(
      within(menue).getByRole('menuitem', {
        name: new RegExp(`^${eigene.duplizieren}`),
      }),
    ).toBeInTheDocument()
  })
})
