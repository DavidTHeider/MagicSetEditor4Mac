// **Das Dichtemaß in befüllten Zuständen, am nativen Fenster** — AP-13
// (#2339),
// Befund B-33 aus dem Audit zur Produktionsreife 0.9.0.
//
// Dieselbe `messung` wie `uirevue` — Bedienelemente, im Bild, Hauptknöpfe
// je Region, Felder, Überschriften, Rolltiefe —, aber in den Zuständen aus
// `scripts/uirevuezustaende.mjs`: ein großes Set mit gewählter Karte, ein
// offenes Deck, eine gefüllte Sammlung, und je ihr leerer Gegenpart. In
// zwei Größen und beiden Erscheinungsbildern; abgelegt als
// `docs/belege/uirevue09-zustaende.json`, die Bilder im Lauf.
//
// # Der Bestand
//
// **Künstlich**: 3.788 erfundene Karten (»Probekarte 0001« …) als MTG JSON
// erzeugt und eingelesen, 60 davon in einem Deck, 1.000 Exemplare über
// `mse sammlung-einlesen`. **Die Pakete** — Spiel und Stil M15 — kommen aus
// dem lokalen Pack (`MSE_PROBE_PACK`, Vorgabe `~/full-magic-pack`) als
// APFS-Klon in den Ort der Probe, wie in `budgets`. Fehlt der Pack, fehlen
// die Zustände, und der Lauf ist rot.
//
// # Was es nicht misst
//
// Ob eine Handlung die richtige ist und wie sich eine Seite bedient — das
// sieht ein Mensch (UXR-01.AK-04). **Die Regeln R1 bis R6 werden hier
// gezeigt, nicht gefordert**: dies ist das Ausgangsmaß, und B-33 sagt
// voraus, dass R2 im befüllten Zustand reißt.
//
// # Bis auf R1 je Blick
//
// **Seit AP-3** (#2360)
// ist R1 je Blick gefordert: höchstens ein gefüllter Hauptknopf im Bild, und
// kein gewählter Umschalter gefüllt (`hauptgewichtBefunde`). Mit
// `MSE_ZUSTAENDE_VORHER` — einem früheren Satz — prüft der Lauf zudem, dass
// keine Handlung verloren ging (`verloreneHandlungen`); `MSE_ZUSTAENDE_ZIEL`
// legt den Satz an einem anderen Ort ab, für den Vorherlauf.
import { execFileSync } from 'node:child_process'
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs'
import { homedir } from 'node:os'
import { join, resolve } from 'node:path'

import { kindumgebung } from '../huellenorte.mjs'
import { erscheinungen, groessen, hauptgewichtBefunde, pflichtzustaende, pruefeZustaende, verloreneHandlungen } from '../uirevuezustaende.mjs'
import { messung } from './uirevue.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const ziel = resolve(wurzel, process.env.MSE_ZUSTAENDE_ZIEL ?? 'docs/belege/uirevue09-zustaende.json')
const pack = process.env.MSE_PROBE_PACK ?? join(homedir(), 'full-magic-pack')
const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))

const KARTEN = 3788
const EXEMPLARE = 1000

// karten erzeugt ein Set erfundener Karten in der Gestalt von
// `testdaten/mtgjson/tisch/TSP.json`.
function karten(anzahl) {
  const arten = [
    { type: 'Creature — Probe', types: ['Creature'], subtypes: ['Probe'], power: '2', toughness: '2' },
    { type: 'Instant', types: ['Instant'], subtypes: [] },
    { type: 'Sorcery', types: ['Sorcery'], subtypes: [] },
    { type: 'Artifact', types: ['Artifact'], subtypes: [] },
    { type: 'Enchantment', types: ['Enchantment'], subtypes: [] },
  ]
  const kosten = ['{1}{W}', '{2}{U}', '{B}', '{3}{R}', '{G}{G}', '{4}']
  const seltenheit = ['common', 'uncommon', 'rare', 'mythic']
  const liste = Array.from({ length: anzahl }, (_, i) => {
    const art = arten[i % arten.length]
    const nummer = String(i + 1).padStart(4, '0')
    return {
      name: `Probekarte ${nummer}`, manaCost: kosten[i % kosten.length], supertypes: [], ...art,
      text: `Eine erfundene Fähigkeit, Nummer ${nummer}.`, rarity: seltenheit[i % seltenheit.length],
      artist: 'Selbst gezeichnet', number: String(i + 1), uuid: `00000000-0000-4000-9000-${(i + 1).toString(16).padStart(12, '0')}`,
    }
  })

  return { meta: { date: '2026-10-13', version: '5.2.2+20261013' }, data: { code: 'P09', name: 'Probesatz 0.9.0', releaseDate: '2026-10-13', type: 'expansion', cards: liste } }
}

