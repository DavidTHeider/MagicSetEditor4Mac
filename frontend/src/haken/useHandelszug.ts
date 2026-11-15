// **Ein Zug an einem Handel, als Haken** — P13
// (#2381).
//
// Was der Zug tut, steht in `handelszug.ts`; hier steht, wie er an einem
// Knopf hängt: **derselbe Riegel wie Schritte, Vorschläge und Buchung** (§7,
// Regel 4) und ein Ausgang, den der Knopf zeigt. Dazu der Ausweg, wenn nur
// der Weg hinaus hängen blieb — in den Ausgang legen und übertragen, auf
// den Klick hin und nicht von selbst (0152 §4).
import { useRef, useState } from 'react'

import { api } from '../fassade'
import { ziehe } from '../handelszug'
import { auspacken } from '../umschlag'

import type { Handelsvorgang as Vorgangsstand } from '../apihandelsleitung'
import type { Zugauftrag, Zugausgang } from '../handelszug'
import type { Vorgang } from './useVorgang'

// Zugbeginn ist, wovon ein Zug ausgeht — ein Stand, oder ein Weg zu ihm, der
// selbst scheitern kann (Eröffnen vor dem Anbieten); dann gibt er nichts.
export interface Zugbeginn extends Omit<Zugauftrag, 'gegenueber' | 'stand'> {
  stand: Vorgangsstand | (() => Promise<Vorgangsstand | undefined>)
}

// Handelszug ist, was ein Raum seinen Kästen reicht: **ein** Ausgang je
// Gegenüber, damit er stehen bleibt, wenn ein Kasten beim Wechsel des
// Vorgangs neu entsteht (nach dem Eröffnen).
export type Handelszug = ReturnType<typeof useHandelszug>

export function useHandelszug(kennung: string, lauf: Vorgang) {
  const [ausgang, setzeAusgang] = useState<Zugausgang>()

  // **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (§7, Regel 2):
  // wer während eines Zugs einen anderen Handel wählt, zählt die Runde
  // weiter, und die Meldung des alten Zugs bleibt aus.
  const runde = useRef(0)

  const kennzeichen = `handelsvorgang/${kennung}`

  // ziehen geht einen Zug und reicht seinen Ausgang weiter — auch einen, in
  // dem nichts gegangen ist: der Aufrufer entscheidet, was er nachlädt.
  const ziehen = (
    beginn: Zugbeginn,
    nachher: (neu: Zugausgang) => void,
  ): Promise<void> =>
    lauf.starte(kennzeichen, async () => {
      setzeAusgang(undefined)
      const meine = runde.current

      const stand =
        typeof beginn.stand === 'function' ? await beginn.stand() : beginn.stand

      if (stand === undefined) {
        return
      }

      const neu = await ziehe({ ...beginn, stand, gegenueber: kennung })

      if (runde.current === meine) {
        setzeAusgang(neu)
      }

      nachher(neu)
    })

  // nochmal bringt hinaus, was hängen blieb — `legen` heißt: es kam nicht
  // einmal in den Ausgang.
  const nochmal = (bisher: Zugausgang, legen: boolean): Promise<void> =>
    lauf.starte(kennzeichen, async () => {
      const { stand, gegangen } = bisher
      const meine = runde.current
      let gelegt = !legen
      let neu: Zugausgang

      try {
        if (legen) {
          const antwort = await api().HandelSchicken(stand.vorgang, kennung)
          auspacken(antwort, () => antwort)
          gelegt = true
        }

        const bericht = await api().HandelZustellen(kennung)
        const getragen = auspacken(bericht, () => bericht.getragen)

        neu = { stand, gegangen, getragen }
      } catch (ursache) {
        neu = gelegt
          ? { stand, gegangen, getragen: 0, zustellung: ursache }
          : {
              stand,
              gegangen,
              getragen: 0,
              gescheitert: { teil: 'ausgang', schritt: '', ursache },
            }
      }

      if (runde.current === meine) {
        setzeAusgang(neu)
      }
    })

  // vergessen räumt die Meldung, wenn ein anderer Vorgang gewählt wird (§7,
  // Regel 1): sie gehörte zu dem vorigen.
  const vergessen = () => {
    runde.current += 1
    setzeAusgang(undefined)
  }

  return {
    ausgang,
    ziehen,
    nochmal,
    vergessen,
    laeuft: lauf.laeuft(kennzeichen),
  }
}
