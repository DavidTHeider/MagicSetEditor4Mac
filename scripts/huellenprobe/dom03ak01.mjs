// **Die eigene Ausfuhr ein zweites Mal, am nativen Fenster** — DOM-03.AK-01
// (#1698):
//
// - *Als CSV* und *CSV zurücklesen* über dieselbe Sammlung: vor dem Lesen
//   steht *Vor dem Einlesen* mit *2 → 4* — nichts ist geschrieben,
// - *Vorhandene auslassen* und *Einlesen*: die Menge bleibt zwei, und der
//   Bericht sagt, was stehen blieb,
// - die Kommandozeile liest danach über denselben Bestand dieselbe Menge.
import { randomUUID } from 'node:crypto'

// **Das Zeichen des Katalogs steht vor dem Text** („+ Einlesen") — gesucht
// wird deshalb nach dem Ende, wo es keine Bezeichnung gibt.
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') === ${JSON.stringify(name)} || (!b.hasAttribute('aria-label') && b.innerText.trim().endsWith(${JSON.stringify(name)}))))`

const vorschau = "return [...document.querySelectorAll('section')].find((s) => s.checkVisibility() && s.getAttribute('aria-label') === 'Vor dem Einlesen')?.innerText ?? ''"

const mengeVon = (p, ort, druck) => {
  const zeile = p.mse(ort, 'sammlung-ausgeben', 'csv').trim().split('\n').find((z) => z.startsWith(druck)) ?? ''
  return zeile.split(',')[1] ?? ''
}

// **Ausgabe und Rücklesen stehen in einer Schublade** — seit UXR-10 (#1918)
// hinter *Einlesen und ausgeben …* im Kopf der Seite, eingeklappt unter der
// Listeneinfuhr. Vorher in einem Kasten im Navigator, der bei leerer
// Sammlung offen stand und mit dem ersten Exemplar zuklappte.
async function zurSammlung(p) {
  await p.taste('3', 'befehl')
  await p.wartetext('Einlesen und ausgeben …', 20000)
  await p.klickeText('Einlesen und ausgeben …')
  await p.warte('die Schublade steht', () => p.seite("return [...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility() && d.getAttribute('aria-label') === 'Einlesen und ausgeben')"), 20000)
  await p.klickeText('Ausgeben und sichern')
  await p.wartetext('Als CSV', 20000)
}

export default async function dom03ak01(p) {
  const ort = p.ort()
  const druck = randomUUID()
  p.mse(ort, 'exemplar', druck, '2')
  await p.starte({ ort })
  await zurSammlung(p)

  p.notiz('AK-01 — die eigene Ausfuhr zurücklesen, und vorher sehen, was sie trifft')
  await p.klickeAuf(knopf('Als CSV'), 'Als CSV')
  await p.warte('das CSV steht im Feld', () => p.seite(`return [...document.querySelectorAll('textarea')].some((t) => t.value.includes(${JSON.stringify(druck)}))`), 15000)
  await p.klickeAuf(knopf('CSV zurücklesen'), 'CSV zurücklesen')

  const text = await p.warte('die Vorschau steht da', async () => (await p.seite(vorschau)) || false, 15000)
  p.zeige(`Vorschau: ${text.replace(/\s+/g, ' ').slice(0, 240)}`)
  p.pruefe('die Vorschau nennt die Zeile mit 2 → 4', text.includes('2 → 4'))
  p.pruefe('vor der Wahl ist nichts geschrieben', mengeVon(p, ort, druck) === '2', mengeVon(p, ort, druck))
  await p.bild('vor-dem-einlesen')

  p.notiz('AK-01 — Vorhandene auslassen')
  // **Erst in den Blick** — die Vorschau steht unten in der schmalen Spalte,
  // und ein nativer Klick trifft nur, was zu sehen ist.
  const auslassen = "[...document.querySelectorAll('label')].find((l) => l.checkVisibility() && l.innerText.trim() === 'Vorhandene auslassen')"
  await p.seite(`${auslassen}.scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(auslassen, 'Vorhandene auslassen')
  await p.warte('Auslassen ist gewählt', () => p.seite(`return ${auslassen}.querySelector('input').checked`), 10000)
  await p.seite(`${knopf('Einlesen')}.scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(knopf('Einlesen'), 'Einlesen')
  await p.wartetext('1 vorhandene Zeilen ausgelassen.', 15000)
  p.pruefe('die Vorschau ist geräumt', (await p.seite(vorschau)) === '')
  p.pruefe('die Menge bleibt zwei', mengeVon(p, ort, druck) === '2', mengeVon(p, ort, druck))
  await p.bild('ausgelassen')

  p.pruefe('⌘Q beendet', await p.beende())
}
