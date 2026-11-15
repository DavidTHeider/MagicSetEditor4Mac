// **Rücknahme nach einem Fokuswechsel, je Handlungsklasse** — G05
// (#2252),
// am nativen Fenster.
//
// ⌘Z hat zwei Herren: **im tippenden Feld** nimmt das Feld seine Eingabe
// zurück, **außerhalb** nimmt die Anwendung den jüngsten Schritt ihres
// Verlaufs zurück (`frontend/src/eingabefeld.ts`). Geprüft wird beides,
// getrennt — und nach jedem Fokuswechsel, denn genau dort lag die alte
// Aussage aus A03 (#2217), die das native Audit nicht wiederfand.
//
//   1. **Eingabe-Undo**: im Feld »Name« tippen und ⌘Z, ohne das Feld zu
//      verlassen — die Eingabe geht zurück, und der Verlauf der Anwendung
//      bleibt, wie er war,
//   2. je Handlungsklasse — **Name** (Text), **Text** (mehrzeilig),
//      **Farbe** (Auswahl; das Saatspiel hat keinen Typ, die Auswahl steht
//      für ihn), **Setangaben** (Titel), **Karte anlegen** (fachliche
//      Handlung) —: ändern, den Fokus auf die Kartenzeile, ⌘Z und ⇧⌘Z.
//      **Verloren oder doppelt wird rot**: nach ⌘Z steht genau der Wert von
//      vorher, nach ⇧⌘Z genau der neue, und keine andere Klasse hat sich
//      bewegt.
//
// **Gelesen wird über die Bindungen**, nicht vom Bildschirm: was der Bestand
// sagt, ist, was gilt.
//
// **Und das Bild**: sein Weg führt über einen nativen Dateidialog — die
// Probe beantwortet ihn (`dialogantwort`) — und das Zuschnittfenster.
import { resolve } from 'node:path'

const binde = (aufruf) => `return await window.go.app.Bindungen.${aufruf}`
const feld = (name) => `document.querySelector('main [aria-label=${JSON.stringify(name)}]')`
const zeile = "document.querySelector('main [role=row][tabindex=\"0\"]')"

async function kennungen(p) {
  const sets = await p.seite(binde('Sets()'))
  const setID = sets.sets.find((s) => s.name === 'Werkstatt-Testkarten').id
  const karten = await p.seite(binde(`Karten(${JSON.stringify(setID)})`))

  return { setID, karteID: karten.karten[0].id, anzahl: karten.karten.length }
}

async function kartenwert(p, karteID, name) {
  const antwort = await p.seite(binde(`Karte(${JSON.stringify(karteID)})`))

  return (antwort.karte?.werte ?? []).find((w) => w.feld === name)?.wert ?? ''
}

async function setwert(p, setID, name) {
  const antwort = await p.seite(binde(`Setangaben(${JSON.stringify(setID)})`))

  return (antwort.werte ?? []).find((w) => w.feld === name)?.wert ?? ''
}

async function kartenzahl(p, setID) {
  return (await p.seite(binde(`Karten(${JSON.stringify(setID)})`))).karten.length
}

// ersetzeImFeld leert ein Feld über den Wertsetzer und tippt den neuen Text
// — dieselbe Form wie in `f03.mjs`: ⌘A verlor dort den ersten Tastendruck.
async function ersetzeImFeld(p, name, text) {
  await p.klickeAuf(feld(name), `das Feld »${name}«`)
  await p.seite(`const e = ${feld(name)}; const setzer = Object.getOwnPropertyDescriptor(Object.getPrototypeOf(e), 'value').set; setzer.call(e, ''); e.dispatchEvent(new Event('input', { bubbles: true })); return true`)
  await p.tippeInsFeld(text)
}

// fokusAufDieZeile wechselt den Fokus aus dem Feld — **ein Klick**, wie ihn
// ein Mensch täte, und kein `blur()` aus dem Skript.
//
// **Verlangt wird, dass er das tippende Feld verlässt**, nicht dass er auf
// der Zeile landet: in WebKit fokussiert ein Klick keinen Knopf, und der
// Knopf liegt über der ganzen Zeile (erster Lauf am 2026-10-10).
async function fokusAufDieZeile(p) {
  await p.klickeAuf(zeile, 'die Kartenzeile')
  await p.warte('der Fokus hat das Feld verlassen', () => p.seite("const e = document.activeElement; return !(e && (e.tagName === 'TEXTAREA' || (e.tagName === 'INPUT' && !['checkbox', 'radio', 'button'].includes(e.type)) || e.closest('[contenteditable=true]')))"), 5000)
}

