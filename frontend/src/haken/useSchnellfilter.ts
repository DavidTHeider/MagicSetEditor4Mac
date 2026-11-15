// **Der Schnellfilter der Kartenliste** — Werkstück U5 von Phase 5.1
// (#1416).
//
// Was jemand ins Filterfeld tippt, geht an `KartenFiltern`; zurück kommen die
// Kennungen, die bleiben. Die Liste selbst bleibt stehen — sie hat das Fenster
// schon —, und was nicht dabei ist, wird verborgen. **Die Sprache entscheidet
// Go**, portiert aus MSE2 (`src/data/filter.cpp`, Fassung f9d9356); hier
// steht nur das Tippen.
//
// # Warum `useLaden` mit `behalten`
//
// Jede Änderung der Anfrage ist ein Ladeeffekt (§7): eine Antwort gehört zu
// dem Stand, aus dem sie stammt, und wer schnell tippt, bekommt sonst die
// vorletzte. `behalten` lässt die vorige Trefferliste stehen, bis die neue da
// ist — sonst flackerte die Liste bei jedem Buchstaben einmal auf „alle".
//
// # Was er nicht tut
//
// **Keine Verzögerung.** MSE2 filtert bei jedem Zeichen
// (`FilterCtrl::onChangeEvent`, `src/gui/control/filter_ctrl.cpp`), und der
// Filter läuft im Speicher über ein Set; eine Frist wäre eine Vermutung über
// eine Langsamkeit, die niemand gemessen hat.

import { useCallback, useState } from 'react'

import type { ListenspalteDTO } from '../api'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from './useLaden'

export interface Schnellfilterstand {
  anfrage: string
  setzeAnfrage: (anfrage: string) => void

  // treffer ist `undefined`, solange kein Filter steht — dann bleibt alles.
  treffer: ReadonlySet<string> | undefined

  laedt: boolean
  fehler: string | undefined
}

// # Warum ein Bezug
//
// **Ein Treffer hängt nicht nur an der Anfrage.** Seit FP-01 (#2155) sieht ein
// freier Begriff die gerechneten Werte der **sichtbaren** Spalten, und jeder
// Begriff die Werte der Karten. `stand` und die sichtbaren `spalten` benennen
// beides — wechselt eines, fragt der Filter neu (Befund 1 des Zweitblicks über
// Phase G: eine ausgeblendete Spalte ließ die Treffer der alten Wahl stehen).
export function useSchnellfilter(
  setID: string,
  stand: number,
  spalten: readonly ListenspalteDTO[],
): Schnellfilterstand {
  const [anfrage, setzeAnfrage] = useState('')
  const leer = anfrage.trim() === ''

  const holen = useCallback(async (): Promise<
    ReadonlySet<string> | undefined
  > => {
    if (leer) {
      return undefined
    }

    const antwort = await api().KartenFiltern(setID, anfrage)

    return new Set(auspacken(antwort, () => antwort.kartenIDs))
  }, [setID, anfrage, leer])

  // **Als Text**, damit ein neu geladener, gleicher Aufbau nicht neu fragt.
  const bezug = `${String(stand)}|${spalten
    .filter((spalte) => spalte.sichtbar)
    .map((spalte) => spalte.feld)
    .join('\u0000')}`

  const { daten, laedt, fehler } = useLaden(
    holen,
    [setID, anfrage, bezug],
    true,
  )

  return {
    anfrage,
    setzeAnfrage,
    treffer: leer ? undefined : daten,
    laedt: !leer && laedt,
    fehler,
  }
}
