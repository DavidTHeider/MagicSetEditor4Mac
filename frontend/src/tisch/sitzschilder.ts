// **Was an einem Sitz steht** — PLAY-03
// (#1711),
// gezeichnet als `tsitz-kopf` im Artboard `Tisch` des Entwurfs.
//
// # Wofür es sie gibt
//
// **Weil die Zahlen eines Sitzes an seinem Sitz stehen gehören.** `leben` und
// `vergiftung` liegen seit G9 am Modell, und seit `M3` zeigt sie die
// `Spielerleiste` — als Leiste **neben** dem Tisch. Bei zwei Sitzen geht das;
// bei vieren (Artboard `TischCommander`) steht dort eine Liste, und wer
// wissen will, wie viel Leben der Sitz gegenüber hat, ordnet Namen und Zahlen
// im Kopf einander zu.
//
// **Der Entwurf löst es anders**: jeder Sitz trägt seinen Kopf an seiner
// Außenkante — Name, Leben, Gift, und ob er am Zug ist.
//
// # Warum das Wort von außen kommt
//
// **Dieselbe Überlegung wie bei `zonenschilder`**: ein Sitzkopf heißt in zwei
// Sprachen zweierlei, und die Kataloge liegen in `sprache/`. Ein Zeichner,
// der sie läse, hätte die Oberflächensprache im Renderer. Hier steht, **wo**
// das Wort hinkommt; **was** es sagt, baut `partiewoerter.ts`.

import { Kopfhoehe, Randbreite } from './masse'

import type { Tischplan } from './tischplan'
import type { Zonenschild } from './zonenschilder'

// Sitzschildgroesse ist die Höhe der Schrift in Weltpunkten — in der Stufe
// „mittel" (`schildmass.ts`).
//
// **Größer als ein Zonenschild** (sechzehn): der Sitzkopf nennt den Spieler
// und seinen Stand, ein Zonenname sagt, was in einem Feld liegt.
// **Zweiundzwanzig in einem Band von zweiunddreißig** (`masse.ts`,
// [Kopfhoehe]) — bei der Anfangskamera eines Fensters von 1.440 Punkten
// knapp dreizehn Punkte.
export const Sitzschildgroesse = 22

// sitzschilder gibt je Sitz sein Schild an seiner Außenkante.
//
// **An der Außenkante und nicht zur Mitte**: zur Mitte hin stünde der Kopf
// zwischen zwei Schlachtfeldern, und niemand wüsste, zu welchem er gehört.
// Der Entwurf zeichnet es genauso — der Kopf des Gegenübers steht **über**
// seinen Zonen, der eigene **unter** den eigenen.
export function sitzschilder(
  plan: Tischplan,
  wort: (sitz: string) => string,
): Zonenschild[] {
  return plan.sitze.map((flaeche) => {
    // **`zurMitte === -1` heißt: die Mitte liegt oben**, der Sitz also unten
    // — seine Außenkante ist die Unterkante, und der Text hängt nach oben
    // hinein (Anker 1).
    const untenSitzend = flaeche.zurMitte === -1

    // **In der Mitte seines Bandes**, bündig mit den Zonen darunter —
    // `masse.ts` gibt ihm die [Kopfhoehe] an der Außenkante des Sitzes.
    return {
      kennung: `${flaeche.sitz}:sitz:kopf`,
      x: flaeche.x + Randbreite,
      y: untenSitzend
        ? flaeche.y + flaeche.hoehe - Randbreite - Kopfhoehe / 2
        : flaeche.y + Randbreite + Kopfhoehe / 2,
      text: wort(flaeche.sitz),
      groesse: Sitzschildgroesse,
      ankerY: 0.5,
      art: 'kopf',
    }
  })
}
