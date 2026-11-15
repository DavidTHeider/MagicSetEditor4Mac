// Prüfungen von scripts/uirevuezustaende.mjs — AP-13
// (#2339).
//
// **Ein Dichtemaß, das den befüllten Zustand nicht sieht, misst den
// Prospekt** (B-33). Geprüft wird, dass jeder Pflichtzustand in jeder Größe
// und jedem Erscheinungsbild dasteht — und dass ein fehlender rot wird.

import { readFileSync } from 'node:fs'
import { test } from 'node:test'
import assert from 'node:assert/strict'

import { erscheinungen, groessen, hauptgewichtBefunde, pflichtzustaende, pruefeZustaende, verloreneHandlungen } from './uirevuezustaende.mjs'

// vollstaendig baut einen Satz, in dem jeder Pflichtzustand gemessen ist.
function vollstaendig() {
  return {
    bauabdruck: 'abcdef012345',
    zeilen: groessen.flatMap((g) => erscheinungen.flatMap((e) => pflichtzustaende.map((z) => ({
      groesse: g.name, erscheinung: e, zustand: z.name, befuellt: z.befuellt, erreicht: true, seite: z.seite,
      bedien: 30, bedienImBild: 25, haupt: 2, hauptImBild: 1, hauptJeRegion: 1, felder: 8, felderImBild: 8,
      ueberschriften: 4, rolltiefe: 1.4, waagerecht: false,
    })))),
  }
}

test('ein vollständiger Satz ist ohne Befund', () => {
  assert.deepEqual(pruefeZustaende(vollstaendig()), [])
})

test('ein fehlendes offenes Deck wird rot — in jeder Größe und jedem Bild', () => {
  const satz = vollstaendig()
  satz.zeilen = satz.zeilen.filter((z) => z.zustand !== 'deck-offen')

  const befunde = pruefeZustaende(satz)

  assert.equal(befunde.length, groessen.length * erscheinungen.length)
  assert.match(befunde[0], /deck-offen: fehlt/)
})

test('ein Zustand, der nicht erreicht wurde, ist nicht gemessen', () => {
  const satz = vollstaendig()
  satz.zeilen.find((z) => z.zustand === 'karten-gross').erreicht = false

  assert.deepEqual(pruefeZustaende(satz), ['1440x838/hell/karten-gross: nicht erreicht'])
})

test('ohne Hauptknöpfe je Blick ist eine Zeile nicht gemessen', () => {
  // **R1 je Blick** (B-33): wie viele gefüllte Hauptknöpfe im Bild stehen,
  // nicht nur je Region — sonst bleiben drei führende Knöpfe untereinander
  // in drei Regionen grün.
  const satz = vollstaendig()
  delete satz.zeilen[0].hauptImBild

  assert.deepEqual(pruefeZustaende(satz), ['1440x838/hell/karten-gross: nicht gemessen'])
})

test('eine Zeile auf der falschen Seite ist nicht erreicht', () => {
  // **Eine gescheiterte Navigation erfüllt sonst einen Leerzustand**
  // (Befund 6 des Zweitblicks über Phase 0): „kein Probekarten-Text“ gilt
  // auf jeder anderen Seite auch.
  const satz = vollstaendig()
  satz.zeilen.find((z) => z.zustand === 'sammlung-leer').seite = 'start'

  assert.deepEqual(pruefeZustaende(satz), ['1440x838/hell/sammlung-leer: auf »start« statt »sammlung«'])
})

test('leere und befüllte Zustände sind beide Pflicht', () => {
  assert.ok(pflichtzustaende.some((z) => z.befuellt))
  assert.ok(pflichtzustaende.some((z) => !z.befuellt))
})

test('ohne Bauabdruck ist der Satz keinem Bündel zuzuordnen', () => {
  const satz = vollstaendig()
  delete satz.bauabdruck

  assert.equal(pruefeZustaende(satz).length, 1)
})

// **Der Wächter über den abgelegten Satz** — was er nicht hält: ob die
// Zahlen stimmen; das zeigen die Bilder des Laufs.
test('der abgelegte Satz AP-13 ist vollständig', () => {
  const satz = JSON.parse(readFileSync(new URL('../docs/belege/uirevue09-zustaende.json', import.meta.url), 'utf8'))

  assert.deepEqual(pruefeZustaende(satz), [])
})

// **Hauptgewicht und Handlungen** — AP-3
// (#2360),
// Befunde B-08 und B-10.

