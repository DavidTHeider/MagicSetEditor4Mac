// **Wo die Felder eines Tisches liegen und wie er ausgemessen wird** —
// herausgelöst am 2026-09-22, als `tischplan.ts` mit der Autoanordnung
// (PLAY-04.AK-03,
// #1712)
// über die Zeilengrenze aus §2 ging.
//
// # Der Schnitt folgt der Frage
//
// Nebenan steht, **welches Objekt wohin gehört** — Zonen, Sitze, Haufen,
// eigene Lagen. Hier steht, **wo das im Raum liegt**: zur Mitte oder von ihr
// weg, wo der gemeinsame Stapel steht, und wie groß der Tisch am Ende ist.

import { Kartenhoehe, Sitzabstand, Sitzhoehe, Stapelbreite } from './masse'

import { zonenrahmen } from './zonenfeld'

import type { Sitzflaeche } from './sitzflaeche'
import type { Tischplan } from './tischplan'
import type { Zonenfeld } from './zonenfeld'

// zurMitteVon sagt, in welche Richtung von einem Sitz aus die Mitte liegt.
//
// **Ein Rückfall auf -1, den kein Aufruf betritt**: ein Schlachtfeldfeld
// entsteht nur aus einer Sitzfläche (`zonenfelder`), und die steht in
// `flaechen`. Er steht da, weil TypeScript es nicht sieht.
export function zurMitteVon(flaechen: Sitzflaeche[], sitz: string): -1 | 1 {
  /* v8 ignore next 2 */
  return flaechen.find((eine) => eine.sitz === sitz)?.zurMitte ?? -1
}

// stapelfeld ist der MagicStack — die Zone, die keinem Sitz gehört (CR 405.1).
//
// **Er liegt in der Lücke zwischen den Sitzreihen.** Gibt es nur einen Sitz,
// liegt die Lücke über ihm, und dann steht das Feld im Negativen — bis
// `normalisiert` den ganzen Plan in den Ursprung schiebt.
export function stapelfeld(flaechen: Sitzflaeche[]): Zonenfeld | undefined {
  if (flaechen.length === 0) return undefined

  const fremde = flaechen.filter((eine) => eine.zurMitte === 1)
  const breiteste = Math.max(...flaechen.map((eine) => eine.x + eine.breite))
  const bandY = fremde.length > 0 ? Sitzhoehe : -Sitzabstand

  return {
    zone: 'stapel',
    sitz: '',
    x: (breiteste - Stapelbreite) / 2,
    y: bandY + (Sitzabstand - Kartenhoehe) / 2,
    breite: Stapelbreite,
    hoehe: Kartenhoehe,
    gestalt: 'faecher',
    verborgen: false,
    geordnet: true,
  }
}

// normalisiert schiebt den ganzen Plan so, dass seine linke obere Ecke im
// Ursprung liegt, und misst ihn aus.
//
// **Gemessen wird an Sitzflächen und Feldern**, nicht an den Lagen: eine Lage
// liegt immer in ihrem Feld, und der Rand einer Sitzfläche liegt außerhalb
// jedes Feldes. Wer nur die Lagen nähme, verlöre den Rand und die leeren
// Sitze gleich mit.
export function normalisiert(plan: Tischplan): Tischplan {
  // **Die Felder mit ihren Namensleisten** — die oberste Leiste läge sonst
  // über dem Plan, und die Kamera passte ihn so ein, dass sie über den Rand
  // ragt (bei einem einzigen Sitz steht der Stapel über ihm).
  const ecken = [...plan.sitze, ...plan.felder.map(zonenrahmen)]
  if (ecken.length === 0) return { ...plan, breite: 0, hoehe: 0 }

  const linke = Math.min(...ecken.map((eines) => eines.x))
  const obere = Math.min(...ecken.map((eines) => eines.y))

  const geschoben = <T extends { x: number; y: number }>(eines: T): T => ({
    ...eines,
    x: eines.x - linke,
    y: eines.y - obere,
  })

  const sitze = plan.sitze.map(geschoben)
  const felder = plan.felder.map(geschoben)
  const aussen = [...sitze, ...felder.map(zonenrahmen)]

  return {
    ...plan,
    sitze,
    felder,
    lagen: plan.lagen.map(geschoben),
    zahlen: plan.zahlen.map(geschoben),
    breite: Math.max(...aussen.map((eines) => eines.x + eines.breite)),
    hoehe: Math.max(...aussen.map((eines) => eines.y + eines.hoehe)),
  }
}
