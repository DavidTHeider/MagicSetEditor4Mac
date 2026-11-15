#!/usr/bin/env node
// **Trägt ein selbst ausgestellter Herausgeber die Bindung im Schlüsselbund
// über eine Fassung?** — I01 (#2259).
//
//     node scripts/bindunggast.mjs [<vm>]
//
// Übersetzt `scripts/huellenprobe/gast/leser.swift` am Wirt in zwei Fassungen
// und `fremd.c` zu einer Bibliothek, die E5 in einen Leser lädt, und fährt `scripts/huellenprobe/gast/bindung.pl` im Gast; was der Treiber
// misst und was nicht, steht in seinem Kopf. **Nie am Wirt** — der Treiber
// richtet Herausgeberzeugnisse ein, und das geschieht nur in einem
// Wegwerf-Gast. Der Beleg geht nach `docs/belege/i01-bindung.md`.
import { execFileSync } from 'node:child_process'
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { fahre, gastziel, imGeteilten, wurzel } from './huellenprobe/gastlauf.mjs'

// **Nie am Wirt** — `gastziel` weist `--wirt` und jeden Namen ab, der kein Gast ist.
const { vm } = gastziel(process.argv[2], { wirtErlaubt: false })

const beginn = new Date()
// **Das geteilte Verzeichnis geht in jedem Fall weg** — samt Vorbereitung,
// nach einer Ausnahme und nach einem Signal (`imGeteilten`; Befund 2 des
// Zweitblicks und Befund 1 des Nachlaufs über Phase I).
const { lauf, ergebnis } = await imGeteilten('i01', async (geteilt) => {
  const ordner = join(geteilt, 'i01')
  mkdirSync(ordner)
  const quelle = join(wurzel, 'scripts/huellenprobe/gast/leser.swift')
  execFileSync('swiftc', ['-O', quelle, '-o', join(ordner, 'leser-v1')], { stdio: 'ignore' })
  execFileSync('swiftc', ['-O', '-D', 'FASSUNG_ZWEI', quelle, '-o', join(ordner, 'leser-v2')], { stdio: 'ignore' })
  execFileSync('xcrun', ['clang', '-Wall', '-Wextra', '-Werror', '-dynamiclib', '-target', 'arm64-apple-macos15.2', '-framework', 'Security',
    '-framework', 'CoreFoundation', join(wurzel, 'scripts/huellenprobe/gast/fremd.c'), '-o', join(ordner, 'fremd.dylib')], { stdio: 'ignore' })
  execFileSync('codesign', ['--force', '--sign', '-', join(ordner, 'fremd.dylib')], { stdio: 'ignore' })
  execFileSync('cp', ['-c', join(wurzel, 'scripts/huellenprobe/gast/bindung.pl'), join(ordner, 'bindung.pl')])
  const lauf = await fahre({ vm, geteilt, name: 'i01', treiber: 'bindung.pl' })
  return { lauf, ergebnis: existsSync(join(geteilt, 'i01-ergebnis.json')) ? JSON.parse(readFileSync(join(geteilt, 'i01-ergebnis.json'), 'utf8')) : null }
})
const { rueckgabe, ausgabe } = lauf
if (!ergebnis) {
  console.error(`Kein Ergebnis (Rückgabe ${String(rueckgabe)}):\n${ausgabe}`)
  process.exit(1)
}

const kopf = execFileSync('git', ['-C', wurzel, 'rev-parse', '--short', 'HEAD']).toString().trim()
const g = ergebnis.maschine
const abweichend = ergebnis.schritte.filter((s) => !s.ok)
const text = [
  '# Beleg: I01 — trägt ein eigener Herausgeber die Bindung über eine Fassung?',
  '',
  `**Stand:** Commit \`${kopf}\` · **Datum:** ${beginn.toISOString().slice(0, 16).replace('T', ' ')} UTC · **Befehl:** \`node scripts/bindunggast.mjs ${vm}\` · **Rückgabe:** ${String(rueckgabe)}`,
  '',
  `**Maschine:** Lume-VM \`${vm}\` — macOS ${g.system} (${g.bau}), ${g.openssl}. **Nur ein künstlicher Bund** mit Probekennwort, kurz vorn in der Suchliste des Gastes; zwei selbst ausgestellte Herausgeber (A, B), im selben Bund und mit ihm gelöscht. **Auf dem Wirt wurde kein Herausgeberzeugnis eingerichtet.** Aufgeräumt: ${ergebnis.aufgeraeumt}.`,
  '',
  `**Ergebnis:** ${abweichend.length === 0 ? `jede der ${String(ergebnis.schritte.length)} Erwartungen trifft` : `**${String(abweichend.length)} von ${String(ergebnis.schritte.length)} Erwartungen treffen nicht**`}.`,
  '',
  '| Fall | Leser | erwartet | gemessen | trifft |',
  '| --- | --- | --- | --- | --- |',
  ...ergebnis.schritte.map((s) => `| ${s.fall} | ${s.wer} | ${s.soll} | ${s.ist} | ${s.ok ? 'ja' : '**nein**'} |`),
  '',
  '**Die Anforderungen der Fassungen** (`codesign -d -r-`, dazu die Signaturflags — `runtime` heißt gehärtet):',
  '',
  ...Object.entries(ergebnis.anforderung ?? {}).map(([name, anforderung]) => `- \`${name}\` (cdhash \`${String(ergebnis.cdhash?.[name] ?? '?').slice(0, 12)}…\`, flags \`${String(ergebnis.flags?.[name] ?? '?')}\`): \`${anforderung}\``),
  '',
  '**Gatekeeper** (`spctl --assess --type execute`), getrennt gemessen:',
  '',
  ...Object.entries(ergebnis.gatekeeper ?? {}).map(([name, urteil]) => `- \`${name}\`: ${String(urteil).trim()}`),
  '',
  `**Die ACL des gebundenen Eintrags** (\`security dump-keychain -a\`, gekürzt): \`${String(ergebnis.acl ?? '—').replace(/\s+/g, ' ').slice(0, 400)}\``,
  ...(ergebnis.grenzen?.length ? ['', '**Grenzen dieses Laufs:**', '', ...ergebnis.grenzen.map((f) => `- ${f}`)] : []),
  '',
  '## Das Protokoll',
  '',
  '```text',
  ...ergebnis.protokoll,
  '```',
  '',
]
writeFileSync(join(wurzel, 'docs/belege/i01-bindung.md'), text.join('\n'))
console.log(text.slice(0, 24).join('\n'))
// **Vollständig ist der Lauf, wenn auch aufgeräumt ist** — dieselbe Regel
// wie bei der Quick-Look-Probe (Befund 7 des Zweitblicks über Phase I).
if (ergebnis.aufgeraeumt !== 'ja') console.error('unvollständig: nicht aufgeräumt')
process.exit(abweichend.length || ergebnis.grenzen?.length || ergebnis.aufgeraeumt !== 'ja' ? 1 : 0)
