// **Das Menü Darstellung am nativen Fenster** — UXR-F14
// (#2029),
// Folge aus UXR-02 (#1910).
//
// *Darstellung* führt Seitenleiste, Navigator und Inspector mit ⌥⌘S, ⌥⌘[ und
// ⌥⌘]; die Einträge heißen nach dem Stand (Apple, Human Interface Guidelines,
// „The menu bar") und sind gegraut, wo die Seite das Teil nicht hat. Was
// `go test` und `vitest` nicht sehen, hier in der echten Leiste, so wie AppKit
// sie hält:
//
// 1. **Der Titel stimmt mit dem Bild überein** — steht die Spalte, heißt der
//    Eintrag „ausblenden", ist sie zu, „einblenden"; gelesen wird beides, die
//    Seite (`aria-expanded` am Knopf) und die Leiste (`isEnabled`, Titel),
// 2. **der Eintrag klappt**, und die Leiste folgt dem neuen Stand,
// 3. **das Kürzel klappt genau einmal** — die Seite hört es und macOS auch;
//    klappte beides, stünde die Spalte danach wieder da (derselbe Weg wie
//    ⌥⌘S in `eintraege`),
// 4. **auf einer Seite ohne Arbeitsplatz sind Navigator und Inspector grau**,
//    ein Klick darauf wird abgewiesen, und die Seitenleiste bleibt wählbar —
//    und beim Zurückkehren steht der Stand von vorhin,
// 5. **unter einer Schublade** sind beide grau und sagen weiter, wie die
//    Spalte darunter steht,
// 6. **bei Enge** (1.024 × 700) weicht der Navigator über das Zentrum: der
//    Eintrag öffnet ihn dort, und Escape schließt ihn — auch dann folgt die
//    Leiste,
// 7. **auf Englisch** dieselben Wege mit „Show/Hide".
//
// **Was das nicht belegt**: ein Rechner mit deutscher Tastatur. ⌥⌘[ liegt dort
// auf einer anderen Taste (das Fenster deutet sie über den Code, nicht über das
// Zeichen), und `p.taste` kennt nur die amerikanische Belegung — wie bei allen
// Kürzeln dieser Probe.
import { oeffnen, zurSeite } from './uirevue.mjs'

const durchlaeufe = [
  { name: 'breit', breite: 1440, hoehe: 900, sprache: 'de' },
  { name: 'eng', breite: 1024, hoehe: 700, sprache: 'de' },
  { name: 'englisch', breite: 1440, hoehe: 900, sprache: 'en' },
]

// woerter sind die Titel der Leiste je Sprache — dieselben wie
// `internal/app/menuetexte.go`, hier ein zweites Mal und mit Absicht: die
// Probe liest, was AppKit anzeigt, und eine gemeinsame Quelle zeigte nur, dass
// Go mit sich selbst übereinstimmt.
const woerter = {
  de: {
    menue: 'Darstellung',
    ausblenden: { seitenleiste: 'Seitenleiste ausblenden', navigator: 'Navigator ausblenden', inspector: 'Inspector ausblenden' },
    einblenden: { seitenleiste: 'Seitenleiste einblenden', navigator: 'Navigator einblenden', inspector: 'Inspector einblenden' },
    hilfe: ['Hilfe', 'MagicSetEditor4Mac-Hilfe'],
    zurueck: 'Zurück',
    schublade: 'Stil',
  },
  en: {
    menue: 'View',
    ausblenden: { seitenleiste: 'Hide Sidebar', navigator: 'Hide Navigator', inspector: 'Hide Inspector' },
    einblenden: { seitenleiste: 'Show Sidebar', navigator: 'Show Navigator', inspector: 'Show Inspector' },
    hilfe: ['Help', 'MagicSetEditor4Mac Help'],
    zurueck: 'Back',
  },
}

const teile = ['seitenleiste', 'navigator', 'inspector']
const kuerzel = { seitenleiste: '⌥⌘S', navigator: '⌥⌘[', inspector: '⌥⌘]' }
const tasten = { seitenleiste: 's', navigator: '[', inspector: ']' }

