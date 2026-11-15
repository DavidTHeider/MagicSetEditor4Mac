// **Worauf die Kamera zielt** — PLAY-05.AK-05
// (#1713).
//
// # Der Befund
//
// `kamera.ts` kannte Schwenken, Zoomen und Drehen — **„auf die Auswahl" und
// „auf den eigenen Sitz" nicht**, und `Tischsteuerung.tsx` bot entsprechend
// nur Alles zeigen, Näher heran, Weiter weg und Tisch drehen. Wer in einer
// Partie mit vier Sitzen eine Karte suchte, musste sie mit der Maus finden.
//
// # Warum eine eigene Datei
//
// **Der Schnitt folgt der Frage** (§2): `kamera.ts` beantwortet, **wie** sich
// eine Kamera bewegt — verschieben, zoomen, drehen, einpassen. Hier steht,
// **worauf** sie zielt, und das braucht den Plan: eine Lage, eine Sitzfläche.
// Zusammen ginge `kamera.ts` außerdem über die Zeilengrenze.
//
// # Und warum `eingepasst` hier nicht doppelt steht
//
// **Weil es derselbe Rechenweg ist.** `eingepasst` legt den **ganzen** Plan in
// die Fläche; [aufFlaeche] tut dasselbe mit einem beliebigen Rechteck, und
// `eingepasst` ist seither der Fall „das Rechteck ist der Plan". Zwei
// Rechnungen nebeneinander liefen auseinander, sobald jemand eine von beiden
// anfasst.

import { begrenzt, imBogen } from './kamera'

import type { Flaeche, Kamera } from './kamera'
import type { Tischplan } from './tischplan'

// Zielrand ist die Luft in Bildpunkten, die um ein Ziel bleibt.
//
// **Dieselbe wie beim Einpassen**, und das ist der Punkt: ein Rand in
// Bildpunkten sagt „so viel Weiß am Bildschirmrand" und nichts über die Größe
// des Ziels.
export const Zielrand = 16

// Auswahlumfeld sagt, wie viel **mehr** als die Karte zu sehen sein soll.
//
// # Warum ein Faktor und kein Rand in Punkten
//
// **Weil ein Rand in Punkten zwei ganz verschiedene Dinge bedeutet.** Der
// erste Anlauf gab beiden Zielen 120 Punkte Luft, und der Lauf über den
// eigenen Sitz wurde rot: eine Sitzfläche ist **1200 × 300** und damit so
// breit wie der ganze Tisch — 120 Punkte auf jeder Seite ließen sie
// *kleiner* erscheinen als beim Einpassen (0,47 gegen 0,64). Eine Karte ist
// **63 × 88**; dieselben 120 Punkte sind dort fast alles.
//
// **Vier**, und das ist gerechnet: bei Faktor 4 steht die gewählte Karte
// mitten in einem Feld von 252 × 352 Weltpunkten — das sind gut drei Karten
// in der Breite und vier in der Höhe. Wer heranzoomt, sieht sie **und** ihre
// Nachbarn; sonst füllte sie den Schirm und sagte nichts über ihren Ort.
export const Auswahlumfeld = 4

// Rechteck ist ein Ausschnitt des Tisches in Weltpunkten.
export interface Rechteck {
  x: number
  y: number
  breite: number
  hoehe: number
}

// aufFlaeche legt ein Rechteck mittig in die Zeichenfläche.
//
// **Die Drehung bleibt und geht in die Rechnung ein** — derselbe Grund wie bei
// `eingepasst`: ein gedrehtes Rechteck braucht mehr Platz als ein aufrechtes,
// und ohne den Term liefe das Ziel bei jeder Drehung über den Rand.
export function aufFlaeche(
  ziel: Rechteck,
  flaeche: Flaeche,
  rand: number,
  drehung: number,
): Kamera {
  const cos = Math.abs(Math.cos(imBogen(drehung)))
  const sin = Math.abs(Math.sin(imBogen(drehung)))
  const breite = Math.max(1, ziel.breite * cos + ziel.hoehe * sin)
  const hoehe = Math.max(1, ziel.breite * sin + ziel.hoehe * cos)

  return {
    mitteX: ziel.x + ziel.breite / 2,
    mitteY: ziel.y + ziel.hoehe / 2,
    massstab: begrenzt(
      Math.min(
        (flaeche.breite - 2 * rand) / breite,
        (flaeche.hoehe - 2 * rand) / hoehe,
      ),
    ),
    drehung,
  }
}

// aufAuswahl zielt auf die gewählte Karte — oder auf nichts.
//
// **Nichts ist die richtige Antwort, wo die Kennung zu keiner Lage gehört.**
// Die Auswahl kann weggeräumt worden sein; ein Sprung auf den Ursprung wäre
// eine Bewegung, die niemand angeordnet hat.
export function aufAuswahl(
  plan: Tischplan,
  gewaehlt: string | undefined,
  flaeche: Flaeche,
  drehung = 0,
): Kamera | undefined {
  if (gewaehlt === undefined) return undefined

  const lage = plan.lagen.find((eine) => eine.kennung === gewaehlt)

  return lage
    ? aufFlaeche(umfeld(lage, Auswahlumfeld), flaeche, Zielrand, drehung)
    : undefined
}

// umfeld gibt ein Rechteck um die Mitte eines anderen, um den Faktor größer.
function umfeld(ziel: Rechteck, faktor: number): Rechteck {
  const breite = ziel.breite * faktor
  const hoehe = ziel.hoehe * faktor

  return {
    x: ziel.x + ziel.breite / 2 - breite / 2,
    y: ziel.y + ziel.hoehe / 2 - hoehe / 2,
    breite,
    hoehe,
  }
}

// aufEigenenSitz zielt auf die eigene Sitzfläche — oder auf nichts.
//
// **Ein Zuschauer hat keinen eigenen Sitz**, und `sitzflaeche.ts` sagt es:
// `eigen` ist dann nirgends wahr. Der Knopf tut dann nichts, statt irgendwohin
// zu springen.
export function aufEigenenSitz(
  plan: Tischplan,
  flaeche: Flaeche,
  drehung = 0,
): Kamera | undefined {
  const eigene = plan.sitze.find((eine) => eine.eigen)

  return eigene ? aufFlaeche(eigene, flaeche, Zielrand, drehung) : undefined
}
