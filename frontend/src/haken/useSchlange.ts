// **Wünsche der Reihe nach** — P5
// (#2373),
// Befunde C-21 und C-24: ein Befehl, während der vorige unterwegs war, wurde
// verworfen, und schnelle Gesten gingen still verloren.
//
// # Einer unterwegs, die anderen warten
//
// **§7, Regel 4 bleibt**: je Datensatz ein Vorgang — aber ein zweiter Wunsch
// wird nicht weggeworfen, sondern wartet, bis der erste beantwortet ist.
// **Was ein Wechsel überholt hat, verwirft die Arbeit selbst** (Regel 1 und
// 2): sie kennt den Stand, aus dem ein Wunsch kam.
import { useCallback, useLayoutEffect, useRef, useState } from 'react'

export interface Schlange<T> {
  // stelle reiht einen Wunsch ein und stößt die Arbeit an.
  stelle: (wunsch: T) => void

  // wartend ist, wie viele Wünsche noch nicht beantwortet sind — der
  // unterwegs eingeschlossen.
  wartend: number
}

export function useSchlange<T>(
  arbeit: (wunsch: T, ersterDerFolge: boolean) => Promise<void>,
): Schlange<T> {
  const reihe = useRef<T[]>([])
  const unterwegs = useRef(false)
  const [wartend, setzeWartend] = useState(0)

  // **Die Arbeit von heute, nicht die vom Einreihen** — sie hängt an der
  // Partie, die sich zwischen zwei Wünschen ändert.
  const heutige = useRef(arbeit)
  useLayoutEffect(() => {
    heutige.current = arbeit
  })

  // abarbeiten stellt die Wünsche nacheinander, bis keiner mehr ansteht —
  // **eine Schleife und kein Selbstaufruf**, sodass ein zweiter Anstoß,
  // während sie läuft, nur einreiht.
  const abarbeiten = useCallback(async (ersterDerFolge: boolean) => {
    if (unterwegs.current) {
      return
    }

    unterwegs.current = true
    let erster = ersterDerFolge

    try {
      for (
        let wunsch = reihe.current.shift();
        wunsch !== undefined;
        wunsch = reihe.current.shift()
      ) {
        await heutige.current(wunsch, erster)
        erster = false
        setzeWartend(reihe.current.length)
      }
    } finally {
      unterwegs.current = false
      setzeWartend(reihe.current.length)
    }
  }, [])

  const stelle = useCallback(
    (wunsch: T) => {
      const ersterDerFolge = !unterwegs.current && reihe.current.length === 0
      reihe.current.push(wunsch)
      setzeWartend((bisher) => bisher + 1)
      void abarbeiten(ersterDerFolge)
    },
    [abarbeiten],
  )

  return { stelle, wartend }
}
