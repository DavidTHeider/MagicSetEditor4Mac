// **Was am Tisch gerade geschehen ist** — PLAY-04.AK-02
// (#1712).
//
// **Die Rechnung ohne einen Baum und ohne eine Grafikkarte**: hier steht,
// *welche* Wirkung ein Stand gegenüber dem vorigen hat — nicht, wie sie
// aussieht.
import { describe, expect, it } from 'vitest'

import { einObjekt, einSitz, einZug, einePartie } from '../seiten/partieprobe'
import {
  Wirkungsarten,
  standVon,
  wirkungen,
  wirkungsbild,
} from './tischwirkungen'

// einStand baut den Vergleichsstand aus einer Partie.
const einStand = (teile: Parameters<typeof einePartie>[0] = {}) =>
  standVon(einePartie(teile))

// arten nennt die Wirkungen einer Rechnung, samt dem, woran sie hängen.
const arten = (ergebnis: ReturnType<typeof wirkungen>) =>
  ergebnis.map((eine) => `${eine.art}:${eine.woran}`)

describe('wirkungen', () => {
  // **Beim ersten Bild geschieht nichts.** Wer einen Tisch aufschlägt, hat
  // nichts verpasst — und ein Brett, das beim Öffnen einmal ganz aufblitzt,
  // wäre kein Hinweis, sondern ein Blendwerk.
  //
  // **Und geprüft wird das an einem vollen Tisch**: eine Karte auf dem
  // Brett, ein laufender Zug, ein Ausgang. Gegen die leere Anfangspartie war
  // der Lauf grün, **auch wenn man den Rückgabewert entfernte** — sie hat
  // nichts, was eine Wirkung auslösen könnte. Eine Gegenprobe, die nichts
  // ändert, ist keine (§10), und diese hat es am 2026-09-22 gezeigt.
  it('meldet nichts für das erste Bild', () => {
    const voll = einStand({
      objekte: [einObjekt({ kennung: 'k1', zone: 'tisch', getappt: true })],
      sitze: [einSitz({ kennung: 's1', leben: 12, vergiftung: 3 })],
      zug: einZug({ nummer: 7 }),
      ausgang: { sieger: 's1', unentschieden: false, ausgeschieden: [] },
    })

    expect(wirkungen(undefined, voll)).toEqual([])
  })

  it('meldet nichts, wo sich nichts geändert hat', () => {
    expect(wirkungen(einStand(), einStand())).toEqual([])
  })

  // **Eintritt und Austritt sind das Brett, nicht die Sichtbarkeit** — CR
  // 403.1: das Schlachtfeld ist die Zone, in der Bleibendes liegt. Eine
  // Karte, die von der Hand auf den Friedhof geht, ist nie eingetreten.
  it('unterscheidet Eintritt, Austritt und bloße Bewegung', () => {
    const aufDerHand = einStand({
      objekte: [einObjekt({ kennung: 'k1', zone: 'hand' })],
    })

    const aufDemTisch = einStand({
      objekte: [einObjekt({ kennung: 'k1', zone: 'tisch' })],
    })

    const imFriedhof = einStand({
      objekte: [einObjekt({ kennung: 'k1', zone: 'friedhof' })],
    })

    expect(arten(wirkungen(aufDerHand, aufDemTisch))).toEqual(['eintritt:k1'])
    expect(arten(wirkungen(aufDemTisch, imFriedhof))).toEqual(['austritt:k1'])
    expect(arten(wirkungen(aufDerHand, imFriedhof))).toEqual(['bewegung:k1'])
  })

  // **Ein Objekt, das es nicht mehr gibt, ist ausgetreten** — es kann in eine
  // Zone gegangen sein, die dieser Sitz nicht sieht.
  it('liest ein verschwundenes Objekt als Austritt', () => {
    const aufDemTisch = einStand({
      objekte: [einObjekt({ kennung: 'k1', zone: 'tisch' })],
    })

    expect(arten(wirkungen(aufDemTisch, einStand({ objekte: [] })))).toEqual([
      'austritt:k1',
    ])
  })

  // **Und eines, das aus dem Nichts auf dem Tisch erscheint, ist
  // eingetreten** — ein Spielstein entsteht dort (CR 111.1).
  it('liest ein neues Objekt auf dem Tisch als Eintritt', () => {
    expect(
      arten(
        wirkungen(
          einStand({ objekte: [] }),
          einStand({ objekte: [einObjekt({ kennung: 't1', zone: 'tisch' })] }),
        ),
      ),
    ).toEqual(['eintritt:t1'])
  })

  // **Und ein neues Objekt anderswo ist kein Ereignis.** Eine Karte, die
  // dieser Sitz zum ersten Mal sieht — weil er in seine Bibliothek geblickt
  // hat oder die Sicht gewechselt ist —, hat nichts getan.
  it('meldet nichts für ein neues Objekt außerhalb des Bretts', () => {
    expect(
      wirkungen(
        einStand({ objekte: [] }),
        einStand({ objekte: [einObjekt({ kennung: 'k9', zone: 'hand' })] }),
      ),
    ).toEqual([])
  })

  it('meldet das Tappen und das Enttappen als Drehung', () => {
    const offen = einStand({
      objekte: [einObjekt({ kennung: 'k1', zone: 'tisch' })],
    })

    const getappt = einStand({
      objekte: [einObjekt({ kennung: 'k1', zone: 'tisch', getappt: true })],
    })

    expect(arten(wirkungen(offen, getappt))).toEqual(['drehung:k1'])
    expect(arten(wirkungen(getappt, offen))).toEqual(['drehung:k1'])
  })

  // **Ein Zug in eine andere Zone verdeckt die Drehung nicht** — eine Karte,
  // die getappt aufs Feld kommt, tut zweierlei (CR 302.2 gegen „schieben mit
  // getappt"), und wer nur eines zeigte, ließe das andere verschwinden.
  it('meldet Eintritt und Drehung zugleich', () => {
    const aufDerHand = einStand({
      objekte: [einObjekt({ kennung: 'k1', zone: 'hand' })],
    })

    const getapptAufDemTisch = einStand({
      objekte: [einObjekt({ kennung: 'k1', zone: 'tisch', getappt: true })],
    })

    expect(arten(wirkungen(aufDerHand, getapptAufDemTisch))).toEqual([
      'eintritt:k1',
      'drehung:k1',
    ])
  })

  it('meldet Leben und Gift je Sitz', () => {
    const vorher = einStand({
      sitze: [einSitz({ kennung: 's1', leben: 20, vergiftung: 0 })],
    })

    const weniger = einStand({
      sitze: [einSitz({ kennung: 's1', leben: 17, vergiftung: 0 })],
    })

    const vergiftet = einStand({
      sitze: [einSitz({ kennung: 's1', leben: 20, vergiftung: 2 })],
    })

    expect(arten(wirkungen(vorher, weniger))).toEqual(['leben:s1'])
    expect(arten(wirkungen(vorher, vergiftet))).toEqual(['gift:s1'])
  })

  // **Ein neuer Sitz bringt keine Wirkung mit.** Sein Lebensstand ist sein
  // Anfangsstand und keine Änderung.
  it('meldet nichts für einen Sitz, den es vorher nicht gab', () => {
    expect(
      wirkungen(
        einStand({ sitze: [] }),
        einStand({ sitze: [einSitz({ kennung: 's1', leben: 20 })] }),
      ),
    ).toEqual([])
  })

  it('meldet den Zugwechsel', () => {
    const erster = einStand({ zug: einZug({ nummer: 1 }) })
    const zweiter = einStand({ zug: einZug({ nummer: 2 }) })

    expect(arten(wirkungen(erster, zweiter))).toEqual(['zugwechsel:'])
  })

  // **Ein Sandkasten hat keinen Zug** (G9, #1592) — null bleibt null, und
  // eine Wirkung dafür gäbe es nicht.
  it('meldet keinen Zugwechsel ohne Zug', () => {
    expect(wirkungen(einStand(), einStand())).toEqual([])
  })

  it('meldet das Spielende genau einmal', () => {
    const laeuft = einStand()
    const vorbei = einStand({
      ausgang: { sieger: 's1', unentschieden: false, ausgeschieden: [] },
    })

    expect(arten(wirkungen(laeuft, vorbei))).toEqual(['spielende:'])
    expect(wirkungen(vorbei, vorbei)).toEqual([])
  })

  // **Unentschieden ist auch ein Ende** — CR 104.4.
  it('liest auch ein Unentschieden als Spielende', () => {
    expect(
      arten(
        wirkungen(
          einStand(),
          einStand({
            ausgang: { sieger: '', unentschieden: true, ausgeschieden: [] },
          }),
        ),
      ),
    ).toEqual(['spielende:'])
  })
})

