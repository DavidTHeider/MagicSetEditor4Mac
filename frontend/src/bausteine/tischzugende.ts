// **Wann ein Zug am Tisch endet, und was er solange sperrt** — Befund 3 des
// Zweitblicks über Phase 6.9 (PLAY-05.AK-04,
// #1713).
//
// # Eine eigene Datei
//
// **Der Schnitt folgt der Frage** (§2): `tischziehen.ts` rechnet, *wohin*
// ein Zug geht — hier steht, *wie lange* er dauert. Der Anlass war die
// Zeilengrenze von `Tischflaeche.tsx`; dass beides zusammenfiel, ist der
// übliche Fall.
import { useCallback, useEffect } from 'react'

import type { Tischlauf } from '../tisch/useTisch'

// useZugende beendet einen Zug, der neben dem Tisch losgelassen oder mit
// Escape abgebrochen wird.
//
// **Die Fläche hört nur auf ihr eigenes Loslassen.** Wer eine Karte über den
// Rand hinaus zog und daneben losließ, hatte einen Zug, der nie endete: die
// Karte blieb verschoben stehen und folgte beim nächsten Überfahren dem
// Zeiger ohne gedrückte Taste. **Gemeldet wird dabei nichts** — neben dem
// Tisch liegt keine Zone.
//
// **Das Fenster hört nach der Fläche**: ein Loslassen auf dem Tisch erreicht
// zuerst deren eigenen Horcher, der den Zug meldet, und erst danach diesen.
export function useZugende(
  zieht: boolean,
  setzeZiehen: (ziehen: undefined) => void,
) {
  useEffect(() => {
    if (!zieht) return

    const ende = () => {
      setzeZiehen(undefined)
    }

    // **Escape bricht ab** — A04 (#2218): die Karte geht zurück, gemeldet
    // wird nichts, und das Loslassen danach findet keinen Zug mehr. **Die
    // Taste ist dann genommen**: eine Schublade unter dem Tisch schlösse
    // sonst mit (G12, das Innerste zuerst).
    const taste = (ereignis: KeyboardEvent) => {
      if (ereignis.key !== 'Escape') return

      ereignis.preventDefault()
      ereignis.stopPropagation()
      ende()
    }

    window.addEventListener('pointerup', ende)
    window.addEventListener('pointercancel', ende)
    window.addEventListener('keydown', taste, true)

    return () => {
      window.removeEventListener('pointerup', ende)
      window.removeEventListener('pointercancel', ende)
      window.removeEventListener('keydown', taste, true)
    }
  }, [zieht, setzeZiehen])
}

// useAbzugOhneZug gibt den Abzug des Tischs — außer während eines Zugs.
//
// **Dann zeigt der Szenenbaum den gezogenen Plan** und nicht den bestätigten,
// und ein Bild davon hielte einen Stand fest, den niemand bestätigt hat.
// Mit der Maus lässt sich das Ausgabefeld währenddessen nicht bedienen — die
// hält die Karte —, über die Tastatur schon.
export function useAbzugOhneZug(
  zieht: boolean,
  tischabzug: Tischlauf['abzug'],
): Tischlauf['abzug'] {
  return useCallback<Tischlauf['abzug']>(
    async (wahl, dpi) => (zieht ? { fehler: 'zieht' } : tischabzug(wahl, dpi)),
    [zieht, tischabzug],
  )
}
