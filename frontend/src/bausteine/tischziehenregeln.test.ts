// **Was beim Ziehen geschieht** — Werkstück G3 von Phase 5.5
// (#1460),
// herausgelöst aus `Tischflaeche.tsx` (§2).
//
// **Der Lauf daneben (`tischziehen.test.tsx`) zieht am Baum**; hier steht die
// Rechnung dahinter, ohne einen Baum.
import { describe, expect, it } from 'vitest'

import { angefasst, geschoben, losgelassen } from './tischziehen'
import { einObjekt, einePartie } from '../seiten/partieprobe'
import { tischplan } from '../tisch/tischplan'
import { zweiSitze } from '../test/planprobe'

// aufDemTisch gibt einen Plan mit den genannten Karten auf dem Schlachtfeld.
const aufDemTisch = (karten: { kennung: string; haufen?: string }[]) =>
  tischplan(
    einePartie({
      ...zweiSitze,
      objekte: karten.map((karte) =>
        einObjekt({
          kennung: karte.kennung,
          zone: 'tisch',
          eigner: 'ich',
          beherrscher: 'ich',
          haufen: karte.haufen ?? '',
        }),
      ),
    }),
  )

const plan = aufDemTisch([{ kennung: 'k' }])

const laufend = { kennung: 'k', vonX: 10, vonY: 20, dx: 0, dy: 0 }

describe('angefasst', () => {
  it('fängt einen Zug an der Hauptzeigertaste an', () => {
    expect(angefasst(0, 'k', { x: 10, y: 20 })).toEqual(laufend)
  })

  // **Die rechte Taste zieht nicht.** Sie öffnet anderswo ein Menü, und ein
  // Zug daran wäre eine Überraschung.
  it('zieht an keiner anderen Taste', () => {
    expect(angefasst(2, 'k', { x: 10, y: 20 })).toBeUndefined()
  })

  it('zieht nicht am leeren Tisch', () => {
    expect(angefasst(0, undefined, { x: 10, y: 20 })).toBeUndefined()
  })
})

describe('geschoben', () => {
  it('rechnet den Versatz vom Anfangspunkt', () => {
    expect(geschoben(laufend, { x: 15, y: 50 })).toEqual({
      ...laufend,
      dx: 5,
      dy: 30,
    })
  })
})

describe('losgelassen', () => {
  const lage = plan.lagen.find((eine) => eine.kennung === 'k')

  it('meldet den Zug in eine andere Zone', () => {
    expect(
      losgelassen(plan, laufend, { zone: 'friedhof', sitz: 'ich' }),
    ).toEqual({ art: 'zone', kennung: 'k', zone: 'friedhof', sitz: 'ich' })
  })

  // **Ein Zug innerhalb derselben Zone ist eine Anordnung und keine
  // Bewegung** — CR 400.5 lässt sie frei, und der Wirt hätte nichts zu
  // entscheiden.
  it('meldet nichts, wo die Karte liegen bleibt', () => {
    expect(
      losgelassen(plan, laufend, {
        zone: lage?.zone ?? '',
        sitz: lage?.sitz ?? '',
      }),
    ).toBeUndefined()
  })

  // **Neben den Feldern liegt Tisch**, und ein Wurf dorthin ist kein Zug.
  it('meldet nichts ohne Ziel', () => {
    expect(losgelassen(plan, laufend, undefined)).toBeUndefined()
  })

  // **Und nichts für ein Objekt, das der Plan nicht führt.** Es kann
  // verschwinden, während jemand zieht — ein anderer Prozess schreibt auf
  // denselben Bestand (#156).
  it('meldet nichts für ein Objekt ohne Lage', () => {
    expect(
      losgelassen(
        plan,
        { ...laufend, kennung: 'weg' },
        {
          zone: 'friedhof',
          sitz: 'ich',
        },
      ),
    ).toBeUndefined()
  })
})