// mitNamen ergänzt einen vollständigen Satz um die Namen, die AP-3 misst.
function mitNamen() {
  const satz = vollstaendig()
  for (const z of satz.zeilen) Object.assign(z, { hauptNamen: ['Anlegen'], gewaehltGefuellt: [], handlungen: ['button: Anlegen', 'tab: Karten'] })
  return satz
}

test('ein Hauptknopf je Blick und kein gefüllter Umschalter ist ohne Befund', () => {
  assert.deepEqual(hauptgewichtBefunde(mitNamen()), [])
})

test('zwei gefüllte Hauptknöpfe im Bild werden mit Namen rot', () => {
  const satz = mitNamen()
  Object.assign(satz.zeilen[0], { hauptImBild: 2, hauptNamen: ['Ausgeben', '+ Karte anlegen'] })

  assert.deepEqual(hauptgewichtBefunde(satz), ['1440x838/hell/karten-gross: 2 gefüllt im Bild (Ausgeben, + Karte anlegen)'])
})

test('ein gewählter Umschalter, der gefüllt ist, wird rot — auch allein im Bild', () => {
  // **Die Verwechslung aus B-08**: „Kacheln“ gedrückt sah aus wie „+ Neues Set“.
  const satz = mitNamen()
  Object.assign(satz.zeilen[1], { hauptImBild: 1, hauptNamen: ['Liste'], gewaehltGefuellt: ['Liste'] })

  assert.deepEqual(hauptgewichtBefunde(satz), ['1440x838/hell/karten-leer: gewählt und gefüllt (Liste)'])
})

test('eine Zeile ohne Namen ist für das Hauptgewicht nicht gemessen', () => {
  const satz = mitNamen()
  delete satz.zeilen[2].hauptNamen

  assert.deepEqual(hauptgewichtBefunde(satz), ['1440x838/hell/deck-offen: ohne Namen gemessen'])
})

test('nachher fehlt keine Handlung von vorher', () => {
  assert.deepEqual(verloreneHandlungen(mitNamen(), mitNamen()), [])
})

test('eine verlorene Handlung wird je Zeile benannt', () => {
  const nachher = mitNamen()
  nachher.zeilen[0].handlungen = ['tab: Karten']

  assert.deepEqual(verloreneHandlungen(mitNamen(), nachher), ['1440x838/hell/karten-gross: button: Anlegen fehlt'])
})

test('eine Zeile, die nachher nicht erreicht ist, verliert alles', () => {
  // **Ein Weg, der nachher scheitert, ist kein Vergleich** — ohne die Zeile
  // stünde dort „nichts verloren“.
  const nachher = mitNamen()
  nachher.zeilen[3].erreicht = false

  assert.deepEqual(verloreneHandlungen(mitNamen(), nachher), ['1440x838/hell/decks-leer: nachher nicht erreicht'])
})

test('ein Satz ohne Handlungen ist nicht vergleichbar', () => {
  const vorher = mitNamen()
  delete vorher.zeilen[4].handlungen

  assert.deepEqual(verloreneHandlungen(vorher, mitNamen()), ['1440x838/hell/sammlung-voll: vorher ohne Handlungen gemessen'])
})

test('die Zeilen von uirevue über alle Seiten haben Größe und Seite als Kennung', () => {
  const vorher = { zeilen: [{ groesse: 'eng', ziel: 'Sets', erreicht: true, handlungen: ['button: + Neues Set'] }] }
  const nachher = { zeilen: [{ groesse: 'eng', ziel: 'Sets', erreicht: true, handlungen: [] }] }

  assert.deepEqual(verloreneHandlungen(vorher, nachher), ['eng/Sets: button: + Neues Set fehlt'])
})

test('eine Uhrzeit im Namen ist dieselbe Handlung', () => {
  // **Der erste Nachherlauf von AP-3 meldete eine verlorene Handlung**, die
  // dastand: der Eintrag auf *Start* nennt, wann das Set zuletzt offen war,
  // und das war im Vorherlauf eine andere Minute.
  const vorher = { zeilen: [{ groesse: 'eng', ziel: 'Start', erreicht: true, handlungen: ['button: Werkstatt-Testkarten Set · geöffnet · 16.10.2026, 18:48'] }] }
  const nachher = { zeilen: [{ groesse: 'eng', ziel: 'Start', erreicht: true, handlungen: ['button: Werkstatt-Testkarten Set · geöffnet · 16.10.2026, 19:19'] }] }

  assert.deepEqual(verloreneHandlungen(vorher, nachher), [])
})

