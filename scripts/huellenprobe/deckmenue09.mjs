// **Das Menü „Deck“ ohne Zeile — am nativen Fenster** — Rest R6 aus dem
// Nachlauf über Phase 6 von 0.9.0
// (#2469).
//
// # Was es prüft
//
// - **Eine Suche, die das offene Deck nicht trifft**: es bleibt in der Liste
//   stehen, und die Menüleiste führt „Deck“.
// - **Der Navigator ist zu**: die Menüleiste führt „Deck“ weiter, und
//   „Deck › Umbenennen“ aus der echten Leiste klappt ihn auf und zeigt das
//   Feld „Neuer Name“ an der Zeile.
//
// # Was es nicht belegt
//
// **Höchstens zwanzig Bedienelemente im Bild** — das zählt
// `uirevue09zustaende`; hier kommt keines hinzu. **VoiceOver** hört ein
// Mensch.
const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? b.innerText).trim() === ${JSON.stringify(name)})`
const marke = "[...document.querySelectorAll('span')].find((e) => e.checkVisibility() && e.innerText.trim() === 'offen')"
const feldMitWert = (wert) => `return [...document.querySelectorAll('input')].some((e) => e.checkVisibility() && e.value === ${JSON.stringify(wert)})`
const leistenmenue = async (p, titel) => (await p.menue()).find((m) => m.titel === titel)

export default async function deckmenue09(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'ersteschritte', 'aus')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Waldschritt', 'probe')
  p.mse(ort, 'deck', 'neu', 'Glutpfad', 'probe')

  await p.starte({ ort })
  await p.groesse(1440, 838)
  await p.taste('4', 'befehl')
  await p.warte('die Deckliste steht', () => p.seite(`return !!${knopf('Öffnen: Waldschritt')}`), 20000)
  await p.klickeAuf(knopf('Öffnen: Waldschritt'), '»Öffnen: Waldschritt«')
  await p.warte('»…« am offenen Deck', () => p.seite(`return !!${knopf('Handlungen — Waldschritt')}`), 20000)
  p.pruefe('die Menüleiste führt »Deck«', !!(await p.warte('»Deck« in der Leiste', () => leistenmenue(p, 'Deck'), 10000).catch(() => null)))

  p.notiz('Eine Suche, die das offene Deck nicht trifft')
  await p.seite("const e = [...document.querySelectorAll('input[type=search]')].find((e) => e.checkVisibility()); e.focus(); return true")
  await p.tippeInsFeld('zzz')
  await p.warte('die Suche nimmt das andere Deck weg', async () => !(await p.seite(`return !!${knopf('Öffnen: Glutpfad')}`)), 10000)
  p.pruefe('das offene Deck bleibt in der gefilterten Liste', !!(await p.seite(`return !!${marke}`)))
  p.pruefe('die Menüleiste führt »Deck« auch dann', !!(await leistenmenue(p, 'Deck')))
  await p.bild('gefiltert')
  await p.seite("const e = [...document.querySelectorAll('input[type=search]')].find((e) => e.checkVisibility()); const setze = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set; setze.call(e, ''); e.dispatchEvent(new Event('input', { bubbles: true })); return true")

  p.notiz('Der Navigator ist zu')
  await p.klickeAuf(knopf('Navigator ausblenden'), '»Navigator ausblenden«')
  await p.warte('der Navigator ist zu', async () => !(await p.seite(`return !!${marke}`)), 10000)
  p.pruefe('die Menüleiste führt »Deck« bei eingeklapptem Navigator', !!(await leistenmenue(p, 'Deck')))
  await p.klick('Deck', 'Umbenennen')
  const feld = await p.warte('das Feld »Neuer Name« steht', () => p.seite(feldMitWert('Waldschritt')), 10000).catch(() => false)
  p.pruefe('»Deck › Umbenennen« aus der Leiste zeigt das Feld an der Zeile', feld === true)
  p.pruefe('der Navigator ist wieder offen', !!(await p.seite(`return !!${marke}`)))
  await p.bild('umbenennen-aus-der-leiste')
  await p.taste('Escape')

  p.pruefe('⌘Q beendet die Anwendung', await p.beende())
}
