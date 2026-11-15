// **Der Weg durch die Anwendung: eine neuere Fassung wird gemeldet und geholt**
// (#1100).
//
// Die Zusagen des Hakens stehen in `haken/useAktualisierung.test.tsx`. Hier
// steht die eine, die dort nicht herzustellen ist: **jemand ruft ihn**. Stufe 1
// stand vor diesem Werkstück als Verfahren auf `*App` da, war getestet, hatte
// volle Abdeckung — und niemand im Fenster erreichte sie (§10, erste
// Dimension).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import type { LadeAntwort } from './api'
import { effekteNachholen } from './test/effekte'
import { einsetzen } from './sprache/einsetzen'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'

const band = kataloge.de.allgemein.fassungsband

// horcher hält die Rückrufe, die das Fenster bei der Laufzeit anmeldet.
const horcher = new Map<string, (...daten: unknown[]) => void>()

const neuere = {
  gefragt: true,
  neuer: true,
  fassung: '9.9.9',
  seite: 'https://beispiel.test/v9',
  holen: true,
  fehler: null,
}

beforeEach(() => {
  horcher.clear()
  localStorage.clear()
  ;(globalThis as { runtime?: unknown }).runtime = {
    EventsOn: (name: string, rueckruf: (...daten: unknown[]) => void) => {
      horcher.set(name, rueckruf)

      return () => horcher.delete(name)
    },
  }
})

afterEach(() => {
  raeumeApi()
  delete (globalThis as { runtime?: unknown }).runtime
})

