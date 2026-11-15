import { chmodSync, readdirSync, rmSync } from 'node:fs'
import { join } from 'node:path'

// Der Kern von `zufallsmutation.mjs` — Ziehen, Intervall, Umgebung und
// Aufräumen, ohne Go, damit er sich prüfen lässt (B05, #2224).

// zufall ist mulberry32: klein, schnell und **auf jeder Maschine gleich** —
// derselbe Samen zieht dieselben Mutanten.
export function zufall(samen) {
  let a = samen >>> 0
  return () => {
    a = (a + 0x6d2b79f5) >>> 0
    let t = a
    t = Math.imul(t ^ (t >>> 15), t | 1)
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61)
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

// geschichtet zieht je Paket `je` Stellen ohne Zurücklegen.
export function geschichtet(stellen, samen, je) {
  const naechste = zufall(samen)
  const nachPaket = new Map()
  for (const s of stellen) nachPaket.set(s.paket, [...(nachPaket.get(s.paket) ?? []), s])
  const gezogen = []
  for (const paket of [...nachPaket.keys()].sort()) {
    const vorrat = [...nachPaket.get(paket)]
    for (let i = vorrat.length - 1; i > 0; i--) {
      const j = Math.floor(naechste() * (i + 1))
      ;[vorrat[i], vorrat[j]] = [vorrat[j], vorrat[i]]
    }
    gezogen.push(...vorrat.slice(0, je))
  }

  return gezogen
}

// wilson gibt das 95-%-Intervall einer Quote — auch bei 0 oder n.
export function wilson(treffer, n, z = 1.96) {
  if (n === 0) return [0, 1]
  const p = treffer / n
  const nenner = 1 + (z * z) / n
  const mitte = (p + (z * z) / (2 * n)) / nenner
  const breite = (z * Math.sqrt((p * (1 - p)) / n + (z * z) / (4 * n * n))) / nenner

  return [Math.max(0, mitte - breite), Math.min(1, mitte + breite)]
}

// mitAktivierung nimmt die Ausgabe von `activate.sh … && env` in eine
// isolierte Umgebung auf — **ohne ihr Zuhause und ihre Go-Schalter**: die
// Aktivierung läuft mit der Umgebung, die man ihr gibt, und ihre Ausgabe
// nennt dann dasselbe `HOME`; ein Lauf, der sie mit der des Rufenden fährt,
// bekäme das echte zurück (Zweitblick über Phase B, Befund 1).
export function mitAktivierung(isoliert, ausgabe) {
  const umgebung = { ...isoliert }

  for (const zeile of ausgabe.split('\n')) {
    const stelle = zeile.indexOf('=')
    if (stelle > 0) umgebung[zeile.slice(0, stelle)] = zeile.slice(stelle + 1)
  }

  umgebung.HOME = isoliert.HOME
  umgebung.GOFLAGS = isoliert.GOFLAGS

  return umgebung
}

// wegraeumen entfernt einen Ort samt Inhalt — **auch den Modulcache, den Go
// schreibgeschützt anlegt**: `rmSync` allein scheiterte daran mit ENOTEMPTY,
// und zwei Läufe ließen je 338 MB in `$TMPDIR` liegen. Erst werden die
// Verzeichnisse wieder beschreibbar, dann geht alles.
export function wegraeumen(ort) {
  const beschreibbar = (pfad) => {
    try {
      chmodSync(pfad, 0o700)
    } catch {
      return
    }
    for (const eintrag of readdirSync(pfad, { withFileTypes: true })) {
      if (eintrag.isDirectory()) beschreibbar(join(pfad, eintrag.name))
    }
  }

  beschreibbar(ort)
  rmSync(ort, { recursive: true, force: true })
}
