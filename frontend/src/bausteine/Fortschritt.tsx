// **Ein Lauf, den man hören kann** — Werkstück U3 von Phase 5.1
// (#1414).
//
// # Was vorher war
//
// Die Bildausgabe zeigte ihren Stand als `role="status"` und daneben einen
// Balken mit `role="presentation"`. Der Satz wurde damit angesagt, der Balken
// war für ein Vorleseprogramm **nicht da** — und wer die Stelle im Lauf
// abfragen wollte, statt auf die nächste Ansage zu warten, hatte nichts zu
// fragen. Ein `progressbar` mit `aria-valuenow` beantwortet genau das.
//
// # Warum der Satz vom Aufrufer kommt
//
// Weil nur er ihn kennt: „3 geschrieben von 5", „Angehalten. 3 geschrieben
// von 5", „5 Bilder geschrieben nach /tmp/bilder". Der Ordner steht in der
// dritten Fassung, in den anderen beiden nicht, und ein Baustein, der die
// Fälle selbst zusammensetzte, bräuchte für jede Seite eine Ausnahme. Was
// hier steht, ist die **Gestalt** — Live-Region, Balken, Anhalteknopf.
//
// # Und warum der Balken denselben Satz als Namen trägt
//
// Ein `progressbar` ohne Namen ist ein Befund (WCAG 4.1.2), und ein zweiter
// Name daneben wäre eine zweite Aussage über dieselbe Sache. Er zeigt deshalb
// mit `aria-labelledby` auf den Satz über sich — genau die Zeile, die sagt,
// was hier vorangeht.
import { useId } from 'react'

interface FortschrittEigenschaften {
  satz: string

  erledigt: number
  gesamt: number

  // unbestimmt heißt: wie viel es wird, weiß niemand — eine Exportvorlage
  // schreibt, so viele Bilder ihr Skript will (FMT-03). Der Balken nennt dann
  // keinen Stand, statt einen vorzutäuschen (`aria-valuenow` fehlt, wie ARIA
  // es für einen unbestimmten Fortschritt verlangt).
  unbestimmt?: boolean

  // anhalten steht nur da, wenn es etwas anzuhalten gibt. Ein Knopf ohne
  // Rückruf wäre eine Zusage ohne Rumpf (§10).
  anhalten?: {
    beschriftung: string
    angehalten?: boolean
    aufKlick: () => void
  }
}

const bahn = 'mt-2 h-2 w-full max-w-md overflow-hidden rounded-sm bg-betont'

const knopf =
  'rounded-sm border border-kante px-3 py-1 focus-visible:outline ' +
  'focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60'

export function Fortschritt({
  satz,
  erledigt,
  gesamt,
  unbestimmt = false,
  anhalten,
}: FortschrittEigenschaften) {
  const satzID = useId()

  // **Ein Lauf ohne Stücke ist fertig, nicht kaputt.** Der Balken steht dann
  // voll, und `aria-valuemax` bleibt über dem Minimum: `valuemax === valuemin`
  // ergäbe einen Balken, dessen Stand sich nicht ausdrücken lässt.
  const spanne = Math.max(gesamt, 1)
  const anteil = gesamt === 0 ? 100 : Math.round((erledigt / gesamt) * 100)

  return (
    <div className="flex flex-col gap-2">
      <p id={satzID} role="status" aria-live="polite" className="text-leise">
        {satz}
      </p>

      {unbestimmt ? (
        <div role="progressbar" aria-labelledby={satzID} className={bahn}>
          <div className="h-full w-full bg-marke opacity-60" />
        </div>
      ) : (
        <div
          role="progressbar"
          aria-labelledby={satzID}
          aria-valuenow={erledigt}
          aria-valuemin={0}
          aria-valuemax={spanne}
          className={bahn}
        >
          <div
            className="h-full bg-marke"
            style={{ width: `${String(anteil)}%` }}
          />
        </div>
      )}

      {anhalten !== undefined && (
        <div>
          <button
            type="button"
            disabled={anhalten.angehalten}
            onClick={anhalten.aufKlick}
            className={knopf}
          >
            {anhalten.beschriftung}
          </button>
        </div>
      )}
    </div>
  )
}
