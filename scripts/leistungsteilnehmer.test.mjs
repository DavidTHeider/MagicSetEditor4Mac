// Prüfungen von scripts/leistungsteilnehmer.mjs — QA-04.AK-01 und AK-02
// (#1735).
//
// **Welcher Prozess mit welcher Eingabe** — gefahren wird nichts; `messe`
// hält fest, was es bekommen hätte, und antwortet, was der Test vorgibt.

import { test } from 'node:test'
import assert from 'node:assert/strict'
import { existsSync, mkdtempSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { mse2, mse4mac } from './leistungsteilnehmer.mjs'

/** aufzeichnung gibt ein `messe`, das jeden Aufruf festhält und `aus` antwortet. */
function aufzeichnung(aus = '') {
  const aufrufe = []
  const messe = async (befehl, argumente, optionen) => {
    aufrufe.push({ befehl, argumente, ...optionen })
    return { wandMs: 1, fehler: null, aus }
  }
  return { aufrufe, messe }
}

test('MSE2 bekommt ein Zuhause, das auch Core Foundation sieht — samt dem Ordner für mse.config', async () => {
  const arbeit = mkdtempSync(join(tmpdir(), 'teilnehmer-'))
  const { aufrufe, messe } = aufzeichnung()
  await mse2({ binaer: '/mse2/magicseteditor', arbeit, messe }).start()
  const [a] = aufrufe
  assert.deepEqual(a.argumente, ['--cli', '--quiet', '--raw'])
  assert.equal(a.eingabe, ':quit\n')
  assert.equal(a.umgebung.CFFIXED_USER_HOME, a.umgebung.HOME, 'HOME allein lenkt wxWidgets auf macOS nicht um')
  assert.ok(a.umgebung.HOME.startsWith(arbeit))
  assert.ok(existsSync(join(a.umgebung.HOME, 'Library/Application Support/magicseteditor')))
})

test('MSE2 öffnet über :info und liest die Kartenzahl daraus', async () => {
  const { aufrufe, messe } = aufzeichnung('0\n4\nset:\nfilename: /x\nrelative: x\n#cards:   70\n0\n0\n')
  const m = await mse2({ binaer: '/m', arbeit: mkdtempSync(join(tmpdir(), 'teilnehmer-')), messe }).oeffnen({ pfad: '/k/A.mse-set' })
  assert.deepEqual(aufrufe[0].argumente, ['--cli', '/k/A.mse-set', '--quiet', '--raw'])
  assert.equal(aufrufe[0].eingabe, ':info\n:quit\n')
  assert.equal(m.karten, 70)
})

test('MSE2 sichert mit write_set_file in das Verzeichnis des Ziels', async () => {
  const { aufrufe, messe } = aufzeichnung()
  await mse2({ binaer: '/m', arbeit: mkdtempSync(join(tmpdir(), 'teilnehmer-')), messe }).sichern({ pfad: '/k/A.mse-set' }, '/aus/mse2-0.mse-set')
  assert.equal(aufrufe[0].ort, '/aus')
  assert.equal(aufrufe[0].eingabe, 'write_set_file(file:"mse2-0.mse-set")\n:quit\n')
})

test('MSE2 zeichnet alle Bilder und führt Vorlagen über seine eigenen Schalter aus', async () => {
  const { aufrufe, messe } = aufzeichnung()
  const t = mse2({ binaer: '/m', arbeit: mkdtempSync(join(tmpdir(), 'teilnehmer-')), messe })
  await t.bilder({ pfad: '/k/A.mse-set' }, '/aus')
  await t.ausfuhr({ pfad: '/k/A.mse-set' }, 'magic-namecheck.mse-export-template', '/aus/a.txt')
  assert.deepEqual(aufrufe[0].argumente, ['--export-images', '/k/A.mse-set', '/aus/{card.name}.png'])
  assert.deepEqual(aufrufe[1].argumente, ['--export', 'magic-namecheck.mse-export-template', '/k/A.mse-set', '/aus/a.txt'])
})

test('MSE4Mac arbeitet an einem eigenen Bestand, im eigenen Zuhause, ohne den Anmeldebund', async () => {
  const { aufrufe, messe } = aufzeichnung('{"sets":[]}')
  await mse4mac({ name: 'heute', verzeichnis: '/bin', zuhause: '/z', messe }).start('/b/bestand.db')
  const [a] = aufrufe
  assert.equal(a.befehl, '/bin/mse')
  assert.deepEqual(a.argumente, ['sets', '--json'])
  assert.equal(a.umgebung.HOME, '/z')
  assert.equal(a.umgebung.MSE_BESTAND, '/b/bestand.db')
  assert.equal(a.umgebung.MSE_SCHLUESSELBUND, '/z/keiner.keychain')
})

test('MSE4Mac: Öffnen nennt Kennung und Kartenzahl, Wiederöffnen zählt die Liste', async () => {
  const { messe } = aufzeichnung('{"setID":"7","karten":70,"befunde":[]}')
  const t = mse4mac({ name: 'heute', verzeichnis: '/bin', zuhause: '/z', messe })
  const m = await t.oeffnen({ pfad: '/k/A.mse-set' }, '/b/bestand.db')
  assert.equal(m.setID, '7')
  assert.equal(m.karten, 70)
  const w = await mse4mac({ name: 'heute', verzeichnis: '/bin', zuhause: '/z', messe: aufzeichnung('{"karten":[{},{}]}').messe })
    .wiederoeffnen({ id: '7' }, '/b/bestand.db')
  assert.equal(w.karten, 2)
})

test('MSE4Mac sichert und führt Vorlagen über mse aus — mit der Kennung aus dem Bestand', async () => {
  const { aufrufe, messe } = aufzeichnung()
  const t = mse4mac({ name: 'v050', verzeichnis: '/bin', zuhause: '/z', messe })
  await t.sichern({ id: '7' }, '/aus/x.mse-set', '/b/bestand.db')
  await t.ausfuhr({ id: '7' }, 'magic-namecheck.mse-export-template', '/aus/a.txt', '/b/bestand.db')
  assert.deepEqual(aufrufe[0].argumente, ['sichern', '7', '/aus/x.mse-set'])
  assert.deepEqual(aufrufe[1].argumente, ['vorlagen', '7', 'ausfuehren', 'magic-namecheck.mse-export-template', '/aus/a.txt'])
})

test('MSE4Mac schaltet die Kontingente im Bestand ab und wieder an — der Langlauf misst die Maschine, die Lastgrenze die Drossel', async () => {
  const { aufrufe, messe } = aufzeichnung()
  const t = mse4mac({ name: 'heute', verzeichnis: '/bin', zuhause: '/z', messe })
  const m = await t.kontingente('/b/vorrat.db', 'aus')
  await t.kontingente('/b/vorrat.db', 'an')
  assert.equal(m.fehler, null)
  assert.equal(aufrufe[0].befehl, '/bin/mse')
  assert.deepEqual(aufrufe[0].argumente, ['einstellungen', 'setzen', 'grenzen.kontingente', 'aus'])
  assert.deepEqual(aufrufe[1].argumente, ['einstellungen', 'setzen', 'grenzen.kontingente', 'an'])
  assert.equal(aufrufe[0].umgebung.MSE_BESTAND, '/b/vorrat.db')
})

test('MSE4Mac liest den Stand der Kontingente aus den Einstellungen desselben Bestands', async () => {
  const { aufrufe, messe } = aufzeichnung('{"einstellungen":[{"schluessel":"grenzen.kontingente","wert":"an"}]}')
  const t = mse4mac({ name: 'heute', verzeichnis: '/bin', zuhause: '/z', messe })
  const m = await t.kontingenteStand('/b/vorrat.db')
  assert.equal(m.aus, '{"einstellungen":[{"schluessel":"grenzen.kontingente","wert":"an"}]}')
  assert.equal(aufrufe[0].befehl, '/bin/mse')
  assert.deepEqual(aufrufe[0].argumente, ['einstellungen', '--json'])
  assert.equal(aufrufe[0].umgebung.MSE_BESTAND, '/b/vorrat.db')
})