// seitenstand liest, wie die Seite dasteht: `true` offen, `false` zu, `null`
// fehlt. **Die Knöpfe sind über ihr Kürzel gefunden** (`aria-keyshortcuts`),
// denn `aria-controls` steht nur am offenen — und die Aufschrift ist je
// Sprache eine andere.
const seitenstand = `
  const offen = (wahl) => {
    const b = [...document.querySelectorAll(wahl)].find((e) => e.checkVisibility())
    return b ? b.getAttribute('aria-expanded') === 'true' : null
  }
  return {
    seitenleiste: offen('button[aria-controls="seitenleiste"]'),
    navigator: offen('main button[aria-keyshortcuts="Alt+Meta+["]'),
    inspector: offen('main button[aria-keyshortcuts="Alt+Meta+]"]'),
  }
`

// ueberZentrum sagt, ob der Navigator über dem Zentrum liegt (bei Enge).
const ueberZentrum = "return !!document.querySelector('main aside[data-ueber=zentrum]')"

// leistenstand liest die drei Einträge aus der echten Leiste — gefunden über
// ihr Kürzel, nicht über den Titel, der ja der Stand ist.
async function leistenstand(p, w) {
  const darstellung = (await p.menue()).find((m) => m.titel === w.menue)
  if (!darstellung) throw new Error(`kein Menü »${w.menue}«`)

  return Object.fromEntries(teile.map((teil) => {
    // **Über die zwei möglichen Titel, nicht über das Kürzel**: AppKit zeigt
    // ⌥⌘[ auf einer deutschen Tastatur als ⌥⌘Ö — die Anpassung der Kürzel an
    // die Belegung, gemessen am 2026-10-03.
    const eintrag = (darstellung.unter ?? []).find((e) => !e.trenner && (e.titel === w.ausblenden[teil] || e.titel === w.einblenden[teil]))

    return [teil, eintrag ? { titel: eintrag.titel, aktiv: eintrag.aktiv } : null]
  }))
}

// stimmtMit prüft, ob die Leiste sagt, was die Seite zeigt.
//
// **`null` auf der Seite heißt: das Teil fehlt, der Eintrag ist grau.** Sonst
// gilt der Titel nach dem Stand und ein Eintrag, den man wählen kann.
function stimmtMit(seite, leiste, w, teile_ = teile) {
  const fehler = []

  for (const teil of teile_) {
    const eintrag = leiste[teil]
    if (!eintrag) {
      fehler.push(`${teil}: kein Eintrag`)
      continue
    }
    if (eintrag.aktiv !== (seite[teil] !== null)) {
      fehler.push(`${teil}: aktiv ${String(eintrag.aktiv)}, Seite ${String(seite[teil])}`)
    }
    if (seite[teil] !== null && eintrag.titel !== (seite[teil] ? w.ausblenden[teil] : w.einblenden[teil])) {
      fehler.push(`${teil}: »${eintrag.titel}« bei ${seite[teil] ? 'offen' : 'zu'}`)
    }
  }

  return fehler
}

// folgt wartet, bis die Leiste dem Bild folgt, und gibt beides.
async function folgt(p, w, was, erwartet = null) {
  try {
    return await p.warte(was, async () => {
      const seite = await p.seite(seitenstand)
      if (erwartet && !erwartet(seite)) return false
      const leiste = await leistenstand(p, w)

      return stimmtMit(seite, leiste, w).length === 0 ? { seite, leiste } : false
    })
  } catch (fehler) {
    // **Was zuletzt dastand** — sonst sagt ein roter Lauf nur „nicht gefolgt".
    const darstellung = (await p.menue()).find((m) => m.titel === w.menue)
    p.zeige(`das Menü ${w.menue}: ${JSON.stringify((darstellung?.unter ?? []).map((e) => (e.trenner ? '—' : `${e.titel} [${e.kuerzel ?? ''}]`)))}`)
    await beiderSeite(p, w, `${was}, zuletzt`)
    throw fehler
  }
}

