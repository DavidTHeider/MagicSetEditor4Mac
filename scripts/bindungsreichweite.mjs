// **Die Reichweite der nativen Bindungen** — B01
// (#2220).
//
//     node scripts/bindungsreichweite.mjs <lauf> [--beleg docs/belege/bindungsreichweite.md]
//
// `<lauf>` ist ein Verzeichnis unter `.huellenprobe/laeufe/`: seine
// `rufbuch/*.json` sagen, welche Bindungen die Drehbücher **am nativen
// Fenster** wirklich gerufen haben (`internal/fenster/rufbuch.go`).
//
// # Woher Nenner und Klassen kommen
//
// - **Der Nenner** ist der Bindungsvertrag (`frontend/src/bindungsvertrag.json`)
//   — jede gebundene Methode, nicht nur die beobachteten. Wer ihn aus den
//   gerufenen Namen bildete, hätte immer 100 %.
// - **Lesend und schreibend** stehen in `docs/SCHREIBWEGE.md`, und
//   `internal/werkzeuge/schreibwege_test.go` hält beide Listen vollständig.
// - **Dateidialog** sind die Bindungen, die `App.waehleDatei` erreichen —
//   aus dem Aufrufgraphen (`scripts/bindungssignaturen.go`). Hier stand bis
//   zum ersten Vertragsrundgang ✋ aus `docs/FAEHIGKEITEN.md`; das sagt aber,
//   dass eine Fähigkeit an einer Fläche **fehlt**, und fünf Dialogbindungen
//   fielen durch.
// - **Ausnahmen** stehen mit Grund in `scripts/bindungsreichweite.ausnahmen.json`
//   und **neben** der Quote, nicht in ihr.
//
// # Was sie nicht sagt
//
// Ob ein Ruf das **Richtige** tat — dafür gibt es die Prüfungen der
// Drehbücher und C01. Und **extern senden** ist keine eigene Klasse: welche
// Bindung das Netz erreicht, steht in keinem gepflegten Katalog; die, die es
// ohne Zugang nicht kann, stehen als Ausnahme mit Grund.
import { execFileSync } from 'node:child_process'
import { readFileSync, readdirSync, writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const name = /`([A-Za-z_][A-Za-z0-9_]*)`/g

// listeNach gibt die Aufzählung unter einer Überschrift — **den letzten
// Absatz** vor der nächsten Überschrift oder Trennlinie, wie
// `schreibwege_test.go` sie liest.
export function listeNach(text, ueberschrift) {
  const stelle = text.indexOf(ueberschrift)
  if (stelle < 0) throw new Error(`keine Überschrift »${ueberschrift}« — die Klassen fehlen`)
  let rest = text.slice(stelle)
  for (const marke of ['\n## ', '\n---']) {
    const ende = rest.indexOf(marke)
    if (ende > 0) rest = rest.slice(0, ende)
  }
  const absaetze = rest.trim().split('\n\n')

  return [...absaetze[absaetze.length - 1].matchAll(name)].map((treffer) => treffer[1])
}

// klasse bilanziert eine Menge von Bindungen.
function klasse(mitglieder, gerufen, gelungen, ausnahmen) {
  const ausgenommen = Object.fromEntries(mitglieder.filter((n) => n in ausnahmen).map((n) => [n, ausnahmen[n]]))
  const gezaehlt = mitglieder.filter((n) => !(n in ausnahmen))
  const erreicht = gezaehlt.filter((n) => gerufen.has(n))

  return {
    nenner: gezaehlt.length,
    erreicht: erreicht.length,
    gelungen: gezaehlt.filter((n) => gelungen.has(n)).length,
    quote: gezaehlt.length === 0 ? 1 : erreicht.length / gezaehlt.length,
    offen: gezaehlt.filter((n) => !gerufen.has(n)).sort(),
    ausgenommen,
  }
}

// reichweite bilanziert gerufene gegen alle Bindungen, je Klasse.
export function reichweite({ vertrag, schreibend, lesend, dialog, ausnahmen, rufe }) {
  const klassen = new Set([...schreibend, ...lesend])
  const ohne = vertrag.filter((n) => !klassen.has(n))
  if (ohne.length > 0) throw new Error(`ohne Klasse in SCHREIBWEGE.md: ${ohne.join(', ')}`)

  const imVertrag = new Set(vertrag)
  // **Erreicht heißt: der Ruf kam in der Bindung an** — gelungen, fachlich
  // abgelehnt oder abgebrochen. Was das Tor abwies, erreichte sie nicht, und
  // was die Störhülle selbst beantwortete (`gestoert`, B03) auch nicht: die
  // Bindung lief dabei gar nicht.
  const angekommen = (z) => z.anzahl - (z.ausgaenge?.abgewiesen ?? 0) - (z.ausgaenge?.gestoert ?? 0) > 0
  const gerufen = new Set(Object.entries(rufe).filter(([n, z]) => imVertrag.has(n) && angekommen(z)).map(([n]) => n))
  // **Gelungen heißt: mindestens ein Ruf kam ohne Fehler zurück** — erreicht
  // ist auch, was fachlich abgelehnt wurde.
  const gelungen = new Set(Object.entries(rufe).filter(([n, z]) => imVertrag.has(n) && (z.ausgaenge?.gelungen ?? 0) > 0).map(([n]) => n))

  return {
    gesamt: klasse(vertrag, gerufen, gelungen, ausnahmen),
    schreibend: klasse(schreibend.filter((n) => imVertrag.has(n)), gerufen, gelungen, ausnahmen),
    lesend: klasse(lesend.filter((n) => imVertrag.has(n)), gerufen, gelungen, ausnahmen),
    dialog: klasse([...new Set(dialog)].filter((n) => imVertrag.has(n)), gerufen, gelungen, ausnahmen),
    fremd: Object.keys(rufe).filter((n) => !imVertrag.has(n)).sort(),
  }
}

// Vertragsrundgang heißt das Drehbuch, das Bindungen **unmittelbar** über die
// Wails-IPC ruft statt über einen Bedienweg (`scripts/huellenprobe/vertrag.mjs`).
// Seine Rufe gehen durch dieselbe Naht, sagen aber nichts darüber, ob ein
// Mensch die Bindung über die Oberfläche erreicht — deshalb steht die Quote
// der Bedienwege daneben.
export const vertragsrundgang = 'vertrag'

// gemischtAus legt die Rufbücher mehrerer Drehbücher zusammen.
export function gemischtAus(je) {
  const gemischt = {}
  for (const [drehbuch, buch] of Object.entries(je)) {
    for (const [bindung, zahl] of Object.entries(buch)) {
      const eintrag = gemischt[bindung] ?? { anzahl: 0, ausgaenge: {}, dauerMs: 0, drehbuecher: [] }
      eintrag.anzahl += zahl.anzahl
      eintrag.dauerMs += zahl.dauerMs
      for (const [art, n] of Object.entries(zahl.ausgaenge)) eintrag.ausgaenge[art] = (eintrag.ausgaenge[art] ?? 0) + n
      eintrag.drehbuecher.push(drehbuch)
      gemischt[bindung] = eintrag
    }
  }

  return gemischt
}

// nurBedienwege lässt den Vertragsrundgang weg.
export const nurBedienwege = (je) => gemischtAus(Object.fromEntries(Object.entries(je).filter(([d]) => d !== vertragsrundgang)))

// rufeAusLauf liest die Rufbücher eines Laufs — je Drehbuch.
export function rufeAusLauf(ort) {
  const verzeichnis = join(ort, 'rufbuch')
  const je = {}
  for (const datei of readdirSync(verzeichnis).filter((d) => d.endsWith('.json')).sort()) {
    je[datei.slice(0, -5)] = JSON.parse(readFileSync(join(verzeichnis, datei), 'utf8'))
  }

  return je
}

const prozent = (quote) => `${(quote * 100).toFixed(1).replace('.', ',')} %`

// bericht schreibt die Bilanz als Markdown.
export function bericht(bilanz, lauf, rufe, bedienwege) {
  const zeile = (titel, k) => `| ${titel} | ${k.nenner} | ${k.erreicht} | ${prozent(k.quote)} | ${k.gelungen} | ${prozent(k.nenner === 0 ? 1 : k.gelungen / k.nenner)} | ${Object.keys(k.ausgenommen).length} |`
  const abgewiesen = Object.entries(rufe).filter(([, z]) => (z.ausgaenge.abgewiesen ?? 0) > 0).map(([n, z]) => `\`${n}\` (${z.ausgaenge.abgewiesen})`)

  return [
    '# Beleg: Reichweite der nativen Bindungen — B01',
    '',
    `**Lauf \`${lauf.kennung}\`** — Commit \`${lauf.head.slice(0, 9)}\`${lauf.offen?.length ? ` **mit ${lauf.offen.length} ungesicherten Änderung(en)**` : ''}, ${(lauf.ende ?? '').slice(0, 16).replace('T', ' ')} UTC, ${lauf.umwelt?.modell ?? '?'}, macOS ${lauf.umwelt?.macos ?? '?'}, ${lauf.faelle.length} Drehbücher.`,
    `Ausgewertet mit \`node scripts/bindungsreichweite.mjs .huellenprobe/laeufe/${lauf.kennung} --beleg docs/belege/bindungsreichweite.md\`.`,
    '',
    'Gezählt ist, was ein Drehbuch **am nativen Fenster** gerufen hat (`internal/fenster/rufbuch.go`), gegen **jede** Bindung des Vertrags. Lesend und schreibend stehen in `docs/SCHREIBWEGE.md`, die Dateidialoge kommen aus dem Aufrufgraphen (`scripts/bindungssignaturen.go`); Ausnahmen mit Grund stehen neben der Quote. **Erreicht** heißt: der Ruf kam in der Bindung an (gelungen, fachlich abgelehnt oder abgebrochen); **gelungen** heißt: mindestens ein Ruf kam ohne Fehler zurück.',
    '',
    '| Klasse | Nenner | erreicht | Quote | gelungen | Quote | ausgenommen |',
    '| --- | --- | --- | --- | --- | --- | --- |',
    zeile('alle', bilanz.gesamt),
    zeile('schreibend', bilanz.schreibend),
    zeile('lesend', bilanz.lesend),
    zeile('Dateidialog', bilanz.dialog),
    '',
    ...(bedienwege ? [
      '**Ohne den Vertragsrundgang** — nur was ein Bedienweg erreichte:',
      '',
      '| Klasse | Nenner | erreicht | Quote | gelungen | Quote | ausgenommen |',
      '| --- | --- | --- | --- | --- | --- | --- |',
      zeile('alle', bedienwege.gesamt),
      zeile('schreibend', bedienwege.schreibend),
      zeile('lesend', bedienwege.lesend),
      zeile('Dateidialog', bedienwege.dialog),
      '',
    ] : []),
    `**Abgewiesen am Tor oder unbekannt** (${abgewiesen.length}): ${abgewiesen.join(', ') || 'keine'}.`,
    '',
    `**Gerufen, aber nicht im Vertrag**: ${bilanz.fremd.map((n) => `\`${n}\``).join(', ') || 'keine'}.`,
    '',
    `## Nicht erreichte schreibende Bindungen (${bilanz.schreibend.offen.length})`,
    '',
    bilanz.schreibend.offen.map((n) => `\`${n}\``).join(' · ') || '—',
    '',
    '## Ausnahmen',
    '',
    ...Object.entries(bilanz.gesamt.ausgenommen).map(([n, grund]) => `- \`${n}\` — ${grund}`),
    '',
  ].join('\n')
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const [ort, schalter, ziel] = process.argv.slice(2)
  if (!ort) {
    console.error('Aufruf: node scripts/bindungsreichweite.mjs <lauf> [--beleg <datei>]')
    process.exit(2)
  }
  const wurzel = resolve(fileURLToPath(import.meta.url), '../..')
  const lesen = (pfad) => readFileSync(join(wurzel, pfad), 'utf8')
  const schreibwege = lesen('docs/SCHREIBWEGE.md')
  const je = rufeAusLauf(ort)
  const gemischt = gemischtAus(je)
  const klassen = {
    vertrag: Object.keys(JSON.parse(lesen('frontend/src/bindungsvertrag.json')).methoden),
    schreibend: listeNach(schreibwege, 'schreibenden Bindungen'),
    lesend: listeNach(schreibwege, 'lesenden Bindungen'),
    dialog: JSON.parse(execFileSync('go', ['run', 'scripts/bindungssignaturen.go', 'internal/app'], { cwd: wurzel, encoding: 'utf8' })).filter((m) => m.dialog).map((m) => m.name),
    ausnahmen: JSON.parse(lesen('scripts/bindungsreichweite.ausnahmen.json')),
  }
  const bilanz = reichweite({ ...klassen, rufe: gemischt })
  const bedienwege = vertragsrundgang in je ? reichweite({ ...klassen, rufe: nurBedienwege(je) }) : undefined
  const text = bericht(bilanz, JSON.parse(readFileSync(join(ort, 'manifest.json'), 'utf8')), gemischt, bedienwege)
  if (schalter === '--beleg' && ziel) writeFileSync(ziel, text)
  console.log(text)
}
