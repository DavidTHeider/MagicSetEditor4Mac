// **Ein Ziehen ist kein Klick** — P5
// (#2373),
// gefunden am nativen Fenster (`tischhand09`, Rufbuch: 27 Befehle statt 24):
// WebKit schickt nach einem Zug, der auf einem Kartenknopf begann, noch einen
// `click` an ihn. Seit P4 löst ein Klick die Primärhandlung aus — wer eine
// Handkarte auf den Friedhof zog, wirkte sie zugleich. **Ein Zug über die
// Schwelle verschluckt den Klick danach**; ein Klick mit etwas Zittern
// bleibt einer.
import { fireEvent, render, screen } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { eingepasst, nachPunkten } from '../tisch/kamera'
import { einObjekt, einePartie } from '../seiten/partieprobe'
import { Sprachraum } from '../sprache/kontext'
import { tischplan } from '../tisch/tischplan'
import type { Tischmalerei } from '../tisch/maler'
import { einMaler } from '../test/malerprobe'
import { Tischflaeche } from './Tischflaeche'

const plan = tischplan(
  einePartie({
    objekte: [
      einObjekt({
        kennung: 'a',
        zone: 'hand',
        name: 'Wald',
        eigner: 'sitz1',
        beherrscher: 'sitz1',
        geheim: false,
      }),
    ],
  }),
)

const amOrt = (x: number, y: number) => {
  const punkt = nachPunkten(
    eingepasst(plan, { breite: 800, hoehe: 600 }, 16),
    { x, y },
    { breite: 800, hoehe: 600 },
  )

  return { clientX: punkt.x, clientY: punkt.y }
}

// zeichneNeu tauscht den Plan, wie es eine Antwort des Kerns tut.
let zeichneNeu: (neu: ReturnType<typeof tischplan>) => void = () => undefined

async function zeige() {
  const aufKarte = vi.fn()
  const aufZug = vi.fn()
  const maler = einMaler({ breite: 800, hoehe: 600 })
  const baum = (welcher: ReturnType<typeof tischplan>) => (
    <Sprachraum sprache="de">
      <Tischflaeche
        plan={welcher}
        bild={() => undefined}
        malerei={() => Promise.resolve<Tischmalerei>({ art: 'da', maler })}
        aufKarte={aufKarte}
        aufZug={aufZug}
      />
    </Sprachraum>
  )
  const gezeichnet = render(baum(plan))
  zeichneNeu = (neu) => {
    gezeichnet.rerender(baum(neu))
  }
  const flaeche = await screen.findByRole('application')
  flaeche.getBoundingClientRect = () =>
    ({ left: 0, top: 0, width: 800, height: 600 }) as DOMRect
  const lage = plan.lagen.find((eine) => eine.kennung === 'a')
  const feld = plan.felder.find(
    (eines) => eines.zone === 'friedhof' && eines.sitz === 'sitz1',
  )
  if (!lage || !feld) throw new Error('Plan passt nicht zum Lauf')

  return {
    aufKarte,
    aufZug,
    flaeche,
    knopf: screen.getByRole('button', { name: /Wald/ }),
    karte: { x: lage.x + lage.breite / 2, y: lage.y + lage.hoehe / 2 },
    friedhof: { x: feld.x + feld.breite / 2, y: feld.y + feld.hoehe / 2 },
  }
}

it('verschluckt den Klick nach einem Zug', async () => {
  const { aufKarte, aufZug, flaeche, knopf, karte, friedhof } = await zeige()

  fireEvent.pointerDown(flaeche, { button: 0, ...amOrt(karte.x, karte.y) })
  fireEvent.pointerMove(flaeche, amOrt(friedhof.x, friedhof.y))
  fireEvent.pointerUp(flaeche, amOrt(friedhof.x, friedhof.y))
  fireEvent.click(knopf, { detail: 1 })

  expect(aufZug).toHaveBeenCalledTimes(1)
  expect(aufKarte).not.toHaveBeenCalled()

  // **Der nächste Druck zählt wieder.**
  fireEvent.pointerDown(flaeche, { button: 0, ...amOrt(karte.x, karte.y) })
  fireEvent.pointerUp(flaeche, amOrt(karte.x, karte.y))
  fireEvent.click(knopf, { detail: 1 })

  expect(aufKarte).toHaveBeenCalledTimes(1)
})

it('lässt einen Klick mit etwas Zittern einen Klick sein', async () => {
  const { aufKarte, flaeche, knopf, karte } = await zeige()
  const start = amOrt(karte.x, karte.y)

  fireEvent.pointerDown(flaeche, { button: 0, ...start })
  fireEvent.pointerMove(flaeche, {
    clientX: start.clientX + 2,
    clientY: start.clientY + 1,
  })
  fireEvent.pointerUp(flaeche, {
    clientX: start.clientX + 2,
    clientY: start.clientY + 1,
  })
  fireEvent.click(knopf, { detail: 1 })

  expect(aufKarte).toHaveBeenCalledTimes(1)
})

// **Ohne Bewegung kein Zug** — gefunden am nativen Fenster (`tischhand09`):
// ein Klick auf eine Handkarte, und zwischen Drücken und Loslassen kam das
// Bild, in dem dieselbe Karte schon auf dem Schlachtfeld lag. Der Tisch sah
// die Karte im Schlachtfeld und den Zeiger über der Hand und schob sie
// zurück — ein bestätigter Landzug war weg.
it('stellt keinen Zug, wenn sich nur das Bild unter dem Zeiger geändert hat', async () => {
  const { aufZug, flaeche, karte } = await zeige()
  const stelle = amOrt(karte.x, karte.y)

  fireEvent.pointerDown(flaeche, { button: 0, ...stelle })
  zeichneNeu(
    tischplan(
      einePartie({
        objekte: [
          einObjekt({
            kennung: 'a',
            zone: 'tisch',
            name: 'Wald',
            eigner: 'sitz1',
            beherrscher: 'sitz1',
            geheim: false,
          }),
        ],
      }),
    ),
  )
  fireEvent.pointerUp(flaeche, stelle)

  expect(aufZug).not.toHaveBeenCalled()
})
