// **Was ein Spieler an sich selbst sieht** — Werkstück M3 der Regelmaschine
// (#1605).
//
// # Der Befund, für den es diese Leiste gibt
//
// **Gemessen am 2026-09-17**: `leben` und `vergiftung` stehen seit G9 am
// Modell, seit M17 setzt eine zustandsbasierte Aktion sie durch — und im
// Fenster kam **keine der beiden Zahlen vor**. Die einzige Stelle, an der
// „Leben" stand, war der Grund **nach** dem Ausscheiden („stand bei null
// Leben oder darunter"). Wer spielte, sah die Zahl nie, auf die es ankommt.
//
// **Das ist die erste Dimension aus §10**: eine Regel, die greift, und
// niemand kann sehen, worauf.
import { render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Spielerleiste } from './Spielerleiste'
import { Sprachraum } from '../sprache/kontext'
import { einePartie, einSitz } from '../seiten/partieprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.spielen

function zeige(sitze: ReturnType<typeof einSitz>[], fuer = 'sitz-1') {
  render(
    <Sprachraum sprache="de">
      <Spielerleiste partie={einePartie({ sitze, fuer })} />
    </Sprachraum>,
  )
}

// **Der Lebensstand steht da, und zwar für jeden am Tisch.**
it('zeigt jeden Sitz mit seinem Lebensstand', () => {
  zeige([
    einSitz({ kennung: 'sitz-1', name: 'Anna', leben: 17 }),
    einSitz({ kennung: 'sitz-2', name: 'Bert', leben: 3 }),
  ])

  expect(screen.getByText('Anna')).toBeInTheDocument()
  expect(screen.getByText('17')).toBeInTheDocument()
  expect(screen.getByText('3')).toBeInTheDocument()
})

// **Vergiftungsmarken stehen nur da, wo es welche gibt** — CR 122.1f.
//
// **Eine Null ist keine Auskunft**: die meisten Partien haben nie eine
// Vergiftungsmarke, und eine Zeile „Gift: 0" an jedem Sitz wäre Lärm, in dem
// die eine Marke untergeht, auf die es ankommt.
it('zeigt Vergiftungsmarken nur, wo welche liegen', () => {
  zeige([
    einSitz({ kennung: 'sitz-1', name: 'Anna', vergiftung: 4 }),
    einSitz({ kennung: 'sitz-2', name: 'Bert', vergiftung: 0 }),
  ])

  const gift = screen.getAllByText(eigene.standGift, { exact: false })

  expect(gift).toHaveLength(1)
  expect(gift[0]).toHaveTextContent('4')
})

// **Und der Manavorrat ebenso** — CR 106.4.
it('zeigt den Manavorrat, solange einer dasteht', () => {
  zeige([
    einSitz({ kennung: 'sitz-1', name: 'Anna', manavorrat: 'RRG' }),
    einSitz({ kennung: 'sitz-2', name: 'Bert', manavorrat: '' }),
  ])

  const vorrat = screen.getAllByText(eigene.standMana, { exact: false })

  expect(vorrat).toHaveLength(1)
  expect(vorrat[0]).toHaveTextContent('RRG')
})

// **Wer gerade zusieht, erkennt sich selbst.**
//
// **Sonst wäre die Leiste bei zwei gleichen Namen unbrauchbar** — und gegen
// sich selbst zu spielen ist in diesem Programm der Normalfall (G8).
it('hebt den eigenen Sitz hervor', () => {
  zeige(
    [
      einSitz({ kennung: 'sitz-1', name: 'Anna' }),
      einSitz({ kennung: 'sitz-2', name: 'Anna' }),
    ],
    'sitz-2',
  )

  expect(screen.getByLabelText(eigene.standDu)).toHaveTextContent('Anna')
})
