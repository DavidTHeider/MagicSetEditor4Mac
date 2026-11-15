// **Warum das erste Rollen durch Black langsam ist** — E07
// (#2279).
//
// B02 fand den P95 des Rollbilds über 50 ms und die langsamen Bilder am
// Anfang; **vermutet** waren Zeilen, die zum ersten Mal ins Fenster der
// Liste kommen. Dieses Drehbuch trennt drei Lagen am selben Bündel:
//
//   - **A — sofort**: wie B02, gleich nachdem die Liste lang ist,
//   - **B — noch einmal**: zurück nach oben, zwei Sekunden Ruhe, dieselbe
//     Strecke — die Zeilen und ihre Bildchen sind schon da,
//   - **C — nach Ruhe**: die Seite neu geladen, das Set geöffnet, fünf
//     Sekunden gewartet, dann gerollt — die Zeilen sind neu, die Anwendung
//     ist still (aber Go hält die Bildchen aus A schon im Speicher),
//   - **D — ohne Bildchen**: ein frischer Prozess wie A, aber `Kartenbilder`
//     scheitert sofort (`stoere`, B03) — es wird keines gerendert.
//
// Ist A langsam und B wie C schnell, liegt es an dem, was das Öffnen
// nacharbeitet; ist C so langsam wie A, an den neuen Zeilen selbst. Dazu
// zählt das Rufbuch, welche Bindungen während A liefen.
//
// Gemessen wird wie in `budgets.mjs`: `scrollTop` um 120 je Bild, 160 Bilder,
// die Zeit zwischen zwei `requestAnimationFrame`.
import { execFileSync } from 'node:child_process'
import { existsSync, mkdirSync } from 'node:fs'
import { homedir } from 'node:os'
import { join } from 'node:path'

import { rufbuchHolen, stoere } from '../huellenrufbuch.mjs'
import { kennzahlen, langeListe, rollzeit } from './budgets.mjs'

const pack = process.env.MSE_PROBE_PACK ?? join(homedir(), 'full-magic-pack')
const kanon = process.env.MSE_PROBE_KANON ?? join(homedir(), 'Library/Application Support/MagicSetEditor4Mac/kanonische-sets')
const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))

// mitschrift rollt wie `rollzeit` und hält für jedes Bild über 40 ms fest,
// welche Teile der Seite sich darin geändert haben — die nächste Fassung
// eines `data-*`-Ahnen oder seine Rolle, je Änderung gezählt.
const mitschrift = `return new Promise((fertig) => {
  const liste = [...document.querySelectorAll('*')].find((e) => e.checkVisibility() && e.scrollHeight > e.clientHeight + 2000 && /auto|scroll/.test(getComputedStyle(e).overflowY))
  if (!liste) { fertig(null); return }
  const name = (k) => { let e = k.nodeType === 1 ? k : k.parentElement; while (e && e !== document.body) { const d = [...e.attributes].find((a) => a.name.startsWith('data-') && a.name !== 'data-karte'); if (d) return d.name + (d.value.length < 20 ? '=' + d.value : ''); if (e.getAttribute('role')) return 'role=' + e.getAttribute('role'); e = e.parentElement } return 'body' }
  let jetzt = {}
  const beob = new MutationObserver((liste2) => { for (const m of liste2) { const n = name(m.target); jetzt[n] = (jetzt[n] ?? 0) + 1 } })
  beob.observe(document.body, { subtree: true, childList: true, attributes: true, characterData: true })
  const lang = []
  let vorher = 0, schritte = 0, beginn = performance.now()
  const schlag = (t) => {
    if (vorher && t - vorher > 40) lang.push({ bei: Math.round(vorher - beginn), ms: Math.round(t - vorher), aenderungen: Object.entries(jetzt).sort((a, b) => b[1] - a[1]).slice(0, 4).map(([k, v]) => k + ' ' + v).join(', ') })
    jetzt = {}
    vorher = t
    liste.scrollTop += 120
    if (++schritte < 160) requestAnimationFrame(schlag)
    else { beob.disconnect(); fertig(lang) }
  }
  requestAnimationFrame(schlag)
})`
const nachOben = "const l = [...document.querySelectorAll('*')].find((e) => e.checkVisibility() && e.scrollHeight > e.clientHeight + 2000 && /auto|scroll/.test(getComputedStyle(e).overflowY)); if (l) l.scrollTop = 0; return !!l"

// zeitraum nennt, wie viele Rufe und wie viel Zeit je Bindung zwischen zwei
// Abzügen des Rufbuchs lagen.
function zeitraum(vor, nach) {
  return Object.entries(nach)
    .map(([name, e]) => ({ name, anzahl: e.anzahl - (vor[name]?.anzahl ?? 0), ms: Math.round(e.dauerMs - (vor[name]?.dauerMs ?? 0)) }))
    .filter((e) => e.anzahl > 0)
    .sort((a, b) => b.ms - a.ms)
}

async function setOeffnen(p, setname) {
  await p.taste('2', 'befehl')
  await p.wartetext(setname, 30000)
  await p.klickeText(setname)
  await p.warte('die Kartenliste ist lang', () => p.seite(langeListe), 60000)
}

