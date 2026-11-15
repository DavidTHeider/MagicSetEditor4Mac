// **Mantine in einer Ebene — am nativen Fenster** —
// #2451,
// gefunden in AP-3 (#2360).
//
// # Was es prüft, hell und dunkel
//
// - **Die Schrift folgt den Klassen**: Probeknöpfe und ein Probefeld mit
//   `font-medium`, `font-semibold` und `text-sm` in `main`, dazu der echte
//   Knopf »Weiter« (`Knopf`, `font-medium`) — berechneter Schnitt und
//   berechnete Größe gegen das, was die Klasse sagt. Bis #2451 schlug
//   Mantines ungeschichtete Regel `font: inherit` jede dieser Klassen.
// - **Die Mantine-Bauteile behalten ihre Gestalt**: je geöffnetem Zustand —
//   die Auswahl mit Suche (»Stil des neuen Sets«), die Palette (⌘K, ein
//   `Dialog`), die Schublade »Eigenschaften« und der Popover »Umbenennen«
//   an einem Deck — die berechnete Gestalt jedes sichtbaren Elements mit
//   einer Klasse `mantine-…`: Grund, Rand, Rundung, Innenabstand, Maße,
//   Schatten, Lage. **Mit `MSE_MANTINE_VORHER`** wird sie gegen einen
//   früheren Lauf verglichen; was sich außerhalb der Schrift ändert, ist rot.
//   `MSE_MANTINE_ZIEL` legt den Satz ab.
//
// # Was es nicht belegt
//
// **Ob die neue Schrift gut aussieht** und ob ein Knopf mit größerer Schrift
// irgendwo umbricht — das zeigen die Bilder und `uirevue`, `h03sprachen`.
// Der Bestand ist künstlich: das Probeset der Saat und ein Deck aus `mse`.
import { readFileSync, writeFileSync } from 'node:fs'
import { resolve } from 'node:path'

const wurzel = resolve(import.meta.dirname, '..', '..')
const ziel = resolve(wurzel, process.env.MSE_MANTINE_ZIEL ?? 'docs/belege/mantineebene.json')

const sichtbar = '(e) => e.checkVisibility({ checkOpacity: true, checkVisibilityCSS: true })'
const knopfMit = (name) => `[...document.querySelectorAll('button')].find((b) => (${sichtbar})(b) && (b.getAttribute('aria-label') ?? b.innerText).trim().replace(/\\s+/g, ' ') === ${JSON.stringify(name)})`

// schriftProben legt Knöpfe und ein Feld mit genau den Klassen aus dem
// Quelltext in `main` — sie stehen in Bauteilen, also erzeugt Tailwind sie.
const schriftProben = `document.getElementById('mantineebene')?.remove()
  const kasten = document.createElement('div'); kasten.id = 'mantineebene'
  const proben = [['button', 'font-medium text-sm', 'mittel-klein'], ['button', 'font-semibold', 'halbfett'], ['input', 'text-sm font-medium', 'feld']]
  for (const [art, klassen, name] of proben) { const e = document.createElement(art); e.className = klassen; e.dataset.probe = name; if (art === 'button') e.textContent = 'Probe'; kasten.appendChild(e) }
  document.querySelector('main').prepend(kasten)
  const wurzelgroesse = parseFloat(getComputedStyle(document.documentElement).fontSize)
  const lies = (e) => ({ schnitt: getComputedStyle(e).fontWeight, groesse: parseFloat(getComputedStyle(e).fontSize) })
  const werte = Object.fromEntries([...kasten.children].map((e) => [e.dataset.probe, lies(e)]))
  kasten.remove()
  return { wurzelgroesse, ...werte }`

