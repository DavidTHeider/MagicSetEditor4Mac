// **Die Kennungen einer Feldzeile** —
// #1278,
// Befund 5.
//
// # Der Fehler, den es zu verhindern gilt
//
// `aria-describedby` ist eine **Liste von Kennungen, an Leerraum getrennt**.
// Die Zeile baute sie als `hilfe-${feld}` — und ein Feldname mit Leerzeichen
// ist der Normalfall: `archenemy.mse-game` nennt sein Textfeld `rule text`
// (Zeile 459), `magic.mse-game` `casting cost` und `card number`.
//
// Aus `hilfe-rule text` werden damit **zwei** Verweise: `hilfe-rule` und
// `text`. Keiner davon zeigt auf einen Knoten, und die zugängliche
// Beschreibung ist **leer**.
//
// **Das ist seit #1274 schlimmer als vorher**: seither steht die Beschreibung
// außerhalb des Fokus `sr-only`, und die Begründung dafür lautete, eine
// Vorlesehilfe lese sie weiterhin über `aria-describedby`. Genau diese Zusage
// ging ins Leere.
//
// # Warum kein Rutschen über einen Zähler
//
// Weil die Kennung **stabil** sein muss: sie steht am Beschreibungsknoten und
// am Feld, und beide werden getrennt gezeichnet. Ein Zähler über die
// gezeichneten Zeilen liefe auseinander, sobald ein Feld wegfällt.

// unbedenklich sind die Zeichen, die eine Kennung unverändert tragen darf.
//
// **Buchstaben, Ziffern und Unterstrich** — sonst nichts. Auch `-` nicht: es
// leitet unten die Ersetzung ein und muss deshalb selbst ersetzt werden, sonst
// wäre die Abbildung nicht umkehrbar.
const unbedenklich = /[A-Za-z0-9_]/

// feldkennung macht aus einem Feldnamen eine Kennung ohne Leerraum.
//
// # Warum jedes Zeichen einzeln und nicht ein `_` für alles
//
// **Weil zwei Namen sonst dieselbe Kennung bekommen.** Der erste Anlauf
// ersetzte jede Folge von Sonderzeichen durch `_` und hängte die Länge des
// Rohnamens an — damit wurden `a b`, `a.b`, `a:b`, `a/b` und `a#b` **alle** zu
// `a_b_3`. Zwei Felder eines fremden Pakets zeigten dann auf denselben
// Beschreibungsknoten, und das zweite wurde mit der Hilfe des ersten
// vorgelesen. Angestrichen vom Nachlauf über die Nacharbeit
// (#1282,
// Befund 6).
//
// Jetzt wird jedes nicht unbedenkliche Zeichen als `-<hex>-` geschrieben. Weil
// `-` selbst so geschrieben wird (`-2d-`), leitet ein `-` **immer** eine
// Ersetzung ein: die Abbildung ist damit umkehrbar und kann nicht kollidieren.
//
// Aus `rule text` wird `rule-20-text`. Lesbar genug, dass jemand im DOM sein
// Feld wiedererkennt — darum kein Hash.
export function feldkennung(feld: string): string {
  let gebaut = ''

  // **Über die Zeichen und nicht über die Bytes**: ein Paketname darf jedes
  // Zeichen führen, und ein halbiertes Ersatzzeichenpaar wäre keine Kennung.
  for (const zeichen of feld) {
    gebaut += unbedenklich.test(zeichen) ? zeichen : `-${alsHex(zeichen)}-`
  }

  return gebaut
}

// alsHex schreibt ein Zeichen als seine UTF-8-Bytes in Hex.
//
// **Über die Bytes und nicht über `codePointAt`**: das gibt `number |
// undefined`, und der Rückfall auf 0 wäre ein Zweig, den kein Aufruf betritt —
// ein Zeichen aus einer Zeichenkette ist nie leer. Ein Zweig, den kein Lauf
// erreicht, ist eine Behauptung darüber, was im Ernstfall geschieht (§3).
//
// **UTF-8 ist umkehrbar**, also bleibt die Abbildung eindeutig.
function alsHex(zeichen: string): string {
  return [...new TextEncoder().encode(zeichen)]
    .map((byte) => byte.toString(16).padStart(2, '0'))
    .join('')
}

// hilfekennung ist die Kennung des Beschreibungsknotens einer Zeile.
export function hilfekennung(feld: string): string {
  return `hilfe-${feldkennung(feld)}`
}

// fehlerkennung ist die Kennung des Fehlerknotens einer Zeile.
export function fehlerkennung(feld: string): string {
  return `fehler-${feldkennung(feld)}`
}

// Feldfehler ordnet eine Meldung dem Feld zu, das sie verursacht hat.
//
// **Am Feld, nicht als anonymer Hinweis** (§6). Ein Hinweis über dem Formular
// zwingt zum Suchen, und bei zwölf Feldern findet ihn niemand.
export interface Feldfehler {
  feld: string
  meldung: string
}
