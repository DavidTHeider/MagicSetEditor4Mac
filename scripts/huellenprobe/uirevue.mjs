// **Das Prüfmaß der Oberfläche, gemessen am nativen Fenster** — UXR-01
// (#1909).
//
// Je Seite, die die Palette anbietet, und je Fenstergröße zählt es im
// Inhaltsbereich (`main`): Bedienelemente, die davon im Bild stehen,
// Hauptknöpfe (gesamt und höchstens je Region), Formularfelder, Überschriften
// und die Rolltiefe (Höhe des Inhalts durch Höhe des Fensters). Dazu, ob
// waagerecht gerollt werden muss. **Die Zahlen sind die Ausgangslage jedes
// UXR-Issues** und stehen in `docs/UIREVIEW.md`.
//
// **Was das nicht misst**: ob eine Handlung die richtige ist, ob ein Feld
// verständlich beschriftet ist, wie sich die Seite bedient — das sieht ein
// Mensch (UXR-01.AK-04). Und es zählt, was **ohne weiteres Klicken** dasteht:
// ein aufgeklappter Bereich zählt nicht mit, ein eingeklappter nur mit seinem
// Knopf.

import { readFileSync, writeFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { hauptgewichtBefunde, verloreneHandlungen } from '../uirevuezustaende.mjs'

// **Der Satz für den Vergleich** (AP-3): `MSE_UIREVUE_ZIEL` legt ihn an
// einem anderen Ort ab, `MSE_UIREVUE_VORHER` nennt einen früheren.
const wurzel = resolve(import.meta.dirname, '..', '..')
const ziel = resolve(wurzel, process.env.MSE_UIREVUE_ZIEL ?? 'docs/belege/uirevue-seiten.json')

const groessen = [
  { name: 'normal', breite: 1280, hoehe: 800 },
  { name: 'eng', breite: 1000, hoehe: 700 },
]

// **Der Roller selbst trägt Namen und Fokus** — N3 des Nachlaufs über 6.13
// (#1906): bis dahin genügte *irgendein* fokussierbarer Nachfahre und als
// Name die Rolle oder gar der Elementname. Ein unbeschrifteter Tabellenroller
// mit einem Knopf in der ersten Spalte bestand, und sein rechter Teil blieb
// der Maus. **Ob er sich mit der Tastatur bewegt**, prüft [rollerBewegen] am
// laufenden Fenster.
const rollerteile = `
  const sichtbarR = (e) => e.checkVisibility({ checkOpacity: true, checkVisibilityCSS: true })
  const innenRoller = (wurzel) => [...wurzel.querySelectorAll('*')]
    .filter((e) => e.scrollWidth > e.clientWidth + 1 && /(auto|scroll)/.test(getComputedStyle(e).overflowX) && sichtbarR(e))
  const rollerangaben = (e) => {
    const verweis = e.getAttribute('aria-labelledby')
    const name = (e.getAttribute('aria-label') ?? (verweis ? document.getElementById(verweis)?.innerText : '') ?? '').trim()
    return { name: name || e.tagName.toLowerCase(), benannt: name !== '', fokussierbar: e.tabIndex >= 0 }
  }
`

export const messung = `${rollerteile}
  const sichtbar = (e) => e.checkVisibility({ checkOpacity: true, checkVisibilityCSS: true })
  // **Eine offene Schublade ist die Fläche**, nicht der Arbeitsplatz dahinter.
  const schublade = [...document.querySelectorAll('[role=dialog]')].find(sichtbar)
  const haupt = schublade ?? document.querySelector('main')
  if (!haupt) return null
  const imBild = (e) => { const r = e.getBoundingClientRect(); return r.width > 0 && r.height > 0 && r.bottom > 0 && r.top < innerHeight && r.right > 0 && r.left < innerWidth }
  // **Die Feldflächen über einer Karte zählen nicht** (UXR-06, #1914): sie
  // sind der Weg der Maus in ein Feld, nicht in der Tabfolge und ohne eigene
  // Gestalt — eine reiche Vorlage hätte R2 sonst allein durch ihre Felder
  // gerissen. Erkannt als Knopf ohne Tabstelle hinter einem Bild.
  const feldflaeche = (e) => e.matches('img ~ button[tabindex="-1"], canvas ~ button[tabindex="-1"]')
  const bedien = [...haupt.querySelectorAll('button, [role=button], a[href], [role=tab], [role=menuitem], summary')].filter(sichtbar).filter((e) => !feldflaeche(e))
  // **Eine Segmentreihe zählt einmal** (UXR-F2, #2008): ihre Optionsfelder sind
  // die Werte einer Frage, und die Frage ist es, die R3 zählt. Optionsfelder
  // außerhalb einer Reihe — eines je Stil, eines je Karte — zählen weiter
  // einzeln, deshalb zählt die Gruppe (role=radiogroup) und nicht das
  // Optionsfeld überhaupt.
  const felder = [...haupt.querySelectorAll('input:not([type=hidden]), select, textarea, [role=combobox], [role=switch], [role=radiogroup]')].filter((e) => !(e.matches('input') && e.closest('[role=radiogroup]'))).filter(sichtbar)
  const hauptknoepfe = bedien.filter((b) => /(^|\\s)bg-marke(\\s|$)/.test(b.className))
  // **Der Name, unter dem eine Handlung dasteht** (AP-3, #2360): so, wie ein
  // Vorleseprogramm ihn nennt — die Beschriftung, sonst der Text.
  const nameVon = (e) => (e.getAttribute('aria-label') ?? e.innerText ?? '').trim().replace(/\\s+/g, ' ')
  // **Gewählt** ist, was sich als gedrückt, ausgewählt oder angekreuzt meldet.
  const gewaehlt = (e) => ['aria-pressed', 'aria-selected', 'aria-checked'].some((a) => e.getAttribute(a) === 'true')
  const regionen = [...haupt.querySelectorAll('section, form, aside, [role=region], [role=tabpanel], [role=dialog]')]
  const regionVon = (b) => b.closest('section, form, aside, [role=region], [role=tabpanel], [role=dialog]')
  const jeRegion = regionen.map((r) => hauptknoepfe.filter((b) => regionVon(b) === r).length)
  // **Ein Hauptknopf ohne Region zählt zur Seite** (Z1 des Zweitblicks über
  // Phase 6.13): sobald es irgendeine Region gab, fielen Knöpfe im Seitenkopf
  // aus der Zählung — zwei führende Knöpfe dort blieben grün.
  jeRegion.push(hauptknoepfe.filter((b) => !regionen.includes(regionVon(b))).length)
  // **Was innen waagerecht rollt** (Z2): ein Verlauf, eine Tabelle. R5 gilt
  // der Seite; ein Bereich, der es tut, ist benannt und mit der Tastatur
  // erreichbar (WCAG 2.1.1) — sonst bleibt sein rechter Teil der Maus.
  const innen = innenRoller(haupt).map(rollerangaben)
  const roller = [document.scrollingElement, haupt, ...(() => { const a = []; let e = haupt.parentElement; while (e) { a.push(e); e = e.parentElement } return a })()]
    .filter(Boolean).filter((e) => e.scrollHeight > e.clientHeight + 1 && (e === document.scrollingElement || /(auto|scroll)/.test(getComputedStyle(e).overflowY)))
  const rolltiefe = roller.length === 0 ? 1 : Math.max(...roller.map((e) => e.scrollHeight / e.clientHeight))
  return {
    titel: (haupt.querySelector('h1, h2') ?? document.querySelector('h1'))?.innerText.trim() ?? '',
    schublade: !!schublade,
    bedien: bedien.length,
    bedienImBild: bedien.filter(imBild).length,
    haupt: hauptknoepfe.length,
    // **Je Blick** (AP-13, #2339): was an gefüllten Hauptknöpfen im Bild
    // steht — drei in drei Regionen hält R1 je Region und konkurriert doch.
    hauptImBild: hauptknoepfe.filter(imBild).length,
    hauptJeRegion: Math.max(0, ...jeRegion),
    // **Welche** gefüllt im Bild stehen (AP-3): eine Zahl sagt nicht, welcher
    // Knopf das Gewicht abgeben muss.
    hauptNamen: hauptknoepfe.filter(imBild).map(nameVon),
    // **Die Verwechslung aus B-08**: ein gewählter Umschalter, der gefüllt ist,
    // sieht aus wie die führende Handlung.
    gewaehltGefuellt: bedien.filter(gewaehlt).filter((b) => hauptknoepfe.includes(b)).map(nameVon),
    // **Jede Handlung der Fläche** mit Rolle und Namen, ob im Bild oder nicht —
    // der Vergleich vorher und nachher zählt, was verloren ging.
    handlungen: [...new Set(bedien.map((e) => (e.getAttribute('role') ?? e.tagName.toLowerCase()) + ': ' + nameVon(e)))].sort(),
    felder: felder.length,
    felderImBild: felder.filter(imBild).length,
    ueberschriften: [...haupt.querySelectorAll('h1, h2, h3, h4')].filter(sichtbar).length,
    rolltiefe: Math.round(rolltiefe * 10) / 10,
    waagerecht: document.scrollingElement.scrollWidth > innerWidth + 1,
    innenWaagerecht: innen,
  }
`

// flaecheImSkript ist die Fläche, wie `messung` sie wählt — die offene
// Schublade oder `main`.
const flaecheImSkript = "([...document.querySelectorAll('[role=dialog]')].find(sichtbarR) ?? document.querySelector('main'))"

// rollerBewegen fokussiert jeden inneren Roller der Fläche und drückt die
// Pfeiltaste nach rechts — **am nativen Fenster**, wie ein Mensch (N3). Es
// gibt je Roller, ob er sich bewegt hat.
async function rollerBewegen(p) {
  const anzahl = await p.seite(`${rollerteile} const f = ${flaecheImSkript}; return f ? innenRoller(f).length : 0`)
  const bewegt = []
  for (let stelle = 0; stelle < anzahl; stelle++) {
    const bereit = await p.seite(`${rollerteile} const e = innenRoller(${flaecheImSkript})[${String(stelle)}]; if (!e || e.tabIndex < 0) return false; e.scrollLeft = 0; e.focus(); return document.activeElement === e`)
    let ja = false
    if (bereit) {
      await p.taste('ArrowRight')
      for (let versuch = 0; versuch < 8 && !ja; versuch++) {
        await new Promise((los) => setTimeout(los, 150))
        ja = await p.seite('return (document.activeElement?.scrollLeft ?? 0) > 0')
      }
    }
    bewegt.push(ja)
  }
  return bewegt
}

// innenPruefen misst die inneren Roller der offenen Fläche samt Bewegung.
async function innenPruefen(p, wert) {
  if (wert.innenWaagerecht.length === 0) return wert
  const bewegt = await rollerBewegen(p)
  return { ...wert, innenWaagerecht: wert.innenWaagerecht.map((b, stelle) => ({ ...b, bewegt: bewegt[stelle] === true })) }
}

// unbedienbar nennt die Roller, die eine der drei Zusagen brechen.
const unbedienbar = (b) => !b.benannt || !b.fokussierbar || !b.bewegt

// **Die Gegenprobe der Rollerprüfung, im selben Lauf** (N3): auf keiner Seite
// rollt mit dem Probeset etwas innen waagerecht, und eine Prüfung, die nie
// etwas findet, belegt nichts. Deshalb zwei überbreite Bereiche in `main` —
// einer benannt mit Tabstelle, einer ohne beides, aber mit einem Knopf
// darin, der Fall, den die alte Prüfung durchließ. Der erste muss bestehen,
// der zweite muss auffallen.
async function rollerGegenprobe(p) {
  await p.seite(`const main = document.querySelector('main')
    for (const [kennung, gut] of [['gegenprobe-gut', true], ['gegenprobe-schlecht', false]]) {
      const roller = document.createElement('div')
      roller.id = kennung
      roller.style.cssText = 'overflow-x:auto;width:240px'
      if (gut) { roller.setAttribute('role', 'region'); roller.setAttribute('aria-label', 'Gegenprobe'); roller.tabIndex = 0 }
      roller.innerHTML = '<button type="button">x</button><div style="width:2400px;height:12px"></div>'
      main.prepend(roller)
    }
    return true`)
  const wert = await innenPruefen(p, await p.seite(messung))
  await p.seite("document.getElementById('gegenprobe-gut')?.remove(); document.getElementById('gegenprobe-schlecht')?.remove(); return true")
  const gefunden = wert.innenWaagerecht
  const gut = gefunden.find((b) => b.name === 'Gegenprobe')
  const schlecht = gefunden.filter((b) => b !== gut)
  p.pruefe('Gegenprobe: ein benannter Roller mit Tabstelle rollt mit der Pfeiltaste', gut !== undefined && !unbedienbar(gut), JSON.stringify(gut))
  p.pruefe('Gegenprobe: ein Roller ohne Namen und Tabstelle fällt auf, auch mit einem Knopf darin', schlecht.length === 1 && unbedienbar(schlecht[0]), JSON.stringify(schlecht))
}

// **Die Beschriftungen der Palette je Sprache** — `uxrabzug` fährt dieselben
// Wege auf Englisch, und die Palette heißt dort anders.
const palette = {
  de: { eingabe: 'Wohin?', seiten: 'Seiten' },
  en: { eingabe: 'Where to?', seiten: 'Pages' },
}

async function paletteOeffnen(p, sprache = 'de') {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite(`return !!document.querySelector('[role=dialog] input[aria-label="${palette[sprache].eingabe}"]')`))
}

