// **Chinesische Zeichen in chinesischen Formen** — G18 aus dem Zweitblick
// über Phase 6.3 (#1664), der Phase 6.10 zugeordnet (#1671).
//
// Die Ketten tragen japanische Schriften vor chinesischen, weil Kana nur
// japanische haben (UX-09.AK-02). **Steht vorn eine Schrift ohne
// Han-Zeichen** — Hoefler Text, Luminari —, nimmt WebKit für ein Han-Zeichen
// die erste Schrift der Kette, die es hat: Hiragino. Ein chinesischer Name
// bekäme japanische Zeichenformen. Die Systemschrift fragt die Sprache des
// Elements selbst; eine benannte Schrift tut es nicht.
//
// Deshalb gibt es je Kette **drei Reihen**, und `:lang()` wählt im Stilblatt
// — für die Oberfläche in ihrer Sprache und für jedes Element, das eine
// eigene trägt.
//
// # Was hier nicht geprüft wird
//
// **Ob WebKit die Regel anwendet**: jsdom löst weder `:lang()` noch
// Variablen auf. Das prüft die Hüllenprobe `g18` am Fenster.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

import { beforeEach, describe, expect, it } from 'vitest'

import {
  anzeigeschriften,
  cjkReihe,
  cjkUmordnen,
  schriftketten,
  setzeAnzeigeschrift,
  setzeSchriftbereich,
} from './erscheinungsbild'
import type { Anzeigeschrift } from './erscheinungsbild'

// cjkFolge nennt die japanischen und chinesischen Schriften einer Kette in
// ihrer Reihenfolge — `ja`, `sc` oder `tc`.
function cjkFolge(kette: string): string[] {
  return [
    ...kette.matchAll(/'(Hiragino [^']+|Songti [ST]C|PingFang [ST]C)'/g),
  ].map(([, name = '']) =>
    name.startsWith('Hiragino') ? 'ja' : name.endsWith('SC') ? 'sc' : 'tc',
  )
}

// ohneCJK ist der Rest der Kette — er bleibt in jeder Reihe derselbe.
function ohneCJK(kette: string): string {
  return kette.replace(/'(Hiragino [^']+|Songti [ST]C|PingFang [ST]C)', /g, '')
}

let wurzel: HTMLElement

beforeEach(() => {
  wurzel = document.createElement('html')
})

describe('cjkReihe', () => {
  it.each([
    ['de', 'ja'],
    ['en', 'ja'],
    ['ja', 'ja'],
    ['zh', 'hans'],
    ['zh-Hans', 'hans'],
    ['zh-CN', 'hans'],
    ['zh-Hant', 'hant'],
    ['zh-TW', 'hant'],
    ['zh-HK', 'hant'],
    ['zh-hant-tw', 'hant'],
  ])('ordnet %s der Reihe %s zu', (sprache, reihe) => {
    expect(cjkReihe(sprache)).toBe(reihe)
  })
})

describe('schriftketten', () => {
  it.each(Object.keys(anzeigeschriften))(
    'stellt für %s Chinesisch vor Japanisch, wo es gilt',
    (wahl) => {
      const ketten = schriftketten(wahl as Anzeigeschrift)

      expect(ketten.ja).toBe(anzeigeschriften[wahl as Anzeigeschrift])
      expect(cjkFolge(ketten.ja)).toEqual(['ja', 'sc', 'tc'])
      expect(cjkFolge(ketten.hans)).toEqual(['sc', 'tc', 'ja'])
      expect(cjkFolge(ketten.hant)).toEqual(['tc', 'sc', 'ja'])
      // **Nur die Han-Schriften wandern** — die Schrift vorn bleibt die
      // gewählte, der Rückfall am Ende derselbe.
      expect(ohneCJK(ketten.hans)).toBe(ohneCJK(ketten.ja))
      expect(ohneCJK(ketten.hant)).toBe(ohneCJK(ketten.ja))
    },
  )

  it('lässt eine Kette ohne Han-Schriften unberührt', () => {
    expect(cjkUmordnen('Georgia, serif', 'hans')).toBe('Georgia, serif')
  })
})

