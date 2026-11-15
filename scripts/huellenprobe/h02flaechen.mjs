// **Bewegung, Kontrast und das kleinste Fenster** — H02
// (#2256),
// am nativen Fenster, in drei Teilen:
//
//   1. **Bewegung reduziert** (`barrierefreiheit.bewegung = reduziert`): die
//      Wurzel trägt `data-bewegung`, und Palette, Schublade und Dialog gehen
//      ohne Übergang und ohne Animation auf,
//   2. **Kontrast hoch** (`barrierefreiheit.kontrast = hoch`): jeder
//      sichtbare Text jeder Seite der Anwendung hält **7 : 1** — die Zusage
//      der Einstellung (`internal/domain/nutzereinstellung.go`),
//   3. **das kleinste Fenster**: im zugesagten Mindestmaß 1024 × 700
//      (`internal/fenster/bildschirm.go`) rollt keine Seite waagerecht, und
//      eine Rückfrage steht mit ihren Knöpfen ganz im Bild; **darunter** —
//      der Gast im Audit hatte 1024 × 558 — wird nur verlangt, dass jedes
//      Ziel durch Rollen erreichbar bleibt.
//
// **Die Systemeinstellungen setzt ein Mensch** (H02, „Tests ändern sie nicht
// still"): gelesen wird nur, ob `prefers-reduced-motion` und
// `prefers-contrast` gerade gelten. Gilt eine nicht, steht das als offener
// Schritt für einen berechtigten Menschen da — die Einstellungen der
// Anwendung übersteuern das System in dieselbe Richtung, und **die** prüft
// dieser Lauf.
import { texteMessen } from '../huellenkontrast.mjs'
import { paletteOffen } from './budgets.mjs'
import { seitenliste } from './zustaende.mjs'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

// ohneBewegung sagt, ob ein Element ohne Übergang und Animation steht —
// **0,01 ms zählt als ohne**: die Regel in `barrierefreiheit.css` setzt mit
// Absicht diese Dauer statt null, damit `transitionend` weiter feuert.
const ohneBewegung = (ausdruck) => `const e = ${ausdruck}; if (!e) return null; const s = getComputedStyle(e); const ms = (t) => t.trim().endsWith('ms') ? parseFloat(t) : parseFloat(t) * 1000; const dauer = (w) => w.split(',').every((t) => ms(t) <= 0.01); return { uebergang: dauer(s.transitionDuration), animation: dauer(s.animationDuration) || s.animationName === 'none', werte: s.transitionDuration + ' / ' + s.animationDuration }`

async function zurSeite(p, schluessel) {
  await p.taste('k', 'befehl')
  await p.warte('die Befehlspalette steht da', () => p.seite(paletteOffen), 15000)
  const zeile = `document.querySelector('[data-palettenzeile][data-seite=${JSON.stringify(schluessel)}]')`
  if (!(await p.seite(`return !!(${zeile})`))) {
    await p.taste('Escape')

    return false
  }
  await p.klickeAuf(zeile, `»${schluessel}« in der Palette`)
  await p.warte(`die Seite ${schluessel} steht da`, () => p.seite(`return document.querySelector('main')?.dataset.seite === ${JSON.stringify(schluessel)}`), 15000)

  return true
}

