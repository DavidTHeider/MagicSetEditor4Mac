// **Was der Zeichner setzen muss** — Werkstück G3 von Phase 5.5
// (#1460).
//
// **Die Probe ist die Kamera selbst.** Was hier als Transform herauskommt,
// muss dieselben Punkte treffen wie `nachPunkten` — sonst zeigt die
// semantische Schicht woanders hin als das Bild, und niemand merkt es, bis
// jemand mit der Tastatur bedient.
import { describe, expect, it } from 'vitest'

import { nachPunkten } from './kamera'
import { einObjekt, einePartie, einSitz } from '../seiten/partieprobe'
import { erstes } from '../test/erstes'
import {
  beziehungslinien,
  kartenbilder,
  markenpunkte,
  weltlage,
} from './weltbild'
import { tischplan } from './tischplan'
import { Tappdrehung } from './zonenlage'

import type { Kamera } from './kamera'

const flaeche = { breite: 900, hoehe: 700 }
const kamera: Kamera = { mitteX: 300, mitteY: 200, massstab: 1.5, drehung: 0 }

const planMit = (...teile: Parameters<typeof einObjekt>[0][]) =>
  tischplan(
    einePartie({
      sitze: [
        einSitz({ kennung: 'ich', name: 'Ich', deckabdruck: 'a', karten: 2 }),
        einSitz({ kennung: 'du', name: 'Du', deckabdruck: 'b', karten: 2 }),
      ],
      fuer: 'ich',
      // `einObjekt` gibt eine Karte in der Bibliothek und **geheim** — das
      // ist der Normalfall einer Probe und hier der Sonderfall.
      objekte: teile.map((eines) =>
        einObjekt({
          eigner: 'ich',
          beherrscher: 'ich',
          geheim: false,
          ...eines,
        }),
      ),
    }),
  )

describe('weltlage', () => {
  it('trifft dieselben Punkte wie die Kamera', () => {
    for (const drehung of [0, 45, 137]) {
      const welt = weltlage({ ...kamera, drehung }, flaeche)
      const ort = { x: 420, y: 90 }

      // Pixi rechnet: Bildschirm = Position + Drehung × Maßstab × (Ort − Pivot).
      const dx = (ort.x - welt.pivotX) * welt.skalierung
      const dy = (ort.y - welt.pivotY) * welt.skalierung
      const cos = Math.cos(welt.drehung)
      const sin = Math.sin(welt.drehung)
      const erwartet = nachPunkten({ ...kamera, drehung }, ort, flaeche)

      expect(welt.x + dx * cos - dy * sin).toBeCloseTo(erwartet.x, 6)
      expect(welt.y + dx * sin + dy * cos).toBeCloseTo(erwartet.y, 6)
    }
  })

  it('setzt den Ursprung in die Mitte der Fläche', () => {
    expect(weltlage(kamera, flaeche)).toMatchObject({ x: 450, y: 350 })
  })

  it('gibt die Drehung im Bogenmaß', () => {
    expect(weltlage({ ...kamera, drehung: 180 }, flaeche).drehung).toBeCloseTo(
      Math.PI,
      9,
    )
  })
})

