#!/usr/bin/env node
// **Zehn Neustarts unter launchd, eine Marke, im Gast** — SEC02
// (#2355).
//
//     node scripts/sec02gast.mjs [<vm>]
//
// **Nie am Wirt** — ein LaunchAgent ist dauerhafte Einrichtung
// (`docs/FREIGABEN.md`), im Gast ist er Wegwerfware. Das Bündel ist der
// Abnahmebau dieses Stands (`./scripts/bauen.sh abnahme`), der Treiber im
// Gast `scripts/huellenprobe/gast/sec02.pl` — was er prüft, steht in seinem
// Kopf. Der Beleg geht nach `docs/belege/huellenprobe/sec02-<system>.md`.
import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { oberflaecheSteht, textziel } from './huellenprobe.mjs'
import { abnahmestand, fahre, fuelleGeteilt, gastziel, imGeteilten, kopfzeilen, schrittzeilen, wurzel } from './huellenprobe/gastlauf.mjs'

const { vm } = gastziel(process.argv[2], { wirtErlaubt: false })
const stand = abnahmestand(vm)

const beginn = new Date()
const { lauf, ergebnis } = await imGeteilten('sec02', async (geteilt) => {
  const ordner = fuelleGeteilt(geteilt, 'sec02', 'sec02.pl')
  writeFileSync(join(ordner, 'messungen.json'), JSON.stringify({
    oberflaecheSteht,
    textziel: textziel('__TEXT__').replaceAll('"__TEXT__"', '__TEXT__'),
  }))
  const lauf = await fahre({ vm, geteilt, name: 'sec02', treiber: 'sec02.pl', amWirt: false })
  const datei = join(geteilt, 'sec02-ergebnis.json')

  return { lauf, ergebnis: existsSync(datei) ? JSON.parse(readFileSync(datei, 'utf8')) : null }
})
const { rueckgabe, ausgabe } = lauf
if (!ergebnis) {
  console.error(`Kein Ergebnis (Rückgabe ${String(rueckgabe)}):\n${ausgabe}`)
  process.exit(1)
}

const g = ergebnis.maschine
const rot = ergebnis.schritte.filter((s) => !s.ok)
const text = [
  `# Beleg: SEC02 — zehn Neustarts unter launchd, eine Marke, macOS ${g.system}`,
  '',
  ...kopfzeilen({ beginn, befehl: `node scripts/sec02gast.mjs ${vm}`, rueckgabe, wo: `in der Lume-VM \`${vm}\``, maschine: g, stand }),
  '',
  `**Ergebnis:** ${rot.length === 0 && !ergebnis.fehler ? `alle ${String(ergebnis.schritte.length)} Schritte gelungen` : `**${String(rot.length)} von ${String(ergebnis.schritte.length)} Schritten gescheitert**`}.`,
  '',
  ...schrittzeilen(ergebnis.schritte, 'Teil'),
  '',
  '**Künstlich**: ein Wegwerfort mit eigenem Zuhause, Bestand und Bund für das Fenster; die plist schreibt die Anwendung selbst über ihre Bindung `SchnittstelleEinrichten`. Der Dienst läuft unter launchd ohne eigene Umgebung — wie eingerichtet —, also über dem Bestand des Gastnutzers, einer Wegwerf-VM ohne echte Daten. Kein echter Bestand, kein Konto.',
  '',
  '```',
  ...ergebnis.ausgaben.join('\n').trim().split('\n'),
  '```',
  '',
  '**Gefragt wurden die Bindungen über die Naht des Fensters**, nicht die gezeichnete Seite: hinter dem Sperrbildschirm des Gastes ist sie verborgen. Wie die Seite *Schnittstellen* aussieht, belegt `ops03` am Wirt.',
  '',
  '**Was der Lauf nicht zeigt**: eine Fehlerausgabe über 1 MiB unter launchd (die Grenze prüfen `internal/app` und `cmd/mse-rest` an einer echten Datei) und einen Neustart nach einem Fehlerende — den belegt `betriebsband09gast`.',
  '',
]
const datei = join(wurzel, `docs/belege/huellenprobe/sec02-${g.system}.md`)
writeFileSync(datei, text.join('\n'))
console.log(text.join('\n'))
process.exit(rot.length || ergebnis.fehler ? 1 : 0)
