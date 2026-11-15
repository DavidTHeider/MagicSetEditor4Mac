// **Pakete mit Herkunft, Plan und voriger Fassung — am nativen Fenster** —
// DOM-01.AK-01 bis AK-04
// (#1696):
//
// - die Kommandozeile bindet einen Stil ein und legt ein Set auf ihm an,
// - AK-01: die Paketseite ist ein Arbeitsplatz; der Inspector nennt Nutzung,
//   Herkunft und Lizenz, der Filter nach Herkunft trifft,
// - AK-02/03: *Paket einbinden …* öffnet den Dialog von macOS; mit einer
//   neuen Fassung derselben Art steht **erst der Plan** da (ersetzen,
//   Fassungen, betroffenes Set), dann wird ersetzt; *Vorige Fassung
//   zurückholen …* tauscht zurück; *Löschen …* zeigt den Plan und löscht
//   nach *Nichts ändern* nicht,
// - AK-04: die Kommandozeile löscht den Stil; die Setangaben nennen ihn,
//   und *Bei den Paketen suchen* springt zur Paketseite mit gefüllter Suche.
import { mkdirSync, writeFileSync } from 'node:fs'
import { join } from 'node:path'

// **Das Zeichen des Katalogs steht vor dem Text** — gesucht wird deshalb nach
// dem Ende, wo es keine Bezeichnung gibt.
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') === ${JSON.stringify(name)} || (!b.hasAttribute('aria-label') && b.innerText.trim().endsWith(${JSON.stringify(name)}))))`

const zeile = (name) => `document.querySelector('tr[aria-label=${JSON.stringify(name)}]')`

async function klicke(p, name) {
  await p.klickeAuf(knopf(name), name)
}

// herkunft wählt im Navigator — **seit UXR-17 eine Auswahl** und keine
// Knopfreihe (#1925); React hört auf `change`.
async function herkunft(p, wert) {
  await p.seite(`const s = [...document.querySelectorAll('select')].find((s) => [...s.labels].some((l) => l.innerText.trim().startsWith('Herkunft'))); const setze = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set; setze.call(s, ${JSON.stringify(wert)}); s.dispatchEvent(new Event('change', { bubbles: true })); return true`)
}

async function zurSeite(p, beschriftung) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld(beschriftung)
  await p.klickeAuf(`[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(beschriftung)})`, `»${beschriftung}« in der Palette`)
}

async function zuDenPaketen(p) {
  await zurSeite(p, 'Pakete')
  await p.warte('der Arbeitsplatz steht', () => p.seite(`return !!${zeile('probe-fremd')}`), 20000)
}

// fassung legt eine Fassung des Stils in `ort` an.
function fassung(ort, version) {
  const paket = join(ort, `v${version}`, 'probe-fremd.mse-style')
  mkdirSync(paket, { recursive: true })
  writeFileSync(join(paket, 'style'), `mse version: 2.0.0\ngame: probe\nshort name: Fremd\nversion: ${version}\n`)

  return paket
}

const fassungIm = async (p) => p.seite(`return ${zeile('probe-fremd')}?.querySelectorAll('td')[2]?.innerText ?? ''`)

