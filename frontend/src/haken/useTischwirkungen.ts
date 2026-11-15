// **Was am Tisch gerade geschehen ist, über zwei Bilder hinweg** —
// PLAY-04.AK-02
// (#1712).
//
// # Warum die Fassung der Schlüssel ist und nicht die Partie selbst
//
// **Weil ein Sichtwechsel keine Handlung ist.** Wer am Tisch die Sicht
// wechselt, bekommt eine andere Projektion **derselben** Fassung — andere
// Objekte, dieselbe Wahrheit (G8,
// #1465).
// Ein Vergleich über die Objektidentität ließe dabei den halben Tisch
// aufblitzen, und zwar für etwas, das niemand getan hat.
//
// **Die Fassung zählt Ereignisse.** Genau dann, wenn sie sich ändert, ist
// etwas geschehen — und genau dann wird verglichen.
import { useState } from 'react'

import { standVon, wirkungen } from '../tisch/tischwirkungen'

import type { Partie } from '../apipartie'
import type { Tischstand, Wirkungsart } from '../tisch/tischwirkungen'

export interface Tischwirkungen {
  // jeDing ordnet einer Objekt- oder Sitzkennung ihre Wirkung zu.
  //
  // **Eine je Ding**: eine Karte, die getappt aufs Feld kommt, tut zweierlei,
  // und eine Marke kann nur eines tragen. Die **erste** gewinnt — der
  // Zonenwechsel ist die größere Auskunft, und `wirkungen` gibt ihn zuerst.
  jeDing: Map<string, Wirkungsart>

  // amTisch sind die Wirkungen, die an keinem Ding hängen — Zugwechsel und
  // Spielende.
  amTisch: Wirkungsart[]
}

interface Gedaechtnis {
  fassung: string
  vorher: Tischstand | undefined
  jetzt: Tischstand
}

export function useTischwirkungen(partie: Partie): Tischwirkungen {
  const [gemerkt, merke] = useState<Gedaechtnis>(() => ({
    fassung: partie.fassung,
    vorher: undefined,
    jetzt: standVon(partie),
  }))

  // **Angepasst beim Zeichnen und nicht in einem Effekt** — React nennt es
  // „adjusting state during render": ein `useEffect` zeichnete das Bild erst
  // ohne Wirkung und gleich darauf mit, und das eine Zwischenbild ist genau
  // das Flackern, das hier niemand will.
  let stand = gemerkt

  if (gemerkt.fassung !== partie.fassung) {
    stand = {
      fassung: partie.fassung,
      vorher: gemerkt.jetzt,
      jetzt: standVon(partie),
    }

    merke(stand)
  }

  const jeDing = new Map<string, Wirkungsart>()
  const amTisch: Wirkungsart[] = []

  for (const wirkung of wirkungen(stand.vorher, stand.jetzt)) {
    if (wirkung.woran === '') {
      amTisch.push(wirkung.art)

      continue
    }

    if (!jeDing.has(wirkung.woran)) {
      jeDing.set(wirkung.woran, wirkung.art)
    }
  }

  return { jeDing, amTisch }
}
