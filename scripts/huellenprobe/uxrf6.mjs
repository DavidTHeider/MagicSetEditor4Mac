// **Ein Balken der Statistik filtert die Kartenliste, am nativen Fenster** —
// UXR-F6 (#2013),
// ein Kompatibilitätsrest aus UXR-08 (§15).
//
// MSE2 zeigt nach der Wahl einer Gruppe nur ihre Karten
// (`StatsPanel::onGraphSelect` → `filterCards`, `src/gui/set/stats_panel.cpp`
// Zeilen 526–547, Fassung f9d9356 vom 2021-07-06). Dieses Drehbuch fährt den
// Weg am Fenster, mit dem Probeset (fünf Karten, `farbe`: drei rot, zwei blau)
// bei 1280 × 800:
//
// - **die Verteilung des Felds `farbe`**: die Balken sind Knöpfe, der Satz
//   „Einen Balken wählen …" steht da, ein Klick wählt (`aria-pressed`), ein
//   zweiter hebt auf, Enter wählt mit der Tastatur;
// - **der Weg zur Liste**: *Karten anzeigen* schließt die Schublade, die Liste
//   zeigt **nur** die zwei blauen Karten (gezählt, benannt), und die Zeile
//   darüber sagt „Aus der Statistik …" und „2 von 5 Karten";
// - **der Rückweg**: *Zur Statistik* öffnet die Schublade mit dem Balken noch
//   gewählt; *Wahl aufheben* und *Filter aufheben* geben die fünf Karten
//   wieder;
// - **die Auszählung nach der erklärten Dimension `farbe`**: dieselbe Wahl;
// - **die Belegung**: ihre Balken sind keine Knöpfe.
//
// **Was es nicht belegt**: ob jemand findet, dass ein Balken wählbar ist (der
// Satz darüber ist die Hilfe, nicht der Beweis) — das sieht ein Mensch an den
// Bildern. Und die Pfeiltasten von MSE2 gibt es hier nicht (`docs/PARITAET.md`).
import { oeffnen, zurSeite } from './uirevue.mjs'

// zeilen sind die Karten, die die Liste gerade zeigt.
const zeilen = "[...document.querySelectorAll('[role=grid] [role=row][data-karte]')].filter((r) => r.checkVisibility())"
const namen = `${zeilen}.map((r) => r.getAttribute('aria-label'))`

// schublade ist die offene Schublade (der Dialog) — oder keine.
const schublade = "[...document.querySelectorAll('[role=dialog]')].find((d) => d.checkVisibility())"

// balken ist der Knopf eines Balkens in der Schublade.
const balken = (name) => `[...(${schublade})?.querySelectorAll('button[aria-pressed]') ?? []].find((b) => b.checkVisibility() && b.innerText.trim().startsWith(${JSON.stringify(name)}))`

// gedrueckt sagt, ob der Balken gewählt ist.
const gedrueckt = (name) => `return ${balken(name)}?.getAttribute('aria-pressed') === 'true'`

// wahl setzt die Wahl „Worüber" — das Feld ist ein natives Auswahlfeld.
const wahl = (wert) => `const s = [...(${schublade})?.querySelectorAll('select') ?? []].find((e) => (e.labels?.[0]?.innerText ?? '').trim() === 'Worüber'); if (!s) return false; Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set.call(s, ${JSON.stringify(wert)}); s.dispatchEvent(new Event('change', { bubbles: true })); return true`

// zeile ist die Zeile über der Kartenliste, wenn die Statistik filtert.
const zeile = "document.querySelector('[role=group][aria-label=\"Filter aus der Statistik\"]')"

const knopf = (text, wurzel = 'document') => `[...${wurzel}.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(text)})`

// warteAufKarten wartet, bis die Liste so viele Karten zeigt.
async function warteAufKarten(p, anzahl, was) {
  await p.warte(was, () => p.seite(`return (${zeilen}).length === ${String(anzahl)}`))
}

