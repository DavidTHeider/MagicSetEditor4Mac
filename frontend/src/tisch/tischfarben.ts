// **Die Farben des Tisches** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Warum sie aus `stil.css` kommen und nicht hier stehen
//
// **Weil eine zweite Palette ein zweites Erscheinungsbild wäre.** Das Fenster
// schaltet zwischen hell und dunkel über die `--mse-*`-Merkmale; eine
// Farbtafel in TypeScript kennte diesen Schalter nicht und bliebe, wo sie
// war. Und sie stünde außerhalb der Reichweite des Kontrastwächters: der
// liest `stil.css` und die Klassennamen, nicht Hexzahlen im Quelltext (§10,
// „ein Wächter, der die halbe Zusage hält").
//
// **Gemessen sind sie**, und zwar jede gegen jeden Untergrund, auf dem sie
// wirklich liegt — `docs/KONTRASTE.md` führt die zehn Paare.
import type { Tischbild } from './tischbilder'

// Tischfarben sind die sechs, die der Zeichner braucht.
export interface Tischfarben {
  grund: number
  sitz: number
  eigenerSitz: number

  // kante ist die Linienfarbe — sie umrandet **jede** Fläche und jede Karte.
  //
  // **Daran hängt die Zugänglichkeit des ganzen Tisches.** Zwei Holztöne
  // nebeneinander tragen keine 3:1; eine Linie dazwischen schon, und WCAG
  // 1.4.11 lässt das ausdrücklich zu (§3: „gilt auch für Ränder ohne
  // Fläche").
  kante: number

  ruecken: number
  wartet: number

  // auswahl ist der Rahmen um die hervorgehobene Karte.
  //
  // **Ein Rahmen und keine Färbung**: eine getönte Karte hieße, ihr Bild zu
  // verändern, und wer zwei Karten derselben Vorlage nebeneinander legte,
  // sähe zwei verschiedene Farben.
  auswahl: number

  // schrift ist die Farbe von Zonenname und Kartenzahl (PLAY-03.AK-06).
  //
  // **Nicht die Kante**: die steht im hellen Bild bei 3,11 : 1 gegen die
  // Tischplatte. Für eine Linie reicht das (WCAG 1.4.11), für Text nicht
  // (1.4.3 verlangt 4,5:1). Der schlechteste Wert dieser Farbe ist 9,44 : 1,
  // nachgerechnet in `docs/KONTRASTE.md`.
  schrift: number
}

type Tafel = Tischfarben

// Tischersatz ist der helle Satz, wie er in `stil.css` steht.
//
// **Er ist der Rückfall, nicht die Wahrheit.** Wer ihn sieht, sieht einen
// Tisch ohne Merkmale — und die einzige Lage, in der das vorkommt, ist ein
// Dokument ohne `stil.css`.
export const Tischersatz: Tafel = {
  grund: 0xd4bb94,
  sitz: 0xdcc4a0,
  eigenerSitz: 0xe4cfae,
  kante: 0x786248,
  ruecken: 0xfdf9f0,
  wartet: 0xe0d0b0,
  auswahl: 0x2d1206,
  schrift: 0x2d1206,
}

const merkmale: [keyof Tafel, string][] = [
  ['grund', '--mse-tisch-grund'],
  ['sitz', '--mse-tisch-sitz'],
  ['eigenerSitz', '--mse-tisch-eigen'],
  ['kante', '--mse-tisch-kante'],
  ['ruecken', '--mse-tisch-ruecken'],
  ['wartet', '--mse-tisch-wartet'],
  ['auswahl', '--mse-tisch-auswahl'],
  ['schrift', '--mse-tisch-schrift'],
]

// dunklesBild sagt, ob das dunkle Erscheinungsbild gilt.
//
// **Gefragt wird der Grund und nicht das Merkmal `data-erscheinungsbild`** —
// das steht nur, wenn jemand ausdrücklich gewählt hat; folgt das Fenster der
// Systemeinstellung, steht es gar nicht da.
export function dunklesBild(lesen: (name: string) => string): boolean {
  const grund = farbzahl(lesen('--mse-grund'), Tischersatz.ruecken)

  return ((grund >> 16) & 0xff) + ((grund >> 8) & 0xff) + (grund & 0xff) < 384
}

const kurz = /^#([0-9a-f])([0-9a-f])([0-9a-f])$/i
const lang = /^#([0-9a-f]{6})$/i
const rgb = /^rgba?\(\s*(\d+)[\s,]+(\d+)[\s,]+(\d+)/i

// farbzahl rechnet eine CSS-Farbe in die Zahl um, die PixiJS nimmt.
//
// **Drei Formen, und die dritte ist der Ernstfall**: ein Browser gibt für ein
// aufgelöstes Merkmal `rgb(…)` zurück, nicht das, was in der Datei stand.
export function farbzahl(css: string, ersatz: number): number {
  const wert = css.trim()

  const voll = wert.replace(kurz, '#$1$1$2$2$3$3')
  const sechsstellig = lang.exec(voll)?.[1]
  if (sechsstellig !== undefined) return Number.parseInt(sechsstellig, 16)

  const kanaele = rgb.exec(wert)?.slice(1, 4).map(Number)
  if (kanaele) return kanaele.reduce((summe, kanal) => (summe << 8) | kanal, 0)

  return ersatz
}

export function tischfarben(lesen: (name: string) => string): Tafel {
  const tafel = { ...Tischersatz }

  for (const [feld, merkmal] of merkmale)
    tafel[feld] = farbzahl(lesen(merkmal), Tischersatz[feld])

  return tafel
}

// mitTischbild tauscht die drei Flächenfarben aus.
//
// **Nur die drei.** Kante, Rücken, Wartefarbe und Auswahl bleiben, wie sie in
// `stil.css` stehen — sie tragen die Zugänglichkeit, und ein Tischbild darf
// sie nicht anfassen. Dass jede Wahl sie trägt, rechnet
// `tischbilder.test.ts` nach.
export function mitTischbild(
  tafel: Tafel,
  bild: Tischbild,
  dunkel: boolean,
): Tafel {
  const satz = dunkel ? bild.dunkel : bild.hell

  return {
    ...tafel,
    grund: satz.grund,
    sitz: satz.sitz,
    eigenerSitz: satz.eigen,
  }
}

// ausDemDokument ist der Griff an das laufende Fenster.
export function ausDemDokument(dokument: Document) {
  return (name: string) =>
    dokument.defaultView
      ?.getComputedStyle(dokument.documentElement)
      .getPropertyValue(name)
      .trim() ?? ''
}
