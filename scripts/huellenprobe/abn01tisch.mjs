// **Hundert Tischbesuche in einer Sitzung** — ABN-01.AK-04
// (#2175).
//
// `tischwiedereinstieg` startet je Erscheinungsbild neu und fährt zehn
// Zyklen. Hier ist es **eine** Sitzung mit hundert Besuchen, und zwischen
// zweien liegt immer eine andere Seite.
//
// **Der Fallmix steht vor dem Lauf fest** (AK-04), nach der Nummer des
// Besuchs:
//
// | Besuche | Weg an den Tisch |
// | --- | --- |
// | 1, 34, 67 | **neu**: »Partie schließen«, das Formular, »Partie anfangen« |
// | jede fünfte | **Rückkehr**: »Partie schließen«, die Liste, »Weiterspielen« einer anderen bestehenden Partie, »Weiterspielen als Ada« |
// | die übrigen | **Fortsetzen**: Decks (⌘4) → »Partie anfangen« → der offene Tisch |
//
// Vor jedem Besuch eine andere Seite — Start, Sets, Sammlung, Handeln und
// Pakete im Wechsel —, bei den Besuchen 3, 13, …, 93 die Einstellungen und
// dort das **andere** Erscheinungsbild: zehn Wechsel, über die Bedienung und
// nicht über `mse`.
//
// **Positiv belegt je Besuch**, bevor irgendetwas verneint wird: die
// erwartete Partie steht im Seitenweg, eine **neue** Leinwand mit lebendem
// WebGL-Kontext, so viele Karten wie beim ersten Besuch dieser Partie, die
// Schrift des Tischs hell im dunklen Bild und dunkel im hellen, und »Näher
// heran« rückt sie. Erst danach: keine weitere Leinwand im Dokument.
//
// **Die Schrift steht dabei, weil sie fehlte**: im ersten Lauf dieses
// Drehbuchs stand der Zonenname auf dem dunklen Tisch bei 1,03 : 1 —
// `stil.css` gab einem gewählten „dunkel" die Tischmerkmale des hellen
// Satzes (`frontend/src/dunkelgewaehlt.test.ts`).
//
// **Die Abbruchregel steht vor dem Lauf fest**: der erste rote Besuch beendet
// den Lauf — keine Wiederholung, kein Überspringen. „Blockiert" heißt: der
// Tisch zeichnet nicht binnen 20 s, oder die Webansicht beantwortet einen
// Ruf nicht in seiner Frist.
//
// **Gesamtfrist und Lastregel stehen ebenso fest** (#2088): vor dem ersten
// Besuch wird auf Ruhe gewartet — Last höchstens die halbe Kernzahl, fremde
// Prozesse höchstens 250 % —, längstens 120 s, sonst gilt der Lauf als nicht
// gemessen und ist rot. Alle hundert Besuche zusammen haben 30 Minuten.
//
// **Bewertet** wird außerdem die Zeit: der Median der Besuche 91–100 höchstens
// doppelt so lang wie der der Besuche 2–11. **Gemessen und nicht bewertet**:
// der Speicher von Anwendung, WebContent und GPU bei Besuch 1, 50 und 100 —
// eine Schwelle dafür gibt es in 0.7.0 nicht; dass der Webprozess am Tisch
// sich einpendelt, belegt #1852 für den Tisch allein.
//
// **Was es nicht belegt**: ob ein Mensch den Tisch nach hundert Besuchen so
// bedienen kann wie nach einem; einen Tisch mit Partnern im Netz; einen
// Lauf über Stunden.
import { execFileSync } from 'node:child_process'
import { cpus, loadavg } from 'node:os'

import { fremdLast, warteAufRuhe } from '../leistungsschnittteile.mjs'

