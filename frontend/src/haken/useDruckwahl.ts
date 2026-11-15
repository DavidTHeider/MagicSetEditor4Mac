// **Der Druck, den die Suche übernommen hat** — Zweitblick über Phase 6 von
// 0.9.0, Befund 1 (#2327, AP-10 #2367).
//
// Hebt die Drucksuche ihre Wahl auf, geht aus dem Entwurf, was noch so
// dasteht, wie sie es gebracht hat; **was jemand danach von Hand geändert
// hat, bleibt**. Bis dahin räumte die Suche nur ihre sichtbare Auswahl: wer A
// wählte und B ohne Treffer suchte, nahm A auf, und die Kennung stand hinter
// „Weitere Angaben“.
//
// # Eine neue Runde nach dem Aufnehmen
//
// **Die Suche entsteht nach einem geglückten Aufnehmen neu** (`runde` als
// `key`): sonst stünde ihr Druck gewählt da, während das Formular leer ist —
// und ein schon gewähltes Optionsfeld meldet keinen zweiten Klick.
import { useState, type Dispatch, type SetStateAction } from 'react'

import type { ExemplarDTO } from '../api'
import type { Listentreffer } from '../antwortenlisteneinfuhr'

// Die Angaben, die ein gewählter Druck in den Entwurf bringt.
const druckangaben = ['printing', 'name', 'setcode', 'nummer'] as const

export function useDruckwahl(
  setzeEntwurf: Dispatch<SetStateAction<ExemplarDTO>>,
) {
  const [uebernommen, setzeUebernommen] = useState<Listentreffer>()
  const [runde, setzeRunde] = useState(0)

  const gewaehlt = (treffer: Listentreffer | undefined) => {
    if (treffer !== undefined) {
      setzeUebernommen(treffer)
      setzeEntwurf((vorher) => ({
        ...vorher,
        printing: treffer.printing,
        name: treffer.name,
        setcode: treffer.setcode,
        nummer: treffer.nummer,
      }))

      return
    }

    if (uebernommen === undefined) {
      return
    }

    setzeUebernommen(undefined)
    setzeEntwurf((vorher) => {
      const geraeumt = { ...vorher }
      for (const angabe of druckangaben) {
        if (vorher[angabe] === uebernommen[angabe]) {
          geraeumt[angabe] = ''
        }
      }

      return geraeumt
    })
  }

  const neueRunde = () => {
    setzeUebernommen(undefined)
    setzeRunde((vorher) => vorher + 1)
  }

  return { runde, gewaehlt, neueRunde }
}
