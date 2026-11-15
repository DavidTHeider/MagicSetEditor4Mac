// **Was um die Karten herum gezeichnet wird** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Eine eigene Datei
//
// **Der Schnitt folgt der Frage** (§2): `pixinaht.ts` beantwortet, *was* eine
// Karte ist und wie sie auf die Leinwand kommt — hier steht, was **um sie
// herum** liegt: ihr Rand, die Marken darauf, die Linie zu dem, woran sie
// hängt, und der Rahmen um die gewählte.
//
// **Und sie steht in `.covignore`**, wie die Nähte daneben: wo eine Marke
// liegt, rechnet `markenpunkte`; wohin eine Linie geht, `beziehungslinien` —
// beides wird gemessen. Hier wird gezeichnet.
import {
  beziehungslinien,
  kartenecken,
  markenpunkte,
  Markengroesse,
  rahmenUm,
} from './weltbild'
import { sitzschilder } from './sitzschilder'
import { haufenschilder } from './haufenschilder'
import { zonenschilder } from './zonenschilder'
import { wirkungsbild } from './tischwirkungen'
import { schriftgroesse } from './schildmass'

import type * as Pixi from 'pixi.js'
import type { Kartenbild } from './weltbild'
import type { Tischfarben } from './tischfarben'
import type { Tischplan } from './tischplan'
import type { Wirkungsart } from './tischwirkungen'
import type { Zonenschild } from './zonenschilder'

// Ebenen sind die Container, auf die das Beiwerk geht.
export interface Ebenen {
  beziehungen: Pixi.Container
  marken: Pixi.Container
  auswahl: Pixi.Container
}

// Beiwerk zeichnet Ränder, Linien, Marken und die Auswahl.
export interface Beiwerk {
  zeichne: (
    plan: Tischplan,
    bilder: Kartenbild[],
    gewaehlt: string | undefined,
    beschriftung: ((zone: string, kurz?: boolean) => string) | undefined,
    sitzkopf: ((sitz: string) => string) | undefined,

    // wirkungen ist, was seit dem letzten Bild geschehen ist — PLAY-04.AK-02
    // (#1712).
    wirkungen: Map<string, Wirkungsart> | undefined,

    // stufe ist der Wert von `tisch.beschriftung` (`schildmass.ts`).
    stufe: string,
  ) => void

  // faerbe nimmt eine neue Tafel — für einen Wechsel des Tischbilds.
  faerbe: (neue: Tischfarben) => void

  raeume: () => void
}

type Pixiteile = Pick<typeof Pixi, 'Graphics' | 'Text'>

// Zeichengroesse ist die Höhe eines Wirkungszeichens in Millimetern.
//
// **Kleiner als eine Marke**: es steht an der Ecke einer Karte und soll sie
// nicht verdecken. Die Zahl ist gewählt und nicht gemessen.
const Zeichengroesse = 4

