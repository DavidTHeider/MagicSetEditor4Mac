#!/usr/bin/env node
// **Die Bearbeitungsformen der Abhängigkeiten, erhoben** — #2393, Weg 1,
// entschieden vom Auftraggeber am 2026-10-16.
//
//     node scripts/bearbeitungsformen.mjs    # schreibt scripts/bearbeitungsformen.tsv
//
// GPL-2.0 §3 verlangt den vollständigen Quelltext jedes Moduls, das das
// Programm enthält, in der Form, in der man es bearbeitet (Vermerk 0190).
// `vendor/` und `node_modules/` legen die Abhängigkeiten bei, wie ihre
// Pakete erscheinen; **wo diese Form gebaut ist**, gehört der Ursprung dazu.
// Diese Datei erhebt ihn — einmal je Hebung und mit Netz — und schreibt je
// Archiv Adresse, Prüfsumme und Größe. Beim Freigabebau lädt
// `releasequelltext.sh` sie und prüft jede Summe; ein Archiv, das sich auf
// dem Server geändert hat, hält die Freigabe an, statt still ein anderes
// beizulegen.
//
// # Was es erhebt
//
// - **Die npm-Pakete im Bündel der Oberfläche** (`npm ls --omit=dev`): je
//   Paket der Commit, aus dem es veröffentlicht ist — `gitHead` aus der
//   Registry, sonst ein Tag (`tagkandidaten`). Ist das Archiv dort nicht zu
//   haben, gilt nur ein **benannter** Ersatzstand (`ERSATZSTAENDE`), sonst
//   bricht es ab. Das Archiv ist das ganze Repositorium an diesem Commit,
//   von `codeload.github.com` — die Bauskripte eines Monorepos liegen an
//   seiner Wurzel. **Ein Paket ohne JavaScript** (`@types/react`, `csstype`
//   …) bekommt kein Archiv: es geht nicht ins Bündel, und seine `.d.ts` sind
//   selbst Quelltext.
// - **Die SQLite-Quelle** zu `go-sqlite3-wasm` an der Fassung aus `go.mod`:
//   genau die Adressen, die `build/download.sh` des Moduls lädt. Dessen
//   SHA3-Prüfung entfällt still, wenn `openssl` kein SHA3 kann; hier gilt
//   eine eigene SHA-256-Summe.
//
// # Was es nicht tut
//
// Es läuft nicht in der Pipeline — ohne Netz ginge es nicht.
// `bearbeitungsformen.test.mjs` hält die Liste ohne Netz gegen die
// installierten Pakete und gegen `download.sh`: eine Hebung ohne neuen Lauf
// hier ist dort rot.
import { execFileSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { existsSync, mkdirSync, readFileSync, renameSync, rmSync, writeFileSync } from 'node:fs'
import { dirname, join, relative, resolve } from 'node:path'

import {
  ERSATZSTAENDE,
  nurTypen,
  repositorium,
  sqliteQuellen,
  tagkandidaten,
} from './bearbeitungsformenteile.mjs'

const wurzel = resolve(import.meta.dirname, '..')
const oberflaeche = join(wurzel, 'frontend')
const speicher = join(wurzel, 'build/quellarchive')
const liste = join(wurzel, 'scripts/bearbeitungsformen.tsv')

// **Mit großem Puffer**: das Inhaltsverzeichnis des React-Archivs allein
// sprengt die Vorgabe von 1 MB.
const lies = (befehl, argumente, optionen = {}) =>
  execFileSync(befehl, argumente, { encoding: 'utf8', maxBuffer: 256 * 1024 * 1024, ...optionen })

// holen lädt eine Adresse einmal in den Speicher und gibt Summe und Größe —
// oder `undefined`, wenn der Server sie nicht hat.
function holen(adresse, name) {
  const datei = join(speicher, name)
  if (!existsSync(datei)) {
    mkdirSync(dirname(datei), { recursive: true })
    try {
      lies('curl', ['-fsSL', '--retry', '3', '--max-time', '900', '-o', `${datei}.teil`, adresse], {
        stdio: 'pipe',
      })
    } catch {
      rmSync(`${datei}.teil`, { force: true })

      return undefined
    }
    renameSync(`${datei}.teil`, datei)
  }
  const inhalt = readFileSync(datei)

  return { sha256: createHash('sha256').update(inhalt).digest('hex'), bytes: inhalt.length }
}

// tags liest die Tags eines Repositoriums einmal — geschälte zuerst, denn
// ein annotiertes Tag zeigt sonst auf sein Tagobjekt und nicht auf den Commit.
const tagspeicher = new Map()
function tagCommit(adresse, tag) {
  if (!tagspeicher.has(adresse)) {
    tagspeicher.set(
      adresse,
      lies('git', ['ls-remote', '--tags', adresse])
        .split('\n')
        .map((zeile) => zeile.split('\t'))
        .filter(([, ref]) => ref),
    )
  }
  const alle = tagspeicher.get(adresse)

  return (alle.find(([, ref]) => ref === `refs/tags/${tag}^{}`) ??
    alle.find(([, ref]) => ref === `refs/tags/${tag}`))?.[0]
}

// archiv versucht einen Commit und gibt die Zeile — oder `undefined`.
//
// **Und es sieht hinein**: das Archiv muss ein lesbares tar sein und das
// `package.json` des Pakets tragen, im Verzeichnis, das das Paket nennt. Ein
// Archiv von 2 KB (`detect-node-es`) ist sonst nicht von einer Fehlerseite
// zu unterscheiden.
function archiv(repo, commit, paket, ausKopf = false) {
  const adresse = `https://codeload.github.com/${repo.besitzer}/${repo.name}/tar.gz/${commit}`
  const name = `npm/${repo.besitzer}-${repo.name}-${commit.slice(0, 12)}.tar.gz`
  const geholt = holen(adresse, name)
  if (!geholt) return undefined
  const verzeichnis = typeof paket.repository === 'object' ? (paket.repository.directory ?? '') : ''
  const inhalt = lies('tar', ['-tzf', join(speicher, name)]).split('\n')
  const ort = inhalt.find((pfad) => pfad.split('/').slice(1).join('/') === (verzeichnis ? `${verzeichnis}/package.json` : 'package.json'))
  if (!ort) {
    throw new Error(`${paket.name}: ${name} trägt kein ${verzeichnis || '.'}/package.json`)
  }
  const gelesen = JSON.parse(lies('tar', ['-xzOf', join(speicher, name), ort]))
  if (gelesen.name !== paket.name) {
    throw new Error(`${paket.name}: das package.json im Archiv heißt ${gelesen.name}`)
  }
  // **Und es trägt dieselbe Fassung** — nur ein benannter Ersatzstand darf
  // abweichen, und sein Grund sagt, warum der Code trotzdem derselbe ist.
  // **Ein Platzhalter von semantic-release** (`ismobilejs`: `0.0.0-development`)
  // gilt nur am `gitHead` — dort ist der Commit der genaue Stand, und die
  // Fassung setzt erst die Veröffentlichung.
  const platzhalter = ausKopf && /^0\.0\.0-(development|semantically-released)$/.test(String(gelesen.version))
  if (gelesen.version !== paket.version && !platzhalter && !ERSATZSTAENDE[`${paket.name}@${paket.version}`]) {
    throw new Error(`${paket.name}@${paket.version}: das Archiv an ${commit.slice(0, 12)} trägt ${String(gelesen.version)}`)
  }

  return { commit, adresse, name, ...geholt }
}

const zeilen = []

const pakete = lies('npm', ['ls', '--omit=dev', '--all', '--parseable'], { cwd: oberflaeche })
  .split('\n')
  .filter((pfad) => pfad.includes('/node_modules/'))
for (const verzeichnis of [...new Set(pakete)].sort()) {
  const paket = JSON.parse(readFileSync(join(verzeichnis, 'package.json'), 'utf8'))
  const schluessel = `${paket.name}@${paket.version}`
  if (nurTypen(verzeichnis)) {
    zeilen.push(['npm-typen', paket.name, paket.version, '-', '-', '-', '-', '-', 'nur Typen, geht nicht ins Bündel'])
    continue
  }
  const ersatz = ERSATZSTAENDE[schluessel]
  if (ersatz?.gomodul) {
    zeilen.push(['npm-gomodul', paket.name, paket.version, ersatz.gomodul, '-', '-', '-', '-', ersatz.grund])
    continue
  }
  const repo = repositorium(paket)
  let treffer
  const kopf = lies('npm', ['view', schluessel, 'gitHead']).trim()
  if (/^[0-9a-f]{40}$/.test(kopf)) {
    treffer = archiv(repo, kopf, paket, true)
  }
  for (const tag of ersatz?.tag ? [] : tagkandidaten(paket.name, paket.version)) {
    if (treffer) break
    const commit = tagCommit(repo.adresse, tag)
    if (commit) treffer = archiv(repo, commit, paket)
  }
  if (!treffer && ersatz?.tag) {
    const commit = tagCommit(repo.adresse, ersatz.tag)
    treffer = commit && archiv(repo, commit, paket)
  }
  if (!treffer) {
    throw new Error(`${schluessel}: kein Archiv am veröffentlichten Stand — Ersatzstand benennen (ERSATZSTAENDE)`)
  }
  zeilen.push([
    'npm',
    paket.name,
    paket.version,
    treffer.commit,
    treffer.adresse,
    treffer.sha256,
    String(treffer.bytes),
    treffer.name,
    ersatz?.grund ?? '',
  ])
  process.stderr.write(`${schluessel} ${treffer.commit.slice(0, 12)} ${String(treffer.bytes)}\n`)
}

const modulfassung = /github\.com\/ncruces\/go-sqlite3-wasm\/v5 (v\S+)/.exec(
  readFileSync(join(wurzel, 'go.mod'), 'utf8'),
)?.[1]
const modulspeicher = lies('go', ['env', 'GOMODCACHE'], { cwd: wurzel }).trim()
const { tag, adressen } = sqliteQuellen(
  readFileSync(
    join(modulspeicher, `github.com/ncruces/go-sqlite3-wasm/v5@${modulfassung}`, 'build/download.sh'),
    'utf8',
  ),
)
for (const [adresse, name] of adressen) {
  const geholt = holen(adresse, name)
  if (!geholt) {
    throw new Error(`${adresse}: nicht zu haben`)
  }
  zeilen.push([
    'sqlite',
    name.replace(/^sqlite\//, ''),
    tag,
    `github.com/ncruces/go-sqlite3-wasm/v5@${modulfassung}`,
    adresse,
    geholt.sha256,
    String(geholt.bytes),
    name,
    '',
  ])
}

const kopfzeilen = [
  '# Die Bearbeitungsformen der Abhängigkeiten (#2393) — erzeugt von',
  '# scripts/bearbeitungsformen.mjs, nicht von Hand. Spalten: Art, Name,',
  '# Fassung, Herkunft (Commit oder Modul), Adresse, SHA-256, Bytes, Ziel unter',
  '# quellen/bearbeitungsformen/, Vermerk. Ein Ziel kann mehrere Zeilen tragen:',
  '# ein Monorepo ist ein Archiv für mehrere Pakete.',
]
writeFileSync(liste, [...kopfzeilen, ...zeilen.map((zeile) => zeile.join('\t'))].join('\n') + '\n')

const archive = new Map(zeilen.filter(([, , , , adresse]) => adresse !== '-').map((zeile) => [zeile[7], Number(zeile[6])]))
const summe = [...archive.values()].reduce((a, b) => a + b, 0)
process.stderr.write(
  `${String(zeilen.length)} Zeilen, ${String(archive.size)} Archive, ${(summe / 1e6).toFixed(1)} MB → ${relative(wurzel, liste)}\n`,
)
