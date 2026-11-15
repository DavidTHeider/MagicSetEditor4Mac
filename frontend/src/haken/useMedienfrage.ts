// **Ob eine Medienabfrage zutrifft** — für die zwei Systemeinstellungen, die
// UX-09 (#1681)
// mitlesen muss: mehr Kontrast (dann gilt keine Farbwelt) und dunkel (dann
// wird die Lesbarkeit neu gemessen).
//
// **Dieselbe Gestalt wie `useSchmalesFenster`**: ohne `matchMedia` gilt
// „nein", und gelesen wird die Liste, nicht das Ereignis.
import { useEffect, useState } from 'react'

interface Fenster {
  matchMedia?: (frage: string) => MediaQueryList
}

function abfrage(frage: string): MediaQueryList | undefined {
  const fenster: Fenster = window

  return fenster.matchMedia?.(frage)
}

export function useMedienfrage(frage: string): boolean {
  const [trifft, setzeTrifft] = useState(() => abfrage(frage)?.matches ?? false)

  useEffect(() => {
    const liste = abfrage(frage)
    if (liste === undefined) {
      return
    }

    const nachsehen = () => {
      setzeTrifft(liste.matches)
    }

    nachsehen()
    liste.addEventListener('change', nachsehen)

    return () => {
      liste.removeEventListener('change', nachsehen)
    }
  }, [frage])

  return trifft
}
