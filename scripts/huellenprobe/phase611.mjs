// **Was die WebKit-Drehbücher versprachen, am nativen Fenster** — M5, M14
// und G2 aus dem Zweitblick über Phase 6.3 (#1664), verlagert in Phase 6.11
// (#1672).
//
// - **M5, M14** — `ux14-groessen` übersprang jeden Knopf in einem Kasten mit
//   `overflow-x: auto` — und das ist jede Spalte der Arbeitsplätze, denn
//   `overflow-y: auto` macht `visible` zu `auto` (CSS Overflow 3). Es nannte
//   keine Zahl und fuhr weder „Leiste oben + Sehr groß + 1.024 × 700" noch
//   eine eigene Schrift. Hier: jede Seite der Palette in genau dieser
//   Belastung, mit der Schrift „Buch" über die ganze Oberfläche. **Ein Knopf
//   ist abgeschnitten**, wenn er über einen Vorfahren hinausragt, der
//   beschneidet und **nicht** waagerecht rollt — wer rollt, macht ihn
//   erreichbar, und das prüft `uirevue` (innere Roller mit Tastatur).
// - **G2, `ux10-darstellung`** — es versprach die berechnete Übergangsdauer
//   eines Knopfs und las ein Attribut. Die Oberfläche hat aber keine
//   Übergänge; hier: „weniger Bewegung" gesetzt und gemessen, ob die Regel
//   in WebKit einen Übergang von 200 ms tatsächlich kürzt — mit Gegenprobe.
// - **G2, `ux08-schubladen`** — „die Fläche darunter ist nicht erreichbar",
//   ohne Schritt dazu. Hier: eine Schublade offen, dreißig Mal Tab — der
//   Fokus bleibt in ihr —, und der Punkt neben ihr trifft keinen Knopf der
//   Seite darunter.
// - **N12** (Nachlauf über 6.3) — Tab aus der Spalte über dem Zentrum: der
//   Fokus steht danach auf etwas, das man sieht.
// - **G20** — die Leiste oben und die Spalte über dem Zentrum trugen beide
//   `z-10`: rollt die Seite, bleibt die Leiste obenauf?
//
// **Was es nicht prüft**: ob ein Knopf, der erreichbar ist, auch gut steht.
// Das sieht ein Mensch an den Bildern.
import { oeffnen, seitennamen, zurSeite } from './uirevue.mjs'

const eng = { name: 'eng', breite: 1024, hoehe: 700 }

// abgeschnitten nennt die Knöpfe der Fläche, die über einen beschneidenden,
// nicht waagerecht rollenden Vorfahren oder über das Fenster hinausragen.
const abgeschnitten = `
  const sichtbar = (e) => e.checkVisibility({ checkOpacity: true, checkVisibilityCSS: true })
  const flaeche = [...document.querySelectorAll('[role=dialog]')].find(sichtbar) ?? document.querySelector('main')
  const knoepfe = [...flaeche.querySelectorAll('button, a[href], [role=tab]')].filter(sichtbar)
  const befunde = []
  for (const k of knoepfe) {
    const r = k.getBoundingClientRect()
    if (r.width === 0) continue
    let raus = r.left < -1 || r.right > innerWidth + 1
    for (let v = k.parentElement; v && !raus; v = v.parentElement) {
      const o = getComputedStyle(v).overflowX
      if (o === 'visible') continue
      if (v.scrollWidth > v.clientWidth + 1 && (o === 'auto' || o === 'scroll')) break
      const rv = v.getBoundingClientRect()
      const rechts = rv.left + v.clientLeft + v.clientWidth
      if (r.right > rechts + 1 || r.left < rv.left - 1) raus = true
    }
    if (raus) befunde.push((k.getAttribute('aria-label') || k.innerText || k.tagName).trim().replace(/\\s+/g, ' ').slice(0, 40))
  }
  return { geprueft: knoepfe.length, befunde, waagerecht: document.scrollingElement.scrollWidth > innerWidth + 1 }
`

