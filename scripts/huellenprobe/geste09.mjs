// **Eine Handlungsliste, drei Orte — am nativen Fenster** — GESTE
// (#2358),
// Artboard `Kontext09`.
//
// # Was es prüft
//
// **Am Kartenobjekt**: das Menü „Diese Karte" (der Überlauf an der Zeile),
// der Rechtsklick auf das Kartenbild, ⇧F10 dort und das Menü **„Karte"** in
// der Menüleiste nennen **dieselben Einträge** in derselben Reihenfolge — und
// „Karte duplizieren" aus der Leiste legt wirklich eine Karte an
// (`mse karten`). Ein Rechtsklick in ein Feld öffnet keines.
//
// **Am Marktobjekt**: ein Entwurf eines Handels, angelegt wie in
// `druckfeld09`; „Mit diesem Handel", der Rechtsklick auf seine Schritte und
// das Menü **„Handel"** nennen dieselben Einträge, und die Leiste hält das
// Menü „Karte" nicht mehr — **das gewählte Objekt hat die Leiste**.
//
// # Was es nicht belegt
//
// **Den verspäteten Klick** — einen Eintrag der Leiste, der nach einem
// Wechsel ankommt. Die Leiste wird mit jedem Wechsel neu gebaut, bevor ein
// Mensch klicken kann; den Fall prüfen `objekthandlungsraum.test.tsx` und
// `useGeste.test.tsx` gegen eine Laufzeit, die Ereignisse auf Zuruf meldet.
// **Die Leserrolle** prüfen dieselben Tests und `objektkontextmenue.test.tsx`;
// am Fenster belegt sie AP-5 an allen sechs Objektarten. **VoiceOver** hört
// ein Mensch.
import { einfuegen, feld, gekoppelt, knopf } from '../huellenhandel.mjs'
import { rechtsklick } from '../huellenrufbuch.mjs'

const menueOffen = "return [...document.querySelectorAll('[role=menu]')].some((m) => m.checkVisibility())"
const sichtbareEintraege = "return [...document.querySelectorAll('[role=menu] [role=menuitem]')].filter((e) => e.checkVisibility()).map((e) => (e.getAttribute('aria-label') ?? e.innerText).replace(/\\s+/g, ' ').trim())"
const mitteVon = (ausdruck) => `const e = ${ausdruck}; if (!e) return null; e.scrollIntoView({ block: 'center' }); const r = e.getBoundingClientRect(); return { x: r.left + r.width / 2, y: r.top + r.height / 2 }`

// namen ist, was ein Eintrag sagt — ohne Zeichen davor und ohne Kürzel
// dahinter, die das Menü der Seite trägt und die Leiste nicht.
const namen = (liste) => liste.map((name) => name.replace(/^[^\p{L}\p{N}]+/u, '').replace(/\s*[⌘⇧⌥⌃⇞⇟].*$/u, '').trim())

async function seitenmenue(p, aufschrift) {
  await p.klickeAuf(knopf(aufschrift), `»${aufschrift}«`)
  await p.warte(`»${aufschrift}« ist offen`, () => p.seite(menueOffen), 10000)
  const liste = namen(await p.seite(sichtbareEintraege))
  await p.taste('Escape')
  await p.warte('das Menü ist zu', async () => !(await p.seite(menueOffen)), 10000)

  return liste
}

async function kontextmenue(p, ausdruck, wo) {
  const stelle = await p.seite(mitteVon(ausdruck))
  p.pruefe(`${wo} steht da`, !!stelle, JSON.stringify(stelle))
  await rechtsklick(p, stelle.x, stelle.y)
  await p.warte(`der Rechtsklick auf ${wo} öffnet ein Menü`, () => p.seite(menueOffen), 10000)
  const liste = namen(await p.seite(sichtbareEintraege))
  await p.bild(`kontextmenue-${wo.replace(/\s+/g, '-')}`)
  await p.taste('Escape')
  await p.warte('Escape schließt es', async () => !(await p.seite(menueOffen)), 10000)

  return liste
}

const leistenmenue = async (p, titel) => (await p.menue()).find((m) => m.titel === titel)
const leistennamen = (menue) => (menue?.unter ?? []).filter((e) => !e.trenner && !e.verborgen).map((e) => e.titel)
const gleich = (a, b) => JSON.stringify(a) === JSON.stringify(b)

