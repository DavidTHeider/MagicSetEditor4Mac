// **Wie hoch ein Bereich sein darf, damit er bis an den Fensterrand reicht**
// — Befund F09 aus der nativen Prüfung von 0.5.0
// (#1675,
// #1680).
//
// # Wofür
//
// **Weil der Tisch nach dem Anfangen außerhalb der Ansicht lag.** Vor ihm
// standen die Überschrift, zwei Erklärabsätze, die Lebensleiste, die
// Zugleiste, der Darstellungsschalter und die Werkzeugzeile; die Leinwand
// begann am unteren Fensterrand. Das Artboard *Tisch* gibt ihr die Höhe, die
// unter dem Kopf bleibt — und die kennt erst das Fenster.
//
// # Warum gemessen und nicht in CSS
//
// **Weil darüber stehen kann, was die Seite nicht kennt**: ein Meldungsband,
// das Band einer neuen Fassung, die Brotkrume. Eine feste Zahl in `calc()`
// stimmte, bis eines davon erscheint. Deshalb misst der Haken nach, wenn
// das Fenster **oder** das Dokument seine Größe ändert.
//
// # Warum vom Anfang des Dokuments
//
// **Weil die Höhe nicht mitscrollen soll.** Wer ein Stück hinunterscrollt,
// um unter den Tisch zu sehen, soll ihn nicht wachsen sehen; gerechnet wird
// deshalb von dem Stand, in dem die Seite oben steht.
import { useCallback, useSyncExternalStore } from 'react'

import type { RefObject } from 'react'

export interface Fensterrand {
  // unten ist der Abstand, der zum Fensterrand frei bleibt — der Innenrand
  // des Inhalts.
  unten: number

  // mindestens ist die Höhe, unter die der Bereich nicht fällt: in einem
  // flachen Fenster scrollt die Seite, statt den Tisch zu zerdrücken.
  mindestens: number
}

// useFensterrest gibt die Höhe in Punkten, mit der ein Bereich vom eigenen
// oberen Rand bis an den unteren Rand des Fensters reicht.
export function useFensterrest(
  knoten: RefObject<HTMLElement | null>,
  { unten, mindestens }: Fensterrand,
): number {
  const abonnieren = useCallback((melden: () => void) => {
    const fenster = globalThis.window

    fenster.addEventListener('resize', melden)

    // **Das Dokument, nicht der Bereich**: was darüber erscheint, ändert
    // seine Größe und nicht die des Fensters.
    //
    // **Und gemeldet wird im nächsten Bild, nicht im Rückruf** (QA01,
    // #2382):
    // `useSyncExternalStore` zeichnet synchron neu, die neue Höhe ändert das
    // Dokument, dessen Größe gerade gemeldet wurde — und WebKit warf beim
    // Start jeder Seite »ResizeObserver loop completed with undelivered
    // notifications«. Gemessen am 2026-10-20 in der WebKit-Probe: `body`
    // meldete 959 statt 900 Punkte, und im selben Bild zogen sich die
    // Spalten des Arbeitsplatzes um dieselben 59 zusammen.
    let bild = 0
    const beobachter = new ResizeObserver(() => {
      fenster.cancelAnimationFrame(bild)
      bild = fenster.requestAnimationFrame(melden)
    })
    beobachter.observe(fenster.document.body)

    return () => {
      fenster.removeEventListener('resize', melden)
      beobachter.disconnect()
      fenster.cancelAnimationFrame(bild)
    }
  }, [])

  const lesen = () => {
    const element = knoten.current
    if (element === null) return mindestens

    const fenster = globalThis.window
    const oben = element.getBoundingClientRect().top + fenster.scrollY

    return Math.max(mindestens, Math.floor(fenster.innerHeight - oben - unten))
  }

  return useSyncExternalStore(abonnieren, lesen)
}
