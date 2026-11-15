// **Spielen mit einem Bestand, wie ihn ein Mensch hat** — UXR-13
// (#1921).
//
// `uirevue` misst mit dem Probebestand, und der hat weder ein Deck noch eine
// Partie: *Spielen* zeigt dort einen Leerzustand und darunter, was sonst
// selten gebraucht wird. Dieses Drehbuch legt ein Deck und acht Partien an —
// zu zweit, zu dritt, zu viert, eine allein; zwei angehalten, zwei beendet —
// und misst die Seite mit **derselben** Messung wie `uirevue` bei 1280 × 800
// und 1024 × 700, dazu die beiden Schubladen (Beitritt und Leihe). Je Größe
// ein Bild von Seite und Schubladen.
//
// **Was es nicht tut**: urteilen, ob die Anordnung dem Artboard *Eine Partie
// anfangen* gleicht. Das sieht ein Mensch an den Bildern.
import { messung, zurSeite } from './uirevue.mjs'

const groessen = [
  { name: 'normal', breite: 1280, hoehe: 800 },
  { name: 'eng', breite: 1024, hoehe: 700 },
]

// Die Sitze der acht Partien — `ada` sitzt an jeder, damit ihr Stand sich
// über sie setzen lässt.
const partien = [
  ['Ada', 'Bo'],
  ['Ada', 'Cleo', 'Bo'],
  ['Ada', 'Dario', 'Cleo', 'Bo'],
  ['Ada', 'Dario'],
  ['Bo', 'Ada'],
  ['Ada', 'Cleo'],
  ['Dario', 'Ada'],
  ['Ada', 'Bo'],
]

// Die Schubladen der Seite, über die Aufschrift ihres Knopfs im Kopf.
const schubladen = [
  { knopf: 'Bei jemandem mitspielen …', name: 'beitritt' },
  { knopf: 'Verleihen oder leihen …', name: 'leihe' },
]

const zeile = (groesse, ziel, z) =>
  `| ${groesse} | ${ziel} | ${String(z.bedien)} (${String(z.bedienImBild)}) | ${String(z.haupt)} (${String(z.hauptJeRegion)}) | ${String(z.felder)} (${String(z.felderImBild)}) | ${String(z.ueberschriften)} | ${String(z.rolltiefe)} | ${z.waagerecht ? 'ja' : 'nein'} |`

export default async function uxr13(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'deck', 'neu', 'Probe-Deck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(ort, 'deck', 'legen', '1', `--karte=${karte}`, '8')
  for (const sitze of partien) {
    p.mse(ort, 'partie-anlegen', 'probe', ...sitze.map((name) => `${name.toLowerCase()}:${name}:1`))
  }
  // **Die jüngste steht vorn** (`mse partien`): angehalten und beendet
  // werden zwei aus der Mitte, damit die Liste alle drei Stände zeigt.
  const kennungen = p.mse(ort, 'partien').split('\n').map((z) => z.split(' — ')[0].trim()).filter((k) => k !== '')
  p.zeige(`${String(kennungen.length)} Partien: ${kennungen.join(', ')}`)
  for (const [stelle, wohin] of [[2, 'pausieren'], [3, 'pausieren'], [5, 'beenden'], [6, 'beenden']]) {
    if (kennungen[stelle]) p.mse(ort, 'partie-stand', kennungen[stelle], 'ada', wohin)
  }

  const zeilen = []
  for (const groesse of groessen) {
    p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
    p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', String(groesse.breite))
    p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', String(groesse.hoehe))
    await p.starte({ ort })

    await zurSeite(p, 'Spielen')
    await p.wartetext('Weiterspielen', 20000)
    await new Promise((los) => setTimeout(los, 800))
    const seite = await p.seite(messung)
    zeilen.push(zeile(groesse.name, 'Spielen (mit Bestand)', seite))
    p.pruefe(`R1: höchstens ein Hauptknopf je Region (${groesse.name})`, seite.hauptJeRegion <= 1, JSON.stringify(seite))
    p.pruefe(`R5: die Seite rollt nicht waagerecht (${groesse.name})`, !seite.waagerecht)
    await p.bild(`spielen-${groesse.name}`)

    for (const schublade of schubladen) {
      await p.klickeText(schublade.knopf)
      await p.warte(`die Schublade ${schublade.name} steht`, () => p.seite("return [...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility())"))
      await new Promise((los) => setTimeout(los, 500))
      const wert = await p.seite(messung)
      zeilen.push(zeile(groesse.name, `Spielen › ${schublade.knopf} (Schublade)`, wert))
      p.pruefe(`R1 in der Schublade ${schublade.name} (${groesse.name})`, wert.schublade && wert.hauptJeRegion <= 1, JSON.stringify(wert))
      await p.bild(`spielen-${schublade.name}-${groesse.name}`)
      await p.taste('Escape')
      await p.warte('die Schublade ist zu', () => p.seite("return ![...document.querySelectorAll('[role=dialog]')].some((d) => d.checkVisibility())"))
    }

    p.pruefe(`⌘Q beendet die Anwendung (${groesse.name})`, await p.beende())
  }

  p.notiz('Die Tabelle')
  p.zeige('| Größe | Ziel | Bedienelemente (im Bild) | Hauptknöpfe (je Region höchstens) | Felder (im Bild) | Überschriften | Rolltiefe | waagerecht |')
  p.zeige('| --- | --- | ---: | ---: | ---: | ---: | ---: | --- |')
  for (const z of zeilen) p.zeige(z)
}
