// Prüfungen von scripts/huellenwege.mjs — P0
// (#2336).
//
// **Ein Zähler, der einen Sichtwechsel übersieht oder einen zu viel zählt,
// macht jeden Vergleich vorher/nachher falsch** — geprüft an einer
// nachgestellten Probe, deren Blick sich auf Zuruf ändert.

import { test } from 'node:test'
import assert from 'node:assert/strict'

import { sichtschluessel, tabBis, wegzaehler } from './huellenwege.mjs'

// probe zeigt `blicke[0]` und nach der n-ten Handlung `blicke[n]` — so oft
// der Zähler auch fragt — und schreibt mit.
function probe(blicke) {
  const rufe = []
  const geprueft = []
  let stelle = 0
  const merke = (name) => async (...werte) => { rufe.push([name, ...werte]); stelle += 1 }

  return {
    rufe,
    geprueft,
    lauf: { bauabdruck: 'abcdef0123456789' },
    seite: async (js) => {
      rufe.push(['seite'])
      if (js.includes('sichtschluessel') || js.includes('main h1')) {
        return blicke[Math.min(stelle, blicke.length - 1)]
      }
      return true
    },
    klickeAuf: merke('klickeAuf'),
    taste: merke('taste'),
    tippeInsFeld: merke('tippeInsFeld'),
    klick: merke('klick'),
    klickeText: merke('klickeText'),
    dialogantwort: merke('dialogantwort'),
    pruefe: (satz, gilt) => { geprueft.push([satz, gilt]) },
    zeige: () => {},
  }
}

test('Handlungen, Sichtwechsel und Tabstellen kommen in die Zeile', async () => {
  // Blick beim Beginn, dann nach jeder der fünf Handlungen.
  const p = probe(['Start', 'Start', 'Spielen', 'Spielen', 'Spielen|Schublade', 'Tisch'])
  // Der Zähler fragt je Handlung mehrmals; gezählt wird nur der Blick.
  const datensatz = { wege: [] }
  let jetzt = 0
  const z = wegzaehler(p, datensatz, () => jetzt, { takt: 0, frist: 20 })

  await z.beginne('S1', 'bereich', 'tastatur')
  await z.taste('Tab')
  await z.taste('4', 'befehl')
  await z.taste('Tab')
  await z.klick('Partie anfangen')
  await z.klick('Anfangen')
  jetzt = 1500

  const zeile = await z.ende({ wirkung: true })

  assert.deepEqual(zeile, {
    weg: 'S1', eingang: 'bereich', art: 'tastatur',
    handlungen: 5, sichtwechsel: 3, tabstellen: 2, fehlversuche: 0,
    maschinenzeitMs: 1500, wirkung: true, bedingt: [],
  })
  assert.deepEqual(datensatz.wege, [zeile])
})

test('der Blick enthält die Sicht am Tisch — Audit C zählt ihren Wechsel', () => {
  // **Im Hot-Seat ist »Sicht für« der Wechsel** (S2, S4, S8): die
  // Überschrift bleibt »Am Tisch«, und ohne die Wahl im Schlüssel zählte
  // die Probe dort null Sichtwechsel, wo das Audit zwei nennt.
  assert.match(sichtschluessel, /Sicht für/)
})

// zeitprobe stellt eine Hülle nach, deren Blick **später** wechselt als die
// Handlung zurückkehrt — wie eine Navigation, die erst einen Takt nach dem
// angekommenen Klick zeichnet (Befund 5 des Zweitblicks über Phase 0).
//
// `ab` ist die Zahl der Abfragen, bis der Wechsel ankommt; `sofort` ist der
// Blick, den eine Taste **ohne** Verzögerung bringt.
function zeitprobe({ ab = 2, sofort = null } = {}) {
  const stand = { blick: 'Start', ausstehend: null, gelesen: 0 }
  const verzoegert = (neu) => async () => { stand.ausstehend = neu; stand.gelesen = 0 }

  return Object.assign(stand, {
    lauf: { bauabdruck: 'abcdef0123456789' },
    seite: async (js) => {
      if (js.includes('sichtschluessel')) {
        stand.gelesen += 1
        if (stand.ausstehend !== null && stand.gelesen >= ab) {
          stand.blick = stand.ausstehend
          stand.ausstehend = null
        }
        return stand.blick
      }
      return true
    },
    klickeText: verzoegert('Sammlung'),
    taste: sofort === null ? verzoegert('Spielen') : async () => { stand.blick = sofort },
    pruefe: () => {},
    zeige: () => {},
  })
}

