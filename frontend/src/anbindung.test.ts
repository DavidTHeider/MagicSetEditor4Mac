// **Die Anbindung an Wails 3** — ENG-02.
//
// Geprüft wird die Gestalt, die Fassade und Fenster lesen: dass jede Methode
// unter dem vollen Namen der Fassade gerufen wird, dass Ereigniswerte als
// Werte ankommen und dass ein Ablegen dieselben drei Angaben trägt wie unter
// Wails 2. **Was hier nicht geprüft wird**: dass Wails 3 die Namen so bildet
// — das hält der Lauf `TestDerVorsatzDerAnbindungIstDerDerFassade` in Go.
import { describe, expect, it, vi } from 'vitest'

import {
  anbinden,
  ereignisAbgelegt,
  fehlermarke,
  vorsatz,
  wailsBindungen,
  wailsLaufzeit,
  type Hoerer,
} from './anbindung'
import { Bindungsfehler } from './bindungsfehler'
import { methoden } from './fassade'

// probehoerer merkt sich jeden Rückruf und gibt eine Abmeldung zurück.
function probehoerer() {
  const rueckrufe = new Map<string, (ereignis: { data: unknown }) => void>()
  const abgemeldet: string[] = []

  const hoerer: Hoerer = {
    On: (name, rueckruf) => {
      rueckrufe.set(name, rueckruf)

      return () => {
        abgemeldet.push(name)
        rueckrufe.delete(name)
      }
    },
  }

  const melde = (name: string, data: unknown) => {
    rueckrufe.get(name)?.({ data })
  }

  return { hoerer, melde, abgemeldet, rueckrufe }
}

describe('die Anbindung an Wails 3', () => {
  it('ruft jede Methode der Fassade unter ihrem vollen Namen', async () => {
    const ByName = vi.fn(() => Promise.resolve('antwort'))
    const gebaut = wailsBindungen({ ByName }) as Record<
      string,
      (...argumente: unknown[]) => Promise<unknown>
    >

    // **Jede Methode genau einmal** — die Liste der Fassade trägt keine
    // Doppelten, sonst stünde hier eine Zahl, die keine Bindung ist.
    expect(Object.keys(gebaut)).toHaveLength(methoden.length)
    expect(new Set(methoden).size).toBe(methoden.length)

    const setAusDatei = gebaut.SetAusDatei
    if (setAusDatei === undefined) {
      throw new Error('SetAusDatei fehlt in der Anbindung')
    }

    await expect(setAusDatei('/wo/set.mse-set')).resolves.toBe('antwort')
    expect(ByName).toHaveBeenCalledWith(
      'github.com/DavidTHeider/MagicSetEditor4Mac/internal/app.Bindungen.SetAusDatei',
      '/wo/set.mse-set',
    )
    expect(vorsatz.endsWith('.Bindungen.')).toBe(true)
  })

  it('reicht die Werte eines Ereignisses einzeln durch — und nichts, wenn keine Liste kam', () => {
    const { hoerer, melde } = probehoerer()
    const laufzeit = wailsLaufzeit(hoerer)
    const bekommen: unknown[][] = []

    laufzeit.EventsOn('menue:set-zuletzt', (...daten) => bekommen.push(daten))

    melde('menue:set-zuletzt', ['7'])
    melde('menue:set-zuletzt', [])
    melde('menue:set-zuletzt', null)

    expect(bekommen).toEqual([['7'], [], []])
  })

  it('meldet ein Ablegen mit Stelle und Pfaden — und nimmt die Anmeldung zurück', () => {
    const { hoerer, melde, abgemeldet } = probehoerer()
    const laufzeit = wailsLaufzeit(hoerer)
    const abgelegt: unknown[][] = []

    laufzeit.OnFileDrop((x, y, pfade) => abgelegt.push([x, y, pfade]))
    melde(ereignisAbgelegt, { x: 12, y: 34, pfade: ['/a.png'] })

    // **Ein zweites Anmelden ersetzt das erste**, wie unter Wails 2.
    laufzeit.OnFileDrop(() => abgelegt.push(['zweiter']))
    melde(ereignisAbgelegt, { x: 0, y: 0, pfade: [] })

    laufzeit.OnFileDropOff()
    laufzeit.OnFileDropOff()
    melde(ereignisAbgelegt, { x: 1, y: 1, pfade: ['/b.png'] })

    expect(abgelegt).toEqual([[12, 34, ['/a.png']], ['zweiter']])
    expect(abgemeldet).toEqual([ereignisAbgelegt, ereignisAbgelegt])
  })

  it('legt Bindungen und Laufzeit dort ab, wo Fassade und Fenster suchen', () => {
    const ziel: {
      go?: { app?: { Bindungen?: Record<string, unknown> } }
      runtime?: { EventsOn?: unknown }
    } = {}

    anbinden(
      { ByName: () => Promise.resolve(null) },
      probehoerer().hoerer,
      ziel,
    )

    expect(typeof ziel.go?.app?.Bindungen?.Datenstand).toBe('function')
    expect(typeof ziel.runtime?.EventsOn).toBe('function')
  })

  // **Der Fenstertitel geht über die Laufzeit** — AP-7 (#2364): kein
  // eigener Ruf nach Go, sondern Wails' `Window.SetTitle`.
  it('setzt den Fenstertitel über das Fenster der Laufzeit', () => {
    const titel: string[] = []
    const ziel: { runtime?: { TitelSetzen?: (titel: string) => void } } = {}

    anbinden(
      { ByName: () => Promise.resolve(null) },
      probehoerer().hoerer,
      ziel,
      {
        SetTitle: (neu: string) => {
          titel.push(neu)
          return Promise.resolve()
        },
      },
    )
    ziel.runtime?.TitelSetzen?.('Probeset — MagicSetEditor4Mac')

    expect(titel).toEqual(['Probeset — MagicSetEditor4Mac'])
  })
})

