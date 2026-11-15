// **Rückmeldung, Laden, Vorschauwechsel und Ablageziel — am nativen Fenster**
// — AP-4 (#2361),
// Befunde B-11, B-12, B-24, Entwurf `Zustaende09`, Vermerk 0065
// Festlegungen 10 und 11.
//
// # Was es prüft, hell und dunkel
//
// - **Ein Knopf in fünf Zuständen**, mit der echten Maus: Ruhe, Hover (der
//   Grund wechselt), Gedrückt (ein Punkt tiefer), Fokus (Ring per Tab),
//   Gesperrt (weder Hover noch Druck) — je Gewicht, dazu die Dauer der
//   Übergänge (120 ms) und die Kontraste unter dem Zeiger.
// - **Eine Kartenzeile** hebt ihren Grund unter dem Zeiger.
// - **Überlagerungen erscheinen in 160 ms**: Palette und Menü tauchen auf,
//   die Schublade schiebt sich herein, der Popover blendet über; Palette,
//   Popover und Menü tragen die eine Schattenstufe.
// - **Laden in Endhöhe**: mit einer um 6 s verzögerten Bindung `Sets`
//   (`stoere`) und neu geladener Seite steht der Überblick als Platzhalter da, und die Abschnitte
//   darunter bleiben, wo sie sind, wenn die Zahlen kommen.
// - **Vorschauwechsel ohne Leerblitz**: mit verzögerter `Vorschau` steht
//   beim Wechsel das vorige Bild gedämpft und `inert` da, ohne Bildfeld.
// - **Das Ablageziel spricht vor dem Loslassen**: eine Datei wird über ein
//   Bildfeld und über die leere Seite gezogen (`ueberziehen`), ohne
//   Loslassen; der Satz nennt, was geschähe, und das Feld trägt eine
//   gestrichelte Kante.
// - **Wenig Bewegung** (die Einstellung der Anwendung): jede Dauer 0.
//
// # Was es nicht belegt
//
// **Die Systemeinstellung** „Bewegung reduzieren“ und „Kontrast erhöhen“ —
// sie gehören dem System, und ein Mensch schaltet sie (Handprüfung). **Ein
// echtes Ziehen aus dem Finder** — `ueberziehen` ruft dieselben Schritte an
// AppKit, nicht den Fensterserver. **Ob es gut aussieht** — die Bilder.
import { kontraste, pruefeKontraste } from '../huellenkontrast.mjs'
import { loslassen, rechtsklick, stoere, ziehe } from '../huellenrufbuch.mjs'

const schlafe = (ms) => new Promise((los) => setTimeout(los, ms))
const sichtbar = '(e) => e.checkVisibility({ checkOpacity: true, checkVisibilityCSS: true })'
const knopfMit = (name) => `[...document.querySelectorAll('button')].find((b) => (${sichtbar})(b) && (b.getAttribute('aria-label') ?? b.innerText).trim().replace(/\\s+/g, ' ') === ${JSON.stringify(name)})`
const mitte = (wer) => `const e = ${wer}; if (!e) return null; const r = e.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2 }`
const zustand = (wer) => `const e = ${wer}; if (!e) return null; const s = getComputedStyle(e)
  return { grund: s.backgroundColor, lage: s.translate, dauer: s.transitionDuration, eigenschaften: s.transitionProperty, deckkraft: s.opacity, schatten: s.boxShadow, animation: s.animationName, animationsdauer: s.animationDuration }`
const ms = (dauer) => (dauer ?? '').split(',').map((d) => parseFloat(d) * (d.trim().endsWith('ms') ? 1 : 1000))
const weg = { x: 8, y: 300 }

