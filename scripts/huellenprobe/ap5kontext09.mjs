// **Kürzel, Fenstertitel und Kontextmenüs — am nativen Fenster** — AP-7
// (#2364)
// und AP-5 (#2362),
// Artboard `Kontext09`.
//
// # Was es prüft
//
// **AP-7**: der Fenstertitel nennt das offene Set; das Menü **„Karte"** der
// Leiste trägt ⌘↩, ⌘D, ⌘⌫, ⇞ und ⇟; ⌘⌫ **in einem Feld** löscht Text und
// keine Karte; außerhalb legt ⌘↩ an, verdoppelt ⌘D, fragt ⌘⌫ zurück, und ⇟
// und ⇞ blättern — gezählt über `mse karten`, nicht an der Oberfläche.
//
// **AP-5**: Rechtsklick an einer **anderen** Karte der Liste wählt sie und
// zeigt **ihr** Menü; ⇧F10 an der Zeile öffnet es, Escape gibt den Fokus an
// sie zurück; ein Wechsel des Objekts über die Leiste schließt ein offenes
// Menü. Dazu Rechtsklick an Setkarte, Paketzeile (und **kein** Menü an einer
// Gruppe), Vorlagendatei (und keines am Paket im Baum) und Partienzeile.
// Alles hell und dunkel.
//
// # Was es nicht belegt
//
// **Die Leserrolle** am Fenster — sie prüfen `paketrechtsklick.test.tsx`,
// `setkarterechtsklick.test.tsx` und `objektkontextmenue.test.tsx`.
// **VoiceOver** hört ein Mensch.
import { rechtsklick } from '../huellenrufbuch.mjs'

const menueOffen = "return [...document.querySelectorAll('[role=menu]')].some((m) => m.checkVisibility())"
const sichtbareEintraege = "return [...document.querySelectorAll('[role=menu] [role=menuitem]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.replace(/\\s+/g, ' ').trim())"
const menueObjekt = "return [...document.querySelectorAll('[role=menu]')].find((m) => m.checkVisibility())?.closest('[data-objekt]')?.getAttribute('data-objekt') ?? null"
// ganzZuSehen prüft das Menü an seinen vier Ecken: liegt dort das Menü oben,
// schneidet es weder der Fensterrand noch ein Vorfahr mit `overflow` ab —
// der erste Lauf maß nur gegen das Fenster und übersah eine Spalte, die
// »Karte anleg…« abschnitt.
const ganzZuSehen = "const m = [...document.querySelectorAll('[role=menu]')].find((m) => m.checkVisibility()); if (!m) return null; const r = m.getBoundingClientRect(); return [[r.left + 3, r.top + 3], [r.right - 3, r.top + 3], [r.left + 3, r.bottom - 3], [r.right - 3, r.bottom - 3]].every(([x, y]) => m.contains(document.elementFromPoint(x, y)))"
const mitteVon = (ausdruck) => `const e = ${ausdruck}; if (!e) return null; e.scrollIntoView({ block: 'center' }); const r = e.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2 }`
const gewaehlteKarte = "return document.querySelector('[role=grid] [role=row][aria-selected=true]')?.getAttribute('data-karte') ?? null"
const zeileVon = (id) => `document.querySelector('[role=grid] [role=row][data-karte="${id}"]')`

// zeichen schreibt ein Äquivalent der Leiste so, wie macOS es zeigt.
const zeichen = (kuerzel) => (kuerzel ?? '').replace('\r', '↩').replace('\b', '⌫')

const leistenmenue = async (p, titel) => (await p.menue()).find((m) => m.titel === titel)

async function zuSeite(p, name, merkmal) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"), 10000)
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld(name)
  await p.klickeAuf(`[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(name)})`, `»${name}« in der Palette`)
  await p.wartetext(merkmal, 20000)
}