describe('wirkungsbild', () => {
  // **Aus dem Modul und nicht aus einer Aufzählung hier** — die liefe bei
  // der neunten Wirkung auseinander, und dieser Lauf prüfte dann sieben von
  // acht, ohne rot zu werden (§10).
  const alle = Wirkungsarten

  // **Und die Zahl steht trotzdem da**: eine neunte Wirkung ist eine
  // Entscheidung, keine Nebenwirkung. PLAY-04.AK-02 zählt sieben Klassen
  // auf, und „Eintritt/Austritt" sind zwei davon.
  it('kennt genau acht Wirkungen', () => {
    expect(alle).toHaveLength(8)
  })

  // **Das ist das Kriterium, wörtlich**: „erhalten unterscheidbare Effekte".
  // Zwei Wirkungen, die gleich aussehen, sind eine.
  it('gibt jeder Wirkung eine eigene Farbe', () => {
    const farben = alle.map((art) => wirkungsbild(art).farbe)

    expect(new Set(farben).size).toBe(alle.length)
  })

  // **Und Farbe allein genügt nicht** — WCAG 1.4.1: sie darf nicht das
  // einzige Mittel sein. Jede Wirkung trägt deshalb ein eigenes Zeichen.
  it('gibt jeder Wirkung ein eigenes Zeichen', () => {
    const zeichen = alle.map((art) => wirkungsbild(art).zeichen)

    expect(new Set(zeichen).size).toBe(alle.length)
  })

  it('gibt jeder Wirkung eine Dauer über null', () => {
    for (const art of alle) {
      expect(wirkungsbild(art).dauer).toBeGreaterThan(0)
    }
  })
})
