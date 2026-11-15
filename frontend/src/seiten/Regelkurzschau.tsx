// **Was der Tisch tut, auf einen Blick** — das Artboard *Eine Partie
// anfangen* (PLAY-02, S11.NAT-04): „setzt durch", „hilft mit", „von Hand",
// je mit ihren Bereichen und der Zahl von allen.
//
// # Warum eine Zusammenfassung und nicht die Liste
//
// **Weil die Liste zwischen den Sitzen und dem Startknopf stand** — mit zwei
// gewählten Decks war die Seite bis zum 2026-09-24 4.890 Punkte hoch
// (Befund F13). Die Liste bleibt, eingeklappt darunter (`Regelschau`); hier
// steht, was man vor dem Anfangen wissen muss.
//
// # Und warum aus der Matrix
//
// **Weil eine zweite Wahrheit über dieselbe Frage auseinanderliefe** (§16):
// die Grade kommen aus dem Kern, und ein Bereich ohne Eintrag ist
// `unbekannt` — dieselbe Vorgabe wie in der Regelschau und im Kern. Er steht
// als „offen" da und nicht als Zusage.
import { einsetzen } from '../sprache/einsetzen'
import { regelbereichwort } from './regelwoerter'
import { useTexte } from '../sprache/kontext'

import type { RegelfreigabeDTO } from '../apiregeln'

// grade sind die vier Grade des Kerns, vom stärksten zum schwächsten.
const grade = ['automatisch', 'begleitet', 'manuell', 'unbekannt'] as const

export function Regelkurzschau({ freigabe }: { freigabe: RegelfreigabeDTO }) {
  const eigene = useTexte().seiten.spielen

  const worte: Record<(typeof grade)[number], string> = {
    automatisch: eigene.gradKurzAutomatisch,
    begleitet: eigene.gradKurzBegleitet,
    manuell: eigene.gradKurzManuell,
    unbekannt: eigene.gradKurzUnbekannt,
  }

  // **Jeder Bereich genau einmal** — aus dem Baum, und was die Matrix
  // darüber hinaus nennt; dieselbe Menge, die die Regelschau zeigt.
  const bereiche = [
    ...new Set([
      ...freigabe.baum.flatMap((gruppe) => gruppe.bereiche),
      ...freigabe.matrix.map((eintrag) => eintrag.bereich),
    ]),
  ]

  const gradVon = (bereich: string) =>
    freigabe.matrix.find((eintrag) => eintrag.bereich === bereich)?.grad ??
    'unbekannt'

  const zeilen = grade
    .map((grad) => ({
      grad,
      bereiche: bereiche.filter((bereich) => gradVon(bereich) === grad),
    }))
    .filter((zeile) => zeile.bereiche.length > 0)

  return (
    <table className="w-full border-collapse text-left text-sm">
      <thead>
        <tr className="border-b border-kante text-leise">
          <th scope="col" className="py-1 pr-3 font-medium whitespace-nowrap">
            {eigene.kurzschauWas}
          </th>
          <th scope="col" className="py-1 font-medium">
            {eigene.kurzschauBereiche}
          </th>
        </tr>
      </thead>

      <tbody>
        {zeilen.map((zeile) => (
          <tr key={zeile.grad} className="border-b border-kante align-top">
            <th scope="row" className="py-1.5 pr-3 font-medium">
              {worte[zeile.grad]}
            </th>
            <td className="py-1.5">
              <span>
                {zeile.bereiche
                  .map((bereich) => regelbereichwort(eigene, bereich))
                  .join(', ')}
              </span>{' '}
              <span className="text-leise">
                {einsetzen(eigene.kurzschauAnteil, {
                  zahl: String(zeile.bereiche.length),
                  gesamt: String(bereiche.length),
                })}
              </span>
            </td>
          </tr>
        ))}
      </tbody>
    </table>
  )
}
