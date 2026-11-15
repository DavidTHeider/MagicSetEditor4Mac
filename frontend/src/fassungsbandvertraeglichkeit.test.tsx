// **Das Band und das Verträglichkeitsurteil** — Werkstück O6 von Phase 5.2
// (#1430).
//
// **Eine eigene Datei**, weil `fassungsband.test.tsx` mit ihr über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort steht, dass
// eine neuere Fassung **gemeldet und geholt** wird, hier, dass sie **zu
// diesem Bestand passen muss**, bevor jemand sie holt.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'

const band = kataloge.de.allgemein.fassungsband

const neuere = {
  gefragt: true,
  neuer: true,
  fassung: '9.9.9',
  seite: 'https://beispiel.test/v9',
  holen: true,
  fehler: null,
}

beforeEach(() => {
  localStorage.clear()
  ;(globalThis as { runtime?: unknown }).runtime = {
    EventsOn: () => () => undefined,
  }
})

afterEach(() => {
  raeumeApi()
  delete (globalThis as { runtime?: unknown }).runtime
})

describe('Das Band und die Verträglichkeit', () => {
  const urteil = kataloge.de.allgemein.vertraeglichkeit

  it('lässt den Ladeknopf weg, wenn die Fassung den Bestand nicht öffnen kann', async () => {
    belegeApi({
      ...belegeFenster('de'),
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      Vertraeglichkeit: () =>
        Promise.resolve({
          fassung: '9.9.9',
          urteil: 'unvertraeglich',
          bereit: false,
          hindernisse: [{ art: 'wanderungFehlt', stelle: '0052_decks.sql' }],
          fehler: null,
        }),
    })

    render(<App />)

    expect(await screen.findByText(urteil.unvertraeglich)).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: band.laden }),
    ).not.toBeInTheDocument()
  })

  // **Und er steht da, wenn sie passt** — ohne diese Hälfte wäre der Lauf
  // oben auch dann grün, wenn es den Knopf nie gäbe.
  it('lässt den Ladeknopf stehen, wenn die Fassung passt', async () => {
    belegeApi({
      ...belegeFenster('de'),
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      Vertraeglichkeit: () =>
        Promise.resolve({
          fassung: '9.9.9',
          urteil: 'vertraeglich',
          bereit: false,
          hindernisse: [],
          fehler: null,
        }),
    })

    render(<App />)

    expect(await screen.findByText(urteil.vertraeglich)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: band.laden })).toBeInTheDocument()
  })

  // **Und der Wechsel lässt sich von hier vorbereiten** — der Weg durch die
  // Anwendung, nicht nur das Bauteil.
  it('bereitet den Wechsel vor und sagt es', async () => {
    const WechselVorbereiten = vi.fn(() =>
      Promise.resolve({
        fassung: '9.9.9',
        urteil: 'vertraeglich',
        bereit: true,
        hindernisse: [],
        fehler: null,
      }),
    )

    belegeApi({
      ...belegeFenster('de'),
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      Vertraeglichkeit: () =>
        Promise.resolve({
          fassung: '9.9.9',
          urteil: 'vertraeglich',
          bereit: false,
          hindernisse: [],
          fehler: null,
        }),
      WechselVorbereiten,
    })

    render(<App />)

    await userEvent.click(
      await screen.findByRole('button', { name: urteil.vorbereiten }),
    )

    expect(await screen.findByText(urteil.bereit)).toBeInTheDocument()
    expect(WechselVorbereiten).toHaveBeenCalledTimes(1)
  })

  // **Und ein Fehler beim Vorbereiten kommt im Band an** (§8) — sonst bliebe
  // eine abgewiesene Zusage ohne Empfänger stehen.
  it('meldet einen Fehler beim Vorbereiten', async () => {
    belegeApi({
      ...belegeFenster('de'),
      NachNeuerFassungSehen: () => Promise.resolve(neuere),
      Vertraeglichkeit: () =>
        Promise.resolve({
          fassung: '9.9.9',
          urteil: 'vertraeglich',
          bereit: false,
          hindernisse: [],
          fehler: null,
        }),
      WechselVorbereiten: () =>
        Promise.resolve({
          fassung: '9.9.9',
          urteil: 'vertraeglich',
          bereit: false,
          hindernisse: [],
          fehler: {
            code: 'validierung',
            schluessel: 'fehler.validierung',
            meldung: 'Die Sicherung ließ sich nicht schreiben.',
            werte: {},
          },
        }),
    })

    render(<App />)

    await userEvent.click(
      await screen.findByRole('button', {
        name: kataloge.de.allgemein.vertraeglichkeit.vorbereiten,
      }),
    )

    expect(
      await screen.findByText('Die Sicherung ließ sich nicht schreiben.'),
    ).toBeInTheDocument()
  })
})
