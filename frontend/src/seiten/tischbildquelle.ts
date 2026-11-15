// **Woher der gezeichnete Tisch seine Bilder nimmt** — Werkstück G3 von Phase
// 5.5 (#1460).
//
// # Sie meldet an und antwortet in einem
//
// **Weil der Maler beides zugleich braucht.** Er fragt für jede Lage nach
// einem Bild; ist keines da, soll es **beschafft** werden, und bis dahin
// zeichnet er den Platzhalter. Zwei getrennte Wege — hier anmelden, dort
// fragen — hießen, die Liste der Lagen zweimal zu durchlaufen und beide Läufe
// synchron zu halten.
//
// **Gerufen wird sie aus einem Effekt**, nicht aus dem Render: `useTisch`
// zeichnet in `useEffect`, und `kartenbilder` fragt erst dort.
import { bildwunsch } from './tischbild'

import type { Bildquelle } from '../tisch/maler'
import type { Partie } from '../apipartie'
import type { Texturwunsch } from '../apitextur'
import type { Texturlauf } from '../haken/useSpieltexturen'

// Tischguete ist die Stufe, in der ein Tisch zeichnet.
//
// **Die niedrige.** Sechzig Karten in voller Auflösung wären sechzig Rasterläufe
// für Bilder, die auf dem Tisch keine drei Zentimeter groß sind; wer eine Karte
// genau ansehen will, schlägt sie auf (G2, #1459).
export const Tischguete = 'tisch'

export function tischbildquelle(
  partie: Partie,
  texturen: Texturlauf,
): Bildquelle {
  return (lage, verdeckt) => {
    // **Ein Rücken gehört dem Spiel und keiner Karte** — PLAY-03.AK-07
    // (#1711).
    //
    // Er steht **vor** dem Nachschlagen des Objekts, und das ist die Zusage:
    // wo verdeckt gemeint ist, wird die Karte gar nicht erst gesucht. Die
    // Geheimhaltung hängt damit nicht daran, dass niemand das Ergebnis
    // benutzt, sondern daran, dass dieser Zweig keine Karte anfasst.
    if (verdeckt) {
      const ruecken: Texturwunsch = {
        art: 'ruecken',
        spiel: partie.spiel,
        guete: Tischguete,
      }

      texturen.belege(ruecken)

      return texturen.adresse(ruecken)
    }

    const objekt = partie.objekte.find(
      (eines) => eines.kennung === lage.kennung,
    )
    if (!objekt) return undefined

    const wunsch = bildwunsch(objekt, partie.spiel, Tischguete, lage.seite)
    texturen.belege(wunsch)

    return texturen.adresse(wunsch)
  }
}
