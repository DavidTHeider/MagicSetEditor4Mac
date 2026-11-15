// Prüfungen von scripts/leistungsteile.mjs — Z1 und Z2 des Zweitblicks über
// Phase 6.11: der Batchfall `bilder` und der Vergleich der Bildmaße.
//
// Gefahren wird kein Programm. Der Läufer reicht jedem Teil seinen Kontext
// (`c`); hier ist es ein Kontext aus Doubles, deren Antworten der Test vorgibt.

import { test } from 'node:test'
import assert from 'node:assert/strict'
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { bilder, masseInSitzungen, navigation } from './leistungsteile.mjs'
import { bilderImVerzeichnis, quelltextwert, sitzungsschritte, wiederhole } from './leistungsphasen.mjs'
import { mse2Dateinamen } from './leistungsmasse.mjs'
import { png } from './pngbau.mjs'

const namen = ['Eins', 'Zwei', 'Drei', 'Vier']
const bild = (breite = 435, hoehe = 621) => png({ breite, hoehe })

/** kontext baut den Kontext des Läufers aus Doubles. */
function kontext({ mse2Bilder, heuteZeichne, v050Zeichne }) {
  const ort = mkdtempSync(join(tmpdir(), 'leistungsteile-'))
  const sitzungsTeilnehmer = (name, zeichne) => ({
    name, ids: { Archenemy: '7' }, vorrat: `/${name}/vorrat.db`,
    sitzung: async () => ({
      karten: namen.map((n, i) => ({ id: i })),
      zeichne: (stelle, ziel) => zeichne(stelle, ziel),
      lauf: { schliessen: async () => ({ wandMs: name === 'heute' ? 2400 : 61900, fehler: null }) },
    }),
  })
  const mse2 = {
    name: 'mse2',
    bilder: async (set, ziel) => {
      mse2Dateinamen(namen).forEach((datei, i) => writeFileSync(join(ziel, datei), mse2Bilder?.(i) ?? bild()))
      return { wandMs: 3000, fehler: null }
    },
  }
  const heute = sitzungsTeilnehmer('heute', heuteZeichne)
  const teilnehmer = [mse2, heute]
  if (v050Zeichne) teilnehmer.push(sitzungsTeilnehmer('v050', v050Zeichne))
  return {
    k: { bilder: 1 }, teilnehmer, mse4macs: teilnehmer.filter((t) => t.name !== 'mse2'),
    satz: (name) => ({ name, pfad: `/kanon/${name}.mse-set` }),
    kartenzahl: { Archenemy: namen.length },
    vorlagen: { Archenemy: namen.map((name) => ({ name, vorlage: 'v' })) },
    ausgabe: (name) => { const o = join(ort, name); mkdirSync(o, { recursive: true }); return o },
    wiederhole, bilderImVerzeichnis, melde: undefined,
    faelle: [], schritte: [], befunde: [], extra: {},
  }
}

/** schreibt gibt ein `zeichne`, das ein Bild ablegt — oder scheitert, wie der Test es will. */
const schreibt = (ergebnis = () => ({ ok: true })) => async (stelle, ziel) => {
  const r = ergebnis(stelle)
  if (r.ok) writeFileSync(ziel, r.daten ?? bild())
  return { ok: r.ok, dauerMs: 5, text: r.text ?? '' }
}

const fallVon = (c, programm) => c.faelle.find((f) => f.programm === programm)

test('alle Bilder gezeichnet, alle heil, alle gleich groß: kein Befund — und der Durchsatz steht da', async () => {
  const c = kontext({ heuteZeichne: schreibt() })
  await bilder(c)
  for (const f of c.faelle) {
    assert.equal(f.befund, null, f.programm)
    assert.equal(f.fehler, null, f.programm)
    assert.equal(f.bilder, 4)
  }
  assert.deepEqual(c.extra.bildgroessen, { mse2: { '435x621': 4 }, heute: { '435x621': 4 } })
  assert.equal(Math.round(fallVon(c, 'heute').durchsatz), Math.round(4 / 2.4))
})

