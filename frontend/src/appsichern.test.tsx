// ⌘S über die ganze Anwendung — und das Band, in dem seine Meldung landet.
//
// **Eigene Datei**, weil `App.test.tsx` sonst über die Zeilengrenze wüchse
// (AGENTS.md §2). Der Schnitt folgt der Frage: dort Sprache und Aufbau, hier
// das Sichern (#162)
// und die Meldung dazu (#180).
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { effekteNachholen } from './test/effekte'
import { leisteneintrag } from './test/leistenweg'

// probeset ist ein Set, wie die Liste es zeigt.
function probeset() {
  return {
    id: '7',
    name: 'Mein Probeset',
    spiel: 'probe',
    stil: 'schlicht',
    herkunft: '/pfad/probeset.mse-set',
    karten: 1,
    herkunftFehlt: false,
  }
}

// band gibt den Bereich, in dem die Meldungen des Fensters stehen.
//
// **`findByRole` und nicht `getByRole`** — seit #245 zeichnet `App` erst, wenn
// die Einstellungen da sind, und die kommen über die Fassade.
function band(): Promise<HTMLElement> {
  return screen.findByRole('status', {
    name: kataloge.de.allgemein.meldungsband.bereich,
  })
}

// waehleErstesSet klickt in der Setliste auf das erste Set.
async function waehleErstesSet(name: string) {
  await userEvent.click(await leisteneintrag('sets'))

  await userEvent.click(
    await screen.findByRole('button', {
      name: `${name} — ${kataloge.de.seiten.sets.oeffnen}`,
    }),
  )
}

beforeEach(() => {
  belegeApi({ ...belegeFenster('de') })
})

afterEach(() => {
  raeumeApi()
})

describe('⌘S über die Anwendung', () => {
  // **Ohne offenes Set tut ⌘S nichts** — so verlangt es #162, und es ist die
  // richtige Antwort: ein Fehler „kein Set gewählt" wäre eine Meldung über
  // etwas, das niemand versucht hat.
  //
  // # Warum erst ⌘F und dann ⌘S
  //
  // **`not.toHaveBeenCalled()` trifft auch dann zu, wenn niemand zugehört
  // hat** (#605).
  // Der Tastenhorcher hängt in einem `useEffect`; feuert die Taste davor,
  // geschieht **nichts** — der Test ist grün, und der Zweig `if (setID)` in
  // `App.tsx` wird nie betreten. Genau das war der Flacker: unter Last mal
  // 853 Zweige, mal 852.
  //
  // Nachgestellt am 2026-08-20, indem die Taste absichtlich vor `await band()`
  // gefeuert wurde — dieselbe Ausgabe wie im Befund: `App.tsx 94,44 %,
  // Zeile 61, Branches 852/853`.
  //
  // **⌘3 ist der Beleg, dass jemand zuhört**: es geht durch denselben Horcher
  // und setzt die Seite — sichtbar an `aria-current` in der Navigation.
  //
  // **Und nicht ⌘F**: die Suche gehört zu einem Set, und ohne offenes steht
  // sie gar nicht in der Leiste. Die Pakete stehen immer da.
  it('tut nichts, solange kein Set offen ist', async () => {
    const SetSpeichern = vi.fn()
    belegeApi({ ...belegeFenster('de'), SetSpeichern })

    render(<App />)

    // **Erst zeichnen lassen** — seit #245 kommt die Sprache aus dem Bestand,
    // und vorher gibt `App` nichts zurück.
    const meldungen = await band()

    // **Und dann die Effekte nachholen lassen**, sonst geht der Tastendruck
    // darunter verloren ([effekteNachholen], #761).
    await effekteNachholen()

    // **⌘6 seit UX-02**
    // (#1674):
    // die Ziffer gehört dem **Bereich**, und *Ressourcen* ist der sechste.
    // **Eine Seite ohne Set** ist hier der Punkt — die Karten wären eine mit.
    fireEvent.keyDown(window, { key: '6', metaKey: true })

    await waitFor(async () => {
      expect(await leisteneintrag('pakete')).toHaveAttribute(
        'aria-current',
        'page',
      )
    })

    fireEvent.keyDown(window, { key: 's', metaKey: true })

    expect(SetSpeichern).not.toHaveBeenCalled()
    expect(meldungen).toBeEmptyDOMElement()
  })

  it('sichert das offene Set und meldet, wohin', async () => {
    const SetSpeichern = vi.fn().mockResolvedValue({
      pfad: '/tmp/mein.mse-set',
      befunde: [],
      abgebrochen: false,
    })

    belegeApi({
      ...belegeFenster('de'),
      Sets: vi.fn().mockResolvedValue({ sets: [probeset()] }),
      SetSpeichern,
      Karten: vi.fn().mockResolvedValue({ karten: [] }),
    })

    render(<App />)

    await waehleErstesSet('Mein Probeset')

    fireEvent.keyDown(window, { key: 's', metaKey: true })

    await waitFor(() => {
      expect(SetSpeichern).toHaveBeenCalledWith('7')
    })

    // **Die Meldung steht im Band und nicht auf der Kartenseite** — sie gehört
    // zu einer Handlung des Fensters, nicht zu einer Seite (#180).
    await waitFor(async () => {
      expect(await band()).toHaveTextContent('/tmp/mein.mse-set')
    })
  })

  it('sichert mit ⇧⌘S unter einem anderen Namen', async () => {
    const SetSpeichernUnter = vi.fn().mockResolvedValue({
      pfad: '/tmp/anders.mse-set',
      befunde: [],
      abgebrochen: false,
    })

    belegeApi({
      ...belegeFenster('de'),
      Sets: vi.fn().mockResolvedValue({ sets: [probeset()] }),
      SetSpeichernUnter,
      Karten: vi.fn().mockResolvedValue({ karten: [] }),
    })

    render(<App />)

    await waehleErstesSet('Mein Probeset')

    fireEvent.keyDown(window, { key: 's', metaKey: true, shiftKey: true })

    await waitFor(() => {
      expect(SetSpeichernUnter).toHaveBeenCalledWith('7')
    })
  })
})