export default async function e07rollen(p) {
  if (!existsSync(join(pack, 'data')) || !existsSync(join(kanon, 'Black.mse-set'))) {
    p.pruefe(`Pack (${pack}) und Black.mse-set (${kanon}) liegen da`, false)
    return
  }

  const ort = p.ort()
  const anwendungsort = join(ort.heim, 'Library/Application Support/MagicSetEditor4Mac/pakete')
  mkdirSync(anwendungsort, { recursive: true })
  execFileSync('cp', ['-cR', `${join(pack, 'data')}/.`, anwendungsort])
  execFileSync('cp', ['-c', join(kanon, 'Black.mse-set'), join(ort.ort, 'Black.mse-set')])
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', join(ort.ort, 'Black.mse-set'))
  const setname = p.mse(ort, 'sets').trim().split('\n')[0].split('\t')[1]

  const reihen = { A: [], B: [], C: [], D: [] }
  const rufe = []
  const lange = []
  const runden = Number(process.env.MSE_E07_RUNDEN ?? 3)

  for (let runde = 1; runde <= runden; runde++) {
    p.notiz(`Runde ${String(runde)}`)
    await p.starte({ ort })
    await setOeffnen(p, setname)

    // A — sofort, und was dabei gerufen wurde.
    const vorA = structuredClone(await rufbuchHolen(p))
    const a = await p.seite(rollzeit)
    const nachA = structuredClone(await rufbuchHolen(p))
    rufe.push(zeitraum(vorA, nachA))

    // B — dieselbe Strecke noch einmal.
    await p.seite(nachOben)
    await schlafe(2000)
    const b = await p.seite(rollzeit)

    // C — neu geladen, geöffnet, fünf Sekunden Ruhe.
    await p.seite('setTimeout(() => location.reload(), 30); return true')
    await schlafe(1500)
    await p.warte('die Leiste steht wieder', () => p.seite("return document.readyState === 'complete' && document.querySelectorAll('nav button').length > 3"), 30000)
    await setOeffnen(p, setname)
    await schlafe(5000)
    const c = await p.seite(rollzeit)

    p.pruefe(`Runde ${String(runde)}: ⌘Q beendet`, await p.beende())

    // E — ein frischer Prozess wie A, mit Mitschrift der langen Bilder.
    if (runde === 1) {
      await p.starte({ ort })
      await setOeffnen(p, setname)
      lange.push(...(await p.seite(mitschrift)))
      p.pruefe('mit Mitschrift: ⌘Q beendet', await p.beende())
    }

    // D — ein frischer Prozess, und kein Bildchen wird gerendert.
    await p.starte({ ort })
    await stoere(p, [{ bindung: 'Kartenbilder', fehler: true }])
    await setOeffnen(p, setname)
    const d = await p.seite(rollzeit)
    await stoere(p, [])
    p.pruefe(`Runde ${String(runde)}, ohne Bildchen: ⌘Q beendet`, await p.beende())

    for (const [name, lauf] of [['A', a], ['B', b], ['C', c], ['D', d]]) {
      if (lauf) reihen[name].push(lauf.zeiten)
    }
  }

  p.notiz('Die drei Lagen')
  const mittel = (xs) => Math.round((xs.reduce((s, x) => s + x, 0) / Math.max(1, xs.length)) * 10) / 10
  p.zeilen.push('| Lage | Bilder | Median | P95 | max | Mittel der ersten 40 | Mittel danach |', '| --- | ---: | ---: | ---: | ---: | ---: | ---: |')
  for (const [name, laeufe] of Object.entries(reihen)) {
    const alle = laeufe.flat()
    const k = kennzahlen(alle)
    p.zeilen.push(`| ${name} | ${String(k.n)} | ${String(k.median)} | ${String(k.p95 ?? '—')} | ${String(k.max)} | ${String(mittel(laeufe.flatMap((z) => z.slice(0, 40))))} | ${String(mittel(laeufe.flatMap((z) => z.slice(40))))} |`)
  }
  p.zeilen.push('', `Rufe während A, je Runde: ${rufe.map((r) => r.slice(0, 6).map((e) => `${e.name} ${String(e.anzahl)}× ${String(e.ms)} ms`).join(', ') || '—').join(' · ')}`)
  // **Je Runde** — E07 verlangt den P95 in fünf aufeinanderfolgenden Läufen,
  // nicht über alle zusammen.
  p.zeilen.push('', `P95 von A je Runde: ${reihen.A.map((z) => String(kennzahlen(z).p95)).join(', ')} ms; Höchstwert je Runde: ${reihen.A.map((z) => String(kennzahlen(z).max)).join(', ')} ms`)
  p.pruefe('A: P95 ≤ 50 ms in jeder Runde', reihen.A.every((z) => kennzahlen(z).p95 <= 50), reihen.A.map((z) => String(kennzahlen(z).p95)).join(', '))
  p.zeilen.push('', '**Die langen Bilder eines frischen Starts, mit dem, was sich darin änderte**:', '', ...lange.map((l) => `- bei ${String(l.bei)} ms: ${String(l.ms)} ms — ${l.aenderungen || 'keine Änderung im DOM'}`))
  p.zeilen.push('', `Rohwerte — ${Object.entries(reihen).map(([n, l]) => `${n}: ${JSON.stringify(l.map((z) => z.map((x) => Math.round(x))))}`).join('; ')}`)
  p.pruefe('jede Lage wurde in jeder Runde gerollt', Object.values(reihen).every((l) => l.length === runden))
}
