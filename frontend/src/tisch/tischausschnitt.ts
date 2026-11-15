// **Was von einem Tisch ausgegeben wird** — PLAY-05.AK-03 und AK-04
// (#1713).
//
// # Zwei Fragen, und beide ohne Leinwand zu beantworten
//
// **Welcher Weltausschnitt** — sichtbarer Bereich, ganzer Tisch oder eine
// Zone — und **wie viele Bildpunkte** daraus werden. Beides ist eine
// Rechnung; dass daraus ein Bild wird, ist die Sache von `pixinaht.ts`.
//
// **Der Schnitt ist derselbe wie zwischen `tischplan` und `pixinaht`**: hier
// die Entscheidung, dort der Ruf an die Grafikkarte. Er trägt hier genauso:
// ein Ausschnitt, den niemand zeichnen kann, ist trotzdem zu prüfen.
import type { Kamera } from './kamera'
import type { Tischplan } from './tischplan'

// Ausschnittart sind die drei, die AK-03 nennt.
export type Ausschnittart = 'sichtbar' | 'tisch' | 'zone'

// Ausschnittwahl ist, was jemand ausgeben will.
export interface Ausschnittwahl {
  art: Ausschnittart

  // zone und sitz meinen ein Feld — nur bei `art: 'zone'`.
  //
  // **Der Sitz gehört dazu**: jeder hat seinen eigenen Friedhof, und „der
  // Friedhof" allein meinte vier verschiedene. Der gemeinsame Stapel trägt
  // die leere Kennung (CR 405.1).
  zone?: string
  sitz?: string
}

// Rechteck ist ein Ausschnitt der Welt, in Millimetern.
export interface Rechteck {
  x: number
  y: number
  breite: number
  hoehe: number
}

// Ausgaberand ist, wie viel um einen Ausschnitt herum mitgeht.
//
// **Vier Millimeter**: eine Karte am Rand steht sonst angeschnitten da, und
// ein Rahmen mit halber Linie sieht aus wie ein Fehler. Die Zahl ist gewählt
// und nicht gemessen.
const Ausgaberand = 4

// ausschnittFuer gibt den Weltausschnitt zu einer Wahl.
//
// **`undefined` heißt: gibt es nicht** — eine Zone, die dieser Tisch nicht
// führt. Ein leeres Rechteck wäre ein weißes Bild, und das hielte jemand für
// einen leeren Tisch (§8).
export function ausschnittFuer(
  plan: Tischplan,
  kamera: Kamera,
  flaeche: { breite: number; hoehe: number },
  wahl: Ausschnittwahl,
): Rechteck | undefined {
  if (wahl.art === 'tisch') {
    return mitRand({ x: 0, y: 0, breite: plan.breite, hoehe: plan.hoehe })
  }

  if (wahl.art === 'zone') {
    const feld = plan.felder.find(
      (eines) => eines.zone === wahl.zone && eines.sitz === wahl.sitz,
    )

    return feld === undefined ? undefined : mitRand(feld)
  }

  return sichtbares(kamera, flaeche)
}

// sichtbares gibt, was gerade auf der Fläche steht — als umschließendes
// Rechteck.
//
// **Umschließend, weil ein gedrehter Blick ein gedrehtes Rechteck sieht.**
// Wer nur Breite und Höhe durch den Maßstab teilte, schnitte die Ecken ab,
// die man sieht. Dieselbe Rechnung wie in [eingepasst], nur andersherum.
function sichtbares(
  kamera: Kamera,
  flaeche: { breite: number; hoehe: number },
): Rechteck {
  const bogen = (kamera.drehung * Math.PI) / 180
  const cos = Math.abs(Math.cos(bogen))
  const sin = Math.abs(Math.sin(bogen))

  const breite = (flaeche.breite * cos + flaeche.hoehe * sin) / kamera.massstab
  const hoehe = (flaeche.breite * sin + flaeche.hoehe * cos) / kamera.massstab

  return {
    x: kamera.mitteX - breite / 2,
    y: kamera.mitteY - hoehe / 2,
    breite,
    hoehe,
  }
}

// mitRand legt den Ausgaberand um ein Rechteck.
function mitRand(rechteck: Rechteck): Rechteck {
  return {
    x: rechteck.x - Ausgaberand,
    y: rechteck.y - Ausgaberand,
    breite: rechteck.breite + 2 * Ausgaberand,
    hoehe: rechteck.hoehe + 2 * Ausgaberand,
  }
}

// Massfehler sagt, warum aus einem Ausschnitt kein Bild wird.
export type Massfehler = 'dpi' | 'zuGross' | 'leer'

// Bildmass ist die Größe eines Abzugs in Bildpunkten.
export interface Bildmass {
  breite: number
  hoehe: number
}

// DPIspanne ist, was zur Wahl steht.
//
// **72 bis 600** — dieselbe Spanne wie am Druckbogen, und aus demselben
// Grund: darunter ist es kein Druck mehr, darüber wächst nur die Datei. Die
// Zahlen sind gewählt und nicht gemessen.
const DPIniedrigstens = 72
const DPIhoechstens = 600

// Kantengrenze ist, wie viele Bildpunkte eine Kante höchstens hat.
//
// **8192** — vorsichtig unter dem, was eine Zeichenfläche in WebGL2 als
// Textur anlegt; verlorener Kontext ist der Preis, wenn es nicht reicht
// (#1460
// hat ihn schon einmal gekostet). Gewählt und nicht gemessen — wer sie hebt,
// misst an einem echten Fenster.
const Kantengrenze = 8192

// Flaechengrenze ist, wie viele Bildpunkte ein Abzug insgesamt hat.
//
// **Fünfzig Millionen** — rund 200 MB im Arbeitsspeicher, solange jeder
// Punkt vier Bytes trägt. Zwei Kanten von je 8000 kämen darüber, und ein
// Abzug, der den Rechner anhält, ist keiner.
const Flaechengrenze = 50_000_000

// Zoll ist, wie viele Millimeter ein Zoll hat — die Definition, keine Wahl.
const Zoll = 25.4

// bildmasse rechnet einen Ausschnitt in Bildpunkte — oder sagt, warum nicht.
export function bildmasse(
  rechteck: Rechteck,
  dpi: number,
): Bildmass | { fehler: Massfehler } {
  if (dpi < DPIniedrigstens || dpi > DPIhoechstens) {
    return { fehler: 'dpi' }
  }

  if (rechteck.breite <= 0 || rechteck.hoehe <= 0) {
    return { fehler: 'leer' }
  }

  const breite = Math.round((rechteck.breite / Zoll) * dpi)
  const hoehe = Math.round((rechteck.hoehe / Zoll) * dpi)

  if (
    breite > Kantengrenze ||
    hoehe > Kantengrenze ||
    breite * hoehe > Flaechengrenze
  ) {
    return { fehler: 'zuGross' }
  }

  return { breite, hoehe }
}
