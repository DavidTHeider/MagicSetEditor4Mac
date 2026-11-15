// **Die entschlackten Flächen als Bild, in zwei Sprachen und zwei
// Erscheinungsbildern** — Sichtprüfung zu Phase 6.13
// (#1906),
// AK-03 jedes UXR-Issues: „DE und EN, hell und dunkel".
//
// `uirevue` zählt, dieses Drehbuch nimmt auf: je Sprache und Erscheinungsbild
// das Fenster mit jeder Fläche, die ein UXR-Werkstück umgebaut hat, bei
// 1280 × 800. **Was es nicht tut**: urteilen. Ob ein Bild stimmt, sieht ein
// Mensch; das Drehbuch sorgt dafür, dass es dasselbe Bild ist, das das native
// Fenster zeigt, und nicht das eines Browsers.
import { oeffnen, seitennamen, zurSeite } from './uirevue.mjs'

// Die Stellen in der Seitenliste der Palette — über die Stelle und nicht über
// den Namen, denn im Englischen heißen die Seiten anders.
// **Je Fläche ihr Name in beiden Sprachen, nicht ihre Stelle** (Z3 des
// Zweitblicks über Phase 6.13): eine verschobene Palette hätte unter dem
// Dateinamen der erwarteten Fläche eine andere fotografiert, und der Zähler
// wäre trotzdem voll gewesen. Steht ein Name nicht in der Palette, ist das
// ein Befund und kein Bild.
const flaechen = [
  { name: 'start', de: 'Start', en: 'Home' },
  { name: 'sets', de: 'Sets', en: 'Sets' },
  { name: 'pakete', de: 'Pakete', en: 'Packages' },
  { name: 'vorlagen', de: 'Vorlagen', en: 'Templates' },
  { name: 'sammlung', de: 'Sammlung', en: 'Collection' },
  { name: 'decks', de: 'Decks', en: 'Decks' },
  { name: 'turniere', de: 'Turniere', en: 'Tournaments' },
  { name: 'spielen', de: 'Spielen', en: 'Play' },
  { name: 'handeln', de: 'Handeln', en: 'Trading' },
  { name: 'einstellungen', de: 'Einstellungen', en: 'Settings' },
  { name: 'hilfe', de: 'Hilfe', en: 'Help' },
  { name: 'nutzer', de: 'Nutzer und Profile', en: 'Users and profiles' },
  { name: 'karten', de: 'Karten', en: 'Cards' },
  { name: 'suche', de: 'Suche', en: 'Search' },
  { name: 'stil', de: 'Stil', en: 'Style' },
  { name: 'symbol', de: 'Symbol', en: 'Symbol' },
  { name: 'statistik', de: 'Statistik', en: 'Statistics' },
  { name: 'ausgeben', de: 'Ausgeben', en: 'Export' },
  { name: 'druckbogen', de: 'Druckbogen', en: 'Print sheet' },
  { name: 'konsole', de: 'Konsole', en: 'Console' },
]

export default async function uxrabzug(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')

  let aufgenommen = 0
  for (const sprache of ['de', 'en']) {
    for (const bild of ['hell', 'dunkel']) {
      p.mse(ort, 'einstellungen', 'setzen', 'sprache', sprache)
      p.mse(ort, 'einstellungen', 'setzen', 'erscheinungsbild', bild)
      await oeffnen(p, ort, { name: `${sprache}-${bild}`, breite: 1280, hoehe: 800 })
      const namen = await seitennamen(p, sprache)
      p.zeige(`${sprache}, ${bild}: ${namen.join(' · ')}`)
      for (const flaeche of flaechen) {
        const name = flaeche[sprache]
        if (!namen.includes(name)) {
          p.pruefe(`${flaeche.name} steht in der Palette (${sprache}, ${bild})`, false)
          continue
        }
        await zurSeite(p, name, sprache)
        if (await p.bild(`${flaeche.name}-${sprache}-${bild}`)) aufgenommen++
        await p.taste('Escape')
      }
      p.pruefe(`⌘Q beendet die Anwendung (${sprache}, ${bild})`, await p.beende())
    }
  }

  p.pruefe('jede Fläche aufgenommen', aufgenommen === flaechen.length * 4, `${String(aufgenommen)} von ${String(flaechen.length * 4)}`)
}
