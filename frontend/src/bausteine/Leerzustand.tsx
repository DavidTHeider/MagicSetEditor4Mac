// Ein leerer Zustand, der den nächsten Schritt erklärt.
//
// **„Keine Daten" ist keine Auskunft** (AGENTS.md §6). Wer eine leere Seite
// sieht, weiß nicht, ob etwas fehlt, ob er etwas falsch gemacht hat oder ob er
// etwas tun soll — und genau das steht hier.
//
// **Die Handlung steht gleich dabei.** Ein leerer Zustand, der auf eine andere
// Seite verweist, verlangt einen Weg, den man sich merken muss.
import type { ReactNode } from 'react'

interface LeerzustandEigenschaften {
  titel: string
  erklaerung: string
  children?: ReactNode
}

export function Leerzustand({
  titel,
  erklaerung,
  children,
}: LeerzustandEigenschaften) {
  return (
    <div
      data-leerzustand=""
      className="rounded-lg border border-dashed border-kante p-10 text-center"
    >
      <p className="text-lg font-medium">{titel}</p>

      <p className="mx-auto mt-2 max-w-prose text-leise">{erklaerung}</p>

      {children !== undefined && (
        <div className="mt-5 flex justify-center gap-2">{children}</div>
      )}
    </div>
  )
}
