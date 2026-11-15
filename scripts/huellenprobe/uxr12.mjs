// **Turniere mit einem Abend, wie ihn ein Mensch hat** — UXR-12
// (#1920).
//
// `uirevue` misst mit dem Probebestand, und der hat kein Turnier: *Turniere*
// zeigt dort die Liste leer. Dieses Drehbuch legt an der Kommandozeile acht
// Turniere an, das letzte mit sechzehn Teilnehmenden und vier Runden — drei
// gespielt, die vierte läuft — und misst die Seite mit **derselben** Messung
// wie `uirevue` bei 1280 × 800 und 1024 × 700: die Liste, dann je Lasche der
// Werkbank. Je Größe und Lasche ein Bild.
//
// **Was es nicht tut**: urteilen, ob die Anordnung dem Artboard *Turniere*
// gleicht. Das sieht ein Mensch an den Bildern.
import { messung, zurSeite } from './uirevue.mjs'

const groessen = [
  { name: 'normal', breite: 1280, hoehe: 800 },
  { name: 'eng', breite: 1024, hoehe: 700 },
]

const laschen = ['Runden', 'Rangliste', 'Teilnehmende', 'Decklisten', 'Protokoll']

const namen = ['Ada', 'Bo', 'Cleo', 'Dario', 'Emil', 'Frida', 'Gus', 'Hugo', 'Ida', 'Jon', 'Kim', 'Lea', 'Mo', 'Nia', 'Ole', 'Pia']

// **Das Zeichen des Katalogs steht vor dem Text** — gesucht wird deshalb nach
// dem Ende, wo es keine Bezeichnung gibt (wie in `dom04ak02`).
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') === ${JSON.stringify(name)} || (!b.hasAttribute('aria-label') && b.innerText.trim().endsWith(${JSON.stringify(name)}))))`
const lasche = (name) => `[...document.querySelectorAll('[role=tab]')].find((l) => l.checkVisibility() && l.innerText.trim() === ${JSON.stringify(name)})`

const zeile = (groesse, ziel, z) =>
  `| ${groesse} | ${ziel} | ${String(z.bedien)} (${String(z.bedienImBild)}) | ${String(z.haupt)} (${String(z.hauptJeRegion)}) | ${String(z.felder)} (${String(z.felderImBild)}) | ${String(z.ueberschriften)} | ${String(z.rolltiefe)} | ${z.waagerecht ? 'ja' : 'nein'} |`

export default async function uxr12(p) {
  const ort = p.ort()
  for (const abend of ['Montag', 'Dienstag', 'Mittwoch', 'Donnerstag', 'Samstag', 'Sonntag', 'Vorabend']) {
    const id = p.mse(ort, 'turnier', 'neu', abend, 'magic').trim()
    for (const name of namen.slice(0, 4)) p.mse(ort, 'turnier', 'anmelden', id, name)
  }
  const turnierID = p.mse(ort, 'turnier', 'neu', 'Freitag', 'magic').trim()
  for (const name of namen) p.mse(ort, 'turnier', 'anmelden', turnierID, name)
  for (let runde = 1; runde <= 4; runde++) {
    const gelegt = JSON.parse(p.mse(ort, 'turnier', 'runde', turnierID, '--json')).runde
    // **Die vierte läuft noch** — ihre Ergebnisse fehlen.
    if (runde < 4) for (const paarung of gelegt.paarungen) p.mse(ort, 'turnier', 'ergebnis', turnierID, paarung.id, '2', '1')
  }

  const zeilen = []
  for (const groesse of groessen) {
    p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
    p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', String(groesse.breite))
    p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', String(groesse.hoehe))
    await p.starte({ ort })

    await zurSeite(p, 'Turniere')
    await p.warte('das Turnier steht in der Liste', () => p.seite(`return !!${knopf('Öffnen: Freitag')}`), 20000)
    await new Promise((los) => setTimeout(los, 800))
    const liste = await p.seite(messung)
    zeilen.push(zeile(groesse.name, 'Turniere (acht in der Liste)', liste))
    p.pruefe(`R1: höchstens ein Hauptknopf je Region, Liste (${groesse.name})`, liste.hauptJeRegion <= 1, JSON.stringify(liste))
    await p.bild(`turniere-${groesse.name}`)

    await p.seite(`${knopf('Öffnen: Freitag')}.scrollIntoView({ block: 'center' }); return true`)
    await p.klickeAuf(knopf('Öffnen: Freitag'), 'Öffnen: Freitag')
    await p.warte('die Werkbank steht', () => p.seite(`return !!${lasche('Runden')}`), 20000)

    for (const name of laschen) {
      await p.klickeAuf(lasche(name), `die Lasche ${name}`)
      await p.warte(`die Lasche ${name} ist gewählt`, () => p.seite(`return ${lasche(name)}?.getAttribute('aria-selected') === 'true'`), 10000)
      await new Promise((los) => setTimeout(los, 600))
      const wert = await p.seite(messung)
      zeilen.push(zeile(groesse.name, `Turniere › Freitag › ${name}`, wert))
      p.pruefe(`R1 in ${name} (${groesse.name})`, wert.hauptJeRegion <= 1, JSON.stringify(wert))
      p.pruefe(`R5: nichts rollt waagerecht in ${name} (${groesse.name})`, !wert.waagerecht)
      await p.bild(`turniere-${name.toLowerCase()}-${groesse.name}`)
    }

    p.pruefe(`⌘Q beendet die Anwendung (${groesse.name})`, await p.beende())
  }

  p.notiz('Die Tabelle')
  p.zeige('| Größe | Ziel | Bedienelemente (im Bild) | Hauptknöpfe (je Region höchstens) | Felder (im Bild) | Überschriften | Rolltiefe | waagerecht |')
  p.zeige('| --- | --- | ---: | ---: | ---: | ---: | ---: | --- |')
  for (const z of zeilen) p.zeige(z)
}