export default async function uxrf6(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  await oeffnen(p, ort, { name: 'normal', breite: 1280, hoehe: 800 })
  // **Zur Kartenliste** — das Öffnen endet auf der Seite des Sets, nicht
  // zwingend auf den Karten.
  await zurSeite(p, 'Karten')
  await new Promise((los) => setTimeout(los, 1500))
  // **Der Navigator trägt die Kartenliste** — steht er zu, erst aufklappen.
  if (await p.seite(`return !!${knopf('Navigator einblenden')}`)) {
    await p.klickeAuf(knopf('Navigator einblenden'), 'Navigator einblenden')
    await new Promise((los) => setTimeout(los, 1000))
  }
  p.zeige(`auf der Seite: ${JSON.stringify(await p.seite(`return { titel: document.querySelector('main h1, main h2')?.innerText.trim() ?? '', zeilen: (${zeilen}).length, gitter: document.querySelectorAll('[role=grid]').length, reihen: document.querySelectorAll('[role=row]').length, karten: document.querySelectorAll('[data-karte]').length, sichtbar: [...document.querySelectorAll('[data-karte]')].filter((e) => e.checkVisibility()).length }`))}`)
  await p.bild('karten-vorher')
  await warteAufKarten(p, 5, 'die Liste zeigt alle fünf Karten')
  p.zeige(`Karten vorher: ${JSON.stringify(await p.seite(`return ${namen}`))}`)
  p.pruefe('ohne Wahl steht keine Zeile über der Liste', await p.seite(`return ${zeile} === null`))

  p.notiz('Die Verteilung des Felds »farbe« — die Balken sind Knöpfe')
  await zurSeite(p, 'Statistik')
  await p.warte('die Schublade steht', () => p.seite(`return !!(${schublade})`), 20000)
  await p.warte('das Feld Worüber steht', () => p.seite(wahl('feld:farbe')))
  await p.warte('die Balken stehen', () => p.seite(`return !!(${balken('rot')}) && !!(${balken('blau')})`))
  const knoepfe = await p.seite(`return [...(${schublade}).querySelectorAll('button[aria-pressed]')].filter((b) => b.checkVisibility()).map((b) => ({ text: b.innerText.replace(/\\s+/g, ' ').trim(), gedrueckt: b.getAttribute('aria-pressed'), tabstelle: b.tabIndex }))`)
  p.zeige(`Balken: ${JSON.stringify(knoepfe)}`)
  p.pruefe('zwei Balken, beide Knöpfe in der Tabfolge und keiner gewählt', knoepfe.length === 2 && knoepfe.every((k) => k.gedrueckt === 'false' && k.tabstelle >= 0))
  p.pruefe('der Satz „Einen Balken wählen …" steht da', await p.seite(`return (${schublade}).innerText.includes('Einen Balken wählen, um seine Karten in der Kartenliste zu sehen')`))
  await p.bild('statistik-balken-ungewaehlt')

  p.notiz('Wählen mit der Maus — und aufheben mit demselben Klick')
  await p.klickeAuf(balken('rot'), 'der Balken »rot«')
  await p.warte('rot ist gewählt', () => p.seite(gedrueckt('rot')))
  p.pruefe('blau bleibt ungewählt', !(await p.seite(gedrueckt('blau'))))
  const gewaehlt = await p.seite(`return (${schublade}).innerText.replace(/\\s+/g, ' ')`)
  p.pruefe('die Wahl steht in Worten: »Gewählt: »rot« · Karten: 3«', gewaehlt.includes('Gewählt: »rot« · Karten: 3'))
  // **Höchstens ein Hauptknopf in der Schublade** (R1): *Karten anzeigen*.
  const haupt = await p.seite(`return [...(${schublade}).querySelectorAll('button')].filter((b) => b.checkVisibility() && /(^|\\s)bg-marke(\\s|$)/.test(b.className)).map((b) => b.innerText.trim())`)
  p.pruefe('*Karten anzeigen* ist der einzige Hauptknopf der Schublade', haupt.length === 1 && haupt[0] === 'Karten anzeigen', JSON.stringify(haupt))
  await p.bild('statistik-balken-rot')

  await p.klickeAuf(balken('rot'), 'der Balken »rot« noch einmal')
  await p.warte('rot ist wieder frei', () => p.seite(`return !(${balken('rot')}).matches('[aria-pressed=true]')`))
  p.pruefe('derselbe Klick hebt die Wahl auf, und der Satz steht wieder da', await p.seite(`return (${schublade}).innerText.includes('Einen Balken wählen')`))

  p.notiz('Wählen mit der Tastatur — Tab-Stelle und Enter')
  await p.seite(`${balken('blau')}.focus(); return true`)
  p.pruefe('der Balken »blau« nimmt den Fokus', await p.seite(`return document.activeElement === ${balken('blau')}`))
  await p.taste('Enter')
  await p.warte('blau ist gewählt', () => p.seite(gedrueckt('blau')))
  p.pruefe('Enter wählt »blau«, und rot ist ungewählt', !(await p.seite(gedrueckt('rot'))))

  p.notiz('Der Weg zur Liste — nur die Karten des Balkens')
  await p.klickeText('Karten anzeigen')
  await p.warte('die Schublade ist zu', () => p.seite(`return !(${schublade})`))
  await warteAufKarten(p, 2, 'die Liste zeigt nur die zwei blauen Karten')
  const blau = await p.seite(`return ${namen}`)
  p.zeige(`Karten gefiltert: ${JSON.stringify(blau)}`)
  p.pruefe('die zwei blauen Karten, und nur sie', blau.length === 2 && blau.every((n) => /Sonderlayout|Ohne Bild/.test(n)))
  const kopf = await p.seite(`return ${zeile}?.innerText.replace(/\\s+/g, ' ') ?? null`)
  p.zeige(`Zeile über der Liste: ${JSON.stringify(kopf)}`)
  p.pruefe('die Zeile nennt den Filter und zählt „2 von 5 Karten"', !!kopf && kopf.includes('Aus der Statistik: »farbe« = »blau«') && kopf.includes('2 von 5 Karten'))
  await p.bild('karten-gefiltert')

  p.notiz('Der Rückweg — die Wahl bleibt, die Schublade geht auf')
  await p.klickeAuf(knopf('Zur Statistik', zeile), 'Zur Statistik')
  await p.warte('die Schublade steht wieder', () => p.seite(`return !!(${schublade})`))
  await p.warte('blau ist noch gewählt', () => p.seite(gedrueckt('blau')))
  p.pruefe('der Balken »blau« steht gewählt da', true)
  await p.klickeText('Wahl aufheben')
  await p.warte('nichts ist gewählt', () => p.seite(`return !(${balken('blau')}).matches('[aria-pressed=true]')`))
  await p.taste('Escape')
  await p.warte('die Schublade ist zu', () => p.seite(`return !(${schublade})`))
  await warteAufKarten(p, 5, 'die Liste zeigt wieder alle fünf Karten')
  p.pruefe('*Wahl aufheben* gibt alle fünf Karten wieder, und die Zeile ist fort', await p.seite(`return ${zeile} === null`))

  p.notiz('Aufheben auf der Liste — der Knopf in der Zeile')
  await zurSeite(p, 'Statistik')
  await p.warte('der Balken »rot« steht', () => p.seite(`return !!(${balken('rot')})`))
  await p.klickeAuf(balken('rot'), 'der Balken »rot«')
  await p.klickeText('Karten anzeigen')
  await warteAufKarten(p, 3, 'die Liste zeigt die drei roten Karten')
  p.pruefe('drei rote Karten', (await p.seite(`return ${zeilen}.length`)) === 3)
  await p.klickeAuf(knopf('Filter aufheben', zeile), 'Filter aufheben')
  await warteAufKarten(p, 5, 'die Liste zeigt wieder alle fünf Karten')
  p.pruefe('*Filter aufheben* gibt alle fünf Karten wieder', await p.seite(`return ${zeile} === null`))

  p.notiz('Die Auszählung nach der erklärten Dimension »farbe« — dieselbe Wahl')
  await zurSeite(p, 'Statistik')
  await p.warte('die Schublade steht', () => p.seite(`return !!(${schublade})`))
  await p.warte('Auszählung gewählt', () => p.seite(wahl('dimension:farbe')))
  await p.warte('die Gruppen stehen', () => p.seite(`return !!(${balken('blau')})`), 30000)
  await p.klickeAuf(balken('blau'), 'die Gruppe »blau«')
  await p.klickeText('Karten anzeigen')
  await warteAufKarten(p, 2, 'die Liste zeigt die zwei blauen Karten')
  const dimension = await p.seite(`return ${zeile}?.innerText.replace(/\\s+/g, ' ') ?? null`)
  p.zeige(`Zeile über der Liste (Dimension): ${JSON.stringify(dimension)}`)
  p.pruefe('dieselben zwei Karten, wenn nach der Dimension gewählt ist', !!dimension && dimension.includes('2 von 5 Karten'))
  await p.klickeAuf(knopf('Filter aufheben', zeile), 'Filter aufheben')
  await warteAufKarten(p, 5, 'die Liste zeigt wieder alle fünf Karten')

  p.notiz('Die Belegung bleibt zum Lesen')
  await zurSeite(p, 'Statistik')
  await p.warte('die Schublade steht', () => p.seite(`return !!(${schublade})`))
  await p.warte('Belegung gewählt', () => p.seite(wahl('belegung')))
  await p.warte('die Belegung steht', () => p.seite(`return (${schublade}).innerText.includes('Belegung je Feld')`))
  p.pruefe('kein Balken der Belegung ist ein Knopf', (await p.seite(`return [...(${schublade}).querySelectorAll('button[aria-pressed]')].length`)) === 0)
  await p.bild('statistik-belegung')

  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