// kontext klickt rechts auf ein Element und gibt Objekt und Einträge des
// Menüs, das aufgeht — oder null, wenn keines aufgeht.
async function kontext(p, ausdruck, wo, { erwartet = true, bild = '' } = {}) {
  // **Sichtbar, nicht nur vorhanden** — `Activity` hält besuchte Seiten
  // verborgen im Baum, und ein Klick auf (0, 0) träfe nichts.
  await p.warte(`${wo} ist sichtbar`, () => p.seite(`return !!(${ausdruck})?.checkVisibility()`), 15000).catch(() => false)
  const stelle = await p.seite(mitteVon(ausdruck))
  p.pruefe(`${wo} steht da`, !!stelle, JSON.stringify(stelle))
  if (!stelle) return null
  await rechtsklick(p, stelle.x, stelle.y)
  if (!erwartet) {
    await new Promise((fertig) => setTimeout(fertig, 900))
    p.pruefe(`an ${wo} geht kein Menü der Seite auf`, !(await p.seite(menueOffen)))
    await p.taste('Escape')

    return null
  }
  await p.warte(`der Rechtsklick an ${wo} öffnet ein Menü`, () => p.seite(menueOffen), 10000)
  const ergebnis = { objekt: await p.seite(menueObjekt), eintraege: await p.seite(sichtbareEintraege) }
  p.pruefe(`das Menü an ${wo} ist ganz zu sehen`, (await p.seite(ganzZuSehen)) === true)
  await p.bild(`kontext-${wo.replace(/[^\p{L}\p{N}]+/gu, '-')}-${bild}`)
  await p.taste('Escape')
  await p.warte('Escape schließt es', async () => !(await p.seite(menueOffen)), 10000)

  return ergebnis
}

