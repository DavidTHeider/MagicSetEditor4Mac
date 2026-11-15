// **Was dieser Prozess gewirkt hat** — Werkstück O9 von Phase 5.2
// (#1433).
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Wirkungszaehler } from './Wirkungszaehler'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { volleMetriken } from '../test/metriken'

import type { MetrikenAntwort } from '../apimetriken'

const texte = kataloge.de.seiten.technik

afterEach(raeumeApi)

// zeige zeichnet ihn über einer Antwort.
function zeige(teil: Partial<MetrikenAntwort> = {}) {
  belegeApi({
    Metriken: vi.fn(() =>
      Promise.resolve({
        ...volleMetriken(),
        reihen: [],
        grenze: 40,
        ...teil,
      }),
    ),
  })

  render(
    <Sprachraum sprache="de">
      <Wirkungszaehler />
    </Sprachraum>,
  )
}

describe('Der Wirkungszähler', () => {
  it('zeigt jede Reihe mit ihrer Zahl', async () => {
    zeige({
      reihen: [
        {
          flaeche: 'fenster',
          wirkung: 'lokalSchreiben',
          ausgang: 'gewaehrt',
          anzahl: 3,
        },
        {
          flaeche: 'dienst',
          wirkung: 'externSenden',
          ausgang: 'abgewiesen',
          anzahl: 1,
        },
      ],
    })

    // **Übersetzt und nicht roh** — dieselbe Zusage wie beim
    // Betriebszustand eine Überschrift weiter oben.
    expect(
      await screen.findByText(texte.wirkungen.lokalSchreiben),
    ).toBeInTheDocument()
    expect(screen.getByText(texte.flaechen.fenster)).toBeInTheDocument()
    expect(screen.getByText(texte.ausgaenge.gewaehrt)).toBeInTheDocument()
    expect(screen.getByText(texte.wirkungen.externSenden)).toBeInTheDocument()
    expect(screen.getByText(texte.ausgaenge.abgewiesen)).toBeInTheDocument()
    expect(screen.getByText('3')).toBeInTheDocument()
    expect(screen.getByText('1')).toBeInTheDocument()
  })

  // **Die Grenze ist die Zusage** (#1433): wer die Tabelle sieht, soll ohne
  // Nachschlagen wissen, dass sie nicht mit dem Bestand wächst.
  // **Ein Bezeichner, den dieses Fenster nicht kennt, steht im Wortlaut da** —
  // er ist immer noch eine Auskunft, eine leere Zelle wäre keine.
  it('zeigt einen unbekannten Bezeichner im Wortlaut', async () => {
    zeige({
      reihen: [
        {
          flaeche: 'satellit',
          wirkung: 'irgendwas',
          ausgang: 'unklar',
          anzahl: 2,
        },
      ],
    })

    expect(await screen.findByText('satellit')).toBeInTheDocument()
    expect(screen.getByText('irgendwas')).toBeInTheDocument()
    expect(screen.getByText('unklar')).toBeInTheDocument()
  })

  it('nennt die Zahl der überhaupt möglichen Reihen', async () => {
    zeige({
      reihen: [
        {
          flaeche: 'mcp',
          wirkung: 'lokalLesen',
          ausgang: 'gewaehrt',
          anzahl: 7,
        },
      ],
    })

    expect(
      await screen.findByText(texte.metrikenGrenze.replace('{grenze}', '40')),
    ).toBeInTheDocument()
  })

  // **Leer ist eine Auskunft und keine Lücke**: wer nur liest, wirkt nichts.
  it('sagt bei keiner Reihe, dass nichts nach außen gewirkt wurde', async () => {
    zeige({ laufzeiten: [] })

    expect(await screen.findByText(texte.metrikenLeer)).toBeInTheDocument()

    // **Keine Tafel** — auch nicht die der Laufzeiten darunter, denn es
    // gibt keine.
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('zeigt einen Fehler als Meldung und keine Tabelle', async () => {
    zeige({
      fehler: {
        code: 'unbekannt',
        schluessel: 'fehler.unbekannt',
        meldung: 'Es ist ein unerwarteter Fehler aufgetreten.',
        werte: {},
      },
    })

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })
})
