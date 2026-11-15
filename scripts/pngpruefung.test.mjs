// Prüfungen von scripts/pngpruefung.mjs — Z2 des Zweitblicks über Phase 6.11.
//
// **Ein Bild ist vollständig, wenn es sich entpacken lässt und die Länge hat,
// die sein Kopf verspricht.** Bis dahin prüfte der Lauf 24 Byte: ein auf
// halber Strecke abgeschnittenes Bild bestand, solange sein Anfang stimmte.

import { test } from 'node:test'
import assert from 'node:assert/strict'
import { deflateSync } from 'node:zlib'

import { chunk, png, rohdaten } from './pngbau.mjs'
import { pngPruefen } from './pngpruefung.mjs'

test('ein vollständiges Bild besteht und nennt seine Maße', () => {
  const p = pngPruefen(png({ breite: 435, hoehe: 621 }))
  assert.deepEqual(p, { ok: true, breite: 435, hoehe: 621, farbart: 6, tiefe: 8, verschachtelt: false })
})

test('jede Farbart und Tiefe, die PNG kennt, wird nach ihrer Zeilenlänge gerechnet', () => {
  const faelle = [
    [0, 1], [0, 2], [0, 4], [0, 8], [0, 16], [2, 8], [2, 16], [3, 1], [3, 2], [3, 4], [3, 8], [4, 8], [4, 16], [6, 8], [6, 16],
  ]
  for (const [farbart, tiefe] of faelle) {
    for (const breite of [1, 3, 9, 13]) {
      const p = pngPruefen(png({ breite, hoehe: 5, farbart, tiefe }))
      assert.equal(p.ok, true, `Farbart ${farbart}, Tiefe ${tiefe}, Breite ${breite}: ${p.fehler}`)
    }
  }
})

test('ein verschachteltes Bild (Adam7) hat die Länge seiner sieben Durchgänge', () => {
  for (const [breite, hoehe] of [[1, 1], [5, 3], [9, 9], [17, 2], [3, 17]]) {
    const p = pngPruefen(png({ breite, hoehe, verschachtelt: true }))
    assert.equal(p.ok, true, `${breite}x${hoehe}: ${p.fehler}`)
    assert.equal(p.verschachtelt, true)
  }
})

test('die Bilddaten dürfen auf mehrere IDAT-Blöcke verteilt sein', () => {
  assert.equal(pngPruefen(png({ breite: 64, hoehe: 64, teile: 5 })).ok, true)
})

test('der Kopf allein — 24 Byte — ist kein Bild', () => {
  const ganz = png({ breite: 4, hoehe: 3 })
  const p = pngPruefen(ganz.subarray(0, 24))
  assert.equal(p.ok, false)
  assert.match(p.fehler, /abgeschnitten/)
})

test('ein auf halber Strecke abgeschnittenes Bild fällt durch, auch wenn sein Anfang stimmt', () => {
  const ganz = png({ breite: 200, hoehe: 200, teile: 4 })
  for (const anteil of [0.5, 0.9, 0.99]) {
    const p = pngPruefen(ganz.subarray(0, Math.floor(ganz.length * anteil)))
    assert.equal(p.ok, false, `bei ${anteil}`)
    assert.ok(p.fehler, 'der Fehler wird genannt')
  }
})

test('ein Bild ohne IEND ist unvollständig', () => {
  const ganz = png()
  const p = pngPruefen(ganz.subarray(0, ganz.length - 12))
  assert.equal(p.ok, false)
  assert.match(p.fehler, /IEND/)
})

test('weniger Zeilen, als der Kopf verspricht, fallen durch — obwohl die Daten sauber gepackt sind', () => {
  const zuWenig = rohdaten(4, 2, 6, 8)
  const p = pngPruefen(png({ breite: 4, hoehe: 3, roh: zuWenig }))
  assert.equal(p.ok, false)
  assert.match(p.fehler, /erwartet/)
  assert.equal(p.breite, 4, 'was der Kopf sagt, bleibt lesbar')
})

test('mehr Daten, als der Kopf verspricht, fallen ebenso durch', () => {
  const p = pngPruefen(png({ breite: 4, hoehe: 3, roh: rohdaten(4, 5, 6, 8) }))
  assert.equal(p.ok, false)
  assert.match(p.fehler, /mehr Bilddaten/, 'entpackt wird nur bis zur erwarteten Länge, nicht bis zum Ende einer Bombe')
})

