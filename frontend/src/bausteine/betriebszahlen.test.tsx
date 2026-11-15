// **Laufzeiten und Verbrauch** — OPS-05.AK-01/04
// (#1722):
// jede Zahl mit ihrer Einheit, der Erhebungszeitraum, der Aufwand und was
// nicht gemessen ist.
import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Betriebszahlen } from './Betriebszahlen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { volleMetriken } from '../test/metriken'

import type { MetrikenAntwort } from '../apimetriken'

const texte = kataloge.de.seiten.technik.betrieb

function zeige(teil: Partial<MetrikenAntwort> = {}) {
  render(
    <Sprachraum sprache="de">
      <Betriebszahlen zahlen={{ ...volleMetriken(), ...teil }} />
    </Sprachraum>,
  )
}

describe('Die Betriebszahlen', () => {
  it('nennen Zeitraum, Aufwand, Speicher, Bestand und Zwischenspeicher', () => {
    zeige()

    // **Der Aufwand der Erhebung als eigener Satz** — AP-11 (#2368), B-27,
    // und sichtbar: OPS-05.AK-04 verlangt ihn, und ein Titel ist mit der
    // Tastatur nicht zu erreichen.
    const zeitraum = screen.getByText(/^Erhoben von /)
    expect(zeitraum).not.toHaveTextContent('µs')
    expect(screen.getByText('Aufwand der Erhebung: 250 µs')).toBeInTheDocument()
    expect(
      screen.getByText(/2 offen, 1 in Benutzung, 3-mal gewartet/),
    ).toBeInTheDocument()
    expect(screen.getByText(/7 Treffer, 2 geholt/)).toBeInTheDocument()
    expect(screen.getByText('12')).toBeInTheDocument()
    expect(screen.getByText(/1,5 s/)).toBeInTheDocument()
  })

  it('sagen, was nicht gemessen ist', () => {
    zeige()

    expect(
      screen.getByText(
        new RegExp(texte.ungemessenNamen.fenster.replace(/[()]/g, '.')),
      ),
    ).toBeInTheDocument()
  })

  it('zeigen die Laufzeiten mit Fehlerrate und Mittelwert', () => {
    zeige()

    const tafel = screen.getByRole('table', { name: texte.laufzeiten })
    const zeile = within(tafel).getAllByRole('row')[1]

    expect(zeile).toHaveTextContent('Dienst')
    expect(zeile).toHaveTextContent(texte.arten.lesend)
    expect(zeile).toHaveTextContent('4')
    expect(zeile).toHaveTextContent('25 %')
    expect(zeile).toHaveTextContent('2,5 ms')
  })

  it('sagen ohne Rufe, warum die Tafel leer ist', () => {
    zeige({ laufzeiten: [] })

    expect(screen.getByText(texte.laufzeitenLeer)).toBeInTheDocument()
    // **Ohne Endpunktnamen** (AP-11, B-27) — die stehen in der Hilfe.
    expect(texte.laufzeitenLeer).not.toMatch(/GET \/metriken/)
    expect(screen.queryByRole('table')).toBeNull()
  })

  it('sagen ohne Bestand, dass es keinen gibt', () => {
    zeige({ datenbank: null })

    expect(screen.getByText(texte.ohneBestand)).toBeInTheDocument()
  })

  it('nennen die Spur mit ihren Grenzen, wenn sie an ist', () => {
    zeige({
      spur: {
        ...volleMetriken().spur,
        an: true,
        abtastung: 10,
        grenze: 500,
        aufbewahrungS: 3600,
        gesehen: 30,
        aufgenommen: 3,
        verdraengt: 0,
        abgelaufen: 0,
        punkte: [],
      },
    })

    expect(screen.getByText(/3 von 30 Rufen, 10 %/)).toBeInTheDocument()
  })

  it('sagen, wo die Spur einzuschalten ist, wenn sie aus ist', () => {
    zeige()

    expect(screen.getByText(texte.spurAus)).toBeInTheDocument()
  })

  // **Was dieses Fenster nicht kennt, steht im Wortlaut da** — dieselbe
  // Regel wie beim Wirkungszähler darüber.
  it('zeigen Unbekanntes im Wortlaut', () => {
    zeige({
      laufzeiten: [
        {
          flaeche: 'satellit',
          art: 'fremd',
          anzahl: 1,
          abgewiesen: 0,
          gescheitert: 0,
          summeMs: 3000,
          // **Ein kürzerer Eimersatz** als die Grenzen — was fehlt, ist
          // null und steht nicht da.
          eimer: [0, 0, 0, 0, 0, 1],
        },
      ],
      ungemessen: ['raumschiff'],
    })

    expect(screen.getByText('satellit')).toBeInTheDocument()
    expect(screen.getByText('fremd')).toBeInTheDocument()
    expect(screen.getByText(/raumschiff/)).toBeInTheDocument()
    expect(screen.getByText(/3 s/)).toBeInTheDocument()
  })
})