// mitEingabe ruft `mse` mit einer Eingabe auf der Standardeingabe.
const mitEingabe = (ort, eingabe, ...argumente) => execFileSync(join(wurzel, '.go/bin/mse'), argumente, { cwd: wurzel, input: eingabe, env: kindumgebung(ort) }).toString()

// mitPack legt die Pakete des lokalen Packs in den Ort — ein Klon, keine Kopie.
function mitPack(p, ort) {
  const pakete = join(ort.heim, 'Library/Application Support/MagicSetEditor4Mac/pakete')
  mkdirSync(pakete, { recursive: true })
  execFileSync('cp', ['-cR', `${join(pack, 'data')}/.`, pakete])
  // **Und seine Schriften**, wie ein Mensch sie mit dem Pack installiert —
  // ohne sie füllt die Spalte »Probleme« den Inspector mit fehlenden
  // Schriften, und die Dichte gälte einem Zustand, den niemand hat.
  const schriften = join(ort.heim, 'Library/Fonts')
  mkdirSync(schriften, { recursive: true })
  if (existsSync(join(pack, 'Magic - Fonts'))) execFileSync('cp', ['-cR', `${join(pack, 'Magic - Fonts')}/.`, schriften])
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
}

// befuellt legt den Ort mit großem Set, Deck und Sammlung an.
function befuellt(p, ort) {
  mitPack(p, ort)
  const datei = join(ort.ort, 'probesatz.json')
  writeFileSync(datei, JSON.stringify(karten(KARTEN)))
  p.mse(ort, 'mtgjson-lesen', datei, 'magic')
  p.zeige(`Stilwechsel: ${p.mse(ort, 'stile', '1', 'wechseln', 'magic-m15').split('\n')[0]}`)
  p.zeige(`Karten im Bestand: ${String(p.mse(ort, 'karten', '1').trim().split('\n').length)}`)
  // **Die Gegenprobe des Vertrags** (AP-13): `MSE_AP13_OHNE_DECK=1` lässt das
  // Deck weg, und der Lauf muss rot werden.
  if (process.env.MSE_AP13_OHNE_DECK !== '1') {
    p.mse(ort, 'deck', 'neu', 'Probedeck', 'magic')
    for (let karte = 1; karte <= 20; karte += 1) p.mse(ort, 'deck', 'legen', '1', `--karte=${String(karte)}`, '3')
  }
  p.mse(ort, 'exemplar', '09-karte-0001', '1', 'nearMint', 'de', 'nonfoil')
  const [kopf, muster] = p.mse(ort, 'sammlung-ausgeben', 'csv').trim().split('\n')
  const spalte = kopf.split(',').findIndex((s) => /printing|druck/i.test(s))
  const zeilen = Array.from({ length: EXEMPLARE - 1 }, (_, i) => muster.split(',').map((w, j) => (j === spalte ? `09-karte-${String(i + 2).padStart(4, '0')}` : w)).join(','))
  mitEingabe(ort, [kopf, ...zeilen].join('\n') + '\n', 'sammlung-einlesen', '--auslassen')
  p.zeige(`Exemplare: ${String(p.mse(ort, 'sammlung-ausgeben', 'csv').trim().split('\n').length - 1)}`)
}

// leer legt einen Ort mit denselben Paketen und einem leeren Set an.
function leer(p, ort) {
  mitPack(p, ort)
  p.mse(ort, 'set-neu', 'Leerset', 'magic', 'magic-m15')
}

const seite = "return document.querySelector('main')?.dataset.seite ?? ''"
const ueberschrift = (text) => `return [...document.querySelectorAll('main h1')].some((e) => e.checkVisibility() && e.innerText.trim() === ${JSON.stringify(text)})`
const sichtbarText = (text) => `return [...document.querySelectorAll('main *')].some((e) => e.checkVisibility() && e.children.length === 0 && (e.innerText ?? '').includes(${JSON.stringify(text)}))`