// **Eine Abweisung aus dem Tor kommt als Bindungsfehler an** — OPS-10
// (#1908): Wails gibt eine Panik nur als Satz weiter, und der Satz trägt den
// Fehler hinter einer Marke.
describe('eine Abweisung aus dem Tor', () => {
  const fehler = {
    code: 'abgeschaltet',
    schluessel: 'bereichAbgeschaltet',
    meldung: 'dieser Bereich ist abgeschaltet: "spielen"',
    werte: { schalter: 'spielen' },
  }

  function gerufenMit(grund: unknown) {
    const gebaut = wailsBindungen({
      // **Auch ein Grund ohne Fehlergestalt** — genau den prüft der dritte Fall.
      // eslint-disable-next-line @typescript-eslint/prefer-promise-reject-errors
      ByName: () => Promise.reject(grund),
    }) as Record<string, () => Promise<unknown>>

    const sets = gebaut.Sets
    if (sets === undefined) {
      throw new Error('Sets fehlt in der Anbindung')
    }

    return sets()
  }

  it('wird als Bindungsfehler ausgepackt', async () => {
    await expect(
      gerufenMit(
        new Error(
          `${vorsatz}Sets: panic: ${fehlermarke}${JSON.stringify(fehler)}`,
        ),
      ),
    ).rejects.toMatchObject({
      name: 'Bindungsfehler',
      code: 'abgeschaltet',
      schluessel: 'bereichAbgeschaltet',
      werte: { schalter: 'spielen' },
    })
  })

  // **Jede andere Ablehnung ist fremd** — Befund 6 des Zweitblicks über
  // Phase 1 von 0.9.0 (F03, #2342). Keine Bindung gibt einen `error` zurück;
  // was dennoch abweist, kommt aus Wails oder dem Transport, etwa
  // `new Error(await response.text())` bei einer Antwort ohne JSON
  // (`runtime.ts`, v3.0.0-beta.27). Bis dahin stand ihr Rohtext im Fenster.
  it.each([
    ['ein schlichter Fehler', new Error('die Verbindung ist weg')],
    [
      'eine Antwort ohne JSON',
      new Error('<html><body>502 Bad Gateway</body></html>'),
    ],
    ['ein Grund ohne Fehlergestalt', 'nur ein Satz'],
  ])('macht aus %s einen Fehler »unbekannt« mit Kennung', async (_, grund) => {
    const ausgang: unknown = await gerufenMit(grund).catch(
      (abweisung: unknown) => abweisung,
    )

    expect(ausgang).toBeInstanceOf(Bindungsfehler)
    expect((ausgang as Bindungsfehler).code).toBe('unbekannt')
    expect((ausgang as Bindungsfehler).message).toBe('')
    expect((ausgang as Bindungsfehler).werte.vorgang).toMatch(
      /^[0-9a-f]{4}-[0-9a-f]{4}$/,
    )
  })

  // **Eine Marke, die sich nicht lesen lässt, ist eine Panik wie jede
  // andere** (F03, #2342) — bis dahin stand ihr Rohtext im Fenster.
  it('macht aus einer unlesbaren Marke einen Fehler »unbekannt« mit Kennung', async () => {
    const unlesbar = new Error(`Sets: panic: ${fehlermarke}{nicht json`)
    const ausgang: unknown = await gerufenMit(unlesbar).catch(
      (grund: unknown) => grund,
    )

    expect(ausgang).toBeInstanceOf(Bindungsfehler)
    expect((ausgang as Bindungsfehler).code).toBe('unbekannt')
    expect((ausgang as Bindungsfehler).werte.vorgang).toMatch(
      /^[0-9a-f]{4}-[0-9a-f]{4}$/,
    )
  })
})
