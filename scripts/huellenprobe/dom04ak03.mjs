// **Kopf, Decklisten, Abschluss, Datei und Bericht am nativen Fenster** —
// DOM-04.AK-02 bis AK-04
// (#1699):
//
// - ein Deck, zwei Namen, eine gespielte Runde — die Kommandozeile legt sie an,
// - AK-02: der Kopf nennt System, Runde und Teilnehmende; die Lasche
//   *Decklisten* schlägt das Deck auf,
// - AK-04: *Abschließen* — danach weist die Kommandozeile ein Anmelden ab,
//   und das Protokoll nennt den Abschluss,
// - AK-03: *Ausgeben …* öffnet den Dialog von macOS, ein Abbruch sagt
//   nichts; mit Ort liegt die Datei da, ohne Kennung des Bestands;
//   *Bericht sichern …* schreibt Markdown; *Turnier einlesen …* legt aus der
//   umbenannten Datei ein neues Turnier an und schlägt es auf.
import { readFileSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

import { rechtsklick } from '../huellenrufbuch.mjs'

// **Das Zeichen des Katalogs steht vor dem Text** — gesucht wird deshalb nach
// dem Ende, wo es keine Bezeichnung gibt.
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') === ${JSON.stringify(name)} || (!b.hasAttribute('aria-label') && b.innerText.trim().endsWith(${JSON.stringify(name)}))))`

async function klicke(p, name) {
  await p.seite(`${knopf(name)}.scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(knopf(name), name)
}

// ausDemMenue wählt einen Eintrag aus dem Menü „…“ des offenen Turniers —
// seit AP-6 (#2363, B-04) stehen Ausgeben und Bericht dort und nicht mehr als
// Knöpfe neben „Abschließen“.
async function ausDemMenue(p, turnier, eintrag) {
  await klicke(p, `Handlungen — ${turnier}`)
  const zeile = `[...document.querySelectorAll('[role=menuitem]')].find((e) => e.checkVisibility() && e.innerText.trim().endsWith(${JSON.stringify(eintrag)}))`
  await p.warte(`»${eintrag}« steht im Menü`, () => p.seite(`return !!${zeile}`), 10000)
  await p.klickeAuf(zeile, eintrag)
}

async function zuDenTurnieren(p) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld('Turniere')
  await p.klickeAuf("[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Turniere')", '»Turniere« in der Palette')
}

export default async function dom04ak03(p) {
  const ort = p.ort()
  const deckID = p.mse(ort, 'deck', 'neu', 'Aggro', 'magic').trim()
  const turnierID = p.mse(ort, 'turnier', 'neu', 'Freitag', 'magic').trim()
  p.mse(ort, 'turnier', 'anmelden', turnierID, 'Anna', deckID)
  p.mse(ort, 'turnier', 'anmelden', turnierID, 'Bert')
  const runde = JSON.parse(p.mse(ort, 'turnier', 'runde', turnierID, '--json')).runde
  for (const paarung of runde.paarungen) {
    if (paarung.b !== '') p.mse(ort, 'turnier', 'ergebnis', turnierID, paarung.id, '2', '0')
  }
  p.zeige(`Turnier ${turnierID}, Deck ${deckID}`)

  await p.starte({ ort })
  await zuDenTurnieren(p)
  await p.warte('das Turnier steht in der Liste', () => p.seite(`return !!${knopf('Öffnen: Freitag')}`), 20000)
  await klicke(p, 'Öffnen: Freitag')
  await p.wartetext('Schweizer System · Runde 1', 20000)

  p.notiz('AK-02 — der Kopf und die Lasche Decklisten')
  const kopf = await p.seite("return [...document.querySelectorAll('p')].find((e) => e.checkVisibility() && e.innerText.startsWith('Schweizer System'))?.innerText ?? ''")
  // **Ohne „von"** — Appendix E der MTR (Fassung 2026-02-27) empfiehlt für zwei Teilnehmende nichts, und
  // „Runde 1 von 0" war der erste Befund dieser Probe.
  p.pruefe('der Kopf nennt System, Runde und Teilnehmende', kopf === 'Schweizer System · Runde 1 · 2 Teilnehmende', kopf)
  await p.klickeText('Decklisten')
  await p.warte('Anna steht mit ihrem Deck da', () => p.seite(`return !!${knopf('Deck aufschlagen: Anna')}`), 10000)
  p.pruefe('Bert hat keine Liste und keinen Knopf', !(await p.seite(`return !!${knopf('Deck aufschlagen: Bert')}`)))
  await p.bild('decklisten')
  await klicke(p, 'Deck aufschlagen: Anna')
  await p.warte('das Deck ist aufgeschlagen', () => p.seite("return [...document.querySelectorAll('h2, h1')].some((e) => e.checkVisibility() && e.innerText.includes('Aggro'))"), 20000)
  await p.bild('deck-aufgeschlagen')

  // **Zurück auf den Turnieren steht es noch offen.** `Seiteninhalt` hält
  // jede besuchte Seite in einem `<Activity>` (UX-03), und seit UXR-12
  // (2026-10-03) ist die Liste gefaltet, solange eines offen ist. Bis zum
  // 2026-10-05 wartete dieses Drehbuch hier auf »Öffnen: Freitag« in der
  // Liste.
  await zuDenTurnieren(p)
  const zurueck = await p.warte('das Turnier steht wieder', () => p.seite(`if ([...document.querySelectorAll('p')].some((e) => e.checkVisibility() && e.innerText.startsWith('Schweizer System'))) return 'offen'; return ${knopf('Öffnen: Freitag')} ? 'in der Liste' : false`), 20000)
  p.pruefe('es ist noch offen, wie vor dem Deck', zurueck === 'offen', zurueck)

  p.notiz('AK-04 — Abschließen, und danach nimmt es nichts an')
  await klicke(p, 'Abschließen')
  await p.wartetext('Abgeschlossen am', 15000)
  let abgewiesen = false
  try {
    p.mse(ort, 'turnier', 'anmelden', turnierID, 'Cleo')
  } catch {
    abgewiesen = true
  }
  p.pruefe('die Kommandozeile weist ein Anmelden ab', abgewiesen)
  await p.klickeText('Protokoll')
  // **Gewartet wird auf die Wirkung** — »Abgeschlossen am« steht schon im
  // Kopf, bevor das Protokoll geladen ist.
  const liste = "[...document.querySelectorAll('ol, ul')].find((e) => e.checkVisibility() && e.innerText.includes('Angelegt als'))"
  await p.warte('das Protokoll ist geladen', () => p.seite(`return !!${liste}`), 10000)
  const protokoll = await p.seite(`return ${liste}.innerText`)
  p.pruefe('das Protokoll nennt Anlegen und Abschluss', protokoll.includes('Angelegt als »Freitag«') && protokoll.includes('Abgeschlossen'), protokoll.replace(/\s+/g, ' ').slice(0, 240))
  await p.bild('abgeschlossen')

  p.notiz('AP-6 (#2363, B-04) — Ausgeben, Bericht und Löschen hinter »…«, dieselbe Liste im Rechtsklick')
  p.pruefe('»Abschließen« oder »Wieder öffnen« steht als Knopf da', await p.seite(`return !!${knopf('Abschließen')} || [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.includes('öffnen'))`))
  for (const alt of ['Bericht sichern …', 'Turnier löschen: Freitag']) {
    p.pruefe(`»${alt}« steht nicht mehr als Knopf da`, !(await p.seite(`return !!${knopf(alt)}`)))
  }
  const turnierkopf = "[...document.querySelectorAll('section[aria-label=\"Freitag\"] h2')].find((h) => h.checkVisibility())"
  const stelle = await p.seite(`const e = ${turnierkopf}; if (!e) return null; e.scrollIntoView({ block: 'center' }); const r = e.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2 }`)
  await rechtsklick(p, stelle.x, stelle.y)
  const eintraege = "return [...document.querySelectorAll('[role=menu] [role=menuitem]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim())"
  await p.warte('das Menü des Turniers ist offen', async () => (await p.seite(eintraege)).length > 0, 10000).catch(() => undefined)
  const genannt = await p.seite(eintraege)
  p.pruefe('der Rechtsklick auf den Kopf nennt Ausgeben, Bericht und Löschen', ['Ausgeben …', 'Bericht sichern …', 'Turnier löschen'].every((n) => genannt.some((e) => e.endsWith(n))) && genannt.length === 3, genannt.join(' · '))
  await p.bild('turniermenue')
  await p.taste('Escape')

  p.notiz('AK-03 — Ausgeben: der Dialog, ein Abbruch sagt nichts')
  await ausDemMenue(p, 'Freitag', 'Ausgeben …')
  await p.dialogZeigtSich('sichern', '»Ausgeben …«')
  p.pruefe('nach dem Abbruch steht kein Ort da', !(await p.seite("return document.body.innerText.includes('Gesichert unter')")))

  const ordner = ort.ort
  const datei = join(ordner, 'freitag.mseturnier')
  await p.dialogantwort(datei)
  await ausDemMenue(p, 'Freitag', 'Ausgeben …')
  await p.wartetext(`Gesichert unter ${datei}.`, 15000)
  const inhalt = readFileSync(datei, 'utf8')
  p.pruefe('die Datei trägt Namen und Verlauf', inhalt.includes('"name": "Freitag"') && inhalt.includes('"runden"') && inhalt.includes('"abgeschlossen"'))
  p.pruefe('und keine Kennung des Bestands', !inhalt.includes('"id"') && !inhalt.includes(`"${deckID}"`))

  p.notiz('AK-03 — Bericht sichern')
  const bericht = join(ordner, 'freitag.md')
  await p.dialogantwort(bericht)
  await ausDemMenue(p, 'Freitag', 'Bericht sichern …')
  await p.wartetext(`Gesichert unter ${bericht}.`, 15000)
  const markdown = readFileSync(bericht, 'utf8')
  p.pruefe('der Bericht nennt Kopf, Rangliste und Runde', markdown.startsWith('# Freitag\n') && markdown.includes('## Rangliste') && markdown.includes('## Runde 1'), markdown.split('\n').slice(0, 6).join(' / '))
  await p.bild('ausgegeben')

  p.notiz('AK-03 — Turnier einlesen')
  writeFileSync(datei, inhalt.replace('"Freitag"', '"Samstag"'))
  await p.dialogantwort(datei)
  await klicke(p, 'Turnier einlesen …')
  await p.wartetext('Eingelesen und aufgeschlagen.', 15000)
  await p.warte('das neue Turnier ist aufgeschlagen', () => p.seite("return [...document.querySelectorAll('section[aria-label]')].some((e) => e.checkVisibility() && e.getAttribute('aria-label') === 'Samstag')"), 15000)
  const turniere = p.mse(ort, 'turniere')
  p.pruefe('die Kommandozeile kennt beide', turniere.includes('Freitag') && turniere.includes('Samstag'), turniere.trim().replace(/\n/g, ' / '))
  await p.bild('eingelesen')

  p.pruefe('⌘Q beendet', await p.beende())
}
