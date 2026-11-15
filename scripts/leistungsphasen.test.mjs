// Prüfungen von scripts/leistungsphasen.mjs — QA-04.AK-03
// (#1735).

import { test } from 'node:test'
import assert from 'node:assert/strict'
import { existsSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { bilderImVerzeichnis, bildpruefung, kartenImSet, quelltextwert, sitzungsschritte, wiederhole } from './leistungsphasen.mjs'
import { png } from './pngbau.mjs'

test('wiederhole dreht die Reihenfolge und hält Platz und Wiederholung fest', async () => {
  const gefahren = []
  const teilnehmer = [{ name: 'a' }, { name: 'b' }]
  const faelle = await wiederhole(teilnehmer, 2, 'start', '', async (t, i) => {
    gefahren.push(`${t.name}${i}`)
    return { wandMs: 1, aus: 'groß und unnötig' }
  })
  assert.deepEqual(gefahren, ['a0', 'b0', 'b1', 'a1'])
  assert.deepEqual(faelle.map((f) => [f.programm, f.wiederholung, f.platz]), [['a', 0, 0], ['b', 0, 1], ['b', 1, 0], ['a', 1, 1]])
  assert.equal(faelle[0].aufgabe, 'start')
  assert.equal('aus' in faelle[0], false, 'die Standardausgabe gehört nicht in den Beleg')
})

test('vor jedem Fall wird gewartet, bis der Rechner frei ist — und wie lange, steht am Fall', async () => {
  const folge = []
  const faelle = await wiederhole([{ name: 'a' }, { name: 'b' }], 1, 'start', '', async (t) => {
    folge.push(`fall ${t.name}`)
    return { wandMs: 1 }
  }, undefined, async () => {
    folge.push('warten')
    return 2.5
  })
  assert.deepEqual(folge, ['warten', 'fall a', 'warten', 'fall b'])
  assert.equal(faelle[0].gewartetS, 2.5)
})

test('die Sitzung zeichnet kalt, warm und die erste Karte noch einmal — und räumt die Bilder weg', async () => {
  const ort = mkdtempSync(join(tmpdir(), 'leistungsphasen-'))
  const geschrieben = []
  const sitzung = {
    zeichne: async (stelle, ziel) => {
      geschrieben.push(ziel)
      if (stelle === 9) return { ok: false, dauerMs: 3, text: 'ERROR: kaputt' }
      writeFileSync(ziel, png({ breite: 375, hoehe: 523 }))
      return { ok: true, dauerMs: stelle, text: '' }
    },
  }
  const karten = [
    { stelle: 4, schicht: 'Normal', satz: 'S', vorlage: 'v' },
    { stelle: 9, schicht: 'Problemfall', satz: 'S', vorlage: 'w' },
  ]
  const schritte = await sitzungsschritte(sitzung, karten, ort, 'mse2-0')
  assert.deepEqual(schritte.map((s) => [s.zustand, s.stelle, s.ok]), [['kalt', 4, true], ['warm', 9, false], ['wiederholung', 4, true]])
  assert.equal(schritte[0].breite, 375)
  assert.equal(schritte[1].befund, 'ERROR: kaputt')
  assert.equal(new Set(geschrieben).size, 3, 'jeder Schritt hat seinen eigenen Dateinamen')
  assert.equal(geschrieben.some((p) => existsSync(p)), false)
})

test('meldet das Programm Erfolg und liegt kein PNG da, ist der Schritt gescheitert', async () => {
  const ort = mkdtempSync(join(tmpdir(), 'leistungsphasen-'))
  const sitzung = { zeichne: async () => ({ ok: true, dauerMs: 1, text: '' }) }
  const [schritt] = await sitzungsschritte(sitzung, [{ stelle: 0, schicht: 'Normal' }], ort, 'x')
  assert.equal(schritt.ok, false)
  assert.equal(schritt.befund, 'kein PNG am Ziel')
})

test('ein Verzeichnis voller Bilder wird gezählt, nach Größe, und Kaputtes benannt', () => {
  const ort = mkdtempSync(join(tmpdir(), 'leistungsphasen-'))
  writeFileSync(join(ort, 'a.png'), png({ breite: 375, hoehe: 523 }))
  writeFileSync(join(ort, 'b.png'), png({ breite: 375, hoehe: 523 }))
  writeFileSync(join(ort, 'c.png'), png({ breite: 523, hoehe: 375 }))
  writeFileSync(join(ort, 'd.png'), 'kein Bild')
  writeFileSync(join(ort, 'manifest.json'), '{}')
  const v = bilderImVerzeichnis(ort)
  assert.deepEqual([v.anzahl, v.kaputt, v.groessen], [4, 1, { '375x523': 2, '523x375': 1 }])
  assert.deepEqual(v.dateien, {
    'a.png': { breite: 375, hoehe: 523 }, 'b.png': { breite: 375, hoehe: 523 }, 'c.png': { breite: 523, hoehe: 375 },
  })
  assert.deepEqual(v.kaputteDateien.map((k) => k.datei), ['d.png'])
  assert.match(v.kaputteDateien[0].fehler, /Signatur/)
  const leer = bilderImVerzeichnis(join(ort, 'fehlt'))
  assert.deepEqual([leer.anzahl, leer.kaputt, leer.groessen, leer.dateien], [0, 0, {}, {}])
})

test('ein abgeschnittenes Bild zählt als kaputt, auch wenn sein Anfang ein PNG ist', () => {
  const ort = mkdtempSync(join(tmpdir(), 'leistungsphasen-'))
  const ganz = png({ breite: 200, hoehe: 200, teile: 3 })
  writeFileSync(join(ort, 'ganz.png'), ganz)
  writeFileSync(join(ort, 'halb.png'), ganz.subarray(0, Math.floor(ganz.length / 2)))
  const v = bilderImVerzeichnis(ort)
  assert.deepEqual([v.anzahl, v.kaputt], [2, 1])
  assert.deepEqual(Object.keys(v.dateien), ['ganz.png'], 'ein kaputtes Bild hat keine Maße, die man vergleicht')
})

test('die Bildprüfung nennt Größe, Bytes und Hash — oder dass nichts da ist', () => {
  const ort = mkdtempSync(join(tmpdir(), 'leistungsphasen-'))
  const bild = png({ breite: 2, hoehe: 3 })
  writeFileSync(join(ort, 'a.png'), bild)
  const p = bildpruefung(join(ort, 'a.png'))
  assert.equal(p.png, true)
  assert.equal(p.breite, 2)
  assert.equal(p.hoehe, 3)
  assert.equal(p.bytes, bild.length)
  assert.equal(p.fehler, null)
  assert.match(p.sha256, /^[0-9a-f]{64}$/)
  assert.deepEqual(bildpruefung(join(ort, 'b.png')), { da: false })
})

test('die Bildprüfung sagt bei einem unvollständigen Bild, woran es fehlt — und gibt keine Maße', () => {
  const ort = mkdtempSync(join(tmpdir(), 'leistungsphasen-'))
  const ganz = png({ breite: 50, hoehe: 50 })
  writeFileSync(join(ort, 'a.png'), ganz.subarray(0, ganz.length - 20))
  const p = bildpruefung(join(ort, 'a.png'))
  assert.equal(p.png, false)
  assert.equal(p.breite, undefined)
  assert.match(p.fehler, /abgeschnitten|IEND/)
})

test('ein Schritt, dessen Bild nur anfängt wie ein PNG, ist gescheitert — mit dem Grund', async () => {
  const ort = mkdtempSync(join(tmpdir(), 'leistungsphasen-'))
  const ganz = png({ breite: 80, hoehe: 80 })
  const sitzung = { zeichne: async (stelle, ziel) => { writeFileSync(ziel, ganz.subarray(0, 60)); return { ok: true, dauerMs: 1, text: '' } } }
  const [schritt] = await sitzungsschritte(sitzung, [{ stelle: 0, schicht: 'Normal' }], ort, 'x')
  assert.equal(schritt.ok, false)
  assert.match(schritt.befund, /^kein vollständiges PNG am Ziel: /)
  assert.equal(schritt.breite, null)
})

test('die Karten einer gesicherten Setdatei werden an ihren Blöcken gezählt', () => {
  const ort = mkdtempSync(join(tmpdir(), 'leistungsphasen-'))
  writeFileSync(join(ort, 'x.mse-set'), '')
  const lesen = () => 'game: magic\ncard:\n\tname: a\ncard:\n\tname: b\nkeyword:\n\tcard: kein Block\n'
  assert.equal(kartenImSet(lesen, join(ort, 'x.mse-set')), 2)
  assert.equal(kartenImSet(lesen, join(ort, 'fehlt.mse-set')), null)
  assert.equal(kartenImSet(() => null, join(ort, 'x.mse-set')), null)
})

test('ein Kartenname aus MSE2 ist Quelltext — mit Anführungszeichen, die zum Namen gehören können', () => {
  assert.equal(quelltextwert('"All in Good Time"'), 'All in Good Time')
  assert.equal(quelltextwert('"\\"Ach! Hans, Run!\\""'), '"Ach! Hans, Run!"', 'die Anführungszeichen im Namen bleiben')
  assert.equal(quelltextwert('"a\\\\b"'), 'a\\b')
  assert.equal(quelltextwert('70'), '70', 'was keine Zeichenkette ist, bleibt, wie es ist')
})
