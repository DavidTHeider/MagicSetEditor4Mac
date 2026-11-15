// **Zwei Ablagen kurz hintereinander** —
// #891,
// Befund 4 des Zweitblicks über Phase Y.
//
// # Warum das eine eigene Datei ist
//
// `menueablage.test.tsx` fragt, **was die Einträge der Ablage tun**. Hier
// steht die andere Frage: **wem die Anzeige gehört**, wenn zwei Gesten
// gleichzeitig unterwegs sind. Der Schnitt folgt der Frage (§2), und ohne ihn
// liefe die Datei über 400 Zeilen.
//
// # Warum ein Lauf über `App`
//
// Zwischen dem Ablegen und dem, was danach dasteht, liegen vier Stellen: die
// Wails-Laufzeit, `useAbgelegteDateien`, `useDateioeffnung` und die Seite. Ein
// Lauf gegen den Haken allein sagte über die Reihenfolge nichts — und genau
// die ist hier die Zusage.
import { render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { effekteNachholen } from './test/effekte'
import { kataloge } from './sprache/kataloge'

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

const texte = kataloge.de.seiten

// Kartenzeilen ist der Beleg dafür, **welches** Set aufgeschlagen ist: die
// Kartenseite lädt seine Zeilen, sobald sie dasteht.
let Kartenzeilen = vi.fn()

// aufgeschlagen gibt die Kennung des zuletzt geladenen Sets.
function aufgeschlagen(): string | undefined {
  const rufe = Kartenzeilen.mock.calls

  return rufe.length === 0 ? undefined : (rufe[rufe.length - 1]?.[0] as string)
}

// geladeneSets gibt jede Kennung, für die je Zeilen geladen wurden.
function geladeneSets(): string[] {
  return Kartenzeilen.mock.calls.map((ruf) => String(ruf[0]))
}

// gelesenesSet ist die Antwort, die ein geglücktes Einlesen gibt.
function gelesenesSet(setID: string) {
  return {
    abgebrochen: false,
    setID,
    karten: 3,
    befunde: [],
    angelegt: [setID],
    unvollstaendig: false,
    schonOffen: false,
    fehler: null,
  }
}

// bereit stellt eine Laufzeit, zeichnet und gibt die angemeldeten Rückrufe.
async function bereit(
  zusatz: {
    SetAusDatei?: unknown
    ablegen?: (
      rueckruf: (x: number, y: number, pfade: string[]) => void,
    ) => void
  } = {},
) {
  const horcher = new Map<string, (...daten: unknown[]) => void>()

  ;(globalThis as MitLaufzeit).runtime = {
    EventsOn: (name, rueckruf) => {
      horcher.set(name, rueckruf)

      return () => undefined
    },

    // **Auch das Ablegen gehört zur Laufzeit** (#891).
    OnFileDrop: zusatz.ablegen ?? (() => undefined),
    OnFileDropOff: () => undefined,
  }

  // **Wessen Karten geladen werden, sagt, welches Set offen ist.**
  Kartenzeilen = vi.fn().mockResolvedValue({ zeilen: [], fehler: null })

  // **Öffnen geht über die Vorschau** (FMT-03, #1692) — ein abgebrochener
  // Dialog gibt keine.
  const SetVorschauWaehlen = vi.fn().mockResolvedValue({
    abgebrochen: true,
    vorschau: null,
    fehler: null,
  })

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
    SetVorschauWaehlen,
    Kartenzeilen,
    ...(zusatz.SetAusDatei === undefined
      ? {}
      : { SetAusDatei: zusatz.SetAusDatei as never }),
  })

  render(<App />)

  await screen.findByRole('heading', { name: texte.start.ueberschrift })
  await effekteNachholen()

  return { horcher, SetVorschauWaehlen }
}

