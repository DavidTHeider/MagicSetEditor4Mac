// **Leistungs- und Bedienbudgets am Fenster** — B02
// (#2221).
//
// # Der Messvertrag, vor dem Lauf festgelegt
//
// - **Start bis bedienbar**: vom Aufruf über LaunchServices (`open`) bis die
//   Leiste antwortet (`Probe.bereit`), alle 20 ms nachgefragt — **ohne** die
//   halbe Sekunde, die `bereit` danach wartet; bis zum vierten Lauf stand sie
//   in jedem Wert. Fünfmal **kalt** — je ein frisches
//   Zuhause, kein Zwischenspeicher dieses Programms — und sechsmal **warm** —
//   dasselbe Zuhause, das Programm zuvor mit ⌘Q beendet. **Kein `purge`**:
//   der Dateicache des Systems bleibt warm, das braucht ein Kennwort.
// - **Feld → dekodiertes Vorschaubild**: 24 Eingaben in das Namensfeld einer
//   Karte, je eine Taste über AppKit — abwechselnd ein Zeichen und sein
//   Löschen, damit jedes Bild sich vom vorigen unterscheidet. Gemessen **in der Seite**: vom
//   `keydown` bis das Vorschaubild eine neue Quelle hat und `decode()`
//   aufgelöst ist — mit der Entprellung, wie ein Mensch sie erlebt.
// - **Seitenwechsel**: 24 Wechsel über die Befehlspalette, in der Seite
//   gemessen vom `mousedown` auf der Zeile bis die Brotkrume eine andere Seite
//   nennt und zwei Bilder gezeichnet sind. **Über die Palette**, weil die
//   Leiste Arbeitsbereiche führt und nicht jede Seite.
// - **Rollen**: die Kartenliste von „Black" (3.788 Karten, kanonisches Set
//   aus MSE2) in 160 Schritten zu je 120 Punkten; je Bild die Zeit zwischen
//   zwei `requestAnimationFrame`.
// - **Speicher**: Go-Prozess und die WebKit-Prozesse, die nach dem Start
//   entstanden (wie `abn01tisch`), nach jedem Abschnitt.
//
// Je Reihe Median, kleinster und größter Wert, P95 erst ab 20 Werten. **Die
// Vorschläge aus dem Plan** — Start ≤ 2 s, Vorschau Median ≤ 150 ms,
// Rollbild ≤ 50 ms — werden **bestätigt oder begründet geändert**, nicht als
// Schwelle gesetzt: dies ist die Ausgangslage. **Diese Maschine, dieser
// Lauf**; Last und Kohorte stehen im Beleg.
import { execFileSync } from 'node:child_process'
import { existsSync, mkdirSync } from 'node:fs'
import { cpus, homedir, loadavg } from 'node:os'
import { join } from 'node:path'

const pack = process.env.MSE_PROBE_PACK ?? join(homedir(), 'full-magic-pack')
const kanon = process.env.MSE_PROBE_KANON ?? join(homedir(), 'Library/Application Support/MagicSetEditor4Mac/kanonische-sets')

// kennzahlen fasst eine Reihe zusammen — P95 erst ab 20 Werten.
export function kennzahlen(werte) {
  const s = [...werte].sort((a, b) => a - b)
  const q = (anteil) => s[Math.min(s.length - 1, Math.floor(anteil * s.length))]
  const rund = (z) => Math.round(z * 10) / 10

  return { n: s.length, min: rund(s[0]), median: rund(q(0.5)), ...(s.length >= 20 ? { p95: rund(q(0.95)) } : {}), max: rund(s.at(-1)) }
}

const zeile = (was, k, budget) => `| ${was} | ${k.n} | ${k.min} | ${k.median} | ${k.p95 ?? '—'} | ${k.max} | ${budget} |`

const prozesse = () => execFileSync('ps', ['-Ao', 'pid=,rss=,comm=']).toString().split('\n')
  .map((z) => z.trim().split(/\s+/)).filter((t) => t.length >= 3)
  .map(([pid, rss, ...name]) => ({ pid, rss: Number(rss), name: name.join(' ') }))

