// **Der lokale Suchindex** — Werkstück A15 von Phase 5.3
// (#1514).
//
// # Wogegen hier geprüft wird
//
// **Dass der Modus dasteht.** Eine lexikalische Antwort ist etwas anderes als
// eine hybride, und an den Treffern sieht man es nicht — wer es verschwiege,
// verkaufte den Rückfall als die Sache selbst.
//
// **Und dass die Herkunft am Treffer steht.** Ohne Quelle, Stand und Abrufzeit
// ist ein Treffer eine Behauptung.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Suchindex } from './Suchindex'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.seiten.technik

// einManifest ist ein vollständiger Indexstand.
function einManifest() {
  return {
    modell: 'all-minilm',
    modellrevision: 'aaaa',
    laufzeit: 'ollama 0.33.3',
    abstandsmass: 'cosinus',
    dimension: 384,
    zerteilfassung: 'karten-seiten-1',
    suchkonfiguration: 'rrf-60',
    korpusstaende: [
      { quelle: 'oracle', revision: '2026-09-01', datum: '2026-09-01' },
    ],
  }
}

// einStand baut die Auskunft.
function einStand(anders: Record<string, unknown> = {}) {
  return {
    aktiv: true,
    manifest: einManifest(),
    aufbau: { laeuft: false, generation: 1, geschrieben: 12, fehler: '' },
    suchbar: true,
    grund: '',
    fehler: null,
    ...anders,
  }
}

// einTreffer ist ein Fund samt Herkunft.
function einTreffer() {
  return {
    rang: 1,
    kennung: 'scryfall-oracle:o1:0',
    quelle: 'oracle',
    bezugsadresse: 'https://beispiel.invalid/o1',
    titel: 'Funkenschlag',
    text: 'Funkenschlag fügt 3 Schadenspunkte zu.',
    ausschnitt: '…3 Schadenspunkte…',
    sprache: 'de',
    regelnummer: '',
    urheber: '',
    quellrevision: '2026-09-01',
    quelldatum: '2026-09-01',
    abgerufen: '2026-09-11T08:00:00Z',
  }
}

// zeige rendert den Bereich im deutschen Sprachraum.
function zeige() {
  render(
    <Sprachraum sprache="de">
      <Suchindex />
    </Sprachraum>,
  )
}

