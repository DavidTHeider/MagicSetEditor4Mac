// **Nach einem Tausch des Bestands bleibt die Anwendung benutzbar** — A02
// (#2216),
// Befund F-01 des Audits zu 0.8.0.
//
// # Was der Audit sah
//
// Eine Sicherung zurückspielen meldete Erfolg, und danach wies jede
// gebundene Methode mit einem „unerwarteten Fehler" ab — bis zum Neustart.
// `ausgangswerte` spielte zurück und endete dort; ob die Anwendung danach
// noch etwas tat, fragte kein Drehbuch.
//
// # Was hier geschieht
//
// Zurückspielen über *Einstellungen › Daten*, dann — **ohne Neustart** —
// ein Set öffnen, eine Karte anlegen und benennen, die Diagnose lesen;
// danach die Verschlüsselung ein- und wieder ausschalten (derselbe Tausch,
// ein anderer Anlass) und nach jedem Mal wieder schreiben. Jede Wirkung wird
// im Bestand nachgesehen, über `mse`, und nicht nur im Fenster.
import { existsSync } from 'node:fs'
import { join } from 'node:path'

const unerwartet = "return document.body.innerText.includes('unerwarteter Fehler')"

// zurDatengruppe öffnet *Einstellungen › Daten*.
async function zurDatengruppe(p) {
  await p.taste(',', 'befehl')
  const gruppe = "[...document.querySelectorAll('button, a, [role=tab]')].find((e) => e.checkVisibility() && e.innerText.replace(/\\s+/g, ' ').trim().startsWith('Daten'))"
  await p.warte('die Gruppe »Daten« steht zur Wahl', () => p.seite(`return !!(${gruppe})`), 20000)
  await p.seite(`(${gruppe}).scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(gruppe, 'die Gruppe »Daten«')
}

// knopf ist der sichtbare Knopf mit genau diesem Text — **ohne Zeichen
// davor** (»+ Karte anlegen«, siehe `ausgangswerte`).
const knopf = (text) => `[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.replace(/\\s+/g, ' ').trim().replace(/^[^\\p{L}\\p{N}]+/u, '') === ${JSON.stringify(text)})`

async function klicke(p, text) {
  await p.warte(`»${text}« steht da`, () => p.seite(`return !!(${knopf(text)}) && !(${knopf(text)}).disabled`), 20000)
  await p.seite(`(${knopf(text)}).scrollIntoView({ block: 'center' }); return true`)
  await p.klickeAuf(knopf(text), `»${text}«`)
}

// karteAnlegen öffnet das Set und legt eine Karte mit diesem Namen an — und
// wartet, bis sie **im Bestand** steht.
async function karteAnlegen(p, ort, name) {
  await p.taste('2', 'befehl')
  await p.wartetext('Werkstatt-Testkarten', 20000)
  await p.klickeText('Werkstatt-Testkarten')
  await klicke(p, 'Karte anlegen')
  const namensfeld = "[...document.querySelectorAll('input')].find((x) => x.checkVisibility() && x.getAttribute('aria-label') === 'name')"
  await p.warte('das Namensfeld der neuen Karte steht', () => p.seite(`return (${namensfeld})?.value === ''`))
  await p.klickeAuf(namensfeld, 'das Namensfeld')
  await p.seite(`(${namensfeld}).focus(); return true`)
  await p.tippeInsFeld(name)
  await p.taste('Tab')
  await p.warte(`»${name}« steht im Bestand`, () => p.mse(ort, 'karten', '1').includes(name), 20000)
  p.pruefe(`nach dem Tausch: »${name}« angelegt, ohne unerwarteten Fehler`, !(await p.seite(unerwartet)))
}

export default async function a02(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  const sicherung = join(ort.ort, 'a02.msesicherung')
  p.mse(ort, 'bestand-sichern', sicherung)
  p.pruefe('die Sicherung liegt da', existsSync(sicherung))
  await p.starte({ ort })

  p.notiz('Zurückspielen über Einstellungen › Daten')
  await zurDatengruppe(p)
  await p.wartetext('Sicherung zurückspielen …', 20000)
  await p.dialogantwort(sicherung)
  await klicke(p, 'Sicherung zurückspielen …')
  await p.wartetext('Diese Sicherung zurückspielen?', 20000)
  await klicke(p, 'Zurückspielen')
  await p.wartetext('Zurückgespielt.', 30000)
  await p.bild('zurueckgespielt')

  p.notiz('Ohne Neustart: Set öffnen, Karte schreiben, Diagnose lesen')
  await karteAnlegen(p, ort, 'Nach dem Zurückspielen')
  await p.bild('karte-nach-dem-zurueckspielen')
  await p.taste(',', 'befehl')
  await p.wartetext('Diagnose', 20000)
  await p.klickeText('Diagnose')
  // **Seit #2465 in einem Abschnitt** — erst aufklappen.
  await p.klickeText('Gewirkt seit dem Start')
  await p.wartetext('Laufzeiten und Verbrauch', 20000)
  p.pruefe('die Diagnose antwortet nach dem Zurückspielen', !(await p.seite(unerwartet)))
  await p.bild('diagnose-nach-dem-zurueckspielen')

  p.notiz('Derselbe Tausch aus einem anderen Anlass: die Verschlüsselung')
  await zurDatengruppe(p)
  await klicke(p, 'Verschlüsselung einschalten …')
  await p.wartetext('Den Bestand jetzt verschlüsseln?', 20000)
  await klicke(p, 'Verschlüsselung einschalten …')
  await p.warte('der Bestand ist verschlüsselt', () => p.seite("return document.body.innerText.includes('Der Bestand liegt verschlüsselt auf der Platte')"), 30000)
  await karteAnlegen(p, ort, 'Nach dem Verschlüsseln')

  await zurDatengruppe(p)
  await klicke(p, 'Verschlüsselung ausschalten …')
  await p.wartetext('Die Verschlüsselung wieder ausschalten?', 20000)
  await klicke(p, 'Verschlüsselung ausschalten …')
  await p.warte('der Bestand liegt im Klartext', () => p.seite("return document.body.innerText.includes('Der Bestand liegt im Klartext auf der Platte')"), 30000)
  await karteAnlegen(p, ort, 'Nach dem Entschlüsseln')
  await p.bild('nach-dem-entschluesseln')

  p.pruefe('⌘Q beendet', await p.beende())
}
