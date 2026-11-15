#!/usr/bin/env node
// **Quarantäne beim ersten Start und der Aufstieg 0.7.0 → Kandidat mit
// befülltem Bestand** — I03 (#2261).
//
//     ./scripts/bauen.sh abnahme && node scripts/aufstieggast.mjs <0.7.0.dmg> [<vm>]
//     node scripts/aufstieggast.mjs <0.7.0.dmg> --wirt   # ohne Gatekeeper
//
// Das Abbild ist das **Freigabeabbild** von 0.7.0 (`gh release download
// v0.7.0`); seine Prüfsumme wird gegen die signierte `pruefsummen.txt`
// daneben gehalten — hier am Wirt mit `gpg`, im Gast noch einmal über die
// Bytes. Der Treiber ist `scripts/huellenprobe/gast/aufstieg.pl`; was er misst
// und was nicht, steht in seinem Kopf. Der Beleg geht nach
// `docs/belege/huellenprobe/aufstieg-<system>.md`.
import { execFileSync } from 'node:child_process'
import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'

import { oberflaecheSteht } from './huellenprobe.mjs'
import { abnahmestand, fahre, gastziel, fuelleGeteilt, imGeteilten, kopfzeilen, schrittzeilen, wurzel } from './huellenprobe/gastlauf.mjs'

const [abbild, ziel] = process.argv.slice(2)
if (!abbild || !existsSync(abbild)) {
  console.error('Aufruf: node scripts/aufstieggast.mjs <MagicSetEditor4Mac-0.7.0.dmg> [<vm>|--wirt]')
  process.exit(2)
}
const { vm, amWirt } = gastziel(ziel, { wirtErlaubt: true })
const stand = abnahmestand(vm)

// **Die Prüfsumme kommt aus der signierten Liste**, nicht aus dem Abbild.
const liste = join(dirname(abbild), 'pruefsummen.txt')
execFileSync('gpg', ['--verify', `${liste}.asc`, liste], { stdio: 'ignore' })
const zeile = readFileSync(liste, 'utf8').split('\n').find((z) => z.endsWith('.dmg'))
if (!zeile) throw new Error('pruefsummen.txt nennt kein Abbild')

const beginn = new Date()
// **Das geteilte Verzeichnis geht in jedem Fall weg** — samt Vorbereitung,
// nach einer Ausnahme und nach einem Signal (`imGeteilten`; Befund 2 des
// Zweitblicks und Befund 1 des Nachlaufs über Phase I).
const { lauf, ergebnis } = await imGeteilten('i03', async (geteilt) => {
  const ordner = fuelleGeteilt(geteilt, 'i03', 'aufstieg.pl')
  execFileSync('cp', ['-c', abbild, join(ordner, 'alt.dmg')])
  writeFileSync(join(ordner, 'pruefsumme.txt'), `${zeile}\n`)
  writeFileSync(join(ordner, 'messungen.json'), JSON.stringify({ oberflaecheSteht }))
  const lauf = await fahre({ vm, geteilt, name: 'i03', treiber: 'aufstieg.pl', amWirt, wirtUmgebung: { MSE_I03_OHNE_GATEKEEPER: '1' } })
  return { lauf, ergebnis: existsSync(join(geteilt, 'i03-ergebnis.json')) ? JSON.parse(readFileSync(join(geteilt, 'i03-ergebnis.json'), 'utf8')) : null }
})
const { rueckgabe, ausgabe } = lauf
if (!ergebnis) {
  console.error(`Kein Ergebnis (Rückgabe ${String(rueckgabe)}):\n${ausgabe}`)
  process.exit(1)
}