const besuche = Number(process.env.MSE_TISCH_BESUCHE ?? 100)
const neuBei = new Set([1, 34, 67])
const bildBei = (nummer) => nummer % 10 === 3
const seiten = [['1', 'start'], ['2', 'sets'], ['3', 'sammlung'], ['5', 'handeln'], ['6', 'pakete']]
const neueGegner = ['Dora', 'Emil', 'Fritz']
const gesamtfrist = 30 * 60 * 1000
const grenzen = { last: cpus().length / 2, fremd: 250 }
const umgebung = () => ({
  last: loadavg().map((w) => Math.round(w * 100) / 100),
  fremd: fremdLast(execFileSync('ps', ['-A', '-o', 'pid=,ppid=,pcpu=,comm=']).toString(), process.pid),
})

export function fallFuer(nummer) {
  if (neuBei.has(nummer)) return 'neu'

  return nummer % 5 === 0 ? 'rueckkehr' : 'fortsetzen'
}

const tischbefund = (p) => p.seite(`
  const tisch = document.querySelector('[role=application]')
  const leinwand = tisch?.querySelector('canvas')
  const gl = leinwand && (leinwand.getContext('webgl2') ?? leinwand.getContext('webgl'))
  const neu = !!leinwand && !leinwand.dataset.gesehen
  if (leinwand) leinwand.dataset.gesehen = 'ja'
  return {
    sichtbar: !!leinwand && leinwand.checkVisibility() && leinwand.width > 0,
    neu,
    verloren: gl ? gl.isContextLost() : null,
    karten: tisch ? tisch.querySelectorAll('li').length : 0,
    leinwaende: document.querySelectorAll('canvas').length,
    bild: document.documentElement.dataset.erscheinungsbild ?? '',
    schriftHell: (() => { const c = getComputedStyle(document.documentElement).getPropertyValue('--mse-tisch-schrift').trim(); const n = parseInt(c.slice(1), 16); return ((n >> 16) & 255) + ((n >> 8) & 255) + (n & 255) > 384 })(),
    text: document.body.innerText,
  }`)

const knopf = (text) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') === ${JSON.stringify(text)} || b.innerText.trim() === ${JSON.stringify(text)}))`

// prozesse liest Kennung, Speicher und Namen jedes Prozesses — die
// WebContent- und GPU-Prozesse gehören launchd und nicht der Anwendung; zu
// ihr gehören die, die nach ihrem Start entstanden sind (wie in
// `scripts/tischframezeit.mjs`).
const prozesse = () => execFileSync('ps', ['-Ao', 'pid=,rss=,comm='])
  .toString().split('\n').map((z) => z.trim().split(/\s+/)).filter((t) => t.length >= 3)
  .map(([pid, rss, ...name]) => ({ pid, rss: Number(rss), name: name.join(' ') }))

function speicher(vorher, anwendung) {
  const neue = prozesse().filter((x) => !vorher.has(x.pid))
  const summe = (name) => Math.round(neue.filter((x) => x.name.includes(name)).reduce((alle, x) => alle + x.rss, 0) / 1024)

  return { anwendung: Math.round((prozesse().find((x) => x.pid === String(anwendung))?.rss ?? 0) / 1024), web: summe('WebKit.WebContent'), gpu: summe('WebKit.GPU') }
}

const median = (werte) => {
  const sortiert = [...werte].sort((a, b) => a - b)

  return sortiert[Math.floor(sortiert.length / 2)]
}

