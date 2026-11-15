// **Quelltext ohne Kommentare, für die Wächter** — ausgezogen aus
// `farbsprache.test.ts` (AP-2), seit `hauptgewicht.test.ts` (AP-3) ihn auch
// braucht: wer eine Testdatei importiert, registriert deren Tests ein
// zweites Mal.
import ts from 'typescript'

// ohneKommentare gibt den Quelltext ohne seine Kommentare zurück. **Jeder
// Kommentar ist Vor- oder Nachtrivia eines Tokens** — auch der in
// `{/* … */}` —, und deshalb genügt es, die Blätter des Baums abzugehen.
// Beide braucht es: was auf der Zeile des Tokens davor steht, rechnet
// TypeScript zu dessen Nachtrivia.
export function ohneKommentare(datei: string, text: string): string {
  if (datei.endsWith('.css')) {
    return text.replace(/\/\*[\s\S]*?\*\//g, ' ')
  }

  const quelle = ts.createSourceFile(
    datei,
    text,
    ts.ScriptTarget.Latest,
    true,
    datei.endsWith('.tsx') ? ts.ScriptKind.TSX : ts.ScriptKind.TS,
  )
  const bereiche = new Map<number, number>()
  const blaetter = (knoten: ts.Node): void => {
    const kinder = knoten.getChildren(quelle)
    if (kinder.length === 0) {
      for (const bereich of [
        ...(ts.getLeadingCommentRanges(text, knoten.getFullStart()) ?? []),
        ...(ts.getTrailingCommentRanges(text, knoten.getEnd()) ?? []),
      ]) {
        bereiche.set(bereich.pos, bereich.end)
      }
    }
    kinder.forEach(blaetter)
  }
  blaetter(quelle)

  return [...bereiche]
    .sort(([a], [b]) => b - a)
    .reduce(
      (rest, [anfang, ende]) =>
        rest.slice(0, anfang) + ' '.repeat(ende - anfang) + rest.slice(ende),
      text,
    )
}
