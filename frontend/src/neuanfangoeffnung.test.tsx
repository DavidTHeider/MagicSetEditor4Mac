// **Ein Neuanfang entwertet auch, was gerade aufgeht** — Befund 8 des
// Zweitblicks über Phase G, zweiter Teil (#2292,
// #2292).
//
// Der Neuanfang räumte Set und Karte, aber nicht den Lauf der Setwahl. Eine
// Datei, die vorher abgelegt worden war und deren Antwort danach ankam,
// schlug ihr Set im neuen Bestand wieder auf — mit einer Kennung, die es
// dort vielleicht gar nicht gibt (§7, Regel 1 und 2).
import { render, screen, waitFor } from '@testing-library/react'
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

const gelesen = (setID: string) => ({
  abgebrochen: false,
  setID,
  karten: 3,
  befunde: [],
  angelegt: [setID],
  unvollstaendig: false,
  schonOffen: false,
  fehler: null,
})

it('schlägt nach einem Neuanfang kein Set auf, dessen Datei vorher abgelegt wurde', async () => {
  let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined
  ;(globalThis as MitLaufzeit).runtime = {
    EventsOn: () => () => undefined,
    OnFileDrop: (rueckruf) => {
      ablegen = rueckruf
    },
    OnFileDropOff: () => undefined,
  }

  let spaeter: (() => void) | undefined
  const SetAusDatei = vi
    .fn()
    .mockImplementationOnce(
      () =>
        new Promise((fertig) => {
          spaeter = () => {
            fertig(gelesen('7'))
          }
        }),
    )
    .mockResolvedValue(gelesen('9'))
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
    SetAusDatei,
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
        fehler: null,
      }),
  })

  const nutzer = userEvent.setup()
  render(<App />)
  await screen.findByRole('heading', {
    name: kataloge.de.seiten.start.ueberschrift,
  })
  await effekteNachholen()

  // **Die Datei geht auf** — und antwortet erst nach dem Neuanfang.
  ablegen?.(0, 0, ['/a/alt.mse-set'])
  await waitFor(() => {
    expect(SetAusDatei).toHaveBeenCalledOnce()
  })

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

  spaeter?.()
  await effekteNachholen()

  expect(Kartenzeilen.mock.calls.map((ruf) => String(ruf[0]))).not.toContain(
    '7',
  )

  // **Der Beleg, dass jemand zuhört** (§3): eine Ablage nach dem Neuanfang
  // schlägt ihr Set auf.
  ablegen?.(0, 0, ['/a/neu.mse-set'])
  await waitFor(() => {
    expect(Kartenzeilen.mock.calls.map((ruf) => String(ruf[0]))).toContain('9')
  })
})