// woanders geht auf eine andere Seite und wartet, bis sie steht.
async function woanders(p, nummer) {
  if (bildBei(nummer)) {
    await p.taste(',', 'befehl')
    await p.warte('die Einstellungen stehen', () => p.seite("return document.querySelector('main')?.dataset.seite === 'einstellungen'"), 20000)
    // **Die Gruppe wählen, wo sie nicht schon offen ist** — die Einstellungen
    // öffnen auf „Allgemein", und das Thema steht unter „Erscheinungsbild".
    const thema = "return !!document.querySelector('[data-einstellung=\"erscheinungsbild\"]')?.checkVisibility()"
    if (!(await p.seite(thema))) {
      await p.klickeText('Erscheinungsbild')
      await p.warte('das Thema steht da', () => p.seite(thema), 20000)
      await p.ruhe()
    }
    const jetzt = await p.seite('return document.documentElement.dataset.erscheinungsbild ?? ""')
    const ziel = jetzt === 'dunkel' ? 'hell' : 'dunkel'
    await p.klickeAuf(`(() => { const i = document.querySelector('[data-einstellung="erscheinungsbild"] input[value=${ziel}]'); return i?.labels?.[0] ?? i })()`, `»${ziel}«`)
    await p.warte(`das Fenster ist ${ziel}`, () => p.seite(`return document.documentElement.dataset.erscheinungsbild === ${JSON.stringify(ziel)}`), 10000)
    await p.warte('das Thema ist geschrieben', () => p.seite("const felder = [...document.querySelectorAll('[data-einstellung=\"erscheinungsbild\"] input')]; return felder.length > 0 && felder.every((f) => !f.disabled)"), 20000)

    return ziel
  }
  const [taste, schluessel] = seiten[nummer % seiten.length]
  await p.taste(taste, 'befehl')
  await p.warte(`»${schluessel}« steht`, () => p.seite(`return document.querySelector('main')?.dataset.seite === ${JSON.stringify(schluessel)} && !document.querySelector('[role=application] canvas')`), 20000)

  return undefined
}

// zumOffenenTisch geht über Decks und »Partie anfangen« an den Tisch, der
// offen steht — der ursprüngliche Weg aus UI-N01.
async function zumOffenenTisch(p) {
  await p.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await p.klickeText('Partie anfangen')
}

// zurListe kommt zu Formular und Partienliste — über den offenen Tisch und
// »Partie schließen«, oder gleich, wenn keiner offen steht.
async function zurListe(p) {
  await zumOffenenTisch(p)
  const weg = await p.warte('der Tisch oder die Liste steht', () => p.seite(`return ${knopf('Partie schließen')} ? 'tisch' : document.body.innerText.includes('Eine Partie anfangen') ? 'liste' : ''`), 20000)
  if (weg === 'tisch') {
    await p.klickeAuf(knopf('Partie schließen'), '»Partie schließen«')
    await p.wartetext('Eine Partie anfangen', 20000)
  }
}

async function neuePartie(p, gegner) {
  await zurListe(p)
  for (const [stelle, name] of [[1, 'Ada'], [2, gegner]]) {
    await p.seite(`[...document.querySelectorAll('input')].find((i) => i.checkVisibility() && (i.getAttribute('aria-label') === 'Name am Tisch ${stelle}' || [...(i.labels ?? [])].some((l) => l.innerText.trim() === 'Name am Tisch ${stelle}'))).focus(); return true`)
    await p.tippeInsFeld(name)
  }
  await p.seite("for (const s of [...document.querySelectorAll('select')].filter((s) => s.checkVisibility() && [...(s.labels ?? [])].some((l) => l.innerText.trim().startsWith('Deck')))) { const setze = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set; setze.call(s, s.options[1].value); s.dispatchEvent(new Event('change', { bubbles: true })) } return true")
  await p.klickeAuf(`[...document.querySelectorAll('main button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Partie anfangen' && !b.disabled)`, '»Partie anfangen« im Formular')
}