// **Z1**: im Beleg standen für MSE4Mac 24 Bilder bei 70 Karten, weil das
// Guthaben den Rest abwies — und die Schleife ignorierte jeden Ruf, der
// scheiterte.
test('ein Zeichenruf, der scheitert, steht als Befund am Fall — und der Fall sagt, wie viele', async () => {
  const abgewiesen = 'schreiben.karte.zeichnen: zu viele Rufe — in 1 s wieder (Vorgang: f70b1637)'
  const c = kontext({ heuteZeichne: schreibt((stelle) => (stelle < 2 ? { ok: true } : { ok: false, text: abgewiesen })) })
  await bilder(c)
  const heute = fallVon(c, 'heute')
  assert.match(heute.befund, /2 von 4 Zeichenrufen gescheitert/)
  assert.match(heute.befund, /zu viele Rufe/, 'der erste Grund steht dabei')
  assert.match(heute.befund, /2 Bilder, 0 kaputt, bei 4 Karten/)
  assert.ok(heute.fehler, 'der Fall zählt als Fehler')
  assert.equal(fallVon(c, 'mse2').befund, null, 'MSE2 hat nichts mit dem Guthaben von heute zu tun')
})

test('jeder Ruf wird abgewartet und geprüft, auch nach dem ersten Scheitern', async () => {
  const gerufen = []
  const c = kontext({
    heuteZeichne: async (stelle, ziel) => { gerufen.push(stelle); return { ok: false, dauerMs: 1, text: 'ERROR: kaputt' } },
  })
  await bilder(c)
  assert.deepEqual(gerufen, [0, 1, 2, 3])
  assert.match(fallVon(c, 'heute').befund, /4 von 4 Zeichenrufen gescheitert/)
})

test('ein Ruf, der Erfolg meldet, aber kein Bild hinterlässt, fehlt in der Zählung', async () => {
  const c = kontext({ heuteZeichne: async () => ({ ok: true, dauerMs: 1, text: '' }) })
  await bilder(c)
  assert.match(fallVon(c, 'heute').befund, /0 Bilder, 0 kaputt, bei 4 Karten/)
})

// **Z2**: ein Bild, das auf halber Strecke abbricht, hat den Anfang eines PNG.
test('ein abgeschnittenes Bild zählt als kaputt und steht mit seinem Grund im Befund', async () => {
  const ganz = bild()
  const c = kontext({
    heuteZeichne: schreibt((stelle) => (stelle === 1 ? { ok: true, daten: ganz.subarray(0, 200) } : { ok: true })),
  })
  await bilder(c)
  assert.match(fallVon(c, 'heute').befund, /4 Bilder, 1 kaputt, bei 4 Karten/)
  assert.match(fallVon(c, 'heute').befund, /1\.png/)
})

test('eine andere Größe bei einer gepaarten Karte ist ein Befund am Fall — mit Karte, Name und beiden Maßen', async () => {
  const c = kontext({ heuteZeichne: schreibt((stelle) => (stelle === 2 ? { ok: true, daten: bild(435, 620) } : { ok: true })) })
  await bilder(c)
  const heute = fallVon(c, 'heute')
  assert.match(heute.befund, /Karte 2 \(Drei\): 435x620, MSE2 435x621/)
  assert.ok(heute.fehler)
  assert.equal(fallVon(c, 'mse2').befund, null)
  assert.equal(heute.bilder, 4, 'die Zahl der Bilder stimmt — nur die Größe nicht')
})

test('jedes Programm neben MSE2 wird gegen MSE2 gehalten, auch 0.5.0', async () => {
  const c = kontext({
    heuteZeichne: schreibt(),
    v050Zeichne: schreibt((stelle) => (stelle === 0 ? { ok: true, daten: bild(400, 621) } : { ok: true })),
  })
  await bilder(c)
  assert.equal(fallVon(c, 'heute').befund, null)
  assert.match(fallVon(c, 'v050').befund, /Karte 0 \(Eins\): 400x621, MSE2 435x621/)
})

test('ein Bild, das bei MSE2 fehlt, ist ungepaart und steht im Befund — kein stilles Überspringen', async () => {
  const c = kontext({ heuteZeichne: schreibt() })
  const alt = c.teilnehmer[0].bilder
  c.teilnehmer[0].bilder = async (set, ziel) => {
    const m = await alt(set, ziel)
    writeFileSync(join(ziel, 'Eins.png'), 'kein Bild')
    return m
  }
  await bilder(c)
  const heute = fallVon(c, 'heute')
  assert.match(heute.befund, /Karte 0 \(Eins\)/)
  assert.match(heute.befund, /MSE2/)
})

