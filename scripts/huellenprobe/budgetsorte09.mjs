// **Die Bestände des Messvertrags, künstlich** — PERF01
// (#2351).
//
// - **Ein großes Set**: 3.788 erfundene Karten (»Probekarte 0001« …), so viele
//   wie `Black`, das größte der kanonischen Sets — als MTG JSON eingelesen,
//   mit dem Stil M15 aus dem lokalen Pack.
// - **Ein Druckset**: 60 erfundene Karten, ein Deck groß — für Ausgeben und
//   Druckbogen. **Warum nicht das große**: 3.788 Bilder sind ein Stapellauf
//   von Minuten und nicht der Weg, den ein Mensch für ein Deck geht; die
//   Zahl steht im Vertrag, bevor gemessen wurde.
// - **Eine Sammlung mit 25.000 Zeilen**: je Zeile ein eigener Druck und ein
//   eigener Name (»Sammelkarte 00001« …), über `mse sammlung-einlesen`.
// - **Eine Sicherung** des ganzen Bestands, über `mse bestand-sichern` — sie
//   spielt der Messlauf am Fenster zurück.
//
// Nichts davon stammt aus einem echten Bestand; die Pakete kommen als
// APFS-Klon aus dem Pack (`MSE_PROBE_PACK`, Vorgabe `~/full-magic-pack`).
import { execFileSync } from 'node:child_process'
import { existsSync, mkdirSync, writeFileSync } from 'node:fs'
import { homedir } from 'node:os'
import { join, resolve } from 'node:path'

import { kindumgebung } from '../huellenorte.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
export const pack = process.env.MSE_PROBE_PACK ?? join(homedir(), 'full-magic-pack')

export const GROSS = { code: 'P09', name: 'Probesatz 0.9.0', karten: 3788, vorsatz: 'Probekarte' }
export const DRUCK = { code: 'D09', name: 'Drucksatz 0.9.0', karten: 60, vorsatz: 'Druckkarte' }
export const ZEILEN = 25000

// mtgjson erzeugt ein Set erfundener Karten in der Gestalt von
// `testdaten/mtgjson/tisch/TSP.json`.
export function mtgjson({ code, name, karten, vorsatz }) {
  const arten = [
    { type: 'Creature — Probe', types: ['Creature'], subtypes: ['Probe'], power: '2', toughness: '2' },
    { type: 'Instant', types: ['Instant'], subtypes: [] },
    { type: 'Sorcery', types: ['Sorcery'], subtypes: [] },
    { type: 'Artifact', types: ['Artifact'], subtypes: [] },
    { type: 'Enchantment', types: ['Enchantment'], subtypes: [] },
  ]
  const kosten = ['{1}{W}', '{2}{U}', '{B}', '{3}{R}', '{G}{G}', '{4}']
  const seltenheit = ['common', 'uncommon', 'rare', 'mythic']
  const stelle = code === GROSS.code ? 9 : 8
  const liste = Array.from({ length: karten }, (_, i) => {
    const nummer = String(i + 1).padStart(4, '0')
    return {
      name: `${vorsatz} ${nummer}`, manaCost: kosten[i % kosten.length], supertypes: [], ...arten[i % arten.length],
      text: `Eine erfundene Fähigkeit, Nummer ${nummer}.`, rarity: seltenheit[i % seltenheit.length],
      artist: 'Selbst gezeichnet', number: String(i + 1), uuid: `00000000-0000-4000-${String(stelle)}000-${(i + 1).toString(16).padStart(12, '0')}`,
    }
  })

  return { meta: { date: '2026-10-15', version: '5.2.2+20261015' }, data: { code, name, releaseDate: '2026-10-15', type: 'expansion', cards: liste } }
}

// sammelzeilen baut aus Kopf und Musterzeile der CSV-Ausgabe `anzahl` Zeilen,
// jede mit eigenem Druck und Namen. **Die Spalten werden gesucht** und nicht
// angenommen; fehlt eine, scheitert die Vorrichtung laut.
export function sammelzeilen(kopf, muster, anzahl) {
  const spalten = kopf.split(',')
  const druck = spalten.findIndex((s) => /printing|druck/i.test(s))
  const name = spalten.findIndex((s) => /^"?(name|kartenname)"?$/i.test(s))
  if (druck < 0 || name < 0) throw new Error(`die Sammlungs-CSV hat keine Spalte für Druck oder Name: ${kopf}`)
  const werte = muster.split(',')

  return Array.from({ length: anzahl }, (_, i) => {
    const nummer = String(i + 1).padStart(5, '0')
    return werte.map((w, j) => (j === druck ? `00000000-0000-4000-a000-${nummer.padStart(12, '0')}` : j === name ? `Sammelkarte ${nummer}` : w)).join(',')
  })
}

const mitEingabe = (ort, eingabe, ...argumente) => execFileSync(join(wurzel, '.go/bin/mse'), argumente, { cwd: wurzel, input: eingabe, env: kindumgebung(ort), maxBuffer: 64 << 20 }).toString()

// mitPack legt Pakete und Schriften des Packs in den Ort — ein Klon, keine Kopie.
export function mitPack(p, ort) {
  const pakete = join(ort.heim, 'Library/Application Support/MagicSetEditor4Mac/pakete')
  mkdirSync(pakete, { recursive: true })
  execFileSync('cp', ['-cR', `${join(pack, 'data')}/.`, pakete])
  const schriften = join(ort.heim, 'Library/Fonts')
  mkdirSync(schriften, { recursive: true })
  if (existsSync(join(pack, 'Magic - Fonts'))) execFileSync('cp', ['-cR', `${join(pack, 'Magic - Fonts')}/.`, schriften])
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
}

// setEinlesen liest ein erfundenes Set ein und gibt ihm den Stil M15.
function setEinlesen(p, ort, set) {
  const datei = join(ort.ort, `${set.code}.json`)
  writeFileSync(datei, JSON.stringify(mtgjson(set)))
  p.mse(ort, 'mtgjson-lesen', datei, 'magic')
  const id = p.mse(ort, 'sets').trim().split('\n').map((z) => z.split('\t')).find((t) => t[1] === set.name)?.[0]
  p.zeige(`${set.name}: Set ${String(id)}, ${p.mse(ort, 'stile', String(id), 'wechseln', 'magic-m15').split('\n')[0]}`)

  return id
}

// befuellt legt großes Set, Druckset, Sammlung und Sicherung an und gibt
// die Kennungen und den Ort der Sicherung zurück.
export function befuellt(p, ort) {
  mitPack(p, ort)
  const gross = setEinlesen(p, ort, GROSS)
  const druck = setEinlesen(p, ort, DRUCK)
  p.mse(ort, 'exemplar', '09-karte-0001', '1', 'nearMint', 'de', 'nonfoil')
  const [kopf, muster] = p.mse(ort, 'sammlung-ausgeben', 'csv').trim().split('\n')
  mitEingabe(ort, [kopf, ...sammelzeilen(kopf, muster, ZEILEN - 1)].join('\n') + '\n', 'sammlung-einlesen', '--auslassen')
  // **Über den großen Puffer** — `p.mse` liest höchstens 1 MB, und 25.000
  // Zeilen sind mehr (ENOBUFS im ersten Lauf).
  const zeilen = mitEingabe(ort, '', 'sammlung-ausgeben', 'csv').trim().split('\n').length - 1
  p.zeige(`Sammlung: ${String(zeilen)} Zeilen`)
  const sicherung = join(ort.ort, 'budgets09.msesicherung')
  p.mse(ort, 'bestand-sichern', sicherung)

  return { gross, druck, zeilen, sicherung }
}
