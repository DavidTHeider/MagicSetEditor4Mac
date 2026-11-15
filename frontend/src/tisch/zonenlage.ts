// **Wo ein Objekt innerhalb seiner Zone liegt** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Linksbündig, und das ist eine Entscheidung
//
// **Weil eine neue Karte die anderen liegen lassen soll.** Zentriert sähe
// ruhiger aus, verschöbe aber bei jeder Ziehkarte die ganze Hand — und wer
// gerade auf eine Karte zeigt, zeigt danach auf eine andere. Die Regel sagt
// dazu nichts (CR 400.5 erlaubt jede Anordnung); die Stabilität entscheidet.
//
// # Ein hoher Stapel bekommt nicht hundert Lagen
//
// **Weil man in eine Bibliothek nicht hineinsieht** — CR 401.2 (Fassung
// 2026-08-19) nennt sie einen einzelnen verdeckten Stapel, dessen Reihenfolge
// niemand ansehen darf. Was unsichtbar und unerreichbar ist, braucht keine
// Lage; es braucht eine **Zahl**. Achtzig Sprites für achtzig ungesehene
// Karten wären der Preis für nichts.
import { Kartenabstand, Kartenbreite, Kartenhoehe } from './masse'

import type { Zonenfeld } from './zonenfeld'

// Tappdrehung ist der Winkel einer getappten Permanente, in Grad.
//
// **CR 701.26a** (Fassung 2026-08-19): tappen heißt, sie aus der aufrechten
// Lage seitwärts zu drehen. Seitwärts ist ein Viertelkreis.
//
// **Grad und nicht Bogenmaß**, weil der Plan gelesen wird: 90 sagt jedem, was
// es ist. Die Naht zum Zeichner rechnet um.
export const Tappdrehung = 90

// Stapelsicht ist, wie viele Karten eines Stapels eine eigene Lage bekommen.
//
// **Acht reichen für die Dicke.** Mehr sieht niemand, und jede weitere wäre
// ein Sprite ohne Betrachter.
export const Stapelsicht = 8

// Stapelversatz ist, um wie viel eine tieferliegende Karte hervorschaut.
export const Stapelversatz = 0.6

// Haufenversatz ist, um wie viel eine Karte im Haufen hervorschaut.
//
// **Größer als der Stapelversatz.** In einem Stapel zeigt der Versatz die
// **Dicke**; in einem Haufen soll jede Karte greifbar bleiben — vier
// Millimeter sind ein Rand, den ein Zeiger trifft, und bei acht Karten
// nehmen sie weniger als eine halbe Kartenbreite.
export const Haufenversatz = 4

// Gelegtes ist, was eine Lage braucht — und nicht mehr.
//
// **Nicht das ganze Spielobjekt**, damit diese Rechnung ohne eine Partie
// prüfbar bleibt.
export interface Gelegtes {
  kennung: string
  getappt: boolean
  verdeckt: boolean

  // haufen ist die Anordnung, in der dieses Objekt liegt — **leer heißt:
  // kein Haufen** (#1756).
  //
  // **Eine Anordnung und keine Zone** (CR 400.5): sieben Wälder übereinander
  // sind keine Unterzone des Schlachtfelds. Was in einem Haufen liegt,
  // belegt **einen** Platz — und behält seine eigene Lage darin, sonst wäre
  // die unterste Karte nicht mehr auflösbar (PLAY-03.AK-03).
  haufen: string
}

// Lage ist der Platz eines Objekts in Millimetern.
export interface Lage {
  kennung: string

  x: number
  y: number
  breite: number
  hoehe: number

  // drehung ist der Winkel in Grad, um den Mittelpunkt.
  drehung: number

  // hoehenrang sagt, was über was liegt — größer ist weiter oben.
  hoehenrang: number

  verdeckt: boolean
}

// Rechteck ist eine Fläche auf dem Tisch, in Millimetern.
export interface Rechteck {
  x: number
  y: number
  breite: number
  hoehe: number
}

// umriss gibt die Fläche, die eine Lage **sichtbar** belegt.
//
// **Eine Lage beschreibt die aufrechte Karte**, und gedreht wird um ihre
// Mitte — so will es der Zeichner (`weltbild.ts`). Quer ist eine Karte aber
// so breit, wie sie aufrecht hoch ist (CR 701.26a). **Wer wissen will, wo
// eine Karte zu sehen ist, fragt hier** und nicht die Lage: der Zeiger, der
// Auswahlrahmen, die Marken, die Kamera. Bis zum 2026-09-24 fragte jeder die
// Lage, und bei einer getappten Karte lag alles daneben.
export function umriss(lage: Rechteck & { drehung: number }): Rechteck {
  if (lage.drehung % 180 === 0) {
    return { x: lage.x, y: lage.y, breite: lage.breite, hoehe: lage.hoehe }
  }

  const bogen = (lage.drehung * Math.PI) / 180
  const cos = Math.abs(Math.cos(bogen))
  const sin = Math.abs(Math.sin(bogen))
  const breite = lage.breite * cos + lage.hoehe * sin
  const hoehe = lage.breite * sin + lage.hoehe * cos

  return {
    x: lage.x + lage.breite / 2 - breite / 2,
    y: lage.y + lage.hoehe / 2 - hoehe / 2,
    breite,
    hoehe,
  }
}

