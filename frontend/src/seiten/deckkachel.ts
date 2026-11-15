// **Was die Kachel „Decks“ unter ihre Zahl schreibt** — AP-8, Teil 2b
// (#2365),
// Entwurf `Start09`.
//
// # Fehlende Karten und nicht „ungültig im Format“
//
// **Das Artboard sagt „1 ungültig im Format“**, und das ist hier nicht zu
// haben, ohne zu fragen: die Zulässigkeit eines Drucks ist eine Auskunft von
// Scryfall (`service/formatpruefung.go`), sie wird nicht gespeichert, und
// eine Startseite, die für jedes Deck ins Netz ginge, täte es ohne Anweisung
// (§16, Regel 5). **Was ohne Netz feststeht, ist der Abgleich mit der
// Sammlung** — die Aufmerksamkeit zählt ihn schon (`deckFehlend`), und die
// Kachel liest dieselbe Antwort. Die Abweichung steht in
// `docs/PROTOTYP-0.9.0.md`.
//
// # Keine Null
//
// **Ein Deck, dessen Abgleich scheiterte, steht nicht in den Hinweisen**,
// sondern in den Befunden. „0 mit fehlenden Karten“ wäre dann eine
// Behauptung; deshalb steht die Zahl nur da, wo sie etwas zählt.
import type { Anschluss } from '../anschluss'
import type { AufmerksamkeitAntwort } from '../apiaufmerksamkeit'
import { einsetzen } from '../sprache/einsetzen'
import type { Katalog } from '../sprache/kataloge'
import { zaehlwort } from '../sprache/zaehlwort'

export function deckkachel(
  daten: AufmerksamkeitAntwort | undefined,
  texte: Katalog,
): { unterzeile?: string; ziel?: Anschluss } {
  const fehlend = (daten?.hinweise ?? []).filter(
    (hinweis) => hinweis.art === 'deckFehlend',
  )
  const [einziges, zweites] = fehlend

  if (einziges === undefined) {
    return {}
  }

  const eigene = texte.seiten.start.ueberblick
  const unterzeile = einsetzen(eigene.mitFehlenden, {
    anzahl: String(fehlend.length),
  })

  // **Ein Deck, und die Kachel öffnet es mit seiner Prüfung** — wie die
  // Zeile der Aufmerksamkeit; eine Liste mit einer Zeile wäre ein Umweg.
  if (zweites === undefined) {
    return {
      unterzeile,
      ziel: {
        art: 'deckansicht',
        name: einziges.objekt.name,
        deckID: einziges.objekt.kennung,
        lasche: 'pruefung',
      },
    }
  }

  return {
    unterzeile,
    ziel: {
      art: 'deckauswahl',
      name: einsetzen(eigene.decksMitFehlenden, {
        decks: zaehlwort(fehlend.length, texte.allgemein.zaehlwoerter.decks),
      }),
      deckIDs: fehlend.map((hinweis) => hinweis.objekt.kennung),
    },
  }
}
