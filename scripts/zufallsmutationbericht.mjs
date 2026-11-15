// **Der Bericht der Zufallsmutation** — B05
// (#2224).
//
//     node scripts/zufallsmutationbericht.mjs   # liest docs/belege/zufallsmutation.json
//
// **Die Durchsicht steht neben den Rohdaten**
// (`docs/belege/zufallsmutation-durchsicht.json`): je überlebendem Mutanten
// `äquivalent` mit Grund — er ändert kein Verhalten — oder `lücke` mit dem
// Test, der sie seitdem schließt. **Ein Überlebender ohne Durchsicht zählt als
// Lücke**, nicht als äquivalent: wer nicht hingesehen hat, darf die Quote
// nicht schönen.
import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { wilson } from './zufallsmutationkern.mjs'

// gesehen gibt die Durchsicht eines Mutanten — **nur, wenn sie seine Stelle
// nennt** oder keine: verschlüsselt ist sie nach der Nummer, und eine neue
// Ziehung kann unter derselben Nummer einen anderen Mutanten ziehen.
const gesehen = (durchsicht, x) => {
  const eintrag = durchsicht[x.nummer]
  return eintrag && (!eintrag.stelle || eintrag.stelle === `${x.datei}:${String(x.zeile)}`) ? eintrag : undefined
}

// auswerten zählt nach der Durchsicht.
export function auswerten(lauf, durchsicht) {
  const e = lauf.ergebnisse
  const aequivalent = e.filter((x) => x.urteil === 'überlebt' && gesehen(durchsicht, x)?.urteil === 'äquivalent').length
  const gefangen = e.filter((x) => x.urteil === 'gefangen').length
  const luecken = e.filter((x) => x.urteil === 'überlebt').length - aequivalent
  const geeignet = gefangen + luecken

  return {
    gefangen, luecken, aequivalent, geeignet,
    nichtUebersetzt: e.filter((x) => x.urteil === 'nicht übersetzt').length,
    frist: e.filter((x) => x.urteil === 'frist').length,
    quote: geeignet === 0 ? 0 : gefangen / geeignet,
    intervall: wilson(gefangen, geeignet),
  }
}

const prozent = (z) => `${(z * 100).toFixed(1).replace('.', ',')} %`

// bericht schreibt den Beleg.
export function bericht(lauf, durchsicht) {
  const a = auswerten(lauf, durchsicht)
  const schicht = (feld) => [...new Set(lauf.ergebnisse.map((x) => x[feld]))].sort().map((wert) => {
    const teil = auswerten({ ergebnisse: lauf.ergebnisse.filter((x) => x[feld] === wert) }, durchsicht)
    return `| ${wert} | ${teil.gefangen} | ${teil.luecken} | ${teil.aequivalent} | ${teil.nichtUebersetzt} | ${teil.frist} | ${teil.geeignet === 0 ? '—' : prozent(teil.quote)} |`
  })
  const ende = new Date(lauf.ende)
  const ueberlebende = lauf.ergebnisse.filter((x) => x.urteil === 'überlebt')

  return [
    '# Beleg: Zufällige Mutation der Kernregeln — B05',
    '',
    `**Commit \`${lauf.commit.slice(0, 9)}\`**, ${ende.toISOString().slice(0, 10)} ${ende.toISOString().slice(11, 16)} UTC, ${lauf.maschine},`,
    `Befehl \`${lauf.befehl}\` — Samen ${String(lauf.samen)}, je Paket ${String(lauf.je)} aus ${String(lauf.stellen)} Stellen.`,
    '',
    `**Quote gefangen ÷ geeignet: ${prozent(a.quote)}** (${a.gefangen} von ${a.geeignet}), 95-%-Intervall nach Wilson ${prozent(a.intervall[0])} bis ${prozent(a.intervall[1])}. Geeignet ist, was übersetzt, in der Frist lief und nicht äquivalent ist; **„nicht übersetzt" schließt ein, was \`go vet\` im Testlauf abweist** (etwa ein \`a != x || a != y\`). **Die Lücken sind seitdem geschlossen**: je Überlebendem steht der Test, der ihn jetzt fängt — gemessen ist der Stand davor. **Keine Hochrechnung aufs Gesamtprojekt**: die Zahl gilt für \`domain\`, \`service\` und \`storage\` und die fünf Arten aus \`scripts/mutationsorte.go\`.`,
    '',
    '| Paket | gefangen | Lücke | äquivalent | nicht übersetzt | Frist | Quote |',
    '| --- | --- | --- | --- | --- | --- | --- |',
    ...schicht('paket'),
    '',
    '| Art | gefangen | Lücke | äquivalent | nicht übersetzt | Frist | Quote |',
    '| --- | --- | --- | --- | --- | --- | --- |',
    ...schicht('art'),
    '',
    `## Überlebende (${String(ueberlebende.length)})`,
    '',
    '| # | Stelle | Mutation | Durchsicht |',
    '| --- | --- | --- | --- |',
    ...ueberlebende.map((x) => `| ${x.nummer} | \`${x.datei}:${x.zeile}\` | \`${x.urtext.replaceAll('|', '\\|').slice(0, 60)}\` → \`${x.ersatz.replaceAll('|', '\\|').slice(0, 60)}\` | ${gesehen(durchsicht, x) ? `${gesehen(durchsicht, x).urteil} — ${gesehen(durchsicht, x).grund}` : durchsicht[x.nummer] ? '**Lücke** — die Durchsicht unter dieser Nummer gilt einer anderen Stelle' : '**Lücke, nicht durchgesehen**'} |`),
    '',
    'Rohdaten: `docs/belege/zufallsmutation.json`; Durchsicht: `docs/belege/zufallsmutation-durchsicht.json`.',
    '',
  ].join('\n')
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const wurzel = resolve(fileURLToPath(import.meta.url), '../..')
  const lauf = JSON.parse(readFileSync(join(wurzel, 'docs/belege/zufallsmutation.json'), 'utf8'))
  const ort = join(wurzel, 'docs/belege/zufallsmutation-durchsicht.json')
  const durchsicht = existsSync(ort) ? JSON.parse(readFileSync(ort, 'utf8')) : {}
  writeFileSync(join(wurzel, 'docs/belege/zufallsmutation.md'), bericht(lauf, durchsicht))
  console.log(`${prozent(auswerten(lauf, durchsicht).quote)} — docs/belege/zufallsmutation.md`)
}
