// **Was ein Beleg als ungesicherte Änderung zählt** — B06 (#2225).
//
// Der volle Lauf vom 2026-10-09 schrieb über `readmebilder` die README-Bilder
// neu, und jeder spätere Beleg trug „mit 64 ungesicherten Änderung(en)" —
// obwohl das Bündel aus genau dem genannten Commit gebaut war. Was der Lauf
// selbst schreibt, ist keine Änderung am geprüften Stand.
import { strict as assert } from 'node:assert'
import { execFileSync } from 'node:child_process'
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'
import { test } from 'node:test'

import { offeneAenderungen } from './huellenbeleg.mjs'

function wegwerfbaum() {
  const wurzel = mkdtempSync(join(tmpdir(), 'huellenbeleg-'))
  const git = (...argumente) => execFileSync('git', ['-C', wurzel, ...argumente], { stdio: 'ignore' })
  git('init', '-q')
  for (const ordner of ['docs/belege', 'docs/bilder', 'internal']) mkdirSync(join(wurzel, ordner), { recursive: true })
  for (const datei of ['docs/belege/a.md', 'docs/bilder/a.png', 'internal/a.go']) writeFileSync(join(wurzel, datei), 'vorher')
  git('add', '.')
  git('-c', 'user.name=probe', '-c', 'user.email=probe@example.org', '-c', 'commit.gpgsign=false', 'commit', '-q', '-m', 'anfang')

  return wurzel
}

test('Belege und README-Bilder zählen nicht, eine Quelldatei schon', () => {
  const wurzel = wegwerfbaum()
  writeFileSync(join(wurzel, 'docs/belege/a.md'), 'nachher')
  writeFileSync(join(wurzel, 'docs/bilder/a.png'), 'nachher')
  assert.equal(offeneAenderungen(wurzel), 0)

  writeFileSync(join(wurzel, 'internal/a.go'), 'nachher')
  assert.equal(offeneAenderungen(wurzel), 1)
})
