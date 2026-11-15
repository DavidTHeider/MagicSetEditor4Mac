// **Eine eigene Vorlage kommt auf der Karte an — am nativen Fenster** —
// FMT-02, S10.NAT-03 und S10.NAT-04
// (#1691):
//
// - eine eigene Kopie des Stils `probe-schlicht` und ein Set darauf; die
//   Werkstatt zeigt die Karte dieses Sets als Vorschau,
// - der Grund der Karte wird im Editor rot — **die Vorschau folgt, ohne dass
//   gesichert ist**; ein Tippfehler bringt einen Befund und markiert die
//   Vorschau als veraltet, die Korrektur nimmt beides weg (NAT-03),
// - wer mit Ungesichertem eine andere Datei wählt, wird gefragt und kann
//   bleiben (NAT-04),
// - nach dem Sichern ist **die Karte im Set rot** — gemessen am Bild, nicht
//   an der Datei.
//
// **Was sie beim ersten Lauf fand**: die Vorschau der Werkstatt zeichnete die
// Kopie als Platzhalter (236,232,222), während das Set sie weiß zeigte — ihre
// Wurzeln standen seit dem Start fest, und das Paketverzeichnis war da leer.
// Behoben an der Klasse (`paketwurzelnFuer`, sechs Nähte). **Und nach dem
// Sichern blieb die Karte im Set weiß**: der Stilspeicher hielt die Vorlage
// seit dem ersten Zeichnen fest; jetzt gilt sie, bis sich ihre Dateien
// ändern (`stilstempel.go`).
//
// **Gemessen wird ein Bildpunkt**, gezeichnet aus dem Bild in eine Leinwand:
// eine Farbe, die niemand sieht, weil die Vorschau das alte Bild behält, wäre
// sonst eine gelesene Datei und kein gezeigtes Ergebnis.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