describe('Eine neuere Fassung im Fenster', () => {
  it('steht im Band, mit dem Weg zu dem, was sich geändert hat', async () => {
    belegeApi({
      ...belegeFenster('de'),
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
    })

    render(<App />)

    await screen.findByText(einsetzen(band.neu, { fassung: '9.9.9' }))

    expect(screen.getByRole('link', { name: band.seite })).toHaveAttribute(
      'href',
      'https://beispiel.test/v9',
    )
  })

  // **Unter „melden“ kein Ladeknopf** — gefunden am 2026-10-17 beim
  // Handbuchbild (AP-12, #2369): das Band bot „Herunterladen“ an, und Go wies
  // es ab („die Aktualisierung steht nicht auf holen“). Ob geladen werden
  // darf, sagt die Antwort.
  it('meldet unter „melden“ nur, ohne Ladeknopf', async () => {
    belegeApi({
      ...belegeFenster('de'),
      NachNeuerFassungSehen: () => Promise.resolve({ ...neuere, holen: false }),
    })

    render(<App />)

    await screen.findByText(einsetzen(band.neu, { fassung: '9.9.9' }))

    expect(screen.getByRole('link', { name: band.seite })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: band.laden })).toBeNull()
  })

  it('bleibt still, wenn es nichts Neues gibt', async () => {
    const NachNeuerFassungSehen = vi.fn().mockResolvedValue({
      gefragt: true,
      neuer: false,
      fassung: '0.1.0',
      seite: '',
      holen: true,
      fehler: null,
    })

    belegeApi({ ...belegeFenster('de'), NachNeuerFassungSehen })

    render(<App />)

    await waitFor(() => {
      expect(NachNeuerFassungSehen).toHaveBeenCalled()
    })

    expect(screen.queryByRole('button', { name: band.laden })).toBeNull()
  })

  it('lädt auf Knopfdruck und nennt danach den Ort', async () => {
    const NeueFassungLaden = vi.fn().mockResolvedValue({
      neuer: true,
      fassung: '9.9.9',
      pfad: '/Users/x/Downloads/neu.dmg',
      fehler: null,
    })

    belegeApi({
      ...belegeFenster('de'),
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      NeueFassungLaden,
    })

    render(<App />)

    const knopf = await screen.findByRole('button', { name: band.laden })

    await userEvent.click(knopf)

    await screen.findByText(
      einsetzen(band.geladen, {
        fassung: '9.9.9',
        pfad: '/Users/x/Downloads/neu.dmg',
      }),
    )

    // **Geladen heißt geladen**: der Knopf verschwindet, statt ein zweites Mal
    // 188 MB anzubieten.
    expect(screen.queryByRole('button', { name: band.laden })).toBeNull()
  })

  it('zeigt den Fortschritt, während geladen wird', async () => {
    let loesen: (wert: LadeAntwort) => void = () => undefined

    belegeApi({
      ...belegeFenster('de'),
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      NeueFassungLaden: () =>
        new Promise((ja) => {
          loesen = ja
        }),
    })

    render(<App />)

    await userEvent.click(
      await screen.findByRole('button', { name: band.laden }),
    )

    // **Erst der Beleg, dass jemand zuhört** (§3): der Horcher hängt in einem
    // Effekt, und `findByRole` sagt über den nichts zu.
    await effekteNachholen()

    horcher.get('aktualisierung:fortschritt')?.(1 << 20, 2 << 20)

    await screen.findByText(
      einsetzen(band.fortschritt, { gelesen: '1.0 MiB', ganz: '2.0 MiB' }),
    )

    loesen({ neuer: true, fassung: '9.9.9', pfad: '/x/neu.dmg', fehler: null })

    await screen.findByText(
      einsetzen(band.geladen, { fassung: '9.9.9', pfad: '/x/neu.dmg' }),
    )
  })

  // **Vor dem ersten Block steht keine Zahl da.**
  //
  // Gefunden bei der Sichtprüfung am laufenden Fenster, nicht von einem Test:
  // dort stand einen Augenblick lang „0,0 MiB geladen" — eine Messung, die
  // nicht stattgefunden hatte (§10). Dass etwas läuft, sagt die Schaltfläche.
  it('zeigt keine Menge, bevor der erste Block da ist', async () => {
    let loesen: (wert: LadeAntwort) => void = () => undefined

    belegeApi({
      ...belegeFenster('de'),
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      NeueFassungLaden: () =>
        new Promise((ja) => {
          loesen = ja
        }),
    })

    render(<App />)

    await userEvent.click(
      await screen.findByRole('button', { name: band.laden }),
    )
    await effekteNachholen()

    await screen.findByRole('button', { name: band.laeuft })

    expect(
      screen.queryByText(
        einsetzen(band.fortschrittOhneLaenge, { gelesen: '0.0 MiB' }),
      ),
    ).toBeNull()

    // **Und der Beleg, dass der Horcher wirklich hängt** — sonst prüfte diese
    // Verneinung ein Ereignis, das niemand gehört hat (§3).
    horcher.get('aktualisierung:fortschritt')?.(1 << 20, -1)

    await screen.findByText(
      einsetzen(band.fortschrittOhneLaenge, { gelesen: '1.0 MiB' }),
    )

    loesen({ neuer: true, fassung: '9.9.9', pfad: '/x/neu.dmg', fehler: null })

    await screen.findByText(
      einsetzen(band.geladen, { fassung: '9.9.9', pfad: '/x/neu.dmg' }),
    )
  })

  it('zeigt die geladene Menge, wenn die Länge fehlt', async () => {
    let loesen: (wert: LadeAntwort) => void = () => undefined

    belegeApi({
      ...belegeFenster('de'),
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      NeueFassungLaden: () =>
        new Promise((ja) => {
          loesen = ja
        }),
    })

    render(<App />)

    await userEvent.click(
      await screen.findByRole('button', { name: band.laden }),
    )
    await effekteNachholen()

    horcher.get('aktualisierung:fortschritt')?.(3 << 20, -1)

    await screen.findByText(
      einsetzen(band.fortschrittOhneLaenge, { gelesen: '3.0 MiB' }),
    )

    loesen({ neuer: true, fassung: '9.9.9', pfad: '/x/neu.dmg', fehler: null })

    await screen.findByText(
      einsetzen(band.geladen, { fassung: '9.9.9', pfad: '/x/neu.dmg' }),
    )
  })

  it('zeigt einen Fehler, statt ihn zu verschlucken', async () => {
    belegeApi({
      ...belegeFenster('de'),
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      NeueFassungLaden: () =>
        Promise.resolve({
          neuer: true,
          fassung: '9.9.9',
          pfad: '',
          fehler: {
            code: 'konflikt',
            meldung: 'die Prüfsumme stimmt nicht',
            schluessel: 'pruefsumme',
          },
        }),
    })

    render(<App />)

    await userEvent.click(
      await screen.findByRole('button', { name: band.laden }),
    )

    await screen.findByText('die Prüfsumme stimmt nicht')
  })

  it('geht weg und kommt für dieselbe Fassung nicht wieder', async () => {
    belegeApi({
      ...belegeFenster('de'),
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
    })

    const erste = render(<App />)

    await userEvent.click(
      await screen.findByRole('button', { name: band.weglegen }),
    )

    expect(
      screen.queryByText(einsetzen(band.neu, { fassung: '9.9.9' })),
    ).toBeNull()

    erste.unmount()

    render(<App />)

    await waitFor(() => {
      expect(
        screen.getByRole('status', { name: band.bereich }),
      ).toBeEmptyDOMElement()
    })
  })

  it('zeigt ein Band ohne Freigabeseite ohne Verweis', async () => {
    belegeApi({
      ...belegeFenster('de'),
      NachNeuerFassungSehen: () => Promise.resolve({ ...neuere, seite: '' }),
    })

    render(<App />)

    await screen.findByText(einsetzen(band.neu, { fassung: '9.9.9' }))

    expect(screen.queryByRole('link', { name: band.seite })).toBeNull()
  })
})

// **Ein Ladeknopf für eine Fassung, die diesen Bestand nicht öffnen kann,
// wäre die Falle, gegen die es O6 gibt** (#1430).
//
// **Gefunden hat es die Sichtprüfung**, nicht ein Lauf: das Band sagte „Diese
// Fassung kann deinen Bestand nicht öffnen" und bot daneben „Herunterladen"
// an. Ein Satz und ein Knopf, die einander widersprechen — und der Knopf
// gewinnt, weil man ihn drückt.
