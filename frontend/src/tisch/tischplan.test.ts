// **Der Tisch als Plan** — Werkstück G3 von Phase 5.5
// (#1460).
import { describe, expect, it } from 'vitest'

import { einObjekt, einePartie, einSitz } from '../seiten/partieprobe'
import { lage, meines, sitzflaeche, zweiSitze } from '../test/planprobe'

import { Tappdrehung } from './zonenlage'
import { mitZug, sitzFuer, tischplan } from './tischplan'

describe('sitzFuer', () => {
  // **CR 400.3** (Fassung 2026-08-19): was in eine fremde Bibliothek, einen
  // fremden Friedhof oder eine fremde Hand ginge, geht in die des Eigners.
  it('legt Bibliothek, Friedhof und Hand beim Eigner ab', () => {
    for (const zone of ['bibliothek', 'friedhof', 'hand'])
      expect(
        sitzFuer(einObjekt({ zone, eigner: 'du', beherrscher: 'ich' })),
      ).toBe('du')
  })

  // **CR 403.1**: Permanente liegen vor dem, der sie kontrolliert.
  it('stellt eine gestohlene Kreatur vor den Dieb', () => {
    expect(
      sitzFuer(einObjekt({ zone: 'tisch', eigner: 'du', beherrscher: 'ich' })),
    ).toBe('ich')
  })
})

