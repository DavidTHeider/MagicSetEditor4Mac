// Regel 2 aus AGENTS.md §7: **eine Antwort gehört zu dem Stand, aus dem sie
// stammt.**
//
// Wer Set oder Karte wechselt, während eine Anfrage läuft, bekommt die Antwort
// der alten. Ohne einen Zähler trägt sie sich in den neuen Stand ein, und dort
// steht dann die Kartenliste des vorigen Sets — ohne Fehlermeldung, ohne
// erkennbaren Anlass.
//
// **Für Erfolg, Fehler und Rücknahme gleichermaßen.** Der Fehler ist die
// unauffälligere Hälfte: eine Meldung „Set nicht lesbar" zu einem Set, das
// niemand mehr offen hat, sieht aus wie ein Fehler des aktuellen.
import { useCallback, useMemo, useRef } from 'react'

export interface Lauf {
  // naechster zählt weiter und gibt die neue Nummer. Vor jedem Aufruf einmal.
  naechster: () => number

  // nummer gibt den heutigen Stand, **ohne weiterzuzählen**.
  //
  // **Für einen Schreibvorgang und nicht für ein Lesen.** Ein Schreiben
  // wechselt das Bezugsobjekt nicht — es merkt sich nur, welches gemeint war,
  // damit sein später Fehler nicht über einem anderen landet. Wer hier
  // `naechster` nähme, verwürfe jede laufende Leseantwort mit.
  nummer: () => number

  // istAktuell sagt, ob eine Antwort noch zum heutigen Stand gehört.
  istAktuell: (nummer: number) => boolean
}

export function useLauf(): Lauf {
  const stand = useRef(0)

  const naechster = useCallback(() => {
    stand.current += 1

    return stand.current
  }, [])

  const nummer = useCallback(() => stand.current, [])

  const istAktuell = useCallback(
    (nummer: number) => nummer === stand.current,
    [],
  )

  // **Ein Verbund, der derselbe bleibt.**
  //
  // Ohne `useMemo` entstünde bei jedem Zeichnen ein neuer, und jeder Effekt mit
  // `lauf` in den Abhängigkeiten liefe erneut. Gefunden hat es die
  // Sprachmeldung: sie ging beim Start zweimal hinaus, weil zwischen den beiden
  // Zeichenvorgängen nichts gleich blieb außer der Absicht.
  return useMemo(
    () => ({ naechster, nummer, istAktuell }),
    [naechster, nummer, istAktuell],
  )
}
