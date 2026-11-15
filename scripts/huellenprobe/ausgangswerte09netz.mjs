// **Ausgangswerte der Netzpartie am nativen Fenster** — P0
// (#2336),
// Befund C-22 aus dem Audit zur Produktionsreife 0.9.0.
//
// S12 und S13 aus Audit C (Abschnitt 3.1): zwei Anwendungen, je mit eigenem
// Bestand, Bund und eigener Identität, auf diesem Mac — wie in `nat04`.
// Gezählt nach `huellenwege.mjs`, abgelegt als
// `docs/belege/ausgangswerte09-netz.json` (Vertrag: `scripts/ausgangswege.mjs`).
//
// # Was gezählt wird
//
// **S12, als Wirt**: die Partie im Fenster anlegen wie S1 und den Tisch
// öffnen. **S13, als Gast**: über »Bei jemandem mitspielen …« an den Tisch,
// und daneben die Eröffnung des Gastes als eigene Zeile. **Ungezählt** ist,
// was einmal je Gegenüber geschieht — Identität, Kontakt, Zweck, Horchposten
// —; es hat seine Belege in `net01` und `id02`.
//
// # Was es nicht belegt
//
// **Zwei Geräte in zwei Netzen** (#1998): beide Enden laufen auf diesem Mac
// über die Schleife. Und Menschenzeit, Fehlgriffe, Lesen und Suchen.
import { writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { bindung } from '../huellenhandel.mjs'
import { anfangen, deckwahl, knopf, namensfeld } from '../huellenspieltisch.mjs'
import { wegzaehler } from '../huellenwege.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const ziel = join(wurzel, 'docs/belege/ausgangswerte09-netz.json')
const leinwand = "return !!document.querySelector('canvas')?.checkVisibility()"

// koppeln macht zwei Anwendungen einander bekannt — bestätigt, mit Zweck.
async function koppeln(hier, dort, name) {
  await bindung(hier, `b.PeerkontaktAufnehmen(${JSON.stringify(dort.kennung)}, ${JSON.stringify(name)})`)
  const bestaetigt = await bindung(hier, `b.PeerkontaktBestaetigen(${JSON.stringify(dort.kennung)}, ${JSON.stringify(dort.fingerabdruck)})`)
  hier.pruefe(`${name} ist bestätigt`, !bestaetigt?.fehler, JSON.stringify(bestaetigt?.fehler ?? ''))
  const erlaubt = await bindung(hier, `b.PeerzweckErlauben(${JSON.stringify(dort.kennung)}, 'partie')`)
  hier.pruefe(`${name} darf eine Partie spielen`, !erlaubt?.fehler, JSON.stringify(erlaubt?.fehler ?? ''))
}

export default async function ausgangswerte09netz(p) {
  const gast = new p.constructor('ausgangswerte09netz-gast')
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
  p.mse(wirtOrt, 'mtgjson-lesen', join(wurzel, 'testdaten/mtgjson/tisch/TSP.json'), 'magic')
  p.mse(wirtOrt, 'deck', 'neu', 'Waldschritt', 'magic')
  for (const [k, n] of [['1', 2], ['3', 1], ['4', 1], ['5', 1], ['6', 2]]) p.mse(wirtOrt, 'deck', 'legen', '1', `--karte=${k}`, String(n))
  p.mse(wirtOrt, 'deck', 'neu', 'Glutpfad', 'magic')
  for (const [k, n] of [['2', 16], ['4', 4], ['6', 4]]) p.mse(wirtOrt, 'deck', 'legen', '2', `--karte=${k}`, String(n))

  const datensatz = { bauabdruck: p.lauf?.bauabdruck ?? '', erhoben: new Date().toISOString(), fenster: '1440 × 900', wege: [] }
  const zw = wegzaehler(p, datensatz)
  const zg = wegzaehler(gast, datensatz)

  p.notiz('Zwei Anwendungen, einander bekannt — ungezählt')
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

  p.notiz('S12 — über das Netz spielen, als Wirt')
  await p.aktiviere()
  await zw.beginne('S12', 'bereich', 'maus')
  await zw.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await zw.text('Partie anfangen')
  await p.wartetext('Wer sitzt am Tisch', 20000)
  for (const [n, name, deck] of [[1, 'Ada', '1'], [2, 'Bo', '2']]) {
    await zw.klickAuf(namensfeld(n), `das Feld »Name am Tisch ${String(n)}«`)
    await p.seite(`(${namensfeld(n)}).focus(); return true`)
    await zw.tippe(name)
    await zw.wahl(deckwahl(n - 1, deck))
  }
  await zw.klickAuf(anfangen, '»Partie anfangen« unter der Anlage')
  await p.warte('die Leinwand des Wirts zeichnet', () => p.seite(leinwand), 30000)
  await p.warte('»Tisch öffnen« steht', () => p.seite(`return !!${knopf('Tisch öffnen')}`), 10000)
  await zw.text('Tisch öffnen')
  await p.wartetext('Dieser Tisch steht offen.', 10000).catch(() => false)
  await zw.ende({ wirkung: await p.seite("return document.body.innerText.includes('Dieser Tisch steht offen.')") })
  await p.bild('s12-tisch-offen')

  p.notiz('S13 — über das Netz spielen, als Gast')
  await gast.aktiviere()
  await zg.beginne('S13', 'bereich', 'maus')
  await zg.taste('4', 'befehl')
  await gast.wartetext('Partie anfangen', 20000)
  await zg.text('Partie anfangen')
  await gast.wartetext('Bei jemandem mitspielen', 20000)
  await zg.text('Bei jemandem mitspielen')
  await zg.klickAuf("[...document.querySelectorAll('label, [role=radio]')].find((e) => e.checkVisibility() && /^Wirt · /.test((e.innerText ?? '').trim()))", 'den Wirt')
  await zg.text('Mitspielen')
  await gast.wartetext('Du sitzt am Tisch von', 20000).catch(() => false)
  await gast.warte('die Leinwand des Gastes zeichnet', () => gast.seite(leinwand), 20000).catch(() => false)
  await zg.ende({ wirkung: await gast.seite("return document.body.innerText.includes('Du sitzt am Tisch von')") })
  await gast.bild('s13-gast-sitzt')

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
  await gast.bild('s13-gast-eroeffnet')

  writeFileSync(ziel, `${JSON.stringify(datensatz, null, 2)}\n`)
  p.zeige(`Datensatz: ${ziel}`)
  gast.pruefe('⌘Q beendet den Gast', await gast.beende())
  p.pruefe('⌘Q beendet den Wirt', await p.beende())
}
