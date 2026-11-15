// **Ein Paket als `.mse-installer` — hinaus und wieder herein, am nativen
// Fenster** — FMT-02.AK-04
// (#1691):
//
// - im Inspector eines eigenen Pakets steht *Als Installer weitergeben …*;
//   der Dialog von macOS schlägt `<paket>.mse-installer` vor, ein Abbruch
//   schreibt nichts, und danach steht die Datei da,
// - unter *Aus einer MSE2-Installation übernehmen* holt *Installer einlesen …*
//   das entfernte Paket zurück — mit dem Installer als Herkunft,
// - ein zweites Einlesen übergeht es.
//
// **Ein eigenes Zuhause**: angelegt, entfernt und eingelesen wird im
// verwalteten Paketverzeichnis dieser Probe.
import { existsSync, mkdtempSync, readFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const zeile = (name) => `document.querySelector('tr[aria-label=${JSON.stringify(name)}]')`
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith(${JSON.stringify(name)}))`

async function zurSeite(p, ziel) {
  await p.taste('k', 'befehl')
  // **Das Feld nimmt den Fokus, oder es wird gewartet** — eine Palette, die
  // gerade zugeht, stand für einen Augenblick noch da, und der Ruf danach
  // fand kein Feld mehr (2026-10-02, „javascript error" mitten im Lauf).
  await p.warte('die Palette nimmt Eingaben', () => p.seite("const i = document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]'); if (!i) return false; i.focus(); return document.activeElement === i"))
  await p.tippeInsFeld(ziel)
  await p.klickeAuf(`[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(ziel)})`, `»${ziel}« in der Palette`)
}

const zuDenPaketen = (p) => zurSeite(p, 'Pakete')

export default async function fmt02e(p) {
  const ort = p.ort({ zuhause: true })
  const ziel = join(mkdtempSync(join(tmpdir(), 'fmt02e-')), 'bausteine.mse-installer')

  await p.starte({ ort })
  p.mse(ort, 'paket-anlegen', 'mse-include', 'bausteine', 'Bausteine')

  p.notiz('Im Inspector: als Installer weitergeben')
  await zuDenPaketen(p)
  await p.warte('das eigene Paket steht in der Liste', () => p.seite(`return !!${zeile('bausteine')}`), 20000)
  await p.klickeAuf(zeile('bausteine'), 'die Zeile bausteine')
  await p.klickeAuf(knopf('Als Installer weitergeben …'), 'Als Installer weitergeben …')
  const blatt = await p.dialogZeigtSich('sichern', '»Als Installer weitergeben …«')
  p.pruefe('der Dialog schlägt den Namen vor, den MSE2 wählte', blatt.name === 'bausteine.mse-installer', blatt.name)
  p.pruefe('ein Abbruch schreibt nichts', !existsSync(ziel))
  await p.dialogantwort(ziel)
  await p.klickeAuf(knopf('Als Installer weitergeben …'), 'Als Installer weitergeben …')
  await p.wartetext(`Als Installer gesichert: ${ziel}`, 15000)
  // **Und sie steht im Blick** — der Inspector ist eine eigene Rollfläche,
  // und im ersten Lauf lag die Meldung unter ihrem sichtbaren Rand
  // (437 gegen 424 Punkte). Ein Punkt Spiel für die Rundung der Kanten.
  await p.warte('die Meldung steht im sichtbaren Teil des Inspectors', () => p.seite("const e = [...document.querySelectorAll('[role=status]')].find((e) => e.innerText.includes('Als Installer gesichert')); if (!e) return false; const r = e.getBoundingClientRect(); let a = e.parentElement; while (a && getComputedStyle(a).overflowY === 'visible') a = a.parentElement; const q = a ? a.getBoundingClientRect() : { top: 0, bottom: innerHeight }; return r.top >= q.top - 1 && r.bottom <= q.bottom + 1 && r.bottom <= innerHeight"), 5000)
  p.pruefe('die Meldung steht im Blick', true)
  const roh = existsSync(ziel) ? readFileSync(ziel) : Buffer.alloc(0)
  p.pruefe('die Datei ist ein ZIP mit dem Paket und der Hauptdatei', roh.subarray(0, 2).toString() === 'PK' &&
    roh.includes('bausteine.mse-include/include') && roh.includes('installer'))
  await p.bild('weitergegeben')

  p.notiz('Entfernt und über den Dialog wieder eingelesen')
  p.mse(ort, 'paket-entfernen', 'mse-include', 'bausteine')
  // **Über eine andere Seite zurück**: die Paketliste lädt beim Aufschlagen,
  // und von einer Änderung an der Kommandozeile erfährt sie sonst nichts.
  await zurSeite(p, 'Vorlagen')
  await zuDenPaketen(p)
  await p.warte('das Paket ist fort', () => p.seite(`return !${zeile('bausteine')}`), 20000)
  // **Seit UXR-17 in der Schublade** *Paket beziehen …*, im eingeklappten
  // Kasten der Übernahme (#1925).
  await p.klickeAuf(knopf('Paket beziehen …'), 'Paket beziehen …')
  await p.klickeAuf(knopf('Aus einer MSE2-Installation übernehmen'), 'den Kasten der Übernahme')
  await p.klickeAuf(knopf('Installer einlesen …'), 'Installer einlesen …')
  await p.dialogZeigtSich('oeffnen', '»Installer einlesen …«')
  await p.dialogantwort(ziel)
  await p.klickeAuf(knopf('Installer einlesen …'), 'Installer einlesen …')
  await p.wartetext('1 Vorlagen übernommen.', 15000)
  await p.warte('es steht wieder in der Liste', () => p.seite(`return !!${zeile('bausteine')}`), 20000)
  const paket = JSON.parse(p.mse(ort, 'pakete', '--json')).pakete.find((eintrag) => eintrag.name === 'bausteine')
  p.pruefe('es trägt den Installer als Herkunft', paket?.herkunft === ziel && paket?.herkunftsart === 'uebernommen',
    `${paket?.herkunft} / ${paket?.herkunftsart}`)
  await p.bild('eingelesen')

  p.notiz('Ein zweites Einlesen übergeht')
  await p.dialogantwort(ziel)
  await p.klickeAuf(knopf('Installer einlesen …'), 'Installer einlesen …')
  await p.wartetext('Schon vorhanden und deshalb übergangen', 15000)
  p.pruefe('übergangen und nicht überschrieben', await p.seite("return document.body.innerText.includes('bausteine.mse-include')"))
  await p.bild('uebergangen')

  await p.beende()
}