// gestalt nimmt jedes sichtbare Element mit einer Mantine-Klasse auf — mit
// der ersten solchen Klasse als Namen.
const gestalt = `const aufnahme = {}
  for (const e of document.querySelectorAll('[class*="mantine-"]')) {
    if (!(${sichtbar})(e)) continue
    const name = [...e.classList].find((k) => k.startsWith('mantine-'))
    if (aufnahme[name]) continue
    const s = getComputedStyle(e); const r = e.getBoundingClientRect()
    aufnahme[name] = { grund: s.backgroundColor, rand: s.borderTopColor + ' ' + s.borderTopWidth + ' ' + s.borderTopStyle, rundung: s.borderTopLeftRadius,
      innen: s.paddingTop + ' ' + s.paddingRight + ' ' + s.paddingBottom + ' ' + s.paddingLeft, breite: Math.round(r.width), schatten: s.boxShadow, lage: s.position,
      schrift: s.fontWeight + ' ' + s.fontSize, hoehe: Math.round(r.height) }
  }
  return aufnahme`

// SCHRIFTFOLGEN sind die Merkmale, die sich mit der Schrift ändern dürfen:
// die Schrift selbst und die Höhe, die aus ihrer Zeile folgt.
const SCHRIFTFOLGEN = new Set(['schrift', 'hoehe', 'breite'])

async function zustand(p, name, oeffnen, schliessen) {
  await oeffnen()
  await new Promise((los) => setTimeout(los, 400))
  const wert = await p.seite(gestalt)
  p.zeige(`${name}: ${Object.keys(wert).length} Mantine-Elemente — ${Object.keys(wert).join(', ')}`)
  await p.bild(name)
  await schliessen()
  await new Promise((los) => setTimeout(los, 300))
  return wert
}

