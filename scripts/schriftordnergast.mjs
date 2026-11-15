#!/usr/bin/env node
// **Aktiviert macOS Schriften in Unterordnern von `~/Library/Fonts`?** — die
// Klärung #2411,
// im Gast: am Wirt wäre ein Unterordner in `~/Library/Fonts` ein Eingriff in
// die Einstellungen des Menschen davor.
//
//     node scripts/schriftordnergast.mjs [<vm>]
//
// Der Treiber ist `scripts/huellenprobe/gast/schriftordner.pl`. Die Schrift
// kommt aus dem Pack (`MSE_PROBE_PACK`, Vorgabe `~/full-magic-pack`) und geht
// nur in den Wegwerf-Gast. Der Beleg geht nach
// `docs/belege/huellenprobe/schriftordner-<system>.md`.
import { execFileSync } from 'node:child_process'
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { homedir } from 'node:os'
import { join } from 'node:path'

import { fahre, gastziel, imGeteilten, schrittzeilen, wurzel } from './huellenprobe/gastlauf.mjs'

const { vm } = gastziel(process.argv[2], { wirtErlaubt: false })
const pack = process.env.MSE_PROBE_PACK ?? join(homedir(), 'full-magic-pack')
const schriften = ['vanilla whale.ttf', 'MagicMedieval.ttf'].map((name) => join(pack, 'Magic - Fonts', name))
for (const schrift of schriften) {
  if (!existsSync(schrift)) {
    console.error(`${schrift} fehlt`)
    process.exit(2)
  }
}

const beginn = new Date()
const { lauf, ergebnis } = await imGeteilten('schrift', async (geteilt) => {
  const ordner = join(geteilt, 'schrift')
  mkdirSync(ordner, { recursive: true })
  for (const schrift of schriften) execFileSync('cp', ['-c', schrift, ordner])
  execFileSync('cp', ['-c', join(wurzel, 'scripts/huellenprobe/gast/schriftordner.pl'), join(ordner, 'schriftordner.pl')])
  const lauf = await fahre({ vm, geteilt, name: 'schrift', treiber: 'schriftordner.pl', amWirt: false })
  const datei = join(geteilt, 'schrift-ergebnis.json')

  return { lauf, ergebnis: existsSync(datei) ? JSON.parse(readFileSync(datei, 'utf8')) : null }
})
if (!ergebnis) {
  console.error(`Kein Ergebnis (Rückgabe ${String(lauf.rueckgabe)}):\n${lauf.ausgabe}`)
  process.exit(1)
}

const rot = ergebnis.schritte.filter((s) => !s.ok)
const kopf = execFileSync('git', ['-C', wurzel, 'rev-parse', '--short', 'HEAD']).toString().trim()
const text = [
  `# Beleg: #2411 — Schriften in Unterordnern von ~/Library/Fonts, macOS ${ergebnis.maschine.system}`,
  '',
  `**Stand:** Commit \`${kopf}\` · **Datum:** ${beginn.toISOString().slice(0, 16).replace('T', ' ')} UTC · **Befehl:** \`node scripts/schriftordnergast.mjs ${vm}\` · **Rückgabe:** ${String(lauf.rueckgabe)}`,
  '',
  `**Maschine:** in der Lume-VM \`${vm}\` — macOS ${ergebnis.maschine.system} (${ergebnis.maschine.bau}).`,
  '',
  `**Ergebnis:** ${ergebnis.unterordner ? '**macOS aktiviert eine Schrift im Unterordner**' : '**macOS aktiviert eine Schrift im Unterordner nicht**'}; die Gegenprobe flach ${rot.length === 0 ? 'gelang' : '**scheiterte**'}.`,
  '',
  ...schrittzeilen(ergebnis.schritte, 'Teil'),
  '',
  '**Gefragt** wurde AppKit (`NSFontManager.availableFonts`) über JXA, bis zu 20 s je Lage. **Nicht gefragt**: Font Book und CoreText unmittelbar; eine Schrift, die AppKit nicht nennt, steht in keinem Schriftmenü.',
  '',
]
mkdirSync(join(wurzel, 'docs/belege/huellenprobe'), { recursive: true })
writeFileSync(join(wurzel, `docs/belege/huellenprobe/schriftordner-${ergebnis.maschine.system}.md`), text.join('\n'))
console.log(text.join('\n'))
process.exit(rot.length ? 1 : 0)