test('die Maße der Schritte in den Sitzungen werden je Karte gegen MSE2 gehalten und tragen sich in die Sitzung ein', () => {
  const schritt = (programm, stelle, breite) => ({ programm, satz: 'Multicolor', stelle, zustand: 'warm', ok: true, breite, hoehe: 621 })
  const c = {
    schritte: [schritt('mse2', 3, 435), schritt('heute', 3, 435), schritt('mse2', 8, 435), schritt('heute', 8, 400), schritt('heute', 9, 435), schritt('mse2', 9, 435)],
    faelle: [
      { aufgabe: 'navigation-sitzung', satz: 'Multicolor', programm: 'heute', wiederholung: 0, fehler: null },
      { aufgabe: 'navigation-sitzung', satz: 'Multicolor', programm: 'mse2', wiederholung: 0, fehler: null },
      { aufgabe: 'navigation-sitzung', satz: 'Archenemy', programm: 'heute', wiederholung: 0, fehler: null },
    ],
  }
  masseInSitzungen(c)
  assert.match(c.faelle[0].befund, /Multicolor\[8\]: heute 400x621, MSE2 435x621/)
  assert.ok(c.faelle[0].fehler)
  assert.equal(c.faelle[1].befund, undefined, 'die Referenz bleibt, wie sie ist')
  assert.equal(c.faelle[2].befund, undefined, 'ein anderer Satz bleibt, wie er ist')
})

test('viele Abweichungen stehen verkürzt da, mit der Zahl der übrigen', () => {
  const c = {
    schritte: Array.from({ length: 6 }, (_, i) => [
      { programm: 'mse2', satz: 'S', stelle: i, ok: true, breite: 435, hoehe: 621 },
      { programm: 'heute', satz: 'S', stelle: i, ok: true, breite: 400, hoehe: 621 },
    ]).flat(),
    faelle: [{ aufgabe: 'navigation-sitzung', satz: 'S', programm: 'heute', wiederholung: 0, fehler: null }],
  }
  masseInSitzungen(c)
  assert.match(c.faelle[0].befund, /S\[0\].*S\[1\].*S\[2\].*und 3 weitere/)
})

/** sitzungskontext baut den Kontext für `navigation`: zwei Programme, zwei Karten eines Satzes. */
function sitzungskontext(heuteBreite) {
  const ort = mkdtempSync(join(tmpdir(), 'leistungsteile-'))
  const sitzung = (breite) => async () => ({
    bereitMs: 5,
    name: async (stelle) => ({ ok: true, text: namen[stelle] }),
    zeichne: async (stelle, ziel) => { writeFileSync(ziel, png({ breite: stelle === 1 ? breite : 435, hoehe: 621 })); return { ok: true, dauerMs: 3, text: '' } },
    lauf: { schliessen: async () => ({ wandMs: 1, aus: 'x', fehler: null }) },
  })
  const mse2 = { name: 'mse2', sitzung: sitzung(435) }
  const heute = { name: 'heute', ids: { Archenemy: '7' }, vorrat: '/v', sitzung: sitzung(heuteBreite) }
  return {
    k: { navigation: 1 }, teilnehmer: [mse2, heute], mse4macs: [heute],
    auswahl: [{ satz: 'Archenemy', stelle: 0, name: 'Eins', schicht: 'Normal' }, { satz: 'Archenemy', stelle: 1, name: 'Zwei', schicht: 'Normal' }],
    satz: (name) => ({ name, pfad: `/kanon/${name}.mse-set` }),
    ausgabe: (name) => { const o = join(ort, name); mkdirSync(o, { recursive: true }); return o },
    warte: async () => 0, pipelineLaeuft: () => false, sitzungsschritte, quelltextwert,
    faelle: [], schritte: [], befunde: [], extra: {},
  }
}

test('die Navigation hält die Bildgrößen der Sitzungen gegen MSE2 und markiert die Sitzung, die abweicht', async () => {
  const c = sitzungskontext(400)
  await navigation(c)
  const heute = c.faelle.find((f) => f.programm === 'heute')
  assert.match(heute.befund, /Archenemy\[1\]: heute 400x621, MSE2 435x621/)
  assert.equal(c.faelle.find((f) => f.programm === 'mse2').befund, undefined)
})

test('gleich große Bilder in beiden Sitzungen: kein Befund', async () => {
  const c = sitzungskontext(435)
  await navigation(c)
  assert.deepEqual(c.faelle.filter((f) => f.befund), [])
  assert.equal(c.befunde.length, 0)
})
