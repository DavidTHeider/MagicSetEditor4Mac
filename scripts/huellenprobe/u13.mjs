// **U13 am nativen Fenster: Weg und Arbeitsplatz zeigen dasselbe Set** —
// #2158.
//
// Das Fenster liest das offene Set einmal und reicht es an beide. Gemessen:
// der Weg über dem Kopf nennt den Namen, den der Arbeitsplatz zeigt; ein
// **fremder** Schreiber (`mse`, ein anderer Prozess) legt eine Karte an, und
// der Kopf zählt sie, ohne dass der Name dabei verschwindet; ⌘Z nimmt sie
// zurück, und der Kopf folgt auch dem Verlauf.
//
// **Was sie nicht belegt**: die Zahl der Rufe — die zählt
// `frontend/src/haken/offenesset.test.tsx`; das Fenster gibt sie nicht her.
// Und nur 1280×800, hell, Deutsch: der Schnitt ändert an Aufbau, Farben und
// Texten nichts, nur daran, woher Kopf und Weg ihr Set haben.
import { oeffnen, zurSeite } from './uirevue.mjs'

const kopf = "(document.querySelector('main h1')?.innerText ?? '')"
const zeile = "(document.querySelector('main h1')?.nextElementSibling?.innerText ?? '')"
const weg = "[...document.querySelectorAll('nav ol li')].map((l) => l.innerText.replace(/^›\\s*/, '').trim()).filter(Boolean)"

export default async function u13(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  await oeffnen(p, ort, { name: 'normal', breite: 1280, hoehe: 800 })
  await zurSeite(p, 'Karten')
  await p.warte('der Kopf nennt das Set', () => p.seite(`return ${kopf}.length > 0 && ${zeile}.includes('5')`), 15000)

  const name = await p.seite(`return ${kopf}`)
  const krumen = await p.seite(`return ${weg}`)
  p.zeige(`Kopf »${name}«, Weg ${JSON.stringify(krumen)}`)
  p.pruefe('der Weg nennt den Namen des Kopfs', krumen.includes(name), JSON.stringify(krumen))

  p.notiz('Ein fremder Schreiber legt eine Karte an')
  p.mse(ort, 'karte', '1', 'anlegen')
  await p.warte('der Kopf zählt sechs Karten', () => p.seite(`return ${zeile}.includes('6')`), 15000)
  p.pruefe('der Name bleibt im Kopf', (await p.seite(`return ${kopf}`)) === name)
  p.pruefe('und im Weg', (await p.seite(`return ${weg}`)).includes(name))
  await p.bild('u13-fremd')

  p.notiz('⌘Z nimmt sie zurück — der Kopf folgt dem Verlauf')
  await p.taste('z', 'befehl')
  await p.warte('der Kopf zählt wieder fünf Karten', () => p.seite(`return ${zeile}.includes('5')`), 15000)
  p.pruefe('der Weg nennt weiter den Namen', (await p.seite(`return ${weg}`)).includes(name))
  await p.bild('u13-zurueck')

  p.pruefe('⌘Q beendet', await p.beende())
}