describe('Der lokale Suchindex', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('nennt Modell, Verfahren und die Stände je Quelle', async () => {
    belegeApi({ Indexstand: () => Promise.resolve(einStand()) })

    zeige()

    expect(await screen.findByText('all-minilm')).toBeInTheDocument()
    expect(screen.getByText(/rrf-60/)).toBeInTheDocument()
    expect(screen.getByText(/oracle · 2026-09-01/)).toBeInTheDocument()
    expect(screen.getByText(/^ja/)).toBeInTheDocument()
  })

  it('sagt, wenn gar kein Index dasteht', async () => {
    belegeApi({
      Indexstand: () =>
        Promise.resolve(einStand({ aktiv: false, suchbar: false })),
    })

    zeige()

    expect(await screen.findByText(texte.suchindexKeiner)).toBeInTheDocument()
  })

  // **Der Satz, für den es diesen Bereich gibt.**
  it('sagt, wie gesucht wurde — und nennt den Grund für den Rückfall', async () => {
    belegeApi({
      Indexstand: () => Promise.resolve(einStand()),
      Indexsuche: () =>
        Promise.resolve({
          treffer: [einTreffer()],
          modus: 'lexikalisch',
          grund: 'Das Modell ist nicht mehr installiert.',
          fehler: null,
        }),
    })

    zeige()

    await screen.findByText('all-minilm')

    await userEvent.type(
      screen.getByLabelText(texte.suchindexFrage),
      'funkenschlag{Enter}',
    )

    expect(
      await screen.findByText(new RegExp(texte.suchindexModusLexikalisch, 'u')),
    ).toBeInTheDocument()

    expect(
      screen.getByText(/Das Modell ist nicht mehr installiert/),
    ).toBeInTheDocument()

    // **Und die Herkunft steht am Treffer** — mit Abrufzeit, die im
    // Korpusstand darüber nicht vorkommt.
    expect(screen.getByText('Funkenschlag')).toBeInTheDocument()
    expect(
      screen.getByText(/oracle · 2026-09-01 2026-09-01 · 2026-09-11/),
    ).toBeInTheDocument()
  })

  it('sagt es, wenn nichts gefunden wurde', async () => {
    belegeApi({
      Indexstand: () => Promise.resolve(einStand()),
      Indexsuche: () =>
        Promise.resolve({
          treffer: [],
          modus: 'hybrid',
          grund: '',
          fehler: null,
        }),
    })

    zeige()

    await screen.findByText('all-minilm')

    await userEvent.click(
      screen.getByRole('button', { name: texte.suchindexSuchen }),
    )

    expect(await screen.findByText(texte.suchindexLeer)).toBeInTheDocument()
  })

  // **Ein unbekannter Modus wird gezeigt und nicht verschluckt** — er kommt
  // aus Go, und ein leeres Feld wäre die Auskunft „hybrid".
  it('zeigt einen unbekannten Modus unverändert', async () => {
    belegeApi({
      Indexstand: () => Promise.resolve(einStand()),
      Indexsuche: () =>
        Promise.resolve({
          treffer: [],
          modus: 'etwas-neues',
          grund: '',
          fehler: null,
        }),
    })

    zeige()

    await screen.findByText('all-minilm')

    await userEvent.click(
      screen.getByRole('button', { name: texte.suchindexSuchen }),
    )

    expect(await screen.findByText(/etwas-neues/)).toBeInTheDocument()
  })

  it('zeigt den vollen Text, wo es keinen Ausschnitt gibt', async () => {
    belegeApi({
      Indexstand: () => Promise.resolve(einStand()),
      Indexsuche: () =>
        Promise.resolve({
          treffer: [{ ...einTreffer(), ausschnitt: '' }],
          modus: 'hybrid',
          grund: '',
          fehler: null,
        }),
    })

    zeige()

    await screen.findByText('all-minilm')

    await userEvent.click(
      screen.getByRole('button', { name: texte.suchindexSuchen }),
    )

    expect(
      await screen.findByText('Funkenschlag fügt 3 Schadenspunkte zu.'),
    ).toBeInTheDocument()
  })

  it('meldet einen Fehler der Suche am Bereich', async () => {
    belegeApi({
      Indexstand: () => Promise.resolve(einStand()),
      Indexsuche: () => Promise.reject(new Error('der Bestand ist zu')),
    })

    zeige()

    await screen.findByText('all-minilm')

    await userEvent.click(
      screen.getByRole('button', { name: texte.suchindexSuchen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'der Bestand ist zu',
    )
  })

  it('meldet einen Fehler der Auskunft', async () => {
    belegeApi({
      Indexstand: () => Promise.reject(new Error('niemand da')),
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent('niemand da')
  })

  // **Ein laufender Aufbau lässt sich abbrechen — und danach steht der
  // frische Stand da.**
  it('bricht einen laufenden Aufbau ab und holt den Stand neu', async () => {
    const abbruch = vi.fn(() => Promise.resolve(true))

    let gefragt = 0

    belegeApi({
      Indexstand: () => {
        gefragt += 1

        return Promise.resolve(
          einStand({
            aufbau: {
              laeuft: gefragt === 1,
              generation: 2,
              geschrieben: 500,
              fehler: '',
            },
          }),
        )
      },
      IndexAbbrechen: abbruch,
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.suchindexAbbrechen }),
    )

    expect(abbruch).toHaveBeenCalled()

    await waitFor(() => {
      expect(
        screen.queryByRole('button', { name: texte.suchindexAbbrechen }),
      ).not.toBeInTheDocument()
    })
  })
})
