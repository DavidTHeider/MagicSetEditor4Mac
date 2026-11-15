// **Die zehn Szenen am Tischplan** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Wogegen hier geprüft wird
//
// **Dass jede der zehn einen Plan ergibt, der zusammenhält.** Kein Objekt
// heimatlos, keine Lage außerhalb ihres Feldes, keine Karte an zwei Stellen.
// Das ist die Regressionsprobe des Tischplans: eine Änderung an der Rechnung,
// die eine dieser Szenen zerbricht, wird hier laut.
//
// **Und dass die Liste vollständig ist.** #1460 nennt zehn Szenen beim Namen;
// eine kürzere Aufzählung wäre grün, und niemand sähe es (§10).
import { describe, expect, it } from 'vitest'

import { an } from '../test/erstes'
import { Kartenbreite, Kartenhoehe } from './masse'
import { Stapelsicht, Stapelversatz } from './zonenlage'
import { szene, szenennamen } from '../test/tischszenen'
import { tischplan } from './tischplan'

import type { Szenenname } from '../test/tischszenen'
import type { Tischplan } from './tischplan'

const plaene = new Map<Szenenname, Tischplan>(
  szenennamen.map((name) => [name, tischplan(szene(name))]),
)

const planVon = (name: Szenenname): Tischplan => {
  const einer = plaene.get(name)
  if (!einer) throw new Error(`keine Szene ${name}`)
  return einer
}

describe('die zehn Szenen', () => {
  it('sind zehn und heißen so, wie das Issue sie nennt', () => {
    expect(szenennamen).toEqual([
      'empty-board',
      'two-player-simple',
      'creature-combat',
      'large-land-piles',
      'stack-heavy',
      'tokens-and-counters',
      'four-player-commander',
      'extreme-board-state',
      'double-faced-cards',
      'many-attachments',
    ])
  })

  it('geben jedes Mal denselben Plan', () => {
    for (const name of szenennamen)
      expect(tischplan(szene(name))).toEqual(tischplan(szene(name)))
  })

  it.each(szenennamen)('%s lässt kein Objekt heimatlos', (name) => {
    expect(planVon(name).heimatlos).toEqual([])
  })

  // **Ein Stapel ragt über sein Feld, und zwar absichtlich**: der Versatz
  // zeigt seine Dicke. Mehr als `Stapelsicht × Stapelversatz` darf es nicht
  // sein — sonst wandert die unterste Karte in das Feld daneben.
  const stapelluft = Stapelsicht * Stapelversatz

  it.each(szenennamen)('%s hält jede Lage in ihrem Feld', (name) => {
    const plan = planVon(name)

    for (const lage of plan.lagen) {
      const feld = plan.felder.find(
        (eines) => eines.zone === lage.zone && eines.sitz === lage.sitz,
      )
      if (!feld) throw new Error(`kein Feld für ${lage.kennung}`)

      expect(lage.x).toBeGreaterThanOrEqual(feld.x - 0.000001)
      expect(lage.y).toBeGreaterThanOrEqual(feld.y - 0.000001)
      expect(lage.x + Kartenbreite).toBeLessThanOrEqual(
        feld.x + Math.max(feld.breite, Kartenbreite) + stapelluft,
      )
      expect(lage.y + Kartenhoehe).toBeLessThanOrEqual(
        feld.y + Math.max(feld.hoehe, Kartenhoehe) + stapelluft,
      )
    }
  })

  it.each(szenennamen)('%s gibt keiner Karte zwei Lagen', (name) => {
    const kennungen = planVon(name).lagen.map((eine) => eine.kennung)

    expect(new Set(kennungen).size).toBe(kennungen.length)
  })

  it.each(szenennamen)('%s bleibt im Ursprung und misst sich aus', (name) => {
    const plan = planVon(name)

    for (const feld of plan.felder) {
      expect(feld.x).toBeGreaterThanOrEqual(0)
      expect(feld.y).toBeGreaterThanOrEqual(0)
      expect(feld.x + feld.breite).toBeLessThanOrEqual(plan.breite + 0.000001)
      expect(feld.y + feld.hoehe).toBeLessThanOrEqual(plan.hoehe + 0.000001)
    }
  })
})

