// **Ein getippter Klartext zurück in den ausgezeichneten Wert** —
// #687, Stufe 2.
//
// # Wofür
//
// Das Eingabefeld zeigte den rohen Wert: `<word-list-type-en>Creature</word-…>`
// stand als Erstes da, was jemand an diesem Feld sah. Wer die Marken löscht,
// verliert die Wortliste; wer sie stehen lässt, tippt um sie herum.
//
// Gezeigt wird jetzt der Klartext ([ohneAuszeichnung]) — und dieser Weg setzt
// ihn zwischen die Marken zurück.
//
// # Wie MSE es tut
//
// Dort schreibt der Editor **in** den ausgezeichneten Wert: er hält zwei
// Cursorstellen, rechnet sie über `untagged_to_index` in rohe Stellen um und
// ersetzt den Bereich dazwischen (`TextValueEditor::replaceSelection`,
// `src/gui/value/text.cpp`; `typing_action`, `src/data/action/value.cpp`
// Zeilen 187–203; `tagged_substr_replace`, `src/util/tagged_string.cpp`).
// Fassung f9d9356 vom 2021-07-06.
//
// **Hier gibt es keinen Cursor** — ein `<textarea>` meldet seinen neuen
// Inhalt, nicht die Stelle. Der Bereich wird deshalb **erschlossen**: der
// längste gemeinsame Anfang und das längste gemeinsame Ende sagen, was sich
// geändert hat. Für eine Eingabe an einer Stelle — tippen, löschen, einfügen —
// ist das genau der Bereich, den MSE bekäme.
//
// **Was das nicht kann**: zwei Änderungen an zwei Stellen in einem Rutsch
// werden als eine gelesen, die beide umschließt. Das kommt aus einem
// Eingabefeld nicht (jede Taste meldet sich einzeln), wohl aber aus einem
// Skript, das den Wert am Stück setzt — und dort ist der Wert ohnehin roh.
import {
  GESCHUETZT,
  istSchliessend,
  markeAn,
  rohStelle,
  VERBINDUNGSLEERZEICHEN,
} from './marken'
import { ohneAuszeichnung } from './ohneauszeichnung'

// mitAuszeichnung baut aus dem rohen Wert und dem getippten Klartext den neuen
// rohen Wert.
export function mitAuszeichnung(roh: string, klartext: string): string {
  const alt = ohneAuszeichnung(roh)
  if (alt === klartext) {
    return roh
  }

  const vorn = gemeinsamerAnfang(alt, klartext)
  const hinten = gemeinsamesEnde(alt, klartext, vorn)

  return mitEinsatz(
    roh,
    vorn,
    alt.length - hinten,
    geschuetzt(klartext.slice(vorn, klartext.length - hinten)),
  )
}

// mitEinsatz ersetzt einen Bereich des Klartextes durch ein **rohes** Stück.
//
// **Der Weg, den auch das Symbol nimmt.** Bei MSE ist das dieselbe Funktion:
// `replaceSelection` bekommt einmal getippten Text und einmal
// `<sym>R</sym>`, und beides geht durch `tagged_substr_replace`
// (`src/gui/value/text.cpp`, Zeilen 497 und 593; `src/util/tagged_string.cpp`).
// Fassung f9d9356 vom 2021-07-06.
//
// **Der Einsatz wird nicht geschützt** — er ist Auszeichnung, kein Text. Wer
// hier tippt, gibt ihn geschützt herein ([geschuetzt]).
export function mitEinsatz(
  roh: string,
  vonKlar: number,
  bisKlar: number,
  einsatz: string,
): string {
  const anfang = rohStelle(roh, vonKlar)
  const ende = rohStelle(roh, bisKlar)

  const { schliessungen, oeffnungen } = markenIm(roh, anfang, ende)

  // **Die Reihenfolge ist MSEs**: erst schließen, was im entfernten Bereich
  // geschlossen wurde, dann der Einsatz, dann wieder öffnen, was dort geöffnet
  // wurde.
  //
  //	substr_replace(input, start, end,
  //	  get_tags(collect_tags, …, false, true) + // close tags
  //	  replacement +
  //	  get_tags(collect_tags, …, true, false))  // open tags
  //
  // `tagged_substr_replace`, `src/util/tagged_string.cpp`, Fassung f9d9356 vom
  // 2021-07-06. Ohne sie bliebe eine Marke offen, sobald jemand über eine
  // Grenze hinweg löscht — und ein unpaariger Wert ist beim nächsten Zeichnen
  // ein Befund.
  return (
    roh.slice(0, anfang) +
    schliessungen.join('') +
    einsatz +
    oeffnungen.join('') +
    roh.slice(ende)
  )
}