describe('tischplan', () => {
  // **Fünf Felder je Sitz** — Tisch, Hand und drei Ablagen. Die
  // Kommandozone gehört der Variante und steht hier nicht (PLAY-03.AK-01,
  // #1711);
  // der Stapel gehört keinem Sitz.
  it('gibt jedem Sitz seine Fläche und seine fünf Zonen', () => {
    const plan = tischplan(einePartie(zweiSitze))

    expect(plan.sitze.map((eine) => eine.sitz).sort()).toEqual(['du', 'ich'])
    expect(plan.felder.filter((eines) => eines.sitz === 'ich')).toHaveLength(5)
    expect(plan.felder.map((eines) => eines.zone)).not.toContain('kommando')
  })

  // **Und sechs in einer Commanderpartie** — CR 903.3: dort steht der
  // Commander, und ein Feld dafür gehört auf den Tisch.
  it('gibt der Commanderpartie ihr Kommandofeld', () => {
    const plan = tischplan(
      einePartie({
        ...zweiSitze,
        regelwerk: {
          kennung: 'commander',
          fassung: 1,
          strenge: 'begleitet',
          abgewaehlt: [],
        },
      }),
    )

    expect(plan.felder.filter((eines) => eines.sitz === 'ich')).toHaveLength(6)
  })

  it('legt den Stapel zwischen die Sitze und keinem zu', () => {
    const plan = tischplan(einePartie(zweiSitze))
    const stapel = plan.felder.find((eines) => eines.zone === 'stapel')

    const meiner = sitzflaeche(plan, 'ich')
    const deiner = sitzflaeche(plan, 'du')

    expect(stapel?.sitz).toBe('')
    expect(stapel?.y).toBeGreaterThan(deiner.y + deiner.hoehe - 1)
    expect((stapel?.y ?? 0) + (stapel?.hoehe ?? 0)).toBeLessThan(meiner.y + 1)
  })

  it('legt den Stapel auch bei einem einzigen Sitz an', () => {
    const plan = tischplan(
      einePartie({
        sitze: [
          einSitz({
            kennung: 'ich',
            name: 'Ich',
            deckabdruck: 'a',
            karten: 60,
          }),
        ],
        fuer: 'ich',
      }),
    )

    expect(plan.felder.some((eines) => eines.zone === 'stapel')).toBe(true)
  })

  it('beginnt im Ursprung und misst sich selbst aus', () => {
    const plan = tischplan(einePartie(zweiSitze))

    expect(
      Math.min(...plan.felder.map((eines) => eines.x)),
    ).toBeGreaterThanOrEqual(0)
    expect(
      Math.min(...plan.felder.map((eines) => eines.y)),
    ).toBeGreaterThanOrEqual(0)
    for (const eines of plan.felder) {
      expect(eines.x + eines.breite).toBeLessThanOrEqual(plan.breite)
      expect(eines.y + eines.hoehe).toBeLessThanOrEqual(plan.hoehe)
    }
  })

  it('gibt jedem Objekt eine Lage in seiner Zone', () => {
    const plan = tischplan(
      einePartie({
        ...zweiSitze,
        objekte: [
          meines({ kennung: 'a', zone: 'tisch' }),
          einObjekt({
            kennung: 'b',
            zone: 'hand',
            eigner: 'du',
            beherrscher: 'du',
          }),
        ],
      }),
    )

    expect(lage(plan, 'a')).toMatchObject({ zone: 'tisch', sitz: 'ich' })
    expect(lage(plan, 'b')).toMatchObject({ zone: 'hand', sitz: 'du' })
  })

  it('ordnet eine offene Zone nach ihrer Ordnung', () => {
    const plan = tischplan(
      einePartie({
        ...zweiSitze,
        objekte: [
          meines({ kennung: 'spaet', zone: 'friedhof', ordnung: 2 }),
          meines({ kennung: 'frueh', zone: 'friedhof', ordnung: 1 }),
        ],
      }),
    )

    expect(lage(plan, 'spaet').hoehenrang).toBeGreaterThan(
      lage(plan, 'frueh').hoehenrang,
    )
  })

  it('dreht eine getappte Karte und lässt die andere stehen', () => {
    const plan = tischplan(
      einePartie({
        ...zweiSitze,
        objekte: [
          meines({ kennung: 'getappt', zone: 'tisch', getappt: true }),
          meines({ kennung: 'frei', zone: 'tisch' }),
        ],
      }),
    )

    expect(lage(plan, 'getappt').drehung).toBe(Tappdrehung)
    expect(lage(plan, 'frei').drehung).toBe(0)
  })

  it('macht aus einem geheimen Objekt eine verdeckte Lage', () => {
    const plan = tischplan(
      einePartie({
        ...zweiSitze,
        objekte: [
          einObjekt({
            kennung: 'fremd',
            zone: 'hand',
            eigner: 'du',
            beherrscher: 'du',
            geheim: true,
            name: '',
          }),
        ],
      }),
    )

    expect(lage(plan, 'fremd')).toMatchObject({ verdeckt: true, geheim: true })
  })

  it('zeigt die zweite Seite, wenn diese Sicht sie gewendet hat', () => {
    const partie = einePartie({
      ...zweiSitze,
      objekte: [meines({ kennung: 'dfc', zone: 'tisch' })],
    })

    expect(lage(tischplan(partie), 'dfc').seite).toBe('vorne')
    expect(lage(tischplan(partie, { gewendet: ['dfc'] }), 'dfc').seite).toBe(
      'zweite',
    )
  })

  it('zählt eine Bibliothek, statt sie zu zeichnen', () => {
    const plan = tischplan(
      einePartie({
        ...zweiSitze,
        objekte: Array.from({ length: 40 }, (_, stelle) =>
          meines({
            kennung: `b${String(stelle)}`,
            zone: 'bibliothek',
            verdeckt: true,
          }),
        ),
      }),
    )
    const zahl = plan.zahlen.find(
      (eine) => eine.zone === 'bibliothek' && eine.sitz === 'ich',
    )

    expect(zahl?.zahl).toBe(40)
    expect(
      plan.lagen.filter((eine) => eine.zone === 'bibliothek').length,
    ).toBeLessThan(40)
  })

  // **Nur die oberste Karte eines Stapels ist greifbar** — #2472: die
  // übrigen liegen 0,6 Punkte darunter, und als eigene Knöpfe waren sie
  // verdeckte Ziele (WCAG 2.2, 2.5.8). Gezeichnet werden sie weiter; die
  // Liste und „Hineinsehen …“ erreichen jede einzeln.
  it('legt in einem Stapel alle bis auf die oberste unter', () => {
    const plan = tischplan(
      einePartie({
        ...zweiSitze,
        objekte: [
          ...['b0', 'b1', 'b2'].map((kennung) =>
            meines({ kennung, zone: 'bibliothek', verdeckt: true }),
          ),
          ...['f0', 'f1'].map((kennung) =>
            meines({ kennung, zone: 'friedhof' }),
          ),
          meines({ kennung: 't0', zone: 'tisch' }),
          meines({ kennung: 't1', zone: 'tisch' }),
        ],
      }),
    )
    const oben = (zone: string) => {
      const lagen = plan.lagen.filter((eine) => eine.zone === zone)
      const hoechste = Math.max(...lagen.map((eine) => eine.hoehenrang))

      return {
        frei: lagen.filter((eine) => !eine.untergelegt).map((e) => e.kennung),
        oberste: lagen.find((eine) => eine.hoehenrang === hoechste)?.kennung,
      }
    }

    expect(oben('bibliothek').frei).toEqual([oben('bibliothek').oberste])
    expect(oben('friedhof').frei).toEqual([oben('friedhof').oberste])
    expect(oben('tisch').frei).toHaveLength(2)
  })

  // **CR 400.1** (Fassung 2026-08-19) kennt neben den sieben Zonen noch die
  // Ante-Zone älterer Karten. Dieses Programm kennt sie nicht — und genau so
  // ein Objekt darf nicht lautlos vom Tisch verschwinden (§8).
  it('führt ein Objekt in einer unbekannten Zone als Befund', () => {
    const plan = tischplan(
      einePartie({
        ...zweiSitze,
        objekte: [meines({ kennung: 'nirgends', zone: 'ante' })],
      }),
    )

    expect(plan.lagen).toHaveLength(0)
    expect(plan.heimatlos).toEqual(['nirgends'])
  })

  it('führt ein Objekt an einem unbekannten Sitz als Befund', () => {
    const plan = tischplan(
      einePartie({
        ...zweiSitze,
        objekte: [
          einObjekt({ kennung: 'fremd', zone: 'hand', eigner: 'niemand' }),
        ],
      }),
    )

    expect(plan.heimatlos).toEqual(['fremd'])
  })

  // **Eine Karte ohne Lage ist nicht heimatlos.** Ein Stapel zeichnet nur
  // seine obersten acht; die übrigen liegen sehr wohl in ihrem Feld und
  // erscheinen als Zahl. Der erste Anlauf zählte sie mit — eine Bibliothek
  // mit 53 Karten meldete 45 Befunde.
  it('meldet eine hohe Bibliothek nicht als Befund', () => {
    const plan = tischplan(
      einePartie({
        ...zweiSitze,
        objekte: Array.from({ length: 53 }, (_, stelle) =>
          meines({
            kennung: `b${String(stelle)}`,
            zone: 'bibliothek',
            ordnung: stelle,
            verdeckt: true,
          }),
        ),
      }),
    )

    expect(plan.heimatlos).toEqual([])
    expect(
      plan.lagen.filter((eine) => eine.zone === 'bibliothek').length,
    ).toBeLessThan(53)
  })

  it('lässt die Liste leer, wenn jedes Objekt seinen Platz hat', () => {
    const plan = tischplan(
      einePartie({ ...zweiSitze, objekte: [meines({ zone: 'tisch' })] }),
    )

    expect(plan.heimatlos).toEqual([])
  })

  // **Der Index einer Lage ist nicht der Index ihres Objekts.** Ein Stapel
  // gibt nur seine obersten Lagen her (`Stapelsicht`); wer sie über die
  // Stelle zuordnete, hängte den obersten acht die **untersten** acht Karten
  // an — und niemand sähe es, weil beide Listen gleich lang sind.
  it('nimmt für einen hohen Stapel die obersten Karten und nicht die ersten', () => {
    const plan = tischplan(
      einePartie({
        ...zweiSitze,
        objekte: Array.from({ length: 12 }, (_, stelle) =>
          meines({
            kennung: `f${String(stelle)}`,
            zone: 'friedhof',
            ordnung: stelle,
            name: `Karte ${String(stelle)}`,
            geheim: false,
          }),
        ),
      }),
    )

    expect(
      plan.lagen
        .filter((eine) => eine.zone === 'friedhof')
        .map((eine) => eine.name),
    ).toEqual([
      'Karte 4',
      'Karte 5',
      'Karte 6',
      'Karte 7',
      'Karte 8',
      'Karte 9',
      'Karte 10',
      'Karte 11',
    ])
  })

  it('kommt mit einer Partie ohne Sitze zurecht', () => {
    const plan = tischplan(einePartie({ sitze: [], objekte: [], fuer: '' }))

    expect(plan).toMatchObject({
      sitze: [],
      felder: [],
      lagen: [],
      zahlen: [],
      heimatlos: [],
    })
  })
})

