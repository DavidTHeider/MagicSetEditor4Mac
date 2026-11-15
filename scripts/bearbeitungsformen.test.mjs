// Prüfungen von `bearbeitungsformen.tsv` und den Teilen dazu — #2393.
//
// **Ohne Netz**: die Liste wird gegen das gehalten, was hier installiert ist
// — die Pakete im Bündel der Oberfläche, `go.mod` und `build/download.sh` von
// go-sqlite3-wasm aus dem Modulzwischenspeicher. Eine Hebung, nach der
// niemand `node scripts/bearbeitungsformen.mjs` gefahren hat, ist hier rot.
//
// **Was er nicht prüft**: ob die Archive auf dem Server noch dieselben sind.
// Das prüft der Freigabebau, der sie lädt und jede Summe vergleicht
// (`releasequelltext.sh`).
import { test } from 'node:test'
import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import { readFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import {
  ERSATZSTAENDE,
  listeLesen,
  nurTypen,
  repositorium,
  sqliteQuellen,
  tagkandidaten,
} from './bearbeitungsformenteile.mjs'

const wurzel = resolve(import.meta.dirname, '..')
const liste = listeLesen(readFileSync(join(wurzel, 'scripts/bearbeitungsformen.tsv'), 'utf8'))
const lies = (befehl, argumente, ort = wurzel) =>
  execFileSync(befehl, argumente, { cwd: ort, encoding: 'utf8', maxBuffer: 64 * 1024 * 1024 })

const skript = `#!/usr/bin/env bash
curl -#OL "https://sqlite.org/2026/sqlite-autoconf-3530400.tar.gz"
GITHUB_TAG="https://github.com/sqlite/sqlite/raw/version-3.53.4"
mkdir -p ext/
cd ext/
curl -#OL "$GITHUB_TAG/ext/misc/decimal.c"
cd ~-
mkdir -p test/
cd test/
curl -#OL "$GITHUB_TAG/test/speedtest1.c"
cd ~-
`

test('sqliteQuellen liest Quelle, Tag und Unterordner', () => {
  assert.deepEqual(sqliteQuellen(skript), {
    tag: 'version-3.53.4',
    adressen: [
      ['https://sqlite.org/2026/sqlite-autoconf-3530400.tar.gz', 'sqlite/sqlite-autoconf-3530400.tar.gz'],
      ['https://github.com/sqlite/sqlite/raw/version-3.53.4/ext/misc/decimal.c', 'sqlite/ext/decimal.c'],
      ['https://github.com/sqlite/sqlite/raw/version-3.53.4/test/speedtest1.c', 'sqlite/test/speedtest1.c'],
    ],
  })
})

test('sqliteQuellen wirft, wenn das Skript eine andere Gestalt hat', () => {
  assert.throws(() => sqliteQuellen('wget https://sqlite.org/x.tar.gz\n'), /nicht in der gelesenen Form/)
})

test('repositorium versteht die Schreibweisen der Pakete', () => {
  for (const repository of [
    'lukeed/clsx',
    'github:lukeed/clsx',
    'git+https://github.com/lukeed/clsx.git',
    'git://github.com/lukeed/clsx',
    'git@github.com:lukeed/clsx',
    { type: 'git', url: 'https://github.com/lukeed/clsx/' },
  ]) {
    assert.deepEqual(repositorium({ name: 'clsx', repository }), {
      besitzer: 'lukeed',
      name: 'clsx',
      adresse: 'https://github.com/lukeed/clsx',
    })
  }
  assert.throws(() => repositorium({ name: 'x', repository: 'https://gitlab.com/a/b' }), /kein GitHub/)
})

test('tagkandidaten nennt den Paketnamen zuerst', () => {
  assert.deepEqual(tagkandidaten('@floating-ui/core', '1.8.0'), ['@floating-ui/core@1.8.0', 'v1.8.0', '1.8.0'])
})

// --- gegen den Baum -------------------------------------------------------

const imBuendel = lies('npm', ['ls', '--omit=dev', '--all', '--parseable'], join(wurzel, 'frontend'))
  .split('\n')
  .filter((pfad) => pfad.includes('/node_modules/'))
  .map((pfad) => ({ pfad, ...JSON.parse(readFileSync(join(pfad, 'package.json'), 'utf8')) }))

test('jedes Paket im Bündel steht in der Liste, mit seiner Fassung', () => {
  const npm = liste.filter((zeile) => zeile.art.startsWith('npm'))
  const erwartet = [...new Set(imBuendel.map((paket) => `${paket.name}@${paket.version}`))].sort()
  const gelistet = npm.map((zeile) => `${zeile.name}@${zeile.fassung}`).sort()

  assert.deepEqual(gelistet, erwartet)
})

test('nur ein Paket ohne JavaScript ist »nur Typen«', () => {
  for (const paket of imBuendel) {
    const zeile = liste.find((z) => z.name === paket.name && z.fassung === paket.version)
    assert.equal(zeile?.art === 'npm-typen', nurTypen(paket.pfad), `${paket.name}: ${String(zeile?.art)}`)
  }
})

test('jedes Archiv hat Commit, Adresse, Summe und Größe — und ein Ziel ist ein Archiv', () => {
  const ziele = new Map()
  for (const zeile of liste.filter((z) => z.art === 'npm' || z.art === 'sqlite')) {
    assert.match(zeile.sha256, /^[0-9a-f]{64}$/, zeile.name)
    assert.ok(zeile.bytes > 0, zeile.name)
    if (zeile.art === 'npm') {
      assert.match(zeile.herkunft, /^[0-9a-f]{40}$/, zeile.name)
      assert.ok(zeile.adresse.startsWith('https://codeload.github.com/') && zeile.adresse.endsWith(`/tar.gz/${zeile.herkunft}`), zeile.name)
      assert.match(zeile.ziel, /^npm\/[\w.-]+\.tar\.gz$/, zeile.name)
    }
    const vorher = ziele.get(zeile.ziel)
    if (vorher) {
      assert.deepEqual([zeile.adresse, zeile.sha256, zeile.bytes], vorher, zeile.ziel)
    }
    ziele.set(zeile.ziel, [zeile.adresse, zeile.sha256, zeile.bytes])
  }
})

test('jeder Ersatzstand wird benutzt und steht mit seinem Grund in der Liste', () => {
  for (const [schluessel, ersatz] of Object.entries(ERSATZSTAENDE)) {
    const zeile = liste.find((z) => `${z.name}@${z.fassung}` === schluessel)
    assert.ok(zeile, `${schluessel} steht in ERSATZSTAENDE, aber nicht im Bündel`)
    assert.equal(zeile.vermerk, ersatz.grund, schluessel)
  }
})

test('ein Paket aus einem Go-Modul liegt über ERZEUGTE_MODULE bei, an derselben Fassung', () => {
  const quelltext = readFileSync(join(wurzel, 'scripts/releasequelltext.sh'), 'utf8')
  const erzeugt = /ERZEUGTE_MODULE=\(([^)]*)\)/.exec(quelltext)?.[1].split('\n').map((z) => z.trim()).filter(Boolean)
  const gomod = readFileSync(join(wurzel, 'go.mod'), 'utf8')
  for (const zeile of liste.filter((z) => z.art === 'npm-gomodul')) {
    const [modul, teil] = zeile.herkunft.split(':')
    assert.ok(
      erzeugt.some((eintrag) => {
        const [m, ...teile] = eintrag.split(':')
        return m === modul && teile.some((t) => teil === t || teil.startsWith(`${t}/`))
      }),
      `${zeile.name}: ${zeile.herkunft} liegt nicht über ERZEUGTE_MODULE bei`,
    )
    assert.match(gomod, new RegExp(`${modul.replaceAll('.', '\\.')} v${zeile.fassung.replaceAll('.', '\\.')}\\b`), zeile.name)
  }
})

test('die SQLite-Quelle ist genau, was download.sh an der Fassung aus go.mod lädt', () => {
  const fassung = /github\.com\/ncruces\/go-sqlite3-wasm\/v5 (v\S+)/.exec(readFileSync(join(wurzel, 'go.mod'), 'utf8'))?.[1]
  const speicher = lies('go', ['env', 'GOMODCACHE']).trim()
  const { tag, adressen } = sqliteQuellen(
    readFileSync(join(speicher, `github.com/ncruces/go-sqlite3-wasm/v5@${fassung}`, 'build/download.sh'), 'utf8'),
  )
  const sqlite = liste.filter((z) => z.art === 'sqlite')

  assert.deepEqual(sqlite.map((z) => [z.adresse, z.ziel]), adressen)
  for (const zeile of sqlite) {
    assert.equal(zeile.fassung, tag)
    assert.equal(zeile.herkunft, `github.com/ncruces/go-sqlite3-wasm/v5@${fassung}`)
  }
})
