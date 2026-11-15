// **Die Leiste an der Karte sagt dasselbe wie Klick und Menü** — P4
// (#2372),
// Befund C-10, gesehen am nativen Fenster (`tischhand09`): eine Handkarte,
// mit ⌘ gewählt, bot unter „Mit dieser Karte“ noch „Tappen“ und „Für Mana
// tappen“ an. Was an einem Objekt gilt, sagt der Kern (`handlungen`, P3) —
// die Kartengruppe der Leiste nimmt nur das; Zone, Sitz und Kampf bleiben,
// wie sie sind.
import { render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Partiebefehle } from './Partiebefehle'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { einLauf, einObjekt, einePartie, ruhenderLauf } from './partieprobe'

import type { Objekthandlung } from '../apipartie'

const namen = kataloge.de.seiten.handlungen.handlungsname

function zeige(handlungen: Objekthandlung[]) {
  const objekt = einObjekt({
    kennung: 'sitz1/baer/0',
    zone: 'hand',
    name: 'Bär',
    geheim: false,
  })
  render(
    <Sprachraum sprache="de">
      <Partiebefehle
        lauf={einLauf({ handlungen })}
        partie={einePartie({ objekte: [objekt] })}
        objekt={objekt}
        entwaehle={() => undefined}
        texturen={ruhenderLauf()}
      />
    </Sprachraum>,
  )
}

it('bietet an einer Handkarte nur an, was der Kern dort nennt', () => {
  zeige([
    {
      objekt: 'sitz1/baer/0',
      handlungen: ['objekt.wirken', 'objekt.ziehen'],
      primaer: 'objekt.wirken',
      zweit: '',
    },
  ])

  expect(
    screen.queryByRole('button', { name: namen['objekt.tappen'] }),
  ).toBeNull()
  expect(
    screen.queryByRole('button', { name: namen['objekt.mana'] }),
  ).toBeNull()
  expect(
    screen.getByRole('button', {
      name: new RegExp(`^${namen['objekt.wirken']}`),
    }),
  ).toBeInTheDocument()
})

// **Ohne Auskunft keine Kartenhandlung** — eine fremde Hand ist geheim, und
// der Kern nennt dort nichts.
it('bietet ohne Auskunft an der Karte nichts an', () => {
  zeige([])

  expect(
    screen.queryByRole('button', { name: namen['objekt.tappen'] }),
  ).toBeNull()
})
