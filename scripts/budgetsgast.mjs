#!/usr/bin/env node
// **B02 in der VM** — der Gastlauf zu `scripts/huellenprobe/budgets.mjs`
// (#2221):
// „VM/Host getrennt" verlangt eine eigene Reihe, und Node gibt es im Gast
// nicht. Gefahren wird `scripts/huellenprobe/gast/budgets.pl`.
//
//     ./scripts/bauen.sh abnahme && node scripts/budgetsgast.mjs [<vm>]
//
// # Was hier geschieht
//
// 1. Das Bündel muss aus diesem Stand gebaut sein — wie bei der Hüllenprobe.
// 2. Ein geteiltes Verzeichnis unter `$TMPDIR` bekommt Bündel, Saatset,
//    `Black.mse-set` und den Pack — als **Klone** (`cp -c`), auf APFS ohne
//    Platz — und `messungen.json`: die JavaScript-Stücke **aus dem Drehbuch**.
// 3. Die VM startet mit diesem Verzeichnis, ohne Anzeige (`MSE_GAST_ANZEIGE`
//    setzt eine); `lume ssh` fährt den Perl-Treiber; danach endet die VM mit
//    `TERM` (`lume stop` hing).
//
// **Was der Gast nicht misst**: alles am Bild. Er steht nach dem Start hinter
// seinem Sperrbildschirm, und dahinter zeichnet WebKit nicht — gemessen am
// 2026-10-09: null `requestAnimationFrame` in zwei Sekunden, mit und ohne
// Anzeige, bei laufenden Zeitgebern (`budgets-gast/sperrbildschirm.jpg`).
// Entsperren hieße ein Kennwort eingeben, und das geschieht nicht (§17). Der
// Lauf misst dann die Starts und nennt den Rest als Grenze.
// 4. Der Beleg geht nach `docs/belege/huellenprobe/budgets-gast.md`.
//
// **Was der Gast nicht ist**: dieselbe Hardware unter einer anderen
// Grafikschicht. Eine VM auf Apples Virtualisierung hat keine GPU des Wirts —
// was sie misst, ist eine **Obergrenze** und kein zweiter Rechner.
import { execFileSync } from 'node:child_process'
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { homedir } from 'node:os'
import { join, resolve } from 'node:path'

import { bauabdruck, dateiabdruck } from './baustand.mjs'
import { oberflaecheSteht, textziel } from './huellenprobe.mjs'
import { kennzahlen, langeListe, namensfeld, paletteOffen, paletteZeile, rollzeit, seitenzeit, vorschauzeit } from './huellenprobe/budgets.mjs'
import { fahre, imGeteilten } from './huellenprobe/gastlauf.mjs'

const wurzel = resolve(import.meta.dirname, '..')
const vm = process.argv[2] ?? 'mse080-152'
const buendel = join(wurzel, 'build/bin/MagicSetEditor4Mac.app')
const pack = process.env.MSE_PROBE_PACK ?? join(homedir(), 'full-magic-pack')
const kanon = process.env.MSE_PROBE_KANON ?? join(homedir(), 'Library/Application Support/MagicSetEditor4Mac/kanonische-sets')
const lumeUmgebung = { ...process.env, LUME_TELEMETRY_ENABLED: 'false', LUME_UPDATE_CHECK: 'false' }
const lume = (...argumente) => execFileSync('lume', argumente, { env: lumeUmgebung, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] })

const stand = JSON.parse(readFileSync(join(buendel, 'Contents/Resources/baustand.json'), 'utf8'))
if (stand.art !== 'abnahme' || stand.abdruck !== bauabdruck(wurzel)) {
  console.error('Das Bündel ist nicht der Abnahmebau dieses Stands — ./scripts/bauen.sh abnahme')
  process.exit(2)
}
if (/running/.test(lume('ls').split('\n').find((z) => z.startsWith(`${vm} `)) ?? '')) {
  console.error(`Die VM ${vm} läuft schon — sie startet hier mit dem geteilten Verzeichnis dieses Laufs`)
  process.exit(2)
}

