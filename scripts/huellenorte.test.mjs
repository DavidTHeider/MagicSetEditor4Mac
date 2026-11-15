// Prüfungen von scripts/huellenorte.mjs — T01
// (#2095).
//
// **Ein Abdruck sieht, was ein Lauf an einem Ort verändert hat** — bis zur
// zweiten Ebene, nach Name, Größe und Zeit. Er ist der positive Beleg, dass
// die echten Orte dieses Rechners unberührt blieben; ohne ihn sagt ein
// grüner Lauf nur, dass er nicht nachgesehen hat.

import { test } from 'node:test'
import assert from 'node:assert/strict'
import { chmodSync, mkdirSync, mkdtempSync, rmSync, utimesSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { abdruck, abweichungen, anwendungLaeuft, echteOrte, kindumgebung, kontrollprobe, unlesbare } from './huellenorte.mjs'

const raum = () => {
  const ort = mkdtempSync(join(tmpdir(), 'huellenorte-'))
  mkdirSync(join(ort, 'a/innen'), { recursive: true })
  writeFileSync(join(ort, 'a/bestand.db'), 'eins')
  writeFileSync(join(ort, 'a/innen/datei'), 'zwei')

  return ort
}

test('ohne Änderung gibt es keine Abweichung', () => {
  const ort = raum()
  const vorher = abdruck([join(ort, 'a'), join(ort, 'gibt-es-nicht')])
  assert.deepEqual(abweichungen(vorher, abdruck([join(ort, 'a'), join(ort, 'gibt-es-nicht')])), [])
  rmSync(ort, { recursive: true })
})

test('eine neue, eine geänderte und eine fehlende Datei werden genannt', () => {
  const ort = raum()
  const orte = [join(ort, 'a')]
  const vorher = abdruck(orte)
  writeFileSync(join(ort, 'a/neu'), 'x')
  writeFileSync(join(ort, 'a/bestand.db'), 'eins und mehr')
  rmSync(join(ort, 'a/innen/datei'))
  const gefunden = abweichungen(vorher, abdruck(orte)).join('\n')
  assert.match(gefunden, /neu: .*a\/neu/)
  assert.match(gefunden, /verändert: .*a\/bestand\.db/)
  assert.match(gefunden, /fehlt: .*a\/innen\/datei/)
  rmSync(ort, { recursive: true })
})

test('eine bloß neu geschriebene Zeit zählt auch', () => {
  const ort = raum()
  const orte = [join(ort, 'a')]
  const vorher = abdruck(orte)
  const spaeter = new Date(Date.now() + 60_000)
  utimesSync(join(ort, 'a/innen/datei'), spaeter, spaeter)
  assert.match(abweichungen(vorher, abdruck(orte)).join('\n'), /verändert: .*innen\/datei/)
  rmSync(ort, { recursive: true })
})

test('ein Ort, den es vorher nicht gab, wird genannt, wenn er entsteht', () => {
  const ort = raum()
  const orte = [join(ort, 'spaeter')]
  const vorher = abdruck(orte)
  mkdirSync(join(ort, 'spaeter'))
  assert.match(abweichungen(vorher, abdruck(orte)).join('\n'), /neu: .*spaeter/)
  rmSync(ort, { recursive: true })
})

test('auch die tiefste Ebene und versteckte Dateien werden gesehen', () => {
  // **A01 (#2215)**: bis dahin reichte der Abdruck zwei Ebenen tief, und
  // eine Datei in `pakete/<paket>/<teil>/` blieb unsichtbar. Gezählt hat der
  // Audit zu 0.8.0 62 729 Einträge unter den echten Orten — viele davon
  // tiefer.
  const ort = raum()
  mkdirSync(join(ort, 'a/innen/tief/tiefer'), { recursive: true })
  const orte = [join(ort, 'a')]
  const vorher = abdruck(orte)
  writeFileSync(join(ort, 'a/innen/tief/tiefer/datei'), 'x')
  assert.match(abweichungen(vorher, abdruck(orte)).join('\n'), /neu: .*tiefer\/datei$/m)
  const zwischen = abdruck(orte)
  writeFileSync(join(ort, 'a/innen/tief/tiefer/.versteckt'), 'x')
  assert.match(abweichungen(zwischen, abdruck(orte)).join('\n'), /neu: .*\.versteckt$/m)
  rmSync(ort, { recursive: true })
})

test('die echten Orte sind alle der Anwendung unter dem gegebenen Zuhause', () => {
  // **Auch die, die macOS unter der Kennung des Bündels führt** (A01): der
  // Audit sah die Vorgaben des Dateidialogs in
  // `Preferences/com.wails.magicseteditor4mac.plist` wachsen, und der Abdruck
  // davor kannte nur die zwei Orte der Anwendung selbst.
  assert.deepEqual(echteOrte('/Users/jemand'), [
    '/Users/jemand/Library/Application Support/MagicSetEditor4Mac',
    '/Users/jemand/Library/Caches/MagicSetEditor4Mac',
    // **Unter der neuen Kennung und unter der alten** (R02, #2334): der
    // Umzug liest dort, und auch dort schreibt keine Probe.
    ...['io.github.davidtheider.magicseteditor4mac', 'com.wails.magicseteditor4mac'].flatMap((k) => [
      `/Users/jemand/Library/Caches/${k}`,
      `/Users/jemand/Library/Preferences/${k}.plist`,
      `/Users/jemand/Library/WebKit/${k}`,
      `/Users/jemand/Library/HTTPStorages/${k}`,
      `/Users/jemand/Library/Saved Application State/${k}.savedState`,
    ]),
  ])
})

test('eine einzelne Datei als Ort wird selbst abgedruckt', () => {
  const ort = raum()
  const datei = join(ort, 'a/bestand.db')
  const vorher = abdruck([datei])
  assert.equal(vorher.size, 1)
  writeFileSync(datei, 'eins und mehr')
  assert.match(abweichungen(vorher, abdruck([datei])).join('\n'), /verändert: .*bestand\.db/)
  rmSync(ort, { recursive: true })
})

test('die Positivkontrolle sieht die absichtlich veränderte Datei', () => {
  // **Ohne sie sagt ein leerer Vergleich nur, dass er nichts gefunden hat**
  // — nicht, dass er hätte finden können (A01, §10 „Gegenprobe").
  const ergebnis = kontrollprobe()
  assert.equal(ergebnis.erkannt, true)
  assert.match(ergebnis.gefunden.join('\n'), /verändert: .*kontrolle/)
})

test('ob die Anwendung dieses Rechners läuft, sagt ein eigener Befund', () => {
  // **Läuft sie nebenher, verändert sie die Orte selbst** — dann ist ein
  // Unterschied kein Beleg gegen die Probe, und keiner kein Beleg für sie.
  assert.equal(typeof anwendungLaeuft(), 'boolean')
  assert.equal(anwendungLaeuft(() => 'ASN:0x0-0x1234:'), true)
  assert.equal(anwendungLaeuft(() => ''), false)
})

test('ein Kindprozess bekommt Zuhause, Bestand und Bund seines Orts', () => {
  const umgebung = kindumgebung({ heim: '/tmp/h', bestand: '/tmp/b.db', bund: '/tmp/k' })
  assert.equal(umgebung.HOME, '/tmp/h')
  assert.equal(umgebung.MSE_BESTAND, '/tmp/b.db')
  assert.equal(umgebung.MSE_SCHLUESSELBUND, '/tmp/k')
  assert.equal(umgebung.PATH, process.env.PATH)
})

test('ein Verzeichnis ohne Leserecht wird als unlesbar genannt und nicht übergangen', () => {
  // **A01, Zweitblick über Phase A**: ein verschluckter Lesefehler gab zwei
  // gleich unvollständige Abdrücke — und damit „unberührt". Eine Datei
  // darin zu überschreiben ändert die Zeit des Verzeichnisses nicht.
  const ort = raum()
  const gesperrt = join(ort, 'a/innen')
  chmodSync(gesperrt, 0o000)
  try {
    const gelesen = abdruck([join(ort, 'a')])
    assert.deepEqual(unlesbare(gelesen), [gesperrt])
  } finally {
    chmodSync(gesperrt, 0o700)
    rmSync(ort, { recursive: true })
  }
})

test('ein lesbarer Abdruck nennt nichts Unlesbares', () => {
  const ort = raum()
  assert.deepEqual(unlesbare(abdruck([join(ort, 'a')])), [])
  rmSync(ort, { recursive: true })
})

test('ein Ort hinter einem gesperrten Verzeichnis ist unlesbar und nicht abwesend', () => {
  // **Restbefund des Nachlaufs über Phase A**: `existsSync` sagt auch dann
  // „nicht da", wenn der Weg dorthin gesperrt ist — und der Ort fiel still
  // aus dem Abdruck.
  const ort = raum()
  const davor = join(ort, 'a')
  const dahinter = join(davor, 'innen')
  chmodSync(davor, 0o000)
  try {
    assert.deepEqual(unlesbare(abdruck([dahinter])), [dahinter])
  } finally {
    chmodSync(davor, 0o700)
    rmSync(ort, { recursive: true })
  }
})

test('ein Ort, den es nicht gibt, ist abwesend und nicht unlesbar', () => {
  const ort = raum()
  assert.deepEqual([...abdruck([join(ort, 'gibt-es-nicht')])], [])
  rmSync(ort, { recursive: true })
})
