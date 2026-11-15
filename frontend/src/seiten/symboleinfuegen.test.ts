// **Das Einsetzen eines Symbols** —
// #564.
//
//	String code = style().symbol_font.font->insertSymbolCode(id);
//	if (!style().always_symbol) code = "<sym>" + code + "</sym>";
//	replaceSelection(code, _ACTION_("insert symbol"));
//
// `TextValueEditor::onCommand`, `src/gui/value/text.cpp`, Zeilen 589–593,
// Fassung f9d9356 vom 2021-07-06.
import { describe, expect, it } from 'vitest'

import { ausgezeichnet, symbolEinsetzen } from './symboleinfuegen'

describe('ausgezeichnet', () => {
  it.each([
    ['T', true, 'T', '`always symbol:` fügt roh ein'],
    ['T', false, '<sym>T</sym>', 'sonst wird gefasst'],
  ])('%s bei roh=%s ergibt %s', (code, roh, erwartet) => {
    expect(ausgezeichnet(code, roh)).toBe(erwartet)
  })
})

describe('symbolEinsetzen', () => {
  it('schreibt in das Feld und gibt den neuen Wert', () => {
    const eingabe = document.createElement('input')
    document.body.append(eingabe)
    eingabe.value = '2'
    eingabe.setSelectionRange(1, 1)

    expect(symbolEinsetzen(eingabe, '2', 'R', true)).toBe('2R')
    expect(eingabe.value).toBe('2R')

    // **Der Fokus geht zurück** — wer ein Symbol einsetzt, tippt weiter.
    expect(document.activeElement).toBe(eingabe)

    eingabe.remove()
  })

  it('zeichnet aus, wo das Feld Fließtext trägt', () => {
    const kasten = document.createElement('textarea')
    kasten.value = ''

    expect(symbolEinsetzen(kasten, '', 'T', false)).toBe('<sym>T</sym>')
  })

  // **Der rohe Wert bleibt roh** —
  // #687.
  //
  // Das Feld zeigt seit Stufe 2 den Klartext; das Symbol geht trotzdem in den
  // **ausgezeichneten** Wert, und zwar an die Stelle, an der die Schreibmarke
  // im Klartext steht. Wer stumpf in den Kasten schriebe, machte aus einer
  // Auszeichnung sichtbaren Text — und schriebe beim nächsten Speichern einen
  // Wert ohne seine Marken.
  it('setzt in den ausgezeichneten Wert ein, nicht in den Klartext', () => {
    const kasten = document.createElement('textarea')
    document.body.append(kasten)
    kasten.value = 'Es brennt'
    kasten.setSelectionRange(2, 2)

    expect(
      symbolEinsetzen(kasten, '<i-flavor>Es brennt</i-flavor>', 'T', false),
    ).toBe('<i-flavor>Es<sym>T</sym> brennt</i-flavor>')

    // **Und der Kasten zeigt weiter Klartext.**
    expect(kasten.value).toBe('EsT brennt')

    kasten.remove()
  })

  // **Ohne Auswahl steht die Stelle am Ende.** Ein Feld, in dem noch niemand
  // war, meldet `null`; dort ist „hinten" die einzige Stelle, die niemanden
  // überrascht. `input type="number"` ist das Steuerelement, das es tut — die
  // Textfelder dieser Anwendung melden immer eine Zahl.
  it('hängt an, wo es keine Auswahl gibt', () => {
    const eingabe = document.createElement('input')
    eingabe.type = 'number'
    eingabe.value = '2'

    expect(symbolEinsetzen(eingabe, '2', 'R', true)).toBe('2R')
  })

  // **Ein Steuerelement ohne Text bekommt nichts untergeschoben.** MSE hängt
  // das Einfügemenü an den `TextValueEditor` und an keinen anderen; eine
  // Auswahl nähme einen Wert an, den ihre eigene Liste nicht kennt.
  it.each([
    ['eine Auswahl', () => document.createElement('select')],
    ['nichts', () => null],
  ])('lässt %s unangetastet', (_name, machen) => {
    expect(symbolEinsetzen(machen(), '', 'T', true)).toBeUndefined()
  })
})
