// **Die Klassen des Quelltextfelds** — wie das Feld aussieht, getrennt von
// dem, was es tut (§2). Ausgelagert, weil `Quelltextfeld.tsx` mit dem Sprung
// aus U7 über die Zeilengrenze ging.

// masse sind die Maße, die Anzeige und Eingabe teilen.
//
// **Eine Konstante und nicht zwei Klassenlisten.** Sie müssen Zeichen für
// Zeichen übereinanderliegen; wer sie getrennt pflegt, merkt die Abweichung
// erst an der Stelle, an der eine Farbe danebensteht.
const masse = 'px-2 py-1 font-mono text-sm leading-5'

// zeilenhoehe ist, wie hoch eine Zeile in `leading-5` steht — in `rem`.
//
// **Aus derselben Klasse abgeleitet wie die Anzeige.** Wer sie hier anders
// annähme, bekäme ein Feld, dessen Höhe nicht durch seine Zeilen teilbar ist.
export const zeilenhoehe = 1.25

// Die drei Klassenlisten stehen **außerhalb des JSX**.
//
// Nicht aus Ordnungsliebe: `keintextimjsx.test.ts` steigt in jedes Attribut
// hinab und meldet eine Zeichenkette mit Platzhalter als möglichen
// Oberflächentext. Das ist richtig so — über diesen Weg fängt es `titel=…` und
// jede andere beschriftende Eigenschaft, die nicht in seiner Liste steht.
// Eine Klassenliste ist keine Aufschrift, und sie gehört deshalb dorthin, wo
// der Wächter nicht sucht.
export const nummernklassen =
  'shrink-0 select-none border-r border-kante bg-betont/60 ' +
  'text-right text-leise/60 ' +
  masse

export const anzeigeklassen = 'm-0 min-h-full whitespace-pre ' + masse

export const eingabeklassen =
  'absolute inset-0 h-full w-full resize-none whitespace-pre border-0 ' +
  'bg-transparent text-transparent caret-slate-900 outline-none ' +
  'placeholder:text-leise dark:caret-slate-100 ' +
  'dark:placeholder:text-leise ' +
  masse
