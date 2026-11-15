// **Eine Wahl gilt für das Bild, aus dem sie stammt** — UXR-F6
// (#2013),
// AGENTS.md §7, Regel 2.
//
// # Wofür
//
// Die Wahl eines Balkens nimmt dessen Karten mit auf die Kartenliste und
// behält sie dort — auch wenn jemand dazwischen eine Karte ändert. Kommt er
// zur Statistik zurück, rechnet sie neu, und die Karten des Balkens können
// andere sein. MSE2 verwirft in diesem Fall die Wahl: `GraphControl::setData`
// leert `current_item` (`src/gui/control/graph.cpp`, Zeile 1084, Fassung
// f9d9356 vom 2021-07-06), und `StatsPanel::onChange` setzt die Daten bei
// jeder Änderung neu.
//
// # Was er tut
//
// Er gleicht die Wahl mit dem **geladenen** Bild ab. Steht der Balken noch da
// und trägt er andere Karten, folgt die Wahl ihm; ist er verschwunden oder
// leer, endet sie. Das ist freundlicher als MSE2, das die Wahl in jedem Fall
// verwirft — **angepasst** und nicht übernommen: wer den Balken gewählt hat,
// meinte ihn, und nicht den Stand von vorhin.
//
// **Solange das Bild lädt, tut er nichts.** Ein leeres Bild vor der Antwort
// ist keine Aussage über den Balken.
import { useEffect } from 'react'

import type { WertzahlDTO } from '../api'
import { useStatistikfilter } from '../statistikfilter'

// gleich vergleicht zwei Kennungslisten nach Inhalt, in der Reihenfolge — beide
// kommen aufsteigend aus dem Dienst.
function gleich(a: readonly string[], b: readonly string[]): boolean {
  return (
    a.length === b.length && a.every((kennung, stelle) => kennung === b[stelle])
  )
}

// useGruppenabgleich gleicht die Wahl dieses Bildes mit seinen Balken ab.
// `eintraege` ist `undefined`, solange das Bild lädt.
export function useGruppenabgleich(
  wahl: string,
  eintraege: readonly WertzahlDTO[] | undefined,
): void {
  const filter = useStatistikfilter()
  const gewaehlt = filter.wahl

  useEffect(() => {
    if (gewaehlt?.wahl !== wahl || eintraege === undefined) {
      return
    }

    const kartenIDs = eintraege.find(
      (eintrag) => eintrag.name === gewaehlt.gruppe,
    )?.kartenIDs

    if (kartenIDs === undefined || kartenIDs.length === 0) {
      filter.aufheben()
    } else if (!gleich(kartenIDs, gewaehlt.kartenIDs)) {
      filter.waehlen({ ...gewaehlt, kartenIDs })
    }
  }, [wahl, eintraege, gewaehlt, filter])
}
