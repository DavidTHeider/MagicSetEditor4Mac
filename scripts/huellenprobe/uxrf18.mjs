// **Eine Sicht mit gewählten Drucken und mit Behaltemenge, am Fenster** —
// UXR-F18 (#2033).
//
// Der Probebestand kennt weder ein Gegenüber noch eine Sammlung. Dieses
// Drehbuch legt sechzehn Drucke an — eins über der Grenze, ab der die Wahl
// ein Suchfeld bekommt (R7) — und ein bestätigtes Gegenüber mit dem Zweck
// `handel` und `teilen`. Dann am Fenster, über *Einstellungen* ›
// *Verbindungen* › *Verwalten*:
//
// 1. „Ausgewählte Drucke": zwei Drucke über die Suche wählen, erteilen — die
//    Zeile nennt „2 Drucke".
// 2. „Dubletten": die Behaltemenge steht auf 1, auf 2 stellen, erteilen —
//    die Zeile nennt „du behältst 2".
//
// Gemessen wird mit **derselben** Messung wie `uirevue`, bei 1280 × 800 und
// 1024 × 700. **Was es nicht tut**: VoiceOver. Rollen und Namen hält axe im
// Test; den Weg mit dem Bildschirmleser sieht ein Mensch.
import { messung, zurSeite } from './uirevue.mjs'

const groessen = [
  { name: 'normal', breite: 1280, hoehe: 800 },
  { name: 'eng', breite: 1024, hoehe: 700 },
]

const drucke = ['Ancestral Recall', 'Birds of Paradise', 'Counterspell', 'Dark Ritual',
  'Earthquake', 'Fireball', 'Giant Growth', 'Hurricane', 'Icy Manipulator', 'Juggernaut',
  'Kudzu', 'Lightning Bolt', 'Mahamoti Djinn', 'Nevinyrral\'s Disk', 'Orcish Artillery',
  'Prodigal Sorcerer']

const zeilenwert = (ausgabe, name) => ausgabe.split('\n').find((z) => z.startsWith(`${name}\t`))?.split('\t')[1] ?? ''
const mitText = (text) => `[...document.querySelectorAll('button, [role=tab]')].find((b) => b.checkVisibility() && b.innerText.trim().split('\\n')[0] === ${JSON.stringify(text)})`
const kaestchen = (name) => `[...document.querySelectorAll('input[type=checkbox]')].find((k) => k.checkVisibility() && k.closest('label')?.innerText.includes(${JSON.stringify(name)}))`
const feldMit = (beschriftung) => `[...document.querySelectorAll('input, select')].find((f) => f.checkVisibility() && [...(f.labels ?? [])].some((l) => l.innerText.trim().startsWith(${JSON.stringify(beschriftung)})))`
const steht = (muster) => `return ${muster}.test(document.querySelector('main').innerText)`

const zeile = (groesse, ziel, z) =>
  `| ${groesse} | ${ziel} | ${String(z.bedien)} (${String(z.bedienImBild)}) | ${String(z.haupt)} (${String(z.hauptJeRegion)}) | ${String(z.felder)} (${String(z.felderImBild)}) | ${String(z.ueberschriften)} | ${String(z.rolltiefe)} | ${z.waagerecht ? 'ja' : 'nein'} |`

// **Die Liste der Umfänge, erkannt an ihrer Option** — die Beschriftung „Was"
// trägt auch die Notiz zum Gegenüber, und ein Setter für Auswahllisten
// scheitert an einem Textfeld.
const umfangsliste = "[...document.querySelectorAll('select')].find((s) => s.checkVisibility() && [...s.options].some((o) => o.value === 'dubletten'))"

async function umfang(p, wert) {
  await p.warte('die Liste der Umfänge steht', () => p.seite(`return !!${umfangsliste}`), 20000)
  await p.seite(`const s = ${umfangsliste}; const setze = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set; setze.call(s, ${JSON.stringify(wert)}); s.dispatchEvent(new Event('change', { bubbles: true })); return true`)
}

