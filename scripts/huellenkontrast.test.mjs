// Prüfungen der Kontrastrechnung der Hüllenprobe — ABN-01.AK-03
// (#2175).
//
// Die Rechnung läuft in der Webansicht; hier läuft sie gegen eine Seite aus
// Attrappen, deren Canvas `rgb()` und `rgba()` versteht — genug, um die
// Schichtung des Grunds, den Rand gegen den Grund außerhalb und das
// fehlende Element zu prüfen. **Was sie nicht prüft**: `oklch()`; das
// übersetzt in der Anwendung WebKit selbst.

import { test } from 'node:test'
import assert from 'node:assert/strict'

import { kontraste, rechnung } from './huellenkontrast.mjs'

function seite() {
  const leinwand = {
    width: 1,
    height: 1,
    getContext: () => {
      let farbe = 'transparent'
      return {
        clearRect: () => {},
        fillRect: () => {},
        set fillStyle(wert) { farbe = wert },
        getImageData: () => {
          const zahlen = farbe.match(/[\d.]+/g)?.map(Number) ?? [0, 0, 0, 0]
          const alpha = farbe === 'transparent' ? 0 : zahlen.length === 4 ? zahlen[3] : 1

          return { data: [zahlen[0] ?? 0, zahlen[1] ?? 0, zahlen[2] ?? 0, Math.round(alpha * 255)] }
        },
      }
    },
  }
  const blaetter = []
  const dokument = {
    styleSheets: blaetter,
    createElement: (art) => (art === 'canvas' ? leinwand : { style: {}, get stil() { return { color: this.style.color } }, remove: () => {} }),
  }
  const stil = (element, pseudo) => element.stil === undefined && 'style' in element
    ? { color: element.style.color }
    : pseudo === '::placeholder'
      ? { color: element.platzhalter ?? element.stil.color ?? 'rgb(0, 0, 0)' }
      : { color: 'rgb(0, 0, 0)', backgroundColor: 'transparent', borderTopColor: 'rgb(0, 0, 0)', borderTopWidth: '1px', fill: 'rgb(0, 0, 0)', stroke: 'none', ...element.stil }

  // eslint-disable-next-line no-new-func -- die Rechnung ist für die Seite geschrieben und läuft hier wie dort als Text
  const { messe } = new Function('document', 'getComputedStyle', `${rechnung}\n return { messe }`)(dokument, stil)

  return { messe, blaetter }
}

const element = (stil, parentElement = null) => ({ stil, parentElement, matches: () => true, appendChild: () => {} })
const weiss = element({ backgroundColor: 'rgb(255, 255, 255)' })

test('schwarz auf weiß ist 21 : 1, das Grau der Grenze 4,54 : 1', () => {
  const { messe } = seite()

  assert.equal(messe('schwarz', element({ color: 'rgb(0, 0, 0)' }, weiss), 'text', 4.5).verhaeltnis, 21)
  assert.equal(messe('grau', element({ color: 'rgb(118, 118, 118)' }, weiss), 'text', 4.5).verhaeltnis, 4.54)
})

test('der Grund kommt vom nächsten Vorfahren mit Farbe, und ein halb durchsichtiger zählt anteilig', () => {
  const { messe } = seite()
  const schwarz = element({ backgroundColor: 'rgb(0, 0, 0)' }, weiss)
  const span = element({ color: 'rgb(255, 255, 255)' }, element({}, schwarz))

  assert.equal(messe('weiß im span', span, 'text', 4.5).verhaeltnis, 21)

  // Schwarz zu fünfzig Prozent über Weiß: ein Canvas hält Alpha in acht
  // Bit, also 128/255, und darunter bleibt ein Grau von 127,0 — gegen das
  // steht schwarzer Text bei 5,24, nicht bei 21.
  const schleier = element({ backgroundColor: 'rgba(0, 0, 0, 0.5)' }, weiss)
  assert.equal(messe('unter dem Schleier', element({ color: 'rgb(0, 0, 0)' }, schleier), 'text', 4.5).verhaeltnis, 5.24)
})

test('ein Rand zählt gegen den Grund außerhalb, nicht gegen die eigene Fläche', () => {
  const { messe } = seite()
  // Ein weißer Rand um eine schwarze Fläche auf weißem Grund ist nicht zu
  // sehen — gegen die eigene Fläche gemessen stünde dort 21.
  const kasten = element({ backgroundColor: 'rgb(0, 0, 0)', borderTopColor: 'rgb(255, 255, 255)' }, weiss)

  assert.equal(messe('Rand', kasten, 'rand', 3).verhaeltnis, 1)
  assert.equal(messe('Trennlinie', element({ backgroundColor: 'rgb(118, 118, 118)' }, weiss), 'flaeche', 3).verhaeltnis, 4.54)
})

