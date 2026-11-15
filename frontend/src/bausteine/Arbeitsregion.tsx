// **Eine Spalte des Arbeitsplatzes** — Navigator oder Inspector.
//
// **Mit einer Überschrift, die nur vorgelesen wird** (#2249): die Abschnitte
// darin tragen h3 — „Marken", „Kartendaten von Scryfall", „Gespeicherte
// Filter" —, und ohne eine h2 darüber sprang die Gliederung von der h1 der
// Seite auf die dritte Ebene. Gesehen hat es der Wächter über die
// Barrierefreiheit erst, als er auf die geladene Seite wartete; davor prüfte
// er den Ladezustand, in dem es die Abschnitte noch nicht gab. Ein
// Vorleseprogramm springt über Überschriften von Region zu Region, und die
// Spalte hat jetzt eine.
import type { CSSProperties, ReactNode } from 'react'

export function Arbeitsregion({
  id,
  bezeichnung,
  inhalt,
  className,
  style,
  ueber = false,
  spalte,
}: {
  id: string
  bezeichnung: string
  inhalt: ReactNode
  className: string
  style?: CSSProperties | undefined

  // ueber sagt, dass die Spalte über dem Zentrum liegt (enges Fenster).
  ueber?: boolean

  // spalte bekommt das Element — für den Fokus beim Öffnen über dem Zentrum.
  spalte?: (element: HTMLElement | null) => void
}) {
  // **`aria-label` bleibt**, auch neben der Überschrift: die Drehbücher der
  // WebKit-Probe finden die Spalte darüber (`ux15-seitenlaengen.probe`), und
  // der Name ist derselbe.
  return (
    <aside
      id={id}
      aria-label={bezeichnung}
      data-ueber={ueber ? 'zentrum' : undefined}
      ref={spalte}
      className={className}
      style={style}
    >
      <h2 className="sr-only">{bezeichnung}</h2>
      {inhalt}
    </aside>
  )
}
