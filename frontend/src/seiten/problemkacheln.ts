// **Was die Prüfstände auf die Kacheln schreiben** — AP-8, Teil 2
// (#2365),
// Befund B-13: eine Unterzeile je Kachel und, wo die Zahl einen trägt, ein
// Ziel mit gesetztem Filter.
//
// **Eine Zahl nur, wenn jedes Set aktuell geprüft ist** — sonst sagt die
// Kachel „1 von 2 geprüft“ und „nicht vollständig geprüft“, und das Ziel ist
// die ganze Liste. Ein Filter über Ungeprüftes wäre eine Behauptung, einer
// ohne Treffer eine Sackgasse.
import type { Anschluss } from '../anschluss'
import { einsetzen } from '../sprache/einsetzen'
import type { Katalog } from '../sprache/kataloge'
import { zaehlwort } from '../sprache/zaehlwort'
import type { Problemlage } from './problemlage'

export type Kachel = 'sets' | 'karten' | 'exemplare' | 'decks'

export function problemkacheln(
  lage: Problemlage | undefined,
  texte: Katalog,
): {
  // **Nur die zwei Kacheln der Prüfstände** — „Sammlung“ und „Decks“
  // schreiben ihre Unterzeile aus anderen Antworten (Teil 2b).
  unterzeile: (kachel: 'sets' | 'karten') => string | undefined
  auswahl: (kachel: Kachel) => Anschluss | undefined
} {
  const eigene = texte.seiten.start.ueberblick
  const woerter = texte.allgemein.zaehlwoerter

  return {
    unterzeile: (kachel) => {
      if (lage === undefined || lage.gesamt === 0) {
        return undefined
      }

      if (kachel === 'sets') {
        return lage.vollstaendig
          ? einsetzen(eigene.mitProblemen, {
              anzahl: String(lage.mitProblemen.length),
            })
          : einsetzen(eigene.geprueftVon, {
              geprueft: String(lage.geprueft),
              gesamt: String(lage.gesamt),
            })
      }

      return lage.vollstaendig
        ? einsetzen(eigene.mitBefund, { anzahl: String(lage.karten) })
        : eigene.nichtGeprueft
    },

    auswahl: (kachel) => {
      if (lage?.vollstaendig !== true || lage.mitProblemen.length === 0) {
        return undefined
      }

      const sets = zaehlwort(lage.mitProblemen.length, woerter.sets)
      const setIDs = lage.mitProblemen.map((befund) => befund.setID)

      if (kachel === 'sets') {
        return {
          art: 'setauswahl',
          name: einsetzen(eigene.setsMitProblemen, { sets }),
          setIDs,
        }
      }

      // **Ein Set, und die Kachel führt in seine Kartenliste** — gefiltert
      // auf die Karten mit Befund und benannt wie die Zeile der
      // Aufmerksamkeit (Teil 2b): die Setliste mit einer Zeile wäre ein
      // Umweg.
      const [einziges, zweites] = lage.mitProblemen
      if (
        kachel === 'karten' &&
        einziges !== undefined &&
        zweites === undefined
      ) {
        return {
          art: 'kartenauswahl',
          name: einsetzen(texte.seiten.start.aufmerksamkeit.setBefunde, {
            karten: zaehlwort(einziges.zahl, woerter.karten),
          }),
          setID: einziges.setID,
          karteIDs: einziges.karteIDs,
        }
      }

      if (kachel === 'karten') {
        return {
          art: 'setauswahl',
          name: einsetzen(eigene.kartenMitBefund, {
            karten: zaehlwort(lage.karten, woerter.karten),
            sets,
          }),
          setIDs,
        }
      }

      return undefined
    },
  }
}