describe('am Wurzelelement', () => {
  it('setzt die Anzeigeschrift in allen drei Reihen', () => {
    setzeAnzeigeschrift(wurzel, 'buch')

    const ketten = schriftketten('buch')
    expect(wurzel.style.getPropertyValue('--font-anzeige')).toBe(ketten.ja)
    expect(wurzel.style.getPropertyValue('--font-anzeige-hans')).toBe(
      ketten.hans,
    )
    expect(wurzel.style.getPropertyValue('--font-anzeige-hant')).toBe(
      ketten.hant,
    )
  })

  it('setzt eine unbekannte Wahl in allen drei Reihen auf Pergament', () => {
    setzeAnzeigeschrift(wurzel, 'comic')

    expect(wurzel.style.getPropertyValue('--font-anzeige-hans')).toBe(
      schriftketten('pergament').hans,
    )
  })

  it('legt die Textschrift in allen drei Reihen über den Text — und nimmt alle drei wieder weg', () => {
    setzeSchriftbereich(wurzel, 'ganz', 'rund')

    const ketten = schriftketten('rund')
    expect(wurzel.style.getPropertyValue('--font-text')).toBe(ketten.ja)
    expect(wurzel.style.getPropertyValue('--font-text-hans')).toBe(ketten.hans)
    expect(wurzel.style.getPropertyValue('--font-text-hant')).toBe(ketten.hant)

    setzeSchriftbereich(wurzel, 'ueberschriften', 'rund')

    expect(wurzel.style.getPropertyValue('--font-text')).toBe('')
    expect(wurzel.style.getPropertyValue('--font-text-hans')).toBe('')
    expect(wurzel.style.getPropertyValue('--font-text-hant')).toBe('')
  })
})

// **Das Stilblatt wählt die Reihe** — und seine Vorgaben gelten, bis die
// Einstellungen geladen sind. Gelesen wird der Text der Datei: jsdom wendet
// `:lang()` nicht an.
describe('das Stilblatt', () => {
  // **Zwei Dateien**: die Vorgaben der japanischen Reihe stehen in
  // `stil.css`, die übrigen Reihen und `:lang()` in `schriftreihen.css`.
  const stil = ['stil.css', 'schriftreihen.css']
    .map((datei) => readFileSync(join(import.meta.dirname, datei), 'utf8'))
    .join('\n')
    .replace(/\s+/g, ' ')

  // vorgabe liest den Wert einer Variablen aus dem Stilblatt.
  function vorgabe(name: string): string {
    return new RegExp(`${name}: ([^;]+);`).exec(stil)?.[1] ?? ''
  }

  // regel liest den Rumpf der Regel mit diesem Wähler.
  function regel(waehler: string): string {
    return stil.split(`${waehler} {`)[1]?.split('}')[0] ?? ''
  }

  it.each([
    [':lang(zh)', 'hans'],
    [':lang(zh-Hant), :lang(zh-TW), :lang(zh-HK), :lang(zh-MO)', 'hant'],
  ])('wählt unter %s die Reihe %s — auch für Mantine', (waehler, reihe) => {
    const rumpf = regel(waehler)

    expect(rumpf).toContain(`--font-anzeige: var(--font-anzeige-${reihe});`)
    expect(rumpf).toContain(`--font-text: var(--font-text-${reihe});`)
    expect(rumpf).toContain(`--mantine-font-family: var(--font-text-${reihe});`)
    expect(rumpf).toContain(
      `--mantine-font-family-headings: var(--font-anzeige-${reihe});`,
    )
  })

  it('hält die Vorgaben der drei Reihen gleich', () => {
    expect(vorgabe('--font-anzeige-hans')).toBe(
      cjkUmordnen(vorgabe('--font-anzeige'), 'hans'),
    )
    expect(vorgabe('--font-anzeige-hant')).toBe(
      cjkUmordnen(vorgabe('--font-anzeige'), 'hant'),
    )
    expect(vorgabe('--font-text-hans')).toBe(
      cjkUmordnen(vorgabe('--font-text'), 'hans'),
    )
    expect(vorgabe('--font-text-hant')).toBe(
      cjkUmordnen(vorgabe('--font-text'), 'hant'),
    )
  })
})
