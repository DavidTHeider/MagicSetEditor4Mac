// **Ein berechnetes Feld wird durch seine Teile ersetzt**
// (#696).
//
// `text` trägt `save value: false` und entsteht aus Regeltext und Flavortext.
// Der Editor bot einen Textkasten „Text" an; wer hineintippte, sah die Karte
// unverändert — die Vorlage liest `text` nie.
import { describe, expect, it } from 'vitest'

import type { FeldbeschreibungDTO, WertDTO } from '../api'
import type { FeldflaecheDTO } from '../formen'
import { angeboteneFelder } from './angebotenefelder'

function flaeche(feld: string, oben: number): FeldflaecheDTO {
  return { feld, links: 0, oben, breite: 10, hoehe: 5 }
}

function form(name: string, teilfelder: string[] = []): FeldbeschreibungDTO {
  return {
    name,
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    aufschrift: name,
    teilfelder,
    editierbar: true,
    gespeichert: true,
    reihe: '',
    stelle: 0,
  }
}

function wert(feld: string, text: string): WertDTO {
  return { feld, wert: text }
}

describe('Ein Feld, das aus anderen entsteht', () => {
  it('bietet seine Teile an seiner Stelle an', () => {
    const felder = angeboteneFelder(
      [
        wert('rule text', 'Füge drei Schaden zu.'),
        wert('flavor text', 'Es brennt.'),
      ],
      [
        form('name'),
        form('text', ['rule text', 'flavor text']),
        form('rule text'),
        form('flavor text'),
      ],
      [flaeche('name', 30), flaeche('text', 300)],
    )

    expect(felder.map((f) => f.feld)).toEqual([
      'name',
      'rule text',
      'flavor text',
    ])
    expect(felder.map((f) => f.wert)).toEqual([
      '',
      'Füge drei Schaden zu.',
      'Es brennt.',
    ])
  })

  // **Ein Teil, den das Spiel nicht erklärt, bleibt weg.** Er hätte kein
  // Steuerelement, keine Aufschrift und keinen Anfangswert — angeboten wäre
  // er ein leeres Kästchen ohne Namen.
  it('lässt einen Teil ohne Erklärung aus', () => {
    const felder = angeboteneFelder(
      [],
      [form('text', ['rule text', 'flavor text']), form('rule text')],
      [flaeche('text', 300)],
    )

    expect(felder.map((f) => f.feld)).toEqual(['rule text'])
  })

  // **Und ein Teil, der ohnehin gezeichnet wird, steht nur einmal da.**
  it('bietet einen Teil nicht zweimal an', () => {
    const felder = angeboteneFelder(
      [],
      [form('text', ['rule text']), form('rule text')],
      [flaeche('rule text', 100), flaeche('text', 300)],
    )

    expect(felder.map((f) => f.feld)).toEqual(['rule text'])
  })
})
