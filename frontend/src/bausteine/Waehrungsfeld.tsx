// **Ein Feld für einen Währungscode** — UX-06.AK-01 bis AK-03
// (#1678).
//
// # Was es tut
//
//   - **Es schlägt vor**, was die Plattform an Währungen kennt, mit ihrem
//     Namen in der Sprache des Fensters — und nimmt trotzdem jeden Code an,
//     der die Gestalt hat (AK-01: offene Werte behalten Freitext).
//   - **Es urteilt wie Go** ([waehrungsbefund]), sobald das Feld verlassen
//     ist, und der Satz sagt, was zu tun ist (AK-03). Der getippte Text bleibt
//     stehen — nichts wird still großgeschrieben, wo die Stelle dahinter es
//     nicht auch tut.
//   - **Ein Befund der Anwendung schlägt das eigene Urteil**, wie bei
//     `Pflichtfeld`: „zu einem Preis gehört seine Währung" weiß nur die
//     Domäne.
import { useMemo, useState } from 'react'

import { Vorschlagsfeld } from './Vorschlagsfeld'

import { einsetzen } from '../sprache/einsetzen'
import { useSprache, useTexte } from '../sprache/kontext'
import { waehrungsbefund, waehrungsvorschlaege } from '../waehrungsgestalt'

interface WaehrungsfeldEigenschaften {
  beschriftung: string
  wert: string
  setzeWert: (wert: string) => void

  // kleinErlaubt folgt der Stelle, an die der Wert geht (siehe
  // [waehrungsbefund]).
  kleinErlaubt: boolean

  befund?: string | undefined
}

export function Waehrungsfeld({
  beschriftung,
  wert,
  setzeWert,
  kleinErlaubt,
  befund,
}: WaehrungsfeldEigenschaften) {
  const texte = useTexte().bausteine.waehrung
  // **Die Namen in der Sprache des Fensters** — sie sind Wörter; wie eine
  // Zahl aussieht, entscheidet die Darstellung, wie ein Wort heißt, die
  // Sprache (LOC-03.AK-03).
  const sprache = useSprache()
  const vorschlaege = useMemo(() => waehrungsvorschlaege(sprache), [sprache])

  const [verlassen, setzeVerlassen] = useState(false)

  const eigener = verlassen ? waehrungsbefund(wert, kleinErlaubt) : undefined
  const eigenerSatz =
    eigener === undefined
      ? undefined
      : eigener.art === 'gestalt'
        ? einsetzen(texte.keineGestalt, { wert: wert.trim() })
        : einsetzen(texte.grossSchreiben, {
            wert: wert.trim(),
            gross: eigener.gross,
          })

  return (
    <Vorschlagsfeld
      beschriftung={beschriftung}
      wert={wert}
      setzeWert={(neu) => {
        setzeVerlassen(false)
        setzeWert(neu)
      }}
      vorschlaege={vorschlaege}
      beispiel={texte.beispiel}
      befund={befund ?? eigenerSatz}
      verlassen={() => {
        setzeVerlassen(true)
      }}
      breite={5}
    />
  )
}
