// **Ein Symbol an der Schreibmarke einsetzen** —
// #564.
//
//	String code = style().symbol_font.font->insertSymbolCode(id);
//	if (!style().always_symbol) {
//	  code = _("<sym>") + code + _("</sym>");
//	}
//	replaceSelection(code, _ACTION_("insert symbol"));
//
// `TextValueEditor::onCommand`, `src/gui/value/text.cpp`, Zeilen 589–593,
// Fassung f9d9356 vom 2021-07-06.
//
// **An der Marke und nicht am Ende.** MSE ersetzt die Auswahl; wer stumpf
// anhängte, machte aus `2R` mit der Marke vorne `2RR` statt `R2R` — und beim
// Regeltext schriebe er das Symbol hinter den Punkt.

import { mitAuszeichnung, mitEinsatz } from './mitauszeichnung'
import { ohneAuszeichnung } from './ohneauszeichnung'

// ausgezeichnet fasst einen Code so, wie das Feld ihn braucht.
//
// **`roh` ist eine Aussage über das Feld** (`always symbol:`), nicht über das
// Symbol: bei den Manakosten besteht der ganze Wert aus Symbolen, im Regeltext
// steht `<sym>T</sym>` mitten im Fließtext.
export function ausgezeichnet(code: string, roh: boolean): string {
  return roh ? code : `<sym>${code}</sym>`
}

// **Der Einsatz geht in den ausgezeichneten Wert**
// (#687): das
// Feld zeigt seit Stufe 2 den Klartext, und wer stumpf in den Kasten schriebe,
// machte aus `<sym>R</sym>` sichtbaren Text — und schriebe beim nächsten
// Speichern einen Wert ohne seine Marken.
//
// `wertRoh` ist der zuletzt **geschriebene** Wert; was im Kasten steht, ist der
// Entwurf. [mitAuszeichnung] bringt beide zusammen, bevor eingesetzt wird —
// dieselbe Rechnung wie beim Tippen, und deshalb dieselbe Antwort.
export function symbolEinsetzen(
  eingabe: Element | null | undefined,
  wertRoh: string,
  code: string,
  roh: boolean,
): string | undefined {
  if (
    !(eingabe instanceof HTMLInputElement) &&
    !(eingabe instanceof HTMLTextAreaElement)
  ) {
    return undefined
  }

  const klartext = eingabe.value

  // **Beide Grenzen werden eingefangen.** `selectionStart` kann größer sein
  // als die Länge, wenn der Wert zwischenzeitlich gekürzt wurde; ohne die
  // Klemme läge die Einsatzstelle hinter dem Text.
  //
  // **Und ohne Auswahl steht sie am Ende.** Ein Feld, in dem noch niemand war,
  // meldet `null` — dann ist „hinten" die einzige Stelle, die niemanden
  // überrascht.
  const von = grenze(eingabe.selectionStart ?? klartext.length, klartext.length)
  const bis = grenze(
    Math.max(eingabe.selectionEnd ?? von, von),
    klartext.length,
  )

  const neu = mitEinsatz(
    mitAuszeichnung(wertRoh, klartext),
    von,
    bis,
    ausgezeichnet(code, roh),
  )

  eingabe.value = ohneAuszeichnung(neu)

  // **Und der Fokus geht zurück.** Wer ein Symbol einsetzt, tippt weiter; ohne
  // diese Zeile stünde die Marke auf der Schaltfläche, und die nächste Taste
  // löste sie erneut aus.
  eingabe.focus()

  return neu
}

function grenze(zahl: number, laenge: number): number {
  return Math.min(Math.max(zahl, 0), laenge)
}