// knopf misst einen Knopf in Ruhe, unter dem Zeiger und gedrückt.
async function knopf(p, bild, name, { gesperrt = false } = {}) {
  const wer = knopfMit(name)
  const stelle = await p.seite(mitte(wer))
  if (stelle === null) {
    p.pruefe(`${bild}: »${name}« steht da`, false)
    return
  }
  await p.frage({ was: 'bewege', x: weg.x, y: weg.y })
  await schlafe(250)
  const ruhe = await p.seite(zustand(wer))
  await p.frage({ was: 'bewege', x: stelle.x, y: stelle.y })
  await schlafe(300)
  const hover = await p.seite(zustand(wer))
  if (!gesperrt) await pruefeKontraste(p, bild, [[`»${name}« unter dem Zeiger`, wer, 'text', 4.5]])
  await ziehe(p, [stelle, stelle], true)
  await schlafe(250)
  const gedrueckt = await p.seite(zustand(wer))
  await loslassen(p, weg.x, weg.y)
  await p.frage({ was: 'bewege', x: weg.x, y: weg.y })
  await schlafe(250)
  p.zeige(`${bild}: »${name}« Ruhe ${JSON.stringify(ruhe)} · Hover ${hover.grund} ${hover.lage} · gedrückt ${gedrueckt.grund} ${gedrueckt.lage}`)
  if (gesperrt) {
    p.pruefe(`${bild}: der gesperrte »${name}« antwortet weder dem Zeiger noch dem Druck`, hover.grund === ruhe.grund && !/1px/.test(gedrueckt.lage), `${ruhe.grund} → ${hover.grund}, ${gedrueckt.lage}`)
    return
  }
  p.pruefe(`${bild}: »${name}« hebt seinen Grund unter dem Zeiger`, hover.grund !== ruhe.grund, `${ruhe.grund} → ${hover.grund}`)
  p.pruefe(`${bild}: »${name}« sinkt gedrückt um einen Punkt`, /(^|\s)1px$/.test(gedrueckt.lage), gedrueckt.lage)
  p.pruefe(`${bild}: »${name}« wechselt in 120 ms, nur Farbe, Deckkraft und Lage`, ms(ruhe.dauer).every((d) => d === 120) && !/all|width|height|margin|padding/.test(ruhe.eigenschaften), `${ruhe.dauer} ${ruhe.eigenschaften}`)
}

// ueberlagerung öffnet eine Überlagerung und liest, wie sie erscheint.
async function ueberlagerung(p, bild, name, oeffnen, wer, schliessen) {
  await oeffnen()
  await p.warte(`${name} steht`, () => p.seite(`return !!(${wer})`), 15000)
  const wert = await p.seite(zustand(wer))
  p.zeige(`${bild}: ${name} ${JSON.stringify(wert)}`)
  await p.bild(`${name}-${bild}`)
  await schliessen()
  await schlafe(300)
  return wert
}

