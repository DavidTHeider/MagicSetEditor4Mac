// **Die Sitze eines Anfangs** — Zustand und Herkunft, aus `Partieanfang.tsx`
// herausgelöst, als P6
// (#2374)
// ihn über die Zeilengrenze aus §2 brachte.
//
// # Woher ein Sitz kommt
//
// **Ein Deck, das mitgereist ist, sitzt schon** — UX-03.AK-01
// (#1675),
// auf dem ersten Sitz. **Die übrigen Lücken füllt die Vorbelegung** (P6,
// Artboard `PartieStart09`), einmal, sobald Decks und Quellen gelesen sind
// (`spiel/vorbelegung.ts`) — und erst dann steht das Formular da.
//
// # Schnell spielen
//
// **Reist ein Deck mit `sofort`**, zählt `schnell` hoch — der Anfang fängt
// dann von selbst an (`Partieanfang`). Ein Zähler und kein Schalter: ein
// zweites „Schnell spielen“ ist ein zweiter Wunsch.
import { useState } from 'react'

import { useAnschluss, useAufgenommen } from '../anschlussraum'
import { eingeladenerSitz, vorbelegung } from '../spiel/vorbelegung'
import { useVorbelegung, type Vorbelegungsdaten } from './useVorbelegung'

import type { Sitzwunsch } from '../seiten/Sitzliste'

const leer: Sitzwunsch = { name: '', deckID: '', profilID: '', leiheID: '' }

export function useSitzwuensche({
  decks,
  gegenueber,
  ersatzname,
}: {
  // decks sind die eigenen — `undefined`, solange sie laden.
  decks: readonly { id: string }[] | undefined
  gegenueber: string | undefined
  ersatzname: string
}) {
  const [wuensche, setzeWuensche] = useState<Sitzwunsch[]>([leer, leer])
  const [schnell, setzeSchnell] = useState(0)

  const setzeWunsch = (stelle: number, neu: Sitzwunsch) => {
    setzeWuensche((bisher) =>
      bisher.map((alt, lage) => (lage === stelle ? neu : alt)),
    )
  }

  const anschluss = useAnschluss()
  const mitgereist =
    anschluss.offen?.art === 'partie' ? anschluss.offen : undefined

  useAufgenommen(mitgereist, (angekommen) => {
    setzeWuensche((bisher) =>
      bisher.map((alt, lage) => {
        if (lage === 0) return { ...alt, deckID: angekommen.deckID }

        // **Mit einer Einladung ist der zweite Sitz der Gast** (P9c).
        return lage === 1 && angekommen.eingeladen !== undefined
          ? eingeladenerSitz(angekommen.eingeladen)
          : alt
      }),
    )
    if (angekommen.sofort === true) {
      setzeSchnell((bisher) => bisher + 1)
    }
  })

  const quellen: Vorbelegungsdaten | undefined = useVorbelegung()
  const [vorbelegt, setzeVorbelegt] = useState(false)
  if (!vorbelegt && decks !== undefined && quellen !== undefined) {
    setzeVorbelegt(true)
    const vorgabe = vorbelegung({
      decks,
      mitgebracht: mitgereist?.deckID,
      eingeladen: mitgereist?.eingeladen,
      ...quellen,
      gegenueber,
      ersatzname,
    })
    // **Sie ersetzt, statt Lücken zu füllen**: das Formular steht erst da,
    // wenn sie gerechnet ist — getippt hat noch niemand, und ein
    // mitgereistes Deck nimmt sie als `mitgebracht` selbst.
    setzeWuensche(vorgabe)
  }

  return {
    wuensche,
    setzeWuensche,
    setzeWunsch,
    mitgereist,

    // verbrauche räumt das Mitgereiste weg — **ein Anfang benutzt es**, und
    // ein Formular, das sich danach neu aufbaut, nähme es sonst ein zweites
    // Mal auf (bei „Schnell spielen“: die nächste Partie).
    verbrauche: anschluss.verbrauche,
    quellen,
    schnell: vorbelegt ? schnell : 0,
  }
}
