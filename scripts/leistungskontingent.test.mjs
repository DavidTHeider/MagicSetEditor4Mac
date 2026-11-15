// Prüfungen von scripts/leistungskontingent.mjs — Z1 des Zweitblicks über
// Phase 6.11.
//
// **Ein Leistungsvergleich misst nicht die Drosselung.** Im Beleg vom
// 2026-10-03 zeichnete der Batchfall 24 von 70 Bildern, weil das Guthaben von
// `mse-mcp` den Rest abwies — und das Dokument nannte 2,4 s und „null
// Befunde“. Abgeschaltet war das Guthaben nur im Langlauf.

import { test } from 'node:test'
import assert from 'node:assert/strict'

import { fahreTeile, kontingentschalter, kontingentstandAus, schalterFuer } from './leistungskontingent.mjs'

/** doppel gibt Lesen und Setzen, die einen Stand führen und jeden Schritt festhalten. */
function doppel(anfang = 'an', { setzenScheitertBeiAus = false } = {}) {
  const protokoll = []
  const stand = { wert: anfang }
  return {
    protokoll, stand,
    lies: async () => { protokoll.push('lies'); return stand.wert },
    setze: async (neu) => {
      protokoll.push(`setze ${neu}`)
      if (setzenScheitertBeiAus && neu === 'aus') throw new Error('Bestand gesperrt')
      stand.wert = neu
    },
  }
}

test('der Stand wird aus der Auskunft der Einstellungen gelesen', () => {
  const antwort = JSON.stringify({ einstellungen: [
    { schluessel: 'sprache', wert: 'de', wirksam: 'de' },
    { schluessel: 'grenzen.kontingente', wert: 'aus', wirksam: 'aus' },
  ] })
  assert.equal(kontingentstandAus(antwort), 'aus')
  assert.equal(kontingentstandAus(JSON.stringify({ einstellungen: [{ schluessel: 'grenzen.kontingente', wert: 'an' }] })), 'an')
})

test('ein Stand, den es nicht gibt, ist ein Fehler und kein Raten', () => {
  assert.throws(() => kontingentstandAus('{"einstellungen":[]}'), /grenzen\.kontingente/)
  assert.throws(() => kontingentstandAus('kein JSON'), /grenzen\.kontingente/)
  assert.throws(() => kontingentstandAus(JSON.stringify({ einstellungen: [{ schluessel: 'grenzen.kontingente', wert: 'vielleicht' }] })), /vielleicht/)
})

test('abgeschaltet wird nach dem Lesen des alten Standes, zurückgestellt auf genau den', async () => {
  const d = doppel('an')
  const schalter = kontingentschalter(d)
  await schalter.aus()
  assert.equal(d.stand.wert, 'aus')
  await schalter.zurueck()
  assert.equal(d.stand.wert, 'an')
  assert.deepEqual(d.protokoll, ['lies', 'setze aus', 'setze an'])
})

test('war es schon aus, bleibt es nach dem Zurückstellen aus', async () => {
  const d = doppel('aus')
  const schalter = kontingentschalter(d)
  await schalter.aus()
  await schalter.zurueck()
  assert.equal(d.stand.wert, 'aus')
})

test('zweimal abschalten liest und setzt einmal, zweimal zurückstellen ebenso', async () => {
  const d = doppel('an')
  const schalter = kontingentschalter(d)
  await schalter.aus()
  await schalter.aus()
  await schalter.zurueck()
  await schalter.zurueck()
  assert.deepEqual(d.protokoll, ['lies', 'setze aus', 'setze an'])
})

test('zurückstellen ohne vorheriges Abschalten ändert nichts', async () => {
  const d = doppel('an')
  await kontingentschalter(d).zurueck()
  assert.deepEqual(d.protokoll, [])
})

test('scheitert das Abschalten, wird zurückgestellt, was unklar sein könnte', async () => {
  const d = doppel('an', { setzenScheitertBeiAus: true })
  const schalter = kontingentschalter(d)
  await assert.rejects(schalter.aus(), /gesperrt/)
  await schalter.zurueck()
  assert.equal(d.stand.wert, 'an')
  assert.deepEqual(d.protokoll, ['lies', 'setze aus', 'setze an'])
})

test('die Teile laufen mit abgeschaltetem Guthaben, nur die Lastgrenze mit eingeschaltetem — einmal je Lauf', async () => {
  const d = doppel('an')
  const schalter = kontingentschalter(d)
  const stande = []
  await fahreTeile(['start', 'navigation', 'bilder', 'langlauf', 'last'], schalter, async (teil) => { stande.push([teil, d.stand.wert]) })
  assert.deepEqual(stande, [['start', 'aus'], ['navigation', 'aus'], ['bilder', 'aus'], ['langlauf', 'aus'], ['last', 'an']])
  assert.deepEqual(d.protokoll, ['lies', 'setze aus', 'setze an'], 'abgeschaltet einmal, zurückgestellt einmal')
  assert.equal(d.stand.wert, 'an')
})

test('wird nur die Lastgrenze gefahren, bleibt das Guthaben unberührt', async () => {
  const d = doppel('an')
  await fahreTeile(['last'], kontingentschalter(d), async () => {})
  assert.deepEqual(d.protokoll, [])
})

test('bricht ein Teil ab, steht das Guthaben danach wie vorher — der Fehler bleibt', async () => {
  const d = doppel('an')
  await assert.rejects(fahreTeile(['start', 'bilder', 'langlauf'], kontingentschalter(d), async (teil) => {
    if (teil === 'bilder') throw new Error('MSE2 stürzt ab')
  }), /stürzt ab/)
  assert.equal(d.stand.wert, 'an')
})

test('steht die Lastgrenze vor anderen Teilen, wird wieder abgeschaltet', async () => {
  const d = doppel('an')
  const stande = []
  await fahreTeile(['last', 'bilder'], kontingentschalter(d), async (teil) => { stande.push([teil, d.stand.wert]) })
  assert.deepEqual(stande, [['last', 'an'], ['bilder', 'aus']])
  assert.equal(d.stand.wert, 'an')
})

test('der Schalter für einen Teilnehmer liest und setzt in dessen Vorrat — und meldet, was misslingt', async () => {
  const aufrufe = []
  const stand = { wert: 'an' }
  const t = {
    name: 'heute', vorrat: '/b/vorrat.db',
    kontingenteStand: async (bestand) => {
      aufrufe.push(['stand', bestand])
      return { fehler: null, aus: JSON.stringify({ einstellungen: [{ schluessel: 'grenzen.kontingente', wert: stand.wert }] }) }
    },
    kontingente: async (bestand, neu) => {
      aufrufe.push(['setze', bestand, neu])
      stand.wert = neu
      return { fehler: null }
    },
  }
  const schalter = schalterFuer(t)
  await schalter.aus()
  await schalter.zurueck()
  assert.deepEqual(aufrufe, [['stand', '/b/vorrat.db'], ['setze', '/b/vorrat.db', 'aus'], ['setze', '/b/vorrat.db', 'an']])

  const kaputt = schalterFuer({
    ...t,
    kontingenteStand: async () => ({ fehler: 'Rückgabe 1', fehlerText: 'Bestand gesperrt', aus: '' }),
  })
  await assert.rejects(kaputt.aus(), /heute.*Kontingente.*Rückgabe 1.*Bestand gesperrt/s)

  const nichtSetzbar = schalterFuer({ ...t, kontingente: async () => ({ fehler: 'Rückgabe 2', fehlerText: 'nein' }) })
  await assert.rejects(nichtSetzbar.aus(), /nicht auf aus stellen.*Rückgabe 2.*nein/s)
})
