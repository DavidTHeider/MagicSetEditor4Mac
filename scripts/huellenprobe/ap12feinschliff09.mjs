// **Mac-Feinschliff am nativen Fenster** — AP-12
// (#2369),
// Befunde B-25, B-26, B-28.
//
// # Was es prüft
//
// 1. **Trenner** (B-25): zwischen Kartenliste und Vorschau steht ein Trenner
//    mit Wert; → geht 16 Punkte, Ende an die obere Grenze, und dem Zentrum
//    bleiben dort 480 Punkte. Ein Zug mit der Maus (AppKit, `ziehe`) macht
//    die Spalte breiter. `mse einstellungen` kennt die Breite, und **nach
//    ⌘Q und neuem Start** steht die Spalte so breit wieder da.
// 2. **Symbol** (B-26): „Größer“ ist ein Umschalter; gedrückt geht die
//    Schublade auf 56rem, die Zeichenfläche auf 32rem.
// 3. **Pakettabelle** (B-28): bei 1.440 × 838 und 1.024 × 700 läuft keine
//    Fassung in den Chip daneben, und kein Name bricht mitten im Wort.
// 4. **Setangaben** (B-28): kein eigener Rahmen in der Schublade.
// 5. **Sammlung** (B-28): zwischen Navigator und Inspector steht jede
//    Bezeichnung ganz da — Sprache, Set und Lagerort stehen schmal darunter —,
//    keine Zeile läuft in ihre Nachbarn, jeder Kopf steht über seiner Zelle,
//    und eine Auswahl schiebt die Liste nicht. Mit Namen wie im Handbuchbild:
//    dort liefen die umgebrochenen Bezeichnungen des ersten Schnitts
//    ineinander, und die kurzen Namen dieses Drehbuchs hatten es nicht
//    gezeigt.
// 6. **Diagnose** (B-28): keine zwei Trennlinien direkt untereinander.
//
// Hell und dunkel, künstlicher Bestand, eigenes Zuhause. Mit
// `MSE_AP12_GEGENPROBE=1` setzt es in der Diagnose zwei Linien drei Punkte
// untereinander — dann muss Prüfung 6 rot werden —, und in der Sammlung
// bricht es die Bezeichnungen um und verteilt die Köpfe gleich: dann müssen
// die drei Prüfungen aus 5 rot werden.
//
// # Was es nicht belegt
//
// **Ob ein Mensch den Trenner findet** — das zeigt das Bild. **VoiceOver**
// hört ein Mensch. **Abgeschnittene Texte in acht Sprachen** misst
// `h03sprachen`, nicht dieses Drehbuch.
import { ziehe } from '../huellenrufbuch.mjs'

const trenner = "document.querySelector('[role=separator][aria-label^=\"Breite von\"]')"
const spalteVon = (t) => `document.getElementById(${t}.getAttribute('aria-controls'))`

