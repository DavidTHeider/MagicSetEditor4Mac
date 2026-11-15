// **Mehrere Dateien auf einmal, und ein Zug, der nirgends endet** —
// #1110,
// #1111 und
// #1112.
//
// Alle drei kommen aus dem Nachlauf des Zweitblicks über Phase Y, und alle
// drei haben dieselbe Wurzel: die Zusagen aus §7 gelten je Weg, nicht je
// Geste.
//
// Eigene Datei, weil `setziehen.test.tsx` mit ihnen über 400 Zeilen ginge
// (§2). Dort steht **wem die Anzeige gehört**, hier **was bei mehreren
// Dateien geschieht**.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
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

let Kartenzeilen = vi.fn()

// aufgeschlagen gibt die Kennung des zuletzt geladenen Sets.
function aufgeschlagen(): string | undefined {
  const rufe = Kartenzeilen.mock.calls

  return rufe.length === 0 ? undefined : (rufe[rufe.length - 1]?.[0] as string)
}

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

interface Seteintrag {
  id: string
  name: string
  spiel: string
  karten: number
}

async function bereit(zusatz: {
  SetAusDatei?: unknown
  ablegen?: (rueckruf: (x: number, y: number, pfade: string[]) => void) => void
  sets?: Seteintrag[]
}) {
  ;(globalThis as MitLaufzeit).runtime = {
    EventsOn: () => () => undefined,
    OnFileDrop: zusatz.ablegen ?? (() => undefined),
    OnFileDropOff: () => undefined,
  }

  Kartenzeilen = vi.fn().mockResolvedValue({ zeilen: [], fehler: null })

  belegeApi({
    ...belegeFenster('de'),
    Startbefund: vi.fn().mockResolvedValue({ offen: true, grund: '' }),
    Sets: vi.fn().mockResolvedValue({ sets: zusatz.sets ?? [], fehler: null }),
    // **Die Startseite listet, woran zuletzt gearbeitet wurde** (#1422) — die
    // Sets aus der Liste stehen dort als zuletzt geöffnet, wie vorher unter
    // „Zuletzt benutzt“.
    Weiterarbeiten: vi.fn().mockResolvedValue({
      zuletzt: null,
      geoeffnet: (zusatz.sets ?? []).map((set) => ({
        objekt: {
          art: 'set' as const,
          kennung: set.id,
          name: set.name,
          angeheftet: false,
        },
        vorgang: 'geoeffnet' as const,
        zeitpunkt: '2026-09-09T10:00:00Z',
      })),
      eingefuehrt: [],
      ausgegeben: [],
      angeheftet: [],
      fehler: null,
    }),
    Pakete: vi.fn().mockResolvedValue({ pakete: [], fehler: null }),
    Karten: vi.fn().mockResolvedValue({ karten: [], fehler: null }),
    Kartenfelder: vi.fn().mockResolvedValue({ felder: [], fehler: null }),
    Setangaben: vi
      .fn()
      .mockResolvedValue({ felder: [], werte: [], fehler: null }),
    Kartenzeilen,
    ...(zusatz.SetAusDatei === undefined
      ? {}
      : { SetAusDatei: zusatz.SetAusDatei as never }),
  })

  render(<App />)

  await screen.findByRole('heading', { name: texte.start.ueberschrift })
  await effekteNachholen()
}

