// **Ein Wert, der erst nach einer Frist nachrückt** — für die Referenzsuche
// der Vorlagenwerkstatt (Werkstück U7 von Phase 5.1,
// #1418).
//
// Der Schreibzeiger wandert mit jedem Anschlag, und mit ihm das Wort darunter.
// Eine Suche nach Verwendungen liest **alle** Dateien des Pakets; sie je
// Anschlag zu starten hieße, das Paket beim Tippen dauernd zu lesen —
// `useLaden` verwirft die überholten Antworten, aber gelesen wären sie. Die
// Frist lässt den Zeiger zur Ruhe kommen, bevor gesucht wird.
//
// **Das `setState` steht im Zeitgeber, nicht im Effekt** — ein Effekt, der
// sofort setzt, zeichnet zweimal und ist genau das Muster, das eslint abweist.
import { useEffect, useState } from 'react'

export function useVerzoegert<T>(wert: T, frist: number): T {
  const [verzoegert, setzeVerzoegert] = useState(wert)

  useEffect(() => {
    const zeitgeber = setTimeout(() => {
      setzeVerzoegert(wert)
    }, frist)

    return () => {
      clearTimeout(zeitgeber)
    }
  }, [wert, frist])

  return verzoegert
}
