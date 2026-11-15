// **Wo der Halt der Seitenleiste liegt** — #1488: auf der aufgeschlagenen
// Seite, und er folgt ihr. Aus `rahmen/Rahmen.tsx` herausgezogen, weil die
// Datei an der Zeilengrenze stand (§2). Der Roving-Tabindex selbst ist
// `useRovierendeZeile` — derselbe wie in der Paketliste.
import { useEffect } from 'react'

import { type Seitenschluessel } from '../navigation'
import { leisteneintraege } from '../leistenordnung'
import { useRovierendeZeile, type RovierendeZeile } from './useRovierendeZeile'

export function useLeistenhalt(aktuell: Seitenschluessel): RovierendeZeile {
  const rovierend = useRovierendeZeile(leisteneintraege.length)
  const { rueckenAuf } = rovierend

  useEffect(() => {
    const stelle = leisteneintraege.findIndex(
      (eintrag) => eintrag.ziel === aktuell,
    )

    // **Nicht jede Seite ist ein Eintrag der Leiste** — seit UX-02 führt die
    // Leiste Bereiche, und *Turniere*, *Vorlagen* und jede Lasche des Sets
    // stehen nicht darin. Dann bleibt der Halt, wo er war: beim ersten
    // Eintrag oder dort, wohin jemand gewandert ist.
    if (stelle >= 0) {
      rueckenAuf(stelle)
    }
  }, [aktuell, rueckenAuf])

  return rovierend
}
