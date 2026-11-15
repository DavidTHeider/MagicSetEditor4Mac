// **Die Kamera über dem Tisch** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Sie gehört diesem Fenster und sonst niemandem
//
// **Ein Schwenk ist keine Spielhandlung.** Wer den Tisch dreht, ändert nichts
// am Spiel — die Kamera geht deshalb nie an den Wirt und steht in keinem
// Ereignis. Sie ist der Grund, warum der Plan in Millimetern rechnet: was auf
// dem Tisch liegt, ist für alle gleich; wie groß es *erscheint*, ist es nicht.

// Kamera ist der Blick auf den Tisch.
export interface Kamera {
  // mitteX und mitteY sind der Punkt in Millimetern, der in der Mitte der
  // Fläche steht.
  mitteX: number
  mitteY: number

  // massstab ist, wie viele Punkte ein Millimeter ist.
  massstab: number

  // drehung ist der Winkel in Grad, um den die Fläche gedreht erscheint.
  drehung: number
}

// Flaeche ist die Zeichenfläche in Punkten.
export interface Flaeche {
  breite: number
  hoehe: number
}

export interface Punkt {
  x: number
  y: number
}

// Kleinstmass und Groesstmass begrenzen den Maßstab.
//
// **Unter 0,2 ist eine Karte 13 Punkte breit** — darunter ist nichts mehr zu
// erkennen, und die Grafikkarte zeichnete Sprites für niemanden. **Über 8**
// füllt eine Karte 504 Punkte; wer mehr will, schlägt sie auf.
export const Kleinstmass = 0.2
export const Groesstmass = 8

// imBogen rechnet Grad in Bogenmaß.
//
// **Exportiert seit PLAY-05.AK-05**: `kamerazielen.ts` rechnet dieselbe
// Drehung in dieselbe Einpassung ein, und zwei Umrechnungen nebeneinander
// liefen auseinander, sobald jemand eine von beiden anfasst.
export const imBogen = (grad: number) => (grad * Math.PI) / 180

// eingepasst legt den ganzen Plan in die Fläche.
//
// **Die Drehung bleibt, wie sie war.** Sie ist eine eigene Entscheidung des
// Bedienenden; ein Einpassen, das sie zurückdrehte, nähme ihm die Sicht, die
// er gerade gewählt hat.
//
// **Und sie geht in die Rechnung ein.** Ein gedrehtes Rechteck braucht mehr
// Platz als ein aufrechtes: |b·cos φ| + |h·sin φ| in der Breite. Ohne diesen
// Term liefe der Tisch bei jeder Drehung über den Rand — und „passt den
// ganzen Tisch ein" wäre eine Zusage, die nur bei null Grad hält (§10).
export function eingepasst(
  plan: { breite: number; hoehe: number },
  flaeche: Flaeche,
  rand = 0,
  drehung = 0,
): Kamera {
  const cos = Math.abs(Math.cos(imBogen(drehung)))
  const sin = Math.abs(Math.sin(imBogen(drehung)))
  const breite = Math.max(1, plan.breite * cos + plan.hoehe * sin)
  const hoehe = Math.max(1, plan.breite * sin + plan.hoehe * cos)

  return {
    mitteX: plan.breite / 2,
    mitteY: plan.hoehe / 2,
    massstab: begrenzt(
      Math.min(
        (flaeche.breite - 2 * rand) / breite,
        (flaeche.hoehe - 2 * rand) / hoehe,
      ),
    ),
    drehung,
  }
}

// **Warum `kamerazielen.aufFlaeche` daneben steht und nicht hier**: jene legt
// ein **beliebiges** Rechteck in die Fläche und braucht dafür den Plan; diese
// legt den Plan hinein und braucht seine Ecke nicht, weil sie im Ursprung
// liegt. Die Rechnung ist dieselbe — und sie steht zweimal da, was ein
// Widerspruch werden kann. **Der Lauf dagegen steht in
// `kamerazielen.test.ts`**: er hält die eine gegen die andere.

export function begrenzt(massstab: number): number {
  return Math.min(Groesstmass, Math.max(Kleinstmass, massstab))
}

// nachPunkten rechnet einen Ort auf dem Tisch in einen Ort auf der Fläche um.
export function nachPunkten(
  kamera: Kamera,
  ort: Punkt,
  flaeche: Flaeche,
): Punkt {
  const dx = (ort.x - kamera.mitteX) * kamera.massstab
  const dy = (ort.y - kamera.mitteY) * kamera.massstab
  const cos = Math.cos(imBogen(kamera.drehung))
  const sin = Math.sin(imBogen(kamera.drehung))

  return {
    x: flaeche.breite / 2 + dx * cos - dy * sin,
    y: flaeche.hoehe / 2 + dx * sin + dy * cos,
  }
}

// nachMillimetern ist der Weg zurück — und er wird gebraucht, sobald jemand
// auf die Fläche zeigt.
export function nachMillimetern(
  kamera: Kamera,
  punkt: Punkt,
  flaeche: Flaeche,
): Punkt {
  const px = punkt.x - flaeche.breite / 2
  const py = punkt.y - flaeche.hoehe / 2
  const cos = Math.cos(imBogen(kamera.drehung))
  const sin = Math.sin(imBogen(kamera.drehung))

  return {
    x: kamera.mitteX + (px * cos + py * sin) / kamera.massstab,
    y: kamera.mitteY + (-px * sin + py * cos) / kamera.massstab,
  }
}

// verschoben schwenkt um eine Strecke in Punkten.
//
// **Die Drehung geht mit ein**: wer den Tisch gedreht hat und nach rechts
// zieht, will nach rechts auf dem *Bildschirm*, nicht nach rechts auf dem
// Tisch.
export function verschoben(kamera: Kamera, dx: number, dy: number): Kamera {
  const cos = Math.cos(imBogen(kamera.drehung))
  const sin = Math.sin(imBogen(kamera.drehung))

  return {
    ...kamera,
    mitteX: kamera.mitteX - (dx * cos + dy * sin) / kamera.massstab,
    mitteY: kamera.mitteY - (-dx * sin + dy * cos) / kamera.massstab,
  }
}

// gezoomt ändert den Maßstab und lässt den Ort unter dem Zeiger stehen.
export function gezoomt(
  kamera: Kamera,
  faktor: number,
  aufPunkt: Punkt,
  flaeche: Flaeche,
): Kamera {
  const vorher = nachMillimetern(kamera, aufPunkt, flaeche)
  const naeher = { ...kamera, massstab: begrenzt(kamera.massstab * faktor) }
  const nachher = nachMillimetern(naeher, aufPunkt, flaeche)

  return {
    ...naeher,
    mitteX: naeher.mitteX + (vorher.x - nachher.x),
    mitteY: naeher.mitteY + (vorher.y - nachher.y),
  }
}

// gedreht dreht den Blick und hält den Winkel im Kreis.
export function gedreht(kamera: Kamera, grad: number): Kamera {
  return { ...kamera, drehung: (((kamera.drehung + grad) % 360) + 360) % 360 }
}
