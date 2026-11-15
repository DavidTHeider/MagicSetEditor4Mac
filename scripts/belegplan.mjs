// **Der Belegplan: jede Freigabezusage mit Beleg, Frist und Schwelle** — B06
// (#2225).
//
//     node scripts/belegplan.mjs [--plan docs/belege/BELEGPLAN.json]
//
// Je Eintrag des Plans: der Beleg ist da, nennt **Commit, Datum, Maschine und
// Befehl** (bei einem Laufbeleg auch die Rückgabe), ist **nicht älter als
// seine Frist**, und ein benannter **Messwert hält seine Schwelle**. Ein
// Beleg der Hüllenprobe meldet **keinen Befund** und stammt **nicht aus
// ungesicherten Änderungen**. Rot ist, was eines davon nicht tut — mit dem
// Grund.
//
// # Was er nicht hält
//
// **Ob der Beleg das Richtige misst** — das sagt der Lauf, der ihn schrieb.
// Und **ob der Commit der heutige ist**: ein Beleg darf älter sein als der
// letzte Commit, solange er in seiner Frist liegt; welche Änderung ihn
// ungültig macht, entscheidet `internal/laufbeleg` für die Läufe, die einen
// Plan mit `misst` haben.
import { existsSync, readFileSync } from 'node:fs'
import { join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const tag = 24 * 60 * 60 * 1000

// angabenAus liest Commit, Datum, Maschine, Befehl und Rückgabe aus einem
// Beleg — Markdown der Hüllenprobe oder JSON eines Laufbelegs.
export function angabenAus(roh, art) {
  if (art === 'json') {
    const beleg = JSON.parse(roh)
    return {
      commit: beleg.commit ?? '',
      datum: beleg.ende ? new Date(beleg.ende) : undefined,
      // **Eine fehlende oder leere Maschine bleibt leer** — `JSON.stringify`
      // machte aus `''`, `{}` und `[]` eine wahre Zeichenkette (Zweitblick
      // über Phase B, Befund 8, und sein Nachlauf).
      maschine: (() => {
        const roh = beleg.werkzeuge ?? beleg.umwelt ?? beleg.maschine
        const leer = roh === undefined || roh === null ||
          (typeof roh === 'string' && roh.trim() === '') ||
          (typeof roh === 'object' && Object.keys(roh).length === 0)
        return leer ? '' : JSON.stringify(roh)
      })(),
      befehl: Array.isArray(beleg.befehl) ? beleg.befehl.join(' ') : (beleg.befehl ?? ''),
      rueckgabe: beleg.rueckgabe,
      status: beleg.status,
      beleg,
    }
  }
  // **Die Köpfe sind nicht gleich gebaut** — `Commit` oder `Stand`, mit
  // oder ohne Sekunden, `UTC` oder ISO —, gelesen wird, was dasteht.
  const kopf = roh.split('\n').slice(0, 12).join('\n')
  const datum = kopf.match(/(\d{4}-\d{2}-\d{2})(?:[ T](\d{2}:\d{2}))?/)

  const befundzahl = kopf.match(/^\*\*(\d+) Befund\(e\)\*\*/m)?.[1]
  const hueller = /^# Hüllenprobe:/m.test(kopf)

  return {
    // **Das Urteil und der Stand der Hüllenprobe** (`huellenbeleg.mjs`): ein
    // Beleg mit Befunden oder aus ungesicherten Änderungen trägt keine Zusage —
    // **und einer ohne Urteil auch nicht**: bis zum Zweitblick über Phase B
    // (Befund 8) hieß ein fehlendes Urteil null Befunde.
    befunde: Number(befundzahl ?? 0),
    ohneUrteil: hueller && befundzahl === undefined && !/^\*\*bestanden\*\*/m.test(kopf),
    veraltet: /das Bündel stammt aus einem anderen Stand/.test(kopf),
    ungesichert: Number(kopf.match(/\*\*mit (\d+) ungesicherten Änderung/)?.[1] ?? 0),
    commit: kopf.match(/(?:Commit|Stand) `([0-9a-f]{7,40})`/)?.[1] ?? '',
    datum: datum ? new Date(`${datum[1]}T${datum[2] ?? '00:00'}:00Z`) : undefined,
    maschine: kopf.match(/Mac[\w,]+[^\n]*|Apple M\d+[^\n]*|macOS [\d.]+|darwin\/arm64/)?.[0] ?? '',
    befehl: kopf.match(/`(scripts\/[^`]+|make [^`]+|node [^`]+)`/)?.[1] ?? '',
  }
}

const wertAus = (angaben, roh, wert) => {
  if (wert.pfad) return wert.pfad.split('.').reduce((o, k) => o?.[k], angaben.beleg)
  const treffer = roh.match(new RegExp(wert.muster))

  return treffer ? Number(treffer[1].replace(',', '.')) : undefined
}

// pruefe gibt die Befunde eines Eintrags — leer, wenn er hält.
export function pruefe(eintrag, roh, heute = new Date()) {
  if (roh === undefined) return [`${eintrag.schritt}: der Beleg ${eintrag.beleg} fehlt`]
  const art = eintrag.beleg.endsWith('.json') ? 'json' : 'md'
  const angaben = angabenAus(roh, art)
  const befunde = []
  if (!angaben.commit) befunde.push(`${eintrag.schritt}: kein Commit im Beleg`)
  if (!angaben.datum || Number.isNaN(angaben.datum.getTime())) befunde.push(`${eintrag.schritt}: kein Datum im Beleg`)
  else if (heute - angaben.datum > eintrag.fristTage * tag) {
    befunde.push(`${eintrag.schritt}: überfällig — gelaufen am ${angaben.datum.toISOString().slice(0, 10)}, Frist ${String(eintrag.fristTage)} Tage`)
  }
  if (!angaben.maschine) befunde.push(`${eintrag.schritt}: keine Maschine im Beleg`)
  if (!angaben.befehl) befunde.push(`${eintrag.schritt}: kein Befehl im Beleg`)
  // **Ein Laufbeleg urteilt mit `status`** (`internal/laufbeleg`): ein
  // Trendfall lässt `go test` mit 1 enden, und der Lauf ist trotzdem PASS.
  // Nur ein Beleg ohne Urteil wird an seiner Rückgabe gemessen.
  if (art === 'json' && angaben.status !== undefined && angaben.status !== 'PASS') befunde.push(`${eintrag.schritt}: Urteil ${String(angaben.status)}`)
  if (art === 'json' && angaben.status === undefined && angaben.rueckgabe !== 0) befunde.push(`${eintrag.schritt}: Rückgabe ${String(angaben.rueckgabe)}`)
  if (angaben.befunde > 0) befunde.push(`${eintrag.schritt}: der Beleg meldet ${String(angaben.befunde)} Befund(e)`)
  if (angaben.ohneUrteil) befunde.push(`${eintrag.schritt}: kein Urteil im Beleg`)
  if (angaben.veraltet) befunde.push(`${eintrag.schritt}: das Bündel stammt aus einem anderen Stand als der Beleg`)
  if (angaben.ungesichert > 0) befunde.push(`${eintrag.schritt}: aus einem Stand mit ${String(angaben.ungesichert)} ungesicherten Änderung(en)`)
  // **Ein Eintrag kann mehrere Grenzen tragen** (`werte`) — die Budgets sind
  // vier, und eine Zusage „Budgets" über eine bloße Messung hielt auch, als
  // das Rollbild sein Budget riss (Zweitblick über Phase B, Befund 9).
  for (const wert of eintrag.werte ?? (eintrag.wert ? [eintrag.wert] : [])) {
    const gemessen = wertAus(angaben, roh, wert)
    if (typeof gemessen !== 'number' || Number.isNaN(gemessen)) befunde.push(`${eintrag.schritt}: Messwert nicht gefunden (${wert.pfad ?? wert.muster})`)
    else if (wert.hoechstens !== undefined && gemessen > wert.hoechstens) befunde.push(`${eintrag.schritt}: ${String(gemessen)} über der Schwelle ${String(wert.hoechstens)}${wert.name ? ` (${wert.name})` : ''}`)
    else if (wert.mindestens !== undefined && gemessen < wert.mindestens) befunde.push(`${eintrag.schritt}: ${String(gemessen)} unter der Schwelle ${String(wert.mindestens)}${wert.name ? ` (${wert.name})` : ''}`)
  }

  return befunde
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const wurzel = resolve(fileURLToPath(import.meta.url), '../..')
  const stelle = process.argv.indexOf('--plan')
  const planpfad = join(wurzel, stelle < 0 ? 'docs/belege/BELEGPLAN.json' : process.argv[stelle + 1])
  const plan = JSON.parse(readFileSync(planpfad, 'utf8'))
  let rot = 0
  for (const eintrag of plan.eintraege) {
    const ort = join(wurzel, 'docs/belege', eintrag.beleg)
    const befunde = pruefe(eintrag, existsSync(ort) ? readFileSync(ort, 'utf8') : undefined)
    console.log(`${befunde.length === 0 ? '✓' : '✗'} ${eintrag.schritt}${befunde.length ? ` — ${befunde.join('; ')}` : ''}`)
    rot += befunde.length > 0 ? 1 : 0
  }
  console.log(rot === 0 ? `✓ Belegplan: ${String(plan.eintraege.length)} Zusagen belegt` : `✗ Belegplan: ${String(rot)} von ${String(plan.eintraege.length)} Zusagen ohne gültigen Beleg`)
  process.exit(rot === 0 ? 0 : 1)
}