test('ohne Uhrzeit bleibt ein anderer Name eine verlorene Handlung', () => {
  const vorher = { zeilen: [{ groesse: 'eng', ziel: 'Start', erreicht: true, handlungen: ['button: Set 12 öffnen'] }] }
  const nachher = { zeilen: [{ groesse: 'eng', ziel: 'Start', erreicht: true, handlungen: ['button: Set 13 öffnen'] }] }

  assert.deepEqual(verloreneHandlungen(vorher, nachher), ['eng/Start: button: Set 12 öffnen fehlt'])
})

// **Die Wächter über die abgelegten Sätze von AP-3** — was sie nicht halten:
// ob der eine gefüllte Knopf der richtige ist; das zeigen die Bilder.
const satzAus = (pfad) => JSON.parse(readFileSync(new URL(pfad, import.meta.url), 'utf8'))

test('im abgelegten Satz führt je Blick ein Knopf, und kein gewählter ist gefüllt', () => {
  assert.deepEqual(hauptgewichtBefunde(satzAus('../docs/belege/uirevue09-zustaende.json')), [])
  assert.deepEqual(hauptgewichtBefunde(satzAus('../docs/belege/uirevue-seiten.json')), [])
})

// **Gegen den Satz, mit dem AP-3 abschloss** (`ap3/…-nachher.json`, aus
// `03189a186`) — und nicht gegen den heutigen: AP-6 hat seither Handlungen
// mit Absicht hinter „…“, in Kästen und Schubladen gelegt, und ein Vergleich
// mit dem jeweils neuesten Satz hielte AP-3 dafür verantwortlich. Was AP-6
// verlegt, hält der Wächter darunter.
test('gegen den Vorherlauf von AP-3 ist keine Handlung verloren', () => {
  assert.deepEqual(verloreneHandlungen(satzAus('../docs/belege/ap3/uirevue09-zustaende-vorher.json'), satzAus('../docs/belege/ap3/uirevue09-zustaende-nachher.json')), [])
  assert.deepEqual(verloreneHandlungen(satzAus('../docs/belege/ap3/uirevue-seiten-vorher.json'), satzAus('../docs/belege/uirevue-seiten.json')), [])
})

test('der Vorherlauf von AP-3 war rot — sonst belegte der Vergleich nichts', () => {
  assert.ok(hauptgewichtBefunde(satzAus('../docs/belege/ap3/uirevue09-zustaende-vorher.json')).length > 0)
  assert.ok(hauptgewichtBefunde(satzAus('../docs/belege/ap3/uirevue-seiten-vorher.json')).length > 0)
})

// **AP-6 Teil 3 verlegt — und nennt, wohin** (#2363). Gegen den Satz davor
// darf nur fehlen, was hier mit seinem neuen Ort steht; `ap6deckwege09`
// öffnet jeden dieser Orte am nativen Fenster. Eine Handlung, die sonst
// fehlt, ist verloren — laut, wie bei AP-3.
const verlegtInAp6 = {
  'button: Liste': 'Segment „Ansicht“ (Optionsfeld)',
  'button: Kacheln': 'Segment „Ansicht“ (Optionsfeld)',
  'tab: Decks': 'die Seite selbst — es gibt keine zweite Lasche mehr',
  'tab: Formate': 'Schublade hinter „Formate …“ im Kopf',
  'button: Ein Deck aus einem fremden Profil holen': 'Kasten „Holen und weitergeben“',
  'button: Als Umschlag weitergeben': 'Kasten „Holen und weitergeben“',
  'button: Meiner Sammlung': 'Kasten „Holen und weitergeben“',
  'button: Einem Set': 'Kasten „Holen und weitergeben“',
  'button: Eine Deckliste einlesen': 'Menü „…“ des offenen Decks, Schublade',
  'button: Öffnen: Probedeck': 'Marke „offen“ — das Deck ist offen',
  'button: Nach Marken': 'erscheint, sobald ein Eintrag eine Marke trägt',
}

test('gegen den Satz vor AP-6 Teil 3 fehlt nur, was verlegt ist', () => {
  const verloren = verloreneHandlungen(satzAus('../docs/belege/uirevue09-zustaende-vor-ap6-teil3.json'), satzAus('../docs/belege/uirevue09-zustaende.json'))
  const unerklaert = verloren.filter((befund) => !(befund.replace(/^[^:]+: /, '').replace(/ fehlt$/, '') in verlegtInAp6))
  assert.deepEqual(unerklaert, [])
})