export default async function phase611(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')

  p.notiz('M5 und M14 — Leiste oben, Sehr groß, Buch über die ganze Oberfläche, 1.024 × 700')
  p.mse(ort, 'einstellungen', 'setzen', 'leistenlage', 'waagerecht')
  p.mse(ort, 'einstellungen', 'setzen', 'schriftgroesse', 'sehrgross')
  p.mse(ort, 'einstellungen', 'setzen', 'schriftart', 'buch')
  p.mse(ort, 'einstellungen', 'setzen', 'schriftbereich', 'ganz')
  await oeffnen(p, ort, eng)

  const namen = await seitennamen(p)
  let gesamt = 0
  const alleBefunde = []
  for (const name of namen) {
    try {
      await zurSeite(p, name)
      const wert = await p.seite(abgeschnitten)
      gesamt += wert.geprueft
      if (wert.befunde.length > 0 || wert.waagerecht) {
        alleBefunde.push(`${name}: ${wert.waagerecht ? 'rollt waagerecht; ' : ''}${wert.befunde.join(' · ')}`)
        await p.bild(`groessen-${name.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`)
      }
    } catch (grund) {
      alleBefunde.push(`${name}: nicht erreicht — ${String(grund.message ?? grund)}`)
      await p.taste('Escape')
    }
  }
  p.zeige(`${String(namen.length)} Seiten, ${String(gesamt)} Bedienelemente geprüft`)
  for (const befund of alleBefunde) p.zeige(befund)
  p.pruefe('kein Knopf abgeschnitten, keine Seite rollt waagerecht', alleBefunde.length === 0, String(alleBefunde.length))

  p.notiz('N12 — Tab aus der Spalte über dem Zentrum; G20 — die Leiste bleibt obenauf')
  await zurSeite(p, 'Karten')
  await p.klickeAuf("[...document.querySelectorAll('button')].find((b) => b.innerText.trim() === 'Navigator einblenden' || b.getAttribute('aria-label') === 'Navigator einblenden')", '»Navigator einblenden«')
  await p.warte('die Spalte liegt über dem Zentrum', () => p.seite("return !!document.querySelector('[data-ueber=zentrum]')"), 10000)
  const leisteOben = await p.seite(`
    const spalte = document.querySelector('[data-ueber=zentrum]')
    for (let v = spalte.parentElement; v; v = v.parentElement) { if (v.scrollHeight > v.clientHeight + 1 && getComputedStyle(v).overflowY !== 'visible') { v.scrollTop = v.scrollHeight; break } }
    document.scrollingElement.scrollTop = document.scrollingElement.scrollHeight
    const leiste = document.getElementById('seitenleiste').getBoundingClientRect()
    const s = spalte.getBoundingClientRect()
    const x = Math.max(s.left + 10, 1), y = leiste.top + leiste.height / 2
    const trifft = document.elementFromPoint(x, y)
    return { ueberlappt: s.top < leiste.bottom, inLeiste: !!trifft?.closest('#seitenleiste'), was: (trifft?.innerText || trifft?.tagName || '').trim().slice(0, 30) }
  `)
  p.zeige(`Spalte reicht unter die Leiste: ${String(leisteOben.ueberlappt)}; am Punkt: ${leisteOben.was}`)
  p.pruefe('die Leiste oben bleibt über der Spalte', !leisteOben.ueberlappt || leisteOben.inLeiste, leisteOben.was)
  await p.seite("document.querySelector('[data-ueber=zentrum] button, [data-ueber=zentrum] a[href], [data-ueber=zentrum] [tabindex]')?.focus(); return true")
  // **Gewartet wird auf die Wirkung** (§3): WebKit rollt das fokussierte
  // Element einen Augenblick nach dem Tab ins Bild; der zweite Lauf am
  // 2026-10-04 maß davor und sah es noch unter der Leiste, das Bild danach
  // zeigte es frei. Verdeckt heißt deshalb: nach zwei Sekunden noch.
  const fokusstand = `
      const a = document.activeElement
      if (!a || a === document.body) return 'nirgends'
      if (a.closest('[data-ueber=zentrum]')) return 'drin'
      const r = a.getBoundingClientRect()
      const e = document.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2)
      return e && (a.contains(e) || e.contains(a)) ? 'sichtbar' : 'verdeckt: ' + (a.getAttribute('aria-label') || a.innerText || a.tagName).trim().slice(0, 30)
  `
  let verborgen = ''
  for (let schritt = 0; schritt < 40; schritt++) {
    await p.taste('Tab')
    const stand = await p.seite(fokusstand)
    if (stand === 'drin') continue
    verborgen = stand
    for (let warten = 0; warten < 20 && verborgen.startsWith('verdeckt'); warten++) {
      await new Promise((los) => setTimeout(los, 100))
      verborgen = await p.seite(fokusstand)
    }
    break
  }
  p.zeige(`nach dem Verlassen der Spalte: ${verborgen}`)
  p.pruefe('der Fokus steht danach auf etwas Sichtbarem', verborgen === 'sichtbar' || verborgen === 'nirgends', verborgen)
  await p.bild('n12-nach-tab')
  p.pruefe('⌘Q beendet die Anwendung (Belastung)', await p.beende())

  // Zurück auf die Vorgaben, damit die Teile danach nicht die Belastung messen.
  p.mse(ort, 'einstellungen', 'setzen', 'leistenlage', 'senkrecht')
  p.mse(ort, 'einstellungen', 'setzen', 'schriftgroesse', 'normal')
  p.mse(ort, 'einstellungen', 'setzen', 'schriftbereich', 'ueberschriften')

  p.notiz('G2 (ux10-darstellung) — weniger Bewegung, gemessen an der berechneten Dauer')
  // **Die Oberfläche hat keine Übergänge** — Schublade, Dialog und Popover
  // stehen auf Dauer 0, Übergangsklassen gibt es nicht; gemessen am
  // 2026-10-04 an 201 sichtbaren Elementen, alle ohne. Was sich bewegt, ist
  // der Tisch (`tisch/bewegung.test.ts`). **Gemessen wird deshalb, ob die
  // Regel greift**: ein eingesetztes Element mit 200 ms Übergang, einmal
  // ohne und einmal mit der Einstellung — die Gegenprobe steckt im ersten.
  const probe = `
    const e = document.createElement('div')
    e.style.transition = 'opacity 200ms linear'
    document.body.append(e)
    const dauer = getComputedStyle(e).transitionDuration
    e.remove()
    return { dauer, attribut: document.documentElement.dataset.bewegung ?? '' }
  `
  await oeffnen(p, ort, { name: 'normal', breite: 1280, hoehe: 800 })
  const vorher = await p.seite(probe)
  p.zeige(`wie im System: ein Element mit 200 ms Übergang misst ${vorher.dauer}`)
  p.pruefe('Gegenprobe: ohne die Einstellung bleibt der Übergang', vorher.dauer === '0.2s', vorher.dauer)
  p.pruefe('⌘Q beendet die Anwendung (wie im System)', await p.beende())

  // **Eine Nutzereinstellung**, keine des Geräts — Persönliches unter
  // *Nutzer und Profile*; `einstellungen setzen` kennt sie nicht.
  const nutzer = JSON.parse(p.mse(ort, 'nutzer', '--json')).nutzer[0].id
  p.mse(ort, 'nutzereinstellung', nutzer, 'barrierefreiheit.bewegung', 'reduziert')
  await oeffnen(p, ort, { name: 'normal', breite: 1280, hoehe: 800 })
  const dauer = await p.seite(probe)
  p.zeige(`reduziert (Attribut »${dauer.attribut}«): dasselbe Element misst ${dauer.dauer}`)
  p.pruefe('die Einstellung kommt am Wurzelelement an', dauer.attribut === 'reduziert', dauer.attribut)
  p.pruefe('der Übergang schrumpft auf höchstens 10 ms', parseFloat(dauer.dauer) * (dauer.dauer.endsWith('ms') ? 1 : 1000) <= 10, dauer.dauer)

  p.notiz('G2 (ux08-schubladen) — die Fläche unter einer Schublade ist nicht erreichbar')
  // **Die Schublade des Sets**, wie in `ux08-schubladen` — die des
  // Turnierassistenten öffnet in einem frischen Bestand anders.
  await zurSeite(p, 'Karten')
  await p.klickeAuf("[...document.querySelectorAll('button')].find((b) => b.innerText.trim() === 'Eigenschaften')", '»Eigenschaften«')
  await p.warte('die Schublade steht', () => p.seite("return !!document.querySelector('[role=dialog]')"), 20000)
  await new Promise((los) => setTimeout(los, 500))
  const draussen = []
  for (let schritt = 0; schritt < 30; schritt++) {
    await p.taste('Tab')
    const stelle = await p.seite("const d = document.querySelector('[role=dialog]'); return d && d.contains(document.activeElement) ? '' : (document.activeElement?.innerText || document.activeElement?.tagName || '').trim().slice(0, 30)")
    if (stelle !== '') draussen.push(stelle)
  }
  p.pruefe('dreißig Mal Tab — der Fokus bleibt in der Schublade', draussen.length === 0, draussen.join(' · '))
  const getroffen = await p.seite(`
    const d = document.querySelector('[role=dialog]').getBoundingClientRect()
    const x = Math.max(5, d.left - 40), y = d.top + 100
    const e = document.elementFromPoint(x, y)
    const knopf = e?.closest('button, a[href], input, select, textarea')
    return knopf && !document.querySelector('[role=dialog]').contains(knopf) ? (knopf.innerText || knopf.tagName).trim().slice(0, 30) : ''
  `)
  p.pruefe('neben der Schublade liegt kein Knopf der Seite darunter', getroffen === '', getroffen)
  await p.bild('schublade-darunter')
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
  p.mse(ort, 'nutzereinstellung', nutzer, 'barrierefreiheit.bewegung', 'system')
}