// wege führt je Zustand an seinen Ort im Fenster und sagt, ob er erreicht ist.
const wege = {
  async 'karten-gross'(p) {
    await zumSet(p, 'Probesatz 0.9.0')
    // **Im engen Fenster ist der Navigator eingeklappt** und die erste Karte
    // schon aufgeschlagen; im weiten wird sie in der Liste gewählt.
    const zeile = "[...document.querySelectorAll('main [role=row], main [role=option], main li')].find((e) => e.checkVisibility() && (e.innerText ?? '').includes('Probekarte 0001'))"
    const vorschau = "return !!document.querySelector('img[alt=\"Vorschau der Karte\"]')?.checkVisibility()"
    await p.warte('die erste Karte oder ihre Vorschau steht', async () => (await p.seite(`return !!(${zeile})`)) || p.seite(vorschau), 20000)
    if (await p.seite(`return !!(${zeile})`)) await p.klickeAuf(zeile, '»Probekarte 0001«')
    return p.warte('die Vorschau von »Probesatz 0.9.0« steht', async () => (await p.seite(vorschau)) && p.seite(ueberschrift('Probesatz 0.9.0')), 20000).then(() => true, () => false)
  },
  async 'karten-leer'(p) {
    await zumSet(p, 'Leerset')
    // **Ein positiver Leerzustand** — das Set, und sein leerer Satz.
    return p.warte('»Leerset« ohne Karte', async () => (await p.seite(ueberschrift('Leerset'))) && p.seite(sichtbarText('Noch keine Karte')), 20000).then(() => true, () => false)
  },
  async 'deck-offen'(p) {
    await p.taste('4', 'befehl')
    const oeffnen = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '') === 'Öffnen: Probedeck')"
    await p.warte('das Deck steht zur Wahl', () => p.seite(`return !!(${oeffnen})`), 20000)
    await p.klickeAuf(oeffnen, '»Öffnen: Probedeck«')
    // **»Karten suchen« steht in beiden Größen**; »Mit diesem Deck spielen«
    // liegt im Inspector, und der ist im engen Fenster eingeklappt.
    // **Und die Kennzahlen stehen** — „Fehlt in der Sammlung“ kommt aus dem
    // Abgleich und später als der Kasten; ein Lauf maß einmal davor (AP-6,
    // #2363).
    return p.warte('das Deck ist offen', () => p.seite(sichtbarText('Karten suchen')), 20000)
      .then(() => p.warte('die Kennzahlen stehen', () => p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim().startsWith('Fehlt in der Sammlung'))"), 20000))
      .then(() => true, () => false)
  },
  async 'decks-leer'(p) {
    await p.taste('4', 'befehl')
    return p.warte('»Noch kein Deck« steht', () => p.seite(sichtbarText('Noch kein Deck')), 20000).then(() => true, () => false)
  },
  async 'sammlung-voll'(p) {
    await p.taste('3', 'befehl')
    // **Die Zählzeile und nicht der Name**: im engen Fenster bleibt die
    // Spalte »Bezeichnung« leer (B-15).
    return p.warte('Exemplare stehen', () => p.seite(sichtbarText(`${String(EXEMPLARE)} von ${String(EXEMPLARE)}`)), 20000).then(() => true, () => false)
  },
  async 'sammlung-leer'(p) {
    await p.taste('3', 'befehl')
    return p.warte('die leere Sammlung steht', () => p.seite(sichtbarText('Noch nichts eingetragen')), 20000).then(() => true, () => false)
  },
}

// zumSet öffnet ein Set über seinen Titel in der Kachel.
async function zumSet(p, name) {
  await p.taste('2', 'befehl')
  const titel = `[...document.querySelectorAll('main a, main button')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(name)})`
  await p.warte(`»${name}« steht`, () => p.seite(`return !!(${titel})`), 20000)
  await p.klickeAuf(titel, `»${name}«`)
  await schlafe(1500)
}

