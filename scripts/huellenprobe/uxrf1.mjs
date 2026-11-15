// **Die Auswahl mit Suche, am nativen Fenster** — UXR-F1
// (#2007),
// R7 aus `docs/UIREVIEW.md`; Migrationsschritt 2 aus
// [0125](../../docs/entscheidungen/0125-mantine-als-bauteilbibliothek.md).
//
// Gefahren wird die **Zeitzone** unter *Einstellungen* — die Auswahl mit
// 420 Zeilen, die einzige der vier Stellen (Stil, Testkarte, Spielwahl,
// Zeitzone), deren Liste nicht vom Bestand abhängt. Der Weg ist der eines
// Menschen, **ab AppKit**: Klick, Tippen und Tasten kommen als Ereignisse in
// die Webansicht (`huellenprobe.mjs`).
//
// - **Die Rollen**: Feld `combobox`, Liste `listbox`, Zeilen `option`; das Feld
//   steuert die Liste (`aria-controls`), meldet `aria-expanded` und nennt die
//   hervorgehobene Zeile in `aria-activedescendant`. Der Fokus bleibt im Feld.
//   **Gemessen werden die Attribute, aus denen WebKit den Zugänglichkeitsbaum
//   baut** — den Baum selbst liest die WebKit-Probe
//   (`scripts/webkitprobe/abnahme/uxrf1-suchauswahl.probe`, Motor `wk1`).
// - **Die Lage**: die Liste steht im Fenster, über allem, so breit wie das Feld.
// - **Die Tasten**: Klick wählt den Text, Tippen filtert, Pfeile wandern,
//   Enter wählt, Escape schließt und stellt den Wert her.
// - **Die Wirkung**: die gewählte Zone steht im Bestand; „wie das System"
//   schreibt den leeren Wert; ein Name, den die Liste nicht kennt, bleibt nach
//   einem Neustart sichtbar gewählt.
//
// **Was sie nicht belegt**: VoiceOvers eigene Ansage (`HANDPRUEFUNGEN.md`) und
// ob die Liste auf einem anderen Bildschirm dieselbe Lage nimmt.
import { readFileSync } from 'node:fs'

import { zurSeite } from './uirevue.mjs'

const zonen = JSON.parse(readFileSync(new URL('../../frontend/src/sprache/zeitzonen.json', import.meta.url), 'utf8'))

// grenze ist die Zahl der Zeilen, die die Liste höchstens zeichnet — gelesen
// aus dem Quelltext des Bausteins und nicht noch einmal hingeschrieben.
const grenze = Number(/export const zeilengrenze = (\d+)/.exec(readFileSync(new URL('../../frontend/src/bausteine/suchwahl.ts', import.meta.url), 'utf8'))?.[1])

const feld = "document.querySelector('main input[data-einstellung=\"zeitzone\"]')"