describe('kartenbilder', () => {
  it('nimmt die Adresse, die die Bildquelle nennt', () => {
    const bild = erstes(
      kartenbilder(
        planMit({ kennung: 'a', zone: 'tisch', karteID: '7' }),
        (lage) => (lage.karteID === '7' ? '/spielbild/abc.png' : undefined),
      ),
    )

    expect(bild).toMatchObject({
      kennung: 'a',
      inhalt: { art: 'bild', adresse: '/spielbild/abc.png' },
    })
  })

  // **Der Rücken ist ein Bild und kein weißes Rechteck** — Befund (a) des
  // Entwurfs der Zieloberfläche, jetzt PLAY-03.AK-07
  // (#1711).
  //
  // Die Aufnahme `048-nativer-spieltisch.png` zeigt genau das: ein einzelnes
  // **weißes** Rechteck, wo eine Bibliothekskarte liegt — während die
  // Listenansicht daneben über `bausteine/Kartenbild.tsx` das echte
  // Rückenbild zeigt. Dieselbe Partie, zwei Wahrheiten.
  it('zeigt das Rückenbild, wo verdeckt liegt', () => {
    const bild = erstes(
      kartenbilder(
        planMit({ kennung: 'a', zone: 'tisch', verdeckt: true }),
        (_lage, verdeckt) => (verdeckt ? '/spielbild/ruecken.png' : undefined),
      ),
    )

    expect(bild.inhalt).toEqual({
      art: 'ruecken',
      adresse: '/spielbild/ruecken.png',
    })
  })

  it('zeigt das Rückenbild auch, wo diese Sicht nicht hineinsehen darf', () => {
    const bild = erstes(
      kartenbilder(
        planMit({ kennung: 'a', zone: 'hand', geheim: true }),
        (_lage, verdeckt) =>
          verdeckt ? '/spielbild/ruecken.png' : '/spielbild/verraeterisch.png',
      ),
    )

    expect(bild.inhalt).toEqual({
      art: 'ruecken',
      adresse: '/spielbild/ruecken.png',
    })
  })

  // **Und die Quelle erfährt, dass verdeckt gefragt wird** — das ist die
  // Zusage, die vorher daran hing, dass niemand fragt.
  //
  // Hier stand: *„Die Bildquelle wird gar nicht erst gefragt, wo nicht
  // hingesehen werden darf."* Das hielt die Geheimhaltung — und kostete das
  // Rückenbild. Gefragt wird jetzt, aber **nach dem Rücken**: die Quelle
  // bekommt den zweiten Parameter und baut damit gar keinen Wunsch nach einer
  // Kartenseite.
  it('sagt der Quelle, dass ein Rücken gemeint ist', () => {
    const gefragt: boolean[] = []

    kartenbilder(
      planMit({ kennung: 'a', zone: 'hand', geheim: true }),
      (_lage, verdeckt) => {
        gefragt.push(verdeckt)

        return undefined
      },
    )

    expect(gefragt).toEqual([true])
  })

  // **Ohne Rückenbild bleibt es verdeckt** — und das ist eine eigene Gestalt.
  //
  // Sie ist nicht `wartet`: dort ist ein **Gesicht** unterwegs, und der
  // Platzhalter sagt „gleich". Hier ist nichts unterwegs, was dieses Bild
  // ändern würde; die Karte liegt verdeckt, und mehr gibt es nicht zu sagen.
  it('bleibt verdeckt, solange kein Rückenbild da ist', () => {
    const bild = erstes(
      kartenbilder(
        planMit({ kennung: 'a', zone: 'tisch', verdeckt: true }),
        () => undefined,
      ),
    )

    expect(bild.inhalt).toEqual({ art: 'verdeckt' })
  })

  it('wartet, solange kein Bild da ist', () => {
    const bild = erstes(
      kartenbilder(
        planMit({ kennung: 'a', zone: 'tisch', karteID: '7' }),
        () => undefined,
      ),
    )

    expect(bild.inhalt).toEqual({ art: 'wartet' })
  })

  it('rechnet die Drehung ins Bogenmaß und setzt den Anker in die Mitte', () => {
    const bild = erstes(
      kartenbilder(
        planMit({ kennung: 'a', zone: 'tisch', getappt: true }),
        () => undefined,
      ),
    )
    const lage = erstes(
      planMit({ kennung: 'a', zone: 'tisch', getappt: true }).lagen,
    )

    expect(bild.drehung).toBeCloseTo((Tappdrehung * Math.PI) / 180, 9)
    expect(bild.x).toBeCloseTo(lage.x + lage.breite / 2, 6)
    expect(bild.y).toBeCloseTo(lage.y + lage.hoehe / 2, 6)
  })

  it('gibt für einen leeren Tisch nichts her', () => {
    expect(kartenbilder(planMit(), () => undefined)).toEqual([])
  })
})

