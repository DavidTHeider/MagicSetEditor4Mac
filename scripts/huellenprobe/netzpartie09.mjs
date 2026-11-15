// **Die Netzpartie mit Einladung und Gastdeck am nativen Fenster** — P9c
// (#2377),
// Befunde C-11 und C-12, Vermerke 0198 und 0199, Artboard `Netzpartie09`.
//
// Zwei Anwendungen auf diesem Mac, je mit eigenem Bestand — wie in `nat04`.
// **Die Bestände kollidieren mit Absicht**: der Wirt liest `tisch/TSP.json`,
// der Gast `probe.json`; unter denselben Zeilen liegen andere Karten.
//
// - **S12 und S13** ab dem Bereich gezählt (`huellenwege.mjs`), gegen
//   `ausgangswerte09-netz.json` (10 und 5 + 2): „Gast einladen“ und „Bei Wirt
//   mitspielen“ auf der Deckseite.
// - **Ohne Deck vor dem Anfang abgewiesen**: zuerst setzt sich der Gast mit
//   „Mit einem Deck des Wirts“ an den eingeladenen Sitz und liest den Grund.
// - **Das Gastdeck**: seine Hand zeigt auf **seine** Zeilen und wird bei ihm
//   gezeichnet; beim Wirt steht dieselbe Karte ohne Zeile und mit dem Bild,
//   das der Gast schickt — byteweise dasselbe.
// - **Kartenzug, Abbruch, Wiederverbindung**: eine Karte des Gastes auf den
//   Tisch, aufstehen, wieder hinsetzen — derselbe Sitz, dieselben Karten.
// - **Ein abgewiesener Wunsch** (#2478): der Grund steht am Tisch, und der
//   Gast sitzt weiter.
//
// **Was sie nicht belegt**: zwei Geräte in zwei Netzen (#1998), und
// VoiceOver.
import { writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { bindung } from '../huellenhandel.mjs'
import { rufbuchHolen } from '../huellenrufbuch.mjs'
import { knopf } from '../huellenspieltisch.mjs'
import { wegzaehler } from '../huellenwege.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const ziel = join(wurzel, 'docs/belege/netzpartie09-wege.json')
const leinwand = "return !!document.querySelector('canvas')?.checkVisibility()"

// koppeln macht zwei Anwendungen einander bekannt — bestätigt, mit Zweck.
async function koppeln(hier, dort, name) {
  await bindung(hier, `b.PeerkontaktAufnehmen(${JSON.stringify(dort.kennung)}, ${JSON.stringify(name)})`)
  const bestaetigt = await bindung(hier, `b.PeerkontaktBestaetigen(${JSON.stringify(dort.kennung)}, ${JSON.stringify(dort.fingerabdruck)})`)
  hier.pruefe(`${name} ist bestätigt`, !bestaetigt?.fehler, JSON.stringify(bestaetigt?.fehler ?? ''))
  const erlaubt = await bindung(hier, `b.PeerzweckErlauben(${JSON.stringify(dort.kennung)}, 'partie')`)
  hier.pruefe(`${name} darf eine Partie spielen`, !erlaubt?.fehler, JSON.stringify(erlaubt?.fehler ?? ''))
}

// abdruck liest ein Bild über seine Adresse — Länge und FNV-1a der Bytes.
const abdruck = (p, adresse) => p.seite(`const r = await fetch(${JSON.stringify(adresse)}); const b = new Uint8Array(await r.arrayBuffer()); let h = 0x811c9dc5; for (const x of b) { h ^= x; h = Math.imul(h, 0x01000193) >>> 0 } return r.ok ? b.length + ':' + h.toString(16) : 'HTTP ' + r.status`)

const rufe = (buch, name) => ({ anzahl: buch[name]?.anzahl ?? 0, gelungen: buch[name]?.ausgaenge?.gelungen ?? 0 })

const meldungen = (p) => p.seite("return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).filter(Boolean)")

export default async function netzpartie09(p) {
  const gast = new p.constructor('netzpartie09-gast')
  gast.zeilen = p.zeilen

  try {
    await zweiFenster(p, gast)
  } finally {
    p.fehler += gast.fehler
    gast.aufraeumen()
  }
}

async function zweiFenster(p, gast) {
  const wirtOrt = p.ort()
  const gastOrt = gast.ort()
  for (const [probe, ort] of [[p, wirtOrt], [gast, gastOrt]]) {
    probe.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
    probe.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  }
  p.mse(wirtOrt, 'mtgjson-lesen', 'testdaten/mtgjson/tisch/TSP.json', 'magic')
  p.mse(wirtOrt, 'deck', 'neu', 'Waldschritt', 'magic')
  for (const k of ['1', '3', '4', '5', '6']) p.mse(wirtOrt, 'deck', 'legen', '1', `--karte=${k}`, '8')
  gast.mse(gastOrt, 'mtgjson-lesen', 'testdaten/mtgjson/probe.json', 'magic')
  gast.mse(gastOrt, 'deck', 'neu', 'Funkenschlag', 'magic')
  for (const k of ['1', '2']) gast.mse(gastOrt, 'deck', 'legen', '1', `--karte=${k}`, '20')

  const datensatz = { bauabdruck: p.lauf?.bauabdruck ?? '', erhoben: new Date().toISOString(), fenster: '1440 × 900', wege: [] }
  const zw = wegzaehler(p, datensatz)
  const zg = wegzaehler(gast, datensatz)

  p.notiz('Zwei Anwendungen, einander bekannt — ungezählt, wie in den Ausgangswerten')
  await p.starte({ ort: wirtOrt })
  await p.groesse(1440, 900)
  await gast.starte({ ort: gastOrt })
  await gast.groesse(1440, 900)
  const wirt = await bindung(p, 'b.PeeridentitaetSicherstellen()')
  const bo = await bindung(gast, 'b.PeeridentitaetSicherstellen()')
  await koppeln(p, bo, 'Gast')
  await koppeln(gast, wirt, 'Wirt')
  const horch = await bindung(p, "b.HorchStarten('127.0.0.1:0')")
  const port = horch?.stand?.adresse?.split(':').pop()
  p.pruefe('der Wirt horcht', Boolean(horch?.stand?.laeuft && port), horch?.stand?.adresse ?? JSON.stringify(horch?.fehler))
  await bindung(gast, `b.PeerkontaktAdresse(${JSON.stringify(wirt.kennung)}, '127.0.0.1:${port}')`)

  p.notiz('S12 — der Wirt lädt ein: »Gast einladen« auf der Deckseite')
  await p.aktiviere()
  await zw.beginne('S12', 'bereich', 'maus')
  await zw.taste('4', 'befehl')
  await p.warte('»Gast einladen« steht', () => p.seite(`return !!${knopf('Gast einladen')}`), 20000)
  await zw.text('Gast einladen')
  await p.warte('die Leinwand des Wirts zeichnet', () => p.seite(leinwand), 30000)
  await p.wartetext('Dieser Tisch steht offen.', 15000).catch(() => false)
  const offen = await p.seite("return document.body.innerText.includes('Dieser Tisch steht offen.')")
  await zw.ende({ wirkung: offen })
  p.pruefe('der Tisch steht offen, ohne »Tisch öffnen«', offen)
  p.pruefe('und sagt, wer erwartet wird', await p.seite("return document.body.innerText.includes('Erwartet: Gast')"))
  await p.bild('s12-erwartet')

  p.notiz('Ohne Deck vor dem Anfang abgewiesen — mit Grund')
  await gast.aktiviere()
  await gast.taste('4', 'befehl')
  await gast.warte('die Deckseite steht', () => gast.seite(`return !!${knopf('Partie anfangen')}`), 20000)
  await gast.klickeText('Partie anfangen')
  await gast.wartetext('Bei jemandem mitspielen', 20000)
  await gast.klickeText('Bei jemandem mitspielen')
  await gast.klickeAuf("[...document.querySelectorAll('label, [role=radio]')].find((e) => e.checkVisibility() && /^Wirt · /.test((e.innerText ?? '').trim()))", 'den Wirt')
  await gast.seite("const s = [...document.querySelectorAll('select')].find((e) => e.checkVisibility() && e.labels?.[0]?.innerText.trim() === 'Mit welchem Deck'); Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(s, ''); s.dispatchEvent(new Event('change', { bubbles: true })); return true")
  await gast.klickeText('Mitspielen')
  await gast.warte('der Grund des Wirts steht da', async () => (await gast.seite("return document.body.innerText")).includes('wartet auf dein Deck'), 20000).catch(() => false)
  const grund = await gast.seite("return document.body.innerText.includes('wartet auf dein Deck')")
  gast.pruefe('ohne Deck: der Grund des Wirts, kein Tisch', grund, JSON.stringify(await meldungen(gast)))
  // **Ohne „Nachsehen“** — Go meldet den verlorenen Platz, und der Satz des
  // Wirts steht unter einem eigenen statt der Zeile über einen Abriss.
  const absage = await gast.seite("return document.body.innerText.includes('Der Wirt hat dich nicht an seinen Tisch gesetzt.') && !document.body.innerText.includes('Die Verbindung zum Wirt ist abgerissen.')")
  gast.pruefe('und er steht als Absage, nicht als Abriss', absage)
  // **Und ohne den Tisch, mit dem der Wirt grüßte** — an ihm saß niemand.
  const tischDa = await gast.seite("return [...document.querySelectorAll('h2')].some((e) => e.checkVisibility() && e.innerText.trim() === 'Am Tisch')")
  gast.pruefe('und kein Tisch steht mehr da', !tischDa)
  await gast.bild('ohne-deck-abgewiesen')
  await gast.taste('Escape')
  await bindung(gast, 'b.PartieBeitrittBeenden()')

  p.notiz('S13 — der Gast setzt sich mit seinem Deck: »Bei Wirt mitspielen«')
  await rufbuchHolen(gast)
  const vorher = { ...gast.rufbuch }
  const vorherRuf = (name) => rufe(vorher, name)
  await zg.beginne('S13', 'bereich', 'maus')
  await zg.taste('4', 'befehl')
  await gast.warte('»Bei Wirt mitspielen« steht', () => gast.seite(`return !!${knopf('Bei Wirt mitspielen')}`), 20000)
  await zg.text('Bei Wirt mitspielen')
  await gast.wartetext('Du sitzt am Tisch von', 20000).catch(() => false)
  await gast.warte('die Leinwand des Gastes zeichnet', () => gast.seite(leinwand), 20000).catch(() => false)
  await zg.ende({ wirkung: await gast.seite("return document.body.innerText.includes('Du sitzt am Tisch von')") })

  p.notiz('S13 — die Eröffnung des Gastes')
  await zg.beginne('S13', 'eroeffnung', 'maus')
  const steht = (text) => gast.seite(`return !!${knopf(text)}`)
  if (await gast.warte('»Starthand ziehen« steht', () => steht('Starthand ziehen'), 15000).then(() => true, () => false)) {
    await zg.text('Starthand ziehen')
    await gast.warte('»Hand behalten« steht', () => steht('Hand behalten'), 15000)
    await zg.text('Hand behalten')
    await gast.warte('behalten', async () => !(await steht('Hand behalten')), 15000).catch(() => false)
    await zg.ende({ wirkung: !(await steht('Hand behalten')) && !(await steht('Starthand ziehen')) })
  } else {
    zg.ohneGang('S13', 'eroeffnung', 'maus', 'FEHLT', 'am Tisch des Gastes steht keine Eröffnung zur Wahl')
  }
  await gast.bild('s13-gast-mit-eigenem-deck')

  p.notiz('Das Gastdeck: beim Gast die eigenen Zeilen, beim Wirt keine')
  const partie = (await bindung(gast, 'b.PartieBeitrittStand()'))?.partie ?? ''
  const beimWirt = await bindung(p, `b.Partie(${JSON.stringify(partie)}, 'sitz2')`)
  const hand = (beimWirt?.partie?.objekte ?? []).filter((o) => o.eigner === 'sitz2' && o.zone === 'hand')
  p.pruefe('Bos Hand liegt beim Wirt ohne Zeile, mit Bildkennung', hand.length === 7 && hand.every((o) => o.karteID === '' && o.bild !== ''), JSON.stringify(hand.map((o) => [o.karteID, o.bild !== ''])))
  // **Keine Sicht trägt eine innere Kennung** (#2480, Vermerk 0201): weder
  // Sitz noch Deckplatz noch Exemplar, nur eine Sichtkennung.
  const kennungen = (beimWirt?.partie?.objekte ?? []).map((o) => o.kennung)
  p.pruefe('jede Kennung in der Sicht des Gastes ist eine Sichtkennung', kennungen.length > 0 && kennungen.every((k) => /^s-[A-Za-z0-9_-]{18}$/.test(k)), JSON.stringify(kennungen.slice(0, 4)))
  await rufbuchHolen(gast)
  const eigen = rufe(gast.rufbuch, 'Spieltextur').gelungen - vorherRuf('Spieltextur').gelungen
  gast.pruefe('der Gast zeichnet seine Hand selbst', eigen >= 1, String(eigen))

  p.notiz('Der Wirt behält seine Hand — erst danach ist Spiel (CR 103.8, Fassung 2026-09-25)')
  await p.aktiviere()
  const beimWirtSteht = (text) => p.seite(`return !!${knopf(text)}`)
  if (await p.warte('»Hand behalten« steht beim Wirt', () => beimWirtSteht('Hand behalten'), 15000).then(() => true, () => false)) {
    await p.klickeText('Hand behalten')
    await p.warte('der Wirt hat behalten', async () => !(await beimWirtSteht('Hand behalten')), 15000).catch(() => false)
  }
  p.pruefe('beim Wirt steht keine Eröffnung mehr zur Wahl', !(await beimWirtSteht('Hand behalten')) && !(await beimWirtSteht('Starthand ziehen')))

  p.notiz('Kartenzug: eine Karte des Gastes auf den Tisch — der Wirt sieht sie mit dem Bild des Gastes')
  const gelegt = hand[0]
  const gezogen = await bindung(gast, `b.PartieWuenschen('zug-1', 'objekt.schieben', { objekt: ${JSON.stringify(gelegt?.kennung ?? '')}, zone: 'tisch' })`)
  gast.pruefe('der Wunsch geht hinaus', !gezogen?.fehler, JSON.stringify(gezogen?.fehler ?? ''))
  await p.warte('beim Wirt liegt die Karte auf dem Tisch', async () => {
    const jetzt = await bindung(p, `b.Partie(${JSON.stringify(partie)}, 'sitz1')`)

    return (jetzt?.partie?.objekte ?? []).some((o) => o.kennung === gelegt?.kennung && o.zone === 'tisch' && o.bild !== '')
  }, 15000)
  const offenBeimWirt = (await bindung(p, `b.Partie(${JSON.stringify(partie)}, 'sitz1')`))?.partie?.objekte?.find((o) => o.kennung === gelegt?.kennung)
  const vomGast = await bindung(p, `b.Fremdkartenbild(${JSON.stringify(partie)}, ${JSON.stringify(gelegt?.kennung ?? '')}, ${JSON.stringify(offenBeimWirt?.bild ?? '')}, 'vorne', 'tisch')`)
  // **Das Deck des Gastes hat zwei Zeilen** — eine davon ist die Karte; der
  // Gast zeichnet beide selbst, und eines der Bilder ist das des Wirts.
  const eigene = []
  for (const zeile of ['1', '2']) {
    const selbst = await bindung(gast, `b.Spieltextur('${zeile}', 'vorne', 'tisch')`)
    eigene.push(selbst?.textur?.adresse ? await abdruck(gast, selbst.textur.adresse) : 'kein Bild')
  }
  const a = await abdruck(p, vomGast?.textur?.adresse ?? '')
  p.zeige(`Karte des Gastes: beim Wirt vom Gast ${a}, beim Gast gezeichnet ${eigene.join(' / ')}`)
  p.pruefe('der Wirt zeigt byteweise ein Bild, das der Gast zeichnet', !a.startsWith('HTTP') && eigene.includes(a), `${a} · ${eigene.join(' / ')}`)
  await p.aktiviere()
  await p.bild('wirt-sieht-gastkarte')

  p.notiz('Ein abgewiesener Wunsch beendet die Sitzung nicht (#2478)')
  await gast.aktiviere()
  const nein = await bindung(gast, "b.PartieWuenschen('zug-nein', 'objekt.tappen', { objekt: 'gibt-es-nicht' })")
  gast.pruefe('der Wunsch geht hinaus', !nein?.fehler, JSON.stringify(nein?.fehler ?? ''))
  const abgewiesenSatz = 'Der Wirt hat deinen Wunsch abgewiesen'
  await gast.warte('der Grund steht am Tisch', async () => (await gast.seite('return document.body.innerText')).includes(abgewiesenSatz), 15000).catch(() => false)
  gast.pruefe('der Grund des Kerns steht am Tisch', await gast.seite(`return document.body.innerText.includes(${JSON.stringify(abgewiesenSatz)})`))
  gast.pruefe('und der Gast sitzt weiter', (await bindung(gast, 'b.PartieBeitrittStand()'))?.laeuft === true)
  await gast.bild('wunsch-abgewiesen')

  p.notiz('Abbruch und Wiederverbindung — derselbe Sitz, dieselben Karten')
  await gast.aktiviere()
  await gast.klickeText('Vom Tisch aufstehen')
  await gast.warte('der Beitritt läuft nicht mehr', async () => (await bindung(gast, 'b.PartieBeitrittStand()')).laeuft === false, 15000)
  await gast.taste('4', 'befehl')
  await gast.warte('»Bei Wirt mitspielen« steht wieder', () => gast.seite(`return !!${knopf('Bei Wirt mitspielen')}`), 20000)
  await gast.klickeText('Bei Wirt mitspielen')
  await gast.wartetext('Du sitzt am Tisch von', 20000)
  const danach = await bindung(p, `b.Partie(${JSON.stringify(partie)}, 'sitz2')`)
  const wieder = (danach?.partie?.objekte ?? []).filter((o) => o.eigner === 'sitz2' && o.zone !== 'bibliothek').map((o) => o.kennung).sort()
  const vorhin = hand.map((o) => o.kennung).sort()
  gast.pruefe('nach dem Wiederverbinden dieselben Karten am selben Sitz', JSON.stringify(wieder) === JSON.stringify(vorhin), `${wieder.length} von ${vorhin.length}`)
  gast.pruefe('und keine Fehlermeldung am Tisch', (await meldungen(gast)).length === 0, JSON.stringify(await meldungen(gast)))
  await gast.bild('wiederverbunden')

  writeFileSync(ziel, `${JSON.stringify(datensatz, null, 2)}\n`)
  p.zeige(`Datensatz: ${ziel}`)
  await bindung(gast, 'b.PartieBeitrittBeenden()')
  gast.pruefe('⌘Q beendet den Gast', await gast.beende())
  p.pruefe('⌘Q beendet den Wirt', await p.beende())
}
