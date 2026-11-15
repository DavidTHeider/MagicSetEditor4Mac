// **Was die Leiste oben verdeckt, darunter rollt der Fokus nicht** —
// gefunden am nativen Fenster in Phase 6.11 (#1672, Hüllenprobe
// `phase611`).
//
// Mit der Leiste oben (UX-14.AK-02), „Sehr groß" und 1.024 × 700 nimmt sie
// drei Zeilen — rund 370 von 700 Punkten —, und sie klebt. Tab setzte den
// Fokus auf „Nächste Karte" unter ihr: WebKit rollt ein fokussiertes Element
// ins Bild, rechnet die klebende Leiste aber nicht ab (WCAG 2.4.11).
//
// **Das Dokument rollt** (`min-h-screen` am Rahmen), und sein
// `scroll-padding-top` sagt jedem Rollen ins Bild, wie viel oben nicht
// zählt. Die Höhe folgt der Leiste, denn sie bricht je nach Schrift und
// Breite in eine bis drei Zeilen um.
import { useEffect, type RefObject } from 'react'

export function useLeistenabstand(
  bereich: RefObject<HTMLElement | null>,
  waagerecht: boolean,
): void {
  useEffect(() => {
    const leiste = bereich.current
    const wurzel = document.documentElement

    if (!waagerecht || leiste === null) {
      return undefined
    }

    const setzen = () => {
      wurzel.style.scrollPaddingTop = `${String(leiste.getBoundingClientRect().height)}px`
    }

    setzen()

    const beobachter = new ResizeObserver(setzen)
    beobachter.observe(leiste)

    return () => {
      beobachter.disconnect()
      wurzel.style.scrollPaddingTop = ''
    }
  }, [bereich, waagerecht])
}