describe('Zwei Ablagen kurz hintereinander', () => {
  // **Zwei Ablagen kurz hintereinander: die jüngere gewinnt** — Befund 4 des
  // Zweitblicks über Phase Y.
  //
  // Jedes Ablegen stößt einen eigenen `SetAusDatei` an, und zwei Ereignisse
  // kommen einzeln aus Wails. Bis zum 2026-08-25 gewann, wessen Antwort
  // **zuletzt** eintraf — bei einem großen Set zuerst und einem kleinen gleich
  // darauf also das große, obwohl niemand es mehr gemeint hat.
  //
  // **Geschrieben ist geschrieben** (§7, Regel 3): beide Sets liegen danach im
  // Bestand. Nur die Anzeige gehört der jüngeren Geste.
  it('lassen die überholte Ablage nicht mehr aufschlagen', async () => {
    const langsam = { ...gelesenesSet('7'), setID: '7' }
    const schnell = { ...gelesenesSet('9'), setID: '9' }

    let spaeterAntworten: (() => void) | undefined

    const SetAusDatei = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((fertig) => {
            spaeterAntworten = () => {
              fertig(langsam)
            }
          }),
      )
      .mockResolvedValueOnce(schnell)
      .mockResolvedValue(langsam)

    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined

    await bereit({
      SetAusDatei,
      ablegen: (rueckruf) => {
        ablegen = rueckruf
      },
    })

    ablegen?.(0, 0, ['/a/gross.mse-set'])
    ablegen?.(0, 0, ['/a/klein.mse-set'])

    await waitFor(() => {
      expect(SetAusDatei).toHaveBeenCalledTimes(2)
    })

    await waitFor(() => {
      expect(aufgeschlagen()).toBe('9')
    })

    // Erst jetzt antwortet der **erste** Lauf — überholt.
    spaeterAntworten?.()
    await effekteNachholen()

    // **Und er trägt das Fenster nicht zurück.**
    //
    // **Die Verneinung steht nach einem Beleg, dass jemand zuhört** (§3): der
    // dritte Zug unten geht denselben Weg und kommt an. Ohne ihn wäre dieser
    // Test auch grün, wenn `SetAusDatei` gar nichts mehr auslöste.
    expect(geladeneSets()).not.toContain('7')

    // **Der Beleg.** Eine dritte, **nicht** überholte Ablage auf dasselbe Set
    // schlägt es sehr wohl auf.
    ablegen?.(0, 0, ['/a/gross.mse-set'])

    await waitFor(() => {
      expect(aufgeschlagen()).toBe('7')
    })
  })

  // **Ein überholter Fehler wird trotzdem gemeldet** —
  // #1110.
  //
  // # Hier stand das Gegenteil, und es war zu teuer
  //
  // Bis zum 2026-08-25 verschluckte der Laufstand auch den **Fehler** einer
  // überholten Ablage. Der Grund war gut: „Set nicht lesbar" zu einem Set, das
  // niemand mehr meint, sah aus wie ein Fehler des aktuellen.
  //
  // **Er trägt nicht mehr**, seit die Meldung ihren Pfad nennt — und er kostete
  // den Fall, für den es #1110 gibt: wer **zwei Sets zusammen** zieht, erzeugt
  // zwei Läufe, und der erste ist immer überholt. Eine Datei, die nicht
  // aufging, verschwand **wortlos**.
  //
  // Der Laufstand entscheidet weiterhin, **was aufschlägt** — der Lauf darüber
  // hält das. Ob etwas schiefging, entscheidet er nicht mehr.
  it('melden den überholten Fehler trotzdem, mit seinem Pfad', async () => {
    let spaeterScheitern: (() => void) | undefined

    const SetAusDatei = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((_fertig, ablehnen) => {
            spaeterScheitern = () => {
              ablehnen(new Error('Die Datei ist leer.'))
            }
          }),
      )
      .mockResolvedValue({ ...gelesenesSet('9'), setID: '9' })

    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined

    await bereit({
      SetAusDatei,
      ablegen: (rueckruf) => {
        ablegen = rueckruf
      },
    })

    ablegen?.(0, 0, ['/a/kaputt.mse-set'])
    ablegen?.(0, 0, ['/a/heil.mse-set'])

    await waitFor(() => {
      expect(SetAusDatei).toHaveBeenCalledTimes(2)
    })

    // **Erst wartet das heile Set ab**, sonst prüft die Verneinung unten einen
    // Stand, in dem noch gar nichts geschehen ist.
    await waitFor(() => {
      expect(aufgeschlagen()).toBe('9')
    })

    spaeterScheitern?.()
    await effekteNachholen()

    // **Und der überholte Fehler kommt an** — mit seinem Pfad, sodass niemand
    // ihn für den des offenen Sets hält.
    expect(await screen.findByText(/Die Datei ist leer/)).toBeInTheDocument()
    expect(document.body.textContent).toContain('/a/kaputt.mse-set')

    // **Und das offene Set bleibt offen** (§7, Regel 3): der Fehler meldet
    // sich, er trägt das Fenster nicht zurück.
    expect(aufgeschlagen()).toBe('9')
  })
})
