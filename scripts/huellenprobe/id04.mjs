// **Kontakte aus Apple Kontakte, am nativen Fenster** — ID-04
// (#1703),
// Vermerk 0137:
//
// - AK-03: der Dialog von macOS geht auf, und ein Abbruch ändert nichts,
// - AK-01: die Vorschau zeigt jede Karte; übernommen wird nur die Auswahl —
//   hier Ada ohne ihre Nummer,
// - AK-02: ein zweites Einlesen derselben Datei legt nichts doppelt an: Ada
//   steht als „schon da … ersetzt ihn“, Bo als „schon da, unverändert“.
//
// Die Datei ist selbst geschrieben, in der Gestalt, die Kontakte ausgibt.
import { mkdtempSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

const karten = [
  'BEGIN:VCARD', 'VERSION:3.0', 'PRODID:-//Apple Inc.//macOS//EN', 'N:Lovelace;Ada;;;', 'FN:Ada Lovelace',
  'item1.EMAIL;type=INTERNET;type=pref:ada@example.org', 'TEL;type=CELL;type=VOICE:+49 170 0000000',
  'UID:probe-ada', 'END:VCARD',
  'BEGIN:VCARD', 'VERSION:3.0', 'FN:Bo', 'UID:probe-bo', 'END:VCARD',
].join('\r\n') + '\r\n'

const region = "[...document.querySelectorAll('section')].find((s) => s.checkVisibility() && (s.getAttribute('aria-labelledby') ?? '') !== '' && s.innerText.includes('Was in der Datei steht'))"

export default async function id04(p) {
  const ort = p.ort()
  const datei = join(mkdtempSync(join(tmpdir(), 'id04-')), 'auswahl.vcf')
  writeFileSync(datei, karten)

  await p.starte({ ort })
  await p.taste('5', 'befehl')
  // **Die Kontakte stehen seit UXR-15 in der Schublade hinter „Neuer Tausch
  // …"** (#1923), neben der Wahl des Gegenübers.
  await p.klickeText('Neuer Tausch …')
  await p.wartetext('Kontakte aus Apple Kontakte', 20000)
  await p.klickeText('Kontakte aus Apple Kontakte')
  await p.wartetext('vCard wählen', 10000)

  p.notiz('AK-03 — der Dialog von macOS, und ein Abbruch ändert nichts')
  await p.klickeText('vCard wählen …')
  await p.dialogZeigtSich('oeffnen', '»vCard wählen …«')
  p.pruefe('nach dem Abbruch steht keine Vorschau da', !(await p.seite(`return !!(${region})`)))

  p.notiz('AK-01 — die Vorschau, und übernommen wird nur die Auswahl')
  await p.dialogantwort(datei)
  await p.klickeText('vCard wählen …')
  await p.warte('die Vorschau steht da', () => p.seite(`return !!(${region})`), 15000)
  const vorschau = await p.seite(`return (${region}).innerText`)
  p.pruefe('beide Karten stehen in der Vorschau', vorschau.includes('Ada Lovelace') && vorschau.includes('Bo'), vorschau.replace(/\s+/g, ' ').slice(0, 160))
  await p.bild('vorschau')
  // **Die Wege einer Karte stehen seit UXR-16 hinter „Wege wählen"** (#1924):
  // gelesen werden sie immer, einzeln gewählt auf Wunsch.
  await p.klickeAuf(`[...(${region}).querySelectorAll('button')].find((b) => b.innerText.trim() === 'Wege wählen')`, 'Wege wählen bei Ada')
  await p.klickeAuf(`[...(${region}).querySelectorAll('label')].find((l) => l.innerText.includes('+49 170'))`, 'die Nummer von Ada')
  await p.klickeText('2 übernehmen')
  await p.wartetext('2 angelegt, 0 ersetzt.', 15000)
  const liste = await p.seite("return document.body.innerText")
  p.pruefe('Ada steht mit ihrer Adresse da', liste.includes('E-Mail: ada@example.org'))
  p.pruefe('und ohne ihre Nummer', !liste.includes('Telefon: +49 170 0000000'))
  await p.bild('uebernommen')

  p.notiz('AK-02 — dieselbe Datei noch einmal legt nichts doppelt an')
  await p.dialogantwort(datei)
  await p.klickeText('vCard wählen …')
  await p.warte('die Vorschau steht wieder da', () => p.seite(`return !!(${region})`), 15000)
  const zweite = await p.seite(`return (${region}).innerText`)
  p.pruefe('Ada ist schon da und würde ersetzt', zweite.includes('schon da als Ada Lovelace'), zweite.replace(/\s+/g, ' ').slice(0, 200))
  // **Unverändert steht als eine Zeile da** (UXR-16) und nicht je Karte.
  p.pruefe('Bo ist schon da und unverändert', zweite.includes('Schon da und unverändert: 1'), zweite.replace(/\s+/g, ' ').slice(0, 200))
  await p.bild('zweites-einlesen')

  p.pruefe('⌘Q beendet', await p.beende())
}
