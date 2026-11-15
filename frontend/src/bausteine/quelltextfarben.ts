// **Welche Bindung eine Sprache fragt** —
// #1181.
//
// **Aus `Quelltextfeld.tsx` herausgelöst**, als es über 400 Zeilen ging (§2).
// Der Schnitt folgt der Frage: dort steht, **wie** ein Quelltextfeld aussieht
// und sich bedienen lässt; hier, **woher** seine Farben kommen.
//
// # Zwei Bindungen und kein Schalter darin
//
// Eine `.mse-style` ist das **Paketformat** und keine Sprache: `game: magic`
// ist ein Schlüssel mit einem Wert, kein Ausdruck. Welche Bindung gefragt
// wird, entscheidet der Aufrufer — ein Parameter „welche Sprache" an der
// Bindung wäre eine fachliche Verzweigung in `internal/app`, und dort gehört
// keine hin (§2).
import { api } from '../fassade'
import { type Merkmalstelle } from '../antworten'
import type { Sprache } from './Quelltextfeld'

// holeFarben gibt die Merkmale zu einem Text in seiner Sprache.
//
// **`keine` fragt nichts.** Was dort steht, ist keine Sprache, die dieses
// Programm kennt — eine Konsole etwa nimmt beliebigen Text.
export async function holeFarben(
  sprache: Sprache,
  wert: string,
): Promise<Merkmalstelle[]> {
  if (sprache === 'script') {
    return (await api().Skriptfarben(wert)).merkmale
  }

  if (sprache === 'paket') {
    return (await api().Paketfarben(wert)).merkmale
  }

  return []
}
