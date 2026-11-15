// **„Bei Ada mitspielen“ kommt an** — P9c
// (#2377),
// S13 in zwei Handlungen: der Beitritt reist von der Deckseite hierher, wo der
// Lauf des Gastes lebt.
//
// **Einmal je Wunsch**, dieselbe Form wie „Schnell spielen“
// (`useSitzwuensche`): aufgenommen wird beim Zeichnen und nur der eigene
// Zustand gesetzt; beigetreten wird danach in einem Effekt, der sich die
// Nummer merkt.
import { useEffect, useRef, useState } from 'react'

import { useAnschluss, useAufgenommen } from '../anschlussraum'

import type { Anschluss } from '../anschluss'
import type { Gastlauf } from './usePartiegast'

type Mitspielen = Extract<Anschluss, { art: 'mitspielen' }>

export function useMitspielen(gast: Gastlauf): void {
  const anschluss = useAnschluss()
  const mitgereist =
    anschluss.offen?.art === 'mitspielen' ? anschluss.offen : undefined

  const [wunsch, setzeWunsch] = useState<{
    nummer: number
    beitritt: Mitspielen
  }>()

  useAufgenommen(mitgereist, (beitritt) => {
    setzeWunsch((bisher) => ({ nummer: (bisher?.nummer ?? 0) + 1, beitritt }))
  })

  const erledigt = useRef(0)
  useEffect(() => {
    if (wunsch === undefined || wunsch.nummer <= erledigt.current) return

    erledigt.current = wunsch.nummer
    anschluss.verbrauche()
    gast.beitreten(wunsch.beitritt.gegenueber, '', {
      deckID: wunsch.beitritt.deckID,
      leiheID: '',
    })
  })
}
