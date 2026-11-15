// **Welche Farbe welche Art bekommt** —
// #1164.
//
// # Warum die Zuordnung hier steht und nicht im Bauteil
//
// Weil sie eine Tabelle ist und keine Darstellung: `docs/KONTRASTE.md` hält
// jede dieser Farben gegen eine gemessene Zahl, und ein Wächter liest die
// Klassennamen aus dem Quelltext. In einer Datei beieinander sind sie zu
// prüfen; über ein JSX verteilt wären sie zu suchen.
//
// # Warum es fünf sind und nicht mehr
//
// Weil der Lexer fünf Arten kennt, die im Text stehen — Bezeichner, Zahl,
// Zeichenkette, Schlüsselwort, Operator — und seit #1164 den Kommentar dazu.
// Eine sechste Farbe für etwas, das er nicht unterscheidet, wäre eine
// Behauptung über die Sprache.
//
// # Und zwei für das Paketformat
//
// Seit #1181
// stehen `Schlüssel` und `Wert` dazu. Sie kommen aus einer **anderen** Sprache
// — dem Paketformat —, teilen sich diese Tabelle aber, weil beide durch
// dasselbe Bauteil gehen: ein `script:`-Wert in einer `.mse-style` trägt die
// Arten des Skript-Lexers **innerhalb** einer Paketdatei.
//
// **`Kommentar` gilt für beide.** In beiden Sprachen ist es `#` bis zum
// Zeilenende, und zwei Farben dafür wären zwei Antworten auf dieselbe Frage.

// farbe gibt die Klasse zu einer Art des Lexers.
//
// **Der Rückfall ist die Textfarbe selbst.** Eine Art, die dieses Fenster noch
// nicht kennt, bleibt lesbar — sie wird nur nicht hervorgehoben. Das ist die
// richtige Seite, auf der man sich irrt: eine unbekannte Art unsichtbar zu
// färben versteckte Quelltext.
export function farbe(art: string): string {
  return farben[art] ?? 'text-text'
}

const farben: Record<string, string> = {
  // **Ein Kommentar tritt zurück.** Er ist da, wenn man ihn sucht, und im Weg,
  // wenn er so laut ist wie der Code.
  Kommentar: 'text-leise',

  // **Ein Schlüsselwort ist die Struktur** — `if`, `then`, `else`. Es trägt
  // dieselbe Farbe wie ein Hauptknopf, weil es dieselbe Rolle hat: es sagt,
  // was als Nächstes geschieht.
  Schlüsselwort: 'text-marke',

  // **Eine Zeichenkette ist ein Wert und keine Rechnung.** Sie hebt sich ab,
  // weil ein fehlendes Anführungszeichen sonst erst der Zerteiler findet.
  Zeichenkette: 'text-zeichenkette',

  Zahl: 'text-zahl',

  // **Ein Operator bleibt blass.** Er steht zwischen allem und würde in
  // voller Farbe den Text zerhacken.
  Operator: 'text-leise',

  // **Ein Bezeichner bekommt die Textfarbe.** Er ist der Normalfall; ihn
  // einzufärben hieße, alles einzufärben.
  Bezeichner: 'text-text',

  // **Ein Schlüssel des Paketformats ist die Struktur** (#1181) — er trägt
  // dieselbe Farbe wie ein Schlüsselwort der Sprache, weil er dieselbe Rolle
  // hat: er sagt, worum es in der Zeile geht.
  Schlüssel: 'text-marke',

  // **Ein Wert bekommt die Textfarbe.** Er ist der Inhalt; ihn einzufärben
  // hieße, den Text einzufärben und die Struktur daneben verblassen zu
  // lassen.
  Wert: 'text-text',
}