async function beiderSeite(p, w, was) {
  const seite = await p.seite(seitenstand)
  const leiste = await leistenstand(p, w)
  const fehler = stimmtMit(seite, leiste, w)
  p.zeige(`${was}: Seite ${JSON.stringify(seite)}, Leiste ${JSON.stringify(leiste)}`)
  p.pruefe(was, fehler.length === 0, fehler.join('; '))

  return { seite, leiste }
}

async function durchlauf(p, lauf) {
  const w = woerter[lauf.sprache]
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  if (lauf.sprache !== 'de') p.mse(ort, 'einstellungen', 'setzen', 'sprache', lauf.sprache)

  await oeffnen(p, ort, lauf)
  await p.warte('die Seite hat einen Arbeitsplatz', async () => (await p.seite(seitenstand)).navigator !== null)

  p.notiz(`${lauf.name} — 1: Titel und Aktivierung stimmen mit dem Bild überein`)
  const anfang = (await folgt(p, w, 'die Leiste folgt dem Bild')).leiste
  p.zeige(`Leiste: ${JSON.stringify(anfang)}`)
  p.pruefe('alle drei Einträge sind wählbar', teile.every((t) => anfang[t]?.aktiv))
  await p.bild(`darstellung-${lauf.name}`)

  for (const teil of ['navigator', 'inspector', 'seitenleiste']) {
    const vorher = (await p.seite(seitenstand))[teil]
    const titel = vorher ? w.ausblenden[teil] : w.einblenden[teil]

    p.notiz(`${lauf.name} — 2: »${titel}« klappt, und die Leiste folgt`)
    await p.klick(w.menue, titel)
    const danach = await folgt(p, w, `${teil} klappt auf den Eintrag`, (s) => s[teil] === !vorher)
    p.pruefe(`»${titel}« klappt (${teil}: ${String(vorher)} → ${String(danach.seite[teil])}), und der Eintrag heißt jetzt »${danach.leiste[teil].titel}«`, true)

    // Zurück über den Eintrag, der jetzt anders heißt.
    await p.klick(w.menue, danach.leiste[teil].titel)
    await folgt(p, w, `${teil} klappt zurück`, (s) => s[teil] === vorher)
    p.pruefe(`derselbe Eintrag klappt ${teil} zurück`, true)

    p.notiz(`${lauf.name} — 3: ${kuerzel[teil]} klappt genau einmal`)
    // **Nur wo AppKit das Kürzel so zeigt, wie die Probe es drücken kann**:
    // die Leitung tippt nach US-Belegung, und AppKit legt ⌥⌘[ auf einer
    // deutschen Tastatur auf ⌥⌘Ö (wie Finder und Mail). Dort gilt das Kürzel
    // für einen Menschen — aber nicht für diese Probe; das steht als
    // Handprüfung im Beleg und nicht als grüner Haken.
    const gezeigt = ((await p.menue()).find((m) => m.titel === w.menue)?.unter ?? [])
      .find((e) => !e.trenner && (e.titel === w.ausblenden[teil] || e.titel === w.einblenden[teil]))?.kuerzel
    if (gezeigt !== kuerzel[teil]) {
      p.zeige(`NICHT GEPRÜFT: ${kuerzel[teil]} — AppKit zeigt es auf dieser Belegung als ${String(gezeigt)}; von Hand prüfen`)
      continue
    }
    await p.taste(tasten[teil], 'wahl', 'befehl')
    const nachTaste = await folgt(p, w, `${kuerzel[teil]} klappt ${teil}`, (s) => s[teil] === !vorher)
    // **Ein zweites Mal hinsehen**: klappte Seite und Leiste, kippte es nach
    // einem Augenblick zurück.
    await new Promise((los) => setTimeout(los, 1500))
    const spaeter = await p.seite(seitenstand)
    p.pruefe(`${kuerzel[teil]} klappt ${teil} einmal und nicht zweimal`, spaeter[teil] === nachTaste.seite[teil], `${String(nachTaste.seite[teil])} → ${String(spaeter[teil])}`)
    await p.taste(tasten[teil], 'wahl', 'befehl')
    await folgt(p, w, `${kuerzel[teil]} klappt ${teil} zurück`, (s) => s[teil] === vorher)
  }

  if (lauf.name === 'eng') {
    p.notiz('eng — 6: der Navigator weicht über das Zentrum, und Escape schließt ihn')
    // Bei dieser Breite ist er zu; der Eintrag öffnet ihn über dem Zentrum.
    const vorher = (await p.seite(seitenstand)).navigator
    p.pruefe('der Navigator steht bei Enge nicht neben dem Zentrum', vorher === false, String(vorher))
    await p.klick(w.menue, w.einblenden.navigator)
    await folgt(p, w, 'der Navigator geht auf', (s) => s.navigator === true)
    p.pruefe('der Eintrag öffnet ihn über dem Zentrum', await p.seite(ueberZentrum))
    await p.taste('Escape')
    await folgt(p, w, 'Escape schließt ihn, und die Leiste folgt', (s) => s.navigator === false)
    p.pruefe('Escape schließt ihn, und der Eintrag heißt wieder »einblenden«', true)
  }

  p.notiz(`${lauf.name} — 4: auf einer Seite ohne Arbeitsplatz sind Navigator und Inspector grau`)
  const davor = await p.seite(seitenstand)
  await p.klick(...w.hilfe)
  await p.warte('die Seite hat keinen Arbeitsplatz', async () => (await p.seite(seitenstand)).navigator === null)
  const ohne = await folgt(p, w, 'die Leiste graut Navigator und Inspector')
  p.pruefe('Navigator und Inspector sind grau', !ohne.leiste.navigator.aktiv && !ohne.leiste.inspector.aktiv)
  p.pruefe('die Seitenleiste bleibt wählbar', ohne.leiste.seitenleiste.aktiv)
  let abgewiesen = ''
  try {
    await p.klick(w.menue, ohne.leiste.navigator.titel)
  } catch (grund) {
    abgewiesen = String(grund.message ?? grund)
  }
  p.pruefe('ein Klick auf den grauen Eintrag wird abgewiesen', abgewiesen.includes('nicht aktiv'), abgewiesen)

  await p.klick(w.menue, w.zurueck)
  await p.warte('der Arbeitsplatz ist wieder da', async () => (await p.seite(seitenstand)).navigator !== null)
  const zurueck = await folgt(p, w, 'die Leiste folgt dem Arbeitsplatz wieder')
  p.pruefe('derselbe Stand wie vorher', zurueck.seite.navigator === davor.navigator && zurueck.seite.inspector === davor.inspector, `${JSON.stringify(davor)} → ${JSON.stringify(zurueck.seite)}`)

  if (lauf.name === 'breit') {
    p.notiz('breit — 5: unter einer Schublade sind beide grau und nennen die Spalte darunter')
    const unten = await p.seite(seitenstand)
    await zurSeite(p, w.schublade, lauf.sprache)
    await p.warte('die Schublade steht', () => p.seite("return !![...document.querySelectorAll('[role=dialog]')].find((e) => e.checkVisibility())"))
    const darueber = await p.warte('die Leiste graut beide', async () => {
      const leiste = await leistenstand(p, w)

      return !leiste.navigator.aktiv && !leiste.inspector.aktiv ? leiste : false
    })
    p.pruefe('unter der Schublade sind Navigator und Inspector grau', true)
    p.pruefe('und sie sagen weiter, wie die Spalte darunter steht',
      darueber.navigator.titel === (unten.navigator ? w.ausblenden.navigator : w.einblenden.navigator) &&
      darueber.inspector.titel === (unten.inspector ? w.ausblenden.inspector : w.einblenden.inspector),
      `${darueber.navigator.titel} · ${darueber.inspector.titel}`)
    await p.klick(w.menue, w.zurueck)
    await folgt(p, w, 'ohne Schublade sind sie wieder wählbar')
    p.pruefe('ohne Schublade sind sie wieder wählbar', true)
  }

  await beiderSeite(p, w, `${lauf.name} — am Ende stimmen Leiste und Bild überein`)
  p.pruefe(`⌘Q beendet die Anwendung (${lauf.name})`, await p.beende())
}

export default async function darstellung(p) {
  for (const lauf of durchlaeufe) {
    await durchlauf(p, lauf)
  }
}
