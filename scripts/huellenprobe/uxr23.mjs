// **Jede Seite mit einem frischen Bestand** — UXR-23
// (#1931).
//
// §6 der Leitlinien: *Leere Zustände erklären den nächsten Schritt, statt
// „Keine Daten" zu zeigen.* Ein frischer Bestand ist der Anfang, den jeder
// Mensch einmal sieht — kein Set, kein Deck, keine Identität. Dieses Drehbuch
// geht jede Seite der Palette damit durch und zählt im Hauptbereich:
//
// - **Fehler** — sichtbare `role=alert`. Am Anfang ist nichts schiefgegangen;
//   ein Fehler dort ist ein Befund (so stand es auf *Handeln*, #2027).
// - **Leerzustände** (`data-leerzustand`) und ob sie einen nächsten Schritt
//   tragen — einen Knopf oder Verweis.
// - **„Erneut laden"** — ein Knopf, der am Anfang nichts ändern kann.
//
// Je Seite ein Bild. **Was es nicht tut**: urteilen, ob ein Satz gut ist.
// Das sieht ein Mensch an den Bildern.
import { seitennamen, zurSeite } from './uirevue.mjs'

const zustand = `
  const main = document.querySelector('main')
  if (!main) return null
  const sichtbar = (e) => e.checkVisibility()
  const alarme = [...main.querySelectorAll('[role=alert]')].filter(sichtbar).map((e) => e.innerText.trim().replace(/\\s+/g, ' ').slice(0, 140)).filter((t) => t !== '')
  const leer = [...main.querySelectorAll('[data-leerzustand]')].filter(sichtbar)
  const erneut = [...main.querySelectorAll('button')].filter((b) => sichtbar(b) && /Erneut laden/.test(b.innerText)).length
  return {
    alarme,
    leer: leer.length,
    leerMitSchritt: leer.filter((l) => l.querySelector('button, a')).length,
    erneut,
  }
`

export default async function uxr23(p) {
  // **Ein Ort ohne Set, ohne Deck, ohne Identität** — nur die Sprache, damit
  // die Seitennamen die der Palette sind.
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', '1280')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', '800')
  await p.starte({ ort })

  const namen = await seitennamen(p)
  p.zeige(`Die Palette nennt ${String(namen.length)} Seiten: ${namen.join(' · ')}`)

  const zeilen = []
  for (const name of namen) {
    try {
      await zurSeite(p, name)
      const wert = await p.seite(zustand)
      if (wert) zeilen.push({ name, ...wert })
      await p.bild(`frisch-${name.toLowerCase().replace(/[^a-zäöüß]+/g, '-')}`)
    } catch (grund) {
      p.zeige(`${name}: nicht erreicht — ${String(grund.message ?? grund)}`)
      await p.taste('Escape')
    }
  }
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())

  p.notiz('Die Tabelle')
  p.zeige('| Seite | Fehler | Leerzustände (mit nächstem Schritt) | „Erneut laden" |')
  p.zeige('| --- | --- | ---: | ---: |')
  for (const z of zeilen) {
    p.zeige(`| ${z.name} | ${z.alarme.length === 0 ? '—' : z.alarme.join(' / ')} | ${String(z.leer)} (${String(z.leerMitSchritt)}) | ${String(z.erneut)} |`)
  }

  p.pruefe('jede Seite gemessen', zeilen.length === namen.length, `${String(zeilen.length)} von ${String(namen.length)}`)
  const mitFehler = zeilen.filter((z) => z.alarme.length > 0).map((z) => z.name)
  p.pruefe('am Anfang zeigt keine Seite einen Fehler', mitFehler.length === 0, mitFehler.join(', '))
  const ohneSchritt = zeilen.filter((z) => z.leer > z.leerMitSchritt).map((z) => z.name)
  p.zeige(`Leerzustände ohne nächsten Schritt: ${ohneSchritt.join(', ') || '—'}`)
  const erneut = zeilen.filter((z) => z.erneut > 0).map((z) => z.name)
  p.pruefe('am Anfang steht kein „Erneut laden"', erneut.length === 0, erneut.join(', '))
}
