// **Die Fußleiste der Ausgabe** — AP-10
// (#2367),
// Befund B-20: bei 1.440 × 838 lag der Knopf, der schreibt, unter der Falz.
//
// **Sie bleibt beim Rollen am unteren Rand stehen**, solange die Fläche des
// Ziels weiter reicht als das Fenster — `sticky` an ihrem eigenen Platz, kein
// Element, das über allem schwebt: rollt man bis zu ihr, steht sie, wo sie
// steht, und was darunter kommt — Lauf, Meldung, Ergebnis —, folgt ihr.
//
// **Fest nur im Assistenten.** Die Exportvorlage steht auch an einer
// Setkarte in einem Kasten; dort bliebe eine Leiste über der Liste der Sets
// stehen. `data-ausgabefuss` sagt, welche Gestalt gilt — die Läufe am
// Fenster und die Tests lesen es.
import { useEffect, type ReactNode } from 'react'

// feste zählt die festen Leisten, deren Effekt läuft — **nur die sichtbaren**:
// `Activity` räumt die Effekte einer verborgenen Seite ab.
let feste = 0

// useMarke setzt `data-ausgabeleiste` an `html`, solange eine feste Leiste zu
// sehen ist — daran hängt der Rand, mit dem der Fokus nicht hinter ihr
// verschwindet (`barrierefreiheit.css`, WCAG 2.4.11). **Nicht über
// `html:has(…)`**: das traf auch die Leiste einer verborgenen Seite und
// verschob das Rollen überall sonst. **Ein eigener Name und nicht
// `data-ausgabefuss`**: an `html` träfe `[data-ausgabefuss=fest] button` sonst
// jeden Knopf der Seite.
function useMarke(fest: boolean) {
  useEffect(() => {
    if (!fest) return undefined

    feste += 1
    document.documentElement.dataset.ausgabeleiste = 'fest'

    return () => {
      feste -= 1
      if (feste === 0) delete document.documentElement.dataset.ausgabeleiste
    }
  }, [fest])
}

export function Ausgabefuss({
  fest = true,
  children,
}: {
  fest?: boolean
  children: ReactNode
}) {
  useMarke(fest)

  return (
    <div
      data-ausgabefuss={fest ? 'fest' : 'frei'}
      className={
        fest
          ? 'sticky bottom-0 z-10 -mx-5 mt-4 flex flex-wrap items-center gap-x-4 gap-y-2 border-t border-kante bg-flaeche px-5 py-3'
          : 'mt-3 flex flex-wrap items-center gap-2'
      }
    >
      {children}
    </div>
  )
}