function speicher(vorher, anwendung) {
  const neue = prozesse().filter((x) => !vorher.has(x.pid))
  const mb = (liste) => Math.round(liste.reduce((s, x) => s + x.rss, 0) / 1024)

  return { anwendung: mb(prozesse().filter((x) => String(x.pid) === String(anwendung))), web: mb(neue.filter((x) => x.name.includes('WebKit.WebContent'))), gpu: mb(neue.filter((x) => x.name.includes('WebKit.GPU'))) }
}

// vorschauzeit **legt die Messung an und kehrt sofort zurück** — erst danach
// geht die Taste hinaus. Lief beides zugleich, kam die Taste manchmal vor dem
// Horcher an, und die Messung wartete auf ein `keydown`, das schon vorbei war.
// Die Antwort nennt deshalb auch, **woran** eine Messung scheiterte.
export const vorschauzeit = `window.__vorschaumessung = new Promise((fertig) => {
  const bild = () => document.querySelector('img[alt="Vorschau der Karte"]')
  const vorher = bild()?.src
  let beginn = 0
  document.addEventListener('keydown', () => { beginn = performance.now() }, { capture: true, once: true })
  const frist = setTimeout(() => fertig({ ms: -1, grund: beginn ? 'das Bild blieb' : 'kein keydown' }), 10000)
  const schau = () => {
    const b = bild()
    if (beginn && b && b.src !== vorher) {
      b.decode().then(() => { clearTimeout(frist); fertig({ ms: performance.now() - beginn }) }, () => { clearTimeout(frist); fertig({ ms: -1, grund: 'decode scheiterte' }) })
    } else requestAnimationFrame(schau)
  }
  requestAnimationFrame(schau)
})
return true`

// seitenzeit misst vom Druck auf die Palettenzeile, bis die Brotkrume eine
// andere Seite nennt und zwei Bilder gezeichnet sind — angelegt **vor** dem
// Klick, aus demselben Grund wie `vorschauzeit`.
export const seitenzeit = `window.__seitenmessung = new Promise((fertig) => {
  const vorher = document.querySelector('nav ol [aria-current=page]')?.innerText.trim() ?? ''
  let beginn = 0
  document.addEventListener('mousedown', () => { beginn = performance.now() }, { capture: true, once: true })
  const frist = setTimeout(() => fertig(-1), 10000)
  const schau = () => {
    const jetzt = document.querySelector('nav ol [aria-current=page]')?.innerText.trim() ?? ''
    if (beginn && jetzt !== vorher) {
      requestAnimationFrame(() => requestAnimationFrame(() => { clearTimeout(frist); fertig(performance.now() - beginn) }))
    } else requestAnimationFrame(schau)
  }
  requestAnimationFrame(schau)
})
return true`

export const rollzeit = `return new Promise((fertig) => {
  const liste = [...document.querySelectorAll('*')].filter((e) => e.checkVisibility() && e.scrollHeight > e.clientHeight + 2000 && /auto|scroll/.test(getComputedStyle(e).overflowY))[0]
  if (!liste) { fertig(null); return }
  const zeiten = []
  let vorher = 0
  let schritte = 0
  const schlag = (jetzt) => {
    if (vorher) zeiten.push(jetzt - vorher)
    vorher = jetzt
    liste.scrollTop += 120
    if (++schritte < 160) requestAnimationFrame(schlag)
    else fertig({ zeiten, hoehe: liste.scrollHeight })
  }
  requestAnimationFrame(schlag)
})`

// namensfeld ist das Feld »name« der aufgeschlagenen Karte; langeListe fragt,
// ob eine Liste weit über das Fenster hinausreicht.
export const namensfeld = "[...document.querySelectorAll('input, textarea, [contenteditable=true]')].find((e) => e.checkVisibility() && e.getAttribute('aria-label') === 'name')"
export const langeListe = "return [...document.querySelectorAll('*')].some((e) => e.checkVisibility() && e.scrollHeight > e.clientHeight + 2000)"

