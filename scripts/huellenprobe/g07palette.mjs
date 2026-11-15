// **Die Befehlspalette am nativen Fenster** — G07
// (#2254):
//
//   1. wer schreiben darf: unter „Tun" steht „Karte erstellen"; „Sets"
//      tippen zeigt die oberste Zeile vorgewählt — vom ersten Augenblick an
//      und ohne Sprung, wenn die Daten nachkommen (#2529) —, und ⏎ im Feld
//      — ohne ↓ — öffnet die Seite,
//   2. Escape schließt die Palette,
//   3. als Leser: unter „Tun" fehlen Karte erstellen, Deck spielen und
//      Tauschen, und unter „Befehle" steht kein schreibender.
//
// **Der Weg vorher** steht im Issue: ⏎ im Feld tat nichts, und der Leser
// bekam alle Handlungen angeboten. Der Leserwechsel ist derselbe wie in
// `g10menue.mjs`.
//
// **Was es nicht hält**: ob die Vorwahl springt, wenn die Sets langsamer
// kommen als die Tasten (#2529). Am 2026-10-20 kamen sie nativ vor der
// vierten Taste, und die Gegenprobe gegen den Stand davor blieb grün — bei
// anderem Bauabdruck. Den Fall hält `frontend/src/rahmen/palettevorwahl.test.tsx`,
// das die Antwort zurückhält.
import { paletteOffen } from './budgets.mjs'

const zeilen = "return [...document.querySelectorAll('[data-palettenzeile]')].filter((z) => z.checkVisibility()).map((z) => z.innerText.replace(/\\s+/g, ' ').trim())"
const vorwahl = "return document.querySelector('[data-vorwahl]')?.innerText.replace(/\\s+/g, ' ').trim() ?? null"
const feld = "document.querySelector('[role=dialog] input[type=text]')"

async function neuLaden(p) {
  await p.seite('setTimeout(() => location.reload(), 30); return true')
  await new Promise((weiter) => { setTimeout(weiter, 1500) })
  await p.warte('die Leiste steht wieder', () => p.seite("return document.readyState === 'complete' && document.querySelectorAll('nav button').length > 3"), 30000)
}

async function oeffnen(p) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht da', () => p.seite(paletteOffen), 15000)
}

async function setOeffnen(p) {
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await p.warte('das Set ist offen', () => p.seite("return document.querySelector('main')?.dataset.seite === 'karten'"), 20000)
}

