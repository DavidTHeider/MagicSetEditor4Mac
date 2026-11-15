// **Ein Fehlersystem, gemessen am nativen Fenster** — AP-2
// (#2359),
// Befunde B-07 und B-09.
//
// Bis AP-2 trugen Fehlermeldungen zwei Farben: `text-warnung` aus der
// Gestaltungssprache und Tailwinds `text-red-800 dark:text-red-300`. Seither
// gibt es `text-fehler`, und es kommt aus demselben Merkmal `--mse-warnung`,
// das die Farbwelten setzen und die Sperre misst.
//
// # Was es prüft, in jeder Farbwelt hell und dunkel
//
// Pergament, Klar, Schiefer und Moos (`themen.css`) — acht Starts. Die
// Farbwelten setzen `--mse-warnung` und die Flächen; die neuen Merkmale
// folgen nur dem Erscheinungsbild und müssen deshalb auf **jedem** dieser
// Gründe halten, nicht nur auf Pergament.
//
// - **Die echte Meldung**: in der Konsole ein Ausdruck, den der Zerteiler
//   abweist. Ihr `role="alert"` trägt die Farbe von `--mse-warnung`, nicht
//   eine zweite, und hält 4,5 : 1 auf ihrem Grund.
// - **Die Farben des Quelltexts** im selben Feld — Zahl und Zeichenkette.
// - **Jedes neue Merkmal auf jedem Grund**: Probeelemente mit genau den
//   Klassen aus dem Quelltext auf `bg-grund`, `bg-flaeche` und `bg-betont`,
//   gemessen mit `huellenkontrast.mjs` — Text 4,5 : 1, Rand und Strich
//   3 : 1. **Eine Klasse, für die Tailwind keine Regel erzeugt hat**, erbte
//   die Textfarbe und fiele hier als dieselbe Farbe wie `--mse-text` auf.
//
// # Was es nicht belegt
//
// Ob die Farben zueinander passen — das sieht ein Mensch an den Bildern.
// Und eigene Farben: für sie gilt die Sperre (`kontrast.ts`), bevor
// gezeichnet wird.
import { kontraste, pruefeKontraste } from '../huellenkontrast.mjs'

