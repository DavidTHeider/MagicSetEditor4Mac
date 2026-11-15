// **Die Marke für einen Druck ohne Alias** — Werkstück A2 von Phase 5.3
// (#1441, 0086).
//
// Ein Exemplar, ein Deckplatz oder ein Verweis kann auf einen Druck zeigen,
// den dieser Bestand nur als Zeile kennt — aus einer Liste ohne
// Scryfall-Kennung. **Das ist ein Zustand und kein Fehler**, und er soll zu
// sehen sein: Preise und Bilder bleiben aus, bis jemand ihn auflöst. Eine
// Marke an vier Stellen statt vier Marken — sie sagen dasselbe.
import { useTexte } from '../sprache/kontext'

export function Vorlaeufig({ vorlaeufig }: { vorlaeufig: boolean }) {
  const eigene = useTexte().bausteine.druck

  if (!vorlaeufig) {
    return null
  }

  return (
    <span
      className="rounded-sm border border-kante px-1 text-xs text-leise"
      title={eigene.vorlaeufigErklaerung}
    >
      {eigene.vorlaeufig}
    </span>
  )
}
