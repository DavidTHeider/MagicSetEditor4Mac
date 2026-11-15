// **Eine Bühne für eine Übergabe** — die Zielfläche steht, und ein Knopf
// übergibt, was die Befehlspalette übergäbe (UX-02.AK-02, #1674).
//
// **Die Fläche steht immer**, gleich welche Seite der Raum für aktuell hält:
// geprüft wird, was sie aufnimmt, nicht, wohin der Raum wechselt — das prüft
// `anschlussraum.test.tsx`.
import { useState, type ReactNode } from 'react'

import { Anschlussraum, useAnschluss } from '../anschlussraum'

import type { Anschluss } from '../anschluss'
import type { Seitenschluessel } from '../navigation'

export const uebergabeknopf = 'übergeben'

function Uebergabe({ anschluss }: { anschluss: Anschluss }) {
  const stelle = useAnschluss()

  return (
    <button
      type="button"
      onClick={() => {
        stelle.uebergib(anschluss)
      }}
    >
      {uebergabeknopf}
    </button>
  )
}

export function Anschlussbuehne({
  anschluss,
  children,
}: {
  anschluss: Anschluss
  children: ReactNode
}) {
  const [aktuell, setzeAktuell] = useState<Seitenschluessel>('start')

  return (
    <Anschlussraum aktuell={aktuell} wechseln={setzeAktuell}>
      <Uebergabe anschluss={anschluss} />
      {children}
    </Anschlussraum>
  )
}
