// **Was von einem Tisch ausgegeben wird** — PLAY-05.AK-03 und AK-04
// (#1713).
//
// **Die Rechnung ohne Leinwand**: hier steht, *welcher Weltausschnitt* und
// *wie viele Bildpunkte* — nicht, wie das Bild entsteht.
import { describe, expect, it } from 'vitest'

import { ausschnittFuer, bildmasse } from './tischausschnitt'
import { einePartie } from '../seiten/partieprobe'
import { tischplan } from './tischplan'
import { zweiSitze } from '../test/planprobe'

const plan = tischplan(einePartie(zweiSitze))

const flaeche = { breite: 800, hoehe: 600 }

// ungedreht steht mittig über dem Plan, ein Punkt je Millimeter.
const ungedreht = {
  mitteX: plan.breite / 2,
  mitteY: plan.hoehe / 2,
  massstab: 1,
  drehung: 0,
}

describe('ausschnittFuer', () => {
  // **Der ganze Tisch ist der Plan** — und ein Rand darum, damit die Karten
  // am Rand nicht angeschnitten stehen.
  it('gibt für den ganzen Tisch den Plan samt Rand', () => {
    const aus = ausschnittFuer(plan, ungedreht, flaeche, { art: 'tisch' })

    expect(aus).toBeDefined()
    expect(aus?.breite).toBeGreaterThan(plan.breite)
    expect(aus?.hoehe).toBeGreaterThan(plan.hoehe)
    expect(aus?.x).toBeLessThan(0)
  })

  // **Der sichtbare Bereich hängt am Maßstab** — bei halber Vergrößerung
  // steht doppelt so viel Welt auf derselben Fläche.
  it('rechnet den sichtbaren Bereich aus Fläche und Maßstab', () => {
    const aus = ausschnittFuer(plan, { ...ungedreht, massstab: 2 }, flaeche, {
      art: 'sichtbar',
    })

    expect(aus).toEqual({
      x: plan.breite / 2 - 200,
      y: plan.hoehe / 2 - 150,
      breite: 400,
      hoehe: 300,
    })
  })

  // **Und an der Drehung.** Ein gedrehter Blick sieht ein gedrehtes Rechteck;
  // ausgegeben wird sein umschließendes — alles andere schnitte Ecken ab,
  // die man sieht.
  it('umschließt den gedrehten Blick', () => {
    const gerade = ausschnittFuer(plan, ungedreht, flaeche, { art: 'sichtbar' })
    const schraeg = ausschnittFuer(
      plan,
      { ...ungedreht, drehung: 90 },
      flaeche,
      { art: 'sichtbar' },
    )

    expect(schraeg?.breite).toBeCloseTo(gerade?.hoehe ?? 0)
    expect(schraeg?.hoehe).toBeCloseTo(gerade?.breite ?? 0)
  })

  // **Eine Zone ist ihr Feld** — samt Rand, aus demselben Grund wie beim
  // ganzen Tisch.
  it('gibt für eine Zone ihr Feld', () => {
    const feld = plan.felder.find(
      (eines) => eines.zone === 'tisch' && eines.sitz === 'ich',
    )

    const aus = ausschnittFuer(plan, ungedreht, flaeche, {
      art: 'zone',
      zone: 'tisch',
      sitz: 'ich',
    })

    expect(aus?.breite).toBeCloseTo((feld?.breite ?? 0) + 8)
    expect(aus?.x).toBeCloseTo((feld?.x ?? 0) - 4)
  })

  // **Eine Zone, die es nicht gibt, gibt nichts** — und keinen leeren
  // Ausschnitt, den jemand für einen weißen Tisch hielte (§8).
  it('gibt nichts für eine Zone, die dieser Tisch nicht hat', () => {
    expect(
      ausschnittFuer(plan, ungedreht, flaeche, {
        art: 'zone',
        zone: 'ante',
        sitz: 'ich',
      }),
    ).toBeUndefined()
  })

  // **Der gemeinsame Stapel gehört keinem Sitz** — CR 405.1, und er ist über
  // die leere Sitzkennung erreichbar.
  it('findet den gemeinsamen Stapel ohne Sitz', () => {
    expect(
      ausschnittFuer(plan, ungedreht, flaeche, {
        art: 'zone',
        zone: 'stapel',
        sitz: '',
      }),
    ).toBeDefined()
  })
})

describe('bildmasse', () => {
  const einhundert = { x: 0, y: 0, breite: 100, hoehe: 50 }

  // **Millimeter mal Punkte je Zoll durch 25,4** — die Zahl ist die
  // Definition des Zolls und keine Wahl.
  it('rechnet Millimeter in Bildpunkte', () => {
    expect(bildmasse(einhundert, 300)).toEqual({
      breite: 1181,
      hoehe: 591,
    })
  })

  // **Eine zu hohe Auflösung wird benannt und nicht gerechnet** — AK-04:
  // „DPI-Grenzen sind begrenzt und fehlerrobust".
  it('weist eine Auflösung außerhalb der Grenzen ab', () => {
    expect(bildmasse(einhundert, 1200)).toEqual({ fehler: 'dpi' })
    expect(bildmasse(einhundert, 50)).toEqual({ fehler: 'dpi' })
  })

  // **Und eine zu große Fläche ebenso.** Ein Tisch von zwei Metern bei 600
  // dpi wären 47.000 Punkte je Kante — keine Grafikkarte legt das an, und
  // ein Versuch endete als verlorener Kontext.
  it('weist eine zu große Fläche ab', () => {
    expect(bildmasse({ x: 0, y: 0, breite: 2000, hoehe: 2000 }, 600)).toEqual({
      fehler: 'zuGross',
    })
  })

  // **Ein Ausschnitt ohne Fläche ist kein Bild** — er entstünde, wo ein Feld
  // die Breite null hat.
  it('weist einen leeren Ausschnitt ab', () => {
    expect(bildmasse({ x: 0, y: 0, breite: 0, hoehe: 50 }, 300)).toEqual({
      fehler: 'leer',
    })
  })
})
