// **Der Arbeitsplatz und die Menüleiste** — UXR-F14
// (#2029).
//
// # Was hier zusammenkommt
//
// *Darstellung* führt „Navigator einblenden/ausblenden" und dasselbe für den
// Inspector, mit ⌥⌘[ und ⌥⌘]. Zwei Wege führen zum Arbeitsplatz:
//
//   - **hin**: der Eintrag meldet ein Ereignis (`menue:navigator`,
//     `menue:inspector`), und die Spalte klappt — **über dasselbe `klappen`**
//     wie Knopf und Taste, damit bei Enge die Spalte über dem Zentrum aufgeht
//     und nicht der Wunsch umspringt;
//   - **zurück**: wie die Spalte danach dasteht, geht an die Menüleiste, die
//     daraus Titel und Aktivierung macht (`darstellungsstand.ts`).
//
// Eine eigene Datei, weil `Arbeitsplatz.tsx` mit beidem über die Zeilengrenze
// aus §2 ginge. Der Schnitt folgt der Frage: dort, **wo** die Spalten stehen;
// hier, **was die Menüleiste davon erfährt und tut**.
//
// # Wann der Eintrag nichts zu tun hat
//
// **Unter einer Schublade** (UX-08, `verdeckung.ts`): die Tasten gehören dann
// nicht der Seite darunter — sie hängen sich gar nicht erst ein —, und derselbe
// Grund gilt für den Menüeintrag. Er ist gegraut (`vorhanden: false`), und ein
// Ereignis, das trotzdem ankäme, wird nicht gehört. Die Spalte darunter steht
// weiter da, und der graue Eintrag sagt, wie.
import {
  ereignisInspector,
  ereignisNavigator,
  useMenueereignis,
} from '../fenster'
import { useFensterteilmeldung, type Fensterteil } from '../darstellungsstand'
import { useCallback, useEffect, useRef } from 'react'

import type { Arbeitsplatzregion } from './useArbeitsplatzgedaechtnis'

// useArbeitsplatzmenue gibt den Ref für die Spalten über dem Zentrum zurück —
// **an ihm** nimmt eine über das Menü geöffnete Spalte den Fokus.
export function useArbeitsplatzmenue(
  spalten: Readonly<Record<Arbeitsplatzregion, Fensterteil>>,
  verdeckt: boolean,
  klappen: (region: Arbeitsplatzregion, offen: boolean) => void,
): (spalte: HTMLElement | null) => void {
  const { navigator, inspector } = spalten

  useFensterteilmeldung(
    'navigator',
    navigator.vorhanden && !verdeckt,
    navigator.offen,
  )
  useFensterteilmeldung(
    'inspector',
    inspector.vorhanden && !verdeckt,
    inspector.offen,
  )

  // **Ein Wunsch für genau einen Zug**: öffnet das Menü die Spalte daneben
  // (bei Platz), hängt sich keine über dem Zentrum ein, und der Effekt nach
  // dem Zug räumt den Wunsch ab — sonst nähme die nächste Überlagerung, die
  // ein Knopf öffnet, den Fokus, um den niemand gebeten hat.
  const fokusGewuenscht = useRef(false)
  useEffect(() => {
    fokusGewuenscht.current = false
  })

  useMenueereignis(ereignisNavigator, () => {
    if (navigator.vorhanden && !verdeckt) {
      fokusGewuenscht.current = !navigator.offen
      klappen('navigator', !navigator.offen)
    }
  })
  useMenueereignis(ereignisInspector, () => {
    if (inspector.vorhanden && !verdeckt) {
      fokusGewuenscht.current = !inspector.offen
      klappen('inspector', !inspector.offen)
    }
  })

  return useCallback((spalte: HTMLElement | null) => {
    if (spalte !== null && fokusGewuenscht.current) {
      fokusGewuenscht.current = false
      fokusHinein(spalte)
    }
  }, [])
}

// fokusHinein gibt einer Spalte, die das Menü **über dem Zentrum** geöffnet
// hat, den Fokus — gemessen am Fenster (`darstellung`, #2029): er blieb sonst,
// wo er war, und Escape schloss sie nicht (`escapeSchliesst` verlangt ihn in
// ihr). Das erste Bedienelement nimmt ihn, sonst die Spalte selbst.
//
// **Beim Einhängen und nicht im nächsten Bild**: mit `requestAnimationFrame`
// stand die Spalte am Fenster schon da, der Fokus nach einem schweren
// Navigator aber erst 300 ms später, und ein Escape dazwischen ging ins Leere.
function fokusHinein(ueber: HTMLElement): void {
  const erstes = ueber.querySelector<HTMLElement>(
    'button, a[href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
  )
  if (erstes !== null) {
    erstes.focus()

    return
  }

  ueber.tabIndex = -1
  ueber.focus()
}
