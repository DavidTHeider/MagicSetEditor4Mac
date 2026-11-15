// **Der Fokus fällt nicht ins Leere** — Werkstück U14 von Phase 5.1 (#1490):
// lag er in der Seitenleiste, als sie wegklappte, steht er danach auf dem
// Knopf. Aus `rahmen/Rahmen.tsx` herausgezogen, weil die Datei an der
// Zeilengrenze stand (§2).
//
// Der Effekt läuft nach dem Zeichnen — die Leiste ist dann schon aus dem
// Baum, und `activeElement` ist auf `body` gefallen; deshalb merkt sich ein
// Horcher **davor**, ob der Fokus drin lag.
import { useEffect, useRef, type RefObject } from 'react'

export interface Leistenfokus {
  leistenknopf: RefObject<HTMLButtonElement | null>
  leistenbereich: RefObject<HTMLDivElement | null>
}

export function useLeistenfokus(leisteOffen: boolean): Leistenfokus {
  const leistenknopf = useRef<HTMLButtonElement>(null)
  const leistenbereich = useRef<HTMLDivElement>(null)
  const fokusLagInDerLeiste = useRef(false)

  useEffect(() => {
    const merken = () => {
      fokusLagInDerLeiste.current =
        leistenbereich.current?.contains(document.activeElement) ?? false
    }
    document.addEventListener('focusin', merken)

    return () => {
      document.removeEventListener('focusin', merken)
    }
  }, [])

  useEffect(() => {
    if (!leisteOffen && fokusLagInDerLeiste.current) {
      fokusLagInDerLeiste.current = false
      leistenknopf.current?.focus()
    }
  }, [leisteOffen])

  return { leistenknopf, leistenbereich }
}
