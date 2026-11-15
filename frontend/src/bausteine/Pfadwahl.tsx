// **Ein Ort, den ein Dateidialog wählt — und der lesbar daneben steht.**
//
// # Wozu
//
// UXR-F15 (#2030):
// die Leihe nahm *Ziel* und *Umschlag* als Pfad in einem Freitextfeld. Ein
// Pfad von Hand ist auf macOS kein Bedienelement (R7 in `docs/UIREVIEW.md`);
// überall sonst wählt ein Dateidialog.
//
// # Was er zeigt
//
//   - **Eine Beschriftung, einen Knopf, den Pfad.** Der Knopf öffnet den
//     Dialog und heißt, solange nichts gewählt ist, *wählen* und danach
//     *anderen wählen* — zwei Aufschriften, damit ein Vorleseprogramm zwei
//     Wahlen auf einer Seite auseinanderhält.
//   - **Der Pfad steht im Ganzen da**, in fester Schrift und umbrechend: ein
//     Pfad, der nach dem dritten Ordner abgeschnitten wäre, ließe sich nicht
//     mehr prüfen. Der Knopf nennt ihn als seine Beschreibung.
//   - **Ist nichts gewählt, steht ein Satz da** und kein leerer Fleck.
//   - **Der Grund eines gescheiterten Dialogs steht am Feld** (§6).
//
// **Er tippt nichts.** Es gibt kein Eingabefeld, das ein Pfad sein könnte; wer
// einen anderen Ort will, öffnet den Dialog noch einmal.
import { useId } from 'react'

import { Knopf } from './Knopf'

interface PfadwahlEigenschaften {
  // bezeichnung nennt, was gewählt wird — der Name der Gruppe.
  bezeichnung: string

  // pfad ist der gewählte Ort; leer heißt: keiner.
  pfad: string

  // kein ist der Satz, der statt des Pfads steht.
  kein: string

  // waehlen und anderes sind die zwei Aufschriften des Knopfs.
  waehlen: string
  anderes: string

  // fehler ist der Grund, aus dem der Dialog scheiterte.
  fehler: string | undefined

  // angehalten heißt: ein Dialog ist offen (§7, Regel 4).
  angehalten: boolean

  aufWahl: () => void
}

export function Pfadwahl({
  bezeichnung,
  pfad,
  kein,
  waehlen,
  anderes,
  fehler,
  angehalten,
  aufWahl,
}: PfadwahlEigenschaften) {
  const kopf = useId()
  const ort = useId()

  return (
    <div role="group" aria-labelledby={kopf} className="flex flex-col gap-1">
      <span id={kopf} className="text-sm font-medium">
        {bezeichnung}
      </span>

      <div className="flex flex-wrap items-center gap-3">
        <Knopf
          gewicht="leise"
          angehalten={angehalten}
          beschriebenVon={ort}
          aufKlick={aufWahl}
        >
          {pfad === '' ? waehlen : anderes}
        </Knopf>

        <span
          id={ort}
          className={
            pfad === ''
              ? 'min-w-0 text-sm text-leise'
              : 'min-w-0 font-mono text-xs break-all'
          }
        >
          {pfad === '' ? kein : pfad}
        </span>
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}
    </div>
  )
}
