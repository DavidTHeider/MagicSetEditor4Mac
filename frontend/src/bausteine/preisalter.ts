// **Wie alt ein Preis ist, in Worten** — Werkstück 3 von Phase 3.4
// (#1165).
//
// # Warum das an einer Stelle steht
//
// Die Zeile eines Drucks und die Übersicht sagen dasselbe über dasselbe. Der
// erste Anlauf hatte zwei Fassungen: die eine sagte „gerade eben", die andere
// „vor 0 h" — nebeneinander im selben Bild, gesehen im laufenden Fenster.
import { einsetzen } from '../sprache/einsetzen'

// Alterstexte sind die beiden Sätze, zwischen denen gewählt wird.
export interface Alterstexte {
  preisAlter: string
  preisFrisch: string
}

// alterstext sagt, wie alt ein Preis ist.
//
// **Unter einer Stunde heißt „gerade eben"** und nicht „vor 0 h" — eine Null
// in einer Altersangabe liest sich wie ein Fehler.
export function alterstext(
  texte: Alterstexte,
  stunden: number,
  quelle: string,
): string {
  if (stunden === 0) {
    return einsetzen(texte.preisFrisch, { quelle })
  }

  return einsetzen(texte.preisAlter, { stunden: String(stunden), quelle })
}
