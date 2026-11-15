// **Phase C am nativen Fenster: ein Entwurf übersteht fremde Änderungen** —
// U01 bis U06 (#2119,
// #2120, #2121, #2122).
//
// Der fremde Schreiber ist **`mse`** gegen denselben Bestand — ein anderer
// Prozess, wie im Betrieb. Das Fenster sieht ihn über `data_version`
// (`useBestandsstand`, alle zwei Sekunden) und liest nach. Gemessen wird je
// Formular:
//
// - **Unabhängig**: ein anderes Feld ändert sich fremd; die Eingabe, in der
//   getippt wird, ist **derselbe Knoten** (eine Marke am Element überlebt nur
//   ohne Neuaufbau), trägt den Entwurf und hat den Fokus.
// - **Dasselbe Feld** (Karteneditor): der Satz am Feld nennt den fremden
//   Wert; „Meinen Entwurf speichern" schreibt den Entwurf, nachgelesen über
//   `mse`.
//
// **Was sie nicht belegt**: Hell/Dunkel, die Größe 1024×700 und Englisch —
// die Formulare ändern dort nichts an Aufbau oder Schlüsseln; und den
// Markenkasten nur im unabhängigen Fall (der gleiche ist im DOM-Lauf).
import { oeffnen, zurSeite } from './uirevue.mjs'

const eingabe = (ausdruck) => `(${ausdruck})`
const nachAria = (name) => `document.querySelector('input[aria-label=${JSON.stringify(name)}], textarea[aria-label=${JSON.stringify(name)}]')`
const nachLabel = (text) => `[...document.querySelectorAll('input, textarea')].find((e) => [...(e.labels ?? [])].some((l) => l.innerText.trim().startsWith(${JSON.stringify(text)})))`

// tippeAmEnde fokussiert die Eingabe, setzt die Marke ans Ende, markiert den
// Knoten und tippt nativ.
async function tippeAmEnde(p, ausdruck, text) {
  const vorher = await p.seite(`const e = ${eingabe(ausdruck)}; e.__probe = 'phasec'; e.focus(); e.setSelectionRange(e.value.length, e.value.length); return e.value`)
  await p.tippeInsFeld(text)

  return `${vorher}${text}`
}

// ueberstanden prüft Knoten, Wert und Fokus nach einer fremden Änderung.
async function ueberstanden(p, was, ausdruck, erwartet) {
  const lage = await p.seite(`const e = ${eingabe(ausdruck)}; return { selbe: e?.__probe === 'phasec', wert: e?.value, fokus: document.activeElement === e }`)
  p.zeige(`${was}: derselbe Knoten ${lage.selbe}, Wert »${lage.wert}«, Fokus ${lage.fokus}`)
  p.pruefe(`${was}: derselbe Knoten`, lage.selbe)
  p.pruefe(`${was}: der Entwurf steht`, lage.wert === erwartet, lage.wert)
  p.pruefe(`${was}: der Fokus bleibt`, lage.fokus)
}

const karteNamens = (p, ort, id) => JSON.parse(p.mse(ort, 'karten', '1', '--json')).karten.find((k) => k.id === id)

