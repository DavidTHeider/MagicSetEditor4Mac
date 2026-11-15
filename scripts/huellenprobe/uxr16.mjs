// **Gegenüber mit einem Bestand, wie ihn ein Mensch hat** — UXR-16
// (#1924).
//
// Der Probebestand kennt kein Gegenüber: der Kasten „Gegenüber aufnehmen oder
// bestätigen" stünde leer da. Dieses Drehbuch legt acht Gegenüber an — jedes
// mit eigener Identität in einem eigenen Ort, eingetragen an der
// Kommandozeile, die Hälfte bestätigt — und misst mit **derselben** Messung
// wie `uirevue` bei 1280 × 800 und 1024 × 700:
//
// - *Handeln* › *Neuer Tausch …* mit aufgeklapptem Kasten,
// - *Einstellungen* › *Verbindungen* mit gezeigten Kontakten.
//
// **Was es nicht tut**: urteilen, ob die Anordnung dem Artboard *Gegenueber*
// gleicht. Das sieht ein Mensch an den Bildern.
import { messung, zurSeite } from './uirevue.mjs'

const groessen = [
  { name: 'normal', breite: 1280, hoehe: 800 },
  { name: 'eng', breite: 1024, hoehe: 700 },
]

const namen = ['Bert', 'Clara', 'Dario', 'Emil', 'Frida', 'Gus', 'Hugo', 'Ida']

const zeilenwert = (ausgabe, name) => ausgabe.split('\n').find((z) => z.startsWith(`${name}\t`))?.split('\t')[1] ?? ''
// **Die erste Zeile**: die Gruppenliste der Einstellungen trägt unter jedem
// Namen einen Untertitel.
const mitText = (text) => `[...document.querySelectorAll('button, [role=tab]')].find((b) => b.checkVisibility() && b.innerText.trim().split('\\n')[0] === ${JSON.stringify(text)})`

const imBild = `
  const main = document.querySelector('main')
  return [...main.querySelectorAll('button, a[href], input, select, textarea, [role=switch]')]
    .filter((e) => { if (!e.checkVisibility() || e.type === 'hidden') return false; const r = e.getBoundingClientRect(); return r.bottom > 0 && r.top < innerHeight })
    .map((e) => (e.getAttribute('aria-label') || e.innerText || e.getAttribute('name') || e.tagName).trim().replace(/\\s+/g, ' ').slice(0, 40))
`

// **Eine Zeile, gleich welche** — seit #2035 faltet die Liste nach fünf, und
// welches Gegenüber hinten steht, entscheidet die Reihenfolge des Bestands.
const eineZeile = "return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && /^(Vergleichen|Verwalten): /.test(b.getAttribute('aria-label') ?? ''))"

const zeile = (groesse, ziel, z) =>
  `| ${groesse} | ${ziel} | ${String(z.bedien)} (${String(z.bedienImBild)}) | ${String(z.haupt)} (${String(z.hauptJeRegion)}) | ${String(z.felder)} (${String(z.felderImBild)}) | ${String(z.ueberschriften)} | ${String(z.rolltiefe)} | ${z.waagerecht ? 'ja' : 'nein'} |`

export default async function uxr16(p) {
  const anna = p.ort()
  p.mse(anna, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(anna, 'peer-anlegen')
  for (const [stelle, name] of namen.entries()) {
    const fremd = p.ort()
    p.mse(fremd, 'peer-anlegen')
    const ich = p.mse(fremd, 'peer')
    const kennung = zeilenwert(ich, 'kennung')
    p.mse(anna, 'kontakt-aufnehmen', kennung, name)
    if (stelle % 2 === 0) p.mse(anna, 'kontakt-bestaetigen', kennung, zeilenwert(ich, 'fingerabdruck'))
  }
  const kontakte = p.mse(anna, 'kontakte')
  p.pruefe('acht Gegenüber, vier bestätigt', (kontakte.match(/^bestaetigt\t/gm) ?? []).length === 4, kontakte)

  const zeilen = []
  for (const groesse of groessen) {
    p.mse(anna, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
    p.mse(anna, 'einstellungen', 'setzen', 'fensterbreite', String(groesse.breite))
    p.mse(anna, 'einstellungen', 'setzen', 'fensterhoehe', String(groesse.hoehe))
    await p.starte({ ort: anna })

    await p.taste('5', 'befehl')
    await p.klickeText('Neuer Tausch …')
    // **Die Falte muss aufgehen, nicht nur geklickt sein** — R04 (#2094).
    await p.klappeAuf("[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Gegenüber aufnehmen oder bestätigen')", 'die Falte »Gegenüber aufnehmen oder bestätigen«')
    await p.warte('die Gegenüber stehen', () => p.seite(eineZeile), 20000)
    await new Promise((los) => setTimeout(los, 800))
    const kasten = await p.seite(messung)
    zeilen.push(zeile(groesse.name, 'Handeln › Neuer Tausch … › Gegenüber (acht)', kasten))
    p.pruefe(`R1: höchstens ein Hauptknopf je Region im Kasten (${groesse.name})`, kasten.hauptJeRegion <= 1, JSON.stringify(kasten))
    p.pruefe(`R5: nichts rollt waagerecht im Kasten (${groesse.name})`, !kasten.waagerecht)
    await p.bild(`gegenueber-kasten-${groesse.name}`)
    await p.taste('Escape')

    await zurSeite(p, 'Einstellungen')
    await p.warte('die Gruppe Verbindungen steht', () => p.seite(`return !!${mitText('Verbindungen')}`), 20000)
    await p.klickeAuf(mitText('Verbindungen'), 'die Gruppe Verbindungen')
    await p.warte('Kontakte zeigen steht', () => p.seite(`return !!${mitText('Kontakte zeigen')}`), 20000)
    await p.seite(`${mitText('Kontakte zeigen')}.scrollIntoView({ block: 'center' }); return true`)
    await p.klickeAuf(mitText('Kontakte zeigen'), 'Kontakte zeigen')
    await p.warte('die Gegenüber stehen', () => p.seite(eineZeile), 20000)
    await new Promise((los) => setTimeout(los, 800))
    const verbindungen = await p.seite(messung)
    // **Was oben steht** — ohne Bild sagt erst die Liste der Namen, was die
    // Bedienelemente im Bild sind (#2035).
    p.zeige(`im Bild unter Verbindungen (${groesse.name}): ${(await p.seite(imBild)).join(' · ')}`)
    zeilen.push(zeile(groesse.name, 'Einstellungen › Verbindungen (Kontakte gezeigt)', verbindungen))
    p.pruefe(`R1 unter Verbindungen (${groesse.name})`, verbindungen.hauptJeRegion <= 1, JSON.stringify(verbindungen))
    await p.bild(`gegenueber-verbindungen-${groesse.name}`)

    p.pruefe(`⌘Q beendet die Anwendung (${groesse.name})`, await p.beende())
  }

  p.notiz('Die Tabelle')
  p.zeige('| Größe | Ziel | Bedienelemente (im Bild) | Hauptknöpfe (je Region höchstens) | Felder (im Bild) | Überschriften | Rolltiefe | waagerecht |')
  p.zeige('| --- | --- | ---: | ---: | ---: | ---: | ---: | --- |')
  for (const z of zeilen) p.zeige(z)
}
