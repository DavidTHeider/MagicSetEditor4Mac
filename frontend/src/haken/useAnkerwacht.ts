// **Ein offenes Menü folgt seinem Anker, auch ohne Rollen** — Rest R2 des
// Nachlaufs zum Zweitblick über Phase 6 von 0.9.0. Rollen und Verkleinern
// hören `Menue` und `Kontextmenue` selbst; eine Liste, die sich umordnet,
// meldet nichts. Die Deckliste sortiert „zuletzt benutzt“ und lädt nach dem
// Öffnen nach, und die Zeile rückte unter einem offenen Menü weg.
//
// # Je Bild ein Blick
//
// **Solange etwas offen ist, liest die Wache je Bild die Lage des Ankers**
// und die Größe des Fensters; ändert sich eines, meldet sie `folgen`. Ein
// Menü ist kurz offen, und ein `getBoundingClientRect` je Bild kostet
// nichts, was man misst. Einen Beobachter für „ein Element rückt“ gibt es
// nicht: `ResizeObserver` sieht nur seine Größe.
import { useEffect, useRef } from 'react'

// lage schreibt, wo der Anker steht — oder dass es ihn nicht mehr gibt.
function lage(anker: Element | null | undefined): string {
  const fenster = `${String(window.innerWidth)}×${String(window.innerHeight)}`
  if (anker?.isConnected !== true) {
    return `weg ${fenster}`
  }

  const rand = anker.getBoundingClientRect()

  return `${String(rand.left)} ${String(rand.top)} ${String(rand.right)} ${String(rand.bottom)} ${fenster}`
}

export function useAnkerwacht(
  aktiv: boolean,
  anker: () => Element | null | undefined,
  folgen: () => void,
): void {
  // **Über Verweise** — sonst begänne die Wache bei jedem Zeichnen neu, und
  // was sie zuletzt sah, wäre vergessen.
  const ankerJetzt = useRef(anker)
  const folgenJetzt = useRef(folgen)
  useEffect(() => {
    ankerJetzt.current = anker
    folgenJetzt.current = folgen
  })

  useEffect(() => {
    if (!aktiv) {
      return
    }

    let zuletzt = lage(ankerJetzt.current())
    const schauen = () => {
      const jetzt = lage(ankerJetzt.current())
      if (jetzt !== zuletzt) {
        zuletzt = jetzt
        folgenJetzt.current()
      }

      bild = requestAnimationFrame(schauen)
    }
    let bild = requestAnimationFrame(schauen)

    return () => {
      cancelAnimationFrame(bild)
    }
  }, [aktiv])
}
