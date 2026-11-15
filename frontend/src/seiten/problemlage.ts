// **Was die Prüfstände der Sets zusammen sagen** — AP-8, Teil 2
// (#2365).
//
// **Vollständig heißt: jedes Set aktuell geprüft.** Nur dann tragen die
// Zahlen — Sets mit Problemen, Karten mit Befund —, und nur dann sind sie
// ein Ziel. Ein ungeprüftes, veraltetes oder abgebrochenes Set macht aus
// „0“ eine Behauptung (`Start09`: „Unvollständig heißt nie 0“).
import type { SetproblemstandDTO } from '../apiprobleme'

// Setbefund ist ein Set mit Karten mit Befund — für die Zeile der
// Aufmerksamkeit und den Filter der Kartenliste (AP-8, Teil 2b).
export interface Setbefund {
  setID: string
  name: string
  zahl: number
  karteIDs: readonly string[]
}

export interface Problemlage {
  // gesamt sind die Sets, geprueft die aktuell geprüften.
  gesamt: number
  geprueft: number
  vollstaendig: boolean

  // mitProblemen sind die aktuell geprüften Sets mit wenigstens einer Karte
  // mit Befund.
  mitProblemen: readonly Setbefund[]
  karten: number

  // kartenGeprueft und kartenGesamt zählen über alle geprüften Sets;
  // alter ist das des ältesten Stands, in Sekunden.
  kartenGeprueft: number
  kartenGesamt: number
  alter: number
}

export function problemlage(
  staende: readonly SetproblemstandDTO[],
): Problemlage {
  const sicher = staende.filter((stand) => stand.sicher)

  return {
    gesamt: staende.length,
    geprueft: sicher.length,
    vollstaendig: staende.length > 0 && sicher.length === staende.length,
    mitProblemen: sicher
      .filter((stand) => stand.zahl > 0)
      .map((stand) => ({
        setID: stand.setID,
        name: stand.name,
        zahl: stand.zahl,
        karteIDs: stand.karten.map((karte) => karte.karteID),
      })),
    karten: sicher.reduce((summe, stand) => summe + stand.zahl, 0),
    kartenGeprueft: sicher.reduce((summe, stand) => summe + stand.geprueft, 0),
    kartenGesamt: sicher.reduce((summe, stand) => summe + stand.gesamt, 0),
    alter: Math.max(0, ...sicher.map((stand) => stand.alter)),
  }
}
