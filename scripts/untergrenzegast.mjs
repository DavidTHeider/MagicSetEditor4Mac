#!/usr/bin/env node
// **Läuft das unveränderte Bündel auf der zugesagten Untergrenze?** — I02
// (#2260).
//
//     ./scripts/bauen.sh abnahme && node scripts/untergrenzegast.mjs [<vm>]
//     node scripts/untergrenzegast.mjs --wirt     # derselbe Treiber am Wirt
//
// Der Treiber ist `scripts/huellenprobe/gast/untergrenze.pl`; was er misst,
// steht in seinem Kopf. **`--wirt` ist die Positivkontrolle**: ein Schritt,
// der dort scheitert, sagt nichts über die Untergrenze. Der Beleg geht nach
// `docs/belege/huellenprobe/untergrenze-<system>.md`.
import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { oberflaecheSteht, textziel } from './huellenprobe.mjs'
import { namensfeld } from './huellenprobe/budgets.mjs'
import { abnahmestand, fahre, gastziel, fuelleGeteilt, imGeteilten, kopfzeilen, schrittzeilen, wurzel } from './huellenprobe/gastlauf.mjs'

const { vm, amWirt } = gastziel(process.argv[2], { wirtErlaubt: true })
const stand = abnahmestand(vm)

const beginn = new Date()
// **Das geteilte Verzeichnis geht in jedem Fall weg** — samt Vorbereitung,
// nach einer Ausnahme und nach einem Signal (`imGeteilten`; Befund 2 des
// Zweitblicks und Befund 1 des Nachlaufs über Phase I).
const { lauf, ergebnis } = await imGeteilten('i02', async (geteilt) => {
  const ordner = fuelleGeteilt(geteilt, 'i02', 'untergrenze.pl')
  writeFileSync(join(ordner, 'messungen.json'), JSON.stringify({
    oberflaecheSteht, namensfeld,
    textziel: textziel('__TEXT__').replaceAll('"__TEXT__"', '__TEXT__'),
  }))
  const lauf = await fahre({ vm, geteilt, name: 'i02', treiber: 'untergrenze.pl', amWirt, wirtUmgebung: { MSE_UNTERGRENZE_OHNE_RUHE: '1' } })
  return { lauf, ergebnis: existsSync(join(geteilt, 'i02-ergebnis.json')) ? JSON.parse(readFileSync(join(geteilt, 'i02-ergebnis.json'), 'utf8')) : null }
})
const { rueckgabe, ausgabe } = lauf
if (!ergebnis) {
  console.error(`Kein Ergebnis (Rückgabe ${String(rueckgabe)}):\n${ausgabe}`)
  process.exit(1)
}

const g = ergebnis.maschine
const rot = ergebnis.schritte.filter((s) => !s.ok)
const text = [
  `# Beleg: I02 — der Kernweg ${amWirt ? 'am Wirt' : `auf macOS ${g.system}`}`,
  '',
  ...kopfzeilen({ beginn, befehl: `node scripts/untergrenzegast.mjs ${amWirt ? '--wirt' : vm}`, rueckgabe, wo: amWirt ? 'am Wirt (Positivkontrolle)' : `in der Lume-VM \`${vm}\``, maschine: g, stand }),
  '',
  `**Ergebnis:** ${rot.length === 0 ? `alle ${String(ergebnis.schritte.length)} Schritte gelungen` : `**${String(rot.length)} von ${String(ergebnis.schritte.length)} Schritten gescheitert**`}.`,
  '',
  ...schrittzeilen(ergebnis.schritte, 'Fläche'),
  '',
  ergebnis.bilder ? `- Die Seite zeichnet: ${String(ergebnis.bilder.bilder)} Bilder in ${String(Math.round(ergebnis.bilder.ms))} ms (\`requestAnimationFrame\`), Sichtbarkeit »${ergebnis.sichtbar}«. Der Weg am Fenster geht über die Seite und nicht über die Maus — hinter dem Sperrbildschirm des Gastes ist das Fenster nicht das vordere.` : '- Ob die Seite zeichnet, wurde nicht erreicht.',
  ...(ergebnis.grenzen?.length ? ['', '**Grenzen dieses Laufs:**', '', ...ergebnis.grenzen.map((f) => `- ${f}`)] : []),
  '',
]
writeFileSync(join(wurzel, `docs/belege/huellenprobe/untergrenze-${amWirt ? 'wirt' : g.system}.md`), text.join('\n'))
console.log(text.join('\n'))
process.exit(rot.length || ergebnis.grenzen?.length ? 1 : 0)
