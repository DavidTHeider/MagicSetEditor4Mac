// **Der Stand der Datei-Läufe einer Art** — FMT-03.AK-02
// (#1692).
//
// Go meldet jeden Lauf beim Beginn, bei jedem gezählten Stück und am Ende
// (`app.EreignisAuftrag`). **Gehört wird, nicht gefragt**: ein Bogen über
// dreihundert Karten meldet dreihundertmal, und dreihundert Rufe nach dem
// ganzen Buch wären Arbeit für eine Zahl.
import { useEffect, useState } from 'react'
import { ereignisAuftrag, horche } from '../fenster'
import type { AuftragDTO } from '../apiaufmerksamkeit'

// istAuftrag prüft die Gestalt, die über die Brücke kam — sie ist
// `unknown`, bis jemand hinsieht.
function istAuftrag(wert: unknown): wert is AuftragDTO {
  return (
    typeof wert === 'object' &&
    wert !== null &&
    typeof (wert as { art?: unknown }).art === 'string'
  )
}

export function useLaufstand(art: string): AuftragDTO | undefined {
  const [stand, setzeStand] = useState<AuftragDTO>()

  useEffect(
    () =>
      horche(ereignisAuftrag, (daten) => {
        const auftrag = daten[0]
        if (istAuftrag(auftrag) && auftrag.art === art) {
          setzeStand(auftrag)
        }
      }),
    [art],
  )

  return stand
}
