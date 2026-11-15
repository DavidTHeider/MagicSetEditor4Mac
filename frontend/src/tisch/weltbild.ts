// **Was der Zeichner setzen muss** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Die letzte Rechnung vor der Naht
//
// **Alles, was hier steht, könnte in der Grafiknaht stehen — und stünde dort
// ungeprüft.** Der Anker in der Mitte, das Bogenmaß, die Entscheidung
// zwischen Bild, Rücken und Warten: das sind Regeln, keine Grafikaufrufe. Was
// danach kommt, ist ein Sprite an eine Stelle setzen, und dafür gibt es
// nichts mehr zu entscheiden.
import type { Bildquelle } from './maler'
import type { Flaeche, Kamera } from './kamera'
import type { Tischlage, Tischplan } from './tischplan'
import { schluesselvergleich } from '../ordnung'
import { umriss } from './zonenlage'

import type { Rechteck } from './zonenlage'

const imBogen = (grad: number) => (grad * Math.PI) / 180

// Weltlage ist, wie der Behälter zu setzen ist, der den ganzen Tisch trägt.
//
// **Die Gestalt kommt von PixiJS**: Position, Pivot, Maßstab und Drehung
// eines Containers. Die Zahlen kommen von der Kamera — und dass beides
// dieselben Punkte trifft, prüft der Lauf daneben gegen `nachPunkten`.
export interface Weltlage {
  x: number
  y: number
  pivotX: number
  pivotY: number
  skalierung: number

  // drehung im **Bogenmaß** — anders als im Plan, der Grad führt, weil er
  // gelesen wird.
  drehung: number
}

export function weltlage(kamera: Kamera, flaeche: Flaeche): Weltlage {
  return {
    x: flaeche.breite / 2,
    y: flaeche.hoehe / 2,
    pivotX: kamera.mitteX,
    pivotY: kamera.mitteY,
    skalierung: kamera.massstab,
    drehung: imBogen(kamera.drehung),
  }
}

// Kartenbildinhalt sagt, was auf einer Karte zu sehen ist.
//
//   - `bild` — die Textur ist da und trägt ihre Adresse.
//   - `ruecken` — es **darf** keine geben: die Karte liegt verdeckt, oder
//     diese Sicht sieht nicht hinein.
//   - `wartet` — es **gibt** noch keine: sie ist unterwegs.
//
// # Warum drei Gestalten und nicht ein Feld mit Lücken
//
// **Dieselbe Überlegung wie bei `Texturwunsch` in G2.** Eine Gestalt mit
// `adresse?` und `platzhalter` hätte an jeder Stelle ein Feld, das die Hälfte
// der Zeit nichts bedeutet — und ein Rücken mit gesetzter Adresse wäre
// darstellbar, obwohl er nie vorkommen darf.
export type Kartenbildinhalt =
  | { art: 'bild'; adresse: string }
  | { art: 'ruecken'; adresse: string }
  | { art: 'verdeckt' }
  | { art: 'wartet' }

// Kartenbild ist ein Sprite, wie der Zeichner es braucht.
export interface Kartenbild {
  kennung: string

  // x und y sind der **Mittelpunkt** — Karten drehen sich um ihre Mitte
  // (CR 701.26a), und ein Anker an der Ecke ließe eine getappte Karte
  // wegwandern.
  x: number
  y: number

  breite: number
  hoehe: number

  drehung: number
  hoehenrang: number

  inhalt: Kartenbildinhalt
}

export function kartenbilder(plan: Tischplan, bild: Bildquelle): Kartenbild[] {
  return plan.lagen.map((lage) => ({
    kennung: lage.kennung,
    x: lage.x + lage.breite / 2,
    y: lage.y + lage.hoehe / 2,
    breite: lage.breite,
    hoehe: lage.hoehe,
    drehung: imBogen(lage.drehung),
    hoehenrang: lage.hoehenrang,
    inhalt: inhaltFuer(lage, bild),
  }))
}

