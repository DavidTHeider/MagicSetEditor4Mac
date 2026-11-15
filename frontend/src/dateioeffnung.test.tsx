// **Ein Doppelklick im Finder, durch die ganze Anwendung** —
// #891.
//
// # Warum das ein Lauf über `App` ist
//
// Der Weg geht über **vier** Stellen: das Ereignis aus Go, die Prüfung der
// Nutzlast in `fenster.ts`, den Haken, der daraus einen Satz macht, und die
// Seite, die danach dasteht. Ein Lauf gegen den Haken allein sagte über die
// drei anderen nichts — und die Naht dazwischen ist genau der Ort, an dem
// dieser Weg bisher gar nicht existierte.
import { render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { effekteNachholen } from './test/effekte'
import { kataloge } from './sprache/kataloge'
import { einsetzen } from './sprache/einsetzen'

interface MitLaufzeit {
  runtime?: {
    EventsOn: (
      name: string,
      rueckruf: (...daten: unknown[]) => void,
    ) => () => void
  }
}

afterEach(() => {
  delete (globalThis as MitLaufzeit).runtime
  raeumeApi()
})

// mitHorchern stellt eine Laufzeit und gibt die angemeldeten Rückrufe.
function mitHorchern() {
  const horcher = new Map<string, (...daten: unknown[]) => void>()

  ;(globalThis as MitLaufzeit).runtime = {
    EventsOn: (name, rueckruf) => {
      horcher.set(name, rueckruf)

      return () => undefined
    },
  }

  return horcher
}

// bereit zeichnet die Anwendung und wartet, bis die Horcher hängen.
async function bereit(horcher: Map<string, (...daten: unknown[]) => void>) {
  render(<App />)

  await screen.findByRole('heading', {
    name: kataloge.de.seiten.start.ueberschrift,
  })

  // **Dass die Überschrift dasteht, heißt nicht, dass der Effekt lief**
  // ([effekteNachholen], #761).
  await effekteNachholen()

  const gemeldet = horcher.get('datei:geoeffnet')
  const gescheitert = horcher.get('datei:fehler')

  if (gemeldet === undefined || gescheitert === undefined) {
    throw new Error('die Horcher für die Datei hängen nicht')
  }

  return { gemeldet, gescheitert }
}

describe('Eine Datei aus dem Finder', () => {
  it('schlägt das Set auf', async () => {
    const horcher = mitHorchern()

    belegeApi({
      ...belegeFenster('de'),
      Startbefund: vi.fn().mockResolvedValue({ offen: true, grund: '' }),
      Sets: vi.fn().mockResolvedValue({
        sets: [{ id: '7', name: 'Probe', spiel: 'probe', karten: 1 }],
        fehler: null,
      }),
      Karten: vi.fn().mockResolvedValue({ karten: [], fehler: null }),
      Kartenfelder: vi.fn().mockResolvedValue({ felder: [], fehler: null }),
      Setangaben: vi
        .fn()
        .mockResolvedValue({ felder: [], werte: [], fehler: null }),
    })

    const { gemeldet } = await bereit(horcher)

    gemeldet('7')

    // **Die Kartenseite steht da**, und zwar für das gemeldete Set.
    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: kataloge.de.seiten.karten.anlegen }),
      ).toBeInTheDocument()
    })
  })

  it('sagt, wenn sie sich nicht öffnen ließ', async () => {
    const horcher = mitHorchern()

    belegeApi({
      ...belegeFenster('de'),
      Startbefund: vi.fn().mockResolvedValue({ offen: true, grund: '' }),
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
    })

    const { gescheitert } = await bereit(horcher)

    gescheitert('/pfad/Weg.mse-set', {
      schluessel: 'nichtGefunden',
      meldung: 'Es ist ein unerwarteter Fehler aufgetreten.',
      werte: null,
    })

    // **Der Satz entsteht im Fenster** — aus dem Pfad und dem übersetzten
    // Schlüssel, nicht aus dem deutschen Rückfallsatz aus Go (§9).
    const erwartet = einsetzen(kataloge.de.allgemein.dateiNichtGeoeffnet, {
      pfad: '/pfad/Weg.mse-set',
      grund: kataloge.de.fehler.nichtGefunden,
    })

    expect(await screen.findByText(erwartet)).toBeInTheDocument()

    // **Ein Schlüssel, den der Katalog nicht kennt, fällt auf den Satz aus Go
    // zurück** (§9). Eine unübersetzte Meldung ist besser als eine leere —
    // und sie steht hier, damit ein neuer Schlüssel aus Go nicht dazu führt,
    // dass gar nichts dasteht.
    gescheitert('/pfad/Neu.mse-set', {
      schluessel: 'einenSolchenSchluesselGibtEsNicht',
      meldung: 'Der Satz aus Go.',
      werte: {},
    })

    const zweiter = einsetzen(kataloge.de.allgemein.dateiNichtGeoeffnet, {
      pfad: '/pfad/Neu.mse-set',
      grund: 'Der Satz aus Go.',
    })

    // **Beide Sätze stehen da, nicht nur der zweite** —
    // #1110.
    //
    // Hier stand `findByText(zweiter)`, und das war die alte Zusage: der
    // jüngste Fehler ersetzt den vorigen. Bei zwei Dateien aus dem Finder
    // heißt das, dass der erste verschwindet, bevor jemand ihn gelesen hat.
    await waitFor(() => {
      expect(document.body.textContent).toContain(zweiter)
    })

    expect(document.body.textContent).toContain(erwartet)
  })

  // **Eine Nutzlast, die nicht passt, tut nichts.**
  //
  // Sie kommt über die Ereignisnaht als `unknown`. Ein blindes `as` machte aus
  // einer geänderten Go-Seite eine Seite, die stumm etwas anderes tut — hier
  // etwa eine Kartenseite ohne Set.
  it('lässt eine unpassende Nutzlast liegen', async () => {
    const horcher = mitHorchern()

    belegeApi({
      ...belegeFenster('de'),
      Startbefund: vi.fn().mockResolvedValue({ offen: true, grund: '' }),
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
    })

    const { gemeldet, gescheitert } = await bereit(horcher)

    gemeldet(42)
    gemeldet('')
    gescheitert(7, { schluessel: 'nichtGefunden', meldung: '', werte: null })
    gescheitert('/pfad/Weg.mse-set', 'kein Fehlerverbund')
    gescheitert('/pfad/Weg.mse-set', null)

    await effekteNachholen()

    // **Die Startseite steht noch da**, und es gibt keine Meldung.
    expect(
      screen.getByRole('heading', {
        name: kataloge.de.seiten.start.ueberschrift,
      }),
    ).toBeInTheDocument()

    // **Kein Satz über die Datei** — geprüft am Pfad und nicht an der Rolle:
    // das Band ist nicht das einzige `role="status"` der Startseite.
    expect(screen.queryByText(/Weg\.mse-set/)).not.toBeInTheDocument()

    // **Der Beleg, dass die Horcher wirklich hängen** (§3): dieselbe Naht
    // meldet, wenn die Nutzlast passt. Eine bloße Verneinung träfe auch zu,
    // wenn gar nichts angekommen wäre.
    gescheitert('/pfad/Weg.mse-set', {
      schluessel: 'nichtGefunden',
      meldung: 'Es ist ein unerwarteter Fehler aufgetreten.',
      werte: null,
    })

    expect(
      await screen.findByText(/Weg\.mse-set/, { exact: false }),
    ).toBeInTheDocument()
  })
})
