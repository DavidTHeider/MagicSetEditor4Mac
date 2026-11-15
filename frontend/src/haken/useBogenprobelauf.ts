// **Wie der Druckbogen würde** — UXR-F7
// (#2015).
//
// Der Probelauf fragt den Dienst mit **denselben Wahlen**, mit denen der Knopf
// danach legt: Papier, Anschnitt, Auflösung, Marken und die Karten. Ändert sich
// eine davon, beginnt eine neue Frage, und die Antwort der vorigen wird
// verworfen (§7, Regel 2 — `useLaden` zählt mit).
//
// **Das Vorige bleibt stehen, bis das Neue da ist** (`behalten`), und die
// Anzeige sagt dabei, dass gerechnet wird: eine Zahl, die bei jedem Klick
// verschwindet und wiederkommt, ist schlechter als eine, die einen Augenblick
// alt ist — aber nur, solange sie sich so zeigt. Ein **Wechsel des Sets**
// räumt trotzdem (§7, Regel 1) — das erledigt der Schlüssel des Assistenten je
// Set.
import { useCallback } from 'react'

import { api } from '../api'
import type { BogenprobelaufAntwort } from '../antwortenprobelauf'
import { auspacken } from '../umschlag'
import { useLaden, type Ladezustand } from './useLaden'

export function useBogenprobelauf(
  setID: string,
  papier: string,
  anschnitt: number,
  aufloesung: number,
  marken: boolean,
  auswahl: readonly string[],
): Ladezustand<BogenprobelaufAntwort> {
  const holen = useCallback(async () => {
    const antwort = await api().BogenProbelauf(
      setID,
      papier,
      anschnitt,
      aufloesung,
      marken,
      [...auswahl],
    )

    return auspacken(antwort, () => antwort)
  }, [setID, papier, anschnitt, aufloesung, marken, auswahl])

  return useLaden(
    holen,
    [setID, papier, anschnitt, aufloesung, marken, auswahl],
    true,
  )
}