// zustand liest, was ein Vorleser am Feld und an der Liste erfährt, und wo
// die Liste steht.
const zustand = `
  const f = ${feld}
  if (!f) return null
  const liste = document.querySelector('[role=listbox]')
  const zeilen = liste ? [...liste.querySelectorAll('[role=option]')] : []
  const aktiv = f.getAttribute('aria-activedescendant')
  const zeile = aktiv ? document.getElementById(aktiv) : null
  const r = f.getBoundingClientRect()
  const l = liste?.getBoundingClientRect()
  const mitte = zeilen[0]?.getBoundingClientRect()
  const oben = mitte ? document.elementFromPoint(mitte.left + mitte.width / 2, mitte.top + mitte.height / 2) : null
  return {
    rolle: f.getAttribute('role'),
    name: (f.labels?.[0]?.innerText ?? '').replace(/\\s+/g, ' ').trim(),
    wert: f.value,
    platzhalter: f.getAttribute('placeholder'),
    offen: f.getAttribute('aria-expanded'),
    steuert: f.getAttribute('aria-controls'),
    haspopup: f.getAttribute('aria-haspopup'),
    autocomplete: f.getAttribute('aria-autocomplete'),
    gesperrt: f.disabled,
    listeRolle: liste?.getAttribute('role') ?? null,
    listeName: liste?.getAttribute('aria-label') ?? null,
    listeId: liste?.id ?? null,
    zeilen: zeilen.length,
    texte: zeilen.slice(0, 3).map((z) => z.innerText.trim()),
    alleTexte: zeilen.length <= 5 ? zeilen.map((z) => z.innerText.trim()) : [],
    stelle: zeile ? zeilen.indexOf(zeile) : -1,
    aktivRolle: zeile?.getAttribute('role') ?? null,
    aktivText: zeile?.innerText.trim() ?? null,
    aktivGewaehlt: zeile?.getAttribute('aria-selected') ?? null,
    fokusImFeld: document.activeElement === f,
    ansage: [...document.querySelectorAll('[data-suchauswahl]')].map((e) => e.innerText.trim()).find(Boolean) ?? '',
    listeImFenster: l ? l.left >= 0 && l.right <= innerWidth + 1 && l.bottom <= innerHeight + 1 : null,
    // **Der Rahmen um die Liste**, nicht die Liste: sie steht in einem Rand
    // von 2 Punkten je Seite und ist deshalb 4 Punkte schmaler als das Feld.
    listeBreit: liste?.parentElement ? Math.round(liste.parentElement.getBoundingClientRect().width) : null,
    feldBreit: Math.round(r.width),
    obenAufliegend: oben && liste ? liste.contains(oben) : null,
    markiert: [f.selectionStart, f.selectionEnd],
  }
`

// gespeichert liest die Zone aus dem Bestand — über die Kommandozeile, die
// dieselbe Bindung benutzt wie das Fenster.
const gespeichert = (p, ort) => JSON.parse(p.mse(ort, 'einstellungen', '--json')).einstellungen.find((e) => e.schluessel === 'zeitzone')?.wert

async function zurZeitzone(p, ort) {
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', '1280')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', '800')
  await p.starte({ ort })
  await zurSeite(p, 'Einstellungen')
  // **Die Gruppe *Allgemein* steht vorn**; fehlt das Feld, wird sie gewählt.
  if (!(await p.seite(`return !!${feld}?.checkVisibility()`))) await p.klickeText('Allgemein')
  await p.warte('die Zeitzone steht als Feld da', () => p.seite(`return !!${feld}?.checkVisibility()`), 20000)
}

// oeffneMitKlick klickt das Feld an, wie eine Maus es täte, und wartet, bis die
// Liste steht.
async function oeffneMitKlick(p) {
  await p.klickeAuf(feld, 'das Feld der Zeitzone')
  await p.warte('die Liste steht', () => p.seite("return document.querySelector('main input[data-einstellung=\"zeitzone\"]')?.getAttribute('aria-expanded') === 'true' && !!document.querySelector('[role=listbox]')"))
}

const geschlossen = (p) => p.warte('die Liste ist zu', () => p.seite(`return ${feld}?.getAttribute('aria-expanded') === 'false' && !document.querySelector('[role=listbox]')`))

