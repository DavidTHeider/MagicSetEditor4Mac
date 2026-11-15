// **Während eines Sichtwechsels steht vom alten Sitz nichts da** — Befund 1
// des Nachlaufs über Phase 6.9 (PLAY-05.AK-02,
// #1713).
//
// Der Kopf bleibt — mit ihm wählt man die Sicht, auch noch einmal nach einem
// Fehlschlag. Was eine Sicht zeigt, geht: Tisch und Zonen, Zug, Spieler,
// Handlungen und Tischbild.
import { render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Partietisch } from './Partietisch'
import { Sprachraum } from '../sprache/kontext'
import { einLauf, einePartie, einObjekt } from './partieprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.spielen

const mitHand = einePartie({
  fuer: 'sitz1',
  objekte: [
    einObjekt({
      kennung: 'h1',
      zone: 'hand',
      name: 'Geheime Karte',
      eigner: 'sitz1',
      beherrscher: 'sitz1',
      geheim: false,
    }),
  ],
})

function zeige(wechselt: boolean, fehler?: string) {
  render(
    <Sprachraum sprache="de">
      <Partietisch
        lauf={einLauf({
          partie: mitHand,
          wechselt,
          ...(fehler === undefined ? {} : { fehler }),
        })}
        partie={mitHand}
      />
    </Sprachraum>,
  )
}

it('zeigt die Hand, solange die Sicht steht', () => {
  zeige(false)

  expect(screen.getAllByText(/Geheime Karte/).length).toBeGreaterThan(0)
  expect(screen.queryByText(eigene.sichtLaedt)).toBeNull()
})

it('nimmt sie weg, sobald eine andere Sicht gewählt ist', () => {
  zeige(true)

  expect(screen.queryByText(/Geheime Karte/)).toBeNull()
  expect(screen.queryByRole('application')).toBeNull()
  expect(screen.queryByRole('button', { name: eigene.abzugTitel })).toBeNull()
  expect(screen.getByRole('status')).toHaveTextContent(eigene.sichtLaedt)

  // **Der Kopf bleibt** — mit ihm wählt man die Sicht.
  expect(screen.getByLabelText(eigene.sicht)).toBeInTheDocument()
})

it('sagt nach einem Fehlschlag, was schiefging, und zeigt die alte Sicht nicht', () => {
  zeige(true, 'Die Leitung ist weg.')

  expect(screen.getByRole('alert')).toHaveTextContent('Die Leitung ist weg.')
  expect(screen.queryByText(/Geheime Karte/)).toBeNull()
})
