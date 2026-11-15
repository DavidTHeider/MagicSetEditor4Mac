// Das Meldungsband: die eine Stelle für eine Meldung, die zu keiner Seite gehört.
//
// # Warum es das gibt
//
// ⌘Z und ⌘S gelten im **ganzen Fenster** — sie hängen am Set, nicht an der
// Seite. Ihre Meldung stand bis hierher auf der Kartenseite beziehungsweise
// nirgends: wer auf Statistik, Suche oder Druckbogen ⌘Z drückte und das
// Schreiben scheiterte, bekam **nichts** zu sehen
// (#180).
//
// # Warum das kein „anonymer Hinweis" nach §6 ist
//
// §6 lehnt ihn ab, und zu Recht: ein Feldfehler gehört ans Feld, sonst sucht
// jemand, welche Eingabe gemeint war. Diese Meldungen haben aber **kein Feld**
// und keine Seite — sie gehören zu einer Handlung des Fensters. Sie irgendwo
// hinzuhängen wäre die Erfindung eines Bezugs, den es nicht gibt.
//
// Die drei verworfenen Möglichkeiten stehen im Issue: ein Hinweis ohne Ort
// (genau der anonyme), die Meldung in jede Seite durchreichen (fünf Seiten, die
// von einer Rücknahme nichts wissen wollen), oder ⌘Z auf die Kartenseite
// beschränken (eine gewohnte Taste, die je nach Seite etwas anderes tut — §6
// lehnt auch das ab).
//
// # Warum es `role="status"` trägt und nicht `role="alert"`
//
// `alert` unterbricht einen Vorleser mitten im Satz. Das ist für eine
// gescheiterte Rücknahme zu viel: sie ist ärgerlich, nicht dringend. `status`
// meldet höflich, sobald der Vorleser Luft holt.
import { useTexte } from '../sprache/kontext'

export type Meldungsart = 'gut' | 'schlecht'

// **Die Farben stehen als Tabelle und nicht als Ausdruck im JSX.**
// `keintextimjsx.test.ts` weist eine Zeichenkette mit Platzhalter im JSX ab —
// zu Recht: eine zusammengesetzte Zeichenkette dort ist meistens ein
// Oberflächentext, den jemand vergessen hat. Hier ist sie es nicht, und die
// Tabelle sagt das, statt eine Ausnahme zu verlangen.
const bandfarben: Record<Meldungsart, string> = {
  schlecht: 'border-fehler text-fehler',
  gut: 'border-erfolg text-erfolg',
}

const bandrahmen =
  'mt-2 flex items-start justify-between gap-4 rounded-sm border px-3 py-2 text-sm'

export interface Bandmeldung {
  text: string
  art: Meldungsart
}

interface MeldungsbandEigenschaften {
  meldung: Bandmeldung | undefined
  schliessen: () => void
}

export function Meldungsband({
  meldung,
  schliessen,
}: MeldungsbandEigenschaften) {
  const texte = useTexte()

  // **Der Bereich steht immer da, auch leer.** Ein Vorleser meldet nur, was
  // sich in einem Bereich ändert, den er schon kennt; einer, der erst mit der
  // Meldung entsteht, bleibt stumm.
  return (
    <div
      role="status"
      aria-live="polite"
      aria-label={texte.allgemein.meldungsband.bereich}
      className="px-6"
    >
      {meldung ? (
        <div className={[bandrahmen, bandfarben[meldung.art]].join(' ')}>
          <span>{meldung.text}</span>
          <button
            type="button"
            onClick={schliessen}
            aria-label={texte.bausteine.aktionen.schliessen}
            className="rounded-sm px-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            ×
          </button>
        </div>
      ) : null}
    </div>
  )
}
