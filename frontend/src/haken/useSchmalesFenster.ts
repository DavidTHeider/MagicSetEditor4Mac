// **Ob das Fenster schmal ist** —
// #1630.
//
// **Die Grenze ist 768 Bildpunkte**, dieselbe wie Tailwinds `md`. Darunter
// bliebe dem Inhalt neben einer Leiste von 240 Bildpunkten weniger als die
// Hälfte des Fensters — bei 370 waren es gemessene **89**, und jede Zeile
// brach nach einem Wort.
//
// **Gefragt wird die Medienabfrage und nicht `innerWidth`.** Sie meldet den
// Wechsel von selbst; ein Horcher auf `resize` liefe bei jedem Bildpunkt und
// müsste entprellt werden.
import { useEffect, useState } from 'react'

export const schmalGrenze = 768

// **Ein Fenster ohne `matchMedia` gilt als breit.** Es gibt eines: jedes
// Dokument, das nicht in einem Browser steht — dieselbe Überlegung wie in
// `tisch/bewegung.ts`, und dieselbe Gestalt: die Typen des Browsers sagen,
// `matchMedia` sei immer da, und dann ist jede Frage danach „unnötig".
interface Fenster {
  matchMedia?: (frage: string) => MediaQueryList
}

function abfrage(grenze: number): MediaQueryList | undefined {
  const fenster: Fenster = window

  return fenster.matchMedia?.(`(max-width: ${String(grenze - 1)}px)`)
}

export function useSchmalesFenster(grenze: number = schmalGrenze): boolean {
  const [schmal, setzeSchmal] = useState(
    () => abfrage(grenze)?.matches ?? false,
  )

  useEffect(() => {
    const liste = abfrage(grenze)

    if (liste === undefined) {
      return
    }

    // **Noch einmal fragen, bevor gehorcht wird**: zwischen dem ersten Lesen
    // und dem Effekt liegt ein Zeichnen, und in ihm kann sich die Breite
    // geändert haben (§7, Regel 2 in ihrer kleinsten Gestalt).
    const nachsehen = () => {
      setzeSchmal(liste.matches)
    }

    nachsehen()

    // **Gelesen wird die Liste, nicht das Ereignis.** Gemessen am 2026-09-19
    // im Vorschaubereich: ein `change` kam mit `matches: false` an, während
    // `matchMedia(…).matches` im selben Augenblick schon `true` sagte — das
    // Ereignis trug den Stand von **vorher**. Wer ihm glaubt, steht mit einer
    // Antwort da, die nicht zu ihrem Stand gehört (§7, Regel 2).
    liste.addEventListener('change', nachsehen)

    // **Und `resize` dazu**, aus demselben Anlass: dort, wo das eine Ereignis
    // zu früh kam, kam das andere richtig. Ein zweiter Weg zum selben Wert
    // kostet nichts — React bricht ab, wenn der Zustand gleich bleibt.
    window.addEventListener('resize', nachsehen)

    return () => {
      liste.removeEventListener('change', nachsehen)
      window.removeEventListener('resize', nachsehen)
    }
  }, [grenze])

  return schmal
}
