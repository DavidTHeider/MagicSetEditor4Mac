// **Derselbe Druck zweimal, am nativen Fenster** — DOM-03, S04.NAT-01 bis 03
// (#1698):
//
// - S04.NAT-01: ein zweites Mal derselbe Druck fragt nach der Menge, statt
//   nur abzusagen — und *Auf N erhöhen* erhöht die vorhandene Zeile,
// - S04.NAT-03: Escape bricht ab, ohne eine Menge zu ändern, und die Eingabe
//   bleibt stehen; *Zur vorhandenen Zeile* wählt sie im Inspector,
// - der Fokus steht beim Erscheinen auf dem Abbrechen.
//
// Die Zeilen selbst — gleiche Sprache und Ausführung, andere Sprache eine
// eigene Zeile (S04.NAT-02) — belegen `internal/domain/dublette_test.go` und
// `internal/app/doppelteaufnahme_test.go`.
import { randomUUID } from 'node:crypto'

const feld = (beschriftung) => `[...document.querySelectorAll('input')].find((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').trim() === ${JSON.stringify(beschriftung)})`

const mengeVon = (p, ort, druck) => {
  const zeile = p.mse(ort, 'sammlung-ausgeben', 'csv').trim().split('\n').find((z) => z.startsWith(druck)) ?? ''
  return zeile.split(',')[1] ?? ''
}

// inspector findet die rechte Spalte, die das gewählte Exemplar zeigt.
const inspector = "const r = [...document.querySelectorAll('aside')].find((a) => a.checkVisibility() && a.getAttribute('aria-label') === 'Exemplar')"

const frage = "return [...document.querySelectorAll('[role=group]')].find((g) => g.checkVisibility() && g.innerText.includes('liegt schon in der Sammlung'))?.innerText ?? ''"

// **Das Formular steht in einer Schublade** — seit UXR-10 (#1918) hinter
// *+ Aufnehmen …* im Kopf der Seite, vorher in einem Kasten im Navigator.
// Das Zeichen des Katalogs steht vor dem Text („+ Aufnehmen …"); gesucht wird
// deshalb nach dem Ende.
const kopfknopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && !b.closest('[inert]') && b.innerText.trim().endsWith(${JSON.stringify(name)}))`

async function zurSammlung(p) {
  await p.taste('3', 'befehl')
  await p.warte('der Kopf der Sammlung steht', () => p.seite(`return !!${kopfknopf('Aufnehmen …')}`), 20000)
  await p.klickeAuf(kopfknopf('Aufnehmen …'), 'Aufnehmen …')
  await p.warte('die Schublade steht', () => p.seite("return [...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility() && d.getAttribute('aria-label') === 'Ein Exemplar aufnehmen')"), 20000)
}

// nochmalAufnehmen tippt den Druck ins leere Feld und drückt *Aufnehmen*.
//
// **Die Kennung steht seit AP-10 (#2367) unter „Weitere Angaben“** — der Name
// führt; aufgeklappt wird nur, wenn sie noch zu ist.
const weitereAngaben = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Weitere Angaben')"
async function nochmalAufnehmen(p, druck) {
  if (!(await p.seite(`return !!${feld('Druck')}`))) {
    await p.klickeAuf(weitereAngaben, '»Weitere Angaben«')
    await p.warte('das Feld »Druck« steht', () => p.seite(`return !!${feld('Druck')}`))
  }
  await p.seite(`const e = ${feld('Druck')}; Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(e, ''); e.dispatchEvent(new Event('input', { bubbles: true })); return true`)
  await p.klickeAuf(feld('Druck'), 'das Feld »Druck«')
  await p.tippe(druck)
  await p.warte('im Feld steht der Druck', () => p.seite(`return ${feld('Druck')}?.value === ${JSON.stringify(druck)}`))
  await p.klickeText('Aufnehmen')
  return p.warte('die Frage nach der Menge steht da', async () => (await p.seite(frage)) || false, 15000)
}

export default async function dom03(p) {
  const ort = p.ort()
  const druck = randomUUID()
  p.mse(ort, 'exemplar', druck, '2')
  await p.starte({ ort })
  await zurSammlung(p)

  p.notiz('S04.NAT-01 — derselbe Druck ein zweites Mal')
  const text = await nochmalAufnehmen(p, druck)
  p.zeige(`Frage: ${text.replace(/\s+/g, ' ').slice(0, 200)}`)
  p.pruefe('die Frage nennt die vorhandene Menge', text.includes('2 Stück'))
  p.pruefe('der Fokus steht auf dem Abbrechen', (await p.seite('return document.activeElement?.innerText?.trim() ?? ""')) === 'Abbrechen')
  await p.bild('frage-nach-der-menge')

  p.notiz('S04.NAT-03 — Escape ändert keine Menge und lässt die Eingabe stehen')
  await p.taste('Escape')
  await p.warte('die Frage ist weg', async () => (await p.seite(frage)) === '', 10000)
  p.pruefe('die Menge bleibt zwei', mengeVon(p, ort, druck) === '2', mengeVon(p, ort, druck))
  p.pruefe('die Eingabe bleibt stehen', await p.seite(`return ${feld('Druck')}?.value === ${JSON.stringify(druck)}`))

  p.notiz('S04.NAT-01 — auf drei erhöhen')
  await nochmalAufnehmen(p, druck)
  await p.klickeText('Auf 3 erhöhen')
  await p.warte('die Menge ist drei', () => mengeVon(p, ort, druck) === '3', 15000)
  p.pruefe('die vorhandene Zeile hat drei Stück', true)
  // **Geräumt heißt seit AP-10 auch: wieder zugeklappt** — das Formular
  // entsteht neu, der Name ist leer, und die Kennung steht leer darunter.
  await p.warte('das Feld »Karte« ist leer', () => p.seite(`return ${feld('Karte')}?.value === ''`), 10000)
  if (!(await p.seite(`return !!${feld('Druck')}`))) await p.klickeAuf(weitereAngaben, '»Weitere Angaben«')
  p.pruefe('und das Formular ist geräumt', await p.warte('das Feld ist leer', () => p.seite(`return ${feld('Druck')}?.value === ''`), 10000))
  // **Auf die Wirkung warten** (§3): der Bestand sagt drei, bevor das
  // Nachladen die Liste erreicht — das erste Bild zeigte noch zwei.
  await p.warte('Liste und Inspector zeigen drei', () => p.seite(`${inspector}; const liste = document.querySelector('[role=grid]')?.innerText ?? ''; return !!r && r.innerText.includes('3 × Near Mint') && liste.includes('3 × Near Mint')`), 15000)
  p.pruefe('Liste und Inspector zeigen drei Stück', true)
  await p.bild('erhoeht')

  p.notiz('S04.NAT-03 — zur vorhandenen Zeile')
  await nochmalAufnehmen(p, druck)
  await p.klickeText('Zur vorhandenen Zeile')
  // **Die Schublade geht dabei zu** (UXR-10) — der Inspector darunter zeigt
  // die Zeile.
  await p.warte('die Schublade ist zu', () => p.seite("return ![...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility())"), 10000)
  await p.warte('der Inspector zeigt die Zeile', () => p.seite(`${inspector}; return !!r && r.innerText.includes(${JSON.stringify(druck)}) && r.innerText.includes('3 × Near Mint')`), 15000)
  p.pruefe('die Zeile steht im Inspector, ohne dass sich etwas änderte', mengeVon(p, ort, druck) === '3')
  await p.bild('zur-zeile')

  p.pruefe('⌘Q beendet', await p.beende())
}