test('ein Sichtwechsel, der nach der Handlung ankommt, zählt für sie', async () => {
  const p = zeitprobe()
  const datensatz = { wege: [] }
  const z = wegzaehler(p, datensatz, () => 0, { takt: 1, frist: 50 })

  await z.beginne('X', 'bereich', 'maus')
  await z.text('Sammlung')
  const zeile = await z.ende({ wirkung: true })

  assert.equal(zeile.sichtwechsel, 1)
})

// **Zwei gleiche alte Blicke sind noch keine Ruhe** (Befund 3 des
// Nachlaufs über Phase 0): kommt der Wechsel erst bei der vierten Abfrage,
// hörte der Zähler nach der zweiten auf, und der Wechsel fiel mit dem der
// nächsten Handlung in einen.
test('ein Wechsel nach mehreren unveränderten Abfragen zählt für seine Handlung', async () => {
  const p = zeitprobe({ ab: 4, sofort: 'Spielen' })
  const datensatz = { wege: [] }
  const z = wegzaehler(p, datensatz, () => 0, { takt: 1, frist: 50 })

  await z.beginne('X', 'bereich', 'maus')
  await z.text('Sammlung')
  await z.taste('4', 'befehl')
  const zeile = await z.ende({ wirkung: true })

  assert.equal(zeile.sichtwechsel, 2)
})

// **Und einer, der erst nach der Frist kommt, zählt vor der nächsten** — der
// Blick wird vor jeder Handlung noch einmal gelesen.
test('ein Wechsel nach der Frist zählt vor der nächsten Handlung und nicht mit ihr', async () => {
  const p = zeitprobe({ ab: 30, sofort: 'Spielen' })
  const datensatz = { wege: [] }
  const z = wegzaehler(p, datensatz, () => 0, { takt: 1, frist: 5 })

  await z.beginne('X', 'bereich', 'maus')
  await z.text('Sammlung')
  while (p.ausstehend !== null) await p.seite('sichtschluessel')
  await z.taste('4', 'befehl')
  const zeile = await z.ende({ wirkung: true })

  assert.equal(zeile.sichtwechsel, 2)
})

test('der Abschlussblick zählt, auch wenn er nach der letzten Handlung kommt', async () => {
  const p = zeitprobe()
  const datensatz = { wege: [] }
  const z = wegzaehler(p, datensatz, () => 0, { takt: 1, frist: 0 })

  await z.beginne('X', 'bereich', 'tastatur')
  await z.taste('4', 'befehl')
  p.blick = 'Spielen'
  p.ausstehend = null
  const zeile = await z.ende({ wirkung: true })

  assert.equal(zeile.sichtwechsel, 1)
})

// **Ein Griff, der vom Stand abhängt, wird genannt** — gegriffen oder
// ausgelassen (Befund 3 des Nachlaufs über Phase 0).
test('bedingte Griffe stehen mit Grund in der Zeile und beginnen je Weg leer', async () => {
  const p = probe(['A'])
  const datensatz = { wege: [] }
  const z = wegzaehler(p, datensatz, () => 0, { takt: 0, frist: 0 })

  await z.beginne('H5', 'handel', 'maus')
  z.bedingt('Schritt 3 war nicht offen', true)
  await z.klick('Schritt 3')
  z.bedingt('der Handel war schon gewählt', false)
  const erste = await z.ende({ wirkung: true })

  await z.beginne('H6', 'handel', 'maus')
  const zweite = await z.ende({ wirkung: true })

  assert.deepEqual(erste.bedingt, [
    { grund: 'Schritt 3 war nicht offen', gegriffen: true },
    { grund: 'der Handel war schon gewählt', gegriffen: false },
  ])
  assert.deepEqual(zweite.bedingt, [])
})