// rundlauf nimmt eine Änderung zurück und stellt sie wieder her — und prüft,
// dass genau sie sich bewegt.
async function rundlauf(p, klasse, lesen, vorher, nachher, andere) {
  const ruhig = async () => Promise.all(andere.map((f) => f()))
  const standDerAnderen = JSON.stringify(await ruhig())

  await p.taste('z', 'befehl')
  await p.warte(`${klasse}: ⌘Z bringt den Wert von vorher`, async () => (await lesen()) === vorher, 10000)
  p.pruefe(`${klasse}: nach ⌘Z steht genau der Wert von vorher`, (await lesen()) === vorher, String(await lesen()))
  p.pruefe(`${klasse}: keine andere Klasse hat sich bewegt`, JSON.stringify(await ruhig()) === standDerAnderen)

  await p.taste('z', 'befehl', 'umschalt')
  await p.warte(`${klasse}: ⇧⌘Z bringt den neuen Wert`, async () => (await lesen()) === nachher, 10000)
  p.pruefe(`${klasse}: nach ⇧⌘Z steht genau der neue Wert`, (await lesen()) === nachher, String(await lesen()))
}

export default async function g05ruecknahme(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  await p.starte({ ort })

  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await p.warte('das Feld »name« steht da', () => p.seite(`return !!${feld('name')}`), 20000)
  const { setID, karteID, anzahl } = await kennungen(p)
  const name = () => kartenwert(p, karteID, 'name')
  const text = () => kartenwert(p, karteID, 'text')
  const farbe = () => kartenwert(p, karteID, 'farbe')
  const titel = () => setwert(p, setID, 'title')
  const kunst = () => kartenwert(p, karteID, 'kunst')
  const zahl = () => kartenzahl(p, setID)

  p.notiz('1 — ⌘Z im Feld nimmt die Eingabe zurück, nicht den Verlauf')
  const nameVorher = await name()
  const schritteVorher = JSON.stringify(await p.seite(binde(`Verlauf(${JSON.stringify(setID)})`)))
  await p.klickeAuf(feld('name'), 'das Feld »name«')
  await p.tippeInsFeld('xyz')
  // **⌘Z erreicht die Seite, und sie lässt es im Feld durch** — dann gehört
  // es AppKit als Kürzel von »Widerrufen«. Den letzten Schritt geht die
  // Probe über den Eintrag selbst (`imFenster`), wie `leiste.mjs`: eine
  // Anwendung, die aus einem Hintergrundprozess startet, hat seit macOS 14
  // kein Schlüsselfenster, an dessen Ersten Antwortenden AppKit schicken
  // könnte. Ein Mensch hätte vorher ins Fenster geklickt.
  await p.seite("window.__z = null; addEventListener('keydown', (e) => { if (e.key === 'z' && e.metaKey) setTimeout(() => { window.__z = e.defaultPrevented ? 'verhindert' : 'durchgelassen' }) }, true); return true")
  await p.taste('z', 'befehl')
  const z = await p.warte('⌘Z kommt an der Seite an', () => p.seite('return window.__z'), 5000)
  p.pruefe('die Seite lässt ⌘Z im Feld durch — die Anwendung nimmt es nicht', z === 'durchgelassen', z)
  await p.imFenster('Bearbeiten', 'Widerrufen')
  await p.warte('das Feld hat seine Eingabe zurückgenommen', () => p.seite(`return ${feld('name')}.value === ${JSON.stringify(nameVorher)}`), 5000)
  p.pruefe('im Feld steht wieder der Name von vorher', await p.seite(`return ${feld('name')}.value === ${JSON.stringify(nameVorher)}`))
  p.pruefe('der Verlauf der Anwendung blieb, wie er war', JSON.stringify(await p.seite(binde(`Verlauf(${JSON.stringify(setID)})`))) === schritteVorher)
  await fokusAufDieZeile(p)

  p.notiz('2 — Name, nach dem Fokuswechsel')
  await ersetzeImFeld(p, 'name', 'Rueckprobe')
  await fokusAufDieZeile(p)
  await p.warte('der Name ist geschrieben', async () => (await name()) === 'Rueckprobe', 10000)
  await rundlauf(p, 'Name', name, nameVorher, 'Rueckprobe', [text, farbe, titel, zahl, kunst])

  p.notiz('3 — Text, nach dem Fokuswechsel')
  const textVorher = await text()
  await ersetzeImFeld(p, 'text', 'Ein Satz zur Probe')
  await fokusAufDieZeile(p)
  await p.warte('der Text ist geschrieben', async () => (await text()) === 'Ein Satz zur Probe', 10000)
  await rundlauf(p, 'Text', text, textVorher, 'Ein Satz zur Probe', [name, farbe, titel, zahl, kunst])

  p.notiz('4 — Farbe (Auswahl), nach dem Fokuswechsel')
  const farbeVorher = await farbe()
  const neueFarbe = await p.seite(`const s = ${feld('farbe')}; return [...s.options].map((o) => o.value).find((v) => v !== '' && v !== s.value) ?? ''`)
  p.pruefe('die Auswahl bietet eine andere Farbe an', neueFarbe !== '', neueFarbe)
  // **Ohne Klick auf die Auswahl**: in WKWebView öffnet er das native
  // Aufklappmenü, und das schluckt danach den Klick auf die Zeile (zweiter
  // Lauf am 2026-10-10). Fokus und Wert wie eine Wahl mit der Tastatur.
  await p.seite(`const s = ${feld('farbe')}; s.focus(); s.value = ${JSON.stringify(neueFarbe)}; s.dispatchEvent(new Event('change', { bubbles: true })); return true`)
  await fokusAufDieZeile(p)
  await p.warte('die Farbe ist geschrieben', async () => (await farbe()) === neueFarbe, 10000)
  await rundlauf(p, 'Farbe', farbe, farbeVorher, neueFarbe, [name, text, titel, zahl, kunst])

  // **Vor »Karte anlegen«**: danach steht die neue Karte offen, und das
  // Bild ginge in sie (zweiter Lauf am 2026-10-10).
  p.notiz('5 — Bild, nach dem Fokuswechsel')
  const kunstVorher = await kunst()
  // **Derselbe Weg wie in `fix-f02.mjs`**: erst zeigt sich der Dialog von
  // macOS, dann bekommt der nächste seine Antwort; der Zuschnitt setzt mit
  // »Einsetzen« ein.
  await p.klickeText('Bild wählen')
  await p.dialogZeigtSich('oeffnen', '»Bild wählen«')
  await p.dialogantwort(resolve('saat/probe-schlicht.mse-style/hell.png'))
  await p.klickeText('Bild wählen')
  await p.klickeText('Einsetzen')
  await p.warte('der Zuschnitt ist zu', () => p.seite("return ![...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Einsetzen')"), 15000)
  await p.warte('das Bild ist eingesetzt', async () => (await kunst()) !== kunstVorher, 15000)
  const kunstNachher = await kunst()
  await fokusAufDieZeile(p)
  await rundlauf(p, 'Bild', kunst, kunstVorher, kunstNachher, [name, text, farbe, titel, zahl])

  p.notiz('6 — Karte anlegen, nach dem Fokuswechsel')
  await p.klickeText('Karte anlegen')
  await p.warte('eine Karte mehr', async () => (await zahl()) === anzahl + 1, 10000)
  await fokusAufDieZeile(p)
  await rundlauf(p, 'Karte anlegen', zahl, anzahl, anzahl + 1, [name, text, farbe, titel, kunst])

  p.notiz('7 — Setangaben, nach dem Fokuswechsel')
  const titelVorher = await titel()
  await p.taste('k', 'befehl')
  await p.tippe('Setangaben')
  await p.taste('Enter')
  await p.warte('das Feld »title« steht da', () => p.seite(`return !!document.querySelector('[aria-label="title"]')`), 15000)
  await p.klickeAuf("document.querySelector('[aria-label=\"title\"]')", 'das Feld »title«')
  await p.seite("const e = document.querySelector('[aria-label=\"title\"]'); const setzer = Object.getOwnPropertyDescriptor(Object.getPrototypeOf(e), 'value').set; setzer.call(e, ''); e.dispatchEvent(new Event('input', { bubbles: true })); return true")
  await p.tippeInsFeld('Titelprobe')
  await p.taste('Tab')
  await p.warte('der Titel ist geschrieben', async () => (await titel()) === 'Titelprobe', 10000)
  await p.taste('Escape')
  await fokusAufDieZeile(p)
  await rundlauf(p, 'Setangaben', titel, titelVorher, 'Titelprobe', [name, text, farbe, zahl, kunst])
  await p.bild('g05-ende')

  p.pruefe('⌘Q beendet', await p.beende())
}
