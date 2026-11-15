// **Die jüngste Vorschau gewinnt, und ein Ort gehört zu seinem Bild** —
// Befund 3 des Nachlaufs über Phase 6.9 (PLAY-05.AK-04,
// #1713).
//
// Zwei Abzüge in derselben Sicht konnten in umgekehrter Reihenfolge
// antworten, und die ältere Antwort ersetzte die jüngere. Und ein Ort, der
// für ein älteres Bild zurückkam, stand unter einem neueren.
import { act, render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { Tischabzugfeld } from './Tischabzugfeld'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'
import { einePartie } from '../seiten/partieprobe'
import { tischplan } from '../tisch/tischplan'
import { zweiSitze } from '../test/planprobe'

import type { Abzug } from '../tisch/tischabzug'

const eigene = kataloge.de.seiten.spielen

const plan = tischplan(einePartie(zweiSitze))

afterEach(() => {
  raeumeApi()
})

const bild = (inhalt: string): Abzug => ({
  bild: `data:image/png;base64,${inhalt}`,
  breite: 12,
  hoehe: 8,
})

async function erfuelle(tun: () => void) {
  await act(async () => {
    tun()
    await new Promise((weiter) => setTimeout(weiter, 0))
  })
}

it('lässt eine ältere Antwort die jüngere nicht ersetzen', async () => {
  const offen: ((abzug: Abzug) => void)[] = []
  const abzug = () =>
    new Promise<Abzug>((erfuellen) => {
      offen.push(erfuellen)
    })

  render(
    <Sprachraum sprache="de">
      <Tischabzugfeld plan={plan} abzug={abzug} sicht="sitz1" />
    </Sprachraum>,
  )
  const hand = nutzerereignis.setup()
  const erzeugen = () =>
    hand.click(screen.getByRole('button', { name: eigene.abzugErzeugen }))

  await erzeugen()
  await erzeugen()

  // **Die jüngere zuerst**, dann die ältere.
  await erfuelle(() => {
    offen[1]?.(bild('NEU'))
  })
  expect(await screen.findByAltText(eigene.abzugVorschau)).toHaveAttribute(
    'src',
    'data:image/png;base64,NEU',
  )

  await erfuelle(() => {
    offen[0]?.(bild('ALT'))
  })
  expect(screen.getByAltText(eigene.abzugVorschau)).toHaveAttribute(
    'src',
    'data:image/png;base64,NEU',
  )
})

it('schreibt einen Ort nicht unter ein neueres Bild', async () => {
  const gesichert: ((antwort: {
    pfad: string
    abgebrochen: boolean
    fehler: null
  }) => void)[] = []
  belegeApi({
    TischabzugSpeichern: () =>
      new Promise((erfuellen) => {
        gesichert.push(erfuellen)
      }),
  })
  const bilder = [bild('EINS'), bild('ZWEI')]
  const abzug = () => Promise.resolve(bilder.shift() ?? bild('LEER'))

  render(
    <Sprachraum sprache="de">
      <Tischabzugfeld plan={plan} abzug={abzug} sicht="sitz1" />
    </Sprachraum>,
  )
  const hand = nutzerereignis.setup()
  const erzeugen = () =>
    hand.click(screen.getByRole('button', { name: eigene.abzugErzeugen }))

  await erzeugen()
  await hand.click(
    await screen.findByRole('button', { name: eigene.abzugSichern }),
  )

  // **Ein neues Bild, bevor der Ort des ersten zurückkommt.**
  await erzeugen()
  await screen.findByRole('img', { name: eigene.abzugVorschau })
  expect(screen.getByAltText(eigene.abzugVorschau)).toHaveAttribute(
    'src',
    'data:image/png;base64,ZWEI',
  )

  await erfuelle(() => {
    gesichert[0]?.({ pfad: '/tmp/eins.png', abgebrochen: false, fehler: null })
  })

  expect(
    screen.queryByText(
      einsetzen(eigene.abzugGesichert, { pfad: '/tmp/eins.png' }),
    ),
  ).not.toBeInTheDocument()

  // **Und dass jemand zuhört**: der Ort des zweiten Bildes erscheint.
  await hand.click(screen.getByRole('button', { name: eigene.abzugSichern }))
  await erfuelle(() => {
    gesichert[1]?.({ pfad: '/tmp/zwei.png', abgebrochen: false, fehler: null })
  })
  expect(
    await screen.findByText(
      einsetzen(eigene.abzugGesichert, { pfad: '/tmp/zwei.png' }),
    ),
  ).toBeInTheDocument()
})
