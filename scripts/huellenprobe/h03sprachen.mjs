// **Alle acht Sprachen im kleinsten Fenster** — H03
// (#2257),
// am nativen Fenster.
//
// Je Sprache ein eigener Bestand mit dieser Sprache, das Fenster im
// Mindestmaß 1024 × 700 (`internal/fenster/bildschirm.go`), das Saatset
// offen, und dann **jede registrierte Seite** (`frontend/src/navigation.ts`,
// gelesen und nicht abgeschrieben) samt der Befehlspalette als Dialog:
//
//   - die Wurzel trägt `lang` dieser Sprache,
//   - keine Seite rollt waagerecht,
//   - **keine Beschriftung ist abgeschnitten, ohne dass ihr voller Text
//     erreichbar bleibt** — als `title` oder als Name (`aria-label`) eines
//     Vorfahren. Abgeschnitten heißt: ein Blatt mit Text, das breiter ist als
//     sein Kasten und nicht überläuft. **Was nur der Vorleser hört**
//     (`.sr-only`), zählt nicht — es ist mit Absicht ein Punkt groß.
//
// Dazu **eine Schublade** (*Paket beziehen …*) und **eine Fehlermeldung**
// je Sprache: die Bindung `Pakete` scheitert fachlich (`stoere`,
// `internal/fenster/rufstoerung.go`), und die Meldung steht in der Sprache
// des Fensters, ohne Abgeschnittenes.
//
// **Was er nicht belegt**: ob eine Übersetzung gut ist, und jede einzelne
// Fehlermeldung — die Probe zeigt eine, die Zustandsprobe fährt die übrigen
// Fehlerwege auf Deutsch. Vorlagen- und Nutzerdaten stehen in der Sprache, in
// der sie kamen — der Name des Saatsets ist in jeder Sprache derselbe.
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import { stoere } from '../huellenrufbuch.mjs'
import { paletteOffen } from './budgets.mjs'
import { seitenliste } from './zustaende.mjs'

const sprachen = ['de', 'en', 'es', 'fr', 'it', 'ja', 'ru', 'zh']

// **Chinesisch trägt die Schrift im Etikett** — `zh-Hans` (`bcp47` in
// `frontend/src/sprache/kataloge.ts`, entschieden in `docs/SPRACHINVENTAR.md`):
// sonst darf WebKit für ein Han-Zeichen die japanische Glyphe nehmen.
const etikett = (sprache) => (sprache === 'zh' ? 'zh-Hans' : sprache)

// befund liest, was an der sichtbaren Fläche — Schublade, sonst `main`, dazu
// die Seitenleiste — abgeschnitten ist oder waagerecht überläuft.
const befund = `return (() => {
  const schublade = [...document.querySelectorAll('[role=dialog]')].find((d) => d.checkVisibility())
  const wurzeln = [schublade ?? document.querySelector('main'), document.querySelector('nav')].filter(Boolean)
  const voll = (e) => (e.textContent ?? '').replace(/\\s+/g, ' ').trim()
  const erreichbar = (e) => {
    const text = voll(e)
    for (let k = e; k; k = k.parentElement) {
      const name = (k.getAttribute('title') ?? '') + ' ' + (k.getAttribute('aria-label') ?? '')
      if (name.replace(/\\s+/g, ' ').includes(text)) return true
    }
    return false
  }
  const zuBreit = wurzeln.flatMap((w) => [...w.querySelectorAll('*')])
    .filter((e) => e.children.length === 0 && voll(e) !== '' && e.checkVisibility() && !e.closest('.sr-only'))
    .filter((e) => e.scrollWidth > e.clientWidth + 1 && getComputedStyle(e).overflowX !== 'visible')
  const nenne = (e) => e.tagName.toLowerCase() + ' »' + voll(e).slice(0, 40) + '«'
  return {
    lang: document.documentElement.lang,
    waagerecht: document.scrollingElement.scrollWidth > window.innerWidth + 1,
    abgeschnitten: [...new Set(zuBreit.filter((e) => !erreichbar(e)).map(nenne))].slice(0, 6),
    mitTitel: [...new Set(zuBreit.filter(erreichbar).map(nenne))].slice(0, 6),
  }
})()`

