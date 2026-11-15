// **Die Deckwege nach AP-6, Teil 3 — am nativen Fenster**
// (#2363),
// Entwurf `Bereiche09`.
//
// # Was es prüft
//
// 1. **Der Kernweg** mit dem offenen Set *Werkstatt-Testkarten* (Spiel
//    `probe`): nach „Deck anlegen“ steht der Fokus im Namen, das Spiel ist
//    `probe`; nach „Anlegen“ ist das neue Deck offen, und die Kartenwahl hat
//    das Set schon gewählt — Karte wählen, „Ins Deck legen“, und `mse deck`
//    kennt die Karte. Wie viele Handlungen das sind, zählt `ausgangswerte`.
// 2. **Das offene Deck** trägt die Marke „offen“ und kein „Öffnen“.
// 3. **Die Ansicht ist ein Segment**: mit ⇥ hinein und → zu „Kacheln“, der
//    Fokus bleibt im Segment; ein Klick auf „Liste“ zurück.
// 4. **„Holen und weitergeben“** ist zu und zeigt aufgeklappt Übernahme,
//    Umschlag, „Meiner Sammlung“ und „Einem Set“; „Meiner Sammlung“ führt in
//    die Sammlung.
// 5. **„Formate …“** öffnet die Schublade „Formate“ mit der Verwaltung;
//    Escape schließt sie, und der Fokus steht wieder auf dem Knopf.
// 6. **„Eine Deckliste einlesen …“** steht im Menü „…“ des offenen Decks
//    und nicht mehr als Kasten in der Mitte; es geht in einer Schublade auf.
//
// Hell und dunkel, künstlicher Bestand, eigenes Zuhause.
//
// # Was es nicht belegt
//
// **Wie viele Bedienelemente im Bild stehen** — das zählt
// `uirevue09zustaende`. **VoiceOver** hört ein Mensch.
import { execFileSync } from 'node:child_process'
import { mkdirSync } from 'node:fs'
import { homedir } from 'node:os'
import { join } from 'node:path'