export default async function uxrf1(p) {
  const ort = p.ort()
  await zurZeitzone(p, ort)

  p.notiz('Zu: die Rolle, der Name, der Wert — und keine Liste im Baum')
  const zu = await p.seite(zustand)
  p.zeige(`zu: ${JSON.stringify(zu)}`)
  p.pruefe('das Feld ist ein Kombinationsfeld', zu.rolle === 'combobox' && zu.haspopup === 'listbox' && zu.autocomplete === 'list')
  p.pruefe('es trägt den Namen der Einstellung', zu.name.startsWith('Zeitzone'))
  p.pruefe('es steht zu, ohne Liste im Baum', zu.offen === 'false' && zu.listeId === null && zu.steuert === null)
  p.pruefe('leer heißt „wie das System", mit Namen im Feld', zu.wert === 'Wie das System')
  await p.bild('zeitzone-zu')

  p.notiz('Ein Klick öffnet die Liste — der Fokus bleibt im Feld, der Text ist gewählt')
  await oeffneMitKlick(p)
  const offen = await p.seite(zustand)
  p.zeige(`offen: ${JSON.stringify(offen)}`)
  p.pruefe('das Feld meldet aufgeklappt und steuert die Liste', offen.offen === 'true' && offen.steuert === offen.listeId)
  p.pruefe('die Liste ist eine benannte listbox', offen.listeRolle === 'listbox' && offen.listeName === 'Zeitzone')
  p.pruefe(`die ersten ${String(grenze)} von ${String(zonen.length + 1)} Zeilen stehen da, die erste ist „wie das System"`, offen.zeilen === grenze && offen.texte[0] === 'Wie das System')
  p.pruefe('die hervorgehobene Zeile ist eine option und die gewählte', offen.aktivRolle === 'option' && offen.aktivText === 'Wie das System' && offen.aktivGewaehlt === 'true')
  p.pruefe('der Fokus bleibt im Feld', offen.fokusImFeld)
  // **Offen ist das Feld leer und nennt die Wahl als Platzhalter** — eine
  // Markierung per `select()` hielt WKWebView nach dem Klick nicht (`fmt02d`).
  p.pruefe('das Feld ist leer, die Wahl steht als Platzhalter da', offen.wert === '' && offen.platzhalter === 'Wie das System', JSON.stringify({ wert: offen.wert, platzhalter: offen.platzhalter }))
  p.pruefe('die Live-Region sagt, wie viele es sind und wie viele da stehen', offen.ansage === `${String(grenze)} von ${String(zonen.length + 1)} Treffern gezeigt — Suche eingrenzen …`)
  p.pruefe('die Liste steht im Fenster und über allem', offen.listeImFenster === true && offen.obenAufliegend === true)
  p.pruefe('sie ist so breit wie das Feld', Math.abs(offen.listeBreit - offen.feldBreit) <= 2)
  await p.bild('zeitzone-offen')

  p.notiz('Tippen filtert — auch mit Leerzeichen für den Unterstrich')
  await p.tippe('new york')
  const gefiltert = await p.warte('die Liste ist eingegrenzt', async () => {
    const z = await p.seite(zustand)
    return z.zeilen === 1 ? z : false
  })
  p.zeige(`gefiltert: ${JSON.stringify(gefiltert)}`)
  p.pruefe('„new york" findet genau America/New_York', gefiltert.alleTexte[0] === 'America/New_York')
  p.pruefe('der erste Treffer ist hervorgehoben', gefiltert.aktivText === 'America/New_York')
  p.pruefe('das Feld zeigt, was getippt ist', gefiltert.wert === 'new york')
  p.pruefe('nichts ist geschrieben, solange nicht gewählt ist', gespeichert(p, ort) === '')
  await p.bild('zeitzone-gefiltert')

  p.notiz('Enter wählt, schließt und schreibt')
  await p.taste('Enter')
  await geschlossen(p)
  await p.warte('die Zone steht im Bestand', async () => gespeichert(p, ort) === 'America/New_York')
  const gewaehlt = await p.seite(zustand)
  p.zeige(`gewählt: ${JSON.stringify(gewaehlt)}`)
  p.pruefe('das Feld zeigt die Zone', gewaehlt.wert === 'America/New_York')
  p.pruefe('der Fokus ist im Feld geblieben', gewaehlt.fokusImFeld)

  p.notiz('Nur mit der Tastatur: Pfeil nach unten öffnet bei der Wahl, die Pfeile wandern')
  await p.taste('ArrowDown')
  await p.warte('die Liste steht', () => p.seite("return !!document.querySelector('[role=listbox]')"))
  const bei = await p.seite(zustand)
  p.pruefe('geöffnet steht die Hervorhebung auf der gewählten Zone', bei.aktivText === 'America/New_York')
  await p.taste('ArrowDown')
  await p.taste('ArrowDown')
  const weiter = await p.warte('die Hervorhebung ist zwei Zeilen weiter', async () => {
    const z = await p.seite(zustand)
    return z.stelle === bei.stelle + 2 ? z : false
  })
  p.zeige(`zwei weiter: ${weiter.aktivText}`)
  await p.taste('ArrowUp')
  const zurueck = await p.warte('die Hervorhebung ist eine Zeile zurück', async () => {
    const z = await p.seite(zustand)
    return z.stelle === bei.stelle + 1 ? z : false
  })
  p.pruefe('die Pfeile wandern, und der Fokus bleibt im Feld', zurueck.fokusImFeld && zurueck.aktivRolle === 'option')

  p.notiz('Escape schließt, verwirft die Eingabe und stellt den Wert her')
  await p.taste('a', 'befehl')
  await p.tippe('tokyo')
  await p.warte('die Liste zeigt Tokio', async () => (await p.seite(zustand)).alleTexte.includes('Asia/Tokyo'))
  await p.taste('Escape')
  await geschlossen(p)
  const verworfen = await p.seite(zustand)
  p.pruefe('der Wert steht wieder da', verworfen.wert === 'America/New_York')
  p.pruefe('nichts ist geschrieben', gespeichert(p, ort) === 'America/New_York')
  await p.bild('zeitzone-escape')

  p.notiz('Die letzte Zone steht hinter der Grenze — und lässt sich trotzdem finden')
  p.pruefe('UTC steht hinter der Grenze', zonen.indexOf('UTC') > grenze)
  await oeffneMitKlick(p)
  await p.tippe('utc')
  await p.warte('die Liste zeigt nur UTC', async () => (await p.seite(zustand)).alleTexte[0] === 'UTC')
  await p.taste('Enter')
  await geschlossen(p)
  await p.warte('UTC steht im Bestand', async () => gespeichert(p, ort) === 'UTC')

  p.notiz('„Wie das System" schreibt den leeren Wert — der Weg zurück')
  await oeffneMitKlick(p)
  await p.tippe('wie das')
  await p.warte('die Liste zeigt die Systemzeile', async () => (await p.seite(zustand)).alleTexte[0] === 'Wie das System')
  await p.taste('Enter')
  await geschlossen(p)
  await p.warte('der Bestand ist wieder leer', async () => gespeichert(p, ort) === '')
  p.pruefe('das Feld nennt „wie das System"', (await p.seite(zustand)).wert === 'Wie das System')

  p.notiz('Nichts passt: ein Satz, keine Liste, kein Schreiben')
  await oeffneMitKlick(p)
  await p.tippe('zzz')
  const leer = await p.warte('die Meldung steht', async () => {
    const z = await p.seite(zustand)
    return z.zeilen === 0 && z.ansage !== '' ? z : false
  })
  p.zeige(`leer: ${leer.ansage}`)
  p.pruefe('die Live-Region sagt, dass nichts passt', leer.ansage.includes('zzz'))
  p.pruefe('das Feld steht nicht als aufgeklappt da', leer.offen === 'false' && leer.steuert === null)
  await p.bild('zeitzone-leer')
  await p.taste('Escape')
  p.pruefe('⌘Q beendet die Anwendung', await p.beende())

  p.notiz('Ein Name, den die Liste nicht kennt, bleibt sichtbar gewählt')
  p.pruefe('US/Pacific steht nicht in der Liste', !zonen.includes('US/Pacific'))
  p.mse(ort, 'einstellungen', 'setzen', 'zeitzone', 'US/Pacific')
  await zurZeitzone(p, ort)
  const fremd = await p.seite(zustand)
  p.pruefe('das Feld zeigt ihn, wie er gespeichert ist', fremd.wert === 'US/Pacific')
  await oeffneMitKlick(p)
  const fremdOffen = await p.seite(zustand)
  p.pruefe('geöffnet ist keine Zeile hervorgehoben, und die Liste steht trotzdem', fremdOffen.zeilen === grenze && fremdOffen.stelle === -1)
  await p.taste('Escape')
  p.pruefe('⌘Q beendet die Anwendung (nach dem Neustart)', await p.beende())
}
