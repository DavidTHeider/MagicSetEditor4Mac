// **U02b am nativen Fenster: ein fremder Wert vor der Annahme** —
// #2179.
//
// `phasec` zeigt den Konflikt, wenn das Fenster den fremden Wert **schon
// gesehen** hat. Hier kommt er **im Augenblick des Verlassens**: im Namen ist
// ein Entwurf getippt, `mse` schreibt dasselbe Feld, und gleich danach geht
// der Fokus mit ⇥ weiter — schneller, als das Fenster über `data_version`
// (alle zwei Sekunden) nachliest. Ohne U02b stand danach der Entwurf im
// Bestand und der fremde Wert war fort.
//
// **Zwei Läufe sind möglich, und beide sind richtig** — die Probe sagt, welcher
// es war: kam das Verlassen vor dem Nachlesen, antwortet der Bestand mit
// `fremderWert` am Feld (der Fall, um den es geht); kam das Nachlesen
// zuerst, zeigt die Zeile den Konflikt wie in `phasec`. **In beiden** steht
// danach im Bestand der fremde Wert, nicht der Entwurf.
import { oeffnen, zurSeite } from './uirevue.mjs'

const nachAria = (name) => `document.querySelector('input[aria-label=${JSON.stringify(name)}], textarea[aria-label=${JSON.stringify(name)}]')`
const karteNamens = (p, ort, id) => JSON.parse(p.mse(ort, 'karten', '1', '--json')).karten.find((k) => k.id === id)
const fremderSatz = 'Dieses Feld hat inzwischen jemand anderes geändert'
const fremd = 'Fremd vor der Annahme'

export default async function u02b(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  await oeffnen(p, ort, { name: 'normal', breite: 1280, hoehe: 800 })
  await zurSeite(p, 'Karten')
  await p.warte('das Namensfeld steht da', () => p.seite(`return !!${nachAria('name')}`), 20000)

  const name = await p.seite(`return ${nachAria('name')}.value`)
  const karte = JSON.parse(p.mse(ort, 'karten', '1', '--json')).karten.find((k) => k.name === name)
  p.zeige(`offen: Karte ${karte.id} »${name}«`)

  for (const runde of [1, 2, 3]) {
    p.notiz(`${String(runde)} — Entwurf im Namen, fremdes Schreiben, sofort ⇥`)
    await p.seite(`const e = ${nachAria('name')}; e.focus(); e.setSelectionRange(e.value.length, e.value.length); return true`)
    await p.tippeInsFeld(` Probe ${String(runde)}`)
    const entwurf = await p.seite(`return ${nachAria('name')}.value`)

    const vorher = Date.now()
    p.mse(ort, 'feld', '1', karte.id, 'name', `${fremd} ${String(runde)}`)
    await p.taste('Tab')
    const abstand = Date.now() - vorher

    const weg = await p.warte('das Fenster sagt es', () => p.seite(`const t = document.body.innerText; return t.includes(${JSON.stringify(fremderSatz)}) ? 'fremderWert' : t.includes('Inzwischen steht hier') ? 'Konflikt nach dem Nachlesen' : ''`), 15000)
    p.zeige(`Runde ${String(runde)}: ${weg}, ⇥ ${String(abstand)} ms nach dem fremden Schreiben`)

    // **Gemessen am Bestand und nicht an der Anzeige** — und nach einer
    // Pause, in der ein Nachlesen und ein zweites Verlassen hätten
    // geschehen können.
    await new Promise((los) => { setTimeout(los, 3000) })
    const gespeichert = karteNamens(p, ort, karte.id)?.name
    p.pruefe(`Runde ${String(runde)}: im Bestand steht der fremde Wert, nicht der Entwurf`, gespeichert === `${fremd} ${String(runde)}`, `gespeichert »${String(gespeichert)}«, Entwurf »${entwurf}«`)
    await p.bild(`u02b-runde-${String(runde)}`)

    // **Und der Entwurf ist nicht verloren**: »Meinen Entwurf speichern«
    // schreibt ihn bewusst.
    await p.warte('der Konflikt bietet seine zwei Wege an', () => p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Meinen Entwurf speichern')"), 15000)
    await p.klickeAuf("[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Meinen Entwurf speichern')", '»Meinen Entwurf speichern«')
    await p.warte('der Bestand trägt den Entwurf', () => karteNamens(p, ort, karte.id)?.name === entwurf, 15000)
    p.pruefe(`Runde ${String(runde)}: »behalten« schreibt den Entwurf`, true)
  }

  p.pruefe('⌘Q beendet', await p.beende())
}