// **Seit #2378 steht die Kennung im Kasten »Kennung von Hand«** — eine Zeile
// entsteht über den Namen der Karte, und wer den Druck tippt, klappt ihn auf.
const kennungskasten = "[...document.querySelectorAll('button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Kennung von Hand')"

export default async function geste09(p) {
  const bert = new p.constructor('geste09-bert')
  bert.zeilen = p.zeilen

  try {
    const ort = p.ort()
    gekoppelt(p, { Anna: ort, Bert: bert.ort() })
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
    p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
    p.mse(ort, 'exemplar', 'p0-blitz', '1', 'nearMint', 'de', 'nonfoil')
    const kartenzahl = () => JSON.parse(p.mse(ort, 'karten', '1', '--json')).karten.length

    await p.starte({ ort })
    await p.groesse(1440, 900)

    p.notiz('Kartenobjekt — ⌘2, das Set, die erste Karte')
    await p.taste('2', 'befehl')
    await p.wartetext('Werkstatt-Testkarten', 20000)
    await p.klickeText('Werkstatt-Testkarten')
    await p.warte('die Karte steht mit Bild', () => p.seite("return !!document.querySelector('img[alt=\"Vorschau der Karte\"]')?.checkVisibility()"), 30000)

    const ueberlauf = await seitenmenue(p, 'Diese Karte')
    p.pruefe('»Diese Karte« nennt Einträge', ueberlauf.length > 3, ueberlauf.join(' · '))
    const amBild = await kontextmenue(p, "document.querySelector('[data-objekt^=\"karte/\"] img[alt=\"Vorschau der Karte\"]')", 'das Kartenbild')
    p.pruefe('der Rechtsklick auf das Bild nennt dieselben Einträge wie »Diese Karte«', gleich(amBild, ueberlauf), amBild.join(' · '))

    // **⇧F10 an der Bühne**: der Fokus steht auf einer Feldfläche über dem
    // Bild; den setzt die Seite, ein Klick fokussiert in WebKit nicht.
    await p.seite("const f = document.querySelector('img[alt=\"Vorschau der Karte\"]')?.parentElement?.querySelector('button, [tabindex]'); (f ?? document.querySelector('img[alt=\"Vorschau der Karte\"]')).focus(); return true")
    await p.taste('F10', 'umschalt')
    await p.warte('⇧F10 öffnet das Menü der Karte', () => p.seite(menueOffen), 10000)
    p.pruefe('⇧F10: der Fokus steht im Menü', (await p.seite("return document.activeElement?.getAttribute('role')")) === 'menuitem')
    await p.taste('Escape')
    await p.warte('Escape schließt es', async () => !(await p.seite(menueOffen)), 10000)

    const karte = await p.warte('die Leiste führt »Karte«', () => leistenmenue(p, 'Karte'), 10000)
    const titel = (await p.menue()).map((m) => m.titel)
    p.pruefe('»Karte« steht zwischen »Set« und »Fenster«', titel.indexOf('Karte') === titel.indexOf('Set') + 1 && titel[titel.indexOf('Karte') + 1] === 'Fenster', titel.join(' · '))
    p.pruefe('das Menü »Karte« nennt dieselben Einträge', gleich(leistennamen(karte), ueberlauf), leistennamen(karte).join(' · '))
    // **Seit AP-7 (#2364) tragen fünf Einträge ihr Kürzel** — die übrigen
    // keines; welche und wie sie wirken, misst `ap5kontext09`.
    const mitKuerzel = (karte.unter ?? []).filter((e) => e.kuerzel).map((e) => e.titel)
    p.pruefe('genau Anlegen, Duplizieren, Löschen und das Blättern tragen ein Kürzel', gleich(mitKuerzel.sort(), ['Karte anlegen', 'Karte duplizieren', 'Karte löschen', 'Nächste Karte', 'Vorige Karte']), JSON.stringify(mitKuerzel))
    await p.leistenbild('menue-karte')

    p.notiz('»Karte duplizieren« aus der Leiste')
    const vorher = kartenzahl()
    await p.klick('Karte', 'Karte duplizieren')
    await p.warte('eine Karte mehr im Bestand', () => kartenzahl() === vorher + 1, 15000).catch(() => false)
    p.pruefe('»Karte duplizieren« aus der Leiste legt eine Karte an', kartenzahl() === vorher + 1, `${String(vorher)} → ${String(kartenzahl())}`)

    p.notiz('Ein Rechtsklick in ein Feld bleibt bei macOS')
    const feldort = await p.seite(mitteVon("[...document.querySelectorAll('main input[type=text], main textarea, input:not([type]), textarea')].find((e) => e.checkVisibility())"))
    await rechtsklick(p, feldort.x, feldort.y)
    await new Promise((fertig) => setTimeout(fertig, 800))
    p.pruefe('im Feld öffnet sich kein Menü der Seite', !(await p.seite(menueOffen)))
    await p.taste('Escape')

    p.notiz('Marktobjekt — ⌘5, ein Entwurf mit Bert')
    await p.taste('5', 'befehl')
    await p.wartetext('Neuer Tausch', 20000)
    await p.klickeAuf(knopf('Neuer Tausch …'), '»Neuer Tausch …«')
    const bertWahl = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && /^Bert aufg?e?schlagen$/i.test((b.getAttribute('aria-label') ?? '').trim()))"
    await p.warte('Bert steht zur Wahl', () => p.seite(`return !!${bertWahl}`), 10000)
    await p.klickeAuf(bertWahl, 'Bert in der Schublade')
    // **Ein Vorgang entsteht mit »Eröffnen«** — vorher gibt es keinen Handel,
    // an dem Handlungen hingen. Der Weg aus `druckfeld09`, ohne das Tippen.
    for (const [seite, druck] of [['Von dir', 'p0-blitz'], ['Vom Gegenüber', 'p0-schock']]) {
      await p.warte(`»Zeile hinzufügen: ${seite}« steht`, () => p.seite(`return !!${knopf(`Zeile hinzufügen: ${seite}`)}`), 10000)
      await p.klickeAuf(knopf(`Zeile hinzufügen: ${seite}`), `»Zeile hinzufügen: ${seite}«`)
      await p.warte('der Kasten »Kennung von Hand« steht', () => p.seite(`return !!${kennungskasten}`), 10000)
      if ((await p.seite(`return ${kennungskasten}.getAttribute('aria-expanded')`)) !== 'true') {
        await p.klickeAuf(kennungskasten, 'der Kasten »Kennung von Hand«')
      }
      await p.warte('das Feld »Druck« steht', () => p.seite(`return !!${feld('Druck')}`), 10000)
      await p.seite(einfuegen(feld('Druck'), druck))
      await p.klickeAuf(knopf('+ Zeile hinzufügen'), '»+ Zeile hinzufügen«')
    }
    await p.klickeAuf(knopf('Eröffnen'), '»Eröffnen«')
    await p.warte('»Mit diesem Handel« steht', () => p.seite(`return !!${knopf('Mit diesem Handel')}`), 20000)

    const handelsueberlauf = await seitenmenue(p, 'Mit diesem Handel')
    p.pruefe('»Mit diesem Handel« nennt Einträge', handelsueberlauf.length > 0, handelsueberlauf.join(' · '))
    const anSchritten = await kontextmenue(p, "document.querySelector('[data-objekt^=\"handel/\"]')", 'die Schritte des Handels')
    p.pruefe('der Rechtsklick auf die Schritte nennt dieselben Einträge', gleich(anSchritten, handelsueberlauf), anSchritten.join(' · '))

    const handelsmenue = await p.warte('die Leiste führt »Handel«', () => leistenmenue(p, 'Handel'), 10000)
    p.pruefe('das Menü »Handel« nennt dieselben Einträge', gleich(leistennamen(handelsmenue), handelsueberlauf), leistennamen(handelsmenue).join(' · '))
    p.pruefe('das Menü »Karte« ist gegangen — das gewählte Objekt hat die Leiste', !(await leistenmenue(p, 'Karte')), (await p.menue()).map((m) => m.titel).join(' · '))
    await p.leistenbild('menue-handel')

    p.pruefe('⌘Q beendet die Anwendung', await p.beende())
  } finally {
    p.fehler += bert.fehler
    bert.aufraeumen()
  }
}
