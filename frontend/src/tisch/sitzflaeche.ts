// **Wo ein Sitz am Tisch liegt** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Was die Regel sagt
//
// **CR 403.1** (Fassung 2026-08-19): das Schlachtfeld ist die Fläche
// *zwischen* den Spielern, und die Permanenten eines Spielers liegen *vor
// ihm*. Daraus folgt die Anordnung — jeder Sitz bekommt eine eigene Fläche,
// und die Flächen liegen einander gegenüber.
//
// **CR 400.5** erlaubt die freie Anordnung ausdrücklich, unter drei
// Bedingungen: wer kontrolliert, ob etwas getappt ist und was daran hängt,
// muss für alle klar bleiben. Diese drei tragen die Darstellung, nicht die
// Lage — sie stehen in `zonenlage.ts`.
//
// # Und was sie nicht sagt
//
// **Wer unten liegt, ist keine Regel.** Es ist eine Entscheidung dieser
// Oberfläche: unten liegt, aus wessen Sicht gerechnet wurde. Ein Zuschauer
// hat keine Sicht — dann liegt der **erste** Sitz unten, und zwar der erste
// in der Reihenfolge, in der die Partie ihn führt. Das ist keine Willkür,
// sondern dieselbe Ordnung, die auch jede Liste nimmt.
import { Sitzabstand, Sitzbreite, Sitzhoehe } from './masse'

// Sitzflaeche ist der Bereich, den ein Sitz einnimmt — in Millimetern.
export interface Sitzflaeche {
  sitz: string

  // eigen sagt, ob für **diesen** Sitz gerechnet wurde. Ein Zuschauer hat
  // keinen eigenen, und dann ist es nirgends wahr.
  eigen: boolean

  x: number
  y: number
  breite: number
  hoehe: number

  // zurMitte sagt, in welche Richtung die Tischmitte liegt: -1 nach oben,
  // +1 nach unten.
  //
  // **Daraus folgt die Reihenfolge der Zonen.** Die Hand liegt am weitesten
  // von der Mitte, das Schlachtfeld daneben — CR 403.1 nennt es „vor ihm",
  // und vor einem Spieler heißt zur Mitte hin.
  zurMitte: -1 | 1
}

export function sitzflaechen(sitze: string[], eigener: string): Sitzflaeche[] {
  const [erster] = sitze
  if (erster === undefined) return []

  const bezug = sitze.includes(eigener) ? eigener : erster
  const fremde = sitze.filter((einer) => einer !== bezug)

  const obenBreite =
    fremde.length === 0
      ? Sitzbreite
      : fremde.length * Sitzbreite + (fremde.length - 1) * Sitzabstand

  const flaechen: Sitzflaeche[] = fremde.map((einer, stelle) => ({
    sitz: einer,
    eigen: false,
    x: stelle * (Sitzbreite + Sitzabstand),
    y: 0,
    breite: Sitzbreite,
    hoehe: Sitzhoehe,
    zurMitte: 1,
  }))

  flaechen.push({
    sitz: bezug,
    eigen: bezug === eigener,
    x: (obenBreite - Sitzbreite) / 2,
    y: fremde.length === 0 ? 0 : Sitzhoehe + Sitzabstand,
    breite: Sitzbreite,
    hoehe: Sitzhoehe,
    zurMitte: -1,
  })

  return flaechen
}
