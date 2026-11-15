// **Die eigene Kopie in der Werkstatt — am nativen Fenster** — FMT-02.AK-04
// und S10.NAT-01 (#1691):
//
// - unter einem mitgelieferten Paket steht im Baum „nur lesbar" und
//   *Eigene Kopie anlegen* — seine Dateien lassen sich nicht aufzählen,
// - die Kopie entsteht als eigenes Paket mit eigenem Kurznamen, und die
//   Werkstatt öffnet ihre Hauptdatei — mit „Das Original bleibt",
// - eine Änderung in der Kopie wird gesichert, und das Original trägt sie
//   nicht (S10.NAT-01).
//
// **Ein eigenes Zuhause**: die Kopie landet im verwalteten Paketverzeichnis.
import { readFileSync } from 'node:fs'
import { join, resolve } from 'node:path'

// **Das Original liegt im Bündel** — `mse` aus `.go/bin` hat keine Saat
// neben sich und kennt es nicht; gelesen wird es deshalb von der Platte.
const original = join(resolve(import.meta.dirname, '..', '..'),
  'build/bin/MagicSetEditor4Mac.app/Contents/Resources/saat/probe-schlicht.mse-style', 'style')

const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith(${JSON.stringify(name)}))`
const zweig = (name) => `[...document.querySelectorAll('button[aria-expanded]')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith(${JSON.stringify(name)}))?.closest('li')`
const feld = "document.querySelector('textarea')"

const paketeVon = (p, ort) => JSON.parse(p.mse(ort, 'pakete', '--json')).pakete
const inhaltVon = (p, ort, paket) => p.mse(ort, 'paketdatei', paket, 'style')

async function zurWerkstatt(p) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld('Vorlagen')
  await p.klickeAuf("[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Vorlagen')", '»Vorlagen« in der Palette')
  await p.wartetext('probe-schlicht.mse-style', 20000)
}

export default async function fmt02(p) {
  const ort = p.ort({ zuhause: true })

  await p.starte({ ort })

  p.notiz('Ein mitgeliefertes Paket in der Werkstatt — nur lesbar')
  await zurWerkstatt(p)
  await p.warte('der Zweig sagt »nur lesbar«', () => p.seite(`return (${zweig('probe-schlicht.mse-style')})?.innerText.includes('nur lesbar') ?? false`), 15000)
  const vorher = readFileSync(original, 'utf8')
  p.pruefe('der Baum bietet die Kopie an', await p.seite(`return !![...(${zweig('probe-schlicht.mse-style')})?.querySelectorAll('button') ?? []].find((b) => b.innerText.trim().endsWith('Eigene Kopie anlegen'))`))
  await p.bild('mitgeliefert')

  p.notiz('AK-04 — die eigene Kopie anlegen, die Werkstatt öffnet sie')
  await p.klickeAuf(`[...(${zweig('probe-schlicht.mse-style')})?.querySelectorAll('button') ?? []].find((b) => b.innerText.trim().endsWith('Eigene Kopie anlegen'))`, 'Eigene Kopie anlegen unter probe-schlicht')
  await p.wartetext('Das Original bleibt.', 15000)
  await p.warte('der Editor steht in der Kopie', () => p.seite(`return !!(${feld})`), 10000)
  const kopie = paketeVon(p, ort).find((paket) => paket.name === 'probe-schlicht-kopie')
  p.pruefe('die Kopie ist ein eigenes Paket', kopie?.herkunftsart === 'eigen', JSON.stringify(kopie ?? {}).slice(0, 200))
  p.pruefe('die Kopie trägt einen eigenen Kurznamen', inhaltVon(p, ort, kopie.pfad).includes('short name: schlicht-kopie'))
  p.pruefe('der Baum führt die Kopie', await p.seite(`return !!(${zweig('probe-schlicht-kopie.mse-style')})`))
  p.pruefe('der Editor zeigt ihren Stil', await p.seite(`return (${feld}).value.includes('short name: schlicht-kopie')`))
  await p.bild('kopie')

  p.notiz('S10.NAT-01 — eine Änderung in der Kopie sichern, das Original bleibt')
  await p.seite(`const f = ${feld}; f.focus(); f.setSelectionRange(f.value.length, f.value.length); return true`)
  await p.tippeInsFeld('# geprobt')
  await p.klickeAuf(knopf('Sichern'), 'Sichern')
  await p.wartetext('Gesichert.', 10000)
  p.pruefe('die Kopie trägt die Änderung', inhaltVon(p, ort, kopie.pfad).includes('# geprobt'))
  p.pruefe('das Original ist Byte für Byte, wie es war', readFileSync(original, 'utf8') === vorher)
  await p.bild('gesichert')

  await p.beende()
}