// **Ein Rand ohne Breite ist keiner** — sonst misst die Rechnung die
// Randfarbe, die CSS dann trägt (`currentColor`, also die Schrift), und ein
// falsch gezieltes Element hielt scheinbar 13 : 1. Gefunden im ersten Lauf
// von `politur` mit Kontrasten.
test('ein Rand ohne Breite fehlt — und gibt keinen Wert der Schriftfarbe', () => {
  const { messe } = seite()
  const ohne = element({ color: 'rgb(0, 0, 0)', borderTopColor: 'rgb(0, 0, 0)', borderTopWidth: '0px' }, weiss)

  assert.deepEqual(messe('kein Rand', ohne, 'rand', 3), { name: 'kein Rand', art: 'rand', grenze: 3, fehlt: true })
})

test('ein Element, das es nicht gibt, fehlt — und ist kein Verhältnis', () => {
  const { messe } = seite()

  assert.deepEqual(messe('weg', null, 'text', 4.5), { name: 'weg', art: 'text', grenze: 4.5, fehlt: true })
})

// **Alle vier Seiten und der Platzhalter** — Nachlauf über Phase A (A03):
// bis dahin maß die Rechnung nur die obere Kante, und ein Feld ohne Wert gar
// nicht.
test('ein Rand wird an der genannten Seite gemessen', () => {
  const { messe } = seite()
  const schwarz = element({ backgroundColor: 'rgb(0, 0, 0)' })
  const linie = element({ borderTopWidth: '0px', borderBottomWidth: '1px', borderBottomColor: 'rgb(255, 255, 255)' }, schwarz)

  assert.equal(messe('unten', linie, 'rand', 3, 'Bottom').verhaeltnis, 21)
  assert.equal(messe('oben', linie, 'rand', 3).fehlt, true)
})

test('ein Platzhalter zählt gegen den Grund des Felds — und ohne Farbe fehlt er', () => {
  const { messe } = seite()
  const feld = { ...element({ backgroundColor: 'rgb(255, 255, 255)' }, weiss), platzhalter: 'rgb(118, 118, 118)' }

  assert.equal(messe('Platzhalter', feld, 'platzhalter', 4.5).verhaeltnis, 4.54)
  assert.equal(messe('ohne', element({ backgroundColor: 'rgb(255, 255, 255)' }, weiss), 'platzhalter', 4.5).fehlt, true)
})

// **WebKit gibt für `::placeholder` die Farbe des Felds zurück** — gemessen
// unter macOS 15.2 (Gegenprobe A03): dann gilt die Regel, die greift, und
// nicht das Pseudo-Element, das die Feldfarbe noch einmal sagte.
test('gibt das WebKit die Feldfarbe zurück, gilt die greifende Regel', () => {
  const { messe, blaetter } = seite()
  blaetter.push({ cssRules: [{ selectorText: 'input::placeholder', style: { color: 'rgb(118, 118, 118)' } }] })
  const feld = element({ color: 'rgb(0, 0, 0)', backgroundColor: 'rgb(255, 255, 255)' }, weiss)

  assert.equal(messe('Platzhalter', feld, 'platzhalter', 4.5).verhaeltnis, 4.54)
})

test('ein Symbol zählt mit seiner Strich- oder Füllfarbe', () => {
  const { messe } = seite()
  const grau = element({ backgroundColor: 'rgb(255, 255, 255)' })

  assert.equal(messe('Strich', element({ stroke: 'rgb(118, 118, 118)', fill: 'none' }, grau), 'symbol', 3).verhaeltnis, 4.54)
  assert.equal(messe('Fläche', element({ stroke: 'none', fill: 'rgb(0, 0, 0)' }, grau), 'symbol', 3).verhaeltnis, 21)
  assert.equal(messe('nichts', element({ stroke: 'none', fill: 'none' }, grau), 'symbol', 3).fehlt, true)
})

// **Die Seite reist mit** — AP-3 (#2360): die Unterlinie einer Lasche liegt
// unten, und `kontraste` rief `messe` bis dahin immer mit der oberen Kante.
test('kontraste reicht die genannte Seite eines Rands durch', async () => {
  let skript = ''
  await kontraste({ seite: (s) => { skript = s; return [] } }, [['unten', 'x', 'rand', 3, 'Bottom'], ['oben', 'y', 'rand', 3]])

  assert.match(skript, /messe\("unten", x, "rand", 3, "Bottom"\)/)
  assert.match(skript, /messe\("oben", y, "rand", 3\)/)
})
