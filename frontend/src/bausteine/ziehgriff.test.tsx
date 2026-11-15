// **Das Kartenbild in den Finder ziehen** — AP-12
// (#2369),
// Befund B-24. Gemessen am 2026-10-17 (`ap12ziehquelle09`): WebKit legt ein
// gezogenes `<img>` im Finder als PNG ab und nennt die Datei nach dem letzten
// Teil der Adresse. Der Griff zeigt deshalb auf `…/<marke>/<Kartenname>.png`
// (der Bildweg sucht weiter nur die Marke, `internal/vorschauweg`).
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Ziehgriff, ziehadresse } from './Ziehgriff'

const worte = kataloge.de.seiten.karten.ziehen

describe('ziehadresse', () => {
  it('hängt den Namen hinter die Marke', () => {
    expect(ziehadresse('/vorschau/474dccaa9edc0be9.png', 'Langtext')).toBe(
      '/vorschau/474dccaa9edc0be9/Langtext.png',
    )
  })

  it('kodiert den Namen und ersetzt, was ein Dateiname nicht trägt', () => {
    expect(ziehadresse('/vorschau/ab.png', 'Ja/Nein: „Ä“?')).toBe(
      `/vorschau/ab/${encodeURIComponent('Ja-Nein- „Ä“?')}.png`,
    )
  })

  it('macht aus einem Steuerzeichen einen Strich', () => {
    expect(ziehadresse('/vorschau/ab.png', 'Zeile\tzwei')).toBe(
      '/vorschau/ab/Zeile-zwei.png',
    )
  })

  it('nimmt „Karte“, wenn der Name leer ist', () => {
    expect(ziehadresse('/vorschau/ab.png', '  ')).toBe('/vorschau/ab/Karte.png')
  })

  it('lässt eine fremde Adresse, wie sie ist', () => {
    expect(ziehadresse('data:image/png;base64,AA==', 'Langtext')).toBe(
      'data:image/png;base64,AA==',
    )
  })
})

describe('der Ziehgriff', () => {
  it('zeigt ein ziehbares Bild der Karte mit seiner Beschriftung', () => {
    render(
      <Sprachraum sprache="de">
        <Ziehgriff bild="/vorschau/ab.png" kartenname="Langtext" />
      </Sprachraum>,
    )

    const griff = screen.getByText(worte.beschriftung)
    expect(griff.closest('[title]')).toHaveAttribute(
      'title',
      'Ziehe das Bild in den Finder, um die Karte als PNG abzulegen. Dasselbe geht unter Ausgeben › Bilder.',
    )
    const bild = griff.parentElement?.querySelector('img')
    expect(bild).toHaveAttribute('src', '/vorschau/ab/Langtext.png')
    expect(bild).toHaveAttribute('draggable', 'true')
    // **Das Bild ist Schmuck** — der Text daneben sagt, was es tut.
    expect(bild).toHaveAttribute('alt', '')
  })
})
