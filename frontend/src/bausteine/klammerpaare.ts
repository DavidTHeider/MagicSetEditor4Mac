// **Das Klammerpaar unter dem Schreibzeiger** — Werkstück B von Phase 4.9
// (#1324).
//
// # Woher das kommt
//
// **Nicht aus MSE2.** Dort gibt es keinen Skripteditor: die Konsole ist ein
// einzeiliges `wxTextCtrl` (`src/gui/set/console_panel.cpp`, Zeile 399,
// Fassung f9d9356 vom 2021-07-06), und Vorlagen werden außerhalb des
// Programms bearbeitet. Gesucht wurde über `src/` nach `wxStyledTextCtrl`
// und `Scintilla` (0 Fundstellen), in `src/gui/value/` nach einem Skriptfeld
// und nach `ScriptEditor`/`script_editor` (0). Diese Einheit ist **`new`**,
// und das steht hier, damit niemand später eine Herkunft sucht, die es nicht
// gibt (§15).
//
// # Warum über die Merkmale und nicht über den Text
//
// **Weil eine Klammer in einer Zeichenkette keine ist.** `f("(")` hat ein
// Paar und nicht zwei; wer über die Zeichen zählt, findet drei öffnende. Der
// Lexer in Go weiß das bereits — er liefert die Merkmale, die das Feld
// ohnehin für seine Farben holt. Ein zweiter Zähler in TypeScript liefe an
// dem Tag auseinander, an dem einer von beiden eine Regel lernt; derselbe
// Satz steht seit #1164 über den Farben.
//
// # Warum es reine Rechnung ist
//
// Wie `wortstelle.ts`: Merkmale, Text und eine Stelle hinein, ein Paar
// heraus. Ohne DOM, ohne React — und damit ohne Vorrichtung zu prüfen.
import type { Merkmalstelle } from '../antworten'

// Klammerpaar ist die Klammer am Zeiger und ihr Gegenstück.
//
// **`partner` darf `null` sein**, und das ist der interessante Fall: eine
// Klammer ohne Gegenstück ist ein Befund, den man sehen soll, bevor der
// Zerteiler ihn meldet.
export interface Klammerpaar {
  hier: Merkmalstelle
  partner: Merkmalstelle | null
}

// paare ordnet jeder Klammer ihr Gegenstück zu — in **beide** Richtungen.
//
// **Eine Tabelle und keine Suche.** Der erste Anlauf drehte die Richtung mit
// `Object.keys(…).find(…)` um; das ist dieselbe Aussage zweimal geschrieben,
// und die zweite kann von der ersten abweichen.
const paare: Record<string, string> = {
  '(': ')',
  '[': ']',
  '{': '}',
  ')': '(',
  ']': '[',
  '}': '{',
}

// oeffnend sagt, ob eine Klammer aufmacht — die Richtung der Suche.
const oeffnend = '(['.concat('{')

// Klammerstelle ist eine Klammer samt allem, was aus ihrem Zeichen folgt.
//
// **Damit die Nachschlage-Rückfälle dorthin wandern, wo es sie gibt.** Der
// erste Anlauf filterte auf Klammern und schlug `paare[zeichen]` später noch
// einmal nach — mit einem `?? ''` für den Fall, dass es keine ist. Den Fall
// gibt es dort nicht mehr, und der Zweig blieb unbetreten. Hier steht die
// Umkehrung: das Nachschlagen **ist** der Filter.
interface Klammerstelle {
  merkmal: Merkmalstelle
  zeichen: string
  gegen: string
  auf: boolean
}

