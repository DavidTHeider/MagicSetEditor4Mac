// **Betrieb und Vertrauen** — OPS-07
// (#1724):
// Fassung, Datenstand, Signatur ohne Beschönigung, Aktualisierung, Orte und
// ein Prüfstand, der Geprüftes von Vorhandenem trennt.
import { render, screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Hilfebetrieb } from './Hilfebetrieb'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { BetriebsauskunftAntwort } from '../apibetrieb'

const texte = kataloge.de.hilfe.betrieb

afterEach(raeumeApi)

function auskunft(
  teil: Partial<BetriebsauskunftAntwort> = {},
): BetriebsauskunftAntwort {
  return {
    fassung: '0.6.0',
    stand: '0123456789abcdef',
    gebaut: '2026-09-26T12:00:00Z',
    veraendert: false,
    migrationen: 91,
    letzteMigration: '0091_etwas.sql',
    signatur: 'adhoc',
    notarisiert: false,
    aktualisierung: 'aus',
    orte: [
      {
        name: 'bestand',
        pfad: '~/Library/Application Support/MagicSetEditor4Mac/bestand.db',
      },
      { name: 'woanders', pfad: '~/woanders' },
    ],
    pruefstand: [
      {
        bereich: 'integration',
        gegenstand: 'Scryfall',
        stand: 'geprueft',
        beleg: 'docs/HAERTUNG.md',
      },
      {
        bereich: 'barrierefreiheit',
        gegenstand: 'VoiceOver',
        stand: 'blockiert',
        // **Zusammengesetzt**, damit ein Werkzeug, das Issue-Verweise in Klartext
        // wandelt, die Vorrichtung nicht mit umschreibt — so geschehen bei 1.0.0.
        beleg:
          'https://github.com/DavidTHeider/MagicSetEditor4Mac' + '/issues/2255',
      },
      {
        bereich: 'plattform',
        gegenstand: 'macOS 15.3 bis 26',
        stand: 'vorhanden',
        beleg: 'build/darwin/Info.plist',
      },
      {
        bereich: 'mond',
        gegenstand: 'Etwas',
        stand: 'unklar',
        beleg: 'docs/X.md',
      },
    ],
    fehler: null,
    ...teil,
  }
}

function zeige(teil: Partial<BetriebsauskunftAntwort> = {}) {
  belegeApi({ Betriebsauskunft: vi.fn(() => Promise.resolve(auskunft(teil))) })

  render(
    <Sprachraum sprache="de">
      <Hilfebetrieb />
    </Sprachraum>,
  )
}

describe('Betrieb und Vertrauen', () => {
  it('nennen Fassung, Daten, Signatur und Aktualisierung', async () => {
    zeige()

    expect(
      await screen.findByText(/0\.6\.0, gebaut aus 0123456/),
    ).toBeInTheDocument()
    expect(
      screen.getByText(/91 Migrationen verbucht, zuletzt 0091_etwas\.sql/),
    ).toBeInTheDocument()
    expect(screen.getByText(/nicht notarisiert/)).toBeInTheDocument()
    expect(screen.getByText(texte.aktualisierung.aus)).toBeInTheDocument()
    expect(screen.queryByText(texte.fassungVeraendert)).toBeNull()
  })

  it('nennen die Orte als Tilde, Unbekanntes im Wortlaut', async () => {
    zeige()

    expect(await screen.findByText(texte.orte.bestand)).toBeInTheDocument()
    expect(
      screen.getByText(/~\/Library\/Application Support/),
    ).toBeInTheDocument()
    expect(screen.getByText('woanders')).toBeInTheDocument()
  })

  it('trennen Geprüftes von Vorhandenem und nennen den Beleg', async () => {
    zeige()

    const tafel = await screen.findByRole('table', {
      name: texte.pruefstandName,
    })
    const zeilen = within(tafel).getAllByRole('row')

    expect(zeilen[1]).toHaveTextContent('Scryfall')
    expect(zeilen[1]).toHaveTextContent(texte.staende.geprueft)
    expect(zeilen[1]).toHaveTextContent('docs/HAERTUNG.md')
    expect(zeilen[2]).toHaveTextContent(texte.staende.blockiert)
    expect(zeilen[2]).toHaveTextContent('#2255')
    expect(zeilen[3]).toHaveTextContent(texte.gegenstaende['macOS 15.3 bis 26'])
    expect(zeilen[4]).toHaveTextContent('mond')
    expect(zeilen[4]).toHaveTextContent('unklar')
  })

  it('sagen, wenn der Bau verändert war und kein Bestand offen ist', async () => {
    zeige({ veraendert: true, migrationen: 0, aktualisierung: 'seltsam' })

    expect(await screen.findByText(texte.fassungVeraendert)).toBeInTheDocument()
    expect(screen.getByText(texte.datenOhne)).toBeInTheDocument()
    expect(screen.getByText('seltsam')).toBeInTheDocument()
  })

  it('zeigen einen Fehler', async () => {
    zeige({
      fehler: { code: 'unbekannt', schluessel: '', meldung: 'Das ging nicht.' },
    })

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})