export default async function g07palette(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Palettendeck', 'probe')
  p.mse(ort, 'nutzer-neu', 'Zweite')
  p.mse(ort, 'freigabe-setzen', '1', '2', 'leser')
  await p.starte({ ort })

  p.notiz('1 — wer schreiben darf: Tun, Vorwahl und ⏎ im Feld')
  await setOeffnen(p)
  await oeffnen(p)
  const tun = await p.seite(zeilen)
  p.pruefe('unter »Tun« steht »Karte erstellen«', tun.some((z) => z.startsWith('Karte erstellen')), tun.join(' | '))
  await p.bild('g07palette-tun')
  await p.tippe('Sets')
  // **Der erste Augenblick zählt, und er muss bleiben** (#2529): bis zum
  // 2026-10-20 bot „Tun" ohne geladene Sets „Ein Set anlegen oder öffnen" an,
  // das auf „Sets" passt — wer schnell ⏎ drückte, öffnete etwas anderes als
  // das, was eine Sekunde später vorgewählt stand. Seither entscheidet das
  // offene Set die erste Handlung, und die Vorwahl springt nicht mehr.
  const beide = `${vorwahl.replace(/^return /, 'const v = ')}; ${zeilen.replace(/^return /, 'const z = ')}; return [v, z[0] ?? null]`
  // **Erst wenn das Feld „Sets" hält**: die Tasten kommen einzeln an, und
  // bei „S" ist „Schnell spielen" zu Recht oben (erster Lauf am 2026-10-20).
  const [erste, oben] = await p.warte('die Vorwahl steht zu „Sets"', async () => {
    const [eingabe, ...paar] = await p.seite(`const e = ${feld}?.value ?? null; ${beide.replace(/return \[/, 'return [e, ')}`)
    return eingabe !== 'Sets' || paar[0] === null ? null : paar
  }, 10000)
  p.pruefe('die oberste Zeile ist vorgewählt', erste === oben, `${String(erste)} | ${String(oben)}`)
  await new Promise((weiter) => { setTimeout(weiter, 1500) })
  const [spaeter] = await p.seite(beide)
  p.pruefe('die Vorwahl bleibt, wenn die Daten nachkommen', spaeter === erste, `${String(erste)} → ${String(spaeter)}`)
  p.pruefe('der Fokus steht im Feld', await p.seite(`return document.activeElement === ${feld}`))
  await p.bild('g07palette-vorwahl')
  await p.taste('Enter')
  await p.warte('die Palette ist zu', async () => !(await p.seite(paletteOffen)), 10000)
  p.pruefe('⏎ hat die Seite »Sets« geöffnet', (await p.seite("return document.querySelector('main')?.dataset.seite ?? ''")) === 'sets')

  p.notiz('2 — Escape schließt')
  await oeffnen(p)
  await p.taste('Escape')
  await p.warte('die Palette ist zu', async () => !(await p.seite(paletteOffen)), 10000)
  p.pruefe('Escape hat die Palette geschlossen', !(await p.seite(paletteOffen)))

  p.notiz('3 — als Leser')
  p.zeige(`Nutzerwechsel: ${JSON.stringify(await p.seite("try { const a = await window.go.app.Bindungen.NutzerWechseln('2', ''); return a.fehler ?? 'gewechselt' } catch (e) { return 'abgewiesen: ' + String(e.message ?? e).slice(0, 200) }"))}`)
  await neuLaden(p)
  p.zeige(`Profilwechsel: ${JSON.stringify(await p.seite("try { const a = await window.go.app.Bindungen.ProfilWechseln('1', ''); return a.fehler ?? 'gewechselt' } catch (e) { return 'abgewiesen: ' + String(e.message ?? e).slice(0, 200) }"))}`)
  await neuLaden(p)
  await setOeffnen(p)
  await oeffnen(p)
  const lesertun = await p.seite(zeilen)
  for (const verboten of ['Karte erstellen', 'Deck spielen', 'Karten tauschen']) {
    p.pruefe(`»${verboten}« fehlt`, !lesertun.some((z) => z.startsWith(verboten)), lesertun.join(' | '))
  }
  p.pruefe('die Sicherung bleibt', lesertun.some((z) => z.includes('Sicherung')), lesertun.join(' | '))
  await p.bild('g07palette-leser')

  // **Die schreibenden Befehle aus dem Katalog** — die Palette zeigt nur ihren
  // Namen; welcher schreibt, sagt der Katalog selbst.
  const schreibende = await p.seite("const a = await window.go.app.Bindungen.Befehlskatalog(); return a.befehle.filter((b) => b.schreibend).map((b) => b.name)")
  await p.tippe('set')
  await p.warte('Befehle stehen da', () => p.seite("return [...document.querySelectorAll('[data-palettenzeile] code')].some((c) => c.checkVisibility())"), 10000)
  const gezeigt = await p.seite("return [...document.querySelectorAll('[data-palettenzeile] code')].filter((c) => c.checkVisibility()).map((c) => c.innerText.trim())")
  p.pruefe('unter »Befehle« steht kein schreibender', gezeigt.length > 0 && !gezeigt.some((n) => schreibende.includes(n)), `${gezeigt.join(', ')} — schreibend im Katalog: ${String(schreibende.length)}`)
  await p.bild('g07palette-leserbefehle')

  p.pruefe('⌘Q beendet', await p.beende())
}
