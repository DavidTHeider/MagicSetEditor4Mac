// **Ein Kasten aus einem Menü** — die Gestalt, die ein Menüeintrag öffnet,
// wenn hinter ihm mehr steht als ein Klick: ein Formular, eine Liste, eine
// Auskunft.
//
// # Warum eine eigene Datei
//
// Er stand in `Kartenwerkzeuge.tsx` für das Menü „Mehr" (UXR-05, #1913).
// Seit UXR-06 (#1914)
// öffnet auch das Menü „Diese Karte" einen — den Bezug zu einem Druck —, und
// zwei Fassungen desselben Kastens liefen auseinander, sobald jemand eine
// ändert.
//
// # Was er zusagt
//
// **Mit seinem Titel als Namen und einem Knopf, der ihn schließt.** Wohin
// der Fokus beim Schließen geht, entscheidet der Aufrufer: er weiß, wo das
// Menü steht.
import { useEffect, useRef, type ReactNode } from 'react'

import { Knopf } from '../bausteine/Knopf'
import { useTexte } from '../sprache/kontext'

export function Werkzeugkasten({
  titel,
  offen,
  schliessen,
  children,
}: {
  titel: string
  offen: boolean
  schliessen: () => void
  children: ReactNode
}) {
  const schliessenText = useTexte().bausteine.aktionen.schliessen
  const kasten = useRef<HTMLElement>(null)

  // **Der Fokus geht hinein, wenn er aufgeht.** Das Menü gibt ihn beim Wählen
  // an seinen Knopf zurück; ohne diesen Zug stünde er dort, und der Kasten
  // darunter wäre für eine Vorlesehilfe nicht angesagt.
  useEffect(() => {
    if (offen) {
      kasten.current?.focus()
    }
  }, [offen])

  return (
    <section
      ref={kasten}
      aria-label={titel}
      tabIndex={-1}
      hidden={!offen}
      className="flex flex-col gap-3 rounded-sm border border-kante p-3 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
    >
      <div className="flex items-center justify-between gap-2">
        <p className="text-sm font-medium">{titel}</p>

        <Knopf gewicht="leise" klein aufKlick={schliessen}>
          {schliessenText}
        </Knopf>
      </div>

      {children}
    </section>
  )
}
