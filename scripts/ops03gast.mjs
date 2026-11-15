#!/usr/bin/env node
// **Alles widerrufen unter launchd, wenn das Programm schon fehlt, im Gast** —
// OPS03 (#2349).
//
//     node scripts/ops03gast.mjs [<vm>]
//
// **Nie am Wirt** — ein LaunchAgent ist dauerhafte Einrichtung
// (`docs/FREIGABEN.md`), im Gast ist er Wegwerfware. Das Bündel ist der
// Abnahmebau dieses Stands (`./scripts/bauen.sh abnahme`); die Einrichtung
// schreibt `scripts/ops03einrichtung.go` mit dem Code der Anwendung, der
// Treiber im Gast ist `scripts/huellenprobe/gast/ops03.pl` — was er prüft,
// steht in seinem Kopf. Der Beleg geht nach
// `docs/belege/huellenprobe/ops03-<system>.md`.
import { execFileSync } from 'node:child_process'
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { abnahmestand, buendel, fahre, gastziel, imGeteilten, schrittzeilen, wurzel } from './huellenprobe/gastlauf.mjs'

const { vm } = gastziel(process.argv[2], { wirtErlaubt: false })
const stand = abnahmestand(vm)

// **Der Werkzeugort im Gast**: dort setzt der Treiber das Bündel ein.
const werkzeug = '/Users/lume/Applications/MagicSetEditor4Mac.app/Contents/MacOS'

const beginn = new Date()
const { lauf, ergebnis, eingerichtet } = await imGeteilten('ops03', async (geteilt) => {
  const ordner = join(geteilt, 'ops03')
  const heim = join(ordner, 'heim')
  mkdirSync(heim, { recursive: true })
  execFileSync('cp', ['-cR', buendel, join(ordner, 'MagicSetEditor4Mac.app')])
  execFileSync('cp', ['-c', join(wurzel, 'scripts/huellenprobe/gast/ops03.pl'), join(ordner, 'ops03.pl')])
  const eingerichtet = execFileSync('go', ['run', 'scripts/ops03einrichtung.go', heim, werkzeug], { cwd: wurzel, encoding: 'utf8' })
  writeFileSync(join(ordner, 'heim.txt'), `${heim}\n`)
  writeFileSync(join(ordner, 'werkzeug.txt'), `${werkzeug}\n`)
  const lauf = await fahre({ vm, geteilt, name: 'ops03', treiber: 'ops03.pl', amWirt: false })
  const datei = join(geteilt, 'ops03-ergebnis.json')

  return { lauf, eingerichtet, ergebnis: existsSync(datei) ? JSON.parse(readFileSync(datei, 'utf8')) : null }
})
if (!ergebnis) {
  console.error(`Kein Ergebnis (Rückgabe ${String(lauf.rueckgabe)}):\n${lauf.ausgabe}`)
  process.exit(1)
}

const rot = ergebnis.schritte.filter((s) => !s.ok)
const g = ergebnis.maschine
const kopf = execFileSync('git', ['-C', wurzel, 'rev-parse', '--short', 'HEAD']).toString().trim()
const text = [
  `# Beleg: OPS03 — alles widerrufen unter launchd, wenn das Programm schon fehlt, macOS ${g.system}`,
  '',
  `**Stand:** Commit \`${kopf}\` · **Datum:** ${beginn.toISOString().slice(0, 16).replace('T', ' ')} UTC · **Befehl:** \`node scripts/ops03gast.mjs ${vm}\` · **Rückgabe:** ${String(lauf.rueckgabe)}`,
  '',
  `**Maschine:** in der Lume-VM \`${vm}\` — macOS ${g.system} (${g.bau}), ${g.modell}.`,
  '',
  `**Bündel:** der Abnahmebau dieses Stands, Bauabdruck \`${stand.abdruck.slice(0, 12)}\`, ohne Quarantäne eingesetzt — einmal unter \`~/Applications\`, nach dem Papierkorb ein zweites Mal unter \`~/Downloads/MagicSetEditor4Mac 2.app\`.`,
  '',
  `**Ergebnis:** ${rot.length === 0 ? `alle ${String(ergebnis.schritte.length)} Schritte gelungen` : `**${String(rot.length)} von ${String(ergebnis.schritte.length)} Schritten gescheitert**`}.`,
  '',
  ...schrittzeilen(ergebnis.schritte, 'Teil'),
  '',
  '**Eingerichtet am Wirt, mit dem Code der Anwendung** (`scripts/ops03einrichtung.go`, `App.SchnittstellenAlleEinrichten` mit eingesetztem Befehlsweg) — je Ziel der Ausgang:',
  '',
  '```',
  ...eingerichtet.trim().split('\n'),
  '```',
  '',
  '**Künstlich**: ein Wegwerfzuhause mit je einem fremden Nachbarn — kein echter Bestand, kein Konto, kein `claude`.',
  '',
  '```',
  ...ergebnis.ausgaben.join('\n').trim().split('\n'),
  '```',
  '',
  '**Was der Lauf nicht zeigt**: Claude Code — ohne `claude` im Gast bleibt sein Eintrag stehen, so steht es in der Administration; und das Einrichten über das Fenster selbst, das die Hüllenprobe `ops03` am Wirt belegt, dort ohne `launchctl`.',
  '',
]
const datei = join(wurzel, `docs/belege/huellenprobe/ops03-${g.system}.md`)
writeFileSync(datei, text.join('\n'))
console.log(text.join('\n'))
process.exit(rot.length ? 1 : 0)
