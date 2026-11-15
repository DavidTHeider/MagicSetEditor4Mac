// **Ein Textfeld für eine ganze Zahl ab null, das seinen Fehler am Feld sagt**
// — F03 (#2245).
//
// # Warum nicht `Zahlenfeld`
//
// `Zahlenfeld` meldet nur gültige Werte nach oben und hält den getippten Text
// selbst — richtig für eine Einstellung, die beim Tippen gilt. Hier liegt der
// Text beim Aufrufer, weil er erst beim **Absenden** gelesen wird: eine
// Kennung, eine Stückzahl, ein Spielstand. Bis dahin las `parseInt` ihn und
// machte aus „12a" ein anderes Exemplar.
//
// # Was es tut
//
// **Es urteilt erst, wenn der Aufrufer `pruefen` setzt** — wer tippt, ist noch
// nicht fertig. Danach steht der Satz am Feld, bis der Text eine Zahl ist.
// **`leerIstNull`** lässt ein leeres Feld als null gelten, wo das die Bedeutung
// ist (ein Spielstand ohne Siege).
import { Feld } from './Feld'

import { ganzzahlAus } from '../ganzzahl'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

interface GanzzahlfeldEigenschaften {
  beschriftung: string
  wert: string
  setzeWert: (wert: string) => void

  // pruefen sagt, dass der Aufrufer abgesendet hat und das Urteil zeigen will.
  pruefen: boolean

  leerIstNull?: boolean

  // kleinstens ist die Untergrenze, die das Feld selbst sagt — 0, wenn
  // niemand etwas sagt. **Was darüber hinaus gilt, entscheidet die Domäne.**
  kleinstens?: number
}

// feldklasse ist die eines beschrifteten Feldes, und ein abgewiesener Wert
// bekommt den Rand, den `Zahlenfeld` ihm gibt. **Als ein Literal**: der
// Wächter über die Fokusringe liest Klassen aus dem Quelltext, und eine
// zusammengesetzte Zeichenkette löst er nicht auf.
const feldklasse =
  'rounded-sm border border-kante bg-flaeche px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-[invalid=true]:border-fehler'

// ganzzahlImFeld liest den Text, wie das Feld ihn beurteilt — eine Zahl
// unter `kleinstens` ist keine, die gilt.
export function ganzzahlImFeld(
  wert: string,
  leerIstNull: boolean,
  kleinstens = 0,
): number | undefined {
  const zahl = leerIstNull && wert.trim() === '' ? 0 : ganzzahlAus(wert)

  return zahl !== undefined && zahl >= kleinstens ? zahl : undefined
}

export function Ganzzahlfeld({
  beschriftung,
  wert,
  setzeWert,
  pruefen,
  leerIstNull = false,
  kleinstens = 0,
}: GanzzahlfeldEigenschaften) {
  const allgemein = useTexte().allgemein

  const falsch =
    pruefen && ganzzahlImFeld(wert, leerIstNull, kleinstens) === undefined

  return (
    <Feld
      beschriftung={beschriftung}
      {...(falsch
        ? {
            befund: einsetzen(allgemein.ganzeZahlAb, {
              kleinstens: String(kleinstens),
            }),
          }
        : {})}
    >
      {(verbindung) => (
        <input
          {...verbindung}
          type="text"
          inputMode="numeric"
          size={8}
          value={wert}
          className={feldklasse}
          onChange={(ereignis) => {
            setzeWert(ereignis.target.value)
          }}
        />
      )}
    </Feld>
  )
}
