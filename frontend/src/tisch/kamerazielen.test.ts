// **Auf die Auswahl und auf den eigenen Sitz** — PLAY-05.AK-05
// (#1713).
//
// `kamera.ts` kannte Schwenken, Zoomen und Drehen — **beides davon nicht**,
// und `Tischsteuerung.tsx` bot entsprechend nur Alles zeigen, Näher heran,
// Weiter weg und Tisch drehen. Wer eine Karte in einer Partie mit vier Sitzen
// suchte, musste sie mit der Maus finden.
//
// **Eine eigene Datei**, weil `kamera.test.ts` mit diesen Läufen über die
// Zeilengrenze aus §2 ginge. Der Schnitt folgt der Frage: dort, **wie** die
// Kamera sich bewegt — hier, **worauf** sie zielt.
import { describe, expect, it } from 'vitest'

import { aufAuswahl, aufEigenenSitz, aufFlaeche } from './kamerazielen'
import { einObjekt, einePartie, einSitz } from '../seiten/partieprobe'
import { eingepasst, Groesstmass } from './kamera'
import { tischplan } from './tischplan'

const flaeche = { breite: 1000, hoehe: 800 }

const einPlan = (...teile: Parameters<typeof einObjekt>[0][]) =>
  tischplan(
    einePartie({
      sitze: [
        einSitz({ kennung: 'ich', name: 'Ich', deckabdruck: 'a', karten: 2 }),
        einSitz({ kennung: 'du', name: 'Du', deckabdruck: 'b', karten: 2 }),
      ],
      fuer: 'ich',
      objekte: teile.map((eines) =>
        einObjekt({ eigner: 'ich', beherrscher: 'ich', ...eines }),
      ),
    }),
  )

describe('aufFlaeche', () => {
  it('setzt die Mitte auf die Mitte des Ziels', () => {
    const kamera = aufFlaeche(
      { x: 100, y: 200, breite: 60, hoehe: 80 },
      flaeche,
      0,
      0,
    )

    expect(kamera.mitteX).toBe(130)
    expect(kamera.mitteY).toBe(240)
  })

  it('passt das Ziel in die Fläche, Rand inbegriffen', () => {
    const kamera = aufFlaeche(
      { x: 0, y: 0, breite: 500, hoehe: 400 },
      flaeche,
      100,
      0,
    )

    // 800 Punkte Breite für 500, 600 Punkte Höhe für 400 — die Höhe ist enger.
    expect(kamera.massstab).toBeCloseTo(1.5, 6)
  })

  // **Die Drehung bleibt und geht in die Rechnung ein** — dieselbe Überlegung
  // wie bei `eingepasst`: ein gedrehtes Rechteck braucht mehr Platz.
  it('behält die Drehung und rechnet mit ihr', () => {
    const gerade = aufFlaeche(
      { x: 0, y: 0, breite: 400, hoehe: 100 },
      flaeche,
      0,
      0,
    )
    const quer = aufFlaeche(
      { x: 0, y: 0, breite: 400, hoehe: 100 },
      flaeche,
      0,
      90,
    )

    expect(quer.drehung).toBe(90)
    expect(quer.massstab).toBeLessThan(gerade.massstab)
  })

  // **Ein Ziel ohne Ausdehnung sprengt den Maßstab nicht.** Eine Division
  // durch null gäbe `Infinity`, und `begrenzt` fängt es — der Lauf hält fest,
  // dass es gefangen **wird**, statt dass jemand es annimmt.
  it('bleibt an der Grenze, wo das Ziel keine Größe hat', () => {
    expect(
      aufFlaeche({ x: 10, y: 10, breite: 0, hoehe: 0 }, flaeche, 0, 0).massstab,
    ).toBe(Groesstmass)
  })
})

// **Dieselbe Rechnung, zweimal geschrieben — und ein Lauf dagegen.**
//
// `eingepasst` legt den **ganzen** Plan in die Fläche, `aufFlaeche` ein
// beliebiges Rechteck. Die Rechnung ist dieselbe; sie zusammenzulegen ginge
// nur über einen Kreis zwischen den zwei Dateien. Also steht sie zweimal da —
// **und dieser Lauf hält sie gleich**. Ohne ihn wäre der Satz im Kommentar von
// `kamera.ts` eine Zusage, die niemand einlöst (§10).
describe('aufFlaeche und eingepasst', () => {
  it('geben für den ganzen Plan dasselbe', () => {
    const plan = einPlan({ kennung: 'a', zone: 'tisch' })

    for (const drehung of [0, 37, 90, 180, 270]) {
      for (const rand of [0, 16, 120]) {
        expect(
          aufFlaeche(
            { x: 0, y: 0, breite: plan.breite, hoehe: plan.hoehe },
            flaeche,
            rand,
            drehung,
          ),
        ).toEqual(eingepasst(plan, flaeche, rand, drehung))
      }
    }
  })
})

describe('aufAuswahl', () => {
  it('zielt auf die Lage der gewählten Karte', () => {
    const plan = einPlan({ kennung: 'a', zone: 'tisch' })
    const lage = plan.lagen.find((eine) => eine.kennung === 'a')

    if (!lage) throw new Error('keine Lage')

    const kamera = aufAuswahl(plan, 'a', flaeche)

    expect(kamera?.mitteX).toBeCloseTo(lage.x + lage.breite / 2, 6)
    expect(kamera?.mitteY).toBeCloseTo(lage.y + lage.hoehe / 2, 6)
  })

  // **Näher als der ganze Tisch, und das ist der Sinn.**
  it('kommt näher heran als das Einpassen', () => {
    const plan = einPlan({ kennung: 'a', zone: 'tisch' })
    const nah = aufAuswahl(plan, 'a', flaeche)
    const weit = aufFlaeche(
      { x: 0, y: 0, breite: plan.breite, hoehe: plan.hoehe },
      flaeche,
      0,
      0,
    )

    expect(nah?.massstab).toBeGreaterThan(weit.massstab)
  })

  // **Eine Karte, die es nicht gibt, bewegt nichts.** Sie zu ignorieren ist
  // richtig: die Auswahl kann weggeräumt worden sein, und ein Sprung auf den
  // Ursprung wäre eine Bewegung, die niemand angeordnet hat.
  it('gibt nichts her, wo die Kennung zu keiner Lage gehört', () => {
    expect(aufAuswahl(einPlan(), 'gibtesnicht', flaeche)).toBeUndefined()
  })

  it('gibt nichts her, wo gar nichts gewählt ist', () => {
    expect(aufAuswahl(einPlan(), undefined, flaeche)).toBeUndefined()
  })
})

describe('aufEigenenSitz', () => {
  it('zielt auf die eigene Sitzfläche', () => {
    const plan = einPlan()
    const eigene = plan.sitze.find((eine) => eine.eigen)

    if (!eigene) throw new Error('kein eigener Sitz')

    const kamera = aufEigenenSitz(plan, flaeche)

    expect(kamera?.mitteX).toBeCloseTo(eigene.x + eigene.breite / 2, 6)
    expect(kamera?.mitteY).toBeCloseTo(eigene.y + eigene.hoehe / 2, 6)
  })

  // **Ein Zuschauer hat keinen eigenen Sitz** — `sitzflaeche.ts` sagt es, und
  // der Knopf darf dann nichts tun statt irgendwohin zu springen.
  it('gibt nichts her, wo niemand einen eigenen Sitz hat', () => {
    const plan = einPlan()

    for (const sitz of plan.sitze) sitz.eigen = false

    expect(aufEigenenSitz(plan, flaeche)).toBeUndefined()
  })
})
