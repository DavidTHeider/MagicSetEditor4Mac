// **Eine Vorschau gehört zu der Sicht, aus der sie stammt** — PLAY-05.AK-02
// (#1713),
// Befund 1 des Zweitblicks über Phase 6.9.
//
// # Der Befund
//
// Das Ausgabefeld hielt sein Bild im Zustand, und der Tisch darüber ist nur
// an die Partie gebunden. Wer aus Adas Sicht eine Vorschau erzeugte und dann
// zur Zuschauersicht wechselte, sah weiter Adas Hand — und konnte sie als PNG
// oder PDF sichern. Eine Antwort, die erst nach dem Wechsel eintraf, setzte
// ihr Bild ebenso ein (§7, Regel 2).
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

import type { Abzug, Abzugsfehler } from '../tisch/tischabzug'

const eigene = kataloge.de.seiten.spielen

const plan = tischplan(einePartie(zweiSitze))

const einBild: Abzug = {
  bild: 'data:image/png;base64,AAA',
  breite: 12,
  hoehe: 8,
}

afterEach(() => {
  raeumeApi()
})

// erfuelle löst eine offene Antwort ein und lässt ihr eine Makroaufgabe, um
// anzukommen — die Wails-IPC antwortet auch nicht im selben Takt.
async function erfuelle(tun: () => void) {
  await act(async () => {
    tun()
    await new Promise((weiter) => setTimeout(weiter, 0))
  })
}

function feld(
  sicht: string,
  abzug: () => Promise<Abzug | { fehler: Abzugsfehler }>,
) {
  return (
    <Sprachraum sprache="de">
      <Tischabzugfeld plan={plan} abzug={abzug} sicht={sicht} />
    </Sprachraum>
  )
}

it('verwirft die Vorschau, wenn die Sicht wechselt', async () => {
  const abzug = () => Promise.resolve(einBild)
  const { rerender } = render(feld('ada', abzug))
  const hand = nutzerereignis.setup()

  await hand.click(screen.getByRole('button', { name: eigene.abzugErzeugen }))
  expect(await screen.findByAltText(eigene.abzugVorschau)).toBeInTheDocument()

  rerender(feld('', abzug))

  expect(screen.queryByAltText(eigene.abzugVorschau)).not.toBeInTheDocument()
  expect(
    screen.queryByRole('button', { name: eigene.abzugSichern }),
  ).not.toBeInTheDocument()
})

it('setzt eine späte Vorschau nicht in eine andere Sicht', async () => {
  const offen: ((abzug: Abzug) => void)[] = []
  const abzug = () =>
    new Promise<Abzug>((erfuellen) => {
      offen.push(erfuellen)
    })
  const { rerender } = render(feld('ada', abzug))
  const hand = nutzerereignis.setup()
  const knopf = () => screen.getByRole('button', { name: eigene.abzugErzeugen })

  // **Aus Adas Sicht angestoßen**, dann in die Zuschauersicht gewechselt und
  // dort noch einmal.
  await hand.click(knopf())
  rerender(feld('', abzug))
  await hand.click(knopf())

  // **Die frische Antwort zuerst** — ihre Vorschau steht da.
  await erfuelle(() => {
    offen[1]?.({ ...einBild, bild: 'data:image/png;base64,BBB' })
  })
  expect(await screen.findByAltText(eigene.abzugVorschau)).toHaveAttribute(
    'src',
    'data:image/png;base64,BBB',
  )

  // **Dann die späte aus Adas Sicht** — sie darf nichts ersetzen.
  await erfuelle(() => {
    offen[0]?.(einBild)
  })
  expect(screen.getByAltText(eigene.abzugVorschau)).toHaveAttribute(
    'src',
    'data:image/png;base64,BBB',
  )
})

it('meldet einen gesicherten Ort nicht in einer anderen Sicht', async () => {
  const offen: ((antwort: {
    pfad: string
    abgebrochen: boolean
    fehler: null
  }) => void)[] = []
  belegeApi({
    TischabzugSpeichern: () =>
      new Promise((erfuellen) => {
        offen.push(erfuellen)
      }),
  })
  // **Je Abzug ein eigenes Bild** — wie im Betrieb; der Ort gehört zu dem
  // Bild, das gesichert wurde, und zwei Abzüge sind zwei Bilder.
  const abzug = () => Promise.resolve({ ...einBild })
  const { rerender } = render(feld('ada', abzug))
  const hand = nutzerereignis.setup()
  const sichern = () =>
    screen.findByRole('button', { name: eigene.abzugSichern })

  // **Aus Adas Sicht gesichert**, und bevor der Ort zurückkommt, steht in der
  // Zuschauersicht schon eine eigene Vorschau.
  await hand.click(screen.getByRole('button', { name: eigene.abzugErzeugen }))
  await hand.click(await sichern())
  rerender(feld('', abzug))
  await hand.click(screen.getByRole('button', { name: eigene.abzugErzeugen }))
  await hand.click(await sichern())

  // **Die eigene Sicherung zuerst** — ihr Ort steht da.
  await erfuelle(() => {
    offen[1]?.({ pfad: '/tmp/tisch.png', abgebrochen: false, fehler: null })
  })
  const eigenerOrt = einsetzen(eigene.abzugGesichert, {
    pfad: '/tmp/tisch.png',
  })
  expect(await screen.findByText(eigenerOrt)).toBeInTheDocument()

  // **Dann der späte Ort aus Adas Sicht** — er darf nichts ersetzen.
  await erfuelle(() => {
    offen[0]?.({ pfad: '/tmp/ada.png', abgebrochen: false, fehler: null })
  })
  expect(screen.getByText(eigenerOrt)).toBeInTheDocument()
  expect(
    screen.queryByText(
      einsetzen(eigene.abzugGesichert, { pfad: '/tmp/ada.png' }),
    ),
  ).not.toBeInTheDocument()
})
