// **Die Decks als Kacheln im Navigator** — #2019, Folge aus UXR-11
// (#1919).
//
// `uirevue` misst mit dem Probebestand, und der hat kein Deck — die
// Kachelansicht der Deckliste stand dort nie da. Aus dem Quelltext stand der
// Verdacht: Kacheln von mindestens 16rem in einem Navigator von 15rem. Dieses
// Drehbuch legt drei Decks an, schaltet auf Kacheln und misst R5 aus
// `docs/UIREVIEW.md` — rollt die Seite oder der Navigator waagerecht?
import { zurSeite } from './uirevue.mjs'

export default async function deckkacheln(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  for (const name of ['Erstes Deck', 'Zweites Deck', 'Drittes Deck']) {
    p.mse(ort, 'deck', 'neu', name, 'probe')
  }
  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', '1280')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', '800')
  await p.starte({ ort })

  await zurSeite(p, 'Decks')
  await p.wartetext('Erstes Deck', 20000)
  await p.klickeText('Kacheln')
  await p.warte('die Decks stehen als Kacheln', () => p.seite("return !!document.querySelector('ul[data-form=kacheln]')"))
  await new Promise((los) => setTimeout(los, 800))

  const mass = await p.seite(`
    const liste = document.querySelector('ul[data-form=kacheln]')
    const rollend = []
    for (let e = liste; e; e = e.parentElement) {
      if (e.scrollWidth > e.clientWidth + 1 && /(auto|scroll)/.test(getComputedStyle(e).overflowX)) rollend.push(e.tagName + '.' + [...e.classList].slice(0, 3).join('.'))
    }
    const r = liste.getBoundingClientRect()
    const eltern = liste.parentElement.getBoundingClientRect()
    return {
      seite: document.scrollingElement.scrollWidth > innerWidth + 1,
      rollend,
      ueberstand: Math.round(r.right - eltern.right),
      kachel: Math.round(liste.firstElementChild.getBoundingClientRect().width),
      platz: Math.round(eltern.width),
    }
  `)
  p.zeige(`Maß: ${JSON.stringify(mass)}`)
  await p.bild('kacheln')
  p.pruefe('R5: die Seite rollt nicht waagerecht', !mass.seite)
  p.pruefe('R5: kein Behälter der Kacheln rollt waagerecht', mass.rollend.length === 0, mass.rollend.join(', '))
  p.pruefe('die Kacheln ragen nicht über ihren Platz hinaus', mass.ueberstand <= 0, `${String(mass.ueberstand)} px, Kachel ${String(mass.kachel)} px in ${String(mass.platz)} px`)
  p.pruefe('⌘Q beendet', await p.beende())
}
