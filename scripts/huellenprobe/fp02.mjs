// **FP-02 am nativen Fenster: das lange Textfeld der Saat ist mehrzeilig** —
// #2156.
//
// Die Werkstattkarte »Sonderlayout« trägt ihren Text **unverändert aus der
// Saat** — mit Leerzeile und vier Zeilen (AK-01). Gemessen wird, ob das Feld
// `text` als mehrzeilige Eingabe dasteht und mehr als eine Zeile zeigt, hell
// und dunkel bei 1280×800 (AK-03); dann wird **mit der Tastatur** eine Zeile
// angefügt, die Anwendung beendet und neu geöffnet, und der Text steht samt
// Absätzen wieder da (AK-02). Feldnamen sind die der Saat: `name`, `text`
// (AK-04).
//
// **Was sie nicht belegt**: 1024×700 und Englisch — das Feld ändert dort
// nichts an Aufbau oder Schlüssel; und `⌘S` in eine Datei, denn das Set kam
// aus der Saat, und der Weg schriebe in den Baum.
import { oeffnen, zurSeite } from './uirevue.mjs'

const textfeld = "document.querySelector('textarea[aria-label=\"text\"]')"
const karte = "[...document.querySelectorAll('main button')].find((b) => b.checkVisibility() && b.innerText.startsWith('Sonderlayout'))"

async function zurKarte(p) {
  await zurSeite(p, 'Karten')
  // **Die Kartenliste steht im Navigator**, und der ist anfangs eingeklappt.
  if (await p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Navigator einblenden')")) {
    await p.klickeText('Navigator einblenden')
  }
  await p.klickeAuf(karte, 'die Karte »Sonderlayout«')
  await p.warte('das Textfeld ist eine mehrzeilige Eingabe', () => p.seite(`return !!${textfeld}`), 15000)
}

async function vermesse(p, was) {
  const lage = await p.seite(`const e = ${textfeld}; const r = e.getBoundingClientRect()
    const zeile = parseFloat(getComputedStyle(e).lineHeight) || parseFloat(getComputedStyle(e).fontSize) * 1.4
    return { wert: e.value, hoehe: r.height, zeile }`)
  p.zeige(`${was}: ${String(Math.round(lage.hoehe))} px hoch bei ${String(Math.round(lage.zeile))} px je Zeile`)
  p.pruefe(`${was}: der Text trägt seine Leerzeile`, lage.wert.includes('Erste Zeile.\n\nDritte Zeile'), JSON.stringify(lage.wert.slice(0, 40)))
  p.pruefe(`${was}: mehr als eine Zeile ist zu sehen`, lage.hoehe > 1.8 * lage.zeile)

  return lage
}

export default async function fp02(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  for (const bild of ['hell', 'dunkel']) {
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
    await oeffnen(p, ort, { name: bild, breite: 1280, hoehe: 800 })
    await zurKarte(p)
    await vermesse(p, bild)
    await p.bild(`fp02-${bild}`)

    if (bild === 'dunkel') {
      p.pruefe('⌘Q beendet', await p.beende())
      break
    }

    p.notiz('Mit der Tastatur: ans Ende, Zeilenumbruch, eine Zeile tippen, das Feld verlassen')
    await p.seite(`const e = ${textfeld}; e.focus(); e.setSelectionRange(e.value.length, e.value.length); return true`)
    await p.taste('Enter')
    await p.tippeInsFeld('Fünfte Zeile mit der Tastatur.')
    await p.taste('Tab')
    await p.warte('die neue Zeile steht im Feld', () => p.seite(`return ${textfeld}.value.endsWith('\\nFünfte Zeile mit der Tastatur.')`), 10000)
    await new Promise((los) => setTimeout(los, 1500))
    p.pruefe('⌘Q beendet', await p.beende())
  }

  p.notiz('Wieder geöffnet: der Text steht samt Absätzen und angefügter Zeile da')
  await oeffnen(p, ort, { name: 'wieder', breite: 1280, hoehe: 800 })
  await zurKarte(p)
  const wieder = await vermesse(p, 'wieder geöffnet')
  p.pruefe('die getippte Zeile hat überdauert', wieder.wert.endsWith('\nFünfte Zeile mit der Tastatur.'), JSON.stringify(wieder.wert.slice(-40)))
  p.pruefe('⌘Q beendet', await p.beende())
}
