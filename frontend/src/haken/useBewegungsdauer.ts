// **Die Dauer einer Bewegung, gelesen aus dem Stilblatt** — AP-4
// (#2361),
// Befund B-11, Vermerk 0065 Festlegung 10.
//
// # Warum keine Zahl hier
//
// **Die Grenze steht an einer Stelle**: `--dauer-kurz` (120 ms) und
// `--dauer-mittel` (160 ms) in `stil.css`, und `barrierefreiheit.css` setzt
// beide auf 0 ms, wenn jemand wenig Bewegung will — im System
// (`prefers-reduced-motion`) oder in der Anwendung (`data-bewegung`). Mantine
// braucht für seine Übergänge eine Zahl; stünde sie hier noch einmal, hätte
// die Sperre zwei Stellen und hielte nur eine.
//
// # Wann neu gelesen wird
//
// Wenn sich die Einstellung am Wurzelelement ändert oder die des Systems —
// sonst bliebe eine Schublade nach dem Umschalten bei ihrer alten Dauer.
import { useSyncExternalStore } from 'react'

// Bewegungsart ist eine der beiden Dauern aus Festlegung 10.
export type Bewegungsart = 'kurz' | 'mittel'

const wenigBewegung = '(prefers-reduced-motion: reduce)'

// inMillisekunden liest `160ms` oder `0.16s`; was keine Dauer ist, ist 0 —
// lieber keine Bewegung als eine geratene.
export function inMillisekunden(wert: string): number {
  const treffer = /^\s*(\d*\.?\d+)(ms|s)\s*$/.exec(wert)
  if (treffer === null) {
    return 0
  }

  const zahl = Number(treffer[1])

  return treffer[2] === 's' ? zahl * 1000 : zahl
}

function abonnieren(melden: () => void): () => void {
  const beobachter = new MutationObserver(melden)
  beobachter.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['data-bewegung', 'style', 'class'],
  })
  const frage =
    typeof window.matchMedia === 'function'
      ? window.matchMedia(wenigBewegung)
      : undefined
  frage?.addEventListener('change', melden)

  return () => {
    beobachter.disconnect()
    frage?.removeEventListener('change', melden)
  }
}

// useBewegungsdauer gibt die Dauer in Millisekunden, wie das Stilblatt sie
// gerade setzt.
export function useBewegungsdauer(art: Bewegungsart): number {
  return useSyncExternalStore(abonnieren, () =>
    inMillisekunden(
      getComputedStyle(document.documentElement).getPropertyValue(
        `--dauer-${art}`,
      ),
    ),
  )
}