describe('Mehrere Setdateien auf einmal', () => {
  // **Ein Fehler geht nicht verloren, weil eine andere Datei glückt** —
  // #1110.
  //
  // Wer drei Dateien zieht und bei einer davon einen Fehler bekommt, muss ihn
  // sehen — auch wenn die dritte aufgeht. Bis zum 2026-08-25 räumte **jeder**
  // Erfolg die Meldung, und der Fehler war fort, bevor jemand ihn gelesen hat.
  it('behalten den Fehler der einen, wenn die andere glückt', async () => {
    const SetAusDatei = vi
      .fn()
      .mockRejectedValueOnce(new Error('Die Datei ist leer.'))
      .mockResolvedValueOnce(gelesenesSet('9'))

    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined

    await bereit({
      SetAusDatei,
      ablegen: (rueckruf) => {
        ablegen = rueckruf
      },
    })

    ablegen?.(0, 0, ['/a/kaputt.mse-set', '/a/heil.mse-set'])

    await waitFor(() => {
      expect(SetAusDatei).toHaveBeenCalledTimes(2)
    })

    // **Erst der Beleg, dass die zweite Datei angekommen ist** — sonst prüft
    // die Zusicherung darunter einen Stand, in dem noch nichts geschehen ist
    // (§3).
    await waitFor(() => {
      expect(Kartenzeilen).toHaveBeenCalled()
    })

    // **Und der Fehler der ersten steht immer noch da.**
    expect(await screen.findByText(/Die Datei ist leer/)).toBeInTheDocument()
  })

  // **Eine Datei, die nach einem Fehlschlag doch aufgeht, räumt ihre eigene
  // Meldung** — und nur ihre.
  it('räumen die Meldung, wenn dieselbe Datei danach aufgeht', async () => {
    const SetAusDatei = vi
      .fn()
      .mockRejectedValueOnce(new Error('Die Datei ist leer.'))
      .mockResolvedValueOnce(gelesenesSet('9'))

    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined

    await bereit({
      SetAusDatei,
      ablegen: (rueckruf) => {
        ablegen = rueckruf
      },
    })

    ablegen?.(0, 0, ['/a/wackelig.mse-set'])

    expect(await screen.findByText(/Die Datei ist leer/)).toBeInTheDocument()

    // Derselbe Pfad, zweiter Versuch — diesmal geht er auf.
    ablegen?.(0, 0, ['/a/wackelig.mse-set'])

    await waitFor(() => {
      expect(screen.queryByText(/Die Datei ist leer/)).not.toBeInTheDocument()
    })
  })
})

describe('Der Laufstand gilt für jeden Weg', () => {
  // **Ein Klick in der Liste überholt eine laufende Ablage** —
  // #1111.
  //
  // Der Laufstand stieg bis zum 2026-08-25 **nur** in `dateiGezogen`. Wer ein
  // großes Set zog und währenddessen ein anderes aus der Liste wählte, bekam
  // am Ende das gezogene — die späte Antwort galt weiter als aktuell und
  // schlug ihr Set über dem auf, das der Mensch zuletzt gemeint hat.
  //
  // §7 Regel 2 spricht vom **Bezugsobjekt**, nicht von einem Weg dorthin.
  it('lässt eine späte Ablage das gewählte Set nicht überschreiben', async () => {
    let spaeterAntworten: (() => void) | undefined

    const SetAusDatei = vi.fn().mockImplementationOnce(
      () =>
        new Promise((fertig) => {
          spaeterAntworten = () => {
            fertig(gelesenesSet('7'))
          }
        }),
    )

    let ablegen: ((x: number, y: number, pfade: string[]) => void) | undefined

    await bereit({
      SetAusDatei,
      ablegen: (rueckruf) => {
        ablegen = rueckruf
      },
      sets: [{ id: '9', name: 'Aus der Liste', spiel: 'magic', karten: 2 }],
    })

    ablegen?.(0, 0, ['/a/gross.mse-set'])

    await waitFor(() => {
      expect(SetAusDatei).toHaveBeenCalled()
    })

    // **Während der Zug läuft, wählt jemand ein Set aus der Liste.**
    await userEvent.click(
      // Der Knopf, der aufschlägt — nicht der daneben, der anheftet.
      await screen.findByRole('button', { name: /^Aus der Liste/ }),
    )

    await waitFor(() => {
      expect(aufgeschlagen()).toBe('9')
    })

    // Erst jetzt antwortet die Ablage — überholt.
    spaeterAntworten?.()
    await effekteNachholen()

    // **Und sie trägt das Fenster nicht zurück.**
    expect(aufgeschlagen()).toBe('9')
  })
})