async function zumGegenueber(p) {
  await zurSeite(p, 'Einstellungen')
  await p.warte('die Gruppe Verbindungen steht', () => p.seite(`return !!${mitText('Verbindungen')}`), 20000)
  await p.klickeAuf(mitText('Verbindungen'), 'die Gruppe Verbindungen')
  await p.warte('Kontakte zeigen steht', () => p.seite(`return !!${mitText('Kontakte zeigen')}`), 20000)
  await p.seite(`${mitText('Kontakte zeigen')}.scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(mitText('Kontakte zeigen'), 'Kontakte zeigen')
  const verwalten = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.getAttribute('aria-label') === 'Verwalten: Bert')"
  await p.warte('Bert steht', () => p.seite(`return !!${verwalten}`), 20000)
  // **Ein einziges Gegenüber steht von selbst offen** (`vorgabe` in
  // `Peerkontakte.tsx`) — ein Klick auf „Verwalten" klappte es zu.
  await new Promise((los) => setTimeout(los, 800))
  if (!(await p.seite(steht('/Was dieses Gegenüber sieht/')))) await p.klickeAuf(verwalten, 'Verwalten: Bert')
  await p.warte('die Sichten stehen', () => p.seite(steht('/Was dieses Gegenüber sieht/')), 20000)
  // **Die Sichten stehen hinter einem eigenen Kasten** — aufklappen, wenn er zu
  // ist.
  const kasten = "[...document.querySelectorAll('button, summary')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Was dieses Gegenüber sieht')"
  await new Promise((los) => setTimeout(los, 500))
  if (!(await p.seite(`return !!${umfangsliste}`))) {
    await p.seite(`${kasten}.scrollIntoView({ block: 'center' }); return true`)
    await p.klickeAuf(kasten, 'den Kasten „Was dieses Gegenüber sieht"')
  }
  await p.warte('die Liste der Umfänge steht', () => p.seite(`return !!${umfangsliste}`), 20000)
  await p.seite(`${umfangsliste}.scrollIntoView({ block: 'center' }); return true`)
}

export default async function uxrf18(p) {
  const anna = p.ort()
  p.mse(anna, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(anna, 'peer-anlegen')
  for (const [stelle, name] of drucke.entries()) {
    const druck = `a1b2c3d4-0000-4000-8000-${String(stelle + 1).padStart(12, '0')}`
    p.mse(anna, 'exemplar', druck, String(1 + (stelle % 3)), 'played', 'de', 'nonfoil', '', '', '', '',
      name, 'lea', String(100 + stelle), '')
  }
  const fremd = p.ort()
  p.mse(fremd, 'peer-anlegen')
  const ich = p.mse(fremd, 'peer')
  const kennung = zeilenwert(ich, 'kennung')
  p.mse(anna, 'kontakt-aufnehmen', kennung, 'Bert')
  p.mse(anna, 'kontakt-bestaetigen', kennung, zeilenwert(ich, 'fingerabdruck'))
  p.mse(anna, 'zweck-erlauben', kennung, 'handel')
  p.mse(anna, 'zweck-erlauben', kennung, 'teilen')

  const zeilen = []
  for (const groesse of groessen) {
    p.mse(anna, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
    p.mse(anna, 'einstellungen', 'setzen', 'fensterbreite', String(groesse.breite))
    p.mse(anna, 'einstellungen', 'setzen', 'fensterhoehe', String(groesse.hoehe))
    await p.starte({ ort: anna })
    await zumGegenueber(p)

    if (groesse.name === 'normal') {
      // **1. Ausgewählte Drucke** — sechzehn sind mehr als vierzehn, also
      // steht ein Suchfeld vor der Wahl.
      await umfang(p, 'auswahl')
      await p.warte('das Suchfeld der Drucke steht', () => p.seite(`return !!${feldMit('Drucke suchen')}`), 20000)
      for (const name of ['Fireball', 'Kudzu']) {
        await p.seite(`const f = ${feldMit('Drucke suchen')}; const setze = Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set; setze.call(f, ''); f.dispatchEvent(new Event('input', { bubbles: true })); f.focus(); return true`)
        await p.tippeInsFeld(name)
        await p.warte(`${name} steht zur Wahl`, () => p.seite(`return !!${kaestchen(name)}`), 20000)
        await p.klickeAuf(kaestchen(name), `das Kästchen ${name}`)
        await p.warte(`${name} ist gewählt`, () => p.seite(`return ${kaestchen(name)}?.checked === true`), 20000)
      }
      await p.klickeText('Sicht erteilen')
      await p.warte('die Zeile nennt zwei Drucke', () => p.seite(steht('/2 Drucke/')), 20000)
      p.pruefe('„Ausgewählte Drucke" erteilt, die Zeile nennt zwei Drucke', true)
      await p.bild('sicht-auswahl')

      // **2. Dubletten** — die Vorgabe ist 1, gestellt wird 2.
      await umfang(p, 'dubletten')
      await p.warte('die Behaltemenge steht', () => p.seite(`return !!${feldMit('Davon behalte ich')}`), 20000)
      p.pruefe('die Behaltemenge steht auf 1', (await p.seite(`return ${feldMit('Davon behalte ich')}.value`)) === '1')
      await p.seite(`const f = ${feldMit('Davon behalte ich')}; f.focus(); f.select(); return true`)
      await p.tippe('2')
      await p.warte('im Feld steht 2', () => p.seite(`return ${feldMit('Davon behalte ich')}.value === '2'`), 20000)
      await p.klickeText('Sicht erteilen')
      await p.warte('die Zeile nennt die Behaltemenge', () => p.seite(steht('/du behältst 2/')), 20000)
      p.pruefe('„Dubletten" erteilt, die Zeile nennt „du behältst 2"', true)
    }

    await new Promise((los) => setTimeout(los, 800))
    const gemessen = await p.seite(messung)
    zeilen.push(zeile(groesse.name, 'Einstellungen › Verbindungen › Bert › Sichten', gemessen))
    p.pruefe(`R1: höchstens ein Hauptknopf je Region (${groesse.name})`, gemessen.hauptJeRegion <= 1, JSON.stringify(gemessen))
    p.pruefe(`R5: nichts rollt waagerecht (${groesse.name})`, !gemessen.waagerecht)
    await p.bild(`sicht-dubletten-${groesse.name}`)

    p.pruefe(`⌘Q beendet die Anwendung (${groesse.name})`, await p.beende())
  }

  // **Und die Kommandozeile liest dasselbe** — die Freigabe hängt nicht an der
  // Fläche, die sie erteilt hat (§16).
  const gelesen = p.mse(anna, 'kontakt-freigaben', kennung)
  p.pruefe('die Kommandozeile liest beide Sichten', /auswahl/.test(gelesen) && /dubletten/.test(gelesen), gelesen)

  p.notiz('Die Tabelle')
  p.zeige('| Größe | Ziel | Bedienelemente (im Bild) | Hauptknöpfe (je Region höchstens) | Felder (im Bild) | Überschriften | Rolltiefe | waagerecht |')
  p.zeige('| --- | --- | ---: | ---: | ---: | ---: | ---: | --- |')
  for (const z of zeilen) p.zeige(z)
}
