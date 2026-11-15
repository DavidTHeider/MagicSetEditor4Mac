// **Anlage und Ansicht in einer ruhigen Zeile** — P03
// (#2174).
//
// Über dem Arbeitsplatz der Karten stehen „Karte anlegen" und „Mehr", und
// darunter in einer eigenen Zeile die Schalter für Navigator und Inspector.
// P03 verlangt zwei Gruppen in **einer** Zeile — Anlage und Ansicht —, die in
// geringerer Breite geordnet bleiben, ohne neuen Hauptknopf und ohne dass ein
// Umschalten oder ein Größenwechsel einen Entwurf verliert.
//
// Gemessen wird am nativen Fenster, je Deutsch hell und Englisch dunkel, bei
// 1280 × 800 und 900 × 700:
//
// - **die Zeilen**: wie viele Reihen die vier Bedienelemente bilden,
// - **die Gruppen**: ob „anlegen" und „Mehr" eine benannte Gruppe teilen und
//   die zwei Schalter eine zweite,
// - **die Tab-Folge** von „anlegen" an,
// - **„Mehr"** öffnet und schließt mit Escape, der Fokus geht zurück,
// - **der Navigator zu**: „anlegen" bleibt sichtbar,
// - **ein Entwurf** im Namen übersteht einen Größenwechsel ungeschrieben, mit
//   dem Fokus im Feld, und ein Umschalten, das das Feld verlässt — dann steht
//   er im Bestand, nicht verloren.
//
// **Was es nicht belegt**: die Bedienung mit dem Vorleseprogramm.
import { zurSeite } from './uirevue.mjs'

const sprachen = {
  de: { karten: 'Karten', anlegen: 'Karte anlegen', mehr: 'Mehr', anlage: 'Karte', ansicht: 'Ansicht' },
  en: { karten: 'Cards', anlegen: 'Add card', mehr: 'More', anlage: 'Card', ansicht: 'View' },
}

const groessen = [
  { name: 'breit', breite: 1280, hoehe: 800 },
  { name: 'schmal', breite: 900, hoehe: 700 },
]

// knopf findet einen Knopf an seinem Text — **ohne das Zeichen davor**: der
// Hauptknopf trägt ein „+".
const knopf = (t) => `[...document.querySelectorAll('main button')].find((b) => b.checkVisibility() && b.innerText.trim().replace(/^[^\\p{L}]+/u, '') === ${JSON.stringify(t)})`
// schalter findet den Schalter einer Region — an `aria-expanded` (`aria-controls` trägt er nur offen) und an dem,
// was er sagt: sichtbarer Text oder Name.
const schalter = (art) => `[...document.querySelectorAll('main button[aria-expanded]')].find((b) => b.checkVisibility() && /${art}/i.test((b.getAttribute('aria-label') ?? '') + ' ' + b.innerText))`
const namensfeld = "document.querySelector('input[aria-label=\"name\"], textarea[aria-label=\"name\"]')"

// zeile misst die vier Bedienelemente und ihre Gruppen.
async function zeile(p, t) {
  return p.seite(`
    const teile = { anlegen: ${knopf(t.anlegen)}, mehr: ${knopf(t.mehr)}, navigator: ${schalter('navigator')}, inspector: ${schalter('inspector')} }
    const lage = {}
    for (const [name, e] of Object.entries(teile)) {
      if (!e) { lage[name] = null; continue }
      const r = e.getBoundingClientRect()
      lage[name] = { oben: Math.round(r.top), links: Math.round(r.left), gruppe: e.closest('[role=group]')?.getAttribute('aria-label') ?? null }
    }
    const oben = Object.values(lage).filter(Boolean).map((x) => x.oben)
    const reihen = [...new Set(oben.map((o) => Math.round(o / 8)))].length
    return { lage, reihen }
  `)
}

