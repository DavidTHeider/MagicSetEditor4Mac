// **Was ein Gast am eigenen Tisch tut, erreicht das eigene Fenster** — Befund
// aus dem Zweitblick über Phase 6.11 (#1672, zu Z1).
//
// # Wofür es das gibt
//
// **Go meldet es seit #1627** (`MitFenster` am Wirt): zieht ein Gast, geht
// es an das Fenster des Wirts. Gehört hat darauf bis hierher nur der Haken
// des **Gastes** — der eigene Tisch zeigte den Stand von vorhin, bis der
// Wirt selbst etwas tat. **Jetzt unter eigenem Namen**: dieselbe Anwendung
// kann zugleich Gast an einem anderen Tisch sein, und dessen Haken nähme
// sonst das Bild dieses Tisches.
//
// # Was es nimmt, und was nicht
//
// - **Nur die offene Partie.** Ein Bild einer anderen gehört nicht hierher.
// - **Nur die Sicht des eigenen Sitzes.** Der Wirt rechnet für den Sitz
//   seiner Freigabe; sieht das Fenster gerade einen anderen, bliebe sonst
//   eine fremde Hand stehen. **Und die private Ergänzung eines Zugs** nur,
//   wenn das letzte Bild bestätigt hat, dass es für diesen Sitz gerechnet
//   ist — der Zug kommt vor seinem Bild (`partietischfunk.go`).
// - **Eine Lücke wird genannt** (Z3): fiel verborgen mehr weg, als
//   nachgeholt wird, gibt der Haken es zurück — für die offene Partie.
// - **Kein älteres Bild über ein neueres.** Wer den Tisch wieder öffnet,
//   bekommt nachgeholt, was dazwischen kam (`nachholen.ts`) — und das ist
//   älter als die Sicht, die `ansehen` gerade geholt hat. **Gemessen an der
//   höchsten angenommenen Fassung**, die sofort gilt, und nicht nur an der
//   gerenderten: zwei Bilder hintereinander, 5 und dann 4, ließen sonst die
//   4 gewinnen (N3 aus dem Nachlauf über 6.11).
// - **Nichts während eines Sitzwechsels** (N1): bis die neue Sicht da ist,
//   steht die alte noch, und ein Zug für den alten Sitz trüge seine private
//   Ergänzung in die Liste, die danach dem neuen gehört.
import {
  useEffect,
  useRef,
  useState,
  type Dispatch,
  type SetStateAction,
} from 'react'

import { epocheBekannt, epocheGesehen, epocheVeraltet } from '../bestandsepoche'
import { horcheNachholend } from '../nachholen'

import type { Partie, PartieAntwort, Spielereignis } from '../apipartie'
import type { Regelstand } from './partielauf'

// Die Ereignisse des eigenen Tisches — **wörtlich dieselben wie in
// `internal/app/partiebeitritt.go`**; `fensterereignisse_test.go` hält beide
// Seiten zusammen.
export const ereignisEigenesTischbild = 'partie:eigen:tischbild'
export const ereignisEigenerZug = 'partie:eigen:zug'

// stelle ist der Name, unter dem der eigene Tisch nachholt.
const stelle = 'partie-eigen'

type Tischbild = Pick<PartieAntwort, 'partie' | 'erlaubt' | 'handlungen'>

export function useEigenerTischfunk(
  partie: Partie | undefined,
  wechselt: boolean,
  setzePartie: Dispatch<SetStateAction<Partie | undefined>>,
  setzeRegeln: Dispatch<SetStateAction<Regelstand>>,
  setzeEreignisse: Dispatch<SetStateAction<Spielereignis[]>>,
): boolean {
  const offen = useRef(partie)
  const imWechsel = useRef(wechselt)
  const hoechste = useRef(0)
  // epoche ist die Epoche des Bestands, zu der [hoechste] gehört (#2307).
  const epoche = useRef(0)
  const [luecke, setzeLuecke] = useState<string>()
  const bestaetigt = useRef<string>(undefined)

  useEffect(() => {
    offen.current = partie
    imWechsel.current = wechselt
  })

  const kennung = partie?.id

  useEffect(() => {
    if (kennung === undefined) return undefined

    bestaetigt.current = undefined
    hoechste.current = 0
    epoche.current = epocheBekannt()

    const abmelden = [
      horcheNachholend(ereignisEigenesTischbild, stelle, (daten) => {
        const [von, bild, aus] = daten as [string, Tischbild, number?]
        const jetzt = offen.current
        if (jetzt?.id !== von || imWechsel.current || epocheVeraltet(aus)) {
          return
        }
        epocheGesehen(aus)

        bestaetigt.current = bild.partie.fuer
        if (bild.partie.fuer !== jetzt.fuer) return

        // **Aus einem neuen Bestand gilt auch eine niedrigere Fassung**
        // (#2307): nach dem Zurückspielen einer älteren Sicherung derselben
        // Partie ist sie die richtige, und die gemerkte von vorher nicht.
        const neuerBestand = aus !== undefined && aus > epoche.current
        const fassung = Number(bild.partie.fassung)
        if (
          !neuerBestand &&
          fassung < Math.max(Number(jetzt.fassung), hoechste.current)
        ) {
          return
        }

        if (aus !== undefined) epoche.current = Math.max(epoche.current, aus)
        hoechste.current = fassung
        // **Die angenommene Partie gilt sofort** (Zweitblick über die Folgen
        // vor 0.8.0, Befund 7): der Nachholpuffer liefert synchron, und bis
        // zum nächsten Effekt wäre die Untergrenze noch die Fassung des alten
        // Bestands — ein zweites Bild des neuen fiele dagegen durch.
        offen.current = bild.partie
        setzePartie(bild.partie)
        setzeRegeln({ erlaubt: bild.erlaubt, handlungen: bild.handlungen })
      }),
      horcheNachholend(
        ereignisEigenerZug,
        stelle,
        (daten) => {
          const [von, ereignis, aus] = daten as [string, Spielereignis, number?]
          const jetzt = offen.current
          if (jetzt?.id !== von || imWechsel.current || epocheVeraltet(aus)) {
            return
          }
          epocheGesehen(aus)

          const gezeigt =
            bestaetigt.current === jetzt.fuer
              ? ereignis
              : { ...ereignis, privat: {} }

          setzeEreignisse((bisher) => [...bisher, gezeigt])
        },
        () => {
          setzeLuecke(kennung)
        },
      ),
    ]

    return () => {
      for (const weg of abmelden) weg()
    }
  }, [kennung, setzePartie, setzeRegeln, setzeEreignisse])

  return kennung !== undefined && luecke === kennung
}