const rot = ergebnis.schritte.filter((s) => !s.ok)
const g = ergebnis.maschine
const text = [
  `# Beleg: I03 — Quarantäne und Aufstieg ${amWirt ? 'am Wirt' : `auf macOS ${g.system}`}`,
  '',
  ...kopfzeilen({ beginn, befehl: `node scripts/aufstieggast.mjs ${abbild.split('/').at(-1)} ${amWirt ? '--wirt' : vm}`, rueckgabe, wo: amWirt ? 'am Wirt (Positivkontrolle, ohne Gatekeeper)' : `in der Lume-VM \`${vm}\``, maschine: g, stand }),
  '',
  `**Altes Bündel:** das Freigabeabbild von 0.7.0, \`${zeile.split(' ')[0].slice(0, 16)}…\` laut der mit \`gpg\` geprüften \`pruefsummen.txt\`.`,
  '',
  `**Ergebnis:** ${rot.length === 0 ? `alle ${String(ergebnis.schritte.length)} Schritte gelungen` : `**${String(rot.length)} von ${String(ergebnis.schritte.length)} Schritten gescheitert**`}.`,
  '',
  ...schrittzeilen(ergebnis.schritte, 'Teil'),
  '',
  ...(amWirt ? [] : [
    `- **Quarantäne**: kenntlich synthetisch, \`com.apple.quarantine\` mit dem Absender \`mse4mac-i03\`; das Bündel aus dem eingehängten Abbild erbt \`${ergebnis.gatekeeper?.erbt ?? '?'}\`.`,
    '- **Der Nutzerweg** — *Datenschutz & Sicherheit → Dennoch öffnen* — ist eine Einstellung und wird von einem Menschen gegeben, nicht von diesem Lauf (`AGENTS.md` §17); **abgenommen wird er im Freigabebeleg** als Handprüfung („Die Schritte 5 und 6 des Erststart-Wegs", `docs/RELEASE.md`). Die Programme der folgenden Teile sind deshalb ohne Quarantäne ausgepackt (`ditto --noqtn`), wie nach dieser Freigabe.',
  ]),
  `- **Schema** (verbuchte Wanderungen): vor dem Aufstieg ${String(ergebnis.schema?.vorher ?? '?')}, danach ${String(ergebnis.schema?.nachher ?? '?')} — ${ergebnis.schema && ergebnis.schema.vorher === ergebnis.schema.nachher ? 'keine Wanderung; der Rückweg ist das alte Bündel selbst' : 'gewandert; der Rückweg führt über die Sicherung davor'}.`,
  `- **Sicherung vor der Wanderung**: ${(ergebnis.sicherungen ?? []).map((s) => `\`${s}\``).join(', ') || 'keine'}.`,
  '- **Inhalt** — **Inhalte, nicht Anzahlen**: je Karte ihre Werte, je Deck seine Plätze, je Bild seine Bytes (als Abdruck), dazu Nutzer samt Profilen:',
  `  vorher \`${ergebnis.vorher ?? '—'}\`; nach dem Aufstieg \`${ergebnis.nachher ?? '—'}\`${ergebnis.geschrieben ? `; nachdem der Kandidat einen Wert schrieb \`${ergebnis.geschrieben}\`` : ''}; nach dem Rückweg \`${ergebnis.zurueck ?? '—'}\`.`,
  ergebnis.schema && ergebnis.schema.vorher === ergebnis.schema.nachher
    ? '- **Nicht gebraucht**: die Sicherung vor einer Wanderung — es gab keine. Der Rückweg ist das alte Bündel, wie `docs/ADMINISTRATION.md` („Die Appfassung tauschen") ihn beschreibt.'
    : '- **Nicht gefahren**: das Zurückspielen über das Fenster von 0.7.0 — sein Freigabebau hat keine Abnahmeleitung. Der Rückweg hier ist der aus `docs/ADMINISTRATION.md` („Die Appfassung tauschen"): altes Bündel, die Sicherung an die Stelle des Bestands.',
  ...(ergebnis.grenzen?.length ? ['', '**Grenzen dieses Laufs:**', '', ...ergebnis.grenzen.map((f) => `- ${f}`)] : []),
  '',
]
const datei = join(wurzel, `docs/belege/huellenprobe/aufstieg-${amWirt ? 'wirt' : g.system}.md`)
writeFileSync(datei, text.join('\n'))
console.log(text.join('\n'))
process.exit(rot.length || ergebnis.grenzen?.length ? 1 : 0)