// paletteZeile ist die Zeile einer Seite in der Befehlspalette.
export const paletteZeile = (schluessel) => `document.querySelector('[data-palettenzeile][data-seite=${JSON.stringify(schluessel)}]')`
export const paletteOffen = "return [...document.querySelectorAll('[data-palettenzeile]')].some((z) => z.checkVisibility())"

export default async function budgets(p) {
  p.zeige(`Last vor dem Lauf ${loadavg().map((z) => z.toFixed(2)).join(' / ')} auf ${String(cpus().length)} Kernen`)
  const vorher = new Set(prozesse().map((x) => x.pid))
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  p.notiz('Start bis bedienbar')
  // **Fünf Kaltstarts, jeder in einem frischen Zuhause** — der Plan verlangt
  // mindestens fünf Werte je Start; danach sechs warme im selben Zuhause.
  const kalt = []
  const warm = []
  for (let i = 0; i < 5; i++) {
    const frisch = p.ort()
    p.mse(frisch, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
    p.mse(frisch, 'oeffnen', 'saat/probe-werkstatt.mse-set')
    const beginn = Date.now()
    await p.starte({ ort: frisch, takt: 20 })
    kalt.push(p.bereitUm - beginn)
    p.pruefe(`Kaltstart ${String(i + 1)}: ⌘Q beendet`, await p.beende())
  }
  for (let i = 0; i < 7; i++) {
    const beginn = Date.now()
    await p.starte({ ort, takt: 20 })
    // **Der erste Start in diesem Zuhause ist kalt** und zählt nicht zu den warmen.
    if (i > 0) warm.push(p.bereitUm - beginn)
    if (i < 6) p.pruefe(`Start ${String(i + 1)}: ⌘Q beendet`, await p.beende())
  }
  const anwendung = p.pids.at(-1)
  const nachStart = speicher(vorher, anwendung)

  p.notiz('Feld → dekodiertes Vorschaubild')
  // **Das Set über ⌘2** — wie `a02` und `zustaende`; die erste Karte ist
  // danach aufgeschlagen.
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  const feld = namensfeld
  await p.warte('das Namensfeld steht da', () => p.seite(`return !!(${feld})`), 20000)
  await p.klickeAuf(feld, 'das Namensfeld')
  const vorschau = []
  const gruende = []
  let daneben = 0
  // **Der Fokus wird vor jeder Taste geprüft** — ein Lauf verlor elf von
  // vierundzwanzig Vorschauen, weil die Taste nicht im Feld ankam; gezählt
  // wird nur, was ankam, und daneben, was nicht.
  for (let i = 0; vorschau.length < 24 && i < 40; i++) {
    if (!(await p.seite(`return document.activeElement === (${feld})`))) await p.klickeAuf(feld, 'das Namensfeld')
    const vorher = await p.seite(`return (${feld}).value ?? (${feld}).innerText`)
    await p.seite(vorschauzeit)
    // **Abwechselnd tippen und löschen** — ein Name, der länger wird, als die
    // Zeile zeigt, gibt ein gleiches Bild, und die Messung wartete auf einen
    // Wechsel, den es nicht gibt (vierter Lauf: fünf von 24 ohne Bild).
    await p.taste(i % 2 === 0 ? 'x' : 'Backspace')
    const { ms, grund } = await p.seite('return window.__vorschaumessung')
    const nachher = await p.seite(`return (${feld}).value ?? (${feld}).innerText`)
    if (nachher === vorher) { daneben++; continue }
    if (ms < 0) gruende.push(grund)
    vorschau.push(ms)
  }
  p.zeige(`${String(daneben)} Tasten kamen nicht im Feld an und zählen nicht`)
  const nachVorschau = speicher(vorher, anwendung)

  p.notiz('Seitenwechsel über die Befehlspalette')
  const ziele = ['start', 'decks', 'sammlung', 'sets', 'turniere', 'einstellungen']
  const wechsel = []
  for (let i = 0; i < 24; i++) {
    const ziel = ziele[i % ziele.length]
    await p.taste('k', 'befehl')
    await p.warte('die Palette steht da', () => p.seite(paletteOffen), 15000)
    await p.seite(`(${paletteZeile(ziel)}).scrollIntoView({ block: 'center' }); return true`)
    await p.seite(seitenzeit)
    await p.klickeAuf(paletteZeile(ziel), `»${ziel}« in der Palette`)
    wechsel.push(await p.seite('return window.__seitenmessung'))
  }
  p.pruefe('Start, Vorschau und Seitenwechsel: ⌘Q beendet', await p.beende())

  p.notiz('Rollen durch »Black« (3.788 Karten)')
  let rollen = null
  if (existsSync(join(pack, 'data')) && existsSync(join(kanon, 'Black.mse-set'))) {
    const grossOrt = p.ort()
    const anwendungsort = join(grossOrt.heim, 'Library/Application Support/MagicSetEditor4Mac/pakete')
    mkdirSync(anwendungsort, { recursive: true })
    execFileSync('cp', ['-cR', `${join(pack, 'data')}/.`, anwendungsort])
    execFileSync('cp', ['-c', join(kanon, 'Black.mse-set'), join(grossOrt.ort, 'Black.mse-set')])
    p.mse(grossOrt, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
    p.mse(grossOrt, 'oeffnen', join(grossOrt.ort, 'Black.mse-set'))
    // **Der Name aus dem Bestand** — der Titel des Sets, nicht der Dateiname.
    const setname = p.mse(grossOrt, 'sets').trim().split('\n')[0].split('\t')[1]
    p.zeige(`das große Set heißt »${setname}«`)
    await p.starte({ ort: grossOrt })
    await p.taste('2', 'befehl')
    await p.wartetext(setname, 30000)
    await p.klickeText(setname)
    await p.warte('die Kartenliste ist lang', () => p.seite(langeListe), 60000)
    rollen = await p.seite(rollzeit)
    p.pruefe('Black: ⌘Q beendet', await p.beende())
  } else {
    p.zeige(`nicht gemessen: Pack (${pack}) oder Black.mse-set (${kanon}) fehlt`)
  }

  p.notiz('Ergebnis')
  const gemessen = (reihe) => reihe.filter((z) => z >= 0)
  p.pruefe('jede Vorschau kam an', vorschau.every((z) => z >= 0), `${String(vorschau.filter((z) => z < 0).length)} ohne Bild: ${gruende.join(', ')}`)
  p.pruefe('jeder Seitenwechsel kam an', wechsel.every((z) => z >= 0))
  p.pruefe('die Kartenliste von Black wurde gerollt', rollen !== null && rollen.zeiten.length >= 20)
  p.zeilen.push('', '| Reihe | n | min | Median | P95 | max | Vorschlag des Plans |', '| --- | --- | --- | --- | --- | --- | --- |',
    zeile('Start kalt (ms)', kennzahlen(kalt), '≤ 2000'),
    zeile('Start warm (ms)', kennzahlen(warm), '≤ 2000'),
    zeile('Feld → Vorschaubild (ms)', kennzahlen(gemessen(vorschau)), 'Median ≤ 150'),
    zeile('Seitenwechsel (ms)', kennzahlen(gemessen(wechsel)), '—'),
    ...(rollen ? [zeile('Rollbild Black (ms)', kennzahlen(rollen.zeiten), '≤ 50')] : []),
    '',
    `Speicher (MB, RSS): nach dem Start ${JSON.stringify(nachStart)}, nach den Vorschauen ${JSON.stringify(nachVorschau)}.`,
    '',
    `Rohwerte — kalt ${JSON.stringify(kalt)}, warm ${JSON.stringify(warm)}, Vorschau ${JSON.stringify(vorschau.map(Math.round))}, Wechsel ${JSON.stringify(wechsel.map(Math.round))}${rollen ? `, Rollen ${JSON.stringify(rollen.zeiten.map((z) => Math.round(z * 10) / 10))}` : ''}.`)
}
