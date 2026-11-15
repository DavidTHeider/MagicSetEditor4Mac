// **Der Scanner nach eingeschmuggelten Geheimnissen** — I05
// (#2263).
//
// Ein negativer Fund allein prüft nichts: ein Scanner, der nichts findet,
// weil er falsch sucht, sieht aus wie ein sauberes Paket. Deshalb hier die
// Gegenproben — ein ungeschwärztes Kontrollstück **wird** gefunden, auch
// verkleidet als URL-Kodierung oder Base64 (ein Basic-Auth-Kopf).
import { strict as assert } from 'node:assert'
import { createHash } from 'node:crypto'
import { mkdtempSync, writeFileSync, mkdirSync } from 'node:fs'
import { createServer } from 'node:net'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'

import { fundeIn, pruefeManifest, durchsucheVerzeichnis, ordneLog } from './geheimnisscan.mjs'

const marke = 'MSEGEHEIM7Q2X9K4WZ'

test('findet die Marke roh, URL-kodiert und in Base64 jeder Ausrichtung', () => {
  for (const vorsatz of ['', 'a', 'ab', 'abc']) {
    const kopf = 'Authorization: Basic ' + Buffer.from(`${vorsatz}nutzer:${marke}`).toString('base64')
    assert.equal(fundeIn(kopf, [marke]).length > 0, true, `Base64 mit Vorsatz »${vorsatz}« nicht gefunden: ${kopf}`)
  }

  assert.deepEqual(fundeIn(`proxy=http://nutzer:${marke}@127.0.0.1:9`, [marke]).map((f) => f.art), ['roh'])
  assert.deepEqual(fundeIn(`?schluessel=${encodeURIComponent(`${marke}/+`)}`, [`${marke}/+`]).map((f) => f.art), ['url'])
})

test('findet nichts in einem sauberen Text — und schwärzt nicht mit', () => {
  assert.deepEqual(fundeIn('proxy=http://nutzer:•••@127.0.0.1:9', [marke]), [])
})

function paket(teile, verfaelscht = {}) {
  const ort = mkdtempSync(join(tmpdir(), 'geheimnisscan-'))
  const angaben = []
  for (const [name, inhalt] of Object.entries(teile)) {
    writeFileSync(join(ort, name), inhalt)
    angaben.push({ name, groesse: Buffer.byteLength(inhalt), sha256: verfaelscht[name] ?? createHash('sha256').update(inhalt).digest('hex') })
  }
  writeFileSync(join(ort, 'manifest.json'), JSON.stringify({ fassung: '0.8.0', erstellt: '2026-10-10T20:00:00Z', teile: angaben }))
  return ort
}

test('das Manifest: jede Prüfsumme gilt, und nichts liegt daneben', () => {
  const ort = paket({ 'fassung.txt': '0.8.0', 'protokoll.txt': 'eine Zeile' })
  assert.deepEqual(pruefeManifest(ort, 'manifest.json'), { gueltig: true, fehler: [] })

  writeFileSync(join(ort, 'beigelegt.txt'), 'nicht im Manifest')
  assert.match(pruefeManifest(ort, 'manifest.json').fehler.join(' '), /beigelegt\.txt/)
})

test('das Manifest: eine falsche Prüfsumme wird genannt', () => {
  const ort = paket({ 'fassung.txt': '0.8.0' }, { 'fassung.txt': '00' })
  const ergebnis = pruefeManifest(ort, 'manifest.json')
  assert.equal(ergebnis.gueltig, false)
  assert.match(ergebnis.fehler.join(' '), /fassung\.txt/)
})

test('trennt Inhalt, Dateiname und Manifest', () => {
  const ort = paket({ 'protokoll.txt': `Fehler bei ${marke}` })
  mkdirSync(join(ort, `ordner-${marke}`))
  writeFileSync(join(ort, `ordner-${marke}`, 'leer.txt'), '')
  const orte = durchsucheVerzeichnis(ort, [marke]).map((f) => `${f.ort}:${f.datei}`).sort()
  assert.deepEqual(orte, [`dateiname:ordner-${marke}/leer.txt`, 'inhalt:protokoll.txt'])
})

test('trennt das Protokoll der Anwendung von dem des Systems', () => {
  const eintraege = [
    { processImagePath: '/Applications/MagicSetEditor4Mac.app/Contents/MacOS/MagicSetEditor4Mac', eventMessage: `Abruf mit ${marke}` },
    { processImagePath: '/usr/libexec/trustd', eventMessage: `fremd ${marke}` },
    { processImagePath: '/usr/libexec/trustd', eventMessage: 'nichts' },
  ]
  const { app, system } = ordneLog(eintraege, 'MagicSetEditor4Mac', [marke])
  assert.equal(app.length, 1)
  assert.equal(system.length, 1)
  assert.equal(app[0].art, 'roh')
})

// **Ein Socket ist keine Datei** — der Ort einer Probe trägt den Socket der
// Abnahmeleitung, und ihn zu lesen bricht ab („Unknown system error -102",
// erster Lauf von i05 am 2026-10-10).
test('überspringt, was keine gewöhnliche Datei ist', async () => {
  const ort = mkdtempSync(join(tmpdir(), 'geheimnisscan-'))
  writeFileSync(join(ort, 'zeile.txt'), `Fund ${marke}`)
  const dienst = createServer()
  await new Promise((fertig) => dienst.listen(join(ort, 'abnahme.sock'), fertig))

  try {
    assert.deepEqual(durchsucheVerzeichnis(ort, [marke]).map((f) => f.datei), ['zeile.txt'])
  } finally {
    await new Promise((fertig) => dienst.close(fertig))
  }
})
