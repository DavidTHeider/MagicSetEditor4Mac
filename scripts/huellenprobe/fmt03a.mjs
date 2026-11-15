// **Vor dem Einlesen sehen, was kommt — am nativen Fenster** — FMT-03.AK-01
// (#1692):
//
// - *Set öffnen …* fragt im Dialog von macOS; ein Abbruch tut nichts,
// - fehlt ein Paket, steht die Vorschau da und nennt es — *Nicht einlesen*
//   lässt den Bestand, wie er war, *Einlesen* liest danach ein,
// - eine saubere Datei geht über ⌘O ohne Vorschau gleich auf.
//
// **Ein eigenes Zuhause**: gelesen wird aus dem Bestand dieser Probe.
import { mkdirSync, mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith(${JSON.stringify(name)}))`
const vorschau = (name) => `document.querySelector('section[aria-label=${JSON.stringify(`Vor dem Einlesen: ${name}`)}]')`

async function zurSeite(p, ziel) {
  await p.taste('k', 'befehl')
  await p.warte('die Palette steht', () => p.seite("return !!document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]')"))
  await p.seite("document.querySelector('[role=dialog] input[aria-label=\"Wohin?\"]').focus(); return true")
  await p.tippeInsFeld(ziel)
  await p.klickeAuf(`[...document.querySelectorAll('[data-palettenzeile]')].find((b) => b.checkVisibility() && b.innerText.trim() === ${JSON.stringify(ziel)})`, `»${ziel}« in der Palette`)
}

function setdatei(ordner, name, inhalt) {
  const pfad = join(ordner, `${name}.mse-set`)
  mkdirSync(pfad)
  writeFileSync(join(pfad, 'set'), inhalt)

  return pfad
}

const setnamen = (p, ort) => JSON.parse(p.mse(ort, 'sets', '--json')).sets.map((set) => set.name)

export default async function fmt03a(p) {
  const ort = p.ort({ zuhause: true })
  const ordner = mkdtempSync(join(tmpdir(), 'fmt03a-'))
  const fremd = setdatei(ordner, 'fremd', 'mse version: 2.0.0\ngame: fremd\nset info:\n\ttitle: Fremdes Set\ncard:\n\tname: Eins\ncard:\n\tname: Zwei\n')
  const sauber = setdatei(ordner, 'sauber', 'mse version: 2.0.0\ngame: probe\nstylesheet: schlicht\nset info:\n\ttitle: Sauberes Set\ncard:\n\tname: Eins\n')

  await p.starte({ ort })

  // **Die Kommandozeile sagt vorher, was das Fenster zeigen soll** — sonst
  // prüfte die Probe eine Annahme über die Saat.
  const vorab = JSON.parse(p.mse(ort, '--json', 'set-vorschau', sauber))
  p.pruefe('an der sauberen Datei findet die Vorschau nichts', vorab.bedarf.length === 0 && vorab.befunde.length === 0,
    JSON.stringify(vorab.bedarf))

  await zurSeite(p, 'Sets')
  await p.klickeAuf(knopf('Set öffnen …'), 'Set öffnen …')
  await p.dialogZeigtSich('oeffnen', '»Set öffnen …«')

  p.notiz('Eine Datei, deren Spiel fehlt')
  await p.dialogantwort(fremd)
  await p.klickeAuf(knopf('Set öffnen …'), 'Set öffnen …')
  await p.warte('die Vorschau steht', () => p.seite(`return !!${vorschau('Fremdes Set')}`), 15000)
  p.pruefe('sie nennt Kartenzahl und das fehlende Spiel',
    await p.seite(`const e = ${vorschau('Fremdes Set')}; return e.innerText.includes('2 Karten') && e.innerText.includes('fremd.mse-game')`))
  p.pruefe('eingelesen ist noch nichts', setnamen(p, ort).length === 0, setnamen(p, ort).join(', '))
  await p.bild('vorschau')

  await p.klickeAuf(knopf('Nicht einlesen'), 'Nicht einlesen')
  await p.warte('die Vorschau ist fort', () => p.seite(`return !${vorschau('Fremdes Set')}`))
  p.pruefe('Nicht einlesen lässt den Bestand, wie er war', setnamen(p, ort).length === 0)

  await p.dialogantwort(fremd)
  await p.klickeAuf(knopf('Set öffnen …'), 'Set öffnen …')
  await p.warte('die Vorschau steht wieder', () => p.seite(`return !!${vorschau('Fremdes Set')}`), 15000)
  await p.klickeAuf(knopf('Einlesen'), 'Einlesen')
  await p.warte('das Set ist eingelesen', async () => setnamen(p, ort).includes('Fremdes Set'), 15000)
  // **Auf die sichtbare Wirkung gewartet** (§3): die Liste nennt es, und der
  // Knopf ist wieder frei.
  await p.warte('die Liste nennt es', () => p.seite(`return document.body.innerText.includes('Fremdes Set') && !${knopf('Set öffnen …')}.disabled`), 15000)
  p.pruefe('Einlesen liest ein', true)
  await p.bild('eingelesen')

  p.notiz('Eine saubere Datei über ⌘O')
  await zurSeite(p, 'Sets')
  await p.dialogantwort(sauber)
  await p.taste('o', 'befehl')
  await p.warte('das saubere Set ist eingelesen', async () => setnamen(p, ort).includes('Sauberes Set'), 15000)
  await p.warte('die Liste nennt es, und der Knopf ist frei', () => p.seite(`return document.body.innerText.includes('Sauberes Set') && document.body.innerText.includes('1 Karte geöffnet') && !${knopf('Set öffnen …')}.disabled`), 15000)
  p.pruefe('ohne Vorschau dazwischen', await p.seite(`return !${vorschau('Sauberes Set')}`))
  await p.bild('sauber')

  await p.beende()
}