async function rueckkehr(p, gegner) {
  await zurListe(p)
  const menue = `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '') === 'Weiterspielen: Ada · ${gegner}')`
  // **Ab der fünften Partie steht sie hinter der Falte** — die Liste zeigt
  // vier (`Partienliste`, `Vorn`), und »Weitere … zeigen« holt den Rest. Ein
  // Mensch klickt dort auch; gefunden im ersten Gesamtlauf, als Bo die fünfte
  // war.
  await p.warte('die Liste steht', () => p.seite("return [...document.querySelectorAll('button')].some((b) => (b.getAttribute('aria-label') ?? '').startsWith('Weiterspielen: '))"), 20000)
  if (!(await p.seite(`return !!${menue}`))) {
    const falte = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && /^Weitere \\d+ zeigen$/.test(b.innerText.trim()))"
    if (await p.seite(`return !!${falte}`)) await p.klickeAuf(falte, '»Weitere … zeigen«')
  }
  try {
    await p.warte(`»Weiterspielen« der Partie gegen ${gegner} steht`, () => p.seite(`return !!${menue}`), 20000)
  } catch (grund) {
    // **Was stattdessen dasteht** — ein Abbruch ohne die Namen der Menüs
    // sagte nur, dass etwas fehlt.
    p.zeige(`Menüs der Liste: ${JSON.stringify(await p.seite("return [...document.querySelectorAll('button')].filter((b) => (b.getAttribute('aria-label') ?? '').startsWith('Weiterspielen')).map((b) => [b.getAttribute('aria-label'), b.checkVisibility(), Math.round(b.getBoundingClientRect().top)])"))}`)
    throw grund
  }
  await p.klickeAuf(menue, `»Weiterspielen: Ada · ${gegner}«`)
  await p.klickeText('Weiterspielen als Ada')
}

// amTisch wartet, bis der Tisch der erwarteten Partie zeichnet.
function amTisch(p, gegner) {
  return p.warte(`der Tisch gegen ${gegner} zeichnet`, async () => {
    const befund = await tischbefund(p)

    return befund.sichtbar && befund.verloren === false && befund.karten > 0 && befund.text.includes(`Partie gegen ${gegner}`) ? befund : false
  }, 20000)
}

async function bedient(p) {
  const stelle = "return document.querySelector('[role=application] li')?.getAttribute('style') ?? ''"
  const vorher = await p.seite(stelle)
  await p.klickeText('Näher heran')

  return Boolean(await p.warte('»Näher heran« rückt die Karten', async () => (await p.seite(stelle)) !== vorher, 5000))
}

function rot(p, beschreibung, zeige) {
  p.pruefe(beschreibung, false, zeige)
  throw new Error(`abgebrochen nach der Abbruchregel: ${beschreibung}`)
}