export default async function phasec(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  await oeffnen(p, ort, { name: 'normal', breite: 1280, hoehe: 800 })
  await zurSeite(p, 'Karten')
  await p.warte('das Namensfeld steht da', () => p.seite(`return !!${nachAria('name')}`), 20000)

  const name = await p.seite(`return ${nachAria('name')}.value`)
  const karte = JSON.parse(p.mse(ort, 'karten', '1', '--json')).karten.find((k) => k.name === name)
  p.zeige(`offen: Karte ${karte.id} »${name}«`)

  p.notiz('U01 — ein anderes Feld derselben Karte ändert sich fremd, während im Namen getippt wird')
  const entwurf = await tippeAmEnde(p, nachAria('name'), ' Probe')
  p.mse(ort, 'feld', '1', karte.id, 'text', 'Fremd geschrieben')
  await p.warte('das Textfeld zeigt den fremden Wert', () => p.seite(`return ${nachAria('text')}?.value === 'Fremd geschrieben'`), 15000)
  await ueberstanden(p, 'Karteneditor', nachAria('name'), entwurf)
  await p.bild('phasec-u01-unabhaengig')

  p.notiz('U02 — dasselbe Feld ändert sich fremd')
  p.mse(ort, 'feld', '1', karte.id, 'name', 'Fremdname')
  await p.warte('der Satz am Feld nennt den fremden Wert', () => p.seite("return document.body.innerText.includes('Inzwischen steht hier »Fremdname«')"), 15000)
  await ueberstanden(p, 'Karteneditor im Konflikt', nachAria('name'), entwurf)
  await p.bild('phasec-u02-konflikt')
  await p.klickeAuf("[...document.querySelectorAll('button')].find((b) => b.innerText.trim() === 'Meinen Entwurf speichern')", '»Meinen Entwurf speichern«')
  await p.warte('der Bestand trägt den Entwurf', () => karteNamens(p, ort, karte.id)?.name === entwurf, 15000)
  p.pruefe('der Entwurf ist geschrieben', true)
  await p.warte('der Satz ist fort', () => p.seite("return !document.body.innerText.includes('Inzwischen steht hier')"), 8000)
  p.pruefe('kein Konflikt mehr', true)

  p.notiz('U04 — die Notiz übersteht eine fremde Änderung anderswo')
  // **Notiz und Marke stehen im Navigator**, und der ist im Arbeitsplatz
  // zunächst eingeklappt.
  await p.klickeAuf("[...document.querySelectorAll('button')].find((b) => b.innerText.trim() === 'Navigator einblenden')", '»Navigator einblenden«')
  await p.warte('der Arbeitsstand steht da', () => p.seite("return [...document.querySelectorAll('details summary')].some((s) => s.innerText.trim() === 'Arbeitsstand dieser Karte')"), 10000)
  await p.seite("const d = [...document.querySelectorAll('details')].find((e) => e.querySelector('summary')?.innerText.trim() === 'Arbeitsstand dieser Karte'); d.open = true; return true")
  const notiz = nachLabel('Notizen zur Karte')
  await p.warte('der Notizkasten steht da', () => p.seite(`return !!${notiz}`))
  const notizentwurf = await tippeAmEnde(p, notiz, 'Notiz Probe')
  p.mse(ort, 'feld', '1', karte.id, 'text', 'Wieder fremd')
  await p.warte('das Textfeld zeigt den zweiten fremden Wert', () => p.seite(`return ${nachAria('text')}?.value === 'Wieder fremd'`), 15000)
  await ueberstanden(p, 'Notiz', notiz, notizentwurf)

  p.notiz('U05 — eine noch nicht gesetzte Marke übersteht eine fremde Änderung')
  const marke = `document.getElementById('marke-${karte.id}')`
  const markenentwurf = await tippeAmEnde(p, marke, 'pruef')
  p.mse(ort, 'feld', '1', karte.id, 'text', 'Ein drittes Mal')
  await p.warte('das Textfeld zeigt den dritten fremden Wert', () => p.seite(`return ${nachAria('text')}?.value === 'Ein drittes Mal'`), 15000)
  await ueberstanden(p, 'Marke', marke, markenentwurf)
  await p.bild('phasec-u04-u05')

  p.notiz('U03 — die Setangaben: der Titel übersteht eine fremde Änderung am Copyright')
  await zurSeite(p, 'Setangaben')
  await p.warte('das Titelfeld steht da', () => p.seite(`return !!${nachAria('title')}`), 10000)
  const titelentwurf = await tippeAmEnde(p, nachAria('title'), ' Probe')
  p.mse(ort, 'setangaben', '1', 'setzen', 'copyright', 'Fremd')
  await p.warte('das Copyright zeigt den fremden Wert', () => p.seite(`return ${nachAria('copyright')}?.value === 'Fremd'`), 15000)
  await ueberstanden(p, 'Setangaben', nachAria('title'), titelentwurf)
  await p.bild('phasec-u03')

  p.notiz('U06 — der Schlüsselworteditor: das Muster übersteht eine fremde Änderung am Erinnerungstext')
  const wortID = p.mse(ort, 'schluesselwort', '1', 'anlegen').trim()
  p.mse(ort, 'schluesselwort', '1', 'setzen', wortID, 'wort', 'Probewort')
  p.mse(ort, 'schluesselwort', '1', 'setzen', wortID, 'muster', 'Probewort')
  await zurSeite(p, 'Schlüsselwörter')
  await p.klickeAuf("[...document.querySelectorAll('[role=row]')].find((r) => r.innerText.includes('Probewort'))?.querySelector('button')", 'das Probewort in der Liste')
  const muster = nachLabel('Muster')
  await p.warte('das Musterfeld steht da', () => p.seite(`return !!${muster}`), 10000)
  const musterentwurf = await tippeAmEnde(p, muster, ' Probe')
  p.mse(ort, 'schluesselwort', '1', 'setzen', wortID, 'erinnerung', 'Fremd erinnert')
  await p.warte('der Erinnerungstext zeigt den fremden Wert', () => p.seite(`return ${nachLabel('Erinnerungstext')}?.value === 'Fremd erinnert'`), 15000)
  await ueberstanden(p, 'Schlüsselworteditor', muster, musterentwurf)
  await p.bild('phasec-u06')

  p.pruefe('⌘Q beendet', await p.beende())
}
