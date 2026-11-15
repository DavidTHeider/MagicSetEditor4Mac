// **Netzdiagnose am nativen Fenster** — D02
// (#2233).
//
// Bis dahin zeigte die Kartensuche für jeden Ausfall von Scryfall „Es ist ein
// unerwarteter Fehler aufgetreten". Jetzt sagt der Satz, welcher Fall es war,
// welcher Dienst, welcher Rat — und unter welcher Kennung das Protokoll
// steht. Drei Fälle über denselben Weg wie `sec17bilder` (*Karten › Mehr ›
// Aus Scryfall anlegen …*):
//
//   1. ein Name, den es nicht gibt (`.invalid`, RFC 6761) — der Resolver
//      antwortet sofort,
//   2. eine Gegenstelle auf 127.0.0.1, die mit `429` antwortet,
//   3. dieselbe Gegenstelle, die den Kopf nicht schickt — die **übliche
//      Kopffrist von 30 s** greift, gemessen am Fenster.
//
// **Was es nicht belegt**: einen Namensdienst, der hängt, und einen echten
// Proxy. Die Klassen selbst hält `internal/netz/netzklassen_test.go`.
import { createServer } from 'node:http'

import { zurSeite } from './uirevue.mjs'

// kennung ist die Gestalt einer Vorgangskennung im Satz.
const kennung = /\b[0-9a-f]{8}\b/

// suche ersetzt den Suchbegriff und sucht.
async function suche(p, begriff) {
  const feld = "[...document.querySelectorAll('input')].find((e) => e.checkVisibility() && e.labels?.[0]?.innerText.startsWith('Bei Scryfall suchen'))"
  await p.klickeAuf(feld, 'das Suchfeld')
  // **Erst der Fokus, dann die Tasten** — und **kein ⌘A davor**: in zwei
  // nativen Läufen (2026-10-10) ging die erste Taste nach dem Kürzel
  // verloren, aus »Probe« wurde »robe«. Geleert wird über den Wertsetzer,
  // den React liest; getippt wird dann mit echten Tasten.
  await p.warte('das Suchfeld hat den Fokus', () => p.seite(`return document.activeElement === (${feld})`), 10000)
  await p.seite(`const f = ${feld}; Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(f, ''); f.dispatchEvent(new Event('input', { bubbles: true })); return f.value === ''`)
  await p.tippeInsFeld(begriff)
  await p.klickeText('Suchen')
}

// zurSuche führt von der Setliste zur Scryfallsuche.
async function zurSuche(p) {
  await p.taste('2', 'befehl')
  await p.warte('die Sets stehen', () => p.seite("return !!document.querySelector('main article')"), 20000)
  await p.klickeAuf("[...document.querySelectorAll('main article :is(h2, h3) button, main article :is(h2, h3) a')].find((b) => b.checkVisibility())", 'das erste Set')
  await zurSeite(p, 'Karten')
  await p.klickeAuf("[...document.querySelectorAll('main [aria-haspopup=menu]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Mehr')", 'das Menü »Mehr«')
  await p.klickeText('Aus Scryfall anlegen …')
}

// meldung wartet auf die Meldung am Suchfeld und prüft sie.
async function meldung(p, fall, { enthaelt, frist = 20000 }) {
  const satz = await p.warte(`die Meldung für ${fall}`, async () => {
    const text = await p.seite("return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText).join(' ')")
    return text.includes(enthaelt) ? text : null
  }, frist, 250)
  p.zeige(satz)
  p.pruefe(`${fall}: kein »unerwarteter Fehler«`, !satz.includes('unerwarteter Fehler'))
  p.pruefe(`${fall}: der Satz nennt den Vorgang`, kennung.test(satz), satz.match(kennung)?.[0] ?? '')
  return satz
}

export default async function d02(p) {
  let modus = 'drossel'
  const offen = new Set()
  const gegenstelle = createServer((anfrage, antwort) => {
    p.zeige(`die Gegenstelle bekam ${anfrage.url.split('?')[0]} im Modus ${modus}`)
    if (modus === 'drossel') {
      antwort.writeHead(429, { 'retry-after': '1', 'content-type': 'application/json' })
      antwort.end('{"object":"error","status":429}')
      return
    }
    // **Stumm**: kein Kopf, bis die Probe endet.
    offen.add(antwort)
  })
  await new Promise((fertig) => { gegenstelle.listen(0, '127.0.0.1', fertig) })
  const adresse = `http://127.0.0.1:${String(gegenstelle.address().port)}`

  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  try {
    p.notiz('1 — ein Name, den es nicht gibt')
    await p.starte({ ort, umgebung: { MSE_SCRYFALL_ADRESSE: 'http://gibtsnicht.invalid' } })
    await zurSuche(p)
    await suche(p, 'Probe')
    const name = await meldung(p, 'der Name', { enthaelt: 'ließ sich nicht auflösen' })
    p.pruefe('der Name nennt den Dienst', name.includes('gibtsnicht.invalid'))
    await p.bild('d02-name')
    p.pruefe('⌘Q beendet', await p.beende())

    p.notiz('2 — eine Gegenstelle, die mit 429 antwortet')
    await p.starte({ ort, umgebung: { MSE_SCRYFALL_ADRESSE: adresse } })
    await zurSuche(p)
    await suche(p, 'Probe')
    await meldung(p, 'das 429', { enthaelt: 'zu viele Anfragen' })
    await p.bild('d02-gedrosselt')

    p.notiz('3 — dieselbe Gegenstelle schweigt; die Kopffrist von 30 s greift')
    modus = 'stumm'
    await new Promise((weiter) => { setTimeout(weiter, 1500) })
    const beginn = Date.now()
    await suche(p, 'Probe')
    const stumm = await meldung(p, 'die Kopffrist', { enthaelt: 'nicht binnen 30 Sekunden', frist: 60000 })
    const dauer = (Date.now() - beginn) / 1000
    p.pruefe('die Frist griff nach rund 30 s', dauer >= 29 && dauer <= 40, `${dauer.toFixed(1)} s`)
    p.pruefe('die Kopffrist nennt den Dienst', stumm.includes('127.0.0.1'))
    await p.bild('d02-kopffrist')
    p.pruefe('⌘Q beendet', await p.beende())
  } finally {
    for (const antwort of offen) antwort.destroy()
    gegenstelle.close()
  }
}