// Gelegte Lage ist eine Lage samt dem, woraus sie entstand.
//
// **Sie trägt es mit, statt dass der Aufrufer über die Stelle zuordnet.** Ein
// Stapel gibt nur seine obersten Lagen her; wer sie über den Index zurück auf
// die Liste abbildete, hängte ihnen die **untersten** an — gleich lange
// Listen, falsche Paare, und kein Typ, der es merkte.
export type GelegteLage<T> = Lage & { gelegtes: T }

// Zonenlagen sind die sichtbaren Lagen einer Zone und ihre wahre Zahl.
export interface Zonenlagen<T> {
  lagen: GelegteLage<T>[]

  // zahl ist, wie viele Objekte wirklich in der Zone liegen.
  //
  // **Sie kann größer sein als `lagen`** — und genau dann sagt eine Zahl am
  // Stapel, wie viele es sind.
  zahl: number
}

export function zonenlagen<T extends Gelegtes>(
  feld: Zonenfeld,
  gelegte: T[],
): Zonenlagen<T> {
  return feld.gestalt === 'stapel'
    ? stapellagen(feld, gelegte)
    : reihenlagen(feld, gelegte)
}

// reihenlagen deckt `reihe` und `faecher` mit derselben Rechnung ab: ein
// Fächerfeld ist eine Kartenhöhe hoch und bekommt dadurch genau eine Reihe.
//
// # Ein Platz ist so breit, wie er zu sehen ist
//
// **Eine getappte Karte liegt quer** und braucht die Breite einer
// Kartenhöhe. Bis zum 2026-09-24 bekam jeder Platz die Breite einer
// aufrechten Karte: zwei getappte überdeckten einander, und die erste ragte
// über den linken Rand ihres Feldes (WebKit, Set „Green").
function reihenlagen<T extends Gelegtes>(
  feld: Zonenfeld,
  gelegte: T[],
): Zonenlagen<T> {
  const reihen = Math.max(
    1,
    Math.floor((feld.hoehe + Kartenabstand) / (Kartenhoehe + Kartenabstand)),
  )

  // **Gerechnet wird über Plätze und nicht über Karten** — ein Haufen
  // belegt einen (#1756).
  // Ohne diesen Schritt lägen fünf Wälder nebeneinander und trügen nur
  // denselben Namen.
  const plaetze = haufenweise(gelegte)

  // **Erst umbrechen, dann enger setzen.** Eine zweite Reihe ist billiger zu
  // lesen als eine Überlappung; erst wenn auch die Reihen voll sind, rücken
  // die Karten zusammen.
  const proReihe =
    reihenbreite(plaetze.map(platzbreite)) <= feld.breite
      ? Math.max(1, plaetze.length)
      : Math.ceil(plaetze.length / reihen)

  const lagen: GelegteLage<T>[] = []

  for (let anfang = 0; anfang < plaetze.length; anfang += proReihe) {
    const y = feld.y + (anfang / proReihe) * (Kartenhoehe + Kartenabstand)

    for (const { platz, x } of reihenorte(
      feld,
      plaetze.slice(anfang, anfang + proReihe),
    )) {
      for (const [tiefe, eines] of platz.entries()) {
        const drehung = eines.getappt ? Tappdrehung : 0

        // **Die Lage beschreibt die aufrechte Karte**, gedreht wird um ihre
        // Mitte; deshalb rückt eine quere um den Überstand nach innen.
        const ueberstand = eines.getappt ? (Kartenhoehe - Kartenbreite) / 2 : 0

        lagen.push({
          kennung: eines.kennung,

          // **Innerhalb des Haufens schaut jede Karte hervor.** Die zuerst
          // gelegte liegt unten; wer später dazukam, liegt darüber — sonst
          // sähe derselbe Haufen bei jedem Bild anders aus.
          x: x + ueberstand + tiefe * Haufenversatz,
          y: y + tiefe * Haufenversatz,
          breite: Kartenbreite,
          hoehe: Kartenhoehe,
          drehung,

          // **Der Rang zählt über die Plätze hinweg weiter**, damit zwei
          // Karten nie denselben tragen und der Zeichner nicht nach Zufall
          // entscheidet.
          hoehenrang: lagen.length,
          verdeckt: eines.verdeckt,
          gelegtes: eines,
        })
      }
    }
  }

  return { zahl: gelegte.length, lagen }
}