// geschuetzt macht aus einem getippten `<` MSEs `ESCAPED_LANGLE`.
//
//	wxUniChar tag_char(wxUniChar c) {
//	  if (c == _('<')) return ESCAPED_LANGLE;
//	  else             return c;
//	}
//
// `src/util/tagged_string.cpp`, Zeilen 19–22, Fassung f9d9356 vom 2021-07-06.
// **Ohne das würde aus einem getippten `3 < 4` beim nächsten Lesen der Anfang
// einer Marke** — und der Text bräche ab.
function geschuetzt(text: string): string {
  // **Und die beiden reservierten Zeichen fallen heraus**
  // (#718).
  //
  // `ohneAuszeichnung` deutet sie um — `ESCAPED_LANGLE` als `<`,
  // `CONNECTION_SPACE` als Leerzeichen. Ein getippter Text, der eines von
  // beiden wirklich enthält, käme deshalb anders zurück, als er hineinging:
  // der Zweitblick durch Codex hat 230 Rundreisen gefahren, und **alle 46
  // Fehlschläge** hingen daran.
  //
  // **Kodiert werden sie nicht, denn MSE hat dieselbe Zweideutigkeit**:
  // `escape()` bildet allein `<` ab (`src/util/tagged_string.cpp`, Zeilen
  // 19–22, Fassung f9d9356 vom 2021-07-06), und ein `\1` in der Eingabe ist
  // danach von einem geschützten `<` nicht mehr zu unterscheiden. Wer eine
  // eigene Darstellung erfände, schriebe etwas, das kein anderes Programm
  // liest.
  //
  // In einem Kartentext sind beide ohnehin nichts: ein Steuerzeichen und ein
  // Zeichen aus dem privaten Bereich.
  return text
    .replaceAll(GESCHUETZT, '')
    .replaceAll(VERBINDUNGSLEERZEICHEN, '')
    .replaceAll('<', GESCHUETZT)
}

// gemeinsamerAnfang gibt die Zahl der Zeichen, die beide Fassungen teilen.
function gemeinsamerAnfang(alt: string, neu: string): number {
  const grenze = Math.min(alt.length, neu.length)

  let stelle = 0
  while (stelle < grenze && alt[stelle] === neu[stelle]) {
    stelle++
  }

  return stelle
}

// gemeinsamesEnde gibt die Zahl der Zeichen am Ende, die beide teilen.
//
// **Der gemeinsame Anfang begrenzt es**: sonst zählte bei `aa` → `aaa`
// derselbe Buchstabe zweimal, und der Bereich wäre negativ.
function gemeinsamesEnde(alt: string, neu: string, vorn: number): number {
  const grenze = Math.min(alt.length, neu.length) - vorn

  let stelle = 0
  while (
    stelle < grenze &&
    alt[alt.length - stelle - 1] === neu[neu.length - stelle - 1]
  ) {
    stelle++
  }

  return stelle
}

// markenIm sammelt, was der entfernte Bereich an Marken unpaarig zurückließe.
//
// **Zwei Sorten, und beide sind ein Loch:**
//
//   - eine **schließende** Marke, deren öffnende vor dem Bereich steht: die
//     öffnende bleibt stehen und hätte keinen Schluss mehr.
//   - eine **öffnende** Marke, deren schließende hinter dem Bereich steht: die
//     schließende bleibt stehen und hätte keinen Anfang mehr.
//
// Was **innerhalb** des Bereichs geöffnet und geschlossen wird, fällt ganz
// weg — das ist der `<sep>` zwischen zwei Feldern, wenn jemand über ihn
// hinwegschreibt.
function markenIm(
  roh: string,
  anfang: number,
  ende: number,
): { schliessungen: string[]; oeffnungen: string[] } {
  const offen: string[] = []
  const schliessungen: string[] = []

  let stelle = anfang
  while (stelle < ende) {
    const laenge = markeAn(roh, stelle)
    if (laenge === 0) {
      stelle++

      continue
    }

    const marke = roh.slice(stelle, stelle + laenge)
    stelle += laenge

    if (!istSchliessend(roh, stelle - laenge)) {
      offen.push(marke)

      continue
    }

    const gegenstueck = offen.lastIndexOf('<' + marke.slice(2))
    if (gegenstueck < 0) {
      schliessungen.push(marke)

      continue
    }

    offen.splice(gegenstueck, 1)
  }

  return { schliessungen, oeffnungen: offen }
}