export function neuesBeiwerk(
  pixi: Pixiteile,
  ebenen: Ebenen,
  anfangsfarben: Tischfarben,
): Beiwerk {
  let farben = anfangsfarben

  const raender = new pixi.Graphics()
  ebenen.marken.addChild(raender)

  const linien = new pixi.Graphics()
  ebenen.beziehungen.addChild(linien)

  const marken = new pixi.Graphics()
  ebenen.marken.addChild(marken)

  // **Die Zahl auf einer Marke ist Text, und Text ist teuer.** Ein `Text` je
  // Marke und Bild wäre eine Neuvermessung der Schrift bei jedem Zug; sie
  // werden deshalb gehalten und nur umgesetzt.
  const zahlen = new Map<string, InstanceType<typeof pixi.Text>>()

  const zeichneLinien = (plan: Tischplan) => {
    linien.clear()
    for (const linie of beziehungslinien(plan))
      linien
        .moveTo(linie.vonX, linie.vonY)
        .lineTo(linie.nachX, linie.nachY)
        .stroke({ color: farben.kante, width: 1.2, alpha: 0.8 })
  }

  // **Die Auswahl ist ein Rahmen und keine Färbung.** Eine getönte Karte
  // hieße, ihr Bild zu verändern — und wer zwei Karten nebeneinander
  // vergleicht, sähe zwei verschiedene Farben derselben Vorlage.
  const auswahl = new pixi.Graphics()
  ebenen.auswahl.addChild(auswahl)

  const zeichneAuswahl = (plan: Tischplan, gewaehlt: string | undefined) => {
    auswahl.clear()
    if (gewaehlt === undefined) return

    // **Um die Karte, wie sie zu sehen ist** — eine getappte liegt quer.
    const rahmen = rahmenUm(plan, gewaehlt, 2)
    if (!rahmen) return

    auswahl
      .rect(rahmen.x, rahmen.y, rahmen.breite, rahmen.hoehe)
      .stroke({ color: farben.auswahl, width: 2.5 })
  }

  const zeichneMarken = (plan: Tischplan) => {
    marken.clear()
    const gesehen = new Set<string>()

    for (const punkt of markenpunkte(plan))
      for (const marke of punkt.marken) {
        const schluessel = `${punkt.kennung}:${marke.art}`
        gesehen.add(schluessel)

        marken
          .circle(marke.x, marke.y, Markengroesse / 2)
          .fill({ color: farben.ruecken })
          .stroke({ color: farben.kante, width: 0.8 })

        let zahl = zahlen.get(schluessel)
        if (!zahl) {
          zahl = new pixi.Text({
            text: '',
            style: { fontSize: 24, fill: farben.kante },
          })
          zahl.anchor.set(0.5)
          zahl.scale.set(Markengroesse / 32)
          zahlen.set(schluessel, zahl)
          ebenen.marken.addChild(zahl)
        }

        zahl.text = String(marke.zahl)
        zahl.position.set(marke.x, marke.y)
      }

    for (const [schluessel, zahl] of zahlen)
      if (!gesehen.has(schluessel)) {
        zahl.destroy()
        zahlen.delete(schluessel)
      }
  }

  // **Die Schilder am Zonenfeld** — Name und Anzahl, PLAY-03.AK-06
  // (#1711).
  //
  // **Gehalten wie die Markenzahlen und aus demselben Grund**: ein `Text` je
  // Bild wäre eine Neuvermessung der Schrift bei jedem Zug.
  const schilder = new Map<string, InstanceType<typeof pixi.Text>>()

  // **Unter jeder Anzahl eine Marke** — wie `tzone-zahl` im Artboard. Auf
  // einem Kartenrücken wäre die Zahl sonst dunkel auf dunkel.
  const zahlmarken = new Map<string, InstanceType<typeof pixi.Graphics>>()

  const zeichneSchilder = (
    plan: Tischplan,
    beschriftung: ((zone: string, kurz?: boolean) => string) | undefined,
    sitzkopf: ((sitz: string) => string) | undefined,
    stufe: string,
  ) => {
    const gesehen = new Set<string>()

    // **Beide Listen gehen denselben Weg.** Ein Sitzkopf ist ein Schild wie
    // ein Zonenname — er steht nur woanders und ist größer.
    const alle = [
      ...(beschriftung ? zonenschilder(plan, beschriftung) : []),
      ...(sitzkopf ? sitzschilder(plan, sitzkopf) : []),

      // **Und die Zahl an jedem Haufen** — PLAY-03.AK-04, grafische Hälfte
      // (#1711).
      //
      // **Ohne Katalog**: eine Zahl ist in jeder Sprache dieselbe. Deshalb
      // hängt sie nicht an `beschriftung` wie die Zonennamen — sieben Wälder
      // sähen sonst aus wie einer, nur weil niemand einen Katalog mitgab.
      ...haufenschilder(plan),
    ]

    // **Ohne Wort bleibt es beim Strich.** Das ist der Zustand vor PLAY-03
    // und kein Fehler — wer keinen Katalog mitgibt, bekommt keinen erfundenen
    // Namen.
    for (const schild of alle) {
      // **Was bei dieser Stufe nicht dasteht, fällt unten weg** — wie ein
      // Schild, dessen Zone es nicht mehr gibt (`schildmass.ts`).
      const groesse = schriftgroesse(schild, stufe)
      if (groesse === undefined) continue

      gesehen.add(schild.kennung)

      let text = schilder.get(schild.kennung)

      if (!text) {
        text = new pixi.Text({ text: '', style: schildstil(schild, farben) })
        schilder.set(schild.kennung, text)
        ebenen.beziehungen.addChild(text)
      }

      text.anchor.set(schild.ankerX ?? 0, schild.ankerY)
      text.scale.set(groesse / 24)
      text.text =
        schild.art === 'name' || schild.art === 'kurzname'
          ? schild.text.toUpperCase()
          : schild.text
      text.position.set(schild.x, schild.y)

      if (schild.art === 'zahl') zeichneZahlmarke(schild, text, groesse)
    }

    for (const [kennung, text] of schilder)
      if (!gesehen.has(kennung)) {
        text.destroy()
        schilder.delete(kennung)
        zahlmarken.get(kennung)?.destroy()
        zahlmarken.delete(kennung)
      }
  }

  const zeichneZahlmarke = (
    schild: Zonenschild,
    text: InstanceType<typeof pixi.Text>,
    groesse: number,
  ) => {
    let marke = zahlmarken.get(schild.kennung)

    if (!marke) {
      marke = new pixi.Graphics()
      zahlmarken.set(schild.kennung, marke)
      ebenen.beziehungen.addChildAt(
        marke,
        ebenen.beziehungen.getChildIndex(text),
      )
    }

    const rand = groesse * 0.3
    const breite = text.width + 2 * rand
    const hoehe = text.height + rand * 0.6

    marke
      .clear()
      .roundRect(
        text.x - (schild.ankerX ?? 0) * text.width - rand,
        text.y - schild.ankerY * text.height - rand * 0.3,
        breite,
        hoehe,
        hoehe / 2,
      )
      .fill({ color: farben.grund })
      .stroke({ color: farben.kante, width: 0.8 })
  }

  // **Der Ring einer Wirkung** — PLAY-04.AK-02
  // (#1712).
  //
  // **Ein Ring und ein Zeichen, nicht nur eine Farbe** — WCAG 1.4.1 gilt auch
  // auf einer Leinwand. Farbe und Zeichen kommen aus derselben Tabelle wie in
  // der Zonenliste (`tischwirkungen.ts`); zwei Quellen für dasselbe Aussehen
  // liefen auseinander (§10).
  //
  // **Und er steht, statt zu blinken.** Eine Wirkung bleibt bis zum nächsten
  // Bild — wer gerade woanders hinsah, findet sie noch vor, und
  // `prefers-reduced-motion` nimmt niemandem eine Auskunft weg (#1460:
  // „abschaltbare Animationen", und eine abgeschaltete darf nichts
  // verschweigen).
  const ringe = new pixi.Graphics()
  ebenen.auswahl.addChild(ringe)

  const wirkzeichen = new Map<string, InstanceType<typeof pixi.Text>>()

  const zeichneWirkungen = (
    plan: Tischplan,
    wirkungen: Map<string, Wirkungsart> | undefined,
  ) => {
    ringe.clear()
    const gesehen = new Set<string>()

    for (const [kennung, art] of wirkungen ?? []) {
      const rahmen = rahmenUm(plan, kennung, 3)
      if (!rahmen) continue

      const bild = wirkungsbild(art)
      gesehen.add(kennung)

      ringe
        .rect(rahmen.x, rahmen.y, rahmen.breite, rahmen.hoehe)
        .stroke({ color: bild.farbe, width: 2 })

      let text = wirkzeichen.get(kennung)
      if (!text) {
        text = new pixi.Text({ text: bild.zeichen })
        wirkzeichen.set(kennung, text)
        ebenen.auswahl.addChild(text)
      }

      text.text = bild.zeichen
      text.style.fill = bild.farbe
      text.scale.set(Zeichengroesse / 24)
      text.position.set(rahmen.x, rahmen.y)
    }

    for (const [kennung, text] of wirkzeichen)
      if (!gesehen.has(kennung)) {
        text.destroy()
        wirkzeichen.delete(kennung)
      }
  }

  const zeichneRaender = (bilder: Kartenbild[]) => {
    raender.clear()
    for (const eines of bilder) {
      raender.setStrokeStyle({ color: farben.kante, width: 0.8 })
      raender.beginPath()

      // **Der Rand liegt, wie die Karte liegt** — gedreht um ihre Mitte
      // (`kartenecken`); ein Rechteck stünde um eine getappte hochkant.
      raender.poly(kartenecken(eines))
      raender.stroke()
    }
  }

  return {
    zeichne(plan, bilder, gewaehlt, beschriftung, sitzkopf, wirkungen, stufe) {
      zeichneSchilder(plan, beschriftung, sitzkopf, stufe)
      zeichneRaender(bilder)
      zeichneLinien(plan)
      zeichneMarken(plan)
      zeichneAuswahl(plan, gewaehlt)
      zeichneWirkungen(plan, wirkungen)
    },

    faerbe(neue) {
      farben = neue
      for (const zahl of zahlen.values()) zahl.style.fill = neue.kante
      for (const text of schilder.values()) text.style.fill = neue.schrift
    },

    raeume() {
      for (const text of wirkzeichen.values()) text.destroy()
      wirkzeichen.clear()
      zahlen.clear()
      schilder.clear()
      zahlmarken.clear()
    },
  }
}

// schildstil ist die Schrift eines Schildes — wie `tsitz-kopf` und
// `tzone-name` im Artboard *Tisch*: der Kopf halbfett, der Name gesperrt in
// Versalien, die Anzahl halbfett.
//
// **Die Schrift des Systems und nicht die von PixiJS**: dessen Vorgabe ist
// Arial, und neben der Oberfläche sah die Beschriftung darin fremd aus.
function schildstil(
  schild: Zonenschild,
  farben: Tischfarben,
): Partial<Pixi.TextStyleOptions> {
  return {
    fontFamily: 'system-ui, -apple-system, "Helvetica Neue", sans-serif',
    fontSize: 24,
    fill: farben.schrift,
    fontWeight:
      schild.art === 'name' || schild.art === 'kurzname' ? '500' : '600',

    // **Gesperrt nur der ganze Name** — über einem Stapel fehlte dafür der
    // Platz (`zonenschilder.ts`, [Kurzschildgroesse]).
    letterSpacing: schild.art === 'name' ? 2 : 0,
  }
}