async function zurSeite(p, schluessel) {
  await p.taste('k', 'befehl')
  await p.warte('die Befehlspalette steht da', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"), 15000)
  const zeile = `document.querySelector('[data-palettenzeile][data-seite=${JSON.stringify(schluessel)}]')`
  await p.warte(`»${schluessel}« steht in der Palette`, () => p.seite(`return !!(${zeile})`), 10000)
  await p.seite(`(${zeile}).scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(zeile, `»${schluessel}« in der Palette`)
}

async function oeffneSet(p) {
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await p.warte('die Kartenliste steht', () => p.seite(`return !!${trenner}`), 20000)
}

// mass liest Trenner, Spalte und Zentrum.
const mass = `const t = ${trenner}; const s = ${spalteVon('t')}; const z = t.nextElementSibling; return { wert: Number(t.getAttribute('aria-valuenow')), hoechstens: Number(t.getAttribute('aria-valuemax')), spalte: Math.round(s.getBoundingClientRect().width), zentrum: Math.round(z.getBoundingClientRect().width), name: t.getAttribute('aria-label') }`

async function trennerweg(p, bild, ort) {
  await oeffneSet(p)
  const vorher = await p.seite(mass)
  p.pruefe(`${bild}: zwischen Liste und Vorschau steht ein Trenner mit Wert`, vorher.wert > 0 && Math.abs(vorher.spalte - vorher.wert) <= 1, JSON.stringify(vorher))

  await p.seite(`${trenner}.focus(); return true`)
  await p.taste('ArrowRight')
  const rechts = await p.warte('→ macht die Spalte 16 Punkte breiter', async () => { const m = await p.seite(mass); return m.wert === vorher.wert + 16 && Math.abs(m.spalte - m.wert) <= 1 ? m : null }, 10000)
  p.pruefe(`${bild}: → geht 16 Punkte`, rechts !== null, JSON.stringify(rechts))

  await p.taste('End')
  const ende = await p.warte('Ende geht an die Grenze', async () => { const m = await p.seite(mass); return m.wert === m.hoechstens && Math.abs(m.spalte - m.wert) <= 1 ? m : null }, 10000)
  p.pruefe(`${bild}: Ende geht an die Grenze, und dem Zentrum bleiben 480 Punkte`, ende !== null && ende.zentrum >= 480, JSON.stringify(ende))
  await p.bild(`trenner-grenze-${bild}`)

  await p.taste('Home')
  await p.warte('Pos1 geht an 120', async () => (await p.seite(mass)).wert === 120, 10000)
  // **Und bleibt dort** — ein Schreiben, das scheiterte, nähme den Wunsch
  // zurück; gemessen über eine Sekunde, bevor der Zug ansetzt.
  const verlauf = []
  for (let schritt = 0; schritt < 5; schritt += 1) {
    verlauf.push((await p.seite(mass)).wert)
    await new Promise((weiter) => { setTimeout(weiter, 250) })
  }
  p.pruefe(`${bild}: Pos1 bleibt bei 120`, verlauf.every((wert) => wert === 120), verlauf.join(' → '))
  const lage = await p.seite(`const r = ${trenner}.getBoundingClientRect(); return { x: Math.round(r.left + r.width / 2), y: Math.round(r.top + r.height / 2) }`)
  await ziehe(p, [lage, { x: lage.x + 40, y: lage.y }, { x: lage.x + 80, y: lage.y }])
  let gezogen = null
  try {
    gezogen = await p.warte('der Zug macht die Spalte breiter', async () => { const m = await p.seite(mass); return m.wert >= 190 && m.wert <= 210 && Math.abs(m.spalte - m.wert) <= 1 ? m : null }, 10000)
  } catch {
    p.zeige(`nach dem Zug: ${JSON.stringify(await p.seite(mass))}, Ansatz ${JSON.stringify(lage)}, Fehler am Arbeitsplatz: ${JSON.stringify(await p.seite("return [...document.querySelectorAll('main [role=alert]')].map((e) => e.innerText.trim()).filter(Boolean)"))}`)
  }
  p.pruefe(`${bild}: ein Zug mit der Maus um 80 Punkte macht die Spalte breiter`, gezogen !== null, JSON.stringify(gezogen ?? (await p.seite(mass))))
  if (gezogen === null) return

  const gemerkt = await p.warte('mse einstellungen kennt die Breite', () => {
    const eintrag = JSON.parse(p.mse(ort, 'einstellungen', '--json')).einstellungen.find((e) => e.schluessel === 'arbeitsplatz.navigatorbreite')
    return eintrag?.wert === String(gezogen?.wert) ? eintrag.wert : null
  }, 10000)
  p.pruefe(`${bild}: mse einstellungen kennt die gezogene Breite`, gemerkt !== null, String(gemerkt))
  await p.bild(`trenner-gezogen-${bild}`)
  p.pruefe(`${bild}: ⌘Q beendet`, await p.beende())

  await p.starte({ ort })
  await p.groesse(1440, 838)
  await oeffneSet(p)
  const danach = await p.seite(mass)
  p.pruefe(`${bild}: nach dem Neustart steht die Spalte so breit wieder da`, danach.wert === gezogen?.wert && Math.abs(danach.spalte - danach.wert) <= 1, JSON.stringify(danach))
}

async function symbol(p, bild) {
  await zurSeite(p, 'symbol')
  const knopf = "[...document.querySelectorAll('[role=dialog] button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Größer')"
  await p.warte('„Größer“ steht da', () => p.seite(`return !!(${knopf})`), 15000)
  // **Erst wenn die Schublade steht** — sie fährt ein, und ein Klick auf den
  // Knopf, der noch wandert, trifft den Abschnitt daneben.
  let lage = null
  await p.warte('die Schublade steht still', async () => {
    const jetzt = await p.seite("return Math.round(document.querySelector('.mantine-Drawer-content').getBoundingClientRect().left)")
    const still = jetzt === lage
    lage = jetzt
    if (!still) await new Promise((weiter) => { setTimeout(weiter, 120) })

    return still
  }, 5000)
  const schmal = await p.seite(`return { gedrueckt: (${knopf}).getAttribute('aria-pressed'), schublade: Math.round(document.querySelector('.mantine-Drawer-content').getBoundingClientRect().width), flaeche: Math.round(document.querySelector('svg[role=group]').getBoundingClientRect().width) }`)
  p.pruefe(`${bild}: „Größer“ ist ein Umschalter, zuerst nicht gedrückt`, schmal.gedrueckt === 'false', JSON.stringify(schmal))
  await p.klickeAuf(knopf, '»Größer«')
  const weit = await p.warte('die Schublade geht weit auf', async () => { const m = await p.seite(`return { gedrueckt: (${knopf}).getAttribute('aria-pressed'), schublade: Math.round(document.querySelector('.mantine-Drawer-content').getBoundingClientRect().width), flaeche: Math.round(document.querySelector('svg[role=group]').getBoundingClientRect().width) }`); return m.gedrueckt === 'true' && m.schublade >= 880 ? m : null }, 10000)
  p.pruefe(`${bild}: gedrückt geht die Schublade auf 56rem und die Fläche auf 32rem`, weit !== null && Math.abs(weit.flaeche - 512) <= 2, `${JSON.stringify(schmal)} → ${JSON.stringify(weit)}`)
  await p.bild(`symbol-groesser-${bild}`)

  await p.klickeAuf("[...document.querySelectorAll('[role=dialog] [role=tab]')].find((t) => t.dataset.lasche === 'setangaben')", 'die Lasche »Setangaben«')
  await p.warte('die Setangaben stehen da', () => p.seite("return [...document.querySelectorAll('[role=dialog] input')].some((e) => e.checkVisibility())"), 15000)
  const rahmen = await p.seite("const feld = document.querySelector('[role=dialog] [role=tabpanel]'); const kaesten = [...feld.querySelectorAll('*')].filter((e) => { const s = getComputedStyle(e); return e.checkVisibility() && !e.matches('input, textarea, select, button') && s.borderTopWidth !== '0px' && s.borderLeftWidth !== '0px' && s.borderRightWidth !== '0px' && e.getBoundingClientRect().width > feld.getBoundingClientRect().width * 0.9 }); return kaesten.map((e) => e.className.toString().slice(0, 60))")
  p.pruefe(`${bild}: die Setangaben stehen ohne eigenen Rahmen in der Schublade`, rahmen.length === 0, rahmen.join(' | ') || 'keiner')
  const schublade = await p.seite("return Math.round(document.querySelector('.mantine-Drawer-content').getBoundingClientRect().width)")
  p.pruefe(`${bild}: bei einer anderen Lasche ist die Schublade wieder schmal`, schublade < 500, String(schublade))
  await p.bild(`setangaben-${bild}`)
  await p.taste('Escape')
}

const paketmass = `return [...document.querySelectorAll('[role=treegrid] tbody tr')].filter((z) => z.cells.length >= 4 && z.checkVisibility()).map((z) => { const fassung = z.cells[2].getBoundingClientRect(); const chip = z.cells[3].querySelector('span')?.getBoundingClientRect(); const text = document.createRange(); text.selectNodeContents(z.cells[2]); const inhalt = text.getBoundingClientRect(); return { name: z.cells[0].innerText.split('\\n')[0].slice(0, 30), fassungRechts: Math.round(inhalt.right), chipLinks: chip ? Math.round(chip.left) : null, ueberlappt: chip ? inhalt.right > chip.left + 0.5 : false } })`

async function pakete(p, bild) {
  for (const [breite, hoehe] of [[1440, 838], [1024, 700]]) {
    await p.groesse(breite, hoehe)
    await zurSeite(p, 'pakete')
    await p.warte('die Pakettabelle steht', () => p.seite("return document.querySelectorAll('[role=treegrid] tbody tr').length > 2"), 20000)
    const zeilen = await p.seite(paketmass)
    const schlecht = zeilen.filter((z) => z.ueberlappt)
    // **Ein Name bricht nur an einem Bindestrich oder Leerzeichen** — jedes
    // Wortstück steht auf einer Zeile. Ein Umbruch mitten im Wort hieß, dass
    // die Spalten den Platz falsch verteilen (am Fenster: „probe-spielskri/pt“).
    const zerrissen = await p.seite("return [...document.querySelectorAll('[role=treegrid] tbody tr')].filter((z) => z.cells.length >= 4 && z.checkVisibility()).map((z) => z.cells[0].querySelector('span')).filter((n) => { const knoten = n?.firstChild; if (!knoten || knoten.nodeType !== 3) return false; const text = knoten.textContent; let anfang = 0; for (const stueck of text.split(/(?<=[-\\s])/)) { const r = document.createRange(); r.setStart(knoten, anfang); r.setEnd(knoten, anfang + stueck.replace(/[-\\s]$/, '').length); anfang += stueck.length; if (r.getClientRects().length > 1) return true } return false }).map((n) => n.innerText)")
    p.pruefe(`${bild}, ${String(breite)} × ${String(hoehe)}: kein Paketname bricht mitten im Wort`, zerrissen.length === 0, zerrissen.join(' · ') || 'keiner')
    p.pruefe(`${bild}, ${String(breite)} × ${String(hoehe)}: keine Fassung läuft in den Chip daneben (${String(zeilen.length)} Zeilen)`, zeilen.length > 0 && schlecht.length === 0, schlecht.length === 0 ? `${zeilen[0]?.name}: ${String(zeilen[0]?.fassungRechts)} < ${String(zeilen[0]?.chipLinks)}` : JSON.stringify(schlecht.slice(0, 3)))
    await p.bild(`pakete-${String(breite)}-${bild}`)
  }
  await p.groesse(1440, 838)
}

// gitter ist die Liste der Sammlung — die sichtbare: besuchte Seiten bleiben
// verborgen im Baum, und die Kartenliste ist auch ein Gitter.
const gitter = "[...document.querySelectorAll('main [role=grid][aria-label=\"Exemplare\"]')].find((g) => g.checkVisibility())"

async function sammlung(p, bild) {
  await zurSeite(p, 'sammlung')
  const inspector = "document.querySelector('aside[aria-label=\"Exemplar\"]')"
  await p.warte('die Sammlung steht', () => p.seite(`return !!(${inspector}) && !!document.querySelector('main [role=gridcell][title]')`), 20000)
  // **Die Bezeichnung steht ganz da** — umgebrochen statt gekürzt; und eine
  // Auswahl schiebt die Liste nicht (der erste Anlauf ließ den leeren
  // Inspector schmal stehen, und das Kontextmenü ging dann zu).
  // **Gegenprobe** (`MSE_AP12_GEGENPROBE=1`): die Bezeichnungen brechen um wie
  // im ersten Schnitt, und die Köpfe teilen sich den Platz wie vorher — alle
  // drei Prüfungen darunter müssen dann rot werden.
  if (process.env.MSE_AP12_GEGENPROBE === '1') await p.seite(`const g = ${gitter}; for (const e of g.querySelectorAll('[role=gridcell][title] .truncate')) { e.classList.remove('truncate'); e.style.overflowWrap = 'anywhere' } for (const e of g.querySelectorAll('[role=gridcell][title] > span:nth-child(2)')) e.style.display = 'none'; for (const e of g.querySelectorAll('[role=gridcell].hidden')) e.style.display = 'block'; for (const k of g.querySelectorAll('[role=columnheader]')) k.style.cssText = 'flex: 1 1 0; width: auto; display: block'; return true`)
  const gekuerzt = "return [...(" + gitter + ").querySelectorAll('[role=gridcell][title] > span:first-child > .truncate')].filter((e) => e.checkVisibility() && (e.scrollWidth > e.clientWidth + 1 || e.getBoundingClientRect().width < 8)).map((e) => e.innerText)"
  const vorher = await p.seite(`return Math.round((${inspector}).getBoundingClientRect().width)`)
  const kurz = await p.seite(gekuerzt)
  const zahl = await p.seite("return [...(" + gitter + ").querySelectorAll('[role=gridcell][title] > span:first-child > .truncate')].filter((e) => e.checkVisibility()).length")
  p.zeige(`Breite der Liste: ${String(await p.seite(`return Math.round((${gitter}).getBoundingClientRect().width)`))} Punkte`)
  p.pruefe(`${bild}: jede Bezeichnung steht ganz da (${String(zahl)} Zeilen)`, zahl >= 5 && kurz.length === 0, kurz.join(' · ') || 'alle')
  // **Keine Zeile läuft in ihre Nachbarn**: jeder sichtbare Knoten einer
  // Zeile bleibt in ihren 40 Punkten.
  const ueber = await p.seite("const raus = []; for (const z of (" + gitter + ").querySelectorAll('[role=row]')) { if (!z.checkVisibility()) continue; const r = z.getBoundingClientRect(); for (const e of z.querySelectorAll('*')) { if (!e.checkVisibility()) continue; const k = e.getBoundingClientRect(); if (k.height > 0 && (k.top < r.top - 1 || k.bottom > r.bottom + 1)) raus.push((e.innerText || e.tagName).slice(0, 30) + ' ' + Math.round(k.top - r.top) + '…' + Math.round(k.bottom - r.top)) } } return raus")
  p.pruefe(`${bild}: keine Zeile der Sammlung läuft in ihre Nachbarn`, ueber.length === 0, ueber.slice(0, 4).join(' | ') || 'keine')
  // **Jeder Kopf steht über seiner Zelle** — die Köpfe waren gleich breit,
  // die Zellen nicht, und „Menge“ stand über der Bezeichnung.
  const versatz = await p.seite("const g = " + gitter + "; const koepfe = [...g.querySelectorAll('[role=columnheader]')].filter((e) => e.checkVisibility()); const zeile = [...g.querySelectorAll('[role=row]')].find((z) => z.querySelector('[role=gridcell][title]')); const zellen = [...zeile.querySelectorAll('[role=gridcell]')].filter((e) => e.checkVisibility()); return { koepfe: koepfe.length, zellen: zellen.length, schief: koepfe.map((k, i) => [k.innerText, Math.round(Math.abs(k.getBoundingClientRect().left - (zellen[i]?.getBoundingClientRect().left ?? -999)))]).filter(([, d]) => d > 1) }")
  p.pruefe(`${bild}: jeder Kopf der Sammlung steht über seiner Zelle (${String(versatz.koepfe)} Köpfe)`, versatz.koepfe === versatz.zellen && versatz.schief.length === 0, JSON.stringify(versatz))
  await p.klickeAuf("[...document.querySelectorAll('main [role=gridcell][title]')].find((z) => z.checkVisibility())", 'das erste Exemplar')
  await new Promise((weiter) => { setTimeout(weiter, 500) })
  const nachher = await p.seite(`return Math.round((${inspector}).getBoundingClientRect().width)`)
  p.pruefe(`${bild}: eine Auswahl schiebt die Liste nicht`, nachher === vorher, `${String(vorher)} → ${String(nachher)}`)
  await p.bild(`sammlung-${bild}`)
}

// linien sucht waagerechte Linien, die weniger als 4 Punkte untereinander
// liegen und sich zur Hälfte überdecken.
const linien = `const l = []; for (const e of document.querySelectorAll('main *')) { if (!e.checkVisibility()) continue; const s = getComputedStyle(e); const r = e.getBoundingClientRect(); if (r.width < 40) continue; for (const [seite, y] of [['Top', r.top], ['Bottom', r.bottom]]) { if (s['border' + seite + 'Style'] !== 'none' && parseFloat(s['border' + seite + 'Width']) >= 1 && !/rgba\\(.*, 0\\)/.test(s['border' + seite + 'Color'])) l.push({ y, links: r.left, rechts: r.right, wer: e.tagName.toLowerCase() + '.' + e.className.toString().split(' ').slice(0, 3).join('.') }) } } const paare = []; for (let i = 0; i < l.length; i++) for (let j = i + 1; j < l.length; j++) { const a = l[i], b = l[j]; const ueber = Math.min(a.rechts, b.rechts) - Math.max(a.links, b.links); if (Math.abs(a.y - b.y) > 0.5 && Math.abs(a.y - b.y) < 4 && ueber > 0.5 * Math.min(a.rechts - a.links, b.rechts - b.links)) paare.push(a.wer + ' / ' + b.wer + ' @' + Math.round(a.y)) } return paare`

async function diagnose(p, bild) {
  await p.taste(',', 'befehl')
  const lasche = "[...document.querySelectorAll('main [role=tab]')].find((t) => t.checkVisibility() && t.innerText.split('\\n')[0].trim() === 'Diagnose')"
  await p.warte('die Einstellungen stehen', () => p.seite(`return !!(${lasche})`), 20000)
  await p.klickeAuf(lasche, 'die Gruppe »Diagnose«')
  await new Promise((weiter) => { setTimeout(weiter, 500) })
  // **Gegenprobe** (`MSE_AP12_GEGENPROBE=1`): zwei Linien drei Punkte
  // untereinander — die Prüfung muss sie finden, sonst sagt ihr Grün nichts.
  if (process.env.MSE_AP12_GEGENPROBE === '1') await p.seite("const m = document.querySelector('main'); for (const seite of ['bottom', 'top']) { const d = document.createElement('div'); d.style.cssText = 'border-' + seite + ': 1px solid currentColor; height: 1px; margin: 0;'; m.prepend(d) } return true")
  const paare = await p.seite(linien)
  p.pruefe(`${bild}: in der Diagnose stehen keine zwei Trennlinien direkt untereinander`, paare.length === 0, paare.slice(0, 4).join(' | ') || 'keine')
  await p.bild(`diagnose-${bild}`)
}

async function lauf(p, bild) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  // **Beide Spalten stehen** — dazwischen ist die Sammlung schmal.
  p.mse(ort, 'einstellungen', 'setzen', 'arbeitsplatz.navigator', 'sichtbar')
  p.mse(ort, 'einstellungen', 'setzen', 'arbeitsplatz.inspektor', 'sichtbar')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  for (const druck of ['11111111-2222-4333-8444-555555555555', '66666666-7777-4888-9999-aaaaaaaaaaaa']) p.mse(ort, 'exemplar', druck, '2')
  // **Namen wie im Handbuchbild** — mit ihnen liefen dort die Zeilen ineinander.
  for (const zeilen of [
    ['counterspell-7ed', '4', 'nearMint', 'de', 'nonfoil', '2026-08-01', '1.20', 'EUR', 'getauscht', 'Counterspell', '7ed', '67', 'Kiste 1'],
    ['lightning-bolt-m10', '3', 'nearMint', 'de', 'nonfoil', '2026-06-03', '2.10', 'EUR', 'gekauft', 'Lightning Bolt', 'm10', '146', 'Kiste 2'],
    ['llanowar-elves-m19', '4', 'played', 'en', 'nonfoil', '2026-05-20', '0.30', 'EUR', 'Booster', 'Llanowar Elves', 'm19', '314', 'Kiste 2'],
  ]) p.mse(ort, 'exemplar', ...zeilen)
  await p.starte({ ort })
  await p.groesse(1440, 838)

  p.notiz(`${bild}: der Trenner`)
  await trennerweg(p, bild, ort)
  p.notiz(`${bild}: das Symbol, größer, und die Setangaben`)
  await symbol(p, bild)
  p.notiz(`${bild}: die Pakettabelle`)
  await pakete(p, bild)
  p.notiz(`${bild}: die Sammlung`)
  await sammlung(p, bild)
  p.notiz(`${bild}: die Diagnose`)
  await diagnose(p, bild)
  p.pruefe(`${bild}: ⌘Q beendet`, await p.beende())
}

export default async function ap12feinschliff09(p) {
  for (const bild of ['hell', 'dunkel']) await lauf(p, bild)
}