export async function seitennamen(p, sprache = 'de') {
  await paletteOeffnen(p, sprache)
  // **Nur die Seitenliste**, nicht *Tun* und *Finden*: dort stehen Befehle,
  // und ein Messlauf soll nichts anlegen oder löschen.
  const namen = await p.seite(`return [...document.querySelectorAll('[role=dialog] ul[aria-label="${palette[sprache].seiten}"] [data-palettenzeile]')].filter((b) => b.checkVisibility()).map((b) => b.innerText.trim().split('\\n')[0])`)
  await p.taste('Escape')
  return namen
}

export async function zurSeite(p, name, sprache = 'de') {
  const { eingabe, seiten } = palette[sprache]
  await paletteOeffnen(p, sprache)
  // **Das Feld nimmt den Fokus, oder es wird gewartet** — ein Lauf fand am
  // 2026-10-02 die ersten vier Seiten nicht: die Palette stand für einen
  // Augenblick und war beim nächsten Ruf fort.
  await p.warte('die Palette nimmt Eingaben', () => p.seite(`const i = document.querySelector('[role=dialog] input[aria-label="${eingabe}"]'); if (!i) return false; i.focus(); return document.activeElement === i`))
  await p.tippeInsFeld(name)
  const zeile = `[...document.querySelectorAll('[role=dialog] ul[aria-label="${seiten}"] [data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim().split('\\n')[0] === ${JSON.stringify(name)})`
  const ziel = await p.warte(`»${name}« steht in der Palette`, () => p.seite(`return ${zeile}?.dataset.seite ?? ''`))
  await p.klickeAuf(`[...document.querySelectorAll('[role=dialog] ul[aria-label="${seiten}"] [data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim().split('\\n')[0] === ${JSON.stringify(name)})`, `»${name}« in der Palette`)
  // **Gewartet wird auf die Seite, nicht auf eine Frist** — N4 des
  // Nachlaufs über 6.13 (#1906): bis dahin standen hier 1,5 Sekunden, und bei
  // einem verzögerten oder wirkungslosen Wechsel lag unter dem neuen
  // Dateinamen das Bild der vorigen Fläche. Angekommen ist, wenn der Inhalt
  // als `data-seite` die Seite trägt, die die Palettenzeile nannte — auch
  // eine Seite, die als Schublade erscheint.
  await p.warte(`»${name}« steht da`, () => p.seite(angekommen(ziel)), 8000)
  await new Promise((los) => setTimeout(los, 300))
}

