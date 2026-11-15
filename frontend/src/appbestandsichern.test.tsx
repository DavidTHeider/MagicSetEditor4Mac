// *Ablage → Bestand sichern …* über die ganze Anwendung
// (#799).
//
// **Eigene Datei**, weil `appsichern.test.tsx` daneben eine andere Frage
// beantwortet: dort geht ein **Set** in eine `.mse-set`-Datei, hier wird die
// **Datenbank** weggetragen. Zwei Gegenstände, zwei Meldungen.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { ereignisBestandSichern } from './fenster'
import { kataloge } from './sprache/kataloge'
import {
  antwortAbwarten,
  belegeApi,
  belegeFenster,
  raeumeApi,
} from './test/api'
import { effekteNachholen } from './test/effekte'

interface MitLaufzeit {
  runtime?: {
    EventsOn?: (name: string, rueckruf: () => void) => () => void
  }
}

afterEach(() => {
  raeumeApi()
  delete (globalThis as MitLaufzeit).runtime
})

// band gibt den Bereich, in dem die Meldungen des Fensters stehen.
function band(): Promise<HTMLElement> {
  return screen.findByRole('status', {
    name: kataloge.de.allgemein.meldungsband.bereich,
  })
}

// menueAusloesen stellt eine Wails-Laufzeit bereit und feuert den Eintrag.
//
// **Über die Laufzeit und nicht über den Haken unmittelbar**: geprüft werden
// soll der Weg, den der Menüeintrag nimmt — dass `App` ihn verdrahtet, nicht
// dass ein Haken für sich funktioniert.
async function menueAusloesen() {
  const horcher = new Map<string, () => void>()

  ;(globalThis as MitLaufzeit).runtime = {
    EventsOn: (name, rueckruf) => {
      horcher.set(name, rueckruf)

      return () => undefined
    },
  }

  render(<App />)

  // **Der Horcher hängt erst nach einem Effekt** (§3, #761): `findByRole` sagt
  // zu, dass ein Knoten im DOM steht — nicht, dass ein `useEffect` gelaufen
  // ist. Ohne das Nachholen fiele das Ereignis ins Leere, und der Test wäre
  // grün, ohne etwas ausgelöst zu haben.
  await band()
  await effekteNachholen()

  horcher.get(ereignisBestandSichern)?.()
}

describe('Bestand sichern', () => {
  it('nennt den Ort, an dem die Sicherung liegt', async () => {
    const BestandSichern = vi.fn(() =>
      Promise.resolve({ pfad: '/Users/probe/mse-sicherung.db' }),
    )

    belegeApi({
      ...belegeFenster('de'),
      Sets: () => Promise.resolve({ sets: [] }),
      BestandSichern,
    })

    await menueAusloesen()

    await waitFor(() => {
      expect(BestandSichern).toHaveBeenCalledOnce()
      // **Aus dem Menü ohne Kennwort** (OPS-06) — das Kennwort steht auf
      // der Seite, nicht hinter einem Menüeintrag.
      expect(BestandSichern).toHaveBeenCalledWith('', false)
    })

    // **Und danach auf die Wirkung, nicht nur auf den Ruf** (§3, #1321).
    //
    // Hier stand die Zusicherung unmittelbar unter dem `waitFor`, und
    // `band()` ist kein Warten darauf: der Meldungsbereich steht **immer**
    // da, `findByRole` findet ihn sofort, und der Text kommt eine
    // Makroaufgabe später. Unter Last war einer von 3.931 Läufen rot.
    await waitFor(async () => {
      expect(await band()).toHaveTextContent('/Users/probe/mse-sicherung.db')
    })
  })

  it('meldet einen Fehler, statt ihn zu verschlucken', async () => {
    belegeApi({
      ...belegeFenster('de'),
      BestandSichern: () =>
        Promise.resolve({
          pfad: '',
          fehler: {
            code: 'validierung' as const,
            schluessel: 'fehler.validierung',
            meldung: 'Die Datei liegt schon da.',
          },
        }),
    })

    await menueAusloesen()

    await waitFor(async () => {
      expect(await band()).toHaveTextContent('Die Datei liegt schon da.')
    })
  })

  // **Und die Meldung lässt sich wegklicken.**
  //
  // Der Knopf am Band räumt beide Meldungen — die des Sets und die des
  // Bestands. Welche gerade dasteht, weiß er nicht, und er soll es nicht wissen
  // müssen; was nicht dasteht, verschwindet nicht.
  it('räumt die Meldung, wenn jemand sie schließt', async () => {
    belegeApi({
      ...belegeFenster('de'),
      BestandSichern: () => Promise.resolve({ pfad: '/tmp/s.db' }),
    })

    await menueAusloesen()

    await waitFor(async () => {
      expect(await band()).toHaveTextContent('/tmp/s.db')
    })

    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.bausteine.aktionen.schliessen,
      }),
    )

    expect(await band()).toBeEmptyDOMElement()
  })

  // **Ein Abbruch hinterlässt nichts** — kein Pfad, keine Meldung.
  //
  // Ohne diesen Fall wäre „Abbrechen" von „geglückt" nicht zu unterscheiden:
  // beide kommen ohne Fehler zurück.
  it('sagt nichts, wenn der Dialog abgebrochen wurde', async () => {
    const BestandSichern = vi.fn(() => Promise.resolve({ pfad: '' }))

    belegeApi({
      ...belegeFenster('de'),
      Sets: () => Promise.resolve({ sets: [] }),
      BestandSichern,
    })

    await menueAusloesen()

    await waitFor(() => {
      expect(BestandSichern).toHaveBeenCalledOnce()
    })

    // **Eine Verneinung wartet die Antwort ausdrücklich ab** (§3).
    //
    // Es gibt keine Wirkung, auf die man warten könnte — das ist die
    // Aussage. `band()` ist dafür kein Anker: der Bereich steht immer da.
    //
    // **Was die Gegenprobe zeigt und was nicht.** Bricht man den Schutz im
    // Haken auf, sodass ein leerer Pfad doch gemeldet wird, wird dieser
    // Test rot — **auch ohne** diese Zeile, weil `findByRole` seine eine
    // Makroaufgabe Frist gibt und die Vorrichtung genau so lange braucht.
    // Sie steht hier trotzdem: **genau diese Frist reicht unter Last
    // nicht**, gemessen in #761, und dann wäre der Test grün, ohne die
    // Antwort gesehen zu haben.
    await antwortAbwarten()

    expect(await band()).toBeEmptyDOMElement()
  })
})
