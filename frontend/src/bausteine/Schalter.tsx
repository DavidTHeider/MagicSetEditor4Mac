// **Der Schalter einer Einstellung** — UXR-F2
// (#2008),
// herausgelöst aus dem `Bereichsschalter` (UXR-19,
// #1927).
//
// # Warum ein Baustein
//
// **Weil dasselbe Bedienelement sonst an zwei Stellen stünde**: die
// Bereichsschalter und jede Einstellung mit zwei Werten, die an oder aus
// sind. Ein Kästchen mit `role="switch"`, damit ein Vorleseprogramm „Schalter,
// ein" sagt statt „Kontrollkästchen, angehakt" — und die Leertaste legt um,
// wie bei jedem Schalter.
//
// **Die Beschriftung bringt der Aufrufer**: er stellt den Schalter in ein
// `<label>` mit dem Namen. Der Baustein weiß nichts von Namen und nichts vom
// Wert, den „an" in der Einstellung hat.
//
// # Der Fokus bleibt
//
// **Ein Schalter ist während des Schreibens `disabled`** (§7, Regel 4), und
// ein fokussiertes Element, das `disabled` wird, verliert den Fokus an den
// Seitenanfang (#1990). Wer mit der Leertaste umlegt, stünde danach am Anfang
// der Seite; der Fokus geht deshalb zurück, wenn er beim Umlegen auf dem
// Schalter war.
import { useRef } from 'react'

import { useFokusUeberAnhalten } from '../haken/useFokusUeberAnhalten'

export function Schalter({
  schluessel,
  an,
  angehalten,
  aendern,
}: {
  // schluessel steht als `data-einstellung` am Schalter, damit sich zählen
  // lässt, dass keine Einstellung zweimal dasteht.
  schluessel: string
  an: boolean
  angehalten: boolean
  aendern: (an: boolean) => void
}) {
  const griff = useRef<HTMLInputElement | null>(null)
  const merkeFokus = useFokusUeberAnhalten(griff, angehalten)

  return (
    <input
      ref={griff}
      type="checkbox"
      role="switch"
      data-einstellung={schluessel}
      checked={an}
      disabled={angehalten}
      onChange={(ereignis) => {
        merkeFokus()
        aendern(ereignis.target.checked)
      }}
      className="size-4 accent-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
    />
  )
}