export default async function uirevue09zustaende(p) {
  const orte = { befuellt: p.ort(), leer: p.ort() }
  const satz = { bauabdruck: p.lauf?.bauabdruck ?? '', erhoben: new Date().toISOString(), karten: KARTEN, exemplare: EXEMPLARE, zeilen: [] }
  if (existsSync(join(pack, 'data'))) {
    befuellt(p, orte.befuellt)
    leer(p, orte.leer)
  } else {
    p.zeige(`kein Pack unter ${pack} — die Zustände fehlen`)
  }

  for (const [art, ort] of Object.entries(orte)) {
    const zustaende = pflichtzustaende.filter((z) => z.befuellt === (art === 'befuellt'))
    for (const groesse of groessen) {
      for (const erscheinung of erscheinungen) {
        p.notiz(`${art} · ${groesse.name} · ${erscheinung}`)
        p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
        p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', String(groesse.breite))
        p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', String(groesse.hoehe))
        p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', erscheinung)
        await p.starte({ ort })
        p.zeige(`Fenster: ${JSON.stringify((await p.fenster()).rahmen)}`)
        for (const zustand of zustaende) {
          let erreicht = false
          try {
            erreicht = await wege[zustand.name](p)
          } catch (grund) {
            p.zeige(`${zustand.name}: ${String(grund.message ?? grund)}`)
          }
          const wert = erreicht ? await p.seite(messung) : null
          satz.zeilen.push({ groesse: groesse.name, erscheinung, zustand: zustand.name, befuellt: zustand.befuellt, erreicht, seite: await p.seite(seite), ...(wert ?? {}) })
          p.pruefe(`${zustand.name} (${groesse.name}, ${erscheinung}) erreicht`, erreicht)
          await p.bild(`${zustand.name}-${groesse.name}-${erscheinung}`)
        }
        p.pruefe(`⌘Q beendet (${art}, ${groesse.name}, ${erscheinung})`, await p.beende())
      }
    }
  }

  p.notiz('Die Tafel')
  p.zeige('| Größe | Bild | Zustand | Bedienelemente (im Bild) | Hauptknöpfe (im Bild · je Region höchstens) | Felder (im Bild) | Überschriften | Rolltiefe | waagerecht |')
  p.zeige('| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | --- |')
  for (const z of satz.zeilen.filter((x) => x.erreicht)) {
    p.zeige(`| ${z.groesse} | ${z.erscheinung} | ${z.zustand} | ${String(z.bedien)} (${String(z.bedienImBild)}) | ${String(z.haupt)} (${String(z.hauptImBild)} · ${String(z.hauptJeRegion)}) | ${String(z.felder)} (${String(z.felderImBild)}) | ${String(z.ueberschriften)} | ${String(z.rolltiefe)} | ${z.waagerecht ? 'ja' : 'nein'} |`)
  }
  p.notiz('Gefüllt im Bild, mit Namen')
  for (const z of satz.zeilen.filter((x) => x.erreicht)) {
    p.zeige(`${z.groesse} · ${z.erscheinung} · ${z.zustand}: ${(z.hauptNamen ?? []).join(' · ') || 'keiner'}${(z.gewaehltGefuellt ?? []).length > 0 ? ` — gewählt und gefüllt: ${z.gewaehltGefuellt.join(' · ')}` : ''}`)
  }
  const gross = satz.zeilen.filter((z) => z.erreicht && z.groesse === groessen[0].name)
  p.zeige(`R1 (höchstens ein Hauptknopf je Region) gerissen: ${gross.filter((z) => z.hauptJeRegion > 1).map((z) => `${z.zustand}/${z.erscheinung}`).join(', ') || 'nirgends'}`)
  p.zeige(`R1 je Blick (mehr als ein Hauptknopf im Bild): ${gross.filter((z) => z.hauptImBild > 1).map((z) => `${z.zustand}/${z.erscheinung} (${String(z.hauptImBild)})`).join(', ') || 'nirgends'}`)
  p.zeige(`R2 (höchstens 20 im Bild) gerissen: ${gross.filter((z) => z.bedienImBild > 20).map((z) => `${z.zustand}/${z.erscheinung} (${String(z.bedienImBild)})`).join(', ') || 'nirgends'}`)
  p.zeige(`R3 (höchstens 12 Felder) gerissen: ${gross.filter((z) => z.felder > 12).map((z) => `${z.zustand}/${z.erscheinung} (${String(z.felder)})`).join(', ') || 'nirgends'}`)
  p.zeige(`R4 (Rolltiefe höchstens 2,5) gerissen: ${gross.filter((z) => z.rolltiefe > 2.5).map((z) => `${z.zustand}/${z.erscheinung} (${String(z.rolltiefe)})`).join(', ') || 'nirgends'}`)

  writeFileSync(ziel, `${JSON.stringify(satz, null, 2)}\n`)
  p.zeige(`Satz: ${ziel}`)
  const befunde = pruefeZustaende(satz)
  p.pruefe('jeder Pflichtzustand ist in jeder Größe und jedem Bild gemessen', befunde.length === 0, befunde.join('; '))
  const gewicht = hauptgewichtBefunde(satz)
  p.pruefe('je Blick höchstens ein gefüllter Hauptknopf, kein gewählter Umschalter gefüllt (AP-3)', gewicht.length === 0, gewicht.join('; '))
  if (process.env.MSE_ZUSTAENDE_VORHER !== undefined) {
    const vorher = JSON.parse(readFileSync(resolve(wurzel, process.env.MSE_ZUSTAENDE_VORHER), 'utf8'))
    p.zeige(`Vorher: Bauabdruck ${String(vorher.bauabdruck)}, erhoben ${String(vorher.erhoben)}`)
    const verloren = verloreneHandlungen(vorher, satz)
    p.pruefe('gegen den Vorherlauf ist keine Handlung verloren (AP-3)', verloren.length === 0, verloren.join('; '))
  }
}
