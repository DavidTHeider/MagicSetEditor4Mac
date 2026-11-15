// **Das Wort unter dem Schreibzeiger** — Werkstück 1 von Phase 3.3
// (#1164).
//
// # Warum das eine eigene Datei ist
//
// Weil es reine Rechnung ist: Text und eine Stelle hinein, ein Wort und seine
// Grenzen heraus. Ohne DOM, ohne React — und damit ohne Vorrichtung zu prüfen.
//
// # Was ein Wort hier ist
//
// **Was der Lexer einen Bezeichner nennt**: ein Buchstabe oder `_` am Anfang,
// danach Buchstaben, Ziffern und `_` (`istNamensanfang`, `istNamenszeichen` in
// `internal/msescript/lexer.go`). Der Punkt gehört **nicht** dazu — in
// `card.name` sind `card` und `name` zwei Wörter, und wer nach `card.na`
// vervollständigt, meint `name`.

// Wortstelle ist ein Wort samt seinen Grenzen im Text.
export interface Wortstelle {
  wort: string
  von: number
  bis: number
}

// wortstelle gibt das Wort, das bei `stelle` endet.
//
// **Es endet dort und beginnt davor.** Wer mitten in einem Wort steht und
// tippt, meint den Teil links vom Zeiger; was rechts steht, hat er schon
// geschrieben.
export function wortstelle(text: string, stelle: number): Wortstelle {
  let von = stelle

  while (von > 0 && istWortzeichen(text.charAt(von - 1))) {
    von -= 1
  }

  return { wort: text.slice(von, stelle), von, bis: stelle }
}

// passende gibt die Namen, die ein Wort fortsetzen — höchstens `hoechstens`.
//
// **Ohne Rücksicht auf Groß- und Kleinschreibung beim Suchen**, aber die
// Vorschläge stehen, wie sie heißen: MSE-Bezeichner sind durchweg klein, und
// wer `TO_` tippt, meint `to_upper`.
//
// **Ein Wort, das genau einem Namen entspricht, schlägt nichts vor.** Wer
// `card` fertig getippt hat, braucht keinen Vorschlag `card` — und eine Liste,
// die immer aufgeht, steht im Weg.
export function passende(
  namen: readonly string[],
  wort: string,
  hoechstens: number,
): string[] {
  if (wort === '') {
    return []
  }

  const klein = wort.toLowerCase()

  const treffer = namen.filter(
    (name) => name.toLowerCase().startsWith(klein) && name !== wort,
  )

  return treffer.slice(0, hoechstens)
}

// istWortzeichen sagt, ob ein Zeichen zu einem Bezeichner gehört.
export function istWortzeichen(zeichen: string): boolean {
  return /[\p{L}\p{N}_]/u.test(zeichen)
}
