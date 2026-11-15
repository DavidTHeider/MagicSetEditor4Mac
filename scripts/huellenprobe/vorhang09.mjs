// **Der Vorhang am gemeinsamen Gerät** — P7
// (#2375),
// Befund C-05, Artboard `Vorhang09`, am Bündel aus `scripts/bauen.sh abnahme`.
//
// # Was es belegt
//
// **S1 und S2 am neuen Bündel** — der Anfang aus P6 und der Beginn bis Zug 1,
// jetzt mit Vorhang: „Hand behalten“, „Ich bin …“, „Hand behalten“. **Der
// Vorhang zeigt keine Hand**, weder die des vorigen noch die des neuen Sitzes.
// **Esc** lässt den bisherigen Sitz weiterhandeln, **↩** bestätigt den neuen
// — der Fokus steht dafür auf „Ich bin …“. **Nachladen**: mit verzögerter
// Sicht (`Partie` um 1,5 s aufgehalten) steht in keinem Bild der Tisch des
// vorigen Sitzes; mit einer gescheiterten steht die Meldung und kein Tisch.
// **Der Zuschauer** übernimmt keinen Sitz, wenn die Priorität wechselt.
// **„Offene Hände“** folgt ohne Vorhang. **Ein Profilwechsel** lässt keinen
// Tisch des vorigen Profils stehen. Dazu axe in WebKit am Vorhang.
//
// # Was es nicht belegt
//
// **VoiceOver** hört ein Mensch — hier steht nur, dass Name, Fokus und
// Tastenkürzel dastehen. **Zwei Geräte und ein echter Netzverlust** gehören
// zu P9; ein Sitz, den ein Gast spielt, ist im Einheitstest belegt.
import { readFileSync, writeFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

import { stoere } from '../huellenrufbuch.mjs'
import { knopf, sichtwahl } from '../huellenspieltisch.mjs'
import { wegzaehler } from '../huellenwege.mjs'

const wurzel = resolve(import.meta.dirname, '..', '..')
const karte = { waldboden: '1', bergpfad: '2', moosbaer: '3', glutlaeufer: '4', funkenregen: '5', rindenwall: '6' }
const namen = ['Waldboden', 'Bergpfad', 'Moosbär', 'Glutläufer', 'Funkenregen', 'Rindenwall']

const partien = (p, ort) => JSON.parse(p.mse(ort, 'partien', '--json')).partien
const zeigen = (p, ort, partie) => JSON.parse(p.mse(ort, 'partie-zeigen', partie, 'sitz1', '--json')).partie
const anfangen = "[...document.querySelectorAll('main button')].filter((b) => b.checkVisibility() && b.innerText.trim() === 'Partie anfangen').at(-1)"
const amTisch = "return !!document.querySelector('canvas')?.checkVisibility()"
const sicht = "([...document.querySelectorAll('label')].find((l) => l.checkVisibility() && l.textContent.startsWith('Sicht für'))?.querySelector('select')?.value ?? null)"
// vorhang ist der sichtbare Vorhang — `wer` leer: irgendeiner.
const vorhang = (wer = '') => `[...document.querySelectorAll('section')].find((s) => s.checkVisibility() && (document.getElementById(s.getAttribute('aria-labelledby') ?? '')?.textContent ?? '').trim().startsWith(${JSON.stringify(wer === '' ? 'Gerät an ' : `Gerät an ${wer} weitergeben`)}))`
// blick ist ein Bild dessen, was ein Mensch vor dem Gerät sieht.
const blick = `return { tisch: [...document.querySelectorAll('[role=application], canvas')].some((e) => e.checkVisibility()), namen: ${JSON.stringify(namen)}.filter((n) => document.body.innerText.includes(n)), sicht: ${sicht}, laedt: document.body.innerText.includes('Die gewählte Sicht wird geladen'), vorhang: !!${vorhang()} }`

function bestand(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  p.mse(ort, 'mtgjson-lesen', join(wurzel, 'testdaten/mtgjson/tisch/TSP.json'), 'magic')
  p.mse(ort, 'deck', 'neu', 'Waldschritt', 'magic')
  for (const [k, n] of [[karte.waldboden, 2], [karte.moosbaer, 1], [karte.glutlaeufer, 1], [karte.funkenregen, 1], [karte.rindenwall, 2]]) p.mse(ort, 'deck', 'legen', '1', `--karte=${k}`, String(n))
  p.mse(ort, 'deck', 'neu', 'Glutpfad', 'magic')
  for (const [k, n] of [[karte.bergpfad, 16], [karte.glutlaeufer, 4], [karte.rindenwall, 4]]) p.mse(ort, 'deck', 'legen', '2', `--karte=${k}`, String(n))
  const nutzer = JSON.parse(p.mse(ort, 'nutzer', '--json')).nutzer[0].id
  p.mse(ort, 'profil-neu', nutzer, 'Zweitprofil')

  return ort
}

// beobachte sammelt Bilder, bis `fertig` gilt oder die Frist um ist.
async function beobachte(p, frist, fertig = () => false) {
  const bilder = []
  const bis = Date.now() + frist
  while (Date.now() < bis) {
    const jetzt = await p.seite(blick)
    bilder.push(jetzt)
    if (fertig(jetzt)) break
    await new Promise((weiter) => setTimeout(weiter, 40))
  }

  return bilder
}

// uebergabe sorgt für einen Übergang an diesem Gerät — der gezeigte Sitz
// passt, und die Priorität geht an den anderen. Hat gerade der andere sie,
// passt er zuerst an der Kommandozeile (Weg, kein Messwert).
async function uebergabe(p, ort, partie) {
  const gezeigt = await p.seite(`return ${sicht}`)
  const zug = zeigen(p, ort, partie).zug
  if (zug.prioritaet !== gezeigt) {
    p.mse(ort, 'partie-befehl', partie, zug.prioritaet, `weg-${String(Date.now())}`, 'zug.passen')
    await p.warte('die Priorität ist beim gezeigten Sitz', () => zeigen(p, ort, partie).zug.prioritaet === gezeigt, 15000)
  }
  await p.warte('»Passen« steht bereit', () => p.seite(`const k = ${knopf('Passen')}; return !!k && !k.disabled`), 15000)
  await p.klickeAuf(knopf('Passen'), '»Passen«')
  await p.warte('der Vorhang ist zu', () => p.seite(`return !!${vorhang()}`), 15000)
}

async function axeAmVorhang(p) {
  const quelle = readFileSync(join(wurzel, 'frontend/node_modules/axe-core/axe.min.js'), 'utf8')
  await p.seite(`${quelle}; return true`)
  const befund = await p.seite("const r = await axe.run(document, { runOnly: ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa', 'wcag22aa'] }); return r.violations.map((v) => `${v.id} (${v.nodes.length}): ${v.nodes.slice(0, 3).map((n) => n.target.join(' ')).join(' | ')}`)")
  p.pruefe('axe in WebKit: keine Verstöße gegen WCAG 2.2 AA am Vorhang', befund.length === 0, befund.join(' · '))
}

export default async function vorhang09(p) {
  const ort = bestand(p)
  const datensatz = { bauabdruck: p.lauf?.bauabdruck ?? '', erhoben: new Date().toISOString(), fenster: '1440 × 900', wege: [] }
  const z = wegzaehler(p, datensatz)

  await p.starte({ ort })
  await p.groesse(1440, 900)

  p.notiz('S1 — aus dem Bereich, am neuen Bündel')
  const vorher = partien(p, ort).length
  await z.beginne('S1', 'bereich', 'maus')
  await z.taste('4', 'befehl')
  await p.warte('»Partie anfangen« steht', () => p.seite(`return !!(${anfangen})`), 20000)
  await z.klickAuf(anfangen, '»Partie anfangen« oben')
  await p.warte('das Formular steht vorbelegt', () => p.seite(`const k = ${anfangen}; return !!k && !k.disabled && [...document.querySelectorAll('input')].some((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').startsWith('Name am Tisch'))`), 20000)
  await z.klickAuf(anfangen, '»Partie anfangen«')
  await p.warte('der Tisch steht', () => p.seite(amTisch), 30000)
  await z.ende({ wirkung: partien(p, ort).length === vorher + 1 })
  const partie = partien(p, ort)[0].id
  const name = Object.fromEntries(zeigen(p, ort, partie).sitze.map((s) => [s.kennung, s.name]))

  p.notiz('S2 — bis Zug 1: behalten, der Vorhang, behalten')
  await z.beginne('S2', 'tisch', 'maus')
  await p.warte('»Hand behalten« steht (erster Sitz)', () => p.seite(`return !!${knopf('Hand behalten')}`), 15000)
  await z.text('Hand behalten')
  await p.warte(`der Vorhang für ${name.sitz2} ist zu`, () => p.seite(`return !!${vorhang(name.sitz2)}`), 15000)
  const zu = await beobachte(p, 600)
  p.pruefe('hinter dem Vorhang kein Tisch und keine Karte', zu.every((b) => !b.tisch && b.namen.length === 0), JSON.stringify(zu.at(-1)))
  p.pruefe('der Vorhang wechselt die Sicht nicht von selbst', zu.every((b) => b.sicht === 'sitz1'), JSON.stringify(zu.map((b) => b.sicht)))
  const fokus = await p.seite(`return { fokus: document.activeElement?.innerText?.trim() ?? '', esc: [...(${vorhang()})?.querySelectorAll('button') ?? []].find((b) => b.getAttribute('aria-keyshortcuts') === 'Escape')?.innerText?.trim() ?? '' }`)
  p.pruefe('der Fokus steht auf »Ich bin …«, Esc ist angesagt', fokus.fokus === `Ich bin ${name.sitz2}` && fokus.esc === `Weiter als ${name.sitz1}`, JSON.stringify(fokus))
  await p.bild('vorhang')
  await axeAmVorhang(p)
  await z.klickAuf(knopf(`Ich bin ${name.sitz2}`), `»Ich bin ${name.sitz2}«`)
  await p.warte('»Hand behalten« steht (zweiter Sitz)', async () => (await p.seite(`return ${sicht}`)) === 'sitz2' && p.seite(`return !!${knopf('Hand behalten')}`), 15000)
  await z.text('Hand behalten')
  await p.warte('Zug 1 läuft', () => zeigen(p, ort, partie).zug.nummer === 1, 15000).catch(() => false)
  await z.ende({ wirkung: zeigen(p, ort, partie).zug.nummer === 1 && zeigen(p, ort, partie).sitze.every((s) => s.eroeffnung === 'behalten') })

  p.notiz('Esc lässt den bisherigen Sitz weiterhandeln')
  if (!(await p.seite(`return !!${vorhang()}`))) await uebergabe(p, ort, partie)
  const bisher = await p.seite(`return ${sicht}`)
  await p.taste('Escape')
  await p.warte('der Vorhang ist auf', () => p.seite(`return !${vorhang()}`), 10000)
  const nachEsc = await p.seite(blick)
  p.pruefe('nach Esc steht der Tisch des bisherigen Sitzes', nachEsc.tisch && nachEsc.sicht === bisher, JSON.stringify(nachEsc))

  p.notiz('↩ und verzögertes Nachladen — kein Bild mit dem Tisch des vorigen Sitzes')
  await uebergabe(p, ort, partie)
  const von = await p.seite(`return ${sicht}`)
  const nach = von === 'sitz1' ? 'sitz2' : 'sitz1'
  p.pruefe('der Fokus steht auf »Ich bin …«', (await p.seite('return document.activeElement?.innerText?.trim() ?? ""')) === `Ich bin ${name[nach]}`)
  await stoere(p, [{ bindung: 'Partie', verzoegerungMs: 1500 }])
  await p.taste('Enter')
  const laden = await beobachte(p, 6000, (b) => b.tisch && b.sicht === nach)
  await stoere(p, [])
  const verletzt = laden.filter((b) => b.tisch && b.sicht !== nach)
  p.pruefe('während des Ladens zeigt kein Bild den Tisch des vorigen Sitzes', verletzt.length === 0, `${String(laden.length)} Bilder, ${String(laden.filter((b) => b.laedt).length)} mit »wird geladen«, ${String(verletzt.length)} mit fremdem Tisch`)
  p.pruefe('das Laden war sichtbar verzögert', laden.filter((b) => b.laedt).length >= 5, String(laden.filter((b) => b.laedt).length))
  p.pruefe('danach steht der neue Sitz', laden.at(-1)?.tisch === true && laden.at(-1)?.sicht === nach, JSON.stringify(laden.at(-1)))
  p.pruefe('der Fokus steht danach an der Überschrift des Tisches', /^H[1-3]$/.test(await p.seite('return document.activeElement?.tagName ?? ""')))

  p.notiz('Gescheitertes Nachladen — die Meldung und kein Tisch')
  await uebergabe(p, ort, partie)
  const vorFehler = await p.seite(`return ${sicht}`)
  await stoere(p, [{ bindung: 'Partie', fehler: true }])
  await p.taste('Enter')
  const fehlbild = await beobachte(p, 2000)
  await stoere(p, [])
  p.pruefe('nach einem gescheiterten Laden kein Tisch und keine Karte', fehlbild.every((b) => !b.tisch && b.namen.length === 0), JSON.stringify(fehlbild.at(-1)))
  p.pruefe('die Meldung steht da', await p.seite("return [...document.querySelectorAll('[role=alert]')].some((e) => e.checkVisibility() && e.innerText.trim() !== '')"))
  await p.seite(sichtwahl(vorFehler === 'sitz1' ? 'sitz2' : 'sitz1'))
  await p.warte('über die Sichtwahl steht wieder ein Tisch', () => p.seite(amTisch), 15000)

  p.notiz('Der Zuschauer übernimmt keinen Sitz')
  await p.seite(sichtwahl(''))
  await p.warte('die Sicht des Zuschauers steht', async () => (await p.seite(`return ${sicht}`)) === '' && p.seite(amTisch), 15000)
  const zug = zeigen(p, ort, partie).zug
  p.mse(ort, 'partie-befehl', partie, zug.prioritaet, `zuschauer-${String(Date.now())}`, 'zug.passen')
  const neu = zeigen(p, ort, partie)
  // **Vor dem Nein der Beleg, dass jemand zuhört** (§3): das Fenster muss
  // die neue Fassung zeigen, sonst sagt „kein Vorhang“ nichts.
  const gesehen = await p.warte('das Fenster kennt die neue Fassung', () => p.seite(`return /Fassung ${String(neu.fassung)}\\b/.test(document.body.innerText)`), 15000).then(() => true, () => false)
  p.pruefe('das Fenster hat den Wechsel der Priorität gelesen', gesehen && neu.zug.prioritaet !== zug.prioritaet, `${zug.prioritaet} → ${neu.zug.prioritaet}, Fassung ${String(neu.fassung)}`)
  const zuschauer = await beobachte(p, 1500)
  p.pruefe('der Zuschauer bleibt Zuschauer, ohne Vorhang', zuschauer.every((b) => b.sicht === '' && !b.vorhang), JSON.stringify(zuschauer.at(-1)))

  p.notiz('„Offene Hände“ folgt ohne Vorhang')
  await p.seite(sichtwahl(zeigen(p, ort, partie).zug.prioritaet))
  await p.warte('der Sitz mit Priorität steht', () => p.seite(amTisch), 15000)
  await p.klickeAuf("[...document.querySelectorAll('input[type=checkbox]')].find((e) => e.checkVisibility() && e.closest('label')?.innerText.trim() === 'Offene Hände')", '»Offene Hände«')
  const vorOffen = await p.seite(`return ${sicht}`)
  await p.warte('»Passen« steht bereit', () => p.seite(`const k = ${knopf('Passen')}; return !!k && !k.disabled`), 15000)
  await p.klickeAuf(knopf('Passen'), '»Passen«')
  const offen = await beobachte(p, 6000, (b) => b.tisch && b.sicht !== vorOffen)
  p.pruefe('die Sicht folgt ohne Vorhang', offen.every((b) => !b.vorhang) && offen.at(-1)?.sicht !== vorOffen, `${String(offen.length)} Bilder, zuletzt ${JSON.stringify(offen.at(-1))}`)

  p.notiz('Ein Profilwechsel lässt keinen Tisch des vorigen Profils stehen')
  const davor = await p.seite(blick)
  p.pruefe('vor dem Wechsel steht ein Tisch mit Karten', davor.tisch && davor.namen.length > 0, JSON.stringify(davor))
  // **Der Knopf trägt den Namen des Profils** — „Profil wechseln …“ ist
  // seine Beschreibung (`Kontextleiste.tsx`).
  const profilknopf = "[...document.querySelectorAll('button[aria-describedby]')].find((b) => b.checkVisibility() && document.getElementById(b.getAttribute('aria-describedby'))?.textContent.trim() === 'Profil wechseln …')"
  await p.klickeAuf(profilknopf, '»Profil wechseln …«')
  const hierhin = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '').endsWith('· Zweitprofil wechseln'))"
  await p.warte('die Profile stehen da', () => p.seite(`return !!${hierhin}`), 15000)
  await p.klickeAuf(hierhin, '»Zu … · Zweitprofil wechseln«')
  const gewechselt = await p.warte('das Zweitprofil ist gezeigt', () => p.mse(ort, 'profil').includes('Zweitprofil'), 15000).then(() => true, () => false)
  p.pruefe('das Profil ist gewechselt', gewechselt, p.mse(ort, 'profil').trim().split('\n')[0])
  await p.taste('4', 'befehl')
  const nachWechsel = await beobachte(p, 2000)
  p.pruefe('nach dem Profilwechsel kein Tisch und keine Karte', nachWechsel.every((b) => !b.tisch && b.namen.length === 0), JSON.stringify(nachWechsel.at(-1)))

  writeFileSync(join(wurzel, 'docs/belege/vorhang09-wege.json'), `${JSON.stringify(datensatz, null, 2)}\n`)
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