// klammerpaar gibt die Klammer bei `stelle` samt Gegenstück.
//
// **Der Zeiger zählt links wie rechts.** Wer hinter eine Klammer tippt, meint
// sie; wer davorsteht, auch. Das ist die Gewohnheit jedes Editors, und wer
// nur eine Seite nähme, ließe die Hervorhebung beim Tippen flackern.
//
// Gibt `null`, wenn dort keine Klammer steht.
export function klammerpaar(
  merkmale: readonly Merkmalstelle[],
  wert: string,
  stelle: number,
): Klammerpaar | null {
  // **Die Zeilenlängen einmal.** `versatz` läuft je Merkmal, und eine Vorlage
  // hat mehrere hundert Zeilen; den Text dort jedes Mal neu zu zerlegen wäre
  // dieselbe Rechnung so oft, wie Klammern dastehen.
  const laengen = wert.split('\n').map((zeile) => zeile.length)

  const klammern = merkmale.flatMap((merkmal) =>
    klammerstelle(merkmal, zeichenVon(wert, laengen, merkmal)),
  )

  const hier =
    klammerBei(klammern, laengen, stelle) ??
    klammerBei(klammern, laengen, stelle - 1)

  if (hier === undefined) {
    return null
  }

  return { hier: hier.merkmal, partner: gegenstueck(klammern, hier) }
}

// klammerstelle macht aus einem Merkmal eine Klammerstelle — oder nichts.
//
// **Der eine Ort, an dem entschieden wird, ob etwas eine Klammer ist.** Was
// hier durchkommt, trägt sein Gegenstück schon bei sich; danach wird nichts
// mehr nachgeschlagen und nichts mehr abgesichert.
function klammerstelle(merkmal: Merkmalstelle, zeichen: string) {
  if (merkmal.art !== 'Operator' || merkmal.breite !== 1) {
    return []
  }

  const gegen = paare[zeichen]
  if (gegen === undefined) {
    return []
  }

  return [{ merkmal, zeichen, gegen, auf: oeffnend.includes(zeichen) }]
}

// gegenstueck sucht die Partnerklammer — vorwärts oder rückwärts.
//
// **Über eine Tiefe und nicht über die nächste gleiche.** `f(g(1), 2)` hat
// zwei schließende, und die erste gehört zur inneren. Die Tiefe ist die
// einzige Rechnung, die das unterscheidet.
//
// **Und sie zählt nur die eigene Art.** `(a]` ist kein Paar: wer die Arten
// gemeinsam zählte, fände in `(a]` eines und behauptete damit, der Text sei
// in Ordnung.
function gegenstueck(
  klammern: readonly Klammerstelle[],
  hier: Klammerstelle,
): Merkmalstelle | null {
  const reihe = hier.auf
    ? klammern.slice(klammern.indexOf(hier) + 1)
    : klammern.slice(0, klammern.indexOf(hier)).reverse()

  let tiefe = 0

  for (const kandidat of reihe) {
    if (kandidat.zeichen === hier.zeichen) {
      tiefe += 1

      continue
    }

    if (kandidat.zeichen !== hier.gegen) {
      continue
    }

    if (tiefe === 0) {
      return kandidat.merkmal
    }

    tiefe -= 1
  }

  return null
}

// klammerBei gibt die Klammer, die bei `stelle` **beginnt**.
function klammerBei(
  klammern: readonly Klammerstelle[],
  laengen: readonly number[],
  stelle: number,
): Klammerstelle | undefined {
  if (stelle < 0) {
    return undefined
  }

  return klammern.find((eine) => versatz(laengen, eine.merkmal) === stelle)
}

// versatz rechnet Zeile und Spalte in eine Stelle im Text um.
//
// **Über die Zeilenlängen und nicht über eine Schleife durch den Text**: die
// Anzeige rechnet ohnehin in Zeilen, und beide zählen dann dasselbe.
function versatz(laengen: readonly number[], merkmal: Merkmalstelle): number {
  let bei = 0

  for (const laenge of laengen.slice(0, merkmal.zeile - 1)) {
    bei += laenge + 1
  }

  return bei + merkmal.spalte - 1
}

// zeichenVon gibt das Zeichen, auf dem ein Merkmal steht.
function zeichenVon(
  wert: string,
  laengen: readonly number[],
  merkmal: Merkmalstelle,
): string {
  const bei = versatz(laengen, merkmal)

  return wert.slice(bei, bei + 1)
}
