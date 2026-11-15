// **Teilen am nativen Fenster** — NET-04
// (#1708):
//
// - eine Einladung entsteht, und „Teilen …" zeigt vorher Datei, Größe und
//   was mitgeht (AK-02),
// - die Kopie liegt im Teilordner **dieser Probe** — mit eigenem Zuhause,
//   damit nichts im echten `Caches` landet —, nur für den Eigentümer,
// - „Mail, Nachrichten, AirDrop …" öffnet den Teilen-Dialog von macOS;
//   er wird mit Escape geschlossen, ohne dass etwas hinausgeht,
// - „Fertig" räumt die Kopie weg (AK-04).
import { existsSync, readdirSync, statSync } from 'node:fs'
import { join } from 'node:path'

export default async function net04(p) {
  const anna = p.ort({ zuhause: true })
  const teilordner = join(anna.heim, 'Library', 'Caches', 'MagicSetEditor4Mac', 'teilen')

  p.notiz('1 — ein Profil mit Identität')
  p.mse(anna, 'einstellungen', 'setzen', 'sprache', 'de')
  p.mse(anna, 'peer-anlegen')

  await p.starte({ ort: anna })
  await p.taste('5', 'befehl')

  p.notiz('2 — eine Einladung entsteht')
  // **Seit UXR-15 in der Schublade hinter „Neuer Tausch …"** (#1923).
  await p.klickeText('Neuer Tausch …')
  await p.klickeText('Gegenüber aufnehmen oder bestätigen')
  await p.klickeText('Meinen Code zeigen')
  await p.klickeAuf("document.querySelector('[data-einladungstausch] input')", 'das Namensfeld')
  await p.tippeInsFeld('Anna')
  await p.klickeText('Code erzeugen')
  await p.warte('der Code steht da', () =>
    p.seite("return (document.querySelector('[data-einladungstausch] textarea[readonly]')?.value ?? '').startsWith('mse4mac-kontakt:1:')"))

  p.notiz('3 — Teilen zeigt vorher, was mitgeht')
  await p.klickeText('Teilen …')
  await p.wartetext('Einladung.txt')
  p.pruefe('die Adresse steht als Mitgehendes da', await p.seite("return document.body.innerText.includes('die Adresse, unter der du erreichbar bist')"))
  p.pruefe('der Satz über die Kopie steht da', await p.seite("return document.body.innerText.includes('Geteilt wird eine Kopie.')"))
  await p.bild('vorschau')

  const teilungen = existsSync(teilordner) ? readdirSync(teilordner) : []
  p.pruefe('genau eine Teilung im Teilordner der Probe', teilungen.length === 1, teilungen.join(', '))
  const datei = join(teilordner, teilungen[0] ?? '-', 'Einladung.txt')
  p.pruefe('die Kopie gehört nur dem Eigentümer', existsSync(datei) && (statSync(datei).mode & 0o077) === 0)

  p.notiz('4 — der Teilen-Dialog von macOS')
  await p.klickeText('Mail, Nachrichten, AirDrop …')
  await new Promise((fertig) => setTimeout(fertig, 1200))
  const lage = await p.fenster()
  p.zeige(`vorn: ${lage.vorn}`)
  await p.bild('dialog')
  await p.taste('Escape')

  // **Kam das Escape in der Webansicht an, ist die Schublade zu** — sie
  // behält, was in ihr stand; ein zweites Öffnen zeigt es wieder.
  if (!(await p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Fertig')"))) {
    await p.klickeText('Neuer Tausch …')
  }

  p.notiz('5 — Fertig räumt die Kopie weg')
  await p.klickeText('Fertig')
  await p.wartetext('Teilen …')
  await p.warte('die Teilung ist fort', () => Promise.resolve(!existsSync(join(teilordner, teilungen[0] ?? '-'))))
  await p.bild('fertig')
}