async function lauf(p, bild, reduziert) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Probedeck', 'probe')
  if (reduziert) {
    const nutzer = JSON.parse(p.mse(ort, 'nutzer', '--json')).nutzer[0].id
    p.mse(ort, 'nutzereinstellung', nutzer, 'barrierefreiheit.bewegung', 'reduziert')
  }
  await p.starte({ ort })
  await p.groesse(1440, 900)
  await p.aktiviere()
  const wo = `${bild}${reduziert ? ', wenig Bewegung' : ''}`

  p.notiz(`${wo}: Sets — Knöpfe in jedem Gewicht`)
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  if (reduziert) {
    const dauer = await p.seite(zustand(knopfMit('Weiter')))
    p.pruefe(`${wo}: der Übergang eines Knopfs dauert höchstens 0,01 ms`, ms(dauer.dauer).every((d) => d <= 0.01), dauer.dauer)
  } else {
    await knopf(p, wo, 'Set öffnen …')
    await knopf(p, wo, 'Weiter')
    await knopf(p, wo, 'Überspringen')
    await p.bild(`sets-${bild}`)
  }

  p.notiz(`${wo}: Überlagerungen`)
  const palette = await ueberlagerung(p, wo, 'palette', () => p.taste('k', 'befehl'), "[...document.querySelectorAll('[role=dialog]')].find((e) => e.checkVisibility())", () => p.taste('Escape'))
  const paletteflaeche = await p.seite("const e = [...document.querySelectorAll('[role=dialog] .shadow-ueberlage')].find((x) => x.checkVisibility()); return e ? getComputedStyle(e).boxShadow : 'none'")
  await p.taste('k', 'befehl')
  await schlafe(400)
  const paletteSchatten = await p.seite("const e = [...document.querySelectorAll('[role=dialog] .shadow-ueberlage')].find((x) => x.checkVisibility()); return e ? getComputedStyle(e).boxShadow : 'none'")
  await p.taste('Escape')
  p.zeige(`${wo}: Schatten der Palette ${paletteSchatten} (vorher gelesen: ${paletteflaeche})`)
  await p.klickeText('Werkstatt-Testkarten')
  await p.warte('das Set ist offen', () => p.seite(`return !!${knopfMit('Eigenschaften')}`), 20000)
  const schublade = await ueberlagerung(p, wo, 'schublade', () => p.klickeAuf(knopfMit('Eigenschaften'), '»Eigenschaften«'), "[...document.querySelectorAll('[role=dialog]')].find((e) => e.checkVisibility())", () => p.taste('Escape'))
  await p.warte('die Karte steht mit Bild', () => p.seite("return !!document.querySelector('img[alt=\"Vorschau der Karte\"]')?.checkVisibility()"), 30000)
  const bildstelle = await p.seite(mitte("document.querySelector('img[alt=\"Vorschau der Karte\"]')"))
  const menue = await ueberlagerung(p, wo, 'menue', () => rechtsklick(p, bildstelle.x, bildstelle.y), "[...document.querySelectorAll('[role=menu]')].find((e) => e.checkVisibility())", () => p.taste('Escape'))

  if (reduziert) {
    for (const [name, wert] of [['Palette', palette], ['Schublade', schublade], ['Menü', menue]]) {
      p.pruefe(`${wo}: ${name} erscheint ohne Bewegung`, ms(wert.animationsdauer).every((d) => d <= 0.01), wert.animationsdauer)
    }
  } else {
    p.pruefe(`${wo}: die Palette taucht in 160 ms auf`, palette.animation === 'auftauchen' && ms(palette.animationsdauer)[0] === 160, `${palette.animation} ${palette.animationsdauer}`)
    p.pruefe(`${wo}: die Schublade schiebt sich in 160 ms herein`, schublade.animation === 'einschieben' && ms(schublade.animationsdauer)[0] === 160, `${schublade.animation} ${schublade.animationsdauer}`)
    p.pruefe(`${wo}: das Menü taucht in 160 ms auf und trägt die Schattenstufe`, menue.animation === 'auftauchen' && ms(menue.animationsdauer)[0] === 160 && menue.schatten !== 'none', `${menue.animation} ${menue.animationsdauer} ${menue.schatten}`)
    p.pruefe(`${wo}: die Palette trägt die Schattenstufe`, paletteSchatten !== 'none', paletteSchatten)
  }

  p.notiz(`${wo}: Kartenzeile unter dem Zeiger`)
  const zeile = "[...document.querySelectorAll('main [role=row][data-karte]')].filter((e) => e.checkVisibility())[1]"
  const zeilenstelle = await p.seite(mitte(zeile))
  if (!reduziert && zeilenstelle !== null) {
    const ruhe = await p.seite(zustand(zeile))
    await p.frage({ was: 'bewege', x: zeilenstelle.x, y: zeilenstelle.y })
    await schlafe(300)
    const hover = await p.seite(zustand(zeile))
    p.pruefe(`${wo}: die Kartenzeile hebt ihren Grund unter dem Zeiger`, hover.grund !== ruhe.grund, `${ruhe.grund} → ${hover.grund}`)
    await p.frage({ was: 'bewege', x: weg.x, y: weg.y })
  }

  p.notiz(`${wo}: Vorschauwechsel mit verzögerter Vorschau`)
  const spalte = "document.querySelector('img[alt=\"Vorschau der Karte\"]')?.closest('div.w-full, div[class*=\"w-2/5\"]')"
  const hoeheVorher = await p.seite(`return Math.round(${spalte}?.getBoundingClientRect().height ?? 0)`)
  const bildVorher = await p.seite("return document.querySelector('img[alt=\"Vorschau der Karte\"]')?.getAttribute('src') ?? ''")
  await stoere(p, [{ bindung: 'Vorschau', verzoegerungMs: 2500 }])
  await p.klickeAuf(`(${zeile})?.querySelector('button')`, 'die zweite Karte')
  await schlafe(600)
  const zwischen = await p.seite(`const b = document.querySelector('img[alt="Vorschau der Karte"]'); const h = b?.closest('[inert]'); return { bild: b?.getAttribute('src') ?? '', inert: !!h, beschaeftigt: h?.getAttribute('aria-busy') ?? '', bildfelder: h ? h.querySelectorAll('[data-bildfeld]').length : -1, schild: h?.querySelector('p')?.checkVisibility() ? h.querySelector('p').innerText.trim() : '', hoehe: Math.round(${spalte}?.getBoundingClientRect().height ?? 0) }`)
  p.zeige(`${wo}: während des Wechsels ${JSON.stringify(zwischen)}, vorher ${bildVorher} ${String(hoeheVorher)}`)
  await p.bild(`vorschauwechsel-${bild}${reduziert ? '-reduziert' : ''}`)
  p.pruefe(`${wo}: beim Wechsel steht das vorige Bild, inert und beschäftigt, ohne Bildfeld`, zwischen.bild === bildVorher && zwischen.inert && zwischen.beschaeftigt === 'true' && zwischen.bildfelder === 0, JSON.stringify(zwischen))
  p.pruefe(`${wo}: über dem vorigen Bild steht sichtbar, dass geladen wird`, zwischen.schild !== '', zwischen.schild)
  p.pruefe(`${wo}: die Spalte behält ihre Höhe`, Math.abs(zwischen.hoehe - hoeheVorher) <= 1, `${String(hoeheVorher)} → ${String(zwischen.hoehe)}`)
  await p.warte('das neue Bild steht', () => p.seite("const b = document.querySelector('img[alt=\"Vorschau der Karte\"]'); return !!b && !b.closest('[inert]')"), 15000)
  await stoere(p, [])

  if (!reduziert) {
    p.notiz(`${wo}: das Ablageziel vor dem Loslassen`)
    const feld = await p.seite("const f = [...document.querySelectorAll('[data-bildfeld]')].find((e) => e.checkVisibility()); if (!f) return null; const r = f.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2, feld: f.dataset.bildfeld }")
    const satz = () => p.seite("return [...document.querySelectorAll('[role=status][aria-live]')].map((e) => e.innerText.trim()).filter(Boolean).join(' | ')")
    if (feld !== null) {
      await p.frage({ was: 'ueberziehen', pfad: ['/tmp/ap4-probe.png'], x: feld.x, y: feld.y })
      await schlafe(400)
      const ueberFeld = await satz()
      const kante = await p.seite(`const f = document.querySelector('[data-bildfeld=${JSON.stringify(feld.feld)}]'); const s = getComputedStyle(f); return { stil: s.outlineStyle, farbe: s.outlineColor, aktiv: f.classList.contains('file-drop-target-active') }`)
      await p.bild(`ablage-bildfeld-${bild}`)
      p.zeige(`${wo}: über »${feld.feld}«: ${ueberFeld} · Kante ${JSON.stringify(kante)}`)
      p.pruefe(`${wo}: über dem Bildfeld nennt der Satz das Feld`, ueberFeld.includes(`»${feld.feld}«`), ueberFeld)
      p.pruefe(`${wo}: das Bildfeld trägt eine gestrichelte Kante`, kante.aktiv && kante.stil === 'dashed', JSON.stringify(kante))
      await p.seite("const s = document.createElement('span'); s.id = 'ap4kante'; s.textContent = 'x'; s.style.color = 'var(--color-marke)'; document.querySelector('main').appendChild(s); return true")
      for (const w of await kontraste(p, [['Ablagekante gegen den Seitengrund', "document.getElementById('ap4kante')", 'text', 3]])) p.zeige(`${wo}: ${w.name} ${String(w.verhaeltnis).replace('.', ',')}`)
      await pruefeKontraste(p, wo, [['Ablagekante gegen den Seitengrund', "document.getElementById('ap4kante')", 'text', 3]])
      await p.seite("document.getElementById('ap4kante')?.remove(); return true")
      await p.frage({ was: 'wegziehen' })
      await schlafe(300)
      p.pruefe(`${wo}: nach dem Wegziehen schweigt der Hinweis`, (await satz()) === '', await satz())
    }
    await p.frage({ was: 'ueberziehen', pfad: ['/tmp/ap4-probe.mse-set'], x: weg.x + 300, y: 120 })
    await schlafe(400)
    const ueberSeite = await satz()
    p.zeige(`${wo}: über der Seite: ${ueberSeite}`)
    p.pruefe(`${wo}: über der Seite sagt der Satz, dass ein Set geöffnet wird`, ueberSeite.includes('.mse-set'), ueberSeite)
    await p.bild(`ablage-seite-${bild}`)
    await p.frage({ was: 'wegziehen' })
    await schlafe(300)

    p.notiz(`${wo}: der Popover am Deck`)
    await p.taste('4', 'befehl')
    await p.warte('das Deck steht zur Wahl', () => p.seite(`return !!${knopfMit('Öffnen: Probedeck')}`), 20000)
    await p.klickeAuf(knopfMit('Öffnen: Probedeck'), '»Öffnen: Probedeck«')
    // **Umbenennen steht seit AP-6 hinter „…“** (#2363) — der Popover geht
    // am Menü auf, nicht mehr an einem eigenen Knopf.
    await p.warte('das Deck ist offen', () => p.seite(`return !!${knopfMit('Handlungen — Probedeck')}`), 20000)
    await knopf(p, wo, 'Ins Deck legen', { gesperrt: true })
    const umbenennen = async () => {
      await p.klickeAuf(knopfMit('Handlungen — Probedeck'), '»…« am Deck')
      await p.klickeAuf("[...document.querySelectorAll('[role=menuitem]')].find((e) => e.checkVisibility() && e.innerText.includes('Umbenennen'))", '»Umbenennen«')
    }
    const popover = await ueberlagerung(p, wo, 'popover', umbenennen, "document.querySelector('.mantine-Popover-dropdown')", () => p.taste('Escape'))
    p.pruefe(`${wo}: der Popover blendet in 160 ms über und trägt die Schattenstufe`, ms(popover.dauer)[0] === 160 && popover.schatten !== 'none', `${popover.dauer} ${popover.schatten}`)

    p.notiz(`${wo}: die Startseite lädt in Endhöhe`)
    // **Neu geladen, damit die Startseite wirklich neu fragt** — besuchte
    // Seiten bleiben sonst im Baum stehen, und die verzögerte Bindung trifft
    // keinen Ladevorgang (erster Lauf). Dieselbe Gestalt wie `zustaende`.
    await stoere(p, [{ bindung: 'Sets', verzoegerungMs: 6000 }])
    await p.seite('setTimeout(() => location.reload(), 30); return true')
    await schlafe(1500)
    await p.warte('die Leiste steht wieder', () => p.seite("return document.readyState === 'complete' && document.querySelectorAll('nav button').length > 3"), 30000)
    await p.taste('1', 'befehl')
    await schlafe(500)
    const abschnitt = "[...document.querySelectorAll('main h2')].find((e) => e.checkVisibility())"
    const frueh = await p.seite(`return { oben: Math.round(${abschnitt}?.getBoundingClientRect().top ?? -1), beschaeftigt: document.querySelector('main ul[aria-busy=true]') !== null }`)
    await p.bild(`start-laedt-${bild}`)
    await p.warte('die Zahlen stehen', () => p.seite("return document.querySelector('main ul[aria-busy=true]') === null"), 15000)
    await schlafe(300)
    const spaet = await p.seite(`return Math.round(${abschnitt}?.getBoundingClientRect().top ?? -1)`)
    await stoere(p, [])
    p.zeige(`${wo}: Startseite laden ${JSON.stringify(frueh)}, danach ${String(spaet)}`)
    p.pruefe(`${wo}: beim Laden steht der Überblick als Platzhalter`, frueh.beschaeftigt, JSON.stringify(frueh))
    p.pruefe(`${wo}: die Abschnitte darunter bleiben, wo sie sind`, frueh.oben >= 0 && Math.abs(frueh.oben - spaet) <= 1, `${String(frueh.oben)} → ${String(spaet)}`)
  }
  p.pruefe(`${wo}: ⌘Q beendet`, await p.beende())
}

export default async function ap4zustaende(p) {
  for (const bild of ['hell', 'dunkel']) await lauf(p, bild, false)
  await lauf(p, 'hell', true)
}
