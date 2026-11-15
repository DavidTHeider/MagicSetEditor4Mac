// **Woher etwas kam** — Werkstück 12 von Phase 4.12
// (#1351),
// nach Vermerk [0071](../../docs/entscheidungen/0071-woher-ein-deck-kommt.md).
//
// # Warum sie verschwindet, statt „—" zu zeigen
//
// **Weil keine Herkunft schon eine Auskunft ist**: es ist hier entstanden.
// Eine Zeile „Herkunft: —" an jedem eigenen Deck wäre eine Auskunft über
// nichts, und §6 verlangt das Gegenteil — ein leerer Zustand erklärt den
// nächsten Schritt oder ist gar nicht da.
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { HerkunftDTO } from '../api'

export function Herkunftszeile({ herkunft }: { herkunft: HerkunftDTO }) {
  const texte = useTexte()
  const eigene = texte.bausteine

  if (herkunft.art === '') {
    return null
  }

  // **Zwei Sätze und kein zusammengesetzter.** „Übernommen aus dem Profil
  // {quelle}" und „Eingelesen aus {quelle}" sind verschiedene Aussagen; ein
  // gemeinsamer Rahmen mit ausgetauschtem Wort ließe sich nicht übersetzen,
  // ohne dass eine der beiden Sprachen schief steht (§9).
  const satz =
    herkunft.art === 'profil'
      ? eigene.herkunftAusProfil
      : eigene.herkunftAusEinfuhr

  return (
    <p className="text-leise">{einsetzen(satz, { quelle: herkunft.quelle })}</p>
  )
}
