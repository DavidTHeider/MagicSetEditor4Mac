// **Erneut versuchen** — F05
// (#2344).
//
// Was die Wiederherstellung tut, wenn die Datei vermutlich heil ist: den
// Start wiederholen und danach den Befund neu holen lassen. **Ein Vorgang**
// (§7, Regel 4): ein zweiter Klick während des ersten öffnete denselben
// Bestand ein zweites Mal.
import { useState } from 'react'

import { api } from '../fassade'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from './useVorgang'

export interface Startversuch {
  wiederholen: () => void
  laeuft: boolean

  // weiterGescheitert sagt, dass der letzte Versuch nicht öffnete — sonst
  // sähe niemand, dass er überhaupt lief: Titel und Satz bleiben dieselben.
  weiterGescheitert: boolean

  fehler: string | undefined
}

export function useStartWiederholen(bestandErsetzt: () => void): Startversuch {
  const texte = useTexte()
  const vorgang = useVorgang()
  const [weiterGescheitert, setzeWeiterGescheitert] = useState(false)
  const [fehler, setzeFehler] = useState<string>()

  const wiederholen = () =>
    void vorgang.starte('start', async () => {
      try {
        const befund = await api().StartWiederholen()

        setzeWeiterGescheitert(!befund.offen)
        setzeFehler(undefined)

        // **Der Befund wird neu geholt**, auch wenn der Versuch scheiterte:
        // die Lage kann eine andere sein — erst belegt, jetzt voll.
        bestandErsetzt()
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  return {
    wiederholen,
    laeuft: vorgang.laeuft('start'),
    weiterGescheitert,
    fehler,
  }
}
