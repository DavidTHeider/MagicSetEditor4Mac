// **„Die 7 eintauschen"** — UX-03.AK-01
// (#1675),
// nach dem Inspector des Artboards *Decks*.
//
// **Was fehlt, soll hereinkommen** — deshalb reist es als Wunsch **vom
// Gegenüber** zum Handel. **Eine Zeile ohne Druck bleibt draußen**: eine
// selbst entworfene Karte lässt sich bei niemandem holen.
import { Knopf } from './Knopf'
import { useAnschluss } from '../anschlussraum'
import { einsetzen } from '../sprache/einsetzen'
import { zaehlwort } from '../sprache/zaehlwort'
import { useTexte } from '../sprache/kontext'

import type { Stueck } from '../anschluss'
import type { ZuordnungszeileDTO } from '../apidecks'

export function Tauschwunschknopf({
  zeilen,
}: {
  zeilen: readonly ZuordnungszeileDTO[]
}) {
  const texte = useTexte()
  const eigene = texte.seiten.decks
  const anschluss = useAnschluss()

  const stuecke: Stueck[] = zeilen
    .filter((zeile) => zeile.fehlend > 0 && zeile.printing !== '')
    .map((zeile) => ({
      printing: zeile.printing,
      name: zeile.bezeichnung,
      menge: zeile.fehlend,
    }))

  if (stuecke.length === 0) {
    return null
  }

  const summe = stuecke.reduce((bisher, stueck) => bisher + stueck.menge, 0)
  const anzahl = String(summe)

  return (
    <div>
      <Knopf
        aufKlick={() => {
          anschluss.uebergib({
            art: 'tauschwunsch',
            name: zaehlwort(summe, texte.allgemein.zaehlwoerter.karten),
            stuecke,
          })
        }}
      >
        {einsetzen(eigene.eintauschen, { anzahl })}
      </Knopf>
    </div>
  )
}