export default async function dom01(p) {
  // **Ein eigenes Zuhause**: die Probe bindet ein, ersetzt und löscht Pakete,
  // und die liegen am Vorgabeort unter `HOME` (QA-10.NAT-01).
  const ort = p.ort({ zuhause: true })
  const alt = fassung(ort.ort, '2026-01-01')
  const neu = fassung(ort.ort, '2026-02-02')
  p.mse(ort, 'paket-installieren', alt)
  const setID = JSON.parse(p.mse(ort, 'set-neu', 'Planset', 'probe', 'probe-fremd', '--json')).id
  p.zeige(`Set ${String(setID)} auf probe-fremd`)

  await p.starte({ ort })
  await zuDenPaketen(p)

  p.notiz('AK-01 — der Inspector und der Filter nach Herkunft')
  await p.klickeAuf(zeile('probe-fremd'), 'die Zeile probe-fremd')
  await p.wartetext('Ein Set benutzt es.', 10000)
  const inspector = await p.seite("return document.body.innerText")
  p.pruefe('der Inspector nennt die Herkunft', inspector.includes('Übernommen aus'), '')
  p.pruefe('der Inspector nennt die Lizenz als nicht angegeben', inspector.includes('nicht angegeben'))
  await p.bild('inspector')
  await herkunft(p, 'eigen')
  await p.warte('der Filter blendet es aus', () => p.seite(`return !${zeile('probe-fremd')}`))
  await herkunft(p, 'uebernommen')
  await p.warte('der Filter zeigt es wieder', () => p.seite(`return !!${zeile('probe-fremd')}`))
  await p.bild('filter-uebernommen')
  await herkunft(p, '')

  p.notiz('AK-02/03 — erst der Plan, dann das Ersetzen')
  await klicke(p, 'Paket einbinden …')
  await p.dialogZeigtSich('oeffnen', '»Paket einbinden …«')
  await p.dialogantwort(neu)
  await klicke(p, 'Paket einbinden …')
  await p.wartetext('probe-fremd.mse-style ersetzen', 15000)
  const plan = await p.seite("return document.body.innerText")
  p.pruefe('der Plan nennt die Fassungen', plan.includes('Fassung 2026-01-01 → 2026-02-02'))
  p.pruefe('der Plan nennt das betroffene Set', plan.includes('Planset'))
  p.pruefe('der Plan nennt die Quelle', plan.includes(neu))
  p.pruefe('vor dem Bestätigen ist nichts ersetzt', (await fassungIm(p)) === '2026-01-01', await fassungIm(p))
  await p.bild('plan-ersetzen')
  await klicke(p, 'Ersetzen')
  await p.warte('die neue Fassung steht in der Zeile', async () => (await fassungIm(p)) === '2026-02-02', 15000)

  p.notiz('AK-02 — zurück zur vorigen Fassung')
  await p.klickeAuf(zeile('probe-fremd'), 'die Zeile probe-fremd')
  await klicke(p, 'Vorige Fassung zurückholen …')
  await p.wartetext('auf die vorige Fassung zurücksetzen', 10000)
  await p.bild('plan-zurueck')
  await klicke(p, 'Zurücksetzen')
  await p.warte('die vorige Fassung steht wieder da', async () => (await fassungIm(p)) === '2026-01-01', 15000)
  const katalog = JSON.parse(p.mse(ort, 'pakete', '--json')).pakete.find((eines) => eines.name === 'probe-fremd')
  p.pruefe('die Kommandozeile sieht dieselbe Fassung', katalog?.version === '2026-01-01', katalog?.version ?? '')

  p.notiz('AK-02 — Löschen zeigt den Plan, und »Nichts ändern« löscht nicht')
  await p.klickeAuf(zeile('probe-fremd'), 'die Zeile probe-fremd')
  await klicke(p, 'Löschen …')
  await p.wartetext('probe-fremd.mse-style löschen', 10000)
  await p.bild('plan-loeschen')
  await klicke(p, 'Nichts ändern')
  await p.warte('der Plan ist fort', () => p.seite("return !document.body.innerText.includes('probe-fremd.mse-style löschen')"))
  p.pruefe('das Paket steht noch da', await p.seite(`return !!${zeile('probe-fremd')}`))

  p.notiz('AK-04 — was dem Set fehlt, und der Sprung zur Suche')
  p.mse(ort, 'paket-entfernen', 'mse-style', 'probe-fremd')
  const bedarf = JSON.parse(p.mse(ort, 'paketbedarf', String(setID), '--json')).befunde
  p.pruefe('die Kommandozeile nennt den fehlenden Stil', bedarf.some((b) => b.braucht === 'probe-fremd.mse-style'), JSON.stringify(bedarf))
  await p.taste('2', 'befehl')
  await p.wartetext('Planset', 20000)
  await p.klickeText('Planset')
  await p.warte('das Set ist offen', () => p.seite("return (document.querySelector('nav ol')?.innerText ?? '').includes('Planset')"), 20000)
  await zurSeite(p, 'Setangaben')
  await p.wartetext('Pakete, die diesem Set fehlen', 20000)
  // **Auf die Liste gewartet, nicht auf die Überschrift** — sie steht, bevor
  // der Bedarf geladen ist; bis zum 2026-10-05 prüfte das Drehbuch sofort.
  await p.warte('die Setangaben nennen den Stil', () => p.seite("return document.body.innerText.includes('probe-fremd.mse-style')"), 10000)
  p.pruefe('die Setangaben nennen den Stil', true)
  await p.bild('setangaben-bedarf')
  await klicke(p, 'Bei den Paketen suchen')
  // **Das sichtbare Suchfeld** — besuchte Seiten bleiben verborgen im Baum
  // (`Activity`), und das erste Suchfeld im Dokument kann ein fremdes sein.
  await p.warte('die Suche ist gefüllt', () => p.seite("return [...document.querySelectorAll('input[type=search]')].find((e) => e.checkVisibility())?.value === 'probe-fremd'"), 15000)
  await p.bild('paketsuche')

  await p.beende()
}