const knopf = (text) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(text)})`
// **Die Konsole liegt in der Werkzeug-Schublade** und nicht in `main` — der
// erste Lauf suchte dort und fand die Meldung nicht, die dastand.
const meldung = "[...document.querySelectorAll('[role=dialog] [role=alert]')].find((e) => e.checkVisibility())"

const texte = ['fehler', 'hinweis', 'erfolg', 'zahl', 'zeichenkette']
const raender = ['fehler', 'erfolg', 'auswahlkante', 'markierungskante']
const striche = ['auswahlkante', 'markierungskante', 'leise']
const gruende = ['grund', 'flaeche', 'betont']

// probenAuslegen legt je Grund eine Zeile mit jeder Klasse aus — oben über
// der Seite, damit kein anderer Grund dazwischen liegt.
const probenAuslegen = `
  document.getElementById('farbsprache09')?.remove()
  const kasten = document.createElement('div')
  kasten.id = 'farbsprache09'
  kasten.className = 'fixed z-50 flex flex-col gap-1 rounded-md border border-kante bg-grund p-2'
  // **Als Stil und nicht als Klasse**: scripts/ liest Tailwind nicht, und
  // eine Lage, die nirgends sonst steht, bekäme keine Regel.
  kasten.style.left = '30%'
  kasten.style.top = '30%'
  for (const grund of ${JSON.stringify(gruende)}) {
    const zeile = document.createElement('div')
    zeile.className = 'flex items-center gap-2 p-1 bg-' + grund
    zeile.dataset.grund = grund
    for (const t of ${JSON.stringify(texte)}) {
      const s = document.createElement('span')
      s.className = 'text-sm text-' + t
      s.dataset.probe = 'text-' + t
      s.textContent = t
      zeile.appendChild(s)
    }
    for (const r of ${JSON.stringify(raender)}) {
      const s = document.createElement('span')
      s.className = 'inline-block size-4 border-2 border-' + r
      s.dataset.probe = 'border-' + r
      zeile.appendChild(s)
    }
    const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg')
    svg.setAttribute('width', '60')
    svg.setAttribute('height', '16')
    const striche = ${JSON.stringify(striche)}
    striche.forEach((st, i) => {
      const linie = document.createElementNS('http://www.w3.org/2000/svg', 'line')
      linie.setAttribute('x1', String(4 + i * 20))
      linie.setAttribute('x2', String(16 + i * 20))
      linie.setAttribute('y1', '8')
      linie.setAttribute('y2', '8')
      linie.setAttribute('stroke-width', '2')
      linie.setAttribute('class', 'stroke-' + st)
      linie.dataset.probe = 'stroke-' + st
      svg.appendChild(linie)
    })
    zeile.appendChild(svg)
    kasten.appendChild(zeile)
  }
  document.body.appendChild(kasten)
  return true`

const probe = (grund, klasse) => `document.querySelector('#farbsprache09 [data-grund=${grund}] [data-probe=${klasse}]')`

export default async function farbsprache09(p) {
  for (const welt of ['pergament', 'klar', 'schiefer', 'moos']) for (const bild of ['hell', 'dunkel']) {
    const wahl = `${welt} ${bild}`
    const ort = p.ort()
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
    p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
    p.mse(ort, 'einstellungen', 'setzen', 'farbwelt', welt)
    p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
    await p.starte({ ort })
    await p.groesse(1440, 900)
    p.notiz(`Erscheinungsbild ${wahl}`)
    const gesetzt = await p.seite('return (document.documentElement.dataset.erscheinungsbild ?? "") + " " + (document.documentElement.dataset.farbwelt ?? "pergament")')
    p.pruefe(`${wahl}: Erscheinungsbild und Farbwelt sind gesetzt`, gesetzt === `${bild} ${welt}`, gesetzt)

    p.notiz('Set öffnen → ⌘K »Konsole« → ein Ausdruck, den der Zerteiler abweist')
    await p.taste('2', 'befehl')
    await p.wartetext('Werkstatt-Testkarten', 20000)
    await p.klickeText('Werkstatt-Testkarten')
    await p.warte('das Set ist offen', () => p.seite("return (document.querySelector('nav ol')?.innerText ?? '').includes('Werkstatt')"), 20000)
    await p.taste('k', 'befehl')
    await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
    await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
    await p.tippeInsFeld('Konsole')
    await p.klickeAuf("[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Konsole')", '»Konsole« in der Palette')
    const feld = "document.querySelector('textarea[aria-label=\"Ausdruck\"]')"
    await p.warte('das Feld »Ausdruck« steht', () => p.seite(`return !!${feld}?.checkVisibility()`), 20000)
    await p.klickeAuf(feld, 'das Feld »Ausdruck«')
    await p.seite(`${feld}.focus(); return true`)
    await p.tippe('"karte" + 12 +')
    await p.klickeAuf(knopf('Auswerten'), '»Auswerten«')
    const satz = await p.warte('die Meldung steht', () => p.seite(`return (${meldung})?.innerText ?? false`), 15000)
    p.zeige(`Meldung: ${satz.replace(/\s+/g, ' ')}`)

    // **Eine Farbe und nicht zwei**: die Meldung trägt, was `--mse-warnung`
    // ergibt — aufgelöst über ein Element mit genau dieser Farbe.
    const farben = await p.seite(`const m = ${meldung}; const s = document.createElement('span'); s.style.color = 'var(--mse-warnung)'; m.parentElement.appendChild(s); const soll = getComputedStyle(s).color; s.remove(); return { ist: getComputedStyle(m).color, soll }`)
    p.pruefe(`${wahl}: die Meldung trägt --mse-warnung`, farben.ist === farben.soll, `${farben.ist} gegen ${farben.soll}`)
    await pruefeKontraste(p, wahl, [
      ['Meldung der Konsole', meldung, 'text', 4.5],
      ['Zahl im Ausdruck', "[...document.querySelectorAll('[role=dialog] .text-zahl')].find((e) => e.checkVisibility())", 'text', 4.5],
      ['Zeichenkette im Ausdruck', "[...document.querySelectorAll('[role=dialog] .text-zeichenkette')].find((e) => e.checkVisibility())", 'text', 4.5],
    ])
    await p.bild(`konsole-fehler-${welt}-${bild}`)

    p.notiz('Jedes neue Merkmal auf Grund, Fläche und betonter Fläche')
    await p.seite(probenAuslegen)
    const liste = gruende.flatMap((grund) => [
      ...texte.map((t) => [`text-${t} auf bg-${grund}`, probe(grund, `text-${t}`), 'text', 4.5]),
      ...raender.map((r) => [`border-${r} auf bg-${grund}`, probe(grund, `border-${r}`), 'rand', 3]),
      ...striche.map((st) => [`stroke-${st} auf bg-${grund}`, probe(grund, `stroke-${st}`), 'symbol', 3]),
    ])
    await pruefeKontraste(p, wahl, liste)
    // **Für den Beleg die Zahlen selbst** — `docs/KONTRASTE.md` nennt sie.
    for (const w of await kontraste(p, liste)) p.zeige(`${wahl}: ${w.name} ${w.fehlt ? 'fehlt' : String(w.verhaeltnis).replace('.', ',')}`)
    // **Keine tote Klasse**: eine Farbe ohne Regel erbte `--mse-text`.
    const geerbt = await p.seite(`const text = getComputedStyle(document.querySelector('#farbsprache09')).color; return [...document.querySelectorAll('#farbsprache09 [data-probe^=text-]')].filter((e) => getComputedStyle(e).color === text).map((e) => e.dataset.probe)`)
    p.pruefe(`${wahl}: jede Klasse hat ihre Farbe und erbt nicht die Textfarbe`, geerbt.length === 0, geerbt.join(', '))
    await p.bild(`farbproben-${welt}-${bild}`)
    await p.seite("document.getElementById('farbsprache09')?.remove(); return true")
    p.pruefe(`${wahl}: ⌘Q beendet`, await p.beende())
  }
}