describe('was die einzelnen Szenen zeigen sollen', () => {
  it('empty-board hat Flächen, aber keine Karte', () => {
    const plan = planVon('empty-board')

    expect(plan.sitze).toHaveLength(2)
    expect(plan.lagen).toHaveLength(0)
    expect(plan.breite).toBeGreaterThan(0)
  })

  it('two-player-simple zeigt die eigene Hand offen und die fremde nicht', () => {
    const haende = planVon('two-player-simple').lagen.filter(
      (eine) => eine.zone === 'hand',
    )

    expect(
      haende.filter((eine) => eine.sitz === 'a').every((eine) => !eine.geheim),
    ).toBe(true)
    expect(
      haende.filter((eine) => eine.sitz === 'b').every((eine) => eine.geheim),
    ).toBe(true)
  })

  it('creature-combat dreht die Angreifer und lässt die Abwehr stehen', () => {
    const lagen = planVon('creature-combat').lagen

    expect(lagen.filter((eine) => eine.drehung !== 0)).toHaveLength(5)
    expect(
      lagen.filter((eine) => eine.sitz === 'b' && eine.drehung === 0),
    ).toHaveLength(4)
  })

  it('large-land-piles setzt enger, statt über den Rand zu laufen', () => {
    const plan = planVon('large-land-piles')
    const feld = plan.felder.find(
      (eines) => eines.zone === 'tisch' && eines.sitz === 'a',
    )
    const lagen = plan.lagen.filter(
      (eine) => eine.zone === 'tisch' && eine.sitz === 'a',
    )

    expect(lagen).toHaveLength(40)
    if (!feld) throw new Error('kein Schlachtfeld für a')

    for (const eine of lagen)
      expect(eine.x + Kartenbreite).toBeLessThanOrEqual(
        feld.x + feld.breite + 0.000001,
      )
  })

  // **CR 405.2** (Fassung 2026-08-19): jedes Objekt kommt oben auf den
  // Stapel — die Reihenfolge ist zu sehen, und der Stapel gehört keinem Sitz.
  it('stack-heavy fächert den Stapel auf und ordnet ihn', () => {
    const lagen = planVon('stack-heavy').lagen.filter(
      (eine) => eine.zone === 'stapel',
    )

    expect(lagen).toHaveLength(12)
    expect(lagen.every((eine) => eine.sitz === '')).toBe(true)
    for (let i = 1; i < lagen.length; i += 1)
      expect(an(lagen, i).hoehenrang).toBeGreaterThan(
        an(lagen, i - 1).hoehenrang,
      )
  })

  it('tokens-and-counters trägt jede Marke an ihre Karte', () => {
    const lagen = planVon('tokens-and-counters').lagen

    expect(
      lagen.filter((eine) => Object.keys(eine.zaehler).length > 0),
    ).toHaveLength(30)
  })

  it('four-player-commander gibt jedem Sitz seine Kommandozone', () => {
    const plan = planVon('four-player-commander')

    expect(plan.sitze).toHaveLength(4)
    expect(plan.lagen.filter((eine) => eine.zone === 'kommando')).toHaveLength(
      4,
    )
  })

  // **#1460 nennt die Leistungsszene mit Zahlen**: „4 Spielern, 150 Karten,
  // 100 Countern, 50 Tokens, 30 Beziehungsgrafiken". Sie werden hier gezählt
  // — eine Szene, die um ein Drittel danebenläge, machte jede Messung darauf
  // unvergleichbar.
  it('extreme-board-state trägt die Zahlen aus der Leistungsvorgabe', () => {
    const partie = szene('extreme-board-state')
    const aufDemTisch = partie.objekte.filter((eines) => eines.zone === 'tisch')

    expect(partie.sitze).toHaveLength(4)
    expect(
      aufDemTisch.filter(
        (eines) => eines.art === 'karte' && eines.haengtAn === '',
      ),
    ).toHaveLength(150)
    expect(
      aufDemTisch.filter((eines) => eines.art === 'spielstein'),
    ).toHaveLength(50)
    expect(
      aufDemTisch.filter((eines) => Object.keys(eines.zaehler).length > 0),
    ).toHaveLength(100)
    expect(aufDemTisch.filter((eines) => eines.haengtAn !== '')).toHaveLength(
      30,
    )
  })

  it('double-faced-cards kann auf die zweite Seite gewendet werden', () => {
    const gewendet = tischplan(szene('double-faced-cards'), {
      gewendet: ['zwei-3'],
    })

    expect(
      gewendet.lagen.find((eine) => eine.kennung === 'zwei-3')?.seite,
    ).toBe('zweite')
    expect(
      gewendet.lagen.filter((eine) => eine.seite === 'vorne'),
    ).toHaveLength(7)
  })

  // **CR 400.5**: was woran hängt, muss für alle klar bleiben.
  it('many-attachments führt jedes Anhängsel auf sein Ziel zurück', () => {
    const lagen = planVon('many-attachments').lagen
    const traeger = new Set(lagen.map((eine) => eine.kennung))
    const haengende = lagen.filter((eine) => eine.haengtAn !== '')

    expect(haengende).toHaveLength(18)
    for (const eine of haengende) expect(traeger.has(eine.haengtAn)).toBe(true)
  })
})