// kartenecken gibt die vier Ecken einer Karte, wie sie liegt — gedreht um
// ihre Mitte, im Uhrzeigersinn ab der linken oberen.
//
// **Und nicht das Rechteck der aufrechten Karte**: eine getappte liegt quer,
// und ein Rechteck stünde hochkant um sie herum. Vier Ecken statt eines
// Rechtecks, damit der Rand auch während der Drehung an der Karte bleibt.
export function kartenecken(
  bild: Pick<Kartenbild, 'x' | 'y' | 'breite' | 'hoehe' | 'drehung'>,
): { x: number; y: number }[] {
  const cos = Math.cos(bild.drehung)
  const sin = Math.sin(bild.drehung)

  return [
    [-1, -1],
    [1, -1],
    [1, 1],
    [-1, 1],
  ].map(([seitX = 0, seitY = 0]) => {
    const dx = (seitX * bild.breite) / 2
    const dy = (seitY * bild.hoehe) / 2

    return { x: bild.x + dx * cos - dy * sin, y: bild.y + dx * sin + dy * cos }
  })
}

// inhaltFuer entscheidet zwischen Gesicht, Rücken, Verdeckt und Warten.
//
// # Die Geheimhaltung hängt an der Frage, nicht am Ergebnis
//
// Die Bildquelle gar nicht erst zu fragen, wo nicht hingesehen werden darf,
// hielte die Zusage — und kostete das Rückenbild: der Zeichner bekäme
// `{ art: 'ruecken' }` ohne Adresse und malte ein **weißes Rechteck**,
// während die Listenansicht daneben über `bausteine/Kartenbild.tsx` den
// echten Rücken zeigt. Dieselbe Partie, zwei Wahrheiten.
//
// **Gefragt wird deshalb, aber nach dem Rücken.** Die Quelle bekommt mit, dass
// verdeckt gemeint ist, und baut daraufhin gar keinen Wunsch nach einer
// Kartenseite — sie fragt nach dem Rücken des **Spiels**, der keiner Karte
// gehört. Die Zusage ist damit nicht schwächer, sondern steht an einer
// Stelle, die man prüfen kann: `kartenbilder`s Lauf zählt nach, mit welchem
// zweiten Parameter gefragt wurde.
//
// # Und warum `verdeckt` eine eigene Gestalt ist
//
// **Weil sie nicht `wartet` ist.** Dort ist ein Gesicht unterwegs, und der
// Platzhalter sagt „gleich". Hier ist nichts unterwegs, was dieses Bild
// ändern würde — die Karte liegt verdeckt, und mehr gibt es nicht zu sagen.
// Sie in `wartet` zu falten hieße, dem Zeichner zwei verschiedene Lagen als
// dieselbe zu melden.
function inhaltFuer(lage: Tischlage, bild: Bildquelle): Kartenbildinhalt {
  const verdeckt = lage.verdeckt || lage.geheim
  const adresse = bild(lage, verdeckt)

  if (verdeckt) {
    return adresse ? { art: 'ruecken', adresse } : { art: 'verdeckt' }
  }

  return adresse ? { art: 'bild', adresse } : { art: 'wartet' }
}

// Marke ist ein Zähler an seinem Platz.
//
// **CR 122.1** (Fassung 2026-08-19): eine Marke liegt auf einem Objekt oder
// einem Spieler und beeinflusst es. Wie viele es sind, steht an ihr; **was
// sie bewirkt, weiß dieses Programm nicht** — die Grenze steht so im
// Grundprofil (Werkstück G4).
export interface Marke {
  art: string
  zahl: number
  x: number
  y: number
}

// Markenpunkt sind die Marken einer Karte.
export interface Markenpunkt {
  kennung: string
  marken: Marke[]
}

// Markengroesse ist der Durchmesser einer Marke in Millimetern.
export const Markengroesse = 11

