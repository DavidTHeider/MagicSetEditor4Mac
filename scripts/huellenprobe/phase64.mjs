// **Was Phase 6.4 am Fenster ändert, am nativen Fenster** — OPS-03.AK-04
// (#1720)
// und die Befunde M15, G13 und G6 des Zweitblicks über Phase 6.3
// (#1665):
//
// - *Einstellungen › Verbindungen*: die Befehlszeile nennt Verzeichnis, Stände
//   und den Prüfbefehl, und neben dem Prüfbefehl steht ein Kopierknopf.
//   **Gedrückt wird er nicht** — die Probe fasst die Zwischenablage dieses
//   Rechners nicht an; dass er kopiert, belegen die Läufe unter jsdom,
// - *Einstellungen › Erscheinungsbild*: die eigenen Farben sagen, woher sie
//   kommen und wann sie wirken, und haben einen Weg zurück zur Vorgabe,
// - *Einstellungen › Allgemein*: dasselbe an der Belegung von ⌘0 bis ⌘9,
// - mit der Leiste oben schaltet ⌥⌘S nichts Unsichtbares um — gelesen am
//   Bestand über `mse`, nicht am Fenster.

// hier ist die Seite, auf der die Brotkrume steht.
const hier = (p) => p.seite("return document.querySelector('nav ol [aria-current=page]')?.innerText.trim() ?? ''")

// gruppe schlägt eine Gruppe der Einstellungen auf und wartet auf einen Text.
async function gruppe(p, name, text) {
  await p.klickeText(name)
  return p.wartetext(text, 20000)
}

const knopfDa = (p, bezeichnung) => p.seite(`return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? b.innerText).trim() === ${JSON.stringify(bezeichnung)})`)

export default async function phase64(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'gestaltung', '--mse-marke: #1f4e79;')
  p.mse(ort, 'einstellungen', 'setzen', 'kuerzel.ziffern', '3=karten')
  // **Die Leiste oben von Anfang an** — das Fenster liest die Lage beim
  // Start; ein Setzen während des Laufs sähe es erst beim nächsten.
  p.mse(ort, 'einstellungen', 'setzen', 'leistenlage', 'waagerecht')
  await p.starte({ ort })

  await p.taste(',', 'befehl')
  await p.warte('die Einstellungen stehen da', async () => (await hier(p)) !== '', 20000)

  p.notiz('OPS-03.AK-04 — der Prüfbefehl lässt sich kopieren')
  // **Seit AP-10 (#2367) unter Verbindungen** und nicht mehr unter Diagnose.
  await p.klickeText('Verbindungen')
  // **Seit UXR-19 (#1927) in einem Abschnitt** — erst aufklappen.
  await gruppe(p, 'Werkzeuge ohne Fenster', 'Die Kommandozeile einrichten')
  // **Der Stand kommt nach dem Aufklappen** — der Abschnitt zeichnet seinen
  // Inhalt erst dann und fragt die Kommandozeile danach (UXR-19).
  await p.wartetext('mse meta', 20000)
  await p.seite("[...document.querySelectorAll('h3')].find((h) => h.innerText.includes('Die Kommandozeile einrichten'))?.scrollIntoView({ block: 'center' }); return true")
  p.pruefe('neben dem Prüfbefehl steht ein Kopierknopf', await knopfDa(p, 'Den Prüfbefehl kopieren'))
  const stand = await p.seite("const h = [...document.querySelectorAll('h3')].find((x) => x.innerText.includes('Die Kommandozeile einrichten')); return h ? h.parentElement.innerText.split('\\n').filter((z) => z.trim() !== '').join(' | ') : ''")
  p.zeige(`die Befehlszeile: ${stand}`)
  p.pruefe('sie nennt den Prüfbefehl »mse meta«', stand.includes('mse meta'), stand)
  await p.bild('befehlszeile')

  p.notiz('M15 — die eigenen Farben sagen, woher sie kommen')
  await p.klickeText('Erscheinungsbild')
  // **Seit #2465 in einem Abschnitt** — erst aufklappen.
  await gruppe(p, 'Eigene Farben', 'Eigene Farben auf die Vorgabe zurücksetzen')
  const farben = await p.seite("const f = document.querySelector('[data-einstellung=gestaltung]'); return f ? f.closest('section').innerText : ''")
  // **Seit AP-11 (#2368) eine Marke** — der ganze Satz ist ihr Titel und
  // steht für das Vorleseprogramm; „Wirkt sofort.“ sagt nichts mehr.
  const farbmarke = await p.seite("const m = document.querySelector('[data-einstellung=gestaltung]')?.closest('section')?.querySelector('[data-herkunftsmarke]'); return m ? `${m.innerText} | ${m.title}` : ''")
  p.pruefe('Herkunft und Ebene stehen am Feld der eigenen Farben, als Marke mit dem ganzen Satz', farben.includes('Gesetzt von diesem Profil, Ebene Profil.') && farbmarke === 'von diesem Profil | Gesetzt von diesem Profil, Ebene Profil.', farbmarke)
  p.pruefe('„Wirkt sofort.“ steht dort nicht mehr', !farben.includes('Wirkt sofort'), farben.split('\n').filter((z) => /Gesetzt|Wirkt/.test(z)).join(' | '))
  await p.seite("[...document.querySelectorAll('button')].find((b) => b.innerText.includes('Eigene Farben auf die Vorgabe'))?.scrollIntoView({ block: 'center' }); return true")
  await p.bild('eigene-farben')

  p.notiz('M15 und G13 — die Belegung sagt dasselbe')
  await p.klickeText('Allgemein')
  // **Seit UXR-19 (#1927) in einem Abschnitt** — erst aufklappen.
  await gruppe(p, 'Tastenkürzel', 'Belegung')
  await p.seite("document.querySelector('[data-einstellung=\"kuerzel.ziffern\"]')?.scrollIntoView({ block: 'center' }); return true")
  const belegung = await p.seite("return document.querySelector('[data-einstellung=\"kuerzel.ziffern\"]')?.innerText ?? ''")
  // **Gesetzt vom Profil, obwohl die Frage dem Menschen gehört** — die
  // Ablage liegt je Profil (`docs/EINSTELLUNGEN.md`, Befund M13).
  p.pruefe('an der Belegung steht die Herkunft, „Wirkt sofort.“ nicht mehr', /Gesetzt von .+, Ebene Nutzer\./.test(belegung) && !belegung.includes('Wirkt sofort'), belegung.split('\n').filter((z) => /Gesetzt|Wirkt|von /.test(z)).join(' | '))
  await p.bild('belegung')

  p.notiz('G6 — mit der Leiste oben schaltet ⌥⌘S nichts um')
  const oben = await p.seite("return document.querySelector('[data-lage=waagerecht]') !== null")
  p.pruefe('die Leiste steht oben', oben)
  await p.taste('s', 'wahl', 'befehl')
  await new Promise((los) => setTimeout(los, 1000))
  const nachher = p.mse(ort, 'einstellungen')
  p.pruefe('die Seitenleiste bleibt im Bestand sichtbar', !/seitenleiste\s+verborgen/.test(nachher), nachher.split('\n').filter((z) => z.startsWith('seitenleiste') || z.startsWith('leistenlage')).join(' | '))

  p.pruefe('⌘Q beendet', await p.beende())
}