// **Ziehen auf eine andere Karte legt einen Haufen** — CR 400.5,
// #1756.
//
// # Warum diese Geste und keine zweite
//
// **Weil der Zug innerhalb einer Zone bisher nichts tat.** CR 400.5 lässt
// die Anordnung frei, und genau das ist ein Haufen — die Geste dafür ist,
// eine Karte auf eine andere zu legen. Die Handlungsleiste kann dasselbe;
// eine andere Bedienung ist keine andere Fachhandlung (#1460).
describe('losgelassen und der Haufen', () => {
  const zwei = aufDemTisch([{ kennung: 'k' }, { kennung: 'z' }])

  const inEinemHaufen = aufDemTisch([
    { kennung: 'k', haufen: 'z' },
    { kennung: 'z', haufen: 'z' },
  ])

  it('legt eine Karte auf die, auf der sie landet', () => {
    expect(
      losgelassen(zwei, laufend, { zone: 'tisch', sitz: 'ich' }, 'z'),
    ).toEqual({ art: 'haufen', kennung: 'k', worauf: 'z' })
  })

  // **Auf sich selbst legt niemand etwas.** Beim Loslassen steht der Zeiger
  // oft noch über dem Platz, von dem die Karte kam.
  it('meldet nichts, wo die Karte auf sich selbst landet', () => {
    expect(
      losgelassen(zwei, laufend, { zone: 'tisch', sitz: 'ich' }, 'k'),
    ).toBeUndefined()
  })

  // **Und nichts, wo sie schon auf demselben Haufen liegt.** Ein Befehl
  // dafür ginge gegen einen Stand, der schon gilt.
  it('meldet nichts innerhalb desselben Haufens', () => {
    expect(
      losgelassen(inEinemHaufen, laufend, { zone: 'tisch', sitz: 'ich' }, 'z'),
    ).toBeUndefined()
  })

  // **Neben den Haufen gezogen heißt: heraus damit.** Das ist die
  // Gegenbewegung, und sie braucht keinen zweiten Griff.
  it('nimmt eine Karte aus dem Haufen, wo sie daneben landet', () => {
    expect(
      losgelassen(inEinemHaufen, laufend, { zone: 'tisch', sitz: 'ich' }),
    ).toEqual({ art: 'enthaufen', kennung: 'k' })
  })

  // **Die Gegenprobe**: wer in keinem Haufen liegt und auf freien Platz
  // zieht, meldet weiter nichts. Sonst schickte jeder Zug innerhalb einer
  // Zone einen Befehl, den der Kern abweist.
  it('meldet weiter nichts für eine Karte ohne Haufen', () => {
    expect(
      losgelassen(zwei, laufend, { zone: 'tisch', sitz: 'ich' }),
    ).toBeUndefined()
  })
})

// **Ein Zug über freien Platz legt ab — wenn der Tisch nicht selbst ordnet**
// — PLAY-04.AK-03
// (#1712).
//
// # Der Satz, der dabei überholt war
//
// Hier stand: „ein Zug über freien Platz ist eine Anordnung, die niemand
// festhalten muss". Das galt, solange der Tisch **immer** selbst ordnete —
// dann wäre eine festgehaltene Lage beim nächsten Zug ohnehin überschrieben.
// Seit die Anordnung abschaltbar ist, ist es genau umgekehrt.
describe('ein Zug über freien Platz', () => {
  // **Der Weg zählt, nicht der Zeiger.** Wo die Karte landet, ist ihre
  // bisherige Lage plus das, was gezogen wurde — die Funktion rechnet es
  // selbst, damit es niemand ein zweites Mal tut.
  const geschobenUm = { kennung: 'k', vonX: 10, vonY: 20, dx: 30, dy: 40 }

  it('legt ab, wo der Tisch nicht selbst ordnet', () => {
    const bisher = plan.lagen.find((eine) => eine.kennung === 'k')

    expect(
      losgelassen(
        plan,
        geschobenUm,
        { zone: 'tisch', sitz: 'ich' },
        undefined,
        true,
      ),
    ).toEqual({
      art: 'ablegen',
      kennung: 'k',
      x: (bisher?.x ?? 0) + 30,
      y: (bisher?.y ?? 0) + 40,
    })
  })

  // **Ordnet der Tisch selbst, meldet er nichts** — eine festgehaltene Lage
  // wäre beim nächsten Zug überschrieben, und eine Meldung, die nichts
  // bewirkt, ist schlimmer als keine.
  it('meldet nichts, wo der Tisch selbst ordnet', () => {
    expect(
      losgelassen(plan, geschobenUm, { zone: 'tisch', sitz: 'ich' }),
    ).toBeUndefined()
  })

  // **Und aus einem Haufen genommen wird weiterhin.** Das ist eine Aussage
  // über die Anordnung im Sinne von CR 400.5 und keine über Millimeter — sie
  // gilt, ob der Tisch selbst ordnet oder nicht.
  it('nimmt weiterhin aus dem Haufen', () => {
    const mitHaufen = aufDemTisch([
      { kennung: 'k', haufen: 'k' },
      { kennung: 'z', haufen: 'k' },
    ])

    expect(
      losgelassen(
        mitHaufen,
        geschobenUm,
        { zone: 'tisch', sitz: 'ich' },
        undefined,
        true,
      ),
    ).toEqual({ art: 'enthaufen', kennung: 'k' })
  })
})
