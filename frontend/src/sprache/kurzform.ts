// **Ein kurzer Satz sichtbar, der Rest hinter dem Fragezeichen** —
// UX-15.AK-02 und AK-03
// (#1771).
//
// # Wofür es das gibt
//
// **Weil unter jedem Feld zwei bis fünf Zeilen standen.** Gemessen am
// 2026-09-22 über `sprache/de/einstellungshinweise.ts`: die Hinweise sind im
// Mittel 171 Zeichen lang und reichen bis 371 — ihre **ersten Sätze** im
// Mittel 25 und höchstens 114. Der erste Satz ist fast immer der, der beim
// Entscheiden hilft; der Rest ist die Begründung, und die liest man einmal.
//
// # Warum getrennt wird und nicht gekürzt
//
// **Weil keine Information verloren geht** — das Issue sagt es ausdrücklich.
// Was hier abgetrennt wird, steht hinter dem Hilfesymbol, und beides zusammen
// ist wieder der ganze Text. Ein Lauf daneben hält genau das.
//
// # Warum hier und nicht im Bauteil
//
// **Weil ein Wächter es messen soll** (AK-03: „an einer Grenze, die im Lauf
// steht"). Eine Teilung, die nur beim Zeichnen geschieht, ließe sich über den
// Katalog nicht prüfen.

// Satzende ist ein Punkt, ein Frage- oder ein Ausrufezeichen, dem ein
// Abstand folgt.
//
// **Der Abstand gehört dazu**: `1.2.3` und `z. B.` trennen sonst mitten im
// Satz. Gemessen am 2026-09-22 kommt in den deutschen Hinweisen keine
// Abkürzung mit Punkt und Abstand vor — käme eine, wäre sie hier zu nennen.
const Satzende = /[.!?]\s/

// kurzform gibt den ersten Satz eines Textes — samt seinem Schlusszeichen.
export function kurzform(text: string): string {
  const stelle = text.search(Satzende)

  return stelle === -1 ? text : text.slice(0, stelle + 1)
}

// Auszeichnung sind die zwei Formen, die die Kataloge kennen — `**fett**`
// und `` `wörtlich` ``.
const Auszeichnung = /\*\*([^*]+)\*\*|`([^`]+)`/g

// ohneAuszeichnung gibt den Text ohne seine Zeichen.
//
// **Für das Hilfesymbol** (UX-15.AK-02): dort steht der Text als
// Zeichenkette in einem `role="tooltip"`, und eine Vorlesehilfe läse die
// Sternchen mit. `bausteine/Satz.tsx` löst sie im sichtbaren Teil auf; hier
// fallen sie weg.
export function ohneAuszeichnung(text: string): string {
  return text.replace(
    Auszeichnung,
    (_, fett: string | undefined, woertlich: string) => fett ?? woertlich,
  )
}

// langform gibt, was nach dem ersten Satz kommt — oder nichts.
//
// **`undefined` und keine leere Zeichenkette**: daran hängt, ob es das
// Hilfesymbol überhaupt gibt, und ein Symbol, das eine leere Erklärung
// öffnet, ist eine Zusage ohne Einlösung (§10).
export function langform(text: string): string | undefined {
  const stelle = text.search(Satzende)

  if (stelle === -1) {
    return undefined
  }

  const rest = text.slice(stelle + 1).trim()

  return rest === '' ? undefined : rest
}
