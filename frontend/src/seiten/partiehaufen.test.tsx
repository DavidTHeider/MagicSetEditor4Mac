// **Die Geste, die einen Haufen legt** — CR 400.5,
// #1756,
// PLAY-04.AK-01 (#1712).
//
// # Wogegen hier geprüft wird
//
// **Dass der Weg vom Zeiger bis zum Befehl trägt.** Die Rechnung dahinter
// steht in `bausteine/tischziehenregeln.test.ts` und kommt ohne einen Baum
// aus; hier zieht ein Zeiger über eine echte Fläche, und am Ende steht ein
// Befehl, den der Kern kennt.
//
// **Eine eigene Datei**, damit `partietischdarstellung.test.tsx` unter der
// Zeilengrenze aus §2 bleibt. Der Schnitt folgt der Frage: dort steht, dass
// **eine** Darstellung dasteht, hier, was ein Zug darin bedeutet.
import { fireEvent, screen, waitFor } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { einLauf, einObjekt, einePartie } from './partieprobe'
import { eingepasst, nachPunkten } from '../tisch/kamera'
import { tischplan } from '../tisch/tischplan'
import { zeige } from './partietischprobe'

// zweiWaelder gibt zwei gleiche Länder auf dem Schlachtfeld.
//
// **Gleich in jedem gezeichneten Zustand** — sonst wiese der Kern den Haufen
// ab, und der Lauf prüfte die Ablehnung statt der Geste.
const zweiWaelder = (haufen = '') =>
  einePartie({
    objekte: [
      einObjekt({
        kennung: 'w1',
        zone: 'tisch',
        name: 'Wald',
        geheim: false,
        karteID: '7',
        haufen,
      }),
      einObjekt({
        kennung: 'w2',
        zone: 'tisch',
        name: 'Wald',
        geheim: false,
        karteID: '7',
        haufen,
      }),
    ],
  })

// ziehe fährt mit dem Zeiger von einer Lage auf einen Punkt.
//
// **Dieselbe Kamera wie der Haken** — und der hat ohne Renderer die Maße
// null: unter jsdom gibt es kein WebGL. Wer hier 800 × 600 annähme,
// rechnete gegen eine Kamera, die es nicht gibt.
function ziehe(
  partie: ReturnType<typeof einePartie>,
  von: string,
  nach: { x: number; y: number },
) {
  const plan = tischplan(partie)
  const lage = plan.lagen.find((eine) => eine.kennung === von)
  if (!lage) throw new Error(`keine Lage für ${von}`)

  const flaeche = screen.getByRole('application')
  flaeche.getBoundingClientRect = () =>
    ({ left: 0, top: 0, width: 800, height: 600 }) as DOMRect

  const ohneFlaeche = { breite: 0, hoehe: 0 }
  const amOrt = (x: number, y: number) => {
    const punkt = nachPunkten(
      eingepasst(plan, ohneFlaeche, 16),
      { x, y },
      ohneFlaeche,
    )

    return { clientX: punkt.x, clientY: punkt.y }
  }

  fireEvent.pointerDown(flaeche, {
    button: 0,
    ...amOrt(lage.x + lage.breite / 2, lage.y + lage.hoehe / 2),
  })
  fireEvent.pointerUp(flaeche, amOrt(nach.x, nach.y))
}

// mitteVon gibt die Mitte einer Lage im Plan.
function mitteVon(
  partie: ReturnType<typeof einePartie>,
  kennung: string,
): { x: number; y: number } {
  const lage = tischplan(partie).lagen.find((eine) => eine.kennung === kennung)
  if (!lage) throw new Error(`keine Lage für ${kennung}`)

  return { x: lage.x + lage.breite / 2, y: lage.y + lage.hoehe / 2 }
}

it('legt eine Karte auf die, auf der sie losgelassen wird', async () => {
  const befehlen = vi.fn()
  const partie = zweiWaelder()

  zeige(einLauf({ befehlen }), partie, false)
  ziehe(partie, 'w1', mitteVon(partie, 'w2'))

  await waitFor(() => {
    expect(befehlen).toHaveBeenCalledWith('objekt.haeufen', {
      objekt: 'w1',
      worauf: 'w2',
    })
  })
})

// **Die Gegenbewegung**: neben den Haufen gezogen heißt heraus damit.
//
// **Angefasst wird die oberste** — und das ist keine Annahme, sondern was am
// Griffpunkt liegt: die Karten eines Haufens überdecken einander, und
// `objektAn` gibt die, die obenauf liegt. Wer die untere wollte, zöge den
// Haufen erst auseinander.
it('nimmt die oberste Karte aus dem Haufen, wo sie daneben landet', async () => {
  const befehlen = vi.fn()
  const partie = zweiWaelder('w2')

  zeige(einLauf({ befehlen }), partie, false)

  // **Ein freier Punkt derselben Zone** — weit genug weg, dass dort keine
  // Karte liegt, und innerhalb des Schlachtfelds.
  const feld = tischplan(partie).felder.find(
    (eines) => eines.zone === 'tisch' && eines.sitz === 'sitz1',
  )
  if (!feld) throw new Error('kein Schlachtfeld')

  ziehe(partie, 'w1', {
    x: feld.x + feld.breite - 1,
    y: feld.y + feld.hoehe - 1,
  })

  await waitFor(() => {
    expect(befehlen).toHaveBeenCalledWith('objekt.enthaeufen', {
      objekt: 'w2',
    })
  })
})
