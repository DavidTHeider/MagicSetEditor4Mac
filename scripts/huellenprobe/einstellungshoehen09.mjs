// **Was in einer Gruppe der Einstellungen die Höhe macht** — #2465
// (Issue (#2465)),
// „zu klären, bevor gebaut wird“.
//
// # Was es misst
//
// Je Gruppe über R4 (Rolltiefe ≤ 2,5) und Fenstergröße die Höhe jedes Teils
// der Gruppe — Überschrift, Einstellungen, Abschnitte, Beiwerk, „Für
// Fachleute“ — in Punkten und als Anteil an der sichtbaren Höhe, dazu die
// Rolltiefe wie `ap11ruhe09` (`messung` aus `uirevue.mjs`). **Ein Teil über
// der halben sichtbaren Höhe wird zerlegt**, bis in die vierte Ebene (je
// Ebene ein ›): erst dort steht, welches Bauteil die Höhe macht. Mit dem
// künstlichen Bestand der Probe, ohne „nur lokal“.
//
// # Was es nicht belegt
//
// **Ob ein Teil entbehrlich ist** — das entscheidet der Bau, und dessen
// Messung vorher und nachher ist `ap11ruhe09`.
import { messung } from './uirevue.mjs'

const gruppen = ['Dienste', 'Daten und Sicherung', 'KI', 'Erscheinungsbild', 'Diagnose']
const groessen = [[1440, 838], [1024, 700]]
const lasche = (name) => `[...document.querySelectorAll('main [role=tab]')].find((t) => t.checkVisibility() && t.innerText.split('\\n')[0].trim() === ${JSON.stringify(name)})`

// teile nennt jedes Kind der offenen Gruppe mit seiner ersten Zeile und Höhe.
const teile = `const h = [...document.querySelectorAll('main h2')].find((e) => e.checkVisibility()); if (!h) return null;
  const roller = (() => { let e = h.parentElement; while (e && !(e.scrollHeight > e.clientHeight + 1 && /(auto|scroll)/.test(getComputedStyle(e).overflowY))) e = e.parentElement; return e ?? document.scrollingElement })();
  const name = (e) => (e.querySelector('h2,h3,h4,legend,summary,button,label') ?? e).innerText.split('\\n')[0].trim().slice(0, 60);
  const hoehe = (e) => Math.round(e.getBoundingClientRect().height);
  const zerlege = (e, tiefe) => [...e.children].filter((k) => k.checkVisibility()).flatMap((k) => [{ was: '›'.repeat(tiefe) + ' ' + name(k), hoehe: hoehe(k) }, ...(hoehe(k) > roller.clientHeight / 2 && tiefe < 4 && k.children.length > 1 ? zerlege(k, tiefe + 1) : [])]);
  return { sichtbar: roller.clientHeight, teile: zerlege(h.parentElement, 0) }`

export default async function einstellungshoehen09(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  await p.starte({ ort })
  await p.taste(',', 'befehl')
  await p.warte('die Einstellungen stehen da', () => p.seite(`return !!${lasche('Allgemein')}`), 20000)

  for (const [breite, hoehe] of groessen) {
    await p.groesse(breite, hoehe)
    for (const name of gruppen) {
      await p.klickeAuf(lasche(name), `die Gruppe »${name}«`)
      await p.warte(`»${name}« ist offen`, () => p.seite(`return ${lasche(name)}?.getAttribute('aria-selected') === 'true'`), 10000)
      await new Promise((weiter) => { setTimeout(weiter, 300) })
      const tiefe = (await p.seite(messung)).rolltiefe
      const gemessen = await p.seite(teile)
      p.zeige(`### ${name} bei ${String(breite)} × ${String(hoehe)} — Rolltiefe ${tiefe.toFixed(1)}, sichtbar ${String(gemessen.sichtbar)} pt`)
      p.zeige('| Teil | Höhe | Anteil an der sichtbaren Höhe |')
      p.zeige('| --- | ---: | ---: |')
      for (const teil of gemessen.teile) p.zeige(`| ${teil.was || '—'} | ${String(teil.hoehe)} | ${(teil.hoehe / gemessen.sichtbar).toFixed(2)} |`)
    }
  }

  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
