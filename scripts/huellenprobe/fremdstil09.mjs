// **Ein fremder Setstilwechsel bei offenem Editor, am nativen Fenster** —
// #2314 (E2314).
//
// Das Probeset trägt `probe-kunst`; seine Karten erben den Stil. Das Fenster
// zeigt eine Karte im Editor, der Fokus steht im Namensfeld. Dann wechselt
// die Kommandozeile — ein zweiter Prozess — den Setstil auf `probe-schlicht`.
// **Erwartet**: die Vorschau zeichnet mit dem neuen Paket, und das Formular
// ordnet nach dessen Lage; Werte, Knoten und Fokus bleiben.
//
// # Woran die Ordnung zu erkennen ist
//
// `probe-kunst` legt das Bildfeld `kunst` über den Namen (oben 60 gegen
// 300); `probe-schlicht` kennt `kunst` gar nicht und zeichnet den Namen
// zuoberst (24). Vor dem Wechsel steht `kunst` vor `name`, danach nicht mehr.
//
// # Was es nicht belegt
//
// Das Budget aus E02 misst dieser Lauf nicht; der Plan nennt sein Paket aus
// dem Stil, den er ohnehin in der Hand hat — eine Abfrage mehr gibt es nicht.
import { execFileSync } from 'node:child_process'
import { resolve } from 'node:path'

import { kindumgebung } from '../huellenorte.mjs'
import { oeffnen, zurSeite } from './uirevue.mjs'

// **Das `mse` aus dem Bündel**: es findet die Saat neben sich
// (`Contents/Resources/saat`), das aus `.go/bin` nicht — und es ist das
// Programm, das jemand neben dem Fenster tatsächlich laufen hätte.
const buendelmse = resolve(import.meta.dirname, '..', '..', 'build/bin/MagicSetEditor4Mac.app/Contents/MacOS/mse')
const ausDemBuendel = (ort, ...argumente) => execFileSync(buendelmse, argumente, { env: kindumgebung(ort) }).toString()

const formular = "document.querySelector('input[aria-label=\"name\"]')?.closest('form')"
const ordnung = `return [...(${formular}?.querySelectorAll('[aria-label]') ?? [])].map((e) => e.getAttribute('aria-label')).filter((n) => ['kunst', 'name', 'text'].includes(n))`
const name = "document.querySelector('input[aria-label=\"name\"]')"

export default async function fremdstil09(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.pruefe('das Set trägt probe-kunst', /probe-kunst/.test(p.mse(ort, 'sets')))

  await oeffnen(p, ort, { name: 'normal', breite: 1280, hoehe: 800 })
  await zurSeite(p, 'Karten')
  await p.warte('das Namensfeld steht da', () => p.seite(`return !!${name}`), 20000)

  const vorher = await p.warte('die Ordnung des alten Pakets steht', async () => {
    const reihe = await p.seite(ordnung)
    return reihe.includes('kunst') && reihe.indexOf('kunst') < reihe.indexOf('name') ? reihe : false
  }, 20000)
  p.zeige(`vorher: ${vorher.join(' · ')}`)
  await p.bild('e2314-vorher')

  const wert = await p.seite(`const e = ${name}; e.__probe = 'e2314'; e.focus(); return e.value`)
  p.pruefe('der Fokus steht im Namen', await p.seite(`return document.activeElement === ${name}`))

  p.notiz('Der fremde Wechsel: die Kommandozeile setzt probe-schlicht')
  p.zeige(ausDemBuendel(ort, 'stile', '1', 'wechseln', 'probe-schlicht').trim().split('\n')[0])

  const nachher = await p.warte('das Formular ordnet nach dem neuen Paket', async () => {
    const reihe = await p.seite(ordnung)
    return !reihe.includes('kunst') || reihe.indexOf('kunst') > reihe.indexOf('name') ? reihe : false
  }, 20000)
  p.zeige(`nachher: ${nachher.join(' · ')}`)
  await p.bild('e2314-nachher')

  const lage = await p.seite(`const e = ${name}; return { selbe: e?.__probe === 'e2314', wert: e?.value, fokus: document.activeElement === e }`)
  p.pruefe('derselbe Knoten', lage.selbe)
  p.pruefe('derselbe Wert', lage.wert === wert, lage.wert)
  p.pruefe('der Fokus bleibt im Namen', lage.fokus)
  p.pruefe('das Set trägt jetzt probe-schlicht', /probe-schlicht/.test(ausDemBuendel(ort, 'sets')))
  p.pruefe('⌘Q beendet', await p.beende())
}
