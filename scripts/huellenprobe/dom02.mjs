// **Die Symbolbibliothek — am nativen Fenster** — DOM-02.AK-01, AK-02 und
// AK-04 (#1697):
//
// - AK-01: im Fach *Symbol* steht unter dem Editor die Bibliothek; das
//   gezeichnete Setsymbol wird aufgenommen, dupliziert, umbenannt, über den
//   Dialog von macOS aus- und wieder eingelesen; die Symbolschriften stehen
//   getrennt daneben,
// - Einsetzen: ein zweites Set bekommt das Symbol als Kopie, und der Editor
//   zeigt es sofort,
// - AK-02: Löschen nennt vorher die Sets, die es tragen, und *Abbrechen*
//   löscht nicht,
// - ein doppelter Name steht am Namensfeld (§6).
import { existsSync } from 'node:fs'
import { join } from 'node:path'

// **Das Zeichen des Katalogs steht vor dem Text** — gesucht wird deshalb nach
// dem Ende.
export const knopfIn = (zeile, name) => `[...(${zeile})?.querySelectorAll('button') ?? []].find((b) => b.checkVisibility() && b.innerText.trim().endsWith(${JSON.stringify(name)}))`
export const eintrag = (name) => `[...document.querySelectorAll('li[aria-labelledby]')].find((l) => document.getElementById(l.getAttribute('aria-labelledby'))?.innerText === ${JSON.stringify(name)})`
export const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith(${JSON.stringify(name)}))`
const zacke = 'Krone (Kopie) zwei'

// loeschfrage ist die Gruppe der Rückfrage — nicht die erste `group` der Seite.
// **`textContent` und nicht `innerText`**: seit #2056 trägt auch die
// Zeichenfläche `role="group"` — ein `<svg>`, und ein SVG-Element hat kein
// `innerText`. Bis zum 2026-10-05 warf die Suche daran.
const loeschfrage = "[...document.querySelectorAll('[role=group]')].find((g) => (g.textContent ?? '').includes('aus der Bibliothek löschen?'))"
const namensfeld = "[...document.querySelectorAll('label')].find((l) => l.innerText.startsWith('Name für das Setsymbol'))?.querySelector('input')"

export async function zumSymbolfach(p, set) {
  // **Erst die Schublade zu**: ihr Schleier fängt jeden Klick daneben.
  if (await p.seite("return document.body.innerText.includes('Eigenschaften des Sets')")) {
    // **Über das ×**: Escape erreichte die Schublade am Fenster nicht.
    await p.klickeAuf("[...document.querySelectorAll('button[aria-label]')].find((b) => b.checkVisibility() && b.getAttribute('aria-label').startsWith('Schließen'))", 'das × der Schublade')
    await p.warte('die Schublade ist zu', () => p.seite("return !document.body.innerText.includes('Eigenschaften des Sets')"))
  }
  await p.taste('2', 'befehl')
  // **Ist schon ein Set offen, führt der Pfad zur Liste zurück** — ⌘2 zeigt
  // dann das offene Set und nicht die Liste.
  const zurListe = "[...document.querySelectorAll('nav ol a, nav ol button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Gestalten')"
  if (await p.seite(`return !!(${zurListe})`)) await p.klickeAuf(zurListe, 'der Pfad zu »Gestalten«')
  await p.wartetext(set, 20000)
  await p.klickeText(set)
  await p.warte(`${set} ist offen`, () => p.seite(`return (document.querySelector('nav ol')?.innerText ?? '').includes(${JSON.stringify(set)})`), 20000)
  await p.klickeText('Eigenschaften')
  await p.klickeText('Symbol')
  await p.wartetext('Symbolbibliothek', 15000)
  // **Die Bibliothek steht seit UXR-07 eingeklappt** (#1915) — aufklappen,
  // wenn sie es noch ist: eine Seite, die stehen blieb, behält sie offen.
  const zu = "[...document.querySelectorAll('button[aria-expanded=false]')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Symbolbibliothek')"
  if (await p.seite(`return !!(${zu})`)) await p.klickeAuf(zu, 'die Symbolbibliothek aufklappen')
  // **Auf die Wirkung gewartet, nicht auf den Klick** (§3) — bis zum
  // 2026-10-05 ging es gleich weiter, und in `dom02ak02` stand die
  // Bibliothek danach mal offen und mal zu.
  await p.warte('die Symbolbibliothek ist offen', () => p.seite("return [...document.querySelectorAll('button[aria-expanded=true]')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Symbolbibliothek')"), 10000)
}

const namen = (p, ort) => JSON.parse(p.mse(ort, 'symbolbibliothek', '--json')).map((s) => `${s.name}:${s.verwendet.join('+')}`).join(', ')

export default async function dom02(p) {
  // **Ein eigenes Zuhause**: die eingesetzten Kopien liegen neben dem
  // Bestand, und die Probe gibt Dateien aus und liest sie ein.
  const ort = p.ort({ zuhause: true })
  p.mse(ort, 'set-neu', 'Kronenset', 'probe', 'probe-schlicht')
  p.mse(ort, 'set-neu', 'Leerset', 'probe', 'probe-schlicht')
  const ausgabe = join(ort.ort, 'zacke-aus.mse-symbol')

  await p.starte({ ort })

  p.notiz('AK-01 — das Setsymbol zeichnen und in die Bibliothek aufnehmen')
  await zumSymbolfach(p, 'Kronenset')
  await p.wartetext('Noch keine Symbole', 10000)
  await p.bild('leer')
  await p.klickeText('Kreis')
  await p.wartetext('Kreis — 4 Punkte')
  await p.klickeText('⤓ Ins Set legen')
  await p.wartetext('abgelegt. Beim Sichern des Sets', 10000)
  await p.klickeAuf(namensfeld, 'das Namensfeld')
  await p.tippeInsFeld('Krone')
  await p.klickeAuf(knopf('Setsymbol aufnehmen'), 'Setsymbol aufnehmen')
  await p.wartetext('Krone steht jetzt in der Bibliothek.', 10000)
  await p.warte('der Eintrag nennt das Set', () => p.seite(`return (${eintrag('Krone')})?.innerText.includes('Getragen von Kronenset') ?? false`))
  p.pruefe('der Eintrag hat ein Vorschaubild', await p.seite(`return !!(${eintrag('Krone')})?.querySelector('img[src^="data:image/png"]')`))
  p.pruefe('die Symbolschriften stehen getrennt daneben', await p.seite("return document.body.innerText.includes('Symbolschriften')"))
  await p.bild('aufgenommen')

  p.notiz('AK-01 — duplizieren und umbenennen, ein doppelter Name am Feld')
  await p.klickeAuf(knopfIn(eintrag('Krone'), 'Duplizieren'), 'Duplizieren an Krone')
  await p.warte('die Kopie steht da', () => p.seite(`return !!(${eintrag('Krone (Kopie)')})`))
  await p.klickeAuf(knopfIn(eintrag('Krone (Kopie)'), 'Umbenennen'), 'Umbenennen an der Kopie')
  await p.klickeAuf(`(${eintrag('Krone (Kopie)')})?.querySelector('input')`, 'das Feld für den neuen Namen')
  // **Angehängt und nicht ersetzt**: eine Rücktaste auf eine Markierung
  // erreichte das Feld am Fenster nur in einem von drei Läufen. Die
  // Schreibmarke steht dafür am Ende.
  await p.seite(`const f = (${eintrag('Krone (Kopie)')}).querySelector('input'); f.focus(); f.setSelectionRange(f.value.length, f.value.length); return true`)
  await p.tippeInsFeld(' zwei')
  await p.klickeAuf(knopfIn(eintrag('Krone (Kopie)'), 'Namen übernehmen'), 'Namen übernehmen')
  await p.warte('die Kopie heißt anders', () => p.seite(`return !!(${eintrag(zacke)})`))
  await p.klickeAuf(namensfeld, 'das Namensfeld')
  await p.tippeInsFeld('krone')
  await p.klickeAuf(knopf('Setsymbol aufnehmen'), 'Setsymbol aufnehmen')
  await p.warte('der doppelte Name steht am Feld', () => p.seite(`return (${namensfeld})?.getAttribute('aria-invalid') === 'true'`))
  await p.bild('doppelter-name')

  p.notiz('Einsetzen — das zweite Set bekommt eine Kopie, und der Editor zeigt sie')
  await zumSymbolfach(p, 'Leerset')
  await p.klickeAuf(knopfIn(eintrag('Krone'), 'In dieses Set einsetzen'), 'Einsetzen an Krone')
  await p.wartetext('Krone ist jetzt das Symbol dieses Sets.', 10000)
  await p.wartetext('Kreis — 4 Punkte', 10000)
  // **In der Reihenfolge der Setliste** — zuletzt geöffnet zuerst.
  await p.warte('der Eintrag nennt beide Sets', () => p.seite(`return (${eintrag('Krone')})?.innerText.includes('Getragen von Leerset, Kronenset') ?? false`))
  p.pruefe('die Kommandozeile sieht dieselben Träger', namen(p, ort).includes('Krone:Leerset+Kronenset'), namen(p, ort))
  await p.bild('eingesetzt')

  p.notiz('AK-02 — Löschen nennt die Folgen, und Abbrechen löscht nicht')
  await p.klickeAuf(knopfIn(eintrag('Krone'), 'Löschen'), 'Löschen an Krone')
  await p.wartetext('Krone aus der Bibliothek löschen?', 10000)
  const frage = await p.seite(`return (${loeschfrage})?.innerText ?? ''`)
  p.pruefe('die Frage nennt die Sets und dass sie ihre Kopie behalten', frage.includes('Leerset, Kronenset') && frage.includes('behalten ihre Kopie'), frage)
  await p.bild('loeschfrage')
  await p.klickeAuf(`[...(${loeschfrage}).querySelectorAll('button')].find((b) => b.innerText.trim() === 'Abbrechen')`, 'Abbrechen')
  await p.warte('die Frage ist zu', () => p.seite("return !document.body.innerText.includes('Krone aus der Bibliothek löschen?')"))
  p.pruefe('Krone steht noch da', await p.seite(`return !!(${eintrag('Krone')})`))
  await p.klickeAuf(knopfIn(eintrag('Krone'), 'Löschen'), 'Löschen an Krone')
  await p.klickeAuf(`[...(${loeschfrage})?.querySelectorAll('button') ?? []].find((b) => b.innerText.trim().endsWith('Löschen'))`, 'Löschen bestätigen')
  await p.warte('Krone ist fort', () => p.seite(`return !(${eintrag('Krone')})`))
  p.pruefe('Leerset trägt sein Symbol weiter', await p.seite("return document.body.innerText.includes('Kreis — 4 Punkte')"))

  p.notiz('AK-01 — ausgeben und einlesen über den Dialog von macOS')
  await p.klickeAuf(knopfIn(eintrag(zacke), 'Ausgeben …'), 'Ausgeben an der Kopie')
  await p.dialogZeigtSich('sichern', '»Ausgeben …«')
  await p.dialogantwort(ausgabe)
  await p.klickeAuf(knopfIn(eintrag(zacke), 'Ausgeben …'), 'Ausgeben an der Kopie')
  await p.warte('die Datei liegt da', () => existsSync(ausgabe), 10000)
  await p.klickeText('Symboldatei einlesen …')
  await p.dialogZeigtSich('oeffnen', '»Symboldatei einlesen …«')
  await p.dialogantwort(ausgabe)
  await p.klickeText('Symboldatei einlesen …')
  await p.wartetext('zacke-aus steht jetzt in der Bibliothek.', 10000)
  p.pruefe('die Kommandozeile sieht die eingelesene', namen(p, ort).includes('zacke-aus:'), namen(p, ort))
  await p.bild('eingelesen')

  await p.beende()
}
