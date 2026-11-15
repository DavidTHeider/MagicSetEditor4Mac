// **Kann ein Bild aus dem Fenster in den Finder gezogen werden?** — AP-12
// (#2369),
// Befund B-24: „Hinausziehen in Finder nach Wails beta.27/SDK-Quelle machbar
// oder begründet abgelehnt“.
//
// # Was die Quellen sagen, und was offen bleibt
//
// Wails v3.0.0-beta.27 bietet keine eigene Ziehquelle: `startDrag` ist das
// Verschieben des Fensters (`pkg/application/webview_window_darwin.go`
// 1911), und die Schicht über der WebView nimmt nur abgelegte Dateien an
// (`webview_window_darwin_drag.m`: `registerForDraggedTypes` für
// `NSFilenamesPboardType`, `hitTest` gibt `nil` — die Maus geht an die
// WKWebView durch). Ob **WebKit selbst** ein gezogenes `<img>` als Datei an
// den Finder gibt, steht nirgends — das misst dieser Lauf.
//
// # Wie
//
// Die Anwendung startet mit künstlichem Bestand und zeigt die Karte
// „Langtext“; gezogen wird der Griff neben den Überlagerungen
// (`bausteine/Ziehgriff.tsx`) — als Kopie unten links, siehe unten. Der erste
// Lauf (Commit d8d362c51) setzte ein eigenes `<img>` ein und belegte die
// Machbarkeit; seitdem gilt der Griff, und die Datei muss „Langtext.png“
// heißen.
// Ein leerer Zielordner geht im Finder auf. **Gezogen wird von außen**:
// mit einer echten Maus über den Fensterserver (im Lauf am 2026-10-17 über
// die Rechnersteuerung), nicht über die Abnahmeleitung — die schickt ihre
// Ereignisse an AppKit in diesem Prozess, und eine Ziehsitzung über zwei
// Anwendungen hinweg sähe sie nicht. Das Drehbuch wartet drei Minuten auf
// eine Datei im Zielordner und nennt, was ankam.
//
// # Was es nicht belegt
//
// Eine Ziehquelle im Produkt — gebaut wird sie nur, wenn dieser Lauf eine
// Datei sieht. Und es läuft nicht unbeaufsichtigt: ohne jemanden, der zieht,
// endet es mit „keine Datei“, und das ist dann kein Befund über WebKit.
import { execFileSync } from 'node:child_process'
import { mkdirSync, readdirSync, readFileSync, statSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

export default async function ap12ziehquelle09(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  const ziel = join(ort.ort, 'ziel')
  mkdirSync(ziel, { recursive: true })
  await p.starte({ ort })
  await p.groesse(1440, 838)
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  // **Der Griff neben den Überlagerungen** — seit dem Machbarkeitslauf
  // (Commit d8d362c51: ein eingesetztes Bild, Datei nach der Marke benannt)
  // zieht das Drehbuch das Bild, das die Anwendung selbst anbietet.
  const griff = "[...document.querySelectorAll('main img[draggable=true]')].find((b) => b.checkVisibility() && b.parentElement?.innerText.includes('Bild ziehen'))"
  await p.warte('der Griff steht', () => p.seite(`return !!(${griff})?.complete`), 30000)
  // **Gezogen wird eine Kopie des Griffs, unten links** (`cloneNode`: dieselbe
  // Adresse, dieselben Eigenschaften). Am Platz des Griffs lag am 2026-10-17
  // ein schwebendes Fenster einer anderen Anwendung über allen Fenstern, und
  // es auszublenden war nicht freigegeben. Belegt ist damit, was der Griff
  // anbietet — Element und Adresse —, nicht das Greifen an seiner Stelle.
  const lage = await p.seite(`const g = (${griff}); g.style.outline = '3px solid red'; const k = g.cloneNode(true); k.id = 'ziehgriffkopie'; k.style.cssText = 'position: fixed; left: 40px; bottom: 40px; height: 120px; width: auto; z-index: 2147483647; outline: 3px solid red;'; document.body.append(k); const r = k.getBoundingClientRect(); return { x: Math.round(r.left + r.width / 2), y: Math.round(r.top + r.height / 2), adresse: k.getAttribute('src'), wieGriff: k.getAttribute('src') === g.getAttribute('src') && k.draggable === g.draggable }`)
  p.pruefe('die Kopie trägt Adresse und Ziehbarkeit des Griffs', lage.wieGriff === true && lage.adresse.endsWith('/Langtext.png'), lage.adresse)
  const fenster = await p.fenster()
  p.zeige(`Ziehprobe in der Seite bei ${JSON.stringify(lage)}, Fenster ${JSON.stringify(fenster.rahmen)}`)
  writeFileSync(join(process.env.TMPDIR ?? '/tmp', 'ap12ziehquelle-lage.json'), JSON.stringify({ lage, rahmen: fenster.rahmen, ziel }))
  execFileSync('open', [ziel])
  await p.bild('ziehprobe-bereit')

  let datei = null
  try {
    datei = await p.warte('eine Datei kommt im Zielordner an', () => { const namen = readdirSync(ziel).filter((n) => !n.startsWith('.')); return namen.length > 0 ? namen[0] : null }, 180000)
  } catch {
    datei = null
  }
  if (datei === null) {
    p.pruefe('eine Datei kam im Zielordner an', false, 'keine — siehe Kopf: ohne Ziehen ist das kein Befund über WebKit')
  } else {
    const pfad = join(ziel, datei)
    const kopf = readFileSync(pfad).subarray(0, 8).toString('hex')
    p.pruefe('eine Datei kam im Zielordner an', true, `${datei}, ${String(statSync(pfad).size)} Bytes, Kopf ${kopf}${kopf === '89504e470d0a1a0a' ? ' (PNG)' : ''}`)
    p.pruefe('sie ist ein PNG und heißt wie die Karte', kopf === '89504e470d0a1a0a' && datei === 'Langtext.png', datei)
  }
  await p.bild('ziehprobe-danach')
  p.pruefe('⌘Q beendet', await p.beende())
}
