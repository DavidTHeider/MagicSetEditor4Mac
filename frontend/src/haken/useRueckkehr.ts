// **Der Fokus eines Überbaus — beim Erscheinen hinein, beim Verschwinden
// zurück** — und nur dann. Angestrichen vom Zweitblick über Phase 6.3
// (G11).
//
// # Warum ein Haken
//
// **Vier Bauteile taten dasselbe mit einem Effekt ohne Abhängigkeiten**:
// `Entscheidung`, `Schublade`, `Popover` und der erste Schritt von `SetNeu`
// merkten sich beim Erscheinen den Auslöser, setzten den Fokus und gaben ihn
// beim Verschwinden zurück. **Seit die Seiten aufbewahrt werden**
// (`<Activity>`, Vermerk 0133) laufen solche Effekte beim Verbergen ab und
// beim Einblenden neu: der Überbau zog den Fokus aus der Leiste, mit der
// jemand zurückkam, und merkte sich die Leiste als Rückkehrziel. Dieselbe
// Klasse fand die WebKit-Probe `ux10-tastatur` an der Listenansicht.
//
// # Wie es verborgen von verschwunden unterscheidet
//
// **Am Knoten**: `<Activity>` lässt ihn beim Verbergen im Baum stehen
// (`display: none`), ein echtes Verschwinden nimmt ihn heraus, bevor React
// die Aufräumarbeit eines Effekts ruft. Gemerkt wird er beim Einrichten des
// Effekts — die Referenz selbst kann beim Aufräumen schon leer sein.
import { useEffect, useRef, type RefObject } from 'react'

import { fokusZurueck, merkeRueckkehr } from '../bausteine/fokus'

// **Steht der Fokus im Überbau, geht er zurück — auch beim Verbergen**: eine
// Schublade des Sets steht selbst in einer `<Activity>`, und Zuklappen ist
// dort Verbergen. Ohne diese Ausnahme blieb der Fokus nach Escape in der
// verborgenen Schublade (gefunden nach dem Umsetzen auf `main`,
// `setschubladen.test.tsx`).
//
// **`nurWennVerloren`** gibt den Fokus nur zurück, wenn er mit dem Überbau
// ging — wer inzwischen woanders arbeitet, dem nimmt ein spätes Schließen
// ihn nicht (das Popover, `Popover.tsx`).
//
// **`ersatz`** nennt das Ziel, wenn das gemerkte nichts taugt — `body` oder
// ein Knoten, den es nicht mehr gibt (QA01, #2382): stand der Fokus beim
// Aufgehen eines Popovers auf `body`, blieb er nach Escape dort, gegen den
// Vertrag „zurück zum Auslöser“. Seit AP-6 (#2363) öffnet ein Menüeintrag
// das Popover der Decks — einer der Wege, auf denen das Gemerkte nicht der
// Auslöser ist.
export function useRueckkehr(
  anker: RefObject<Element | null>,
  fokussieren?: () => void,
  nurWennVerloren = false,
  ersatz?: () => Element | null | undefined,
): void {
  const rueckkehr = useRef<Element | null>(null)
  const erschienen = useRef(false)
  const fokus = useRef(fokussieren)
  const zweites = useRef(ersatz)

  useEffect(() => {
    fokus.current = fokussieren
    zweites.current = ersatz
  })

  useEffect(() => {
    const knoten = anker.current

    if (!erschienen.current) {
      erschienen.current = true
      rueckkehr.current = merkeRueckkehr()
      fokus.current?.()
    }

    return () => {
      const aktiv = document.activeElement
      const verloren = aktiv === null || aktiv === document.body
      const drinnen = aktiv !== null && knoten?.contains(aktiv) === true
      if (
        !drinnen &&
        (knoten?.isConnected === true || (nurWennVerloren && !verloren))
      ) {
        return
      }

      const ziel = rueckkehr.current
      const taugt = ziel?.isConnected === true && ziel !== document.body
      fokusZurueck(taugt ? ziel : (zweites.current?.() ?? ziel))
    }
  }, [anker, nurWennVerloren])
}
