#!/usr/bin/env node
// **Der Umzug der Kennung, im Gast** — R02
// (#2334),
// Vermerk 0191.
//
//     node scripts/kennunggast.mjs <altes Bündel> <neues Bündel> [<vm>]
//
// **Zwei Freigabebauten** (`scripts/bauen.sh` ohne Art): der Umzug betrifft
// die Kennung des ausgelieferten Bündels, und der Abnahmebau trägt eine eigene.
// Das alte entsteht in einem Arbeitsbaum ohne R02, das neue in diesem. **Nie
// am Wirt** — der Lauf lädt und entlädt Dienste und räumt Vorgaben.
//
// Der Treiber ist `scripts/huellenprobe/gast/kennung.pl`; was er prüft und
// was nicht, steht in seinem Kopf. Der Beleg geht nach
// `docs/belege/huellenprobe/kennung-<system>.md`.
import { execFileSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { fahre, gastziel, imGeteilten, schrittzeilen, wurzel } from './huellenprobe/gastlauf.mjs'

const [alt, neu, ziel] = process.argv.slice(2)
if (!alt || !neu || !existsSync(alt) || !existsSync(neu)) {
  console.error('Aufruf: node scripts/kennunggast.mjs <altes Bündel> <neues Bündel> [<vm>]')
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
if (!kennung(alt).startsWith('com.wails.') || kennung(neu) !== 'io.github.davidtheider.magicseteditor4mac') {
  console.error(`alt trägt ${kennung(alt)}, neu ${kennung(neu)} — erwartet com.wails.* und io.github.davidtheider.magicseteditor4mac`)
  process.exit(2)
}
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
const { lauf, ergebnis } = await imGeteilten('r02', async (geteilt) => {
  const ordner = join(geteilt, 'r02')
  execFileSync('mkdir', ['-p', join(ordner, 'alt'), join(ordner, 'neu')])
  execFileSync('cp', ['-cR', alt, join(ordner, 'alt/MagicSetEditor4Mac.app')])
  execFileSync('cp', ['-cR', neu, join(ordner, 'neu/MagicSetEditor4Mac.app')])
  execFileSync('cp', ['-c', join(wurzel, 'scripts/huellenprobe/gast/kennung.pl'), join(ordner, 'kennung.pl')])
  writeFileSync(join(ordner, 'dienst.plist'), dienstplist)
  const lauf = await fahre({ vm, geteilt, name: 'r02', treiber: 'kennung.pl', amWirt: false })
  const datei = join(geteilt, 'r02-ergebnis.json')

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
  `# Beleg: R02 — der Umzug der Kennung auf macOS ${g.system}`,
  '',
  `**Stand:** Commit \`${kopf}\` · **Datum:** ${beginn.toISOString().slice(0, 16).replace('T', ' ')} UTC · **Befehl:** \`node scripts/kennunggast.mjs <alt> <neu> ${vm}\` · **Rückgabe:** ${String(lauf.rueckgabe)}`,
  '',
  `**Maschine:** in der Lume-VM \`${vm}\` — macOS ${g.system} (${g.bau}), ${g.modell}.`,
  '',
  `**Bündel:** alt \`${kennung(alt)}\`, Programm \`${stand(alt).abdruck.slice(0, 12)}\`; neu \`${kennung(neu)}\`, Programm \`${stand(neu).abdruck.slice(0, 12)}\` — beide Freigabebauten, ohne Quarantäne eingesetzt (die Freigabe unter „Dennoch öffnen" gibt ein Mensch, §12).`,
  '',
  `**Ergebnis:** ${rot.length === 0 ? `alle ${String(ergebnis.schritte.length)} Schritte gelungen` : `**${String(rot.length)} von ${String(ergebnis.schritte.length)} Schritten gescheitert**`}.`,
  '',
  ...schrittzeilen(ergebnis.schritte, 'Teil'),
  '',
  '**Künstlich**: die Probevorgabe `R02Probe`, die Marke im Webspeicher und die plist des Dienstes unter dem alten Etikett — sie stehen da, wie das alte Fenster sie hinterließe; kein echter Bestand und kein Konto.',
  '',
  '**Was der Lauf nicht zeigt**: die neue Rückfrage von macOS für das lokale Netz und die Kontakte — sie erscheint erst, wenn die Anwendung danach fragt, und ein Fenster des Freigabebaus lässt sich im Gast nicht fernsteuern. Das bleibt eine Handprüfung am Fenster.',
  ...(ergebnis.grenzen?.length ? ['', '**Grenzen dieses Laufs:**', '', ...ergebnis.grenzen.map((f) => `- ${f}`)] : []),
  '',
]
const datei = join(wurzel, `docs/belege/huellenprobe/kennung-${g.system}.md`)
writeFileSync(datei, text.join('\n'))
console.log(text.join('\n'))
process.exit(rot.length || ergebnis.grenzen?.length ? 1 : 0)