describe('markenpunkte', () => {
  it('gibt jeder Markenart einen eigenen Punkt', () => {
    const karte = erstes(
      markenpunkte(
        planMit({
          kennung: 'a',
          zone: 'tisch',
          zaehler: { plus: 3, aufladung: 1 },
        }),
      ),
    )

    expect(karte.kennung).toBe('a')
    expect(karte.marken.map((eine) => eine.art)).toEqual(['aufladung', 'plus'])
    expect(karte.marken.map((eine) => eine.zahl)).toEqual([1, 3])
  })

  it('ordnet die Arten, damit sie nicht springen', () => {
    // **Ein `Record` hat keine Reihenfolge** — wer sie aus der Einfügung
    // nähme, ließe dieselben Marken bei jeder Antwort anders liegen.
    const erst = erstes(
      markenpunkte(
        planMit({ kennung: 'a', zone: 'tisch', zaehler: { zzz: 1, aaa: 2 } }),
      ),
    )
    const dann = erstes(
      markenpunkte(
        planMit({ kennung: 'a', zone: 'tisch', zaehler: { aaa: 2, zzz: 1 } }),
      ),
    )

    expect(erst.marken).toEqual(dann.marken)
    expect(erstes(erst.marken).art).toBe('aaa')
  })

  it('legt die Marken an den unteren Rand ihrer Karte', () => {
    const plan = planMit({ kennung: 'a', zone: 'tisch', zaehler: { plus: 1 } })
    const lage = erstes(plan.lagen)
    const marke = erstes(erstes(markenpunkte(plan)).marken)

    expect(marke.y).toBeGreaterThan(lage.y + lage.hoehe / 2)
    expect(marke.y).toBeLessThanOrEqual(lage.y + lage.hoehe)
  })

  it('geht an einer Karte ohne Marken vorbei', () => {
    expect(markenpunkte(planMit({ kennung: 'a', zone: 'tisch' }))).toEqual([])
  })

  it('lässt eine Marke mit der Zahl null weg', () => {
    // **Null Marken sind keine Marken** (CR 122.1, Fassung 2026-08-19): ein
    // Objekt mit null Zählern einer Art hat keine dieser Art.
    expect(
      markenpunkte(
        planMit({ kennung: 'a', zone: 'tisch', zaehler: { plus: 0 } }),
      ),
    ).toEqual([])
  })
})

describe('beziehungslinien', () => {
  it('zieht eine Linie vom Anhängsel zu seinem Ziel', () => {
    const plan = planMit(
      { kennung: 'ziel', zone: 'tisch', name: 'Bär' },
      { kennung: 'aura', zone: 'tisch', name: 'Segen', haengtAn: 'ziel' },
    )
    const linie = erstes(beziehungslinien(plan))
    const ziel = plan.lagen.find((eine) => eine.kennung === 'ziel')
    if (!ziel) throw new Error('kein Ziel im Plan')

    expect(linie.von).toBe('aura')
    expect(linie.nach).toBe('ziel')
    expect(linie.nachX).toBeCloseTo(ziel.x + ziel.breite / 2, 6)
  })

  it('lässt ein Anhängsel ohne auffindbares Ziel weg', () => {
    // **Es verschwindet trotzdem nicht** — die Karte selbst liegt da, nur
    // ihre Linie führt ins Leere und wird nicht gezogen.
    const plan = planMit({
      kennung: 'aura',
      zone: 'tisch',
      haengtAn: 'gibtesnicht',
    })

    expect(beziehungslinien(plan)).toEqual([])
    expect(plan.lagen).toHaveLength(1)
  })

  it('gibt für einen Tisch ohne Anhängsel nichts her', () => {
    expect(beziehungslinien(planMit({ kennung: 'a', zone: 'tisch' }))).toEqual(
      [],
    )
  })
})