const beginn = new Date()
// **Das geteilte Verzeichnis geht in jedem Fall weg** — samt Vorbereitung,
// nach einer Ausnahme und nach einem Signal (`imGeteilten`; Befund 1 des
// Nachlaufs über Phase I: hier lag die Entfernung außerhalb jedes `finally`).
const { ergebnis, rueckgabe, ausgabe } = await imGeteilten('b02', async (geteilt) => {
  // --- das geteilte Verzeichnis
  const b02 = join(geteilt, 'b02')
  mkdirSync(join(b02, 'pack'), { recursive: true })
  execFileSync('cp', ['-cR', buendel, join(b02, 'Abnahme.app')])
  execFileSync('cp', ['-cR', join(wurzel, 'saat/probe-werkstatt.mse-set'), join(b02, 'probe-werkstatt.mse-set')])
  for (const datei of ['budgets.pl', 'Gast.pm']) execFileSync('cp', ['-c', join(wurzel, 'scripts/huellenprobe/gast', datei), join(b02, datei)])
  if (existsSync(join(pack, 'data')) && existsSync(join(kanon, 'Black.mse-set'))) {
    execFileSync('cp', ['-cR', join(pack, 'data'), join(b02, 'pack/data')])
    execFileSync('cp', ['-c', join(kanon, 'Black.mse-set'), join(b02, 'Black.mse-set')])
  }
  writeFileSync(join(b02, 'messungen.json'), JSON.stringify({
    oberflaecheSteht, vorschauzeit, seitenzeit, rollzeit, paletteOffen, namensfeld, langeListe,
    paletteZeile: paletteZeile('__SEITE__'),
    textziel: textziel('__TEXT__').replaceAll('"__TEXT__"', '__TEXT__'),
  }))

  // --- die VM, der Lauf, das Ende — **über `fahre`**, mit Festschreiben vor
  // dem Stopp; hier stand bis zum 2026-10-11 eine eigene Kopie ohne.
  const { rueckgabe, ausgabe } = await fahre({ vm, geteilt, name: 'b02', treiber: 'budgets.pl' })

  return { rueckgabe, ausgabe, ergebnis: existsSync(join(geteilt, 'b02-ergebnis.json')) ? JSON.parse(readFileSync(join(geteilt, 'b02-ergebnis.json'), 'utf8')) : null }
})