async function lauf(p, bild) {
  // **Ein eigenes Zuhause**: die Werkstatt legt eine eigene Kopie eines
  // Pakets an, und die gehört nicht in die Pakete dieses Rechners.
  const ort = p.ort({ zuhause: true })
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Probedeck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')
  p.mse(ort, 'partie-anlegen', 'probe', 'ada:Ada:1', 'bo:Bo:1')
  const karten = () => JSON.parse(p.mse(ort, 'karten', '1', '--json')).karten

  await p.starte({ ort })
  await p.groesse(1440, 900)

  p.notiz(`${bild}: der Fenstertitel ohne und mit offenem Set`)
  const ohne = (await p.fenster()).titel
  p.pruefe('ohne offenes Set heißt das Fenster wie die Anwendung', ohne === 'MagicSetEditor4Mac', ohne)
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await p.warte('die Karte steht mit Bild', () => p.seite("return !!document.querySelector('img[alt=\"Vorschau der Karte\"]')?.checkVisibility()"), 30000)
  const mit = await p.warte('der Titel nennt das Set', async () => {
    const titel = (await p.fenster()).titel

    return titel === 'Werkstatt-Testkarten — MagicSetEditor4Mac' ? titel : null
  }, 10000).catch(async () => (await p.fenster()).titel)
  p.pruefe('mit offenem Set nennt der Titel das Set vor der Anwendung', mit === 'Werkstatt-Testkarten — MagicSetEditor4Mac', mit)

  p.notiz(`${bild}: die Kürzel im Menü »Karte«`)
  const karte = await p.warte('die Leiste führt »Karte«', () => leistenmenue(p, 'Karte'), 10000)
  const kuerzel = Object.fromEntries((karte.unter ?? []).filter((e) => !e.trenner).map((e) => [e.titel, zeichen(e.kuerzel)]))
  for (const [titel, erwartet] of [['Karte anlegen', '⌘↩'], ['Karte duplizieren', '⌘D'], ['Karte löschen …', '⌘⌫'], ['Vorige Karte', '⇞'], ['Nächste Karte', '⇟']]) {
    const gefunden = Object.entries(kuerzel).find(([name]) => name.startsWith(titel.replace(' …', '')))
    p.pruefe(`»${titel}« trägt ${erwartet}`, gefunden?.[1] === erwartet, JSON.stringify(gefunden ?? kuerzel))
  }
  await p.leistenbild(`menue-karte-${bild}`)

  p.notiz(`${bild}: ⌘⌫ in einem Feld löscht Text, keine Karte`)
  const zahl = karten().length
  // **Das Namensfeld** — seine Eingabe trägt den Feldschlüssel als Namen
  // (`Feldeingabe.tsx`); ⌘⌫ braucht Text, den es nehmen kann.
  const feld = "[...document.querySelectorAll('input[aria-label=name], textarea[aria-label=name]')].find((e) => e.checkVisibility())"
  p.pruefe('das Namensfeld steht da', await p.seite(`return !!${feld}`))
  await p.seite(`const e = ${feld}; e.focus(); e.setSelectionRange(e.value.length, e.value.length); return true`)
  await p.tippeInsFeld(' Probe')
  const vorherText = await p.seite(`return ${feld}.value`)
  p.pruefe('das Feld hat den Fokus und Text', vorherText.endsWith('Probe') && (await p.seite(`return document.activeElement === ${feld}`)), JSON.stringify(vorherText))
  await p.taste('Backspace', 'befehl')
  await new Promise((fertig) => setTimeout(fertig, 1200))
  const nachherText = await p.seite(`return ${feld}.value`)
  p.pruefe('⌘⌫ im Feld nimmt Text weg', nachherText !== vorherText, `${JSON.stringify(vorherText)} → ${JSON.stringify(nachherText)}`)
  p.pruefe('⌘⌫ im Feld lässt die Karten stehen', karten().length === zahl, `${String(zahl)} → ${String(karten().length)}`)
  p.pruefe('⌘⌫ im Feld stellt keine Löschfrage', !(await p.seite("return document.body.innerText.includes('Diese Karte löschen?')")))

  p.notiz(`${bild}: ⌘D, ⌘↩, ⇟, ⇞ und ⌘⌫ außerhalb der Felder`)
  await p.seite('document.activeElement?.blur(); return true')
  await p.taste('d', 'befehl')
  await p.warte('⌘D verdoppelt die Karte', () => karten().length === zahl + 1, 15000).catch(() => false)
  p.pruefe('⌘D verdoppelt die Karte', karten().length === zahl + 1, `${String(zahl)} → ${String(karten().length)}`)
  await p.seite('document.activeElement?.blur(); return true')
  await p.taste('Enter', 'befehl')
  await p.warte('⌘↩ legt eine Karte an', () => karten().length === zahl + 2, 15000).catch(() => false)
  p.pruefe('⌘↩ legt eine Karte an', karten().length === zahl + 2, `${String(zahl + 1)} → ${String(karten().length)}`)

  await p.seite('document.activeElement?.blur(); return true')
  const vorBlaettern = await p.seite(gewaehlteKarte)
  await p.taste('PageUp')
  const nachOben = await p.warte('⇞ wählt eine andere Karte', async () => {
    const jetzt = await p.seite(gewaehlteKarte)

    return jetzt !== vorBlaettern ? jetzt : null
  }, 10000).catch(() => null)
  p.pruefe('⇞ blättert zur vorigen Karte', nachOben !== null, `${String(vorBlaettern)} → ${String(nachOben)}`)
  await p.taste('PageDown')
  const nachUnten = await p.warte('⇟ wählt wieder die von vorhin', async () => ((await p.seite(gewaehlteKarte)) === vorBlaettern ? vorBlaettern : null), 10000).catch(() => null)
  p.pruefe('⇟ blättert zur nächsten Karte', nachUnten === vorBlaettern, `${String(nachOben)} → ${String(await p.seite(gewaehlteKarte))}`)

  const vorLoeschen = karten().length
  await p.seite('document.activeElement?.blur(); return true')
  await p.taste('Backspace', 'befehl')
  await new Promise((fertig) => setTimeout(fertig, 1200))
  p.pruefe('⌘⌫ außerhalb löscht nicht ohne Rückfrage', karten().length === vorLoeschen, `${String(vorLoeschen)} → ${String(karten().length)}`)
  const frage = await p.warte('die Löschfrage steht', () => p.seite("return document.body.innerText.includes('Diese Karte löschen?')"), 10000).catch(() => false)
  p.pruefe('⌘⌫ außerhalb stellt die Löschfrage', frage === true)
  await p.bild(`loeschfrage-${bild}`)
  await p.taste('Escape')

  p.notiz(`${bild}: Rechtsklick an einer anderen Karte der Liste`)
  const andere = await p.seite("return [...document.querySelectorAll('[role=grid] [role=row][aria-selected=false]')].find((z) => z.checkVisibility())?.getAttribute('data-karte') ?? null")
  const amListeneintrag = await kontext(p, zeileVon(andere), 'einer anderen Karte der Liste', { bild })
  p.pruefe('der Rechtsklick wählt die andere Karte', (await p.seite(gewaehlteKarte)) === andere, String(await p.seite(gewaehlteKarte)))
  p.pruefe('das Menü gehört ihr', amListeneintrag?.objekt === `karte/${andere}`, String(amListeneintrag?.objekt))
  p.pruefe('es nennt Anlegen, Verdoppeln und Löschen', ['anlegen', 'duplizieren', 'löschen'].every((wort) => amListeneintrag?.eintraege.some((e) => e.toLowerCase().includes(wort))), amListeneintrag?.eintraege.join(' · '))

  // **Die unterste sichtbare Zeile** — dort rückt das Menü nach oben, statt
  // über den Rand zu laufen.
  const unten = "[...document.querySelectorAll('[role=grid] [role=row]')].filter((z) => { const r = z.getBoundingClientRect(); return z.checkVisibility() && r.bottom <= innerHeight }).at(-1)"
  const amUnteren = await kontext(p, unten, 'der untersten Zeile', { bild })
  p.pruefe('auch dort gehört das Menü der Karte darunter', amUnteren?.objekt?.startsWith('karte/') === true, String(amUnteren?.objekt))

  p.notiz(`${bild}: ⇧F10 an der Zeile, Escape gibt ihr den Fokus zurück`)
  await p.seite(`${zeileVon(andere)}.focus(); return true`)
  await p.taste('F10', 'umschalt')
  await p.warte('⇧F10 öffnet das Menü der Zeile', () => p.seite(menueOffen), 10000)
  p.pruefe('⇧F10: der Fokus steht im Menü', (await p.seite("return document.activeElement?.getAttribute('role')")) === 'menuitem')
  await p.taste('Escape')
  await p.warte('Escape schließt es', async () => !(await p.seite(menueOffen)), 10000)
  p.pruefe('Escape gibt der Zeile den Fokus zurück', (await p.seite("return document.activeElement?.getAttribute('data-karte')")) === andere, String(await p.seite("return document.activeElement?.outerHTML.slice(0, 80)")))

  p.notiz(`${bild}: ein Wechsel über die Leiste schließt ein offenes Menü`)
  await p.seite(`${zeileVon(andere)}.focus(); return true`)
  await p.taste('F10', 'umschalt')
  await p.warte('das Menü ist offen', () => p.seite(menueOffen), 10000)
  await p.klick('Karte', 'Nächste Karte')
  await p.warte('der Wechsel schließt das Menü', async () => !(await p.seite(menueOffen)), 10000).catch(() => false)
  p.pruefe('ein Objektwechsel schließt das Menü', !(await p.seite(menueOffen)), String(await p.seite(gewaehlteKarte)))

  p.notiz(`${bild}: Rechtsklick an einer Setkarte`)
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  const amSet = await kontext(p, "[...document.querySelectorAll('[data-objekt^=\"set/\"] p[title]')].find((e) => e.checkVisibility())", 'einer Setkarte', { bild })
  p.pruefe('das Menü der Setkarte gehört dem Set', amSet?.objekt?.startsWith('set/') === true, String(amSet?.objekt))
  p.pruefe('es nennt Sichern und Löschen', ['sichern', 'löschen'].every((wort) => amSet?.eintraege.some((e) => e.toLowerCase().includes(wort))), amSet?.eintraege.join(' · '))

  p.notiz(`${bild}: Rechtsklick an einer Paketzeile und einer Gruppe`)
  await p.taste('6', 'befehl')
  await p.warte('die Pakettabelle steht', () => p.seite("return !!document.querySelector('[role=treegrid] tbody tr[aria-selected]')"), 20000)
  const paketzeile = "[...document.querySelectorAll('[role=treegrid] tbody tr[aria-selected]')].find((z) => z.checkVisibility() && z.getAttribute('aria-selected') === 'false')"
  const paketname = await p.seite(`return ${paketzeile}?.getAttribute('aria-label') ?? null`)
  const amPaket = await kontext(p, paketzeile, 'einer Paketzeile', { bild })
  p.pruefe('das Menü der Paketzeile gehört ihrem Paket', amPaket?.objekt?.endsWith(`/${paketname}`) === true, `${String(paketname)} → ${String(amPaket?.objekt)}`)
  p.pruefe('der Rechtsklick wählt das Paket', (await p.seite(`return [...document.querySelectorAll('[role=treegrid] tbody tr[aria-selected=true]')].map((z) => z.getAttribute('aria-label'))`))?.includes(paketname), paketname)
  await kontext(p, "[...document.querySelectorAll('[role=treegrid] tbody tr[aria-expanded]')].find((z) => z.checkVisibility())", 'einer Paketgruppe', { erwartet: false })

  p.notiz(`${bild}: Rechtsklick an einer Vorlagendatei und am Paket im Baum`)
  await zuSeite(p, 'Vorlagen', 'probe-schlicht.mse-style')
  // **Ein mitgeliefertes Paket zeigt keine Dateien** — nur das Angebot einer
  // eigenen Kopie. Die Probe legt sie an, wie `fmt02`.
  const zweig = (name) => `[...document.querySelectorAll('button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith(${JSON.stringify(name)}))`
  await p.klickeAuf(`[...(${zweig('probe-schlicht.mse-style')})?.closest('li')?.querySelectorAll('button') ?? []].find((b) => b.innerText.trim().endsWith('Eigene Kopie anlegen'))`, 'Eigene Kopie anlegen unter probe-schlicht')
  await p.wartetext('Das Original bleibt.', 15000)
  const kopie = zweig('probe-schlicht-kopie.mse-style')
  if (await p.seite(`return ${kopie}?.getAttribute('aria-expanded') === 'false'`)) await p.klickeAuf(kopie, 'den Zweig der Kopie')
  const datei = `[...(${kopie}?.closest('li')?.querySelectorAll(':scope > ul button[data-baumknoten]') ?? [])].find((b) => b.checkVisibility() && b.getAttribute('aria-current') !== 'true')`
  await p.warte('eine zweite Datei der Kopie steht', () => p.seite(`return !!${datei}`), 20000)
  const dateiname = await p.seite(`return ${datei}.innerText.replace(/^[^\\p{L}\\p{N}]+/u, '').trim().split(/\\s/)[0]`)
  const anDatei = await kontext(p, datei, 'einer Vorlagendatei', { bild })
  p.pruefe('das Menü gehört der Datei', anDatei?.objekt?.endsWith(`/${dateiname}`) === true, `${dateiname} → ${String(anDatei?.objekt)}`)
  p.pruefe('der Rechtsklick hat sie geöffnet', (await p.seite(`return [...(${kopie}?.closest('li')?.querySelectorAll('[aria-current=true]') ?? [])].map((b) => b.innerText.replace(/^[^\\p{L}\\p{N}]+/u, '').trim().split(/\\s/)[0])`))?.includes(dateiname), dateiname)
  p.pruefe('es nennt »Neue Datei« und »Umbenennen«', ['neue datei', 'umbenennen'].every((wort) => anDatei?.eintraege.some((e) => e.toLowerCase().includes(wort))), anDatei?.eintraege.join(' · '))
  await kontext(p, kopie, 'einem Paket im Baum', { erwartet: false })

  p.notiz(`${bild}: Rechtsklick an einer Partienzeile`)
  await zuSeite(p, 'Spielen', 'Deine Partien')
  const partie = await p.warte('die Partie steht in der Liste', () => p.seite("return [...document.querySelectorAll('li[data-objekt^=\"partie/\"]')].find((z) => z.checkVisibility())?.getAttribute('data-objekt') ?? null"), 20000)
  const anPartie = await kontext(p, `document.querySelector('li[data-objekt="${partie}"] strong')`, 'einer Partienzeile', { bild })
  p.pruefe('das Menü gehört der Partie', anPartie?.objekt === partie, String(anPartie?.objekt))
  p.pruefe('es nennt Weiterspielen, Zuschauen und Duplizieren', ['weiterspielen als', 'zuschauen', 'duplizieren'].every((wort) => anPartie?.eintraege.some((e) => e.toLowerCase().includes(wort))), anPartie?.eintraege.join(' · '))

  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}

export default async function ap5kontext09(p) {
  for (const bild of ['hell', 'dunkel']) await lauf(p, bild)
}
