// **Ein Feld, das niemand bearbeiten darf, wird nicht angeboten** —
// #708.
//
//	bool is_enabled(ValueViewerP const& v) {
//	  return v->getField()->editable && v->isVisible();
//	}
//
// `src/gui/control/card_editor.cpp`, Zeilen 155–156, Fassung f9d9356 vom
// 2021-07-06. Der Kommentar in `angebotenefelder.ts` führte diese Regel schon
// an — umgesetzt war davon nur die **Sichtbarkeit**.
//
// 24 der 249 Kartenfelder von `magic.mse-game` tragen `editable: false`, und
// jedes davon stand als Eingabefeld da, das der Schreibweg danach abwies.
import { describe, expect, it } from 'vitest'

import type { FeldbeschreibungDTO, WertDTO } from '../api'
import type { FeldflaecheDTO } from '../formen'
import { angeboteneFelder } from './angebotenefelder'

function flaeche(feld: string, oben: number): FeldflaecheDTO {
  return { feld, links: 0, oben, breite: 10, hoehe: 5 }
}

function form(
  name: string,
  teile: Partial<FeldbeschreibungDTO> = {},
): FeldbeschreibungDTO {
  return {
    name,
    art: 'text',
    mehrzeilig: false,
    wahlen: [],
    beschreibung: '',
    vorgabe: '',
    aufschrift: name,
    teilfelder: [],
    reihe: '',
    stelle: 0,
    editierbar: true,
    gespeichert: true,
    ...teile,
  }
}

describe('Ein nicht editierbares Feld', () => {
  it('steht nicht im Formular', () => {
    const felder = angeboteneFelder(
      [],
      [form('name'), form('cmc', { editierbar: false })],
      [flaeche('name', 30), flaeche('cmc', 60)],
    )

    expect(felder.map((f) => f.feld)).toEqual(['name'])
  })

  // **Auch dann nicht, wenn die Karte einen Wert dafür führt.** Ein Set aus
  // MTG JSON trägt `cmc`; der Wert bleibt in der Datei, aber ein Eingabefeld
  // wäre eine Zusage, die der Schreibweg nicht hält.
  it('steht auch dann nicht da, wenn die Karte ihn führt', () => {
    const werte: WertDTO[] = [{ feld: 'cmc', wert: '3' }]

    const felder = angeboteneFelder(
      werte,
      [form('cmc', { editierbar: false })],
      [],
    )

    expect(felder.map((f) => f.feld)).toEqual([])
  })

  // **Seine Teile gelten weiter.** `text` ist bei `magic.mse-game`
  // editierbar; wäre es das nicht, blieben Regeltext und Flavortext trotzdem
  // die Felder, die die Eingabe annehmen.
  it('lässt seine Teilfelder gelten', () => {
    const felder = angeboteneFelder(
      [],
      [
        form('text', { editierbar: false, teilfelder: ['rule text'] }),
        form('rule text'),
      ],
      [flaeche('text', 300)],
    )

    expect(felder.map((f) => f.feld)).toEqual(['rule text'])
  })

  // **Und ein Feld ohne Erklärung bleibt, was es war.** Ein Set aus MTG JSON
  // trägt Felder, die das Spiel nicht kennt; sie zu verstecken hieße, sie
  // stillschweigend zu verlieren.
  it('lässt ein Feld ohne Erklärung stehen', () => {
    const felder = angeboteneFelder([{ feld: 'eigenes', wert: 'x' }], [], [])

    expect(felder.map((f) => f.feld)).toEqual(['eigenes'])
  })
})