async function zurSeite(p, schluessel) {
  await p.taste('k', 'befehl')
  await p.warte('die Befehlspalette steht da', () => p.seite(paletteOffen), 15000)
  const zeile = `document.querySelector('[data-palettenzeile][data-seite=${JSON.stringify(schluessel)}]')`
  if (!(await p.seite(`return !!(${zeile})`))) {
    await p.taste('Escape')

    return false
  }
  await p.seite(`(${zeile}).scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(zeile, `»${schluessel}« in der Palette`)
  await p.warte(`die Seite ${schluessel} steht da`, () => p.seite(`return document.querySelector('main')?.dataset.seite === ${JSON.stringify(schluessel)}`), 15000)

  return true
}

export default async function h03sprachen(p) {
  const { anwendung, set } = seitenliste(readFileSync(resolve('frontend/src/navigation.ts'), 'utf8'))

  for (const sprache of sprachen) {
    p.notiz(sprache)
    const ort = p.ort()
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', sprache)
    p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
    p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'normal')
    p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
    await p.starte({ ort })
    await p.groesse(1024, 700)
    await new Promise((weiter) => { setTimeout(weiter, 1000) })

    // **Die Palette als Dialog** — sie ist die eine Fläche, die jede Seite
    // gleich überdeckt.
    await p.taste('k', 'befehl')
    await p.warte('die Befehlspalette steht da', () => p.seite(paletteOffen), 15000)
    const palette = await p.seite(befund)
    p.pruefe(`${sprache}: die Wurzel trägt lang=${etikett(sprache)}`, palette.lang === etikett(sprache), palette.lang)
    p.pruefe(`${sprache}, Palette: nichts abgeschnitten`, palette.abgeschnitten.length === 0, palette.abgeschnitten.join(' · '))
    await p.bild(`h03-${sprache}-palette`)
    await p.taste('Escape')

    // **Gekürzt, aber lesbar** — was zu breit ist und seinen vollen Text als
    // Titel trägt, wird genannt: das ist der Weg nachher zum Befund aus dem
    // ersten Lauf („Nutzer und Profile" auf fr, ja, ru ohne vollen Text).
    p.zeige(`${sprache}: gekürzt mit vollem Text als Titel — ${palette.mitTitel.join(' · ') || 'nichts'}`)

    // **Erst das Set**, dann gibt es die Setseiten.
    await p.taste('2', 'befehl')
    await p.wartetext('Werkstatt-Testkarten', 20000)
    await p.klickeText('Werkstatt-Testkarten')

    const ohneBefund = []
    for (const seite of [...anwendung, ...set]) {
      if (!(await zurSeite(p, seite))) {
        p.pruefe(`${sprache}, ${seite}: die Seite kam`, false)

        continue
      }
      await p.ruhe()
      const l = await p.seite(befund)
      if (l.waagerecht || l.abgeschnitten.length > 0) {
        p.pruefe(`${sprache}, ${seite}: rollt nicht waagerecht, nichts abgeschnitten`, false, `${l.waagerecht ? 'rollt waagerecht · ' : ''}${l.abgeschnitten.join(' · ')}`)
        await p.bild(`h03-${sprache}-${seite}`)
      } else {
        ohneBefund.push(seite)
      }
    }
    await p.bild(`h03-${sprache}-leiste`)

    // **Eine Schublade** — die führende Handlung der Paketseite, der letzte
    // Knopf in ihrem Kopf: ihre Aufschrift ist in jeder Sprache eine andere.
    if (await zurSeite(p, 'pakete')) {
      await p.klickeAuf("[...document.querySelectorAll('main header button')].filter((b) => b.checkVisibility()).at(-1)", 'den Hauptknopf der Paketseite')
      await p.warte('die Schublade steht da', () => p.seite("return [...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility())"), 15000)
      await p.ruhe()
      const schublade = await p.seite(befund)
      p.pruefe(`${sprache}, Schublade: nichts abgeschnitten, kein waagerechtes Rollen`, !schublade.waagerecht && schublade.abgeschnitten.length === 0, schublade.abgeschnitten.join(' · '))
      await p.bild(`h03-${sprache}-schublade`)
      await p.taste('Escape')
    }

    // **Eine Fehlermeldung** — `Pakete` scheitert fachlich, die Seite wird neu
    // geladen, und die Meldung steht am Ort, in dieser Sprache.
    await stoere(p, [{ bindung: 'Pakete', fehler: true }])
    try {
      await p.seite('setTimeout(() => location.reload(), 30); return true')
      await new Promise((weiter) => { setTimeout(weiter, 1500) })
      await p.warte('die Leiste steht wieder', () => p.seite("return document.readyState === 'complete' && document.querySelectorAll('nav button').length > 3"), 30000)
      await zurSeite(p, 'pakete')
      await p.warte('eine Meldung steht da', () => p.seite("return [...document.querySelectorAll('main [role=alert]')].some((e) => e.checkVisibility() && e.innerText.trim() !== '')"), 15000)
      const meldung = await p.seite("return [...document.querySelectorAll('main [role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).join(' / ').slice(0, 160)")
      const l = await p.seite(befund)
      p.pruefe(`${sprache}, Fehlermeldung: steht ganz da — »${meldung}«`, !l.waagerecht && l.abgeschnitten.length === 0, l.abgeschnitten.join(' · '))
      await p.bild(`h03-${sprache}-fehler`)
    } finally {
      await stoere(p, [])
    }
    p.pruefe(`${sprache}: ${String(ohneBefund.length)} von ${String(anwendung.length + set.length)} Seiten ohne Überlauf und ohne Abgeschnittenes`, ohneBefund.length === anwendung.length + set.length)
    p.pruefe(`${sprache}: ⌘Q beendet`, await p.beende())
  }
}
