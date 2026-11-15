// **„3 Exemplare → Deck"** — UX-03.AK-01
// (#1675),
// nach dem Band im Navigator des Artboards *Sammlung*.
//
// **Die Zahl zählt Stücke und nicht Zeilen**: eine Zeile „3 × Near Mint" sind
// drei Exemplare. Zwei Zeilen desselben Drucks — etwa zwei Zustände — gehen
// als **ein** Druck mit ihrer Summe ins Deck: ein Deckplatz nennt einen Druck
// und eine Menge, keinen Zustand.
import { Knopf } from './Knopf'
import { useAnschluss } from '../anschlussraum'
import { einsetzen } from '../sprache/einsetzen'
import { zaehlwort } from '../sprache/zaehlwort'
import { useTexte } from '../sprache/kontext'

import type { Stueck } from '../anschluss'
import type { ExemplarDTO } from '../api'

// stueckeAus fasst Exemplare je Druck zusammen, in der Reihenfolge der Auswahl.
export function stueckeAus(exemplare: readonly ExemplarDTO[]): Stueck[] {
  const nachDruck = new Map<string, Stueck>()

  for (const exemplar of exemplare) {
    const bisher = nachDruck.get(exemplar.printing)

    nachDruck.set(exemplar.printing, {
      printing: exemplar.printing,
      name: bisher?.name ?? (exemplar.name || exemplar.bezeichnung),
      menge: (bisher?.menge ?? 0) + exemplar.menge,
    })
  }

  return [...nachDruck.values()]
}

export function Auswahlband({
  exemplare,
}: {
  exemplare: readonly ExemplarDTO[]
}) {
  const texte = useTexte()
  const eigene = texte.seiten.sammlung
  const anschluss = useAnschluss()

  if (exemplare.length === 0) {
    return null
  }

  const stuecke = stueckeAus(exemplare)
  const anzahl = zaehlwort(
    stuecke.reduce((summe, stueck) => summe + stueck.menge, 0),
    texte.allgemein.zaehlwoerter.exemplare,
  )

  return (
    <div
      role="group"
      aria-label={eigene.auswahlTitel}
      className="flex items-center gap-2 rounded-sm border border-kante bg-betont px-3 py-2"
    >
      <Knopf
        gewicht="haupt"
        aufKlick={() => {
          anschluss.uebergib({
            art: 'deckstapel',
            name: anzahl,
            stuecke,
          })
        }}
      >
        {einsetzen(eigene.auswahlInsDeck, { anzahl })}
      </Knopf>
    </div>
  )
}
