// **Der Vorhang am gemeinsamen Gerät** — P7
// (#2375),
// Befund C-05, Artboard `Vorhang09`.
//
// # Wann er zugeht
//
// **Beim Übergang**: entscheidet jetzt ein anderer Sitz als der gezeigte
// ([entscheider]) — nach dem Behalten, nach einem Passen, wenn jemand
// Auslöser ordnen muss —, und sitzen beide an **diesem** Gerät. Bis P5 hieß
// das, den Sitz in der Sichtwahl zu suchen; P6 ließ die Sicht ohne Nachfrage
// folgen und deckte dabei die nächste Hand vor den Augen der vorigen auf.
//
// # Wann nicht
//
// - **Für den Zuschauer nie** — er übernimmt keinen Sitz, auch keinen, der
//   gerade entscheidet.
// - **Nie zu einem Gast und nie aus der Sicht eines Gastes**: der spielt an
//   einem anderen Gerät, und seine Hand gehört nicht auf diesen Bildschirm.
//   Solange unbekannt ist, wer Gast ist, wartet der Übergang.
// - **Nach einer Wahl von Hand nicht** — wer in der Sichtwahl einen Sitz
//   nimmt, hat entschieden, wessen Sicht er sehen will.
// - **Mit „Offenen Händen“ nicht** — dann folgt die Sicht ohne Vorhang, und
//   das ist eine ausdrückliche Wahl, keine Vorgabe.
//
// # Was er nicht ist
//
// **Ein Schutz gegen den, dem das Gerät gehört.** Er ist Bedienung am
// gemeinsamen Bildschirm: die Sichtwahl daneben bleibt, und wer sie nimmt,
// sieht, was er wählt.
import { useCallback, useEffect, useRef, useState } from 'react'

import type { Partie } from '../apipartie'
import { entscheider } from '../spiel/entscheider'

export interface Vorhangstand {
  // vorhang ist der Sitz, an den das Gerät gehen soll — `undefined`: offen.
  vorhang: string | undefined
  bestaetigen: () => void
  bleiben: () => void
}

// Ziel ist der Sitz, an den das Gerät gehen soll, und die Sicht, aus der der
// Übergang kam — gilt eine andere, ist er überholt.
interface Ziel {
  sitz: string
  bei: string
}

export function useVorhang({
  partie,
  ansehen,
  gastsitze,
  offen,
}: {
  partie: Partie
  ansehen: (kennung: string, fuer: string) => void

  // gastsitze sind die Sitze, die Gäste spielen — `undefined`: unbekannt.
  gastsitze: readonly string[] | undefined

  // offen heißt: „Offene Hände“ ist gewählt.
  offen: boolean
}): Vorhangstand {
  const wer = entscheider(partie)
  const bei = `${partie.id}:${partie.fuer}`
  const stand = `${bei}:${wer ?? ''}`
  const [vorher, setzeVorher] = useState(stand)
  const [ziel, setzeZiel] = useState<Ziel>()

  // **Im Rendern, nicht in einem Effekt** — der Übergang ist eine Ableitung
  // aus zwei Ständen, und ein Effekt zeichnete zuerst einmal ohne Vorhang.
  if (vorher !== stand) {
    setzeVorher(stand)

    if (
      vorher.startsWith(`${bei}:`) &&
      wer !== undefined &&
      wer !== partie.fuer &&
      partie.fuer !== ''
    ) {
      setzeZiel({ sitz: wer, bei })
    }
  }

  const amGeraet = (sitz: string) =>
    gastsitze !== undefined && !gastsitze.includes(sitz)
  const offenesZiel =
    ziel?.bei === bei &&
    ziel.sitz === wer &&
    amGeraet(ziel.sitz) &&
    amGeraet(partie.fuer)
      ? ziel
      : undefined

  // **Mit „Offenen Händen“ folgt die Sicht** — einmal je Übergang.
  const gefolgt = useRef<Ziel>(undefined)
  useEffect(() => {
    if (
      !offen ||
      offenesZiel === undefined ||
      gefolgt.current === offenesZiel
    ) {
      return
    }

    gefolgt.current = offenesZiel
    ansehen(partie.id, offenesZiel.sitz)
  }, [offen, offenesZiel, partie.id, ansehen])

  const bestaetigen = useCallback(() => {
    if (offenesZiel !== undefined) {
      ansehen(partie.id, offenesZiel.sitz)
    }

    setzeZiel(undefined)
  }, [offenesZiel, partie.id, ansehen])

  const bleiben = useCallback(() => {
    setzeZiel(undefined)
  }, [])

  return {
    vorhang: offen ? undefined : offenesZiel?.sitz,
    bestaetigen,
    bleiben,
  }
}
