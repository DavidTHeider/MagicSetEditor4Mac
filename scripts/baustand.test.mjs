// Prüfungen von scripts/baustand.mjs — T06
// (#2099).
//
// **Ein Bündel von gestern am selben HEAD misst nichts.** Der Bauabdruck
// sagt, aus welchem Stand der Baueingänge ein Bündel entstand; die Probe hält
// ihn gegen den Baum, bevor sie misst. Geprüft gegen ein Wegwerf-Repositorium.

import { test } from 'node:test'
import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, readdirSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

import { bauabdruck, laeufeRaeumen } from './baustand.mjs'

const git = (ort, ...argumente) => execFileSync('git', ['-C', ort, ...argumente], { stdio: 'pipe' }).toString()

// einRepo legt ein Repositorium mit je einer Datei in einem Baueingang und
// außerhalb an.
function einRepo() {
  const ort = mkdtempSync(join(tmpdir(), 'baustand-'))
  git(ort, 'init', '-q')
  git(ort, 'config', 'user.email', 'probe@beispiel.invalid')
  git(ort, 'config', 'user.name', 'Probe')
  git(ort, 'config', 'commit.gpgsign', 'false')
  mkdirSync(join(ort, 'internal'))
  mkdirSync(join(ort, 'docs'))
  writeFileSync(join(ort, 'internal/a.go'), 'package a\n')
  writeFileSync(join(ort, 'docs/b.md'), '# b\n')
  git(ort, 'add', '.')
  git(ort, 'commit', '-q', '-m', 'eins')

  return ort
}

test('derselbe Baum gibt denselben Abdruck', () => {
  const ort = einRepo()
  assert.equal(bauabdruck(ort), bauabdruck(ort))
  assert.match(bauabdruck(ort), /^[0-9a-f]{64}$/)
})

test('eine Änderung an einem Baueingang ändert den Abdruck — am selben HEAD', () => {
  const ort = einRepo()
  const vorher = bauabdruck(ort)
  writeFileSync(join(ort, 'internal/a.go'), 'package a\n\nvar x = 1\n')
  assert.notEqual(bauabdruck(ort), vorher)
})

test('eine neue, noch nicht versionierte Datei im Baueingang zählt mit', () => {
  const ort = einRepo()
  const vorher = bauabdruck(ort)
  writeFileSync(join(ort, 'internal/neu.go'), 'package a\n')
  assert.notEqual(bauabdruck(ort), vorher)
})

test('eine Änderung außerhalb der Baueingänge ändert ihn nicht', () => {
  const ort = einRepo()
  const vorher = bauabdruck(ort)
  writeFileSync(join(ort, 'docs/b.md'), '# b, geändert\n')
  writeFileSync(join(ort, 'docs/neu.md'), '# neu\n')
  assert.equal(bauabdruck(ort), vorher)
})

test('ein Commit nur an der Doku ändert ihn nicht — einer an einem Eingang schon', () => {
  // **Der Inhalt zählt, nicht HEAD**: sonst hieße jedes Bündel nach einem
  // Doku-Commit veraltet, obwohl aus demselben Stand gebaut.
  const ort = einRepo()
  const vorher = bauabdruck(ort)
  writeFileSync(join(ort, 'docs/b.md'), '# b, geändert\n')
  git(ort, 'commit', '-q', '-am', 'zwei')
  assert.equal(bauabdruck(ort), vorher)
  writeFileSync(join(ort, 'internal/a.go'), 'package a\n\nvar y = 2\n')
  git(ort, 'commit', '-q', '-am', 'drei')
  assert.notEqual(bauabdruck(ort), vorher)
})

test('von den Läufen bleiben die jüngsten', () => {
  const ort = mkdtempSync(join(tmpdir(), 'laeufe-'))
  for (const name of ['2026-10-01T10-00-00Z-1', '2026-10-02T10-00-00Z-2', '2026-10-03T10-00-00Z-3', '2026-10-04T10-00-00Z-4']) {
    mkdirSync(join(ort, name))
  }
  assert.deepEqual(laeufeRaeumen(ort, 2), ['2026-10-01T10-00-00Z-1', '2026-10-02T10-00-00Z-2'])
  assert.deepEqual(readdirSync(ort).sort(), ['2026-10-03T10-00-00Z-3', '2026-10-04T10-00-00Z-4'])
})
