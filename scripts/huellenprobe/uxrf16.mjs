// **Die Marke „dran" in der Partienliste, am Fenster** — UXR-F16
// (#2031).
//
// Eigen ist ein Sitz, der an das gezeigte Profil gebunden ist; die Bindung
// bietet das Startformular nur an, wenn es mehr als ein Profil gibt. Dieses
// Drehbuch legt deshalb ein zweites Profil an, beginnt am Fenster eine Partie
// mit dem ersten Sitz am gezeigten Profil und prüft in der Liste:
//
// 1. vor dem Beginn ist der eigene Sitz dran (die Starthand steht aus),
// 2. nach „Partie schließen" steht die Marke noch da — die Liste liest sich
//    beim Verlassen des Tisches neu.
//
// Dazu ein Bild der Liste mit Marke — und derselbe Weg mit **einem** Profil
// ohne jede Wahl (Vermerk 0164). **Was es nicht prüft**: den Wechsel der
// Priorität zum anderen Sitz — den hält `internal/service/partiedran_test.go`.
import { zurSeite } from './uirevue.mjs'

const zeilenwerte = (ausgabe) => ausgabe.split('\n').filter((z) => z.trim() !== '')

const marke = "[...document.querySelectorAll('main *')].some((e) => e.children.length === 0 && e.checkVisibility() && e.textContent.trim() === 'dran')"
const profilwahl = "[...document.querySelectorAll('select')].find((s) => s.checkVisibility() && [...(s.labels ?? [])].some((l) => l.innerText.trim().startsWith('Profil')))"

// namenUndDecks füllt beide Sitze — ohne sie bleibt „Partie anfangen" gesperrt.
async function namenUndDecks(p) {
  for (const [stelle, name] of [[1, 'Ada'], [2, 'Bo']]) {
    await p.seite(`[...document.querySelectorAll('input')].find((i) => i.checkVisibility() && (i.getAttribute('aria-label') === 'Name am Tisch ${stelle}' || [...(i.labels ?? [])].some((l) => l.innerText.trim() === 'Name am Tisch ${stelle}'))).focus(); return true`)
    await p.tippeInsFeld(name)
  }
  await p.seite("for (const s of [...document.querySelectorAll('select')].filter((s) => s.checkVisibility() && [...(s.labels ?? [])].some((l) => l.innerText.trim().startsWith('Deck')))) { const setze = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set; setze.call(s, s.options[1].value); s.dispatchEvent(new Event('change', { bubbles: true })) } return true")
}

export default async function uxrf16(p) {
  const ort = p.ort()
  p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Probe-Deck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')

  // **Ein zweites Profil**, damit die Wahl am Sitz erscheint.
  const nutzer = zeilenwerte(p.mse(ort, 'nutzer')).find((z) => !z.startsWith('\t'))?.split('\t')[0] ?? ''
  p.pruefe('es gibt einen Nutzer', nutzer !== '', nutzer)
  p.mse(ort, 'profil-neu', nutzer, 'Zweitprofil')
  const profile = zeilenwerte(p.mse(ort, 'nutzer')).filter((z) => z.startsWith('\t')).map((z) => z.trim().split('\t'))
  p.pruefe('zwei Profile', profile.length === 2, JSON.stringify(profile))
  const gezeigt = profile[0][0]

  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', '1280')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', '800')
  await p.starte({ ort })

  await zurSeite(p, 'Spielen')
  await p.wartetext('Eine Partie anfangen', 20000)
  await p.warte('die Wahl des Profils steht', () => p.seite(`return !!${profilwahl}`), 20000)
  await p.seite(`const s = ${profilwahl}; const setze = Object.getOwnPropertyDescriptor(HTMLSelectElement.prototype, 'value').set; setze.call(s, ${JSON.stringify(gezeigt)}); s.dispatchEvent(new Event('change', { bubbles: true })); return true`)
  p.pruefe('der erste Sitz ist an das gezeigte Profil gebunden', (await p.seite(`return ${profilwahl}.value`)) === gezeigt)
  await namenUndDecks(p)
  await p.bild('anfang-mit-profil')

  await p.klickeText('Partie anfangen')
  await p.warte('der Tisch steht', () => p.seite("return !!document.querySelector('canvas') && document.querySelector('canvas').checkVisibility()"), 20000)
  await new Promise((los) => setTimeout(los, 1000))

  await p.klickeText('Partie schließen')
  await p.wartetext('Weiterspielen', 20000)
  await p.warte('die Marke steht in der Liste', () => p.seite(`return ${marke}`), 20000)
  p.pruefe('die Partienliste trägt die Marke „dran"', true)
  await p.bild('liste-dran')

  // **Und die Kommandozeile sagt dasselbe** (§16).
  const liste = p.mse(ort, 'partien')
  p.pruefe('die Kommandozeile nennt den eigenen Sitz und dran', /dein Sitz: .*dran/.test(liste), liste)

  p.pruefe('⌘Q beendet die Anwendung', await p.beende())

  // **Und mit einem einzigen Profil, ohne jede Wahl** — UXR-F21
  // (#2041),
  // Vermerk 0164: kein Sitz ist gebunden, also sitzt das gezeigte Profil am
  // Sitz des Wirts. Bis dahin erschien die Marke für niemanden mit einem
  // Profil — das Formular bietet die Wahl dann gar nicht an.
  p.notiz('Ein Profil, keine Wahl — und trotzdem dran')
  const allein = p.ort()
  p.mse(allein, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(allein, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(allein, 'deck', 'neu', 'Probe-Deck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(allein, 'deck', 'legen', '1', `--karte=${karte}`, '8')
  p.pruefe('es gibt genau ein Profil', zeilenwerte(p.mse(allein, 'nutzer')).filter((z) => z.startsWith('\t')).length === 1)
  p.mse(allein, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  p.mse(allein, 'einstellungen', 'setzen', 'fensterbreite', '1280')
  p.mse(allein, 'einstellungen', 'setzen', 'fensterhoehe', '800')
  await p.starte({ ort: allein })

  await zurSeite(p, 'Spielen')
  await p.wartetext('Eine Partie anfangen', 20000)
  p.pruefe('das Formular bietet keine Profilwahl an', !(await p.seite(`return !!${profilwahl}`)))
  await namenUndDecks(p)
  await p.klickeText('Partie anfangen')
  await p.warte('der Tisch steht', () => p.seite("return !!document.querySelector('canvas') && document.querySelector('canvas').checkVisibility()"), 20000)
  await new Promise((los) => setTimeout(los, 1000))
  await p.klickeText('Partie schließen')
  await p.wartetext('Weiterspielen', 20000)
  await p.warte('die Marke steht in der Liste', () => p.seite(`return ${marke}`), 20000)
  p.pruefe('mit einem Profil trägt die Partienliste die Marke „dran"', true)
  await p.bild('liste-dran-ein-profil')

  const alleinListe = p.mse(allein, 'partien')
  p.pruefe('die Kommandozeile nennt den Sitz des Wirts als den eigenen', /dein Sitz: .*dran/.test(alleinListe), alleinListe)

  p.pruefe('⌘Q beendet die Anwendung (ein Profil)', await p.beende())
}