// markenpunkte legt die Marken an den unteren Rand ihrer Karte.
//
// **Unten und nicht oben**, weil oben der Name steht — auf jeder Vorlage, die
// dieses Projekt kennt. Eine Marke über dem Namen verdeckte genau das, was
// die Karte kenntlich macht.
//
// **Nach Art sortiert**, weil ein `Record` keine Reihenfolge hat: wer sie aus
// der Einfügung nähme, ließe dieselben Marken bei jeder Antwort anders liegen.
export function markenpunkte(plan: Tischplan): Markenpunkt[] {
  const punkte: Markenpunkt[] = []

  for (const lage of plan.lagen) {
    // **Null Marken sind keine Marken** (CR 122.1): ein Objekt mit null
    // Zählern einer Art hat keine dieser Art.
    const arten = Object.entries(lage.zaehler)
      .filter(([, zahl]) => zahl > 0)
      .sort(([eine], [andere]) => schluesselvergleich(eine, andere))

    if (arten.length === 0) continue

    // **Am sichtbaren Rand** — eine getappte Karte liegt quer, und ihr
    // unterer Rand ist nicht der der aufrechten Lage (`umriss`).
    const sichtbar = umriss(lage)
    const schritt = Markengroesse + 1
    const anfang =
      sichtbar.x + sichtbar.breite / 2 - ((arten.length - 1) * schritt) / 2

    punkte.push({
      kennung: lage.kennung,
      marken: arten.map(([art, zahl], stelle) => ({
        art,
        zahl,
        x: anfang + stelle * schritt,
        y: sichtbar.y + sichtbar.hoehe - Markengroesse / 2 - 2,
      })),
    })
  }

  return punkte
}

// rahmenUm gibt das Rechteck um eine Karte, wie sie zu sehen ist — um `rand`
// größer, für die Auswahl und die Wirkungen. **Nichts**, wo die Kennung zu
// keiner Lage gehört: die Karte kann weggeräumt worden sein.
export function rahmenUm(
  plan: Tischplan,
  kennung: string,
  rand: number,
): Rechteck | undefined {
  const lage = plan.lagen.find((eine) => eine.kennung === kennung)
  if (!lage) return undefined

  const sichtbar = umriss(lage)

  return {
    x: sichtbar.x - rand,
    y: sichtbar.y - rand,
    breite: sichtbar.breite + 2 * rand,
    hoehe: sichtbar.hoehe + 2 * rand,
  }
}

// Beziehungslinie ist die Verbindung von einem Anhängsel zu seinem Ziel.
//
// **CR 400.5** (Fassung 2026-08-19) verlangt sie: für alle Spieler muss
// erkennbar bleiben, was an welchem Objekt hängt. Eine Aura, die
// irgendwo auf dem Schlachtfeld läge, ohne dass man sähe, woran sie hängt,
// bräche die Regel — und zwar für alle am Tisch.
export interface Beziehungslinie {
  von: string
  nach: string
  vonX: number
  vonY: number
  nachX: number
  nachY: number
}

export function beziehungslinien(plan: Tischplan): Beziehungslinie[] {
  const mitte = new Map(
    plan.lagen.map((eine) => [
      eine.kennung,
      { x: eine.x + eine.breite / 2, y: eine.y + eine.hoehe / 2 },
    ]),
  )

  const linien: Beziehungslinie[] = []

  for (const lage of plan.lagen) {
    if (lage.haengtAn === '') continue

    const ziel = mitte.get(lage.haengtAn)
    const eigene = mitte.get(lage.kennung)

    // **Ein Ziel, das nicht auf dem Tisch liegt, bekommt keine Linie** — die
    // Karte selbst bleibt trotzdem stehen. Das kommt vor: ein Anhängsel kann
    // an etwas hängen, das diese Sicht nicht führt.
    if (!ziel || !eigene) continue

    linien.push({
      von: lage.kennung,
      nach: lage.haengtAn,
      vonX: eigene.x,
      vonY: eigene.y,
      nachX: ziel.x,
      nachY: ziel.y,
    })
  }

  return linien
}
