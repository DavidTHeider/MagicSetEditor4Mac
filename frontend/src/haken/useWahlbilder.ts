// Die Wahlbilder einer Stiloption
// (#1225).
//
// # Warum ein eigener Haken
//
// **Weil er sparsam sein muss.** Über `full-magic-pack/data` stehen 1.097
// `choice images:`-Blöcke; jedes Bild ist ein eigener Rasterlauf. Gefragt wird
// deshalb nur dort, wo die Option selbst sagt, dass es etwas gibt
// (`hatWahlbilder`) — und je Option einmal.
import { useCallback } from 'react'

import { api, auspacken, type WahlbildDTO } from '../api'
import { useLaden } from './useLaden'

// keineBilder ist die Antwort für eine Option ohne Bilder.
//
// **Eine feste Kennung und kein `[]` an der Aufrufstelle**: ein frisches
// leeres Feld je Durchgang wäre eine neue Abhängigkeit und stieße das Laden
// bei jedem Zeichnen erneut an (§7, Regel 2).
const keineBilder: WahlbildDTO[] = []

// useWahlbilder holt die Bilder einer Stiloption — oder nichts.
export function useWahlbilder(
  setID: string,
  option: string,
  gefragt: boolean,
): WahlbildDTO[] {
  const holen = useCallback(async () => {
    if (!gefragt) {
      return keineBilder
    }

    const antwort = await api().Stilwahlbilder(setID, option)

    return auspacken(antwort, () => antwort.bilder)
  }, [setID, option, gefragt])

  // **Ein Fehler bleibt hier liegen, und das ist Absicht.** Die Bilder sind
  // eine Zugabe: geht der Rasterlauf schief, steht die Wahl mit ihrem Namen
  // da — genau wie vorher. Eine rote Meldung neben einer bedienbaren Auswahl
  // sagte, etwas sei kaputt, obwohl nichts fehlt (§8).
  const { daten } = useLaden<WahlbildDTO[]>(holen, [setID, option, gefragt])

  return daten ?? keineBilder
}