// angekommen prüft im Fenster, ob die Seite mit diesem Schlüssel steht und
// die Palette zu ist.
export function angekommen(schluessel) {
  return `if ([...document.querySelectorAll('[data-palettenzeile]')].some((e) => e.checkVisibility())) return false
    return document.querySelector('main')?.dataset.seite === ${JSON.stringify(schluessel)}`
}

// oeffnen startet die Anwendung in dieser Größe und öffnet das erste Set.
//
// **Über die gemerkte Größe und nicht über `window_control`**: ein
// `set_size` am maximierten Fenster ließ Wails' MCP-Server verstummen
// (erster Lauf am 2026-10-02). „Wie zuletzt" mit gesetzten Maßen ist der Weg,
// den ein Mensch auch nimmt.
export async function oeffnen(p, ort, groesse) {
  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', String(groesse.breite))
  p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', String(groesse.hoehe))
  await p.starte({ ort })
  p.zeige(`${groesse.name}: ${JSON.stringify((await p.fenster()).rahmen)}`)
  await p.taste('2', 'befehl')
  await p.warte('die Sets stehen', () => p.seite("return !!document.querySelector('main article')"), 20000)
  // **Ein Set ist offen**, sonst bietet die Palette die Seiten des Sets nicht an.
  await p.klickeAuf("[...document.querySelectorAll('main article :is(h2, h3) button, main article :is(h2, h3) a')].find((b) => b.checkVisibility())", 'das erste Set')
  await new Promise((los) => setTimeout(los, 2000))
}