export default async function mantineebene(p) {
  const satz = { bauabdruck: p.lauf?.bauabdruck ?? '', erhoben: new Date().toISOString(), bilder: {} }
  for (const bild of ['hell', 'dunkel']) {
    const ort = p.ort()
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
    p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
    p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
    // **Im Spiel des Probesets**: ohne das Paket `magic` im Ort fehlte das
    // Deck im ersten Lauf auf der Seite.
    p.zeige(`Deck: ${p.mse(ort, 'deck', 'neu', 'Probedeck', 'probe').trim().split('\n')[0]}`)
    await p.starte({ ort })
    await p.groesse(1440, 900)
    const stand = { schrift: {}, gestalt: {} }

    p.notiz(`${bild}: die Schrift folgt den Klassen`)
    await p.taste('2', 'befehl')
    await p.wartetext('Werkstatt-Testkarten', 20000)
    const proben = await p.seite(schriftProben)
    const klein = Math.round(proben.wurzelgroesse * 0.875 * 100) / 100
    p.zeige(`${bild}: Proben ${JSON.stringify(proben)}`)
    p.pruefe(`${bild}: ein Knopf mit font-medium text-sm misst 500 und ${String(klein)} px`, proben['mittel-klein'].schnitt === '500' && proben['mittel-klein'].groesse === klein, JSON.stringify(proben['mittel-klein']))
    p.pruefe(`${bild}: ein Knopf mit font-semibold misst 600`, proben.halbfett.schnitt === '600', JSON.stringify(proben.halbfett))
    p.pruefe(`${bild}: ein Feld mit text-sm font-medium misst 500 und ${String(klein)} px`, proben.feld.schnitt === '500' && proben.feld.groesse === klein, JSON.stringify(proben.feld))
    const weiter = await p.seite(`const e = ${knopfMit('Weiter')}; return e ? getComputedStyle(e).fontWeight : null`)
    p.pruefe(`${bild}: »Weiter« (Knopf, font-medium) misst 500`, weiter === '500', String(weiter))
    stand.schrift = { ...proben, weiter }

    p.notiz(`${bild}: die Mantine-Bauteile`)
    const stilfeld = "[...document.querySelectorAll('main input[role=combobox], main [role=combobox]')].find((e) => e.checkVisibility())"
    stand.gestalt.suche = await zustand(p, `suche-${bild}`, () => p.klickeAuf(stilfeld, 'die Auswahl »Stil des neuen Sets«'), () => p.taste('Escape'))
    stand.gestalt.palette = await zustand(p, `palette-${bild}`, () => p.taste('k', 'befehl'), () => p.taste('Escape'))
    await p.klickeText('Werkstatt-Testkarten')
    await p.warte('das Set ist offen', () => p.seite(`return !!${knopfMit('Eigenschaften')}`), 20000)
    stand.gestalt.schublade = await zustand(p, `schublade-${bild}`, () => p.klickeAuf(knopfMit('Eigenschaften'), '»Eigenschaften«'), () => p.taste('Escape'))
    await p.taste('4', 'befehl')
    // **»Umbenennen« steht erst am geöffneten Deck** — der dritte Lauf fand
    // in der Liste nur »Öffnen« und »+ Marke«.
    await p.warte('das Deck steht zur Wahl', () => p.seite(`return !!${knopfMit('Öffnen: Probedeck')}`), 20000)
    await p.klickeAuf(knopfMit('Öffnen: Probedeck'), '»Öffnen: Probedeck«')
    await p.warte('das Deck steht', () => p.seite(`return !!${knopfMit('Umbenennen: Probedeck')}`), 20000).catch(async (grund) => {
      p.zeige(`Knöpfe mit »Probedeck«: ${await p.seite("return [...document.querySelectorAll('button')].filter((b) => ((b.getAttribute('aria-label') ?? '') + b.innerText).includes('Probedeck')).map((b) => (b.getAttribute('aria-label') ?? '') + ' | ' + b.innerText.trim() + ' | ' + b.checkVisibility()).join(' ; ')")}`)
      p.zeige(`Seite: ${await p.seite("return document.querySelector('main')?.dataset.seite ?? ''")}`)
      throw grund
    })
    stand.gestalt.popover = await zustand(p, `popover-${bild}`, () => p.klickeAuf(knopfMit('Umbenennen: Probedeck'), '»Umbenennen: Probedeck«'), () => p.taste('Escape'))
    for (const [art, wert] of Object.entries(stand.gestalt)) p.pruefe(`${bild}: ${art} zeigt Mantine-Elemente`, Object.keys(wert).length > 0, Object.keys(wert).join(', '))
    satz.bilder[bild] = stand
    p.pruefe(`${bild}: ⌘Q beendet`, await p.beende())
  }

  writeFileSync(ziel, `${JSON.stringify(satz, null, 2)}\n`)
  p.zeige(`Satz: ${ziel}`)

  if (process.env.MSE_MANTINE_VORHER !== undefined) {
    const vorher = JSON.parse(readFileSync(resolve(wurzel, process.env.MSE_MANTINE_VORHER), 'utf8'))
    p.notiz(`Gegen den Vorherlauf (Bauabdruck ${String(vorher.bauabdruck).slice(0, 12)})`)
    const gestaltwechsel = []
    for (const [bild, stand] of Object.entries(satz.bilder)) {
      for (const [art, elemente] of Object.entries(stand.gestalt)) {
        const alt = vorher.bilder?.[bild]?.gestalt?.[art] ?? {}
        for (const name of new Set([...Object.keys(alt), ...Object.keys(elemente)])) {
          const a = alt[name]; const n = elemente[name]
          if (a === undefined || n === undefined) { gestaltwechsel.push(`${bild}/${art}/${name}: ${a === undefined ? 'neu' : 'fehlt'}`); continue }
          for (const merkmal of Object.keys(a)) {
            if (a[merkmal] === n[merkmal]) continue
            const zeile = `${bild}/${art}/${name}.${merkmal}: ${String(a[merkmal])} → ${String(n[merkmal])}`
            if (SCHRIFTFOLGEN.has(merkmal)) p.zeige(`Schriftfolge ${zeile}`)
            else gestaltwechsel.push(zeile)
          }
        }
      }
    }
    p.pruefe('jedes Mantine-Element behält Grund, Rand, Rundung, Innenabstand, Schatten und Lage', gestaltwechsel.length === 0, gestaltwechsel.join('; '))
  }
}
