// Die Kästchen einer Mehrfachwahl
// (#637).
//
// # Warum Kästchen und keine Mehrfachliste
//
// **Ein `<select multiple>` verlangt gedrückte Zusatztasten**, um mehr als
// eines zu wählen, und zeigt nie alle Möglichkeiten auf einmal. MSE zeichnet
// eine Reihe von Kästchen (`MultipleChoiceValueEditor`), und das ist auch hier
// die Bedienung, die ohne Vorwissen auskommt (§6).
//
// # Was hier nicht entschieden wird
//
// **Was aus einer leeren Auswahl wird.** Die Vorlage darf über `empty choice:`
// sagen, was dann im Wert steht; das weiß nur der Dienst. Diese Stelle schickt
// die leere Zeichenkette und lässt ihn antworten.
import { type StiloptionDTO } from '../api'
import { Wahltrenner, wahlenAus } from '../mehrfachwahl'

interface MehrfachwahlEigenschaften {
  option: StiloptionDTO
  laeuft: boolean

  // beschriebenVon nennt den Grund, wenn die Wahl für einen Leser gesperrt
  // ist — **an jedem Kästchen** (H01, #2255).
  beschriebenVon?: string | undefined
  setzen: (wert: string) => void
}

export function Mehrfachwahl({
  option,
  laeuft,
  beschriebenVon,
  setzen,
}: MehrfachwahlEigenschaften) {
  const gewaehlt = wahlenAus(option.wert)

  return (
    <div className="flex flex-wrap gap-x-4 gap-y-1">
      {option.moeglichkeiten.map((moeglich) => (
        <label key={moeglich.wert} className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={gewaehlt.includes(moeglich.wert)}
            disabled={laeuft}
            aria-describedby={beschriebenVon}
            onChange={(ereignis) => {
              // **Geschickt wird die Reihenfolge der Erklärung.** Der Dienst
              // ordnet sie ohnehin in die Normalform; eine Seite, die etwas
              // anderes schickt, als sie zeigt, ist schwerer zu lesen.
              const naechste = option.moeglichkeiten
                .map((eine) => eine.wert)
                .filter((wert) =>
                  wert === moeglich.wert
                    ? ereignis.target.checked
                    : gewaehlt.includes(wert),
                )

              setzen(naechste.join(Wahltrenner))
            }}
            className="size-4 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
          />

          {moeglich.beschriftung}
        </label>
      ))}
    </div>
  )
}
