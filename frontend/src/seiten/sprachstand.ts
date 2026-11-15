// **Die Sätze über eine Sprache** — H04
// (#2258):
// woher sie kommt und ob ein Mensch sie redigiert hat, aus `sprachherkunft`.
import { einsetzen } from '../sprache/einsetzen'
import { alsKalendertag } from '../zeitpunkt'

import type { kataloge, Sprachherkunft } from '../sprache/kataloge'

type Staendetexte = (typeof kataloge)['de']['seiten']['sprachstaende']

// herkunftstext sagt, woher die Texte einer Sprache kommen.
export function herkunftstext(
  texte: Staendetexte,
  herkunft: Sprachherkunft,
  darstellung: string,
): string {
  return einsetzen(texte.herkunft[herkunft.herkunft], {
    datum: alsKalendertag(herkunft.uebersetzt, darstellung),
  })
}

// redaktionstext sagt, ob und wann ein Mensch sie gelesen hat — **und
// behauptet keine Redaktion, die nicht stattfand**.
export function redaktionstext(
  texte: Staendetexte,
  herkunft: Sprachherkunft,
  darstellung: string,
): string {
  return herkunft.geprueft === ''
    ? texte.ungeprueft
    : einsetzen(texte.geprueft, {
        datum: alsKalendertag(herkunft.geprueft, darstellung),
      })
}
