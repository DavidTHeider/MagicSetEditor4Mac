#!/usr/bin/env node
// **Der Mutationslauf** — QA-05.AK-03
// (#1736).
//
//     node scripts/mutationslauf.mjs            # alle Mutationen
//     node scripts/mutationslauf.mjs M3 M7      # nur diese
//
// # Was er fragt
//
// **Erkennen die Tests einen echten Fehler an einer kritischen Invariante?**
// Jede Mutation schaltet eine Wache ab — eine Rollenprüfung, eine Nonce, eine
// Längengrenze — und fährt die Tests des Pakets. Wird der Lauf rot, ist die
// Mutation **gefangen**; bleibt er grün, hat der Katalog eine Lücke, und sie
// steht im Beleg mit Namen.
//
// # Was er nicht ist
//
// **Kein Werkzeug, das selbst mutiert.** Die Mutationen sind von Hand
// gewählt, je eine an einer Invariante aus `docs/RISIKOKATALOG.md` — die
// Frage ist, ob **diese** Wachen geprüft sind, nicht wie viele Zeilen eine
// Maschine verändern kann. Und die Gegenprobe gilt (§10): eine Mutation, die
// nichts verschiebt, beweist nichts; deshalb muss jede Fundstelle **genau
// einmal** passen, sonst bricht der Lauf ab, statt eine Mutation zu fahren,
// die gar nicht stattfand.
//
// **Gefangen ist nur, was der erwartete Test fängt** — bei grünem Grundlauf
// desselben Befehls (T05, #2108). Frist, Signal, roter Grundlauf und ein
// anderer roter Test sind benannt unentschieden; das Urteil steht in
// `mutationskern.mjs` und ist dort geprüft.
//
// **Der Baum wird immer zurückgesetzt** — aus der Kopie im Speicher, auch nach
// einem Fehler oder Strg-C. Gefahren wird mit dem Go des Projekts
// (`activate.sh`) und einem leeren eigenen Zuhause. **Der Beleg**
// (`docs/belege/mutationen.md`) entsteht nur aus einem vollen Lauf; eine
// Auswahl schreibt ihren Bericht in das Zwischenverzeichnis und lässt den
// letzten vollständigen stehen.

import { execFileSync, spawnSync } from 'node:child_process'
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { mutationen, mutiere, urteil, waehle, zuruecksetzen } from './mutationskern.mjs'

const wurzel = resolve(dirname(fileURLToPath(import.meta.url)), '..')

let zu
try {
  zu = waehle(mutationen, process.argv.slice(2))
} catch (fehler) {
  console.error(String(fehler.message))
  process.exit(2)
}
for (const signal of ['SIGINT', 'SIGTERM']) {
  process.on(signal, () => {
    zuruecksetzen()
    process.exit(130)
  })
}

const heim = mkdtempSync(join(tmpdir(), 'mutationslauf-'))

// gotest fährt `go test` mit dem Go des Projekts und dem eigenen Zuhause.
const gotest = (m) => {
  const lauf = spawnSync('bash', ['-c', 'source scripts/activate.sh >/dev/null 2>&1 && exec go "$@"', '--',
    'test', m.paket, '-count=1', ...(m.lauf ? ['-run', m.lauf] : [])], {
    cwd: wurzel, encoding: 'utf8', timeout: 15 * 60 * 1000, env: { ...process.env, HOME: heim },
  })

  return { status: lauf.status, signal: lauf.signal, fehler: lauf.error?.code, ausgabe: `${lauf.stdout ?? ''}${lauf.stderr ?? ''}` }
}

const grundlaeufe = new Map()
const ergebnisse = []
try {
  for (const m of zu) {
    const schluessel = `${m.paket} ${m.lauf ?? ''}`
    if (!grundlaeufe.has(schluessel)) grundlaeufe.set(schluessel, gotest(m))
    const anfang = Date.now()
    const mutiert = await mutiere(m, wurzel, gotest)
    const sekunden = ((Date.now() - anfang) / 1000).toFixed(1)
    const ergebnis = urteil(m, grundlaeufe.get(schluessel), mutiert)
    ergebnisse.push({ ...m, ...ergebnis, sekunden })
    console.log(`${m.kennung} ${ergebnis.art}${ergebnis.von ? ` von ${ergebnis.von}` : ''}${ergebnis.grund ? ` — ${ergebnis.grund}` : ''} (${sekunden} s)`)
  }
} finally {
  rmSync(heim, { recursive: true, force: true })
}

const commit = execFileSync('git', ['rev-parse', '--short', 'HEAD'], { cwd: wurzel, encoding: 'utf8' }).trim()
const offen = execFileSync('git', ['status', '--porcelain', '--', '.', ':!docs/belege'], { cwd: wurzel, encoding: 'utf8' }).split('\n').filter(Boolean).length
const goFassung = execFileSync('bash', ['-c', 'source scripts/activate.sh >/dev/null 2>&1 && go version'], { cwd: wurzel, encoding: 'utf8' }).trim()
const gefangen = ergebnisse.filter((e) => e.art === 'gefangen').length
const ergebnistext = (e) => ({
  gefangen: `**gefangen** von \`${e.von}\``,
  'anders rot': `**anders rot** — erwartet \`${e.erwartet}\`, rot war \`${e.von}\``,
  überlebt: '**überlebt**',
  unentschieden: `**unentschieden** — ${e.grund}`,
  'nicht übersetzt': '**nicht übersetzt**',
})[e.art]
const zeilen = ergebnisse.map(
  (e) => `| ${e.kennung} | ${e.invariante} | \`${e.datei}\` | \`${e.paket}\`${e.lauf ? ` (\`-run ${e.lauf}\`)` : ''} | ${ergebnistext(e)} | ${e.sekunden} s |`,
)
const voll = zu.length === mutationen.length
const ziel = voll ? join(wurzel, 'docs/belege/mutationen.md') : join(tmpdir(), `mutationen-${commit}.md`)

writeFileSync(
  ziel,
  `# Mutationslauf

**${gefangen} von ${ergebnisse.length} gefangen** — Commit \`${commit}\`${offen > 0 ? ` **mit ${String(offen)} ungesicherten Änderung(en)**` : ''},
${new Date().toISOString().slice(0, 16).replace('T', ' ')} UTC, ${goFassung}.
Gefahren mit \`node scripts/mutationslauf.mjs${voll ? '' : ` ${zu.map((m) => m.kennung).join(' ')}`}\`; gefangen ist nur, was der
erwartete Test bei grünem Grundlauf fängt. Was eine Mutation ist und was der
Lauf nicht belegt, steht im Kopf des Skripts.

| | Invariante | Fundstelle | Lauf | Ergebnis | Dauer |
| --- | --- | --- | --- | --- | ---: |
${zeilen.join('\n')}
`,
)
console.log(`${gefangen} von ${ergebnisse.length} gefangen → ${ziel}`)
process.exit(gefangen === ergebnisse.length ? 0 : 1)