// --- der Beleg
if (!ergebnis) {
  console.error(`Kein Ergebnis aus dem Gast (Rückgabe ${String(rueckgabe)}):\n${ausgabe}`)
  process.exit(1)
}
const gemessen = (reihe) => (reihe ?? []).filter((z) => z >= 0)
const zeile = (was, reihe, budget) => {
  if (gemessen(reihe).length === 0) return `| ${was} | 0 | — | — | — | — | ${budget} — **nicht gemessen** |`
  const k = kennzahlen(gemessen(reihe))

  return `| ${was} | ${k.n} | ${k.min} | ${k.median} | ${k.p95 ?? '—'} | ${k.max} | ${budget} |`
}
const g = ergebnis.maschine
const kopf = execFileSync('git', ['-C', wurzel, 'rev-parse', '--short', 'HEAD']).toString().trim()
const text = [
  '# Beleg: B02 im Gast — Leistungsbudgets in der VM',
  '',
  `**Stand:** Commit \`${kopf}\` · **Datum:** ${beginn.toISOString().slice(0, 16).replace('T', ' ')} UTC · **Befehl:** \`node scripts/budgetsgast.mjs ${vm}\` · **Rückgabe:** ${String(rueckgabe)}`,
  '',
  `**Maschine:** Lume-VM \`${vm}\` — macOS ${g.system} (${g.bau}), ${g.modell}, ${g.kerne} Kerne, ${String(g.speicher)} MB; Last im Gast nach dem Start ${g.last}, beim Messen ${g.lastBeimMessen ?? '?'}. Wirt: ${execFileSync('sysctl', ['-n', 'machdep.cpu.brand_string']).toString().trim()}, macOS ${execFileSync('sw_vers', ['-productVersion']).toString().trim()}.`,
  '',
  `**Bündel:** Abnahmebau, Baustand \`${stand.abdruck.slice(0, 12)}\`, Programm \`${dateiabdruck(join(buendel, 'Contents/MacOS/MagicSetEditor4Mac')).slice(0, 16)}…\`.`,
  '',
  'Gemessen mit **denselben JavaScript-Stücken** wie am Wirt (`scripts/huellenprobe/budgets.mjs`, übergeben als `messungen.json`); getrieben von `scripts/huellenprobe/gast/budgets.pl`, weil der Gast kein Node hat. **Eine VM hat keine GPU des Wirts** — die Zahlen sind eine Obergrenze, kein zweiter Rechner.',
  '',
  '| Reihe | n | min | Median | P95 | max | Vorschlag des Plans |',
  '| --- | --- | --- | --- | --- | --- | --- |',
  zeile('Start kalt (ms)', ergebnis.kalt, '≤ 2000'),
  zeile('Start warm (ms)', ergebnis.warm, '≤ 2000'),
  zeile('Feld → Vorschaubild (ms)', ergebnis.vorschau, 'Median ≤ 150'),
  zeile('Seitenwechsel (ms)', ergebnis.wechsel, '—'),
  zeile('Rollbild Black (ms)', ergebnis.rollen?.zeiten, '≤ 50'),
  '',
  `- ${String(ergebnis.daneben ?? 0)} Tasten kamen nicht im Feld an und zählen nicht; ${String((ergebnis.vorschau ?? []).filter((z) => z < 0).length)} Vorschauen ohne Bild${ergebnis.gruende?.length ? ` (${ergebnis.gruende.join(', ')})` : ''}; ${String((ergebnis.wechsel ?? []).filter((z) => z < 0).length)} Seitenwechsel ohne Ankunft.`,
  `- Speicher (MB, RSS): nach dem Start ${JSON.stringify(ergebnis.speicherNachStart)}${ergebnis.speicherNachVorschau ? `, nach den Vorschauen ${JSON.stringify(ergebnis.speicherNachVorschau)}` : ''}.`,
  ergebnis.bilder ? `- Die Seite zeichnet: ${String(ergebnis.bilder.bilder)} Bilder in ${String(Math.round(ergebnis.bilder.ms))} ms (\`requestAnimationFrame\`).` : '- Ob die Seite zeichnet, wurde nicht erreicht.',
  ...(ergebnis.setname ? [`- Das große Set heißt »${ergebnis.setname}«; Höhe der Liste ${String(ergebnis.rollen?.hoehe ?? '?')} Punkte.`] : []),
  ...(ergebnis.grenzen?.length ? ['', '**Grenzen dieses Laufs:**', '', ...ergebnis.grenzen.map((f) => `- ${f}`), '- Warum: der Gast steht hinter seinem Sperrbildschirm ([Bild](budgets-gast/sperrbildschirm.jpg)); entsperren hieße ein Kennwort eingeben, und das geschieht nicht (`AGENTS.md` §17).'] : []),
  ...(ergebnis.fehler.length ? ['', '**Fehler im Lauf:**', '', ...ergebnis.fehler.map((f) => `- ${f}`)] : []),
  '',
  '## Rohwerte',
  '',
  '```json',
  JSON.stringify({ kalt: ergebnis.kalt, warm: ergebnis.warm, vorschau: ergebnis.vorschau, wechsel: ergebnis.wechsel, rollen: ergebnis.rollen?.zeiten }),
  '```',
  '',
]
writeFileSync(join(wurzel, 'docs/belege/huellenprobe/budgets-gast.md'), text.join('\n'))
console.log(text.slice(0, 22).join('\n'))
process.exit(ergebnis.fehler.length ? 1 : 0)
