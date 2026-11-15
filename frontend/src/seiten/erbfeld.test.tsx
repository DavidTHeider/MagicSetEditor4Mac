// **Ein freies Feld, das erbt, zeigt nicht das Wort „erben"** —
// #1993.
//
// Gemessen am 2026-10-01 in der Abnahme-VM (englisches System, #1871):
// VoiceOver las „erben, contents selected, Width of the left column, edit
// text". Die beiden Spaltenbreiten sind die einzigen freien Einstellungen mit
// Erbwert (`pruefeBreiteOderErben`, `internal/domain/arbeitsplatz.go`); eine
// Auswahl übersetzt `erben` über ihren Katalog, ein Textfeld zeigte es roh.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { Wahl, type Wahlbarer } from './Wahl'

import type { EinstellungDTO } from '../api'

function breite(wert: string): EinstellungDTO {
  return {
    schluessel: 'arbeitsplatz.navigatorbreite',
    wert,
    wirksam: wert === 'erben' ? '240' : wert,
    werte: [],
    bereich: 'arbeitsplatz',
    ebene: 'profil',
    herkunft: wert === 'erben' ? 'nutzer' : 'profil',
    vorgabe: 'erben',
  }
}

function zeige(eintrag: Wahlbarer, sprache: 'de' | 'en', aendern = vi.fn()) {
  render(
    <Sprachraum sprache={sprache}>
      <Wahl eintrag={eintrag} angehalten={false} aendern={aendern} />
    </Sprachraum>,
  )

  return aendern
}

describe('ein freies Feld, das erbt', () => {
  it.each(['de', 'en'] as const)(
    'zeigt in %s den übersetzten Erbwert statt des Schlüssels',
    (sprache) => {
      zeige(breite('erben'), sprache)

      const feld = screen.getByRole('textbox')
      const werte: Record<string, string | undefined> =
        kataloge[sprache].seiten.einstellungen.werte
      const erbe = werte['arbeitsplatz.navigatorbreite.erben']

      expect(feld).toHaveValue('')
      expect(erbe).toBeDefined()
      expect(feld).toHaveAttribute(
        'placeholder',
        expect.stringContaining(erbe ?? 'fehlt'),
      )
      expect(feld).toHaveAttribute(
        'placeholder',
        expect.stringContaining('240'),
      )
    },
  )

  // **Ohne geltenden Wert nur das Wort** — `wirksam` ist am Verbund
  // wahlweise, und „Wie beim Nutzer ()" läse sich wie ein Fehler.
  it('zeigt ohne geltenden Wert nur den Erbwert', () => {
    const ohneWirksam: Wahlbarer = {
      schluessel: 'arbeitsplatz.navigatorbreite',
      wert: 'erben',
      werte: [],
      vorgabe: 'erben',
    }

    zeige(ohneWirksam, 'de')

    const werte: Record<string, string | undefined> =
      kataloge.de.seiten.einstellungen.werte

    expect(screen.getByRole('textbox')).toHaveAttribute(
      'placeholder',
      werte['arbeitsplatz.navigatorbreite.erben'],
    )
  })

  it('schreibt nichts, wenn das leere Feld leer verlassen wird', async () => {
    const nutzer = userEvent.setup()
    const aendern = zeige(breite('erben'), 'de')

    await nutzer.click(screen.getByRole('textbox'))
    await nutzer.tab()

    expect(aendern).not.toHaveBeenCalled()
  })

  // **Leeren heißt: wieder erben** — sonst gäbe es keinen Weg zurück, und
  // ein leerer Wert ist für eine Breite keiner (`pruefeBreite`).
  it('führt ein geleertes Feld zurück zum Erbwert', async () => {
    const nutzer = userEvent.setup()
    const aendern = zeige(breite('300'), 'de')

    await nutzer.clear(screen.getByRole('textbox'))
    await nutzer.tab()

    expect(aendern).toHaveBeenCalledWith('erben')
  })
})