interface MitLaufzeit {
  runtime?: { EventsOn: (name: string, rueckruf: () => void) => () => void }
}

describe('die native Menüleiste', () => {
  afterEach(() => {
    delete (globalThis as MitLaufzeit).runtime
  })

  // **Der zweite Weg zu denselben Handlungen** (#162): im Fenster gehen die
  // Kürzel über `tastatur.ts`, die Leiste meldet Ereignisse. Beide landen in
  // derselben Funktion — sonst täte ⌘S aus dem Menü etwas anderes als ⌘S auf
  // der Tastatur, und niemand könnte sagen, warum.
  it('öffnet die Einstellungen über ihr Ereignis', async () => {
    const horcher = new Map<string, () => void>()

    ;(globalThis as MitLaufzeit).runtime = {
      EventsOn: (name, rueckruf) => {
        horcher.set(name, rueckruf)

        return () => undefined
      },
    }

    belegeApi({ ...belegeFenster('de') })
    render(<App />)

    // **Erst zeichnen lassen.** Der Horcher wird beim ersten Lauf gesetzt; ein
    // Aufruf davor liefe ins Leere.
    await screen.findByRole('heading', {
      name: kataloge.de.seiten.start.ueberschrift,
    })

    // **Dass die Überschrift dasteht, heißt nicht, dass der Effekt lief**
    // ([effekteNachholen], #761). Genau hier war einer der vier roten Läufe.
    await effekteNachholen()

    // **Ein fehlender Horcher wird laut.** `?.()` allein ließe ihn still
    // fallen, und der Test liefe stattdessen in die Zeitgrenze der Abfrage
    // darunter — fünf Sekunden für eine Meldung, die nicht sagt, was fehlt.
    const oeffnen = horcher.get('menue:einstellungen')

    expect(oeffnen).toBeDefined()
    oeffnen?.()

    expect(
      await screen.findByRole('heading', {
        name: kataloge.de.seiten.einstellungen.titel,
      }),
    ).toBeInTheDocument()
  })
})

describe('das Meldungsband', () => {
  // **Es lässt sich schließen** (#180) — eine Meldung, die stehen bleibt, bis
  // die nächste kommt, ist nach der dritten Handlung eine Tapete.
  it('räumt die Meldung auf Wunsch weg', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Sets: vi.fn().mockResolvedValue({ sets: [probeset()] }),
      SetSpeichern: vi.fn().mockResolvedValue({
        pfad: '/tmp/mein.mse-set',
        befunde: [],
        abgebrochen: false,
      }),
      Karten: vi.fn().mockResolvedValue({ karten: [] }),
    })

    render(<App />)

    await waehleErstesSet('Mein Probeset')

    fireEvent.keyDown(window, { key: 's', metaKey: true })

    await waitFor(async () => {
      expect(await band()).toHaveTextContent('/tmp/mein.mse-set')
    })

    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.bausteine.aktionen.schliessen,
      }),
    )

    await waitFor(async () => {
      expect(await band()).toBeEmptyDOMElement()
    })
  })
})

describe('Meldungen, die zu keiner Seite gehören', () => {
  // **Die Zusage aus #180**: eine gescheiterte Rücknahme ist sichtbar, gleich
  // auf welcher Seite. Der Test steht deshalb ausdrücklich **nicht** auf der
  // Kartenliste — dort war sie schon vorher zu sehen, und genau das war der
  // Befund: überall sonst sah man nichts.
  it('zeigt eine gescheiterte Rücknahme auch fern der Kartenseite', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Sets: vi.fn().mockResolvedValue({ sets: [probeset()] }),
      Karten: vi.fn().mockResolvedValue({ karten: [] }),
      Zuruecknehmen: vi
        .fn()
        .mockRejectedValue(new Error('Die Platte ist voll.')),
    })

    render(<App />)

    await waehleErstesSet('Mein Probeset')

    // Fort von der Kartenseite — auf die Statistik. **Seit UX-08 in der
    // Schublade „Werkzeuge"** (#1680), die bei der Statistik aufgeht.
    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.allgemein.arbeitsplatz.schubladen.werkzeuge.knopf,
      }),
    )
    await screen.findByRole('tab', {
      name: kataloge.de.seiten.statistik.titel,
      selected: true,
    })

    fireEvent.keyDown(window, { key: 'z', metaKey: true })

    await waitFor(async () => {
      expect(await band()).toHaveTextContent('Die Platte ist voll.')
    })
  })

  it('zeigt ein gescheitertes Sichern', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Sets: vi.fn().mockResolvedValue({ sets: [probeset()] }),
      Karten: vi.fn().mockResolvedValue({ karten: [] }),
      SetSpeichern: vi.fn().mockRejectedValue(new Error('Kein Schreibrecht.')),
    })

    render(<App />)

    await waehleErstesSet('Mein Probeset')

    fireEvent.keyDown(window, { key: 's', metaKey: true })

    await waitFor(async () => {
      expect(await band()).toHaveTextContent('Kein Schreibrecht.')
    })
  })
})
