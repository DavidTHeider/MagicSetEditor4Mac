#!/usr/bin/env node
// **Ein Dienst, der endet, meldet sich im Betriebsband, im Gast** — OPS02
// (#2348).
//
//     node scripts/betriebsband09gast.mjs [<vm>]
//
// **Nie am Wirt** — ein LaunchAgent ist dauerhafte Einrichtung
// (`docs/FREIGABEN.md`), im Gast ist er Wegwerfware. Das Bündel ist der
// Abnahmebau dieses Stands (`./scripts/bauen.sh abnahme`), der Treiber im
// Gast `scripts/huellenprobe/gast/betriebsband.pl` — was er prüft, steht in
// seinem Kopf. Der Beleg geht nach
// `docs/belege/huellenprobe/betriebsband09-<system>.md`.
import { existsSync, readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { oberflaecheSteht, textziel } from './huellenprobe.mjs'
import { abnahmestand, fahre, fuelleGeteilt, gastziel, imGeteilten, kopfzeilen, schrittzeilen, wurzel } from './huellenprobe/gastlauf.mjs'

const { vm } = gastziel(process.argv[2], { wirtErlaubt: false })
const stand = abnahmestand(vm)

const beginn = new Date()
const { lauf, ergebnis } = await imGeteilten('ops02', async (geteilt) => {
  const ordner = fuelleGeteilt(geteilt, 'ops02', 'betriebsband.pl')
  writeFileSync(join(ordner, 'messungen.json'), JSON.stringify({
    oberflaecheSteht,
    textziel: textziel('__TEXT__').replaceAll('"__TEXT__"', '__TEXT__'),
  }))
  const lauf = await fahre({ vm, geteilt, name: 'ops02', treiber: 'betriebsband.pl', amWirt: false })
  const datei = join(geteilt, 'ops02-ergebnis.json')

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
  `# Beleg: OPS02 — ein Dienst, der endet, meldet sich im Betriebsband, macOS ${g.system}`,
  '',
  ...kopfzeilen({ beginn, befehl: `node scripts/betriebsband09gast.mjs ${vm}`, rueckgabe, wo: `in der Lume-VM \`${vm}\``, maschine: g, stand }),
  '',
  `**Ergebnis:** ${rot.length === 0 && !ergebnis.fehler ? `alle ${String(ergebnis.schritte.length)} Schritte gelungen` : `**${String(rot.length)} von ${String(ergebnis.schritte.length)} Schritten gescheitert**`}.`,
  '',
  ...schrittzeilen(ergebnis.schritte, 'Teil'),
  '',
  '**Künstlich**: zwei Wegwerforte mit eigenem Zuhause, Bestand und Bund — einer für das Fenster, einer für den Dienst, dessen Ordner zuerst keine Rechte hat. Die plist folgt der der Anwendung (`dienstplist` in `internal/app/schnittstellenziele.go`), dazu nur die Umgebung des Wegwerforts. Kein echter Bestand, kein Konto.',
  '',
  '```',
  ...ergebnis.ausgaben.join('\n').trim().split('\n'),
  '```',
  '',
  '**Gefragt wurde die Bindung `Betriebsband` über die Naht des Fensters**, nicht die gezeichnete Seite: hinter dem Sperrbildschirm des Gastes ist sie verborgen, und das Fenster fragt dann nicht — so ist es gebaut. Wie der Hinweis aussieht und wohin er führt, belegt `betriebsband09` am Wirt.',
  '',
  '**Was der Lauf nicht zeigt**: Platz, Sicherung und Webprozess — die belegt die Hüllenprobe `betriebsband09` am Wirt; und das Einrichten des Dienstes über das Fenster, das `ops03gast` belegt.',
  '',
]
const datei = join(wurzel, `docs/belege/huellenprobe/betriebsband09-${g.system}.md`)
writeFileSync(datei, text.join('\n'))
console.log(text.join('\n'))
process.exit(rot.length || ergebnis.fehler ? 1 : 0)
