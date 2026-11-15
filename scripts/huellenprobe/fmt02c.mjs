// **Der Dateibaum der Werkstatt — am nativen Fenster** — FMT-02.AK-01
// (#1691):
//
// - in der eigenen Kopie einer mitgelieferten Vorlage eine **neue Datei**
//   anlegen — sie steht danach auf der Platte und im Editor,
// - sie **umbenennen** — der alte Name ist fort, der neue da,
// - sie **löschen** nach der Rückfrage — sie liegt danach in der Sicherung,
//   und der Editor steht wieder auf der Hauptdatei,
// - ein **Bild hinzufügen** über den Dialog von macOS — es steht im Paket
//   und im Baum als Beiwerk; die Hauptdatei bietet weder Umbenennen noch
//   Löschen an.
//
// **Ein eigenes Zuhause**: alles geschieht im verwalteten Paketverzeichnis.
import { existsSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith(${JSON.stringify(name)}))`
const zweig = (name) => `[...document.querySelectorAll('button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith(${JSON.stringify(name)}))?.closest('li')`
const zeile = "document.querySelector('[role=group][aria-label=\"Dateien des Pakets\"]')"
const aktuell = "document.querySelector('nav [aria-current=true]')?.innerText.trim() ?? ''"

// **Die Handgriffe stehen seit dem 2026-10-02 im Menü »Datei«**
// (#2014) —
// bis zum 2026-10-05 suchte dieses Drehbuch Knöpfe in der Zeile, und die
// Verneinungen an der Hauptdatei bestanden, weil **gar kein** Knopf mehr da
// war. Gelesen werden deshalb die Einträge des offenen Menüs, und daneben
// steht »Neue Datei« als Beleg, dass das Menü überhaupt gelesen wurde (§3).
const dateimenue = `(${zeile})?.querySelector('button[aria-haspopup=menu]')`
const eintrag = (name) => `[...document.querySelectorAll('[role=menu] [role=menuitem]')].find((e) => e.checkVisibility() && e.innerText.trim().endsWith(${JSON.stringify(name)}))`

async function menueeintraege(p) {
  await p.klickeAuf(dateimenue, 'das Menü »Datei«')
  const liste = await p.warte('das Menü steht', () => p.seite("const m = [...document.querySelectorAll('[role=menu]')].find((e) => e.checkVisibility()); return m ? [...m.querySelectorAll('[role=menuitem]')].map((e) => e.innerText.replace(/\\s+/g, ' ').trim()) : false"))
  await p.taste('Escape')
  return liste
}

async function imMenue(p, name) {
  await p.klickeAuf(dateimenue, 'das Menü »Datei«')
  await p.warte(`»${name}« steht im Menü`, () => p.seite(`return !!(${eintrag(name)})`))
  await p.klickeAuf(eintrag(name), `»${name}« im Menü`)
}

// setzeFeld setzt den Wert des Namensfelds, wie ein Einfügen es täte — der
// alte Name soll ersetzt und nicht ergänzt werden.
async function setzeFeld(p, wert) {
  await p.seite(`const f = (${zeile}).querySelector('input'); const setze = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set; setze.call(f, ${JSON.stringify(wert)}); f.dispatchEvent(new Event('input', { bubbles: true })); f.focus(); return true`)
}

async function zurWerkstatt(p) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld('Vorlagen')
  await p.klickeAuf("[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Vorlagen')", '»Vorlagen« in der Palette')
  await p.wartetext('probe-schlicht.mse-style', 20000)
}

export default async function fmt02c(p) {
  const ort = p.ort({ zuhause: true })

  await p.starte({ ort })
  await p.groesse(1280, 820)

  p.notiz('Die eigene Kopie — die Hauptdatei bietet weder Umbenennen noch Löschen')
  await zurWerkstatt(p)
  await p.warte('der Zweig sagt »nur lesbar«', () => p.seite(`return (${zweig('probe-schlicht.mse-style')})?.innerText.includes('nur lesbar') ?? false`), 15000)
  await p.klickeAuf(`[...(${zweig('probe-schlicht.mse-style')})?.querySelectorAll('button') ?? []].find((b) => b.innerText.trim().endsWith('Eigene Kopie anlegen'))`, 'Eigene Kopie anlegen')
  await p.wartetext('Das Original bleibt.', 15000)
  await p.warte('die Zeile der Handgriffe steht', () => p.seite(`return !!(${zeile})`), 10000)
  const paket = JSON.parse(p.mse(ort, 'pakete', '--json')).pakete.find((eintrag) => eintrag.name === 'probe-schlicht-kopie').pfad
  const angeboten = await menueeintraege(p)
  p.pruefe('das Menü der Hauptdatei bietet »Neue Datei«', angeboten.some((e) => e.endsWith('Neue Datei')), angeboten.join(' · '))
  p.pruefe('an der Hauptdatei kein Umbenennen', !angeboten.some((e) => e.endsWith('Umbenennen')))
  p.pruefe('an der Hauptdatei kein Löschen', !angeboten.some((e) => e.endsWith('Löschen …')))
  await p.bild('hauptdatei')

  p.notiz('Eine neue Datei anlegen')
  await imMenue(p, 'Neue Datei')
  await p.warte('das Namensfeld hat den Fokus', () => p.seite(`return document.activeElement === (${zeile}).querySelector('input')`))
  await p.tippeInsFeld('skripte/farben')
  await p.taste('Enter')
  await p.warte('der Baum zeigt die neue Datei gewählt', async () => (await p.seite(`return ${aktuell}`)) === 'skripte/farben', 15000)
  p.pruefe('sie liegt auf der Platte', existsSync(join(paket, 'skripte', 'farben')))
  await p.bild('neue-datei')

  p.notiz('Umbenennen')
  await imMenue(p, 'Umbenennen')
  await p.warte('das Feld trägt den alten Namen', () => p.seite(`return (${zeile}).querySelector('input')?.value === 'skripte/farben'`))
  await setzeFeld(p, 'skripte/farbe')
  await p.taste('Enter')
  await p.warte('der Baum zeigt den neuen Namen gewählt', async () => (await p.seite(`return ${aktuell}`)) === 'skripte/farbe', 15000)
  p.pruefe('der neue Name liegt auf der Platte', existsSync(join(paket, 'skripte', 'farbe')))
  p.pruefe('der alte ist fort', !existsSync(join(paket, 'skripte', 'farben')))
  await p.bild('umbenannt')

  p.notiz('Löschen, mit Rückfrage in der Zeile')
  await imMenue(p, 'Löschen …')
  await p.wartetext('Die Datei liegt danach in der Sicherung neben dem Paket.', 10000)
  await p.bild('loeschfrage')
  await p.klickeAuf(knopf('Ja, löschen'), 'Ja, löschen')
  await p.warte('der Editor steht wieder auf der Hauptdatei', async () => (await p.seite(`return ${aktuell}`)) === 'style', 15000)
  p.pruefe('die Datei ist aus dem Paket fort', !existsSync(join(paket, 'skripte', 'farbe')))
  p.pruefe('sie liegt in der Sicherung', existsSync(join(`${paket}.bak`, 'skripte', 'farbe')))

  p.notiz('Ein Bild über den Dialog von macOS hinzufügen')
  const bild = join(mkdtempSync(join(tmpdir(), 'fmt02c-')), 'maske.png')
  writeFileSync(bild, Buffer.from('89504e470d0a1a0a', 'hex'))
  await imMenue(p, 'Bild oder Schrift hinzufügen …')
  await p.dialogZeigtSich('oeffnen', '»Bild oder Schrift hinzufügen …«')
  p.pruefe('ein Abbruch legt nichts ab', !existsSync(join(paket, 'maske.png')))
  await p.dialogantwort(bild)
  await imMenue(p, 'Bild oder Schrift hinzufügen …')
  await p.wartetext('Kein Text', 15000)
  p.pruefe('das Bild liegt im Paket', existsSync(join(paket, 'maske.png')))
  p.pruefe('der Baum führt es als Beiwerk', await p.seite("return !!document.querySelector('nav button[aria-label=\"maske.png — Beiwerk\"]')"))
  await p.bild('beiwerk')

  await p.beende()
}
