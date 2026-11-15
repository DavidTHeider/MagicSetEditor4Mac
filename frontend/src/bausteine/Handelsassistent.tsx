// **Ein Handel in vier Phasen** — TRADE-02
// (#1716),
// nach dem Artboard *Handeln*: „Mit wem", „Was tauscht ihr", „Zusagen",
// „Verbuchen".
//
// # Was er ist und was nicht
//
// **Eine Anordnung und kein neues Zustandsmodell.** Welche Phase gilt, sagt
// der Vorgang selbst (`phase` aus `domain`); der Assistent zeigt die Tafel
// dazu und darüber, was jetzt geht. Die dreizehn Zustände bleiben, wie sie
// sind — sie erscheinen als Fortschritt und nicht als Formular (AK-01).
//
// # Warum die Tafeln eingehängt bleiben
//
// **Weil Zurück den Entwurf nicht kosten darf** (AK-03). Wer in „Mit wem"
// nachsieht und zurückkommt, findet die halb zusammengestellten Positionen
// wieder; eine ausgehängte Tafel hätte sie verloren. Verborgen wird über
// `hidden` — nicht sichtbar, nicht im Tabweg, nicht im Zugänglichkeitsbaum.
import { useState, type ReactNode } from 'react'

import { useTexte } from '../sprache/kontext'

// schrittstil ist der eines gewählten Segments (`Segmente`, `Knopf`): die
// gewählte Phase mit Rand in der Marke und dem Grund `betont`, ohne das
// Hauptgewicht. **Bis AP-3 (#2360) der einer Lasche** im Kastenstil; seit
// jede Lasche eine Unterlinie trägt, ist eine gedrückte Phase ein Segment —
// sie ist ein Umschalter (`aria-pressed`) und keine Lasche. Die Paare sind
// gemessen (`docs/KONTRASTE.md`: Text auf `bg-betont`, `border-marke` gegen
// `bg-flaeche`).
const schrittstil =
  'inline-flex items-center gap-2 rounded-md border border-transparent px-3 py-1.5 hover:bg-betont focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-pressed:border-marke aria-pressed:bg-betont'

// Die vier Phasen, die ein Mensch erlebt — `ende` ist keine davon.
export const dieHandelsphasen = [
  'gegenueber',
  'angebot',
  'zusagen',
  'verbuchen',
] as const

export type Handelsphase = (typeof dieHandelsphasen)[number]

interface Handelsassistenteigenschaften {
  // phase kommt vom Vorgang; `ende` zeigt die Tafel des Angebots.
  phase: string

  // vorgang ist der gewählte Handel — wer einen anderen wählt, sieht dessen
  // Phase, auch wenn sie dieselbe Nummer trägt.
  vorgang: string

  // lage ist der fertige Satz — „Du bist dran: zustimmen", „Wartet auf …".
  lage: string | undefined

  // schritte sind die Knöpfe, die jetzt gehen.
  schritte: ReactNode

  tafeln: Record<Handelsphase, ReactNode>
}

export function Handelsassistent({
  phase,
  vorgang,
  lage,
  schritte,
  tafeln,
}: Handelsassistenteigenschaften) {
  const eigene = useTexte().seiten.handeln.assistent

  const aktuell: Handelsphase = dieHandelsphasen.includes(phase as Handelsphase)
    ? (phase as Handelsphase)
    : 'angebot'
  const stelle = dieHandelsphasen.indexOf(aktuell)

  // **Die Ansicht folgt dem Vorgang**, wenn er weiterrückt oder ein anderer
  // gewählt wird — und bleibt, wo sie ist, solange er steht. Gemerkt wird,
  // wem sie zuletzt gefolgt ist; ein Wechsel stellt sie beim Zeichnen um.
  const folge = `${vorgang}/${aktuell}`
  const [ansicht, setzeAnsicht] = useState<Handelsphase>(aktuell)
  const [gefolgt, setzeGefolgt] = useState(folge)

  if (gefolgt !== folge) {
    setzeGefolgt(folge)
    setzeAnsicht(aktuell)
  }

  return (
    <section data-handelsassistent="" className="flex flex-col gap-4">
      <nav aria-label={eigene.schrittleiste}>
        <ol className="flex flex-wrap gap-2">
          {dieHandelsphasen.map((name, index) => (
            <li key={name} aria-current={index === stelle ? 'step' : undefined}>
              {/* **Auch eine kommende Phase lässt sich ansehen** — ihr
                  Journal, was unterwegs ist, was festzuhalten wäre: die
                  Tafel sagt selbst, was dort noch nicht geht. */}
              {/* **Gewählt wie eine Lasche und nicht wie ein Hauptknopf**
                  (UXR-15, #1923; R1 in `docs/UIREVIEW.md`): ein gedrückter
                  `Knopf` ist gefüllt, und neben dem nächsten Schritt
                  standen zwei gefüllte Knöpfe im selben Bereich. */}
              <button
                type="button"
                aria-pressed={ansicht === name}
                onClick={() => {
                  setzeAnsicht(name)
                }}
                className={schrittstil}
              >
                <span aria-hidden="true" className="tabular-nums">
                  {String(index + 1)}
                </span>{' '}
                {eigene.phasen[name]}
                <span className="sr-only">
                  {index < stelle
                    ? eigene.erledigt
                    : index === stelle
                      ? eigene.jetzt
                      : eigene.kommt}
                </span>
              </button>
            </li>
          ))}
        </ol>
      </nav>

      {lage !== undefined && (
        <p role="status" className="text-sm font-medium">
          {lage}
        </p>
      )}

      {schritte}

      {dieHandelsphasen.map((name) => (
        <div
          key={name}
          data-tafel={name}
          hidden={ansicht !== name}
          className="flex flex-col gap-4 [&>*:not(:first-child)]:border-t [&>*:not(:first-child)]:border-kante [&>*:not(:first-child)]:pt-4"
        >
          {tafeln[name]}
        </div>
      ))}
    </section>
  )
}