describe('mitZug', () => {
  const einPlan = () =>
    tischplan(
      einePartie({
        ...zweiSitze,
        objekte: [
          meines({ kennung: 'a', zone: 'tisch', name: 'Bär' }),
          meines({ kennung: 'b', zone: 'tisch', name: 'Wolf' }),
        ],
      }),
    )

  it('verschiebt nur die gezogene Lage', () => {
    const plan = einPlan()
    const gezogen = mitZug(plan, { kennung: 'a', dx: 40, dy: -20 })

    expect(lage(gezogen, 'a').x).toBeCloseTo(lage(plan, 'a').x + 40, 6)
    expect(lage(gezogen, 'a').y).toBeCloseTo(lage(plan, 'a').y - 20, 6)
    expect(lage(gezogen, 'b')).toEqual(lage(plan, 'b'))
  })

  it('hebt die gezogene Lage über alle anderen', () => {
    const gezogen = mitZug(einPlan(), { kennung: 'a', dx: 0, dy: 0 })

    expect(lage(gezogen, 'a').hoehenrang).toBeGreaterThan(
      lage(gezogen, 'b').hoehenrang,
    )
  })

  it('gibt ohne Zug denselben Plan zurück', () => {
    const plan = einPlan()

    expect(mitZug(plan, undefined)).toBe(plan)
  })

  it('lässt einen Zug auf etwas Unbekanntes wirkungslos', () => {
    const plan = einPlan()

    expect(
      mitZug(plan, { kennung: 'gibtesnicht', dx: 99, dy: 99 }).lagen,
    ).toEqual(plan.lagen)
  })
})
