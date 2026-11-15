#!/usr/bin/env node
// **Der Aufstieg mit einem eingerichteten Dienst unter launchd, im Gast** —
// M02 (#2346).
//
//     node scripts/m02gast.mjs <Bündel von 0.8.0> <Kandidat> [<vm>]
//
// **Zwei Freigabebauten** (`scripts/bauen.sh` ohne Art): 0.8.0 aus dem Tag
// und der Kandidat. **Nie am Wirt** — ein LaunchAgent ist dauerhafte
// Einrichtung (`docs/FREIGABEN.md`), im Gast ist er Wegwerfware.
//
// Der Treiber ist `scripts/huellenprobe/gast/m02.pl`; was er prüft und was
// nicht, steht in seinem Kopf. Der Beleg geht nach
// `docs/belege/huellenprobe/m02-<system>.md`.
import { execFileSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { fahre, gastziel, imGeteilten, schrittzeilen, wurzel } from './huellenprobe/gastlauf.mjs'

const [alt, neu, ziel] = process.argv.slice(2)
if (!alt || !neu || !existsSync(alt) || !existsSync(neu)) {
  console.error('Aufruf: node scripts/m02gast.mjs <Bündel von 0.8.0> <Kandidat> [<vm>]')
  process.exit(2)
}
const { vm } = gastziel(ziel, { wirtErlaubt: false })

// kennung und stand lesen ein Bündel — **die Kennung entscheidet**, welches
// das alte ist; eine Verwechslung prüfte den Umzug rückwärts.
const kennung = (app) => execFileSync('plutil', ['-extract', 'CFBundleIdentifier', 'raw', '-o', '-', join(app, 'Contents/Info.plist')]).toString().trim()
// stand nennt die Bauschalter und den Abdruck des Programms — ein
// Freigabebau legt keinen `baustand.json` (`scripts/bauen.sh`).
const stand = (app) => {
  const programm = join(app, 'Contents/MacOS/MagicSetEditor4Mac')
  const schalter = execFileSync('go', ['version', '-m', programm]).toString().match(/-tags=(\S+)/)?.[1] ?? ''

  return { schalter, abdruck: createHash('sha256').update(readFileSync(programm)).digest('hex') }
}
// **Der Dienst von 0.8.0 läuft unter dessen Etikett** — auch mit einem
// Kandidaten, der eine neue Kennung trägt (R02): genau diese Lage muss der
// Wärter vor der Wanderung finden.
if (!kennung(alt).startsWith('com.wails.')) {
  console.error(`0.8.0 trägt ${kennung(alt)} — erwartet com.wails.*`)
  process.exit(2)
}
const etikett = `${kennung(alt)}.dienst`
for (const app of [alt, neu]) {
  if (stand(app).schalter !== 'production' || !existsSync(join(app, 'Contents/MacOS/mse-rest'))) {
    console.error(`${app} ist kein Freigabebau mit mse-rest — ./scripts/bauen.sh`)
    process.exit(2)
  }
}

// **Die plist, wie das Einrichten am alten Fenster sie schrieb**
// (`internal/app/schnittstellenziele.go`, `dienstplist`).
const dienstplist = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>__ETIKETT__</string>
	<key>ProgramArguments</key>
	<array>
		<string>__PROGRAMM__</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<dict>
		<key>SuccessfulExit</key>
		<false/>
	</dict>
	<key>StandardOutPath</key>
	<string>__ANSAGE__</string>
	<key>StandardErrorPath</key>
	<string>__ANSAGE__</string>
</dict>
</plist>
`

const beginn = new Date()
const { lauf, ergebnis } = await imGeteilten('m02', async (geteilt) => {
  const ordner = join(geteilt, 'm02')
  execFileSync('mkdir', ['-p', join(ordner, 'alt'), join(ordner, 'neu')])
  execFileSync('cp', ['-cR', alt, join(ordner, 'alt/MagicSetEditor4Mac.app')])
  execFileSync('cp', ['-cR', neu, join(ordner, 'neu/MagicSetEditor4Mac.app')])
  execFileSync('cp', ['-c', join(wurzel, 'scripts/huellenprobe/gast/m02.pl'), join(ordner, 'm02.pl')])
  execFileSync('cp', ['-cR', join(wurzel, 'saat/probe-werkstatt.mse-set'), join(ordner, 'probe-werkstatt.mse-set')])
  writeFileSync(join(ordner, 'dienst.plist'), dienstplist)
  writeFileSync(join(ordner, 'etikett.txt'), `${etikett}\n`)
  const lauf = await fahre({ vm, geteilt, name: 'm02', treiber: 'm02.pl', amWirt: false })
  const datei = join(geteilt, 'm02-ergebnis.json')

  return { lauf, ergebnis: existsSync(datei) ? JSON.parse(readFileSync(datei, 'utf8')) : null }
})
if (!ergebnis) {
  console.error(`Kein Ergebnis (Rückgabe ${String(lauf.rueckgabe)}):\n${lauf.ausgabe}`)
  process.exit(1)
}

const rot = ergebnis.schritte.filter((s) => !s.ok)
const g = ergebnis.maschine
const kopf = execFileSync('git', ['-C', wurzel, 'rev-parse', '--short', 'HEAD']).toString().trim()
const text = [
  `# Beleg: M02 — der Aufstieg mit eingerichtetem Dienst unter launchd, macOS ${g.system}`,
  '',
  `**Stand:** Commit \`${kopf}\` · **Datum:** ${beginn.toISOString().slice(0, 16).replace('T', ' ')} UTC · **Befehl:** \`node scripts/m02gast.mjs <0.8.0> <Kandidat> ${vm}\` · **Rückgabe:** ${String(lauf.rueckgabe)}`,
  '',
  `**Maschine:** in der Lume-VM \`${vm}\` — macOS ${g.system} (${g.bau}), ${g.modell}.`,
  '',
  `**Bündel:** 0.8.0 \`${kennung(alt)}\`, Programm \`${stand(alt).abdruck.slice(0, 12)}\`; Kandidat \`${kennung(neu)}\`, Programm \`${stand(neu).abdruck.slice(0, 12)}\` — beide Freigabebauten, ohne Quarantäne eingesetzt. **Der Dienst** unter \`${etikett}\`, wie das Fenster von 0.8.0 ihn einrichtete.`,
  '',
  `**Ergebnis:** ${rot.length === 0 ? `alle ${String(ergebnis.schritte.length)} Schritte gelungen` : `**${String(rot.length)} von ${String(ergebnis.schritte.length)} Schritten gescheitert**`}.`,
  '',
  ...schrittzeilen(ergebnis.schritte, 'Teil'),
  '',
  `**Wanderungen:** vor dem Aufstieg ${String(ergebnis.wanderungen?.vorher ?? '?')}, danach ${String(ergebnis.wanderungen?.nachher ?? '?')}, zuletzt \`${ergebnis.wanderungen?.zuletzt ?? '?'}\`.`,
  '',
  '**Künstlich**: das Probeset der Saat und die plist des Dienstes, wie das Einrichten am Fenster sie schreibt — kein echter Bestand und kein Konto. Die Ansage des Dienstes steht unten, die Marke geschwärzt.',
  '',
  '```',
  ...(ergebnis.protokoll ?? '').split('\n'),
  '```',
  '',
  '**Was der Lauf nicht zeigt**: einen Dienst, der **selbst** schon die Fassungswache kennt und über einem noch neueren Bestand endet — das belegt `internal/flaechen/neuererbestand_test.go` mit echten Prozessen, nicht unter launchd.',
  ...(ergebnis.grenzen?.length ? ['', '**Grenzen dieses Laufs:**', '', ...ergebnis.grenzen.map((f) => `- ${f}`)] : []),
  '',
]
const datei = join(wurzel, `docs/belege/huellenprobe/m02-${g.system}.md`)
writeFileSync(datei, text.join('\n'))
console.log(text.join('\n'))
process.exit(rot.length || ergebnis.grenzen?.length ? 1 : 0)
