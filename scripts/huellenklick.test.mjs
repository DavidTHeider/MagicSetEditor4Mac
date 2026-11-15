// Prüfungen von scripts/huellenklick.mjs — R04
// (#2094).
//
// **Ein Klick, der nicht traf, ist ein Befund mit Namen** — nicht ein
// Drehbuch, das drei Sekunden später an einer Falte scheitert, die nie
// aufging. Geprüft gegen eine Ersatzprobe, die meldet, was die Seite meldete.

import { test } from 'node:test'
import assert from 'node:assert/strict'

import { klappeAuf, klickeAuf } from './huellenklick.mjs'

// ersatz ist eine Probe, deren Seite `klick` als Treffer meldet und deren
// Falte `offen` sagt.
function ersatz({ klick, offen = 'true', lesefehler = false, neueSeite = false }) {
  const probe = {
    geklickt: 0,
    async warte(beschreibung, bedingung, frist = 1000) {
      const ende = Date.now() + frist
      for (;;) {
        const wert = await bedingung()
        if (wert) return wert
        if (Date.now() > ende) throw new Error(`abgebrochen: ${beschreibung}`)
        await new Promise((los) => setTimeout(los, 10))
      }
    },
    async seite(js) {
      if (js.startsWith('const e = (')) return { x: 10, y: 20 }
      if (js.includes('window.__mseKlick') && !js.includes('addEventListener')) {
        if (lesefehler) throw new Error('TimeoutError: die Leitung antwortet nicht')
        if (neueSeite && probe.geklickt > 0) return { traf: true, was: 'eine neue Seite' }

        return probe.geklickt > 0 ? klick : null
      }
      if (js.includes("getAttribute('aria-expanded')")) return offen === 'true'

      return true
    },
    async frage(befehl) {
      probe.geklickt += 1
      probe.befehl = befehl

      return { ok: true }
    },
    klickeAuf(ausdruck, beschreibung) { return klickeAuf(probe, ausdruck, beschreibung) },
  }

  return probe
}

test('ein Klick, der sein Ziel trifft, geht durch', async () => {
  await klickeAuf(ersatz({ klick: { traf: true, was: 'Neues Paket anlegen' } }), 'z', 'den Kasten')
})

test('ein Klick, der etwas anderes trifft, nennt es', async () => {
  await assert.rejects(klickeAuf(ersatz({ klick: { traf: false, was: 'Pakete' } }), 'z', 'den Kasten'), /der Klick auf den Kasten traf »Pakete« \(10, 20\)/)
})

// **Ein ⌘-Klick wählt am Tisch, ein Doppelklick stellt die Zweithandlung**
// (P4, #2372) — Umschalter und Klickzahl gehen mit dem Klick hinaus.
test('ein Klick trägt Umschalter und Klickzahl hinaus', async () => {
  const probe = ersatz({ klick: { traf: true, was: 'Moosbär' } })
  await klickeAuf(probe, 'z', 'den Bären', { mit: ['befehl'], anzahl: 2 })

  assert.deepEqual(probe.befehl, { was: 'klicke', x: 10, y: 20, mit: ['befehl'], anzahl: 2 })
})

test('ein schlichter Klick schickt keine Umschalter', async () => {
  const probe = ersatz({ klick: { traf: true, was: 'Moosbär' } })
  await klickeAuf(probe, 'z', 'den Bären')

  assert.deepEqual(probe.befehl, { was: 'klicke', x: 10, y: 20 })
})

test('ein Klick, der nicht ankommt, sagt es', async () => {
  await assert.rejects(klickeAuf(ersatz({ klick: null }), 'z', 'den Kasten'), /der Klick auf den Kasten kam nicht an/)
})

test('eine Falte, die nach dem Klick zu bleibt, ist ein Befund', async () => {
  const probe = ersatz({ klick: { traf: true, was: 'Neues Paket anlegen' }, offen: 'false' })
  await assert.rejects(klappeAuf(probe, 'z', 'den Kasten'), /abgebrochen: den Kasten ist offen/)
})

// **Ein Lesefehler ist kein Treffer** — Zweitblick über A und B, Befund 5:
// der Leser fing jeden Fehler und meldete „traf", auch eine Leitung, die
// gar nicht antwortete.
test('ein Klickbeleg, der sich nicht lesen lässt, ist ein Befund und kein Treffer', async () => {
  await assert.rejects(klickeAuf(ersatz({ klick: null, lesefehler: true }), 'z', 'den Kasten'), /der Klickbeleg zu den Kasten ließ sich nicht lesen/)
})

// **Ein Klick, der die Seite wechselt, gilt** — die neue Seite kennt die
// Marke der alten nicht, und genau daran ist er zu erkennen.
test('ein Klick, der eine neue Seite lädt, gilt als angekommen', async () => {
  await klickeAuf(ersatz({ klick: null, neueSeite: true }), 'z', 'den Kasten')
})