test('ein verdrehtes Bit in den Bilddaten verrät die Prüfsumme', () => {
  const bild = Buffer.from(png({ breite: 30, hoehe: 30 }))
  bild[bild.indexOf('IDAT', 0, 'latin1') + 6] ^= 0x01
  const p = pngPruefen(bild)
  assert.equal(p.ok, false)
  assert.match(p.fehler, /Prüfsumme/)
})

test('ein Filterbyte außerhalb von 0 bis 4 ist kein Bild', () => {
  const roh = rohdaten(4, 3, 6, 8)
  roh[(1 + 16) * 2] = 9
  const p = pngPruefen(png({ breite: 4, hoehe: 3, roh }))
  assert.equal(p.ok, false)
  assert.match(p.fehler, /Filter/)
})

test('Bilddaten, die sich nicht entpacken lassen, fallen durch', () => {
  const kopf = png({ breite: 4, hoehe: 3 }).subarray(0, 8 + 25)
  const kaputt = Buffer.concat([kopf, chunk('IDAT', Buffer.from('kein zlib-Strom')), chunk('IEND')])
  const p = pngPruefen(kaputt)
  assert.equal(p.ok, false)
  assert.match(p.fehler, /entpack/)
})

test('ein Bild ohne IDAT fällt durch', () => {
  const kopf = png({ breite: 4, hoehe: 3 }).subarray(0, 8 + 25)
  const p = pngPruefen(Buffer.concat([kopf, chunk('IEND')]))
  assert.equal(p.ok, false)
  assert.match(p.fehler, /IDAT/)
})

test('Bilddaten ohne Palette bei Farbart 3 fallen durch', () => {
  const p = pngPruefen(png({ breite: 4, hoehe: 3, farbart: 3, tiefe: 8, palette: false }))
  assert.equal(p.ok, false)
  assert.match(p.fehler, /PLTE/)
})

test('was keine PNG-Signatur trägt, ist kein Bild', () => {
  assert.match(pngPruefen(Buffer.from('GIF89a…')).fehler, /Signatur/)
  assert.match(pngPruefen(Buffer.alloc(0)).fehler, /Signatur/)
  assert.equal(pngPruefen(Buffer.from('kein Bild')).ok, false)
})

test('ein Kopf mit unmöglichen Angaben fällt durch', () => {
  for (const [name, aendere] of [
    ['Breite null', (k) => k.writeUInt32BE(0, 0)],
    ['Höhe null', (k) => k.writeUInt32BE(0, 4)],
    ['Tiefe 3', (k) => { k[8] = 3 }],
    ['Farbart 1', (k) => { k[9] = 1 }],
    ['Farbart 2 mit Tiefe 4', (k) => { k[9] = 2; k[8] = 4 }],
    ['Verfahren 1', (k) => { k[10] = 1 }],
    ['Filterverfahren 1', (k) => { k[11] = 1 }],
    ['Verschachtelung 2', (k) => { k[12] = 2 }],
  ]) {
    const p = pngPruefen(png({ breite: 4, hoehe: 3, kopfAendern: aendere }))
    assert.equal(p.ok, false, name)
    assert.doesNotMatch(p.fehler, /Prüfsumme/, `${name}: gefallen am Kopf und nicht an der Summe`)
  }
})

test('ein Bild, das mehr Speicher verlangte als erlaubt, wird nicht entpackt', () => {
  const kopf = Buffer.alloc(13)
  kopf.writeUInt32BE(2 ** 30, 0)
  kopf.writeUInt32BE(2 ** 30, 4)
  kopf[8] = 8
  kopf[9] = 6
  const bild = Buffer.concat([
    Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
    chunk('IHDR', kopf), chunk('IDAT', deflateSync(Buffer.alloc(8))), chunk('IEND'),
  ])
  const p = pngPruefen(bild)
  assert.equal(p.ok, false)
  assert.match(p.fehler, /Größe/)
})

test('der erste Block muss der Kopf sein', () => {
  const ganz = png()
  const ohneKopf = Buffer.concat([ganz.subarray(0, 8), ganz.subarray(8 + 25)])
  assert.match(pngPruefen(ohneKopf).fehler, /IHDR/)
})

test('Bytes hinter dem Ende stören ein sonst vollständiges Bild nicht', () => {
  assert.equal(pngPruefen(Buffer.concat([png(), Buffer.from('Nachsatz')])).ok, true)
})

test('ein Block mit unmöglicher Länge fällt durch', () => {
  const bild = Buffer.from(png())
  bild.writeUInt32BE(0xffffffff, 8)
  assert.equal(pngPruefen(bild).ok, false)
})