const pack = process.env.MSE_PROBE_PACK ?? join(homedir(), 'full-magic-pack')
// sichtbarerKnopf findet einen Knopf an seinem Namen — **ohne das Zeichen
// davor**: „Anlegen“ trägt das Zeichen seiner Handlung (`aktionskatalog.ts`).
const sichtbarerKnopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? b.innerText).trim().replace(/^[^\\p{L}]+/u, '') === ${JSON.stringify(name)})`
const wahlNach = (beschriftung) => `[...document.querySelectorAll('select')].find((w) => w.checkVisibility() && (w.labels?.[0]?.innerText ?? '').trim().startsWith(${JSON.stringify(beschriftung)}))`

async function kernweg(p, ort) {
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await p.wartetext('Karte anlegen', 20000)

  await p.taste('4', 'befehl')
  await p.warte('»Deck anlegen« steht da', () => p.seite(`return !!${sichtbarerKnopf('Deck anlegen')}`), 20000)
  await p.klickeAuf(sichtbarerKnopf('Deck anlegen'), '»Deck anlegen«')
  await p.warte('der Fokus steht im Namen', () => p.seite("const e = document.activeElement; return e?.tagName === 'INPUT' && (e.labels?.[0]?.innerText ?? '').trim().startsWith('Name')"), 10000)
    .catch(() => undefined)
  p.pruefe('nach »Deck anlegen« steht der Fokus im Namen', await p.seite("const e = document.activeElement; return e?.tagName === 'INPUT' && (e.labels?.[0]?.innerText ?? '').trim().startsWith('Name')"))
  p.pruefe('das Spiel ist das des offenen Sets', (await p.seite(`return ${wahlNach('Spiel')}?.value`)) === 'probe', String(await p.seite(`return ${wahlNach('Spiel')}?.value`)))

  await p.tippeInsFeld('Wegdeck')
  await p.klickeAuf(sichtbarerKnopf('Anlegen'), '»Anlegen«')
  await p.warte('das neue Deck ist offen', () => p.seite("return [...document.querySelectorAll('h2, h3')].some((h) => h.checkVisibility() && h.innerText.trim() === 'Wegdeck')"), 20000)
  await p.warte('die Kartenwahl hat das Set gewählt', () => p.seite(`const w = ${wahlNach('Set')}; return !!w && w.value !== '' && w.selectedOptions[0]?.innerText.trim().startsWith('Werkstatt-Testkarten')`), 15000)
    .catch(() => undefined)
  const gewaehlt = await p.seite(`return ${wahlNach('Set')}?.selectedOptions[0]?.innerText.trim() ?? ''`)
  // **Die Option nennt die Zahl der Karten dazu** („· 5 Karten“).
  p.pruefe('die Kartenwahl hat das offene Set schon gewählt', gewaehlt.startsWith('Werkstatt-Testkarten'), gewaehlt)

  const karte = wahlNach('Karte')
  await p.warte('die Karten des Sets stehen zur Wahl', () => p.seite(`return (${karte})?.options.length > 1`), 15000)
  const ersteKarte = await p.seite(`const w = ${karte}; w.value = w.options[1].value; w.dispatchEvent(new Event('change', { bubbles: true })); return w.options[1].innerText.trim()`)
  await p.klickeAuf(sichtbarerKnopf('Ins Deck legen'), '»Ins Deck legen«')
  const imDeck = () => (JSON.parse(p.mse(ort, 'decks', '--json')).decks ?? []).find((d) => d.name === 'Wegdeck')?.hauptdeck ?? 0
  await p.warte('die Karte liegt im Deck', () => imDeck() > 0, 15000).catch(() => undefined)
  p.pruefe(`»${ersteKarte}« liegt im neuen Deck`, imDeck() > 0, `Hauptdeck ${String(imDeck())}`)
}

async function offenesDeck(p) {
  const liste = "[...document.querySelectorAll('section')].find((s) => s.checkVisibility() && s.getAttribute('aria-label') === 'Decks')"
  p.pruefe('das offene Deck trägt die Marke »offen«', await p.seite(`return [...(${liste})?.querySelectorAll('span') ?? []].some((e) => e.checkVisibility() && e.innerText.trim() === 'offen')`))
  p.pruefe('und kein »Öffnen: Wegdeck«', !(await p.seite(`return !!${sichtbarerKnopf('Öffnen: Wegdeck')}`)))
}

async function segment(p, bild) {
  const radio = (name) => `[...document.querySelectorAll('[role=radiogroup] input[type=radio]')].find((r) => r.checkVisibility({ checkVisibilityCSS: false }) !== false && r.closest('label')?.innerText.trim() === ${JSON.stringify(name)})`
  p.pruefe('die Ansicht ist eine Optionsgruppe', await p.seite("return [...document.querySelectorAll('[role=radiogroup]')].some((g) => g.checkVisibility() && g.innerText.includes('Kacheln'))"))
  await p.seite(`(${radio('Liste')}).focus(); return true`)
  await p.taste('ArrowRight')
  await p.warte('die Liste zeigt Kacheln', () => p.seite("return !!document.querySelector('ul[data-form=kacheln]')"), 10000).catch(() => undefined)
  p.pruefe('→ im Segment wählt »Kacheln«', await p.seite("return !!document.querySelector('ul[data-form=kacheln]')"))
  // **Der Fokus über zwei Sekunden** — einmal gelesen sieht nicht, ob ein
  // Nachladen nach dem Schreiben der Einstellung ihn später wegnimmt.
  const wo = "const a = document.activeElement; return a === null || a === document.body ? 'body' : a.tagName.toLowerCase() + ' ' + (a.getAttribute('aria-label') ?? a.closest('label')?.innerText ?? a.innerText ?? '').trim().slice(0, 30)"
  const verlauf = []
  for (let i = 0; i < 10; i++) {
    verlauf.push(await p.seite(wo))
    await new Promise((weiter) => { setTimeout(weiter, 200) })
  }
  // **Im Segment heißt: auf einem seiner Felder** — gemessen am Fenster
  // stand der Fokus beim ersten Blick mitunter noch auf „Liste“, ehe er auf
  // „Kacheln“ ankam; er verließ das Segment dabei nie.
  p.pruefe('und der Fokus bleibt im Segment und kommt auf »Kacheln«, auch zwei Sekunden später', verlauf.every((eintrag) => eintrag === 'input Liste' || eintrag === 'input Kacheln') && verlauf.at(-1) === 'input Kacheln', [...new Set(verlauf)].join(' → '))
  await p.bild(`kacheln-${bild}`)
  await p.klickeAuf(`(${radio('Liste')})?.closest('label')`, '»Liste«')
  await p.warte('die Liste zeigt Zeilen', () => p.seite("return !!document.querySelector('ul[data-form=liste]')"), 10000).catch(() => undefined)
  p.pruefe('ein Klick auf »Liste« stellt zurück', await p.seite("return !!document.querySelector('ul[data-form=liste]')"))
}

async function weitereWege(p, bild) {
  const kasten = sichtbarerKnopf('Holen und weitergeben')
  p.pruefe('»Holen und weitergeben« ist zu', (await p.seite(`return ${kasten}?.getAttribute('aria-expanded')`)) === 'false')
  const vier = ['Ein Deck aus einem fremden Profil holen', 'Als Umschlag weitergeben', 'Meiner Sammlung', 'Einem Set']
  p.pruefe('zugeklappt steht keiner der vier Wege da', !(await p.seite(`return ${JSON.stringify(vier)}.some((n) => !!(${sichtbarerKnopf('__N__').replace('"__N__"', 'n')}))`)))
  await p.klickeAuf(kasten, '»Holen und weitergeben«')
  await p.warte('die vier Wege stehen da', () => p.seite(`return ${JSON.stringify(vier)}.every((n) => !!(${sichtbarerKnopf('__N__').replace('"__N__"', 'n')}))`), 10000).catch(() => undefined)
  p.pruefe('aufgeklappt stehen Übernahme, Umschlag, »Meiner Sammlung« und »Einem Set« da', await p.seite(`return ${JSON.stringify(vier)}.every((n) => !!(${sichtbarerKnopf('__N__').replace('"__N__"', 'n')}))`))
  await p.bild(`wege-${bild}`)
  await p.klickeAuf(sichtbarerKnopf('Meiner Sammlung'), '»Meiner Sammlung«')
  await p.warte('die Sammlung steht da', () => p.seite("return [...document.querySelectorAll('h1')].some((h) => h.checkVisibility() && h.innerText.trim() === 'Sammlung')"), 15000).catch(() => undefined)
  p.pruefe('»Meiner Sammlung« führt in die Sammlung', await p.seite("return [...document.querySelectorAll('h1')].some((h) => h.checkVisibility() && h.innerText.trim() === 'Sammlung')"))
  await p.taste('4', 'befehl')
  await p.warte('zurück bei den Decks', () => p.seite(`return !!${sichtbarerKnopf('Formate …')}`), 15000)
}

async function formate(p, bild) {
  const knopf = sichtbarerKnopf('Formate …')
  await p.klickeAuf(knopf, '»Formate …«')
  const schublade = "[...document.querySelectorAll('[role=dialog]')].find((d) => d.checkVisibility() && d.getAttribute('aria-label') === 'Formate')"
  await p.warte('die Schublade »Formate« ist offen', () => p.seite(`return !!${schublade}`), 10000).catch(() => undefined)
  p.pruefe('»Formate …« öffnet die Schublade »Formate«', await p.seite(`return !!${schublade}`))
  p.pruefe('darin steht die Verwaltung der Formate', await p.seite(`return !!(${schublade})?.querySelector('section[aria-label="Formate"]')`))
  await p.bild(`formate-${bild}`)
  await p.taste('Escape')
  await p.warte('die Schublade ist zu', async () => !(await p.seite(`return !!${schublade}`)), 10000).catch(() => undefined)
  p.pruefe('Escape schließt sie', !(await p.seite(`return !!${schublade}`)))
  p.pruefe('und der Fokus steht wieder auf »Formate …«', await p.seite(`return document.activeElement === ${knopf}`))
}

async function einlesen(p, bild) {
  p.pruefe('»Eine Deckliste einlesen« steht nicht mehr als Kasten in der Mitte', !(await p.seite(`return !!${sichtbarerKnopf('Eine Deckliste einlesen')}`)))
  await p.klickeAuf(sichtbarerKnopf('Handlungen — Wegdeck'), '»…« am offenen Deck')
  const eintrag = "[...document.querySelectorAll('[role=menuitem]')].find((e) => e.checkVisibility() && e.innerText.trim().startsWith('Eine Deckliste einlesen'))"
  await p.warte('das Menü nennt das Einlesen', () => p.seite(`return !!${eintrag}`), 10000).catch(() => undefined)
  p.pruefe('das Menü des Decks nennt »Eine Deckliste einlesen …«', await p.seite(`return !!${eintrag}`))
  await p.klickeAuf(eintrag, '»Eine Deckliste einlesen …«')
  const schublade = "[...document.querySelectorAll('[role=dialog]')].find((d) => d.checkVisibility() && d.getAttribute('aria-label') === 'Eine Deckliste einlesen')"
  await p.warte('die Schublade zum Einlesen ist offen', () => p.seite(`return !!${schublade}`), 10000).catch(() => undefined)
  p.pruefe('es geht in einer Schublade auf, mit einem Feld für die Liste', await p.seite(`return !!(${schublade})?.querySelector('textarea')`))
  await p.bild(`einlesen-${bild}`)
  await p.taste('Escape')
  await p.warte('die Schublade ist zu', async () => !(await p.seite(`return !!${schublade}`)), 10000).catch(() => undefined)
  p.pruefe('Escape schließt sie', !(await p.seite(`return !!${schublade}`)))
}

async function lauf(p, bild) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'einstellungen', 'setzen', 'arbeitsplatz.navigator', 'sichtbar')
  p.mse(ort, 'einstellungen', 'setzen', 'arbeitsplatz.inspektor', 'sichtbar')
  // **Zuerst ein Set eines anderen Spiels** — die Vorgabe ohne offenes Set
  // ist das Spiel des ersten Sets, und stünde dort `probe`, belegte die
  // Prüfung des Spiels nichts.
  // Dafür braucht es Spiel und Stil — ein Klon aus dem lokalen Pack
  // (`MSE_PROBE_PACK`, Vorgabe `~/full-magic-pack`), keine Kopie.
  const pakete = join(ort.heim, 'Library/Application Support/MagicSetEditor4Mac/pakete')
  mkdirSync(pakete, { recursive: true })
  for (const paket of ['magic.mse-game', 'magic-m15.mse-style']) execFileSync('cp', ['-cR', join(pack, 'data', paket), pakete])
  p.mse(ort, 'set-neu', 'Glutsatz', 'magic', 'magic-m15')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Altdeck', 'magic')
  await p.starte({ ort })
  await p.groesse(1440, 838)

  p.notiz(`${bild}: der Kernweg`)
  await kernweg(p, ort)
  p.notiz(`${bild}: das offene Deck`)
  await offenesDeck(p)
  await p.bild(`deck-offen-${bild}`)
  p.notiz(`${bild}: das Segment`)
  await segment(p, bild)
  p.notiz(`${bild}: die seltenen Wege`)
  await weitereWege(p, bild)
  p.notiz(`${bild}: die Formate`)
  await formate(p, bild)
  p.notiz(`${bild}: eine Deckliste einlesen`)
  await einlesen(p, bild)
  p.pruefe(`⌘Q beendet die Anwendung (${bild})`, await p.beende())
}

export default async function ap6deckwege09(p) {
  for (const bild of ['hell', 'dunkel']) await lauf(p, bild)
}
