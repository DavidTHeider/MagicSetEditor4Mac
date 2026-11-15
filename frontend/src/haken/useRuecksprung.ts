// **Was zurückgehen kann und was eine Rückkehr hinterlassen hat** — Werkstück
// O7 von Phase 5.2
// (#1431).
//
// # Die Regeln aus §7, die er hält
//
//  1. **Ein Vorgang je Datensatz** (Regel 4): der Datensatz ist die Epoche,
//     und `useVorgang` hält je Kennung genau einen Lauf.
//  2. **Geschrieben ist geschrieben** (Regel 3): nach einer geglückten
//     Quittung wird nachgeladen; scheitert das, bleibt die Quittung stehen und
//     der Lesefehler wird gemeldet. Nach einer **gescheiterten** wird gar
//     nicht nachgeladen.
//  3. **Regel 1 und 2**: es gibt kein Bezugsobjekt, das wechseln könnte — die
//     Rückkehren gehören dem Bestand, und es gibt genau einen.
import { useCallback, useState } from 'react'

import type { RueckkehrDTO, RuecksprungartDTO } from '../apirueckkehr'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useLaden } from './useLaden'
import { useTexte } from '../sprache/kontext'
import type { Vorgang } from './useVorgang'

export interface Ruecksprungstand {
  arten: RuecksprungartDTO[]
  rueckkehren: RueckkehrDTO[]
  laedt: boolean
  fehler: string | undefined

  // quittieren erklärt eine Rückkehr für nachgesehen.
  quittieren: (kennung: number) => void

  // quittiert sagt, welche Kennung gerade läuft.
  laeuft: (kennung: number) => boolean
}

export function useRuecksprung(vorgang: Vorgang): Ruecksprungstand {
  const texte = useTexte()

  const [runde, setzeRunde] = useState(0)
  const [schreibfehler, setzeSchreibfehler] = useState<string>()

  const arten = useLaden(async () => {
    const antwort = await api().Ruecksprungarten()

    return auspacken(antwort, () => antwort.arten)
  }, [])

  const rueckkehren = useLaden(async () => {
    const antwort = await api().Rueckkehren()

    return auspacken(antwort, () => antwort.rueckkehren)
  }, [runde])

  const quittieren = useCallback(
    (kennung: number) => {
      void vorgang.starte(`rueckkehr:${String(kennung)}`, async () => {
        try {
          const antwort = await api().RueckkehrQuittieren(kennung)

          auspacken(antwort, () => undefined)
          setzeSchreibfehler(undefined)

          // **Nachgeladen wird nur nach einem geglückten Schreiben** (§7,
          // Regel 3): sonst überschriebe ein Lesefehler die Meldung, die
          // sagt, was zu tun ist.
          setzeRunde((bisher) => bisher + 1)
        } catch (ursache: unknown) {
          setzeSchreibfehler(meldungAus(ursache, texte))
        }
      })
    },
    [texte, vorgang],
  )

  return {
    arten: arten.daten ?? [],
    rueckkehren: rueckkehren.daten ?? [],
    laedt: arten.laedt || rueckkehren.laedt,
    fehler: schreibfehler ?? arten.fehler ?? rueckkehren.fehler,
    quittieren,
    laeuft: (kennung) => vorgang.laeuft(`rueckkehr:${String(kennung)}`),
  }
}
