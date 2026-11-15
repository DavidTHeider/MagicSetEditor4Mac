// Der Farbwähler einer Stiloption
// (#634).
//
// # Warum er fast immer dasteht
//
// **`allow custom` ist ohne Angabe wahr** (`ColorField::ColorField`,
// `src/data/field/color.cpp`, Zeilen 15–18, Fassung f9d9356 vom 2021-07-06),
// und über das ganze Pack ist es nirgends abgeschaltet: alle 105
// Farbstiloptionen lassen eine eigene Farbe zu. Für die 16 ohne `choice:` ist
// er das Einzige, was es gibt.
//
// # Warum ein `input type="color"` und keine eigene Aufklappliste
//
// **Er ist der Farbwähler des Systems.** Auf macOS öffnet er die Farbpalette,
// die jede andere Anwendung auch benutzt — samt Pipette, Schiebereglern und
// den gesicherten Farben. Ein nachgebauter wäre weniger und müsste jede
// Tastaturbedienung selbst mitbringen (§6).
import { alsBrowserfarbe, alsMsefarbe } from '../farben'
import { useTexte } from '../sprache/kontext'
import { type StiloptionDTO } from '../api'

interface FarbwahlEigenschaften {
  option: StiloptionDTO
  laeuft: boolean

  // beschriebenVon nennt den Grund, wenn die Wahl für einen Leser gesperrt
  // ist (H01, #2255).
  beschriebenVon?: string | undefined
  setzen: (wert: string) => void
}

export function Farbwahl({
  option,
  laeuft,
  beschriebenVon,
  setzen,
}: FarbwahlEigenschaften) {
  const texte = useTexte()

  const farbe = alsBrowserfarbe(option.wert)

  return (
    <input
      type="color"
      // **Ohne lesbaren Wert steht Schwarz da, und der Wert bleibt, was er
      // war.** Der Farbwähler kann nichts anderes zeigen; die Angabe aus der
      // Vorlage geht dadurch nicht verloren — sie wird erst überschrieben,
      // wenn jemand wählt.
      value={farbe ?? '#000000'}
      disabled={laeuft}
      aria-describedby={beschriebenVon}
      aria-label={`${option.name} — ${texte.seiten.stil.eigeneFarbe}`}
      onChange={(ereignis) => {
        setzen(alsMsefarbe(ereignis.target.value))
      }}
      className="h-8 w-16 self-start rounded-sm border border-kante bg-transparent focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
    />
  )
}