export default async function uirevue(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  const zeilen = []
  let anzahl = 0
  for (const groesse of groessen) {
    await oeffnen(p, ort, groesse)
    if (groesse === groessen[0]) await rollerGegenprobe(p)
    const namen = await seitennamen(p)
    anzahl += namen.length
    p.zeige(`Die Palette nennt ${String(namen.length)} Seiten: ${namen.join(' · ')}`)
    for (const name of namen) {
      try {
        await zurSeite(p, name)
        const wert = await p.seite(messung)
        if (wert) zeilen.push({ groesse: groesse.name, ziel: name, erreicht: true, ...(await innenPruefen(p, wert)) })
      } catch (grund) {
        p.zeige(`${name} (${groesse.name}): nicht erreicht — ${String(grund.message ?? grund)}`)
        await p.taste('Escape')
      }
    }
    p.pruefe(`⌘Q beendet die Anwendung (${groesse.name})`, await p.beende())
  }

  p.notiz('Die Tabelle')
  p.zeige('| Größe | Ziel | Bedienelemente (im Bild) | Hauptknöpfe (je Region höchstens) | Felder (im Bild) | Überschriften | Rolltiefe | waagerecht |')
  p.zeige('| --- | --- | ---: | ---: | ---: | ---: | ---: | --- |')
  for (const z of zeilen) {
    p.zeige(`| ${z.groesse} | ${z.ziel}${z.schublade ? ' (Schublade)' : ''} | ${String(z.bedien)} (${String(z.bedienImBild)}) | ${String(z.haupt)} (${String(z.hauptJeRegion)}) | ${String(z.felder)} (${String(z.felderImBild)}) | ${String(z.ueberschriften)} | ${String(z.rolltiefe)} | ${z.waagerecht ? 'ja' : 'nein'} |`)
  }
  p.pruefe('jede Seite gemessen', zeilen.length === anzahl, `${String(zeilen.length)} von ${String(anzahl)}`)
  // **R1 bis R6 aus `docs/UIREVIEW.md`** — seit UXR-24 (#1932) hält jede
  // Seite alle sechs; bis dahin standen nur R1 und R5 hier, und die übrigen
  // waren Befunde einzelner Issues. R2, R3, R4 und R6 gelten bei 1280 × 800
  // (so sind sie gesetzt); R7 ist eine Frage an jedes Bedienelement und steht
  // in `wahlkatalog.test.tsx`.
  const zweiHaupt = zeilen.filter((z) => z.hauptJeRegion > 1).map((z) => `${z.ziel} (${z.groesse})`)
  p.pruefe('R1: höchstens ein Hauptknopf je Region', zweiHaupt.length === 0, zweiHaupt.join(', '))
  // **Und je Blick** — AP-3 (#2360): drei Knöpfe in drei Regionen halten R1
  // je Region und konkurrieren doch; ein gedrückter Umschalter, der gefüllt
  // ist, zählt mit, weil er gleich aussieht (B-08).
  p.notiz('Gefüllt im Bild, mit Namen')
  for (const z of zeilen) p.zeige(`${z.groesse} · ${z.ziel}: ${z.hauptNamen.join(' · ') || 'keiner'}`)
  const gewicht = hauptgewichtBefunde({ zeilen })
  p.pruefe('R1 je Blick: höchstens ein gefüllter Hauptknopf im Bild, kein gewählter Umschalter gefüllt', gewicht.length === 0, gewicht.join('; '))
  const satz = { bauabdruck: p.lauf?.bauabdruck ?? '', erhoben: new Date().toISOString(), zeilen }
  writeFileSync(ziel, `${JSON.stringify(satz, null, 2)}\n`)
  p.zeige(`Satz: ${ziel}`)
  if (process.env.MSE_UIREVUE_VORHER !== undefined) {
    const vorher = JSON.parse(readFileSync(resolve(wurzel, process.env.MSE_UIREVUE_VORHER), 'utf8'))
    p.zeige(`Vorher: Bauabdruck ${String(vorher.bauabdruck)}, erhoben ${String(vorher.erhoben)}`)
    const verloren = verloreneHandlungen(vorher, satz)
    p.pruefe('gegen den Vorherlauf ist keine Handlung verloren (AP-3)', verloren.length === 0, verloren.join('; '))
  }
  const normal = zeilen.filter((z) => z.groesse === 'normal')
  const ueber = (bedingung) => normal.filter(bedingung).map((z) => z.ziel).join(', ')
  p.pruefe('R2: höchstens 20 Bedienelemente im Bild', ueber((z) => z.bedienImBild > 20) === '', ueber((z) => z.bedienImBild > 20))
  p.pruefe('R3: höchstens 12 Felder gezeichnet', ueber((z) => z.felder > 12) === '', ueber((z) => z.felder > 12))
  p.pruefe('R4: Rolltiefe höchstens 2,5', ueber((z) => z.rolltiefe > 2.5) === '', ueber((z) => z.rolltiefe > 2.5))
  p.pruefe('R6: höchstens 200 gezeichnete Bedienelemente', ueber((z) => z.bedien > 200) === '', ueber((z) => z.bedien > 200))
  const waagerecht = zeilen.filter((z) => z.waagerecht).map((z) => `${z.ziel} (${z.groesse})`)
  p.pruefe('R5: keine Seite rollt waagerecht', waagerecht.length === 0, waagerecht.join(', '))
  const innen = zeilen.flatMap((z) => (z.innenWaagerecht ?? []).map((b) => ({ ...b, wo: `${z.ziel} (${z.groesse})` })))
  p.zeige(`innen waagerecht: ${innen.map((b) => `${b.wo}: ${b.name}`).join(', ') || 'keiner'}`)
  const mangel = innen.filter(unbedienbar).map((b) => `${b.wo}: ${b.name}`)
  p.pruefe('R5: was innen waagerecht rollt, ist benannt, nimmt den Fokus und rollt mit der Pfeiltaste', mangel.length === 0, mangel.join(', '))
}