export default async function p03(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  for (const [sprache, bild] of [['de', 'hell'], ['en', 'dunkel']]) {
    const t = sprachen[sprache]
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', sprache)
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)

    for (const groesse of groessen) {
      const wo = `${sprache}, ${bild}, ${groesse.name}`
      p.notiz(`${wo} (${String(groesse.breite)} × ${String(groesse.hoehe)})`)
      p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
      p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', String(groesse.breite))
      p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', String(groesse.hoehe))
      await p.starte({ ort })
      await p.taste('2', 'befehl')
      await p.warte('die Sets stehen', () => p.seite("return !!document.querySelector('main article')"), 20000)
      await p.klickeAuf("[...document.querySelectorAll('main article :is(h2, h3) button, main article :is(h2, h3) a')].find((b) => b.checkVisibility())", 'das erste Set')
      await zurSeite(p, t.karten, sprache)
      await p.warte('die Zeile steht', () => p.seite(`return !!${knopf(t.anlegen)} && !!${schalter('inspector')}`), 20000)

      const gemessen = await zeile(p, t)
      p.zeige(`${wo}: ${JSON.stringify(gemessen)}`)
      const { lage } = gemessen
      p.pruefe(`${wo}: alle vier Bedienelemente stehen da`, Object.values(lage).every(Boolean))
      p.pruefe(`${wo}: „${t.anlegen}" und „${t.mehr}" bilden die Gruppe „${t.anlage}"`, lage.anlegen?.gruppe === t.anlage && lage.mehr?.gruppe === t.anlage, `${String(lage.anlegen?.gruppe)} / ${String(lage.mehr?.gruppe)}`)
      p.pruefe(`${wo}: die zwei Schalter bilden die Gruppe „${t.ansicht}"`, lage.navigator?.gruppe === t.ansicht && lage.inspector?.gruppe === t.ansicht, `${String(lage.navigator?.gruppe)} / ${String(lage.inspector?.gruppe)}`)
      p.pruefe(`${wo}: jede Gruppe steht auf einer Linie`, Math.abs((lage.anlegen?.oben ?? 0) - (lage.mehr?.oben ?? -99)) <= 4 && Math.abs((lage.navigator?.oben ?? 0) - (lage.inspector?.oben ?? -99)) <= 4)
      if (groesse.name === 'breit') p.pruefe(`${wo}: beide Gruppen in einer Zeile`, gemessen.reihen === 1, `${String(gemessen.reihen)} Reihen`)
      await p.bild(`p03-${sprache}-${bild}-${groesse.name}`)

      // **Die Tab-Folge** von „anlegen" an — drei Drücke.
      await p.seite(`${knopf(t.anlegen)}.focus(); return true`)
      const folge = []
      for (let druck = 0; druck < 3; druck += 1) {
        await p.taste('Tab')
        await new Promise((los) => { setTimeout(los, 150) })
        folge.push(await p.seite("const a = document.activeElement; return ((a?.getAttribute('aria-label') ?? '') || (a?.innerText ?? '')).trim().replace(/^[^\\p{L}]+/u, '').slice(0, 40)"))
      }
      p.zeige(`${wo}: Tab von „${t.anlegen}": ${folge.join(' → ')}`)
      p.pruefe(`${wo}: nach „${t.anlegen}" folgen „${t.mehr}" und die zwei Schalter`, folge[0] === t.mehr && /navigator/i.test(folge[1]) && /inspector/i.test(folge[2]), folge.join(' → '))

      // **„Mehr" auf und mit Escape zu**, der Fokus zurück.
      await p.klickeAuf(knopf(t.mehr), `»${t.mehr}«`)
      const auf = await p.warte('das Menü steht', () => p.seite("return !!document.querySelector('[role=menu]')?.checkVisibility()"), 5000)
      await p.taste('Escape')
      const zu = await p.warte('das Menü ist zu und der Fokus zurück', () => p.seite(`return !document.querySelector('[role=menu]')?.checkVisibility() && document.activeElement === ${knopf(t.mehr)}`), 5000)
      p.pruefe(`${wo}: „${t.mehr}" öffnet und schließt, der Fokus kehrt zurück`, Boolean(auf) && Boolean(zu))

      // **Der Navigator zu**, und „anlegen" bleibt.
      const navigatorOffen = await p.seite(`return ${schalter('navigator')}.getAttribute('aria-expanded') === 'true'`)
      if (navigatorOffen) {
        await p.klickeAuf(schalter('navigator'), 'den Navigator schließen')
        await p.warte('der Navigator ist zu', () => p.seite(`return ${schalter('navigator')}.getAttribute('aria-expanded') === 'false'`), 5000)
      }
      p.pruefe(`${wo}: bei geschlossenem Navigator steht „${t.anlegen}" da`, await p.seite(`return !!${knopf(t.anlegen)}`))

      // **Ein Entwurf übersteht den Größenwechsel**, ungeschrieben — der Fokus
      // bleibt im Feld, also hat niemand es verlassen.
      await p.warte('das Namensfeld steht', () => p.seite(`return !!${namensfeld}`), 10000)
      const vorher = await p.seite(`return ${namensfeld}.value`)
      const bestand = () => JSON.parse(p.mse(ort, 'karten', '1', '--json')).karten.map((k) => k.name)
      await p.seite(`const e = ${namensfeld}; e.focus(); e.setSelectionRange(e.value.length, e.value.length); return true`)
      await p.tippeInsFeld(' Entwurf')
      await p.groesse(groesse.breite - 100, groesse.hoehe)
      await new Promise((los) => { setTimeout(los, 600) })
      await p.groesse(groesse.breite, groesse.hoehe)
      await new Promise((los) => { setTimeout(los, 600) })
      const nachGroesse = await p.seite(`return { wert: ${namensfeld}?.value, fokus: document.activeElement === ${namensfeld} }`)
      p.pruefe(`${wo}: nach dem Größenwechsel steht der Entwurf im Feld, mit dem Fokus`, nachGroesse.wert === `${vorher} Entwurf` && nachGroesse.fokus, JSON.stringify(nachGroesse))
      p.pruefe(`${wo}: und ungeschrieben`, bestand().includes(vorher) && !bestand().includes(`${vorher} Entwurf`))

      // **Und das Umschalten verliert ihn nicht**: der Klick auf den Schalter
      // verlässt das Feld, und Verlassen schreibt (U01) — danach steht der
      // Entwurf im Feld **und** im Bestand, nicht der alte Wert.
      await p.klickeAuf(schalter('inspector'), 'den Inspector umschalten')
      await new Promise((los) => { setTimeout(los, 400) })
      await p.klickeAuf(schalter('inspector'), 'den Inspector zurückschalten')
      await p.warte('das Namensfeld steht wieder', () => p.seite(`return !!${namensfeld}`), 10000)
      await p.warte('der Entwurf ist im Bestand', () => bestand().includes(`${vorher} Entwurf`), 10000)
      p.pruefe(`${wo}: nach dem Umschalten steht der Entwurf im Feld und im Bestand`, (await p.seite(`return ${namensfeld}.value`)) === `${vorher} Entwurf`)
      await p.seite(`const e = ${namensfeld}; e.focus(); const setze = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set; setze.call(e, ${JSON.stringify(vorher)}); e.dispatchEvent(new Event('input', { bubbles: true })); e.blur(); return true`)
      await p.warte('der alte Name ist zurück', () => bestand().includes(vorher), 10000)

      p.pruefe('⌘Q beendet', await p.beende())
    }
  }
}