test('ohne Wirkung steht die Zeile da, und die Prüfung ist rot', async () => {
  const p = probe(['A'])
  const datensatz = { wege: [] }
  const z = wegzaehler(p, datensatz, () => 0, { takt: 0, frist: 0 })

  await z.beginne('S5', 'tisch', 'maus')
  await z.klick('Tappen')
  await z.ende({ wirkung: false })

  assert.equal(datensatz.wege[0].wirkung, false)
  assert.deepEqual(p.geprueft.at(-1), ['S5 (tisch, maus): Wirkung beobachtet', false])
})

test('ein Weg ohne Gang steht mit Status und Grund da', () => {
  const datensatz = { wege: [] }
  const z = wegzaehler(probe(['A']), datensatz, () => 0)

  z.ohneGang('S11', 'tisch', 'maus', 'FEHLT', 'kein Befehl nimmt einen Zug zurück')

  assert.deepEqual(datensatz.wege, [{
    weg: 'S11', eingang: 'tisch', art: 'maus', status: 'FEHLT', grund: 'kein Befehl nimmt einen Zug zurück',
  }])
})

// fokusprobe stellt eine Hülle nach, deren Taste **später ankommt**: ein
// Druck wird erst wirksam, wenn die Seite das zweite Mal gefragt wird, ob
// sich der Fokus bewegt hat. So war es am 2026-10-13 im echten Fenster —
// tabBis sah den alten Fokus, drückte noch einmal, und der Name landete in
// der Deckwahl statt im Namensfeld.
function fokusprobe(ziel) {
  let unterwegs = 0
  let gefragt = 0
  let vorTab = 0
  const stand = { gedrueckt: 0, fokus: 0 }

  stand.taste = async () => { stand.gedrueckt += 1; unterwegs += 1; gefragt = 0 }
  stand.seite = async (js) => {
    if (js.includes('__vorTab =')) { vorTab = stand.fokus; return true }
    if (js.includes('!== window.__vorTab')) {
      gefragt += 1
      if (gefragt >= 2 && unterwegs > 0) { stand.fokus += unterwegs; unterwegs = 0 }
      return stand.fokus !== vorTab
    }
    return stand.fokus === ziel
  }

  return stand
}

test('tabBis drückt erst weiter, wenn der vorige Druck angekommen ist', async () => {
  const p = fokusprobe(3)

  assert.equal(await tabBis(p, 'ziel', 10, 200), 3)
  assert.equal(p.gedrueckt, 3)
  assert.equal(p.fokus, 3)
})

test('tabBis gibt -1, wenn das Ziel nie kommt, und hört nach `hoechstens` auf', async () => {
  const p = fokusprobe(99)

  assert.equal(await tabBis(p, 'ziel', 5, 200), -1)
  assert.equal(p.gedrueckt, 5)
})

test('tabBis wartet höchstens die Frist, wenn ein Druck den Fokus nicht bewegt', async () => {
  let gedrueckt = 0
  const p = {
    taste: async () => { gedrueckt += 1 },
    seite: async (js) => (js.includes('!== window.__vorTab') ? false : js.includes('__vorTab =') ? true : false),
  }
  const beginn = Date.now()

  assert.equal(await tabBis(p, 'ziel', 2, 50), -1)
  assert.equal(gedrueckt, 2)
  assert.ok(Date.now() - beginn < 1000)
})

test('tabBis drückt über den übergebenen Weg, damit ein Zähler die Tabstellen sieht', async () => {
  const p = fokusprobe(2)
  let gezaehlt = 0

  assert.equal(await tabBis(p, 'ziel', 10, 200, async () => { gezaehlt += 1; await p.taste('Tab') }), 2)
  assert.equal(gezaehlt, 2)
})
