// **Eine Kennung aus dem Finder gehört zu ihrem Bestand** — Befund 1 des
// nachgeholten Zweitblicks über die Folgepakete (J05, #2269, zu
// #2292).
//
// Go meldet den Doppelklick als Ereignis, und Wails stellt es in eine eigene
// Schlange; die Antwort des Neuanfangs kann es überholen. Kam die Kennung aus
// dem alten Bestand danach an, schlug das Fenster sie im neuen auf. **Die
// Epoche macht sie erkennbar**: die Meldung trägt ihre, die Antwort des
// Neuanfangs die neue, und was älter ist, wird verworfen.
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { effekteNachholen } from './test/effekte'
import { zuAbschnitt } from './test/gebiete'

interface MitLaufzeit {
  runtime?: {
    EventsOn: (
      name: string,
      rueckruf: (...daten: unknown[]) => void,
    ) => () => void
    OnFileDrop?: (
      rueckruf: (x: number, y: number, pfade: string[]) => void,
      ueberZiel?: boolean,
    ) => void
    OnFileDropOff?: () => void
  }
}

afterEach(() => {
  delete (globalThis as MitLaufzeit).runtime
  raeumeApi()
})

const worte = kataloge.de.seiten.ruecksetzung

it('schlägt nach einem Neuanfang keine Kennung aus dem alten Bestand auf', async () => {
  const hoerer = new Map<string, (...daten: unknown[]) => void>()
  ;(globalThis as MitLaufzeit).runtime = {
    EventsOn: (name, rueckruf) => {
      hoerer.set(name, rueckruf)
      return () => undefined
    },
    OnFileDrop: () => undefined,
    OnFileDropOff: () => undefined,
  }

  const Kartenzeilen = vi.fn().mockResolvedValue({ zeilen: [], fehler: null })

  belegeApi({
    ...belegeFenster('de'),
    Startbefund: vi.fn().mockResolvedValue({ offen: true, grund: '' }),
    Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
    Pakete: vi.fn().mockResolvedValue({ pakete: [], fehler: null }),
    Karten: vi.fn().mockResolvedValue({ karten: [], fehler: null }),
    Kartenfelder: vi.fn().mockResolvedValue({ felder: [], fehler: null }),
    Setangaben: vi
      .fn()
      .mockResolvedValue({ felder: [], werte: [], fehler: null }),
    Kartenzeilen,
    Einstellungen: () =>
      Promise.resolve({
        einstellungen: [
          {
            schluessel: 'sprache',
            wert: 'de',
            wirksam: 'de',
            werte: ['de', 'en'],
            bereich: 'anwendung',
          },
        ],
      }),
    Ruecksetzungsarten: () => Promise.resolve(['anwendung']),
    Ruecksetzungsvorschau: () =>
      Promise.resolve({
        art: 'anwendung',
        betroffen: [],
        hindernisse: [],
        bereit: true,
        nameNoetig: false,
        nachweisNoetig: true,
        sicherungEmpfohlen: true,
        revision: 'marke',
        fehler: null,
      }),
    RuecksetzungFreigabeErteilen: () =>
      Promise.resolve({
        kennung: 'freigabe',
        laeuftab: '2026-10-10T20:00:00Z',
        fehler: null,
      }),
    Zuruecksetzen: () =>
      Promise.resolve({
        art: 'anwendung',
        zurueckgesetzt: 0,
        beiseite: '/Users/probe/vorher.db',
        bestandsepoche: 2,
        fehler: null,
      }),
  })

  const nutzer = userEvent.setup()
  render(<App />)
  await screen.findByRole('heading', {
    name: kataloge.de.seiten.start.ueberschrift,
  })
  await effekteNachholen()

  await nutzer.click(
    await screen.findByRole('button', {
      name: kataloge.de.seiten.einstellungen.titel,
    }),
  )
  await zuAbschnitt(
    kataloge.de.seiten.einstellungen.gruppen.daten,
    kataloge.de.seiten.einstellungen.abschnitte.eingriffe,
  )
  await nutzer.click(
    await screen.findByRole('button', { name: worte.nachsehen }),
  )
  await nutzer.click(
    await screen.findByRole('button', { name: worte.freigeben }),
  )
  await nutzer.click(await screen.findByRole('button', { name: worte.tun }))
  expect(await screen.findByText(/vorher\.db/)).toBeInTheDocument()

  // **Die Kennung aus dem alten Bestand kommt nach der Antwort an.**
  act(() => {
    hoerer.get('datei:geoeffnet')?.('7', 1)
  })
  await effekteNachholen()

  expect(Kartenzeilen.mock.calls.map((ruf) => String(ruf[0]))).not.toContain(
    '7',
  )

  // **Der Beleg, dass jemand zuhört** (§3): eine Kennung aus dem neuen
  // Bestand schlägt ihr Set auf.
  act(() => {
    hoerer.get('datei:geoeffnet')?.('9', 2)
  })
  await waitFor(() => {
    expect(Kartenzeilen.mock.calls.map((ruf) => String(ruf[0]))).toContain('9')
  })
})