// platzbreite ist, wie breit ein Platz zu sehen ist: die breiteste Karte
// darin, und jede weitere schaut um den Haufenversatz hervor.
function platzbreite(platz: Gelegtes[]): number {
  const breiteste = Math.max(
    ...platz.map((eines) => (eines.getappt ? Kartenhoehe : Kartenbreite)),
  )

  return breiteste + (platz.length - 1) * Haufenversatz
}

// reihenbreite ist, wie breit Plätze nebeneinander mit vollem Abstand sind.
function reihenbreite(breiten: number[]): number {
  return breiten.reduce(
    (summe, breite, stelle) =>
      summe + breite + (stelle > 0 ? Kartenabstand : 0),
    0,
  )
}

// reihenorte gibt jedem Platz einer Reihe seinen linken Rand.
//
// **Passt die Reihe nicht, rückt sie gleichmäßig zusammen** — so weit, dass
// der letzte Platz am rechten Rand endet. Eine Karte über den Rand zu setzen
// hieße, sie in ein Feld zu legen, in dem sie nicht liegt. **Ein einzelner
// Platz, der breiter ist als sein Feld, kann nicht zusammenrücken**; er
// beginnt am linken Rand.
function reihenorte<T extends Gelegtes>(
  feld: Zonenfeld,
  reihe: T[][],
): { platz: T[]; x: number }[] {
  let versatz = 0

  const gelegt = reihe.map((platz) => {
    const ort = versatz
    versatz += platzbreite(platz) + Kartenabstand

    return { platz, ort }
  })

  const gesamt = versatz - Kartenabstand
  const letzter = Math.max(0, ...gelegt.map((eines) => eines.ort))
  // **`Math.max` fängt den einzelnen Platz**: ohne einen zweiten gibt es
  // nichts, worum er zusammenrücken könnte, der Faktor wird null, und er
  // bleibt am linken Rand.
  const enger =
    gesamt <= feld.breite
      ? 1
      : Math.max(0, (feld.breite - (gesamt - letzter)) / letzter)

  return gelegt.map(({ platz, ort }) => ({ platz, x: feld.x + ort * enger }))
}

// haufenweise fasst zusammen, was auf einem Haufen liegt.
//
// **Die Reihenfolge bleibt die der Liste.** Der erste Platz gehört dem
// ersten Objekt, und wer später auf einen schon offenen Haufen kommt, reiht
// sich dort ein — **ohne** den Platz nach vorn zu ziehen. Eine Sortierung
// hier wäre eine zweite Ordnung über einer Zone, die nach CR 400.5 keine
// hat.
function haufenweise<T extends Gelegtes>(gelegte: T[]): T[][] {
  const plaetze: T[][] = []
  const stelleVon = new Map<string, number>()

  for (const eines of gelegte) {
    const schon = eines.haufen === '' ? undefined : stelleVon.get(eines.haufen)

    if (schon === undefined) {
      if (eines.haufen !== '') {
        stelleVon.set(eines.haufen, plaetze.length)
      }

      plaetze.push([eines])

      continue
    }

    // **Der Platz steht fest**, und der Zugriff ist sicher: `stelleVon`
    // trägt nur Stellen, die vorher gefüllt wurden.
    plaetze[schon]?.push(eines)
  }

  return plaetze
}

// stapellagen legt die obersten übereinander. Die oberste liegt exakt auf dem
// Feld — gleich wie hoch der Stapel ist, damit sie nicht wandert, wenn eine
// Karte gezogen wird.
//
// # Hier wird nichts gedreht, und das ist eine Regel
//
// **CR 110.1** (Fassung 2026-08-19): eine Permanente ist eine Karte oder ein
// Token **auf dem Schlachtfeld**. Getappt sein können nur Permanenten
// (CR 701.26a) — in Bibliothek, Friedhof, Exil und Kommando gibt es das
// nicht. Ein `getappt ? … : …` an dieser Stelle wäre ein Zweig, den kein Lauf
// je beträte, und ein unbetretener Zweig ist entweder ein fehlender Test oder
// toter Code (§3). Hier ist er das zweite.
function stapellagen<T extends Gelegtes>(
  feld: Zonenfeld,
  gelegte: T[],
): Zonenlagen<T> {
  const sichtbar = gelegte.slice(-Stapelsicht)

  return {
    zahl: gelegte.length,
    lagen: sichtbar.map((eines, stelle) => {
      const tiefe = sichtbar.length - 1 - stelle

      return {
        kennung: eines.kennung,
        x: feld.x + tiefe * Stapelversatz,
        y: feld.y + tiefe * Stapelversatz,
        breite: Kartenbreite,
        hoehe: Kartenhoehe,
        drehung: 0,
        hoehenrang: stelle,
        verdeckt: eines.verdeckt,
        gelegtes: eines,
      }
    }),
  }
}