export default async function h02flaechen(p) {
  const { anwendung } = seitenliste(readFileSync(resolve('frontend/src/navigation.ts'), 'utf8'))
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  // **Nicht maximiert**: ein maximiertes Fenster nimmt keine Größe an —
  // dieselbe Einstellung wie in `bk07a.mjs`.
  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'normal')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'nutzereinstellung', '1', 'barrierefreiheit.bewegung', 'reduziert')
  p.mse(ort, 'nutzereinstellung', '1', 'barrierefreiheit.kontrast', 'hoch')
  await p.starte({ ort })

  p.notiz('0 — was das System gerade sagt (gesetzt von einem Menschen, hier nur gelesen)')
  const system = await p.seite("return { bewegung: matchMedia('(prefers-reduced-motion: reduce)').matches, kontrast: matchMedia('(prefers-contrast: more)').matches }")
  p.zeige(`System: reduzierte Bewegung ${system.bewegung ? 'an' : 'aus'}, erhöhter Kontrast ${system.kontrast ? 'an' : 'aus'}${system.bewegung && system.kontrast ? '' : ' — der Lauf mit den Systemeinstellungen bleibt ein Schritt für einen berechtigten Menschen'}`)

  p.notiz('1 — Bewegung reduziert')
  p.pruefe('die Wurzel trägt data-bewegung=reduziert', await p.seite("return document.documentElement.dataset.bewegung === 'reduziert'"))
  await p.taste('k', 'befehl')
  await p.warte('die Befehlspalette steht da', () => p.seite(paletteOffen), 15000)
  const palette = await p.seite(ohneBewegung("document.querySelector('[role=dialog]')"))
  p.pruefe('die Palette geht ohne Übergang und Animation auf', palette?.uebergang === true && palette?.animation === true, JSON.stringify(palette))
  await p.taste('Escape')
  await zurSeite(p, 'pakete')
  await p.klickeText('Paket beziehen …')
  await p.warte('die Schublade steht da', () => p.seite("return !!document.querySelector('[role=dialog]')"), 15000)
  const schublade = await p.seite(ohneBewegung("document.querySelector('[role=dialog]')"))
  p.pruefe('die Schublade geht ohne Übergang und Animation auf', schublade?.uebergang === true && schublade?.animation === true, JSON.stringify(schublade))
  await p.bild('h02-schublade')
  await p.taste('Escape')

  p.notiz('2 — Kontrast hoch: jeder Text jeder Seite hält 7 : 1')
  p.pruefe('die Wurzel trägt data-kontrast=hoch', await p.seite("return document.documentElement.dataset.kontrast === 'hoch'"))
  for (const seite of anwendung) {
    if (!(await zurSeite(p, seite))) {
      p.pruefe(`${seite}: die Seite kam`, false)

      continue
    }
    await p.ruhe()
    const messungen = await texteMessen(p, 'main, nav', { klein: 7, gross: 7 })
    const unter = messungen.filter((m) => m.fehlt || m.verhaeltnis < 7)
    p.pruefe(`${seite}: ${String(messungen.length)} Texte halten 7 : 1`, messungen.length > 0 && unter.length === 0, unter.slice(0, 4).map((m) => `${m.name} ${String(m.verhaeltnis)}`).join(' · ') || `kleinstes ${String(Math.min(...messungen.map((m) => m.verhaeltnis)))}`)
  }
  await p.bild('h02-kontrast')

  p.notiz('3 — das kleinste zugesagte Fenster, 1024 × 700')
  const vorher = await p.seite('return [window.innerWidth, window.innerHeight]')
  await p.groesse(1024, 700)
  await new Promise((weiter) => { setTimeout(weiter, 1500) })
  const nachher = await p.seite('return [window.innerWidth, window.innerHeight]')
  p.zeige(`Fenster innen vorher ${vorher.join(' × ')}, nach set_size 1024 × 700: ${nachher.join(' × ')}`)
  p.pruefe('das Fenster hat das Mindestmaß', Math.abs(nachher[0] - 1024) <= 2, nachher.join(' × '))
  for (const seite of anwendung) {
    if (!(await zurSeite(p, seite))) {
      continue
    }
    await p.ruhe()
    const waagerecht = await p.seite('return document.scrollingElement.scrollWidth > window.innerWidth + 1')
    p.pruefe(`${seite}: rollt nicht waagerecht`, !waagerecht)
  }

  // **Eine Rückfrage ganz im Bild** — die Karte löschen, im Saatset.
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await p.warte('»Diese Karte« steht da', () => p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.includes('Diese Karte'))"), 20000)
  await p.klickeText('Diese Karte')
  await p.klickeAuf("[...document.querySelectorAll('[role=menuitem]')].find((e) => e.checkVisibility() && e.innerText.includes('Karte löschen'))", '»Karte löschen«')
  // **Die Rückfrage steht in der Karte**, nicht als Dialog — gesucht wird
  // sie an ihren beiden Knöpfen.
  const knoepfe = "[...document.querySelectorAll('main button')].filter((b) => b.checkVisibility() && (b.innerText.trim().endsWith('Löschen') || b.innerText.trim() === 'Abbrechen'))"
  await p.warte('die Rückfrage steht da', () => p.seite(`return ${knoepfe}.length === 2`), 10000)
  const rueckfrage = await p.seite(`return ${knoepfe}.map((b) => { const r = b.getBoundingClientRect(); return { text: b.innerText.trim(), drin: r.top >= 0 && r.left >= 0 && r.bottom <= innerHeight && r.right <= innerWidth } })`)
  p.pruefe('jeder Knopf der Rückfrage steht ganz im Bild', rueckfrage.length > 0 && rueckfrage.every((k) => k.drin), JSON.stringify(rueckfrage))
  await p.bild('h02-rueckfrage-1024x700')
  await p.taste('Escape')

  p.notiz('4 — darunter: 1024 × 558, wie der Gast im Audit')
  await p.groesse(1024, 558)
  const hoehe = await p.seite('return window.innerHeight')
  p.zeige(`innere Höhe ${String(hoehe)} — das Mindestmaß hält ${hoehe >= 700 - 60 ? 'das Fenster' : 'es nicht, und die Seite muss rollen'}`)
  for (const seite of anwendung) {
    if (!(await zurSeite(p, seite))) {
      continue
    }
    await p.ruhe()
    const unerreichbar = await p.seite("const main = document.querySelector('main'); return [...main.querySelectorAll('button, a[href], input, select, textarea')].filter((e) => e.checkVisibility()).filter((e) => { e.scrollIntoView({ block: 'nearest' }); const r = e.getBoundingClientRect(); return r.bottom > innerHeight + 1 || r.top < -1 }).map((e) => (e.innerText || e.getAttribute('aria-label') || e.tagName).trim().slice(0, 30))")
    p.pruefe(`${seite}: jedes Ziel ist durch Rollen erreichbar`, unerreichbar.length === 0, unerreichbar.slice(0, 5).join(' · '))
  }
  await p.bild('h02-1024x558')

  p.pruefe('⌘Q beendet', await p.beende())
}