// **Das sichtbare Feld** — die Seite des Sets bleibt eingehängt, und ihr
// Textfeld stünde sonst als erstes da.
const feld = "[...document.querySelectorAll('textarea')].find((t) => t.checkVisibility())"
const bildMit = (alt) => `[...document.querySelectorAll('img')].find((b) => b.checkVisibility() && b.alt === ${JSON.stringify(alt)})`
const zweig = (name) => `[...document.querySelectorAll('button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith(${JSON.stringify(name)}))?.closest('li')`
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith(${JSON.stringify(name)}))`

// punkt liest die Farbe eines Bildpunkts des gezeigten Bildes — **nach
// `decode()`**: gemessen im Gesamtlauf der Abnahme von 0.7.0, Bild für Bild
// alle 50 ms in neun Läufen. Einmal trug dasselbe Bild (dieselbe Adresse,
// 375 × 523, `complete`) bei 0 ms noch den durchsichtigen Punkt `0,0,0,0`
// und 80 ms später den weißen: es war geladen, aber noch nicht dekodiert, und
// die Leinwand der Probe zeichnete nichts. Das war die Probe, nicht die Karte.
//
// **Und `decode()` allein reicht nicht** — gemessen am Freigabestand von
// 0.7.0 (M01, #2192), in vier von 13 Läufen: nach `decode()` gab das erste
// Zeichnen an der Ecke **und** in der Mitte `0,0,0,0`, dasselbe Bild (gleiche
// Adresse, 375 × 523) 300 ms später weiß. Ein Bild, das an beiden Stellen
// völlig durchsichtig ankommt, hat die Leinwand nicht gezeichnet; dann wird
// weiter gewartet, statt den leeren Abzug als Farbe zu werten. Eine Karte, die
// wirklich durchsichtig wäre, bliebe so bis zur Frist leer — und rot.
const punkt = (bild) => `const b = ${bild}; if (!b || !b.complete || b.naturalWidth === 0) return null; await b.decode(); const c = document.createElement('canvas'); c.width = b.naturalWidth; c.height = b.naturalHeight; const k = c.getContext('2d'); k.drawImage(b, 0, 0); const d = k.getImageData(6, 6, 1, 1).data; const m = k.getImageData(b.naturalWidth >> 1, b.naturalHeight >> 1, 1, 1).data; if (d[3] === 0 && m[3] === 0) return null; return [d[0], d[1], d[2]]`
const rot = (farbe) => Array.isArray(farbe) && farbe[0] > 180 && farbe[1] < 80 && farbe[2] < 80
const weiss = (farbe) => Array.isArray(farbe) && farbe.every((kanal) => kanal > 240)

async function einfuegen(p, text) {
  await p.seite(`const t = ${feld}; const setze = Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype, 'value').set; setze.call(t, ${JSON.stringify(text)}); t.dispatchEvent(new Event('input', { bubbles: true })); return true`)
}

async function zurSeite(p, ziel) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld(ziel)
  await p.klickeAuf(`[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(ziel)})`, `»${ziel}« in der Palette`)
}

async function zumSet(p) {
  await p.taste('2', 'befehl')
  await p.warte('das Set steht in der Liste', () => p.seite("return document.body.innerText.includes('Rotprobe')"), 20000)
  await p.klickeText('Rotprobe')
  await p.warte('das Set ist offen', () => p.seite("return (document.querySelector('nav ol')?.innerText ?? '').includes('Rotprobe')"), 20000)
}

export default async function fmt02f(p) {
  const ort = p.ort({ zuhause: true })

  await p.starte({ ort })

  // **Die Kopie entsteht im Fenster** — die Saat liegt im Bündel, und die
  // Kommandozeile dieser Probe steht nicht darin; danach liegt die Kopie im
  // verwalteten Verzeichnis, und dort sieht sie auch die Kommandozeile.
  p.notiz('Eine eigene Kopie des Stils, im Fenster')
  await zurSeite(p, 'Vorlagen')
  await p.warte('der Zweig sagt »nur lesbar«', () => p.seite(`return (${zweig('probe-schlicht.mse-style')})?.innerText.includes('nur lesbar') ?? false`), 20000)
  await p.klickeAuf(`[...(${zweig('probe-schlicht.mse-style')})?.querySelectorAll('button') ?? []].find((b) => b.innerText.trim().endsWith('Eigene Kopie anlegen'))`, 'Eigene Kopie anlegen')
  await p.wartetext('Das Original bleibt.', 15000)
  const setID = JSON.parse(p.mse(ort, 'set-neu', 'Rotprobe', 'probe', 'probe-schlicht-kopie', '--json')).id
  p.mse(ort, 'karte', String(setID), 'anlegen')
  const kopie = JSON.parse(p.mse(ort, 'pakete', '--json')).pakete.find((paket) => paket.name === 'probe-schlicht-kopie')
  const datei = join(kopie.pfad, 'style')
  const gelesen = readFileSync(datei, 'utf8')

  p.notiz('Das Set ist offen, und die Karte ist weiß')
  await zumSet(p)
  const vorher = await p.warte('die Karte ist gezeichnet', () => p.seite(punkt(bildMit('Vorschau der Karte'))), 20000)
  p.pruefe('die Karte hat den weißen Grund des Stils', weiss(vorher), JSON.stringify(vorher))

  p.notiz('NAT-03 — die Vorschau folgt dem Editor, bevor gesichert ist')
  await zurSeite(p, 'Vorlagen')
  await p.warte('die Kopie steht im Baum', () => p.seite(`return !!(${zweig('probe-schlicht-kopie.mse-style')})`), 20000)
  await p.klickeAuf(`[...(${zweig('probe-schlicht-kopie.mse-style')}).querySelectorAll('button')].find((b) => b.innerText.trim() === 'style')`, 'style in der Kopie')
  await p.warte('der Editor zeigt den Stil', () => p.seite(`return (${feld})?.value.includes('rgb(255,255,255)') ?? false`), 15000)
  const vorschauVorher = await p.warte('die Vorschau ist gezeichnet', () => p.seite(punkt(bildMit('Vorschau'))), 20000)
  p.pruefe('die Vorschau zeigt die Karte des Sets weiß', weiss(vorschauVorher), JSON.stringify(vorschauVorher))
  const roterStil = gelesen.replace('else rgb(255,255,255)', 'else rgb(200,30,30)')
  await einfuegen(p, roterStil)
  const vorschauRot = await p.warte('die Vorschau wird rot', async () => {
    const farbe = await p.seite(punkt(bildMit('Vorschau')))
    return rot(farbe) ? farbe : false
  }, 20000)
  p.pruefe('die Vorschau ist rot, ungesichert', rot(vorschauRot) && readFileSync(datei, 'utf8') === gelesen, JSON.stringify(vorschauRot))
  await p.bild('vorschau-rot')

  p.notiz('NAT-03 — ein Tippfehler bringt einen Befund, die Korrektur nimmt ihn weg')
  await einfuegen(p, gelesen.replace('else rgb(255,255,255)', 'else rgb(200,30,30'))
  await p.wartetext('Veraltet', 15000)
  p.pruefe('die Vorschau ist als veraltet markiert', true)
  await p.bild('veraltet')
  await einfuegen(p, roterStil)
  await p.wartetext('Kein Befund in dieser Datei.', 15000)
  p.pruefe('nach der Korrektur ist sie nicht mehr veraltet', !(await p.seite("return document.body.innerText.includes('Veraltet')")))

  p.notiz('NAT-04 — eine andere Datei wählen, mit Ungesichertem')
  await p.klickeAuf(`[...(${zweig('probe-schlicht-kopie.mse-style')}).querySelectorAll('button')].find((b) => b.innerText.includes('hell.png'))`, 'hell.png in der Kopie')
  await p.wartetext('Hier bleiben', 10000)
  p.pruefe('die Frage bietet Bleiben und Verwerfen', await p.seite("return document.body.innerText.includes('Verwerfen und öffnen')"))
  await p.bild('frage')
  await p.klickeAuf(knopf('Hier bleiben'), 'Hier bleiben')
  await p.warte('der Entwurf steht noch', () => p.seite(`return (${feld})?.value.includes('rgb(200,30,30)') ?? false`), 10000)
  p.pruefe('Bleiben behält den Entwurf', true)

  p.notiz('Gesichert — und die Karte im Set ist rot')
  await p.klickeAuf(knopf('Sichern'), 'Sichern')
  await p.wartetext('Gesichert.', 15000)
  p.pruefe('die Datei trägt den roten Grund', readFileSync(datei, 'utf8').includes('rgb(200,30,30)'))
  await zumSet(p)
  const nachher = await p.warte('die Karte ist rot', async () => {
    const farbe = await p.seite(punkt(bildMit('Vorschau der Karte')))
    return rot(farbe) ? farbe : false
  }, 20000)
  p.pruefe('die eigene Vorlage ist auf der Karte angekommen', rot(nachher), JSON.stringify(nachher))
  await p.bild('karte-rot')

  await p.beende()
}
