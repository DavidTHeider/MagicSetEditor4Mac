// **Segmente für eine Einstellung mit zwei bis vier Werten** — UXR-F2
// (#2008),
// Regel R7 aus `docs/UIREVIEW.md`: *Segmente bis vier Werte, Auswahl darüber.*
//
// # Optionsfelder und keine Knöpfe
//
// **Ein gedrückter `Knopf` trug damals das Gewicht des Hauptknopfs**
// (`gedrueckt` in `Knopf`), und jede Region hat höchstens einen (R1):
// Einstellungen mit einer Segmentreihe je Zeile hätten dutzendfach „die
// führende Handlung". Seit AP-3 (#2360) zeichnet er den gewählten Stand wie
// ein Segment hier — Rand in der Marke, Grund `betont` —, und die Füllung
// gehört der einen führenden Handlung je Blick.
// Dazu ist eine Wahl unter mehreren, was ein Optionsfeld ist — mit **einer**
// Tabstelle für die ganze Reihe und den Pfeiltasten darin, ohne dass dafür
// Quelltext nötig wäre. Gezeichnet wird das Optionsfeld selbst nicht
// (`sr-only`), sondern sein Beschriftungsfeld: ein Segment, dessen gewählter
// Stand an Rand und Grund zu sehen ist.
//
// **Der Name der Gruppe kommt von außen** (`beschriftetVon`): die Zeile der
// Einstellung hat ihn schon, mit Marke für einen Treffer der Suche; ein
// zweiter stünde neben dem ersten.
//
// # Der Fokus bleibt in der Reihe
//
// **Gewählt wird mit jedem Pfeil, und jedes Wählen schreibt** — für die Zeit
// des Schreibens sind die Felder `disabled` (§7, Regel 4), und ein
// fokussiertes Element, das `disabled` wird, verliert den Fokus an den
// Seitenanfang (#1990). Ohne Rückgabe wählte der erste Pfeil, und der zweite
// ginge ins Leere. Der Fokus geht deshalb an das Feld zurück, das ihn beim
// Wählen hatte.
import { useEffect, useId, useRef } from 'react'

import { useFokusUeberAnhalten } from '../haken/useFokusUeberAnhalten'

// Segment ist ein Wert samt dem Namen, unter dem er dasteht.
export interface Segment {
  wert: string
  name: string
}

// **Der gewählte Stand an Rand und Grund und nicht an der Schrift**: ein
// Wechsel des Schriftschnitts verschöbe beim Wählen alle Nachbarn. `betont`
// ist der Grund, den die Oberfläche für eine hervorgehobene Zeile nimmt;
// `text` darauf ist gemessen (`docs/KONTRASTE.md`).
const segment =
  'block rounded-sm border border-transparent px-3 py-1 text-center text-sm ' +
  'text-text peer-checked:border-marke peer-checked:bg-betont ' +
  'peer-aria-disabled:opacity-60'

// **Der Ring liegt um das Segment und nicht am Optionsfeld**, das man nicht
// sieht — und er hängt an `:focus`, nicht an `:focus-visible`: gemessen am
// Fenster (`uxrf2`, #2008) hat ein per Pfeiltaste fokussiertes Optionsfeld in
// WKWebView kein `:focus-visible`, und der Ring blieb aus. Ein Mausklick
// fokussiert dort kein Optionsfeld. `fokusring_test.go` kennt die Form als
// Ring am Rahmen.
const beschriftungsfeld =
  'relative flex min-w-0 rounded-sm ' +
  'focus-within:outline focus-within:outline-2 ' +
  'focus-within:outline-offset-2 focus-within:outline-ring'

export function Segmente({
  beschriftetVon,
  schluessel,
  optionen,
  wert,
  angehalten,
  aendern,
  fokusBleibt = false,
}: {
  // beschriftetVon ist die Kennung des Elements, das die Reihe benennt.
  beschriftetVon: string

  // schluessel steht als `data-einstellung` an der Reihe — einmal je
  // Einstellung, nicht je Feld. Ohne ihn ist die Reihe keine Einstellung
  // der Einstellungsseite, sondern eine Wahl an Ort und Stelle (AP-6,
  // #2363: die Ansicht einer Liste).
  schluessel?: string | undefined
  optionen: readonly Segment[]
  wert: string
  angehalten: boolean
  aendern: (wert: string) => void

  // fokusBleibt hält den Fokus beim Mausdruck, wo er war — für eine Wahl,
  // nach der jemand anderes ihn zurückholt (QA01, #2382: in WebKit nahm der
  // Druck auf die Beschriftung ihn sonst aus der Liste, und er landete auf
  // `main`). Knöpfe nehmen ihn in WebKit ohnehin nicht.
  fokusBleibt?: boolean
}) {
  // name bindet die Optionsfelder zu einer Gruppe: erst damit sind die Pfeile
  // eine Wanderung innerhalb der Reihe und die Reihe eine Tabstelle.
  const name = useId()
  const huelle = useRef<HTMLDivElement | null>(null)
  const merkeFokus = useFokusUeberAnhalten(huelle, angehalten)

  // **Ein Horcher und kein `onMouseDown`**: die Reihe ist keine Tabstelle,
  // ihre Optionsfelder sind es — und `jsx-a11y` verlangt von einem Element
  // mit Handler, dass es selbst eine ist.
  useEffect(() => {
    if (!fokusBleibt) {
      return undefined
    }

    const element = huelle.current
    /* v8 ignore next 3 -- die Reihe steht, wenn der Effekt läuft. */
    if (element === null) {
      return undefined
    }

    const halten = (ereignis: MouseEvent) => {
      ereignis.preventDefault()
    }
    element.addEventListener('mousedown', halten)

    return () => {
      element.removeEventListener('mousedown', halten)
    }
  }, [fokusBleibt])

  return (
    <div
      ref={huelle}
      role="radiogroup"
      aria-labelledby={beschriftetVon}
      data-einstellung={schluessel}
      className="inline-flex w-fit max-w-full gap-0.5 self-start rounded-md border border-kante bg-flaeche p-0.5"
    >
      {optionen.map((option) => (
        <label key={option.wert} className={beschriftungsfeld}>
          <input
            type="radio"
            name={name}
            value={option.wert}
            checked={option.wert === wert}
            // **Gesperrt mit `aria-disabled`, nicht `disabled`** (`uxrf2`):
            // so bleibt der Fokus stehen, und der Ring auch.
            aria-disabled={angehalten}
            onChange={(ereignis) => {
              if (angehalten) {
                return
              }

              merkeFokus(ereignis.target)
              aendern(option.wert)
            }}
            className="peer sr-only"
          />
          <span className={segment}>{option.name}</span>
        </label>
      ))}
    </div>
  )
}
