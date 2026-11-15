// **Der Markttisch am nativen Fenster** — P12
// (#2380),
// Befunde C-14, C-18 aus dem Audit zur Produktionsreife 0.9.0.
//
// # Was es prüft
//
// **Zwei getrennte Peers**: Anna im Fenster, Bert als eigener Prozess
// `mse horchen` (wie `p11sicht09`). Anna hat Glutläufer und Moosbär, Bert
// zwei Sturmrufer und gibt Anna seine Sammlung zum Handeln frei.
//
// - **H1** mit der Maus, gezählt nach `huellenwege.mjs`: ⌘5, „Neuer Tausch
//   mit Bert“, Glutläufer nach „Von dir“ ziehen, Sturmrufer nach „Vom
//   Gegenüber“ ziehen, „Anbieten und zustimmen“ — Schwelle 5 (P12) und 7 (P10).
// - **H2**: ein Stück aus dem eigenen Fach in den Tausch — ein Zug.
// - **H3**: was das Gegenüber freigibt **und was davon zusammenpasst**, steht
//   beim Aufschlagen da — null Handlungen (#2494). Anna sucht einen
//   Sturmrufer, Bert zeigt zwei.
// - **Mehrfachwahl** mit ⌘ und ein Zug; **Tastatur** (Leertaste, Eingabe) für
//   beide Ziele; ein falsches Ziel legt nichts; **„Als Liste“** führt zum
//   bisherigen Assistenten. Hell und dunkel.
//
// # Was es nicht belegt
//
// **Zwei Rechner** (#1998), **VoiceOver**, ein Finger auf einem Trackpad: das
// Ziehen kommt als Mausereignis ab AppKit (`ziehe`), nicht vom Fensterserver.
import { spawn } from 'node:child_process'
import { writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { kindumgebung } from '../huellenorte.mjs'
import { gekoppelt, handel, knopf } from '../huellenhandel.mjs'
import { ziehe } from '../huellenrufbuch.mjs'
import { mitteVon } from '../huellenspieltisch.mjs'
import { wegzaehler } from '../huellenwege.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const ziel = join(wurzel, 'docs/belege/p12markttisch09-wege.json')
const sturm = '00000000-0000-4000-8000-0000000b1002'
const glut = '00000000-0000-4000-8000-0000000a1001'
const moos = '00000000-0000-4000-8000-0000000c1003'

const region = (name) => `[...document.querySelectorAll('section')].find((s) => s.checkVisibility() && (s.getAttribute('aria-label') === ${JSON.stringify(name)} || s.querySelector('p')?.innerText.trim() === ${JSON.stringify(name)}))`
const fachkarte = (fach, name) => `[...((${region(fach)})?.querySelectorAll('button[aria-pressed]') ?? [])].find((b) => b.checkVisibility() && b.getAttribute('aria-label') === ${JSON.stringify(name)})`
const spaltentext = (titel) => `return (${region(titel)})?.innerText.replace(/\\s+/g, ' ').trim() ?? ''`
const schnellwahl = knopf('Neuer Tausch mit Bert')

const exemplar = (p, ort, druck, menge, name, nummer, absicht) =>
  p.mse(ort, 'exemplar', druck, String(menge), 'nearMint', 'de', 'nonfoil', '', '', '', '', name, 'p12', nummer, '', absicht)

export function horche(p, ort) {
  const prozess = spawn(join(process.cwd(), '.go/bin/mse'), ['horchen', '127.0.0.1:0'], {
    env: kindumgebung(ort),
    stdio: ['ignore', 'pipe', 'ignore'],
  })
  p.pids.push(prozess.pid)

  return new Promise((erfuellt, verworfen) => {
    let gelesen = ''
    prozess.stdout.on('data', (stueck) => {
      gelesen += stueck.toString()
      if (gelesen.includes('\n')) erfuellt(gelesen.split('\n')[0].trim())
    })
    prozess.on('exit', (code) => verworfen(new Error(`mse horchen endete mit ${String(code)}`)))
  })
}

// ziehen zieht eine Karte über AppKit auf eine Spalte und zählt den Zug als
// eine Handlung (dieselbe Zählung wie `ausgangswerte09spielen`).
async function ziehen(p, z, von, nach) {
  const a = await mitteVon(p, von, false)
  const b = await mitteVon(p, nach, false)
  if (!a || !b) return false
  await ziehe(p, [a, { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 }, b])
  if (z) await z.wahl('return true')

  return true
}

async function aufschlagen(p, z) {
  if (z) await z.klickAuf(schnellwahl, '»Neuer Tausch mit Bert«')
  else await p.klickeAuf(schnellwahl, '»Neuer Tausch mit Bert«')
  // **Der Tisch ist die Vorgabe, gemerkt wird je Rechner** (`useHandelsansicht`)
  // — und der Speicher des Abnahmebündels überlebt das Drehbuch davor:
  // `p11sicht09` stellt auf »Als Liste«. Zurückgestellt wird ohne Zählung,
  // denn auf einem frischen Rechner steht der Tisch von selbst da (QA01, #2382).
  await p.warte('»Als Tisch« steht', () => p.seite(`return !!${knopf('Als Tisch')}`), 10000)
  if ((await p.seite(`return ${knopf('Als Tisch')}.getAttribute('aria-pressed')`)) !== 'true') {
    await p.klickeAuf(knopf('Als Tisch'), '»Als Tisch«')
  }
  await p.warte('beide Fächer stehen', async () => Boolean(await p.seite(`return !!${fachkarte('Mein Fach', 'Glutläufer · P12 #1')} && !!${fachkarte('Fach von Bert', 'Sturmrufer · P12 #3')}`)), 20000)
}

export default async function p12markttisch09(p) {
  const annaOrt = p.ort()
  const bertOrt = p.ort()
  const wer = gekoppelt(p, { Anna: annaOrt, Bert: bertOrt })
  p.mse(annaOrt, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  exemplar(p, annaOrt, glut, 2, 'Glutläufer', '1', 'besitz')
  exemplar(p, annaOrt, moos, 1, 'Moosbär', '2', 'besitz')
  exemplar(p, annaOrt, sturm, 1, 'Sturmrufer', '3', 'wunsch')
  exemplar(p, bertOrt, sturm, 2, 'Sturmrufer', '3', 'besitz')
  p.mse(bertOrt, 'kontakt-freigeben', wer.Anna.kennung, 'handel', 'sammlung')
  const adresse = await horche(p, bertOrt)
  p.mse(annaOrt, 'kontakt-adresse', wer.Bert.kennung, adresse)

  const datensatz = { bauabdruck: p.lauf?.bauabdruck ?? '', erhoben: new Date().toISOString(), fenster: '1440 × 900', wege: [] }
  const z = wegzaehler(p, datensatz)
  await p.starte({ ort: annaOrt })
  await p.groesse(1440, 900)

  p.notiz('H1 — einen Tausch mit je einer Karte pro Seite anbieten, am Tisch')
  await z.beginne('H1', 'bereich', 'maus')
  await z.taste('5', 'befehl')
  await p.warte('»Neuer Tausch mit Bert« steht', () => p.seite(`return !!${schnellwahl}`), 20000)
  await aufschlagen(p, z)
  p.pruefe('H3: Berts Fach steht beim Aufschlagen da', true)
  // **Und was zusammenpasst, ohne Klick** (#2494) — in der Mitte, aus der
  // Sicht, die das Fach schon geholt hat.
  const passung = await p.warte('was zusammenpasst steht da', () => p.seite(`return [...document.querySelectorAll('[role=group]')].find((g) => g.checkVisibility() && g.getAttribute('aria-label') === 'Was dein Gegenüber freigibt')?.innerText.replace(/\\s+/g, ' ').trim() ?? null`), 15000).catch(() => null)
  p.pruefe('H3: was zusammenpasst, steht beim Aufschlagen da', passung !== null && passung.includes('Sturmrufer') && passung.includes('Vorschlag übernehmen'), String(passung))
  p.pruefe('H3: kein Knopf »Freigegebene Sicht anfragen« am Tisch', !(await p.seite(`return !!${knopf('Freigegebene Sicht anfragen')}`)))
  await ziehen(p, z, fachkarte('Mein Fach', 'Glutläufer · P12 #1'), region('Von dir'))
  await p.warte('Glutläufer liegt bei „Von dir“', async () => (await p.seite(spaltentext('Von dir'))).includes('Glutläufer'), 10000).catch(() => false)
  await ziehen(p, z, fachkarte('Fach von Bert', 'Sturmrufer · P12 #3'), region('Vom Gegenüber'))
  await p.warte('Sturmrufer liegt bei „Vom Gegenüber“', async () => (await p.seite(spaltentext('Vom Gegenüber'))).includes('Sturmrufer'), 10000).catch(() => false)
  await p.bild('tisch-hell')
  // **„Anbieten und zustimmen“ seit P13** (#2381, Vermerk 0200): ein Zug je
  // Entscheidung. Hier stand »Anbieten«, und seit P13 lief diese Probe nicht.
  await z.klickAuf(knopf('Anbieten und zustimmen'), '»Anbieten und zustimmen«')
  const angeboten = await p.warte('angeboten', () => handel(p, annaOrt).zeilen.find((x) => x.vorgang.zustand === 'angeboten')?.vorgang ?? null, 15000).catch(() => null)
  const h1 = await z.ende({ wirkung: angeboten !== null })
  p.zeige(`H1: ${String(h1.handlungen)} Handlungen (Ausgang 13, P10 ≤ 7, P12 ≤ 5)`)
  p.pruefe('H1 ≤ 5', h1.handlungen <= 5, String(h1.handlungen))
  const positionen = angeboten === null ? [] : JSON.parse(p.mse(annaOrt, 'handel', 'stand', angeboten.vorgang, '--json')).positionen
  p.pruefe('angeboten mit genau diesen zwei Zeilen', positionen.length === 2
    && positionen.some((x) => x.von === wer.Anna.kennung && x.printing === glut)
    && positionen.some((x) => x.von === wer.Bert.kennung && x.printing === sturm), JSON.stringify(positionen))

  p.notiz('H2 — ein Stück aus dem eigenen Fach in den Tausch')
  await aufschlagen(p, null)
  await z.beginne('H2', 'tisch', 'maus')
  await ziehen(p, z, fachkarte('Mein Fach', 'Glutläufer · P12 #1'), region('Von dir'))
  const h2wirkung = await p.warte('es liegt', async () => (await p.seite(spaltentext('Von dir'))).includes('Glutläufer'), 10000).catch(() => false)
  const h2 = await z.ende({ wirkung: h2wirkung })
  p.pruefe('H2 ≤ 2', h2.handlungen <= 2, String(h2.handlungen))

  p.notiz('Ein falsches Ziel legt nichts')
  await ziehen(p, null, fachkarte('Fach von Bert', 'Sturmrufer · P12 #3'), region('Von dir'))
  p.pruefe('Berts Karte liegt nicht bei „Von dir“', !(await p.seite(spaltentext('Von dir'))).includes('Sturmrufer'))

  p.notiz('Mehrfachwahl mit ⌘ und ein Zug')
  await aufschlagen(p, null)
  await p.klickeAuf(fachkarte('Mein Fach', 'Glutläufer · P12 #1'), 'Glutläufer mit ⌘', { mit: ['befehl'] })
  await p.klickeAuf(fachkarte('Mein Fach', 'Moosbär · P12 #2'), 'Moosbär mit ⌘', { mit: ['befehl'] })
  await ziehen(p, null, fachkarte('Mein Fach', 'Moosbär · P12 #2'), region('Von dir'))
  const beide = await p.warte('beide liegen', async () => { const t = await p.seite(spaltentext('Von dir')); return t.includes('Glutläufer') && t.includes('Moosbär') ? t : null }, 10000).catch(() => '')
  p.pruefe('zwei Karten in einem Zug', beide !== '', beide)

  p.notiz('Mit der Tastatur: Leertaste greift, Eingabe legt')
  await p.seite(`(${fachkarte('Fach von Bert', 'Sturmrufer · P12 #3')}).focus(); return true`)
  await p.taste(' ')
  await p.taste('Enter')
  p.pruefe('Tastatur legt nach „Vom Gegenüber“', await p.warte('Sturmrufer liegt', async () => (await p.seite(spaltentext('Vom Gegenüber'))).includes('Sturmrufer'), 10000).catch(() => false))

  p.notiz('Der bisherige Assistent als Liste')
  await p.klickeAuf(knopf('Als Liste'), '»Als Liste«')
  p.pruefe('die Liste steht ohne Fächer', await p.warte('keine Fächer', async () => !(await p.seite(`return !!${region('Mein Fach')}`)), 5000).catch(() => false))
  await p.klickeAuf(knopf('Als Tisch'), '»Als Tisch«')
  p.pruefe('⌘Q beendet', await p.beende())

  p.notiz('Dunkel')
  p.mse(annaOrt, 'einstellungen', 'setzen', 'erscheinungsbild', 'dunkel')
  await p.starte({ ort: annaOrt })
  await p.groesse(1440, 900)
  await p.taste('5', 'befehl')
  await p.warte('»Neuer Tausch mit Bert« steht', () => p.seite(`return !!${schnellwahl}`), 20000)
  await aufschlagen(p, null)
  await ziehen(p, null, fachkarte('Mein Fach', 'Glutläufer · P12 #1'), region('Von dir'))
  await p.bild('tisch-dunkel')
  p.pruefe('⌘Q beendet', await p.beende())

  writeFileSync(ziel, `${JSON.stringify(datensatz, null, 2)}\n`)
}
