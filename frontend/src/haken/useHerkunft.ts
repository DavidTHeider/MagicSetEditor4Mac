// **Von welcher Seite man in die Hilfe kam** — UXR-22
// (#1930).
//
// Das Artboard stellt der Hilfe „Wo du gerade warst" voran: wer sie
// aufschlägt, hat eine Frage zu der Seite, von der er kommt. Die Hilfe selbst
// weiß das nicht — sie bleibt aufgehoben stehen (UX-03.AK-02) und sieht nur,
// dass sie wieder sichtbar wird.
import { useState } from 'react'

import type { Seitenschluessel } from '../navigation'

// useHerkunft gibt die Seite, die vor dem letzten Wechsel zur Hilfe offen
// war.
//
// **Beim Zeichnen und nicht in einem Effekt** — dieselbe Gestalt wie
// `gesehen` in der Hilfe: die Herkunft hängt allein an der Folge der Seiten,
// und die steht beim Zeichnen schon fest. **Beim Weggehen bleibt sie**: die
// aufgehobene Hilfe zeigt sie weiter, bis man von einer anderen Seite kommt.
export function useHerkunft(
  seite: Seitenschluessel,
): Seitenschluessel | undefined {
  const [stand, setzeStand] = useState<{
    seite: Seitenschluessel
    herkunft: Seitenschluessel | undefined
  }>({ seite, herkunft: undefined })

  if (stand.seite !== seite) {
    const herkunft = seite === 'hilfe' ? stand.seite : stand.herkunft
    setzeStand({ seite, herkunft })

    return herkunft
  }

  return stand.herkunft
}