export default async function abn01tisch(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Probe-Deck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')
  p.mse(ort, 'partie-anlegen', 'probe', 'ada:Ada:1', 'bo:Bo:1')
  p.mse(ort, 'partie-anlegen', 'probe', 'ada:Ada:1', 'cy:Cy:1')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', 'hell')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', '1280')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', '800')

  const mix = Array.from({ length: besuche }, (_, i) => fallFuer(i + 1))
  p.notiz(`${String(besuche)} Besuche in einer Sitzung bei 1280 × 800: ${String(mix.filter((f) => f === 'neu').length)} neu, ${String(mix.filter((f) => f === 'rueckkehr').length)} Rückkehr, ${String(mix.filter((f) => f === 'fortsetzen').length)} Fortsetzen; ${String(mix.filter((_, i) => bildBei(i + 1)).length)} Wechsel des Erscheinungsbilds`)

  const ruhe = await warteAufRuhe({ messen: umgebung, schlafe: (s) => new Promise((los) => { setTimeout(los, s * 1000) }), grenzen, frist: 120, takt: 5 })
  p.zeige(`Umgebung vor dem Lauf: Last ${String(ruhe.umgebung.last[0])}, fremd ${String(ruhe.umgebung.fremd.summe)} % (${ruhe.umgebung.fremd.groesste.map((x) => `${x.name} ${String(x.cpu)}`).join(', ')}), gewartet ${String(ruhe.gewartetS)} s`)
  if (!ruhe.ruhig) rot(p, 'die Umgebung ist ruhig genug für den Lauf', ruhe.grund)

  const vorher = new Set(prozesse().map((x) => x.pid))
  await p.starte({ ort })
  const ende = Date.now() + gesamtfrist
  const anwendung = (await p.fenster()).prozess

  const partien = [{ gegner: 'Bo' }, { gegner: 'Cy' }]
  let offen
  let rueckkehrZeiger = 0
  let leinwaende = 0
  const zeiten = []
  const speicherstaende = []
  const bilder = []

  for (let nummer = 1; nummer <= besuche; nummer += 1) {
    const fall = fallFuer(nummer)
    const anfang = Date.now()
    if (anfang > ende) rot(p, `alle Besuche binnen ${String(gesamtfrist / 60000)} Minuten`, `bei Besuch ${String(nummer)} abgelaufen`)
    if (nummer > 1) {
      const bild = await woanders(p, nummer)
      if (bild) bilder.push(bild)
    }

    let ziel
    if (fall === 'neu') {
      ziel = { gegner: neueGegner[[...neuBei].indexOf(nummer)] }
      partien.push(ziel)
      await neuePartie(p, ziel.gegner)
    } else if (fall === 'rueckkehr' || offen === undefined) {
      const andere = partien.filter((x) => x !== offen)
      ziel = andere[rueckkehrZeiger % andere.length]
      rueckkehrZeiger += 1
      await rueckkehr(p, ziel.gegner)
    } else {
      ziel = offen
      await zumOffenenTisch(p)
    }

    const befund = await amTisch(p, ziel.gegner)
    zeiten.push(Date.now() - anfang)
    ziel.karten ??= befund.karten
    if (befund.schriftHell !== (befund.bild === 'dunkel')) rot(p, `Besuch ${String(nummer)}: die Schrift des Tischs folgt dem Bild »${befund.bild}«`, `hell: ${String(befund.schriftHell)}`)
    if (!befund.neu || befund.karten !== ziel.karten) rot(p, `Besuch ${String(nummer)} (${fall}, gegen ${ziel.gegner}): eine neue Leinwand mit ${String(ziel.karten)} Karten`, JSON.stringify({ ...befund, text: undefined }))
    if (!(await bedient(p))) rot(p, `Besuch ${String(nummer)}: »Näher heran« wirkt`, '')
    if (nummer > 1 && befund.leinwaende > leinwaende) rot(p, `Besuch ${String(nummer)}: keine weitere Leinwand`, `${String(leinwaende)} → ${String(befund.leinwaende)}`)
    leinwaende = Math.max(leinwaende, befund.leinwaende)
    offen = ziel

    if (nummer === 1 || nummer === 50 || nummer === besuche) speicherstaende.push({ nummer, ...speicher(vorher, anwendung) })
    if ([1, 3, 50, besuche].includes(nummer)) await p.bild(`abn01-besuch-${String(nummer)}-${befund.bild}`)
  }

  p.pruefe(`${String(besuche)} Besuche: jeder mit der erwarteten Partie, neuer Leinwand, lebendem Kontext, ihren Karten und wirksamem »Näher heran«`, true)
  p.pruefe(`${String(bilder.length)} Wechsel des Erscheinungsbilds über die Einstellungen, jeder am Wurzelelement angekommen`, bilder.length === mix.filter((_, i) => i > 0 && bildBei(i + 1)).length, bilder.join(', '))
  p.zeige(`Leinwände im Dokument am Tisch: höchstens ${String(leinwaende)}, über alle Besuche`)

  const frueh = median(zeiten.slice(1, 11))
  const spaet = median(zeiten.slice(-10))
  p.zeige(`Zeit je Besuch (ms): ${zeiten.map(String).join(', ')}`)
  p.pruefe(`die Besuche 91–100 brauchen im Median höchstens doppelt so lange wie 2–11`, spaet <= 2 * frueh, `${String(frueh)} ms → ${String(spaet)} ms`)
  p.zeige(`Speicher in MB (Anwendung · WebContent · GPU): ${speicherstaende.map((s) => `Besuch ${String(s.nummer)}: ${String(s.anwendung)} · ${String(s.web)} · ${String(s.gpu)}`).join('; ')}`)
  const danach = umgebung()
  p.zeige(`Umgebung nach dem Lauf, die Anwendung eingeschlossen: Last ${String(danach.last[0])}, fremd ${String(danach.fremd.summe)} %`)
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
