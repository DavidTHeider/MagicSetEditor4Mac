// **Was gerade geschah, steht an der Karte** — PLAY-04.AK-02
// (#1712).
//
// # Wogegen hier geprüft wird
//
// **Dass die Wirkung ankommt** — dass sie berechnet wird, steht in
// `tisch/tischwirkungen.test.ts`, und wie sie auf der Leinwand aussieht,
// sieht nur, wer davorsitzt (§3, „Sichtprüfung").
import { fireEvent, render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Partietisch } from './Partietisch'
import { Sprachraum } from '../sprache/kontext'
import { einLauf, einObjekt, einSitz, einePartie } from './partieprobe'
import { kataloge } from '../sprache/kataloge'
import { wirkungsbild } from '../tisch/tischwirkungen'

import type { Partie } from '../apipartie'

const eigene = kataloge.de.seiten.spielen

// zeigeZweimal zeichnet den Tisch erst mit dem einen Stand und dann mit dem
// anderen — eine Wirkung entsteht nur zwischen zwei Bildern.
function zeigeZweimal(vorher: Partie, nachher: Partie) {
  const { rerender } = render(
    <Sprachraum sprache="de">
      <Partietisch lauf={einLauf({ partie: vorher })} partie={vorher} />
    </Sprachraum>,
  )

  // **Die Zonenliste, nicht die Leinwand**: ohne Grafikkarte zeichnet jsdom
  // nichts, und was dort stünde, könnte kein Lauf lesen.
  fireEvent.click(screen.getByRole('button', { name: eigene.tischListe }))

  rerender(
    <Sprachraum sprache="de">
      <Partietisch lauf={einLauf({ partie: nachher })} partie={nachher} />
    </Sprachraum>,
  )
}

const aufDerHand = einePartie({
  fassung: '1',
  objekte: [
    einObjekt({ kennung: 'k1', zone: 'hand', name: 'Wald', geheim: false }),
  ],
  sitze: [einSitz({ kennung: 'sitz1', name: 'Anna', leben: 20 })],
})

const aufDemTisch = einePartie({
  fassung: '2',
  objekte: [
    einObjekt({ kennung: 'k1', zone: 'tisch', name: 'Wald', geheim: false }),
  ],
  sitze: [einSitz({ kennung: 'sitz1', name: 'Anna', leben: 20 })],
})

// **Die Marke hängt an der Karte und nennt ihre Klasse.**
it('markiert eine Karte, die gerade eingetreten ist', () => {
  zeigeZweimal(aufDerHand, aufDemTisch)

  const karte = screen.getByRole('button', { name: /Wald/ })

  expect(karte).toHaveAttribute('data-wirkung', 'eintritt')
})

// **Und sie trägt ein Zeichen, nicht nur eine Farbe** — WCAG 1.4.1.
it('zeigt das Zeichen der Wirkung', () => {
  zeigeZweimal(aufDerHand, aufDemTisch)

  expect(screen.getByRole('button', { name: /Wald/ })).toHaveTextContent(
    wirkungsbild('eintritt').zeichen,
  )
})

// **Ohne Änderung keine Marke** — sonst stünde sie an jeder Karte, und dann
// sagte sie nichts.
it('markiert nichts, wo sich nichts geändert hat', () => {
  zeigeZweimal(aufDerHand, einePartie({ ...aufDerHand, fassung: '2' }))

  expect(screen.getByRole('button', { name: /Wald/ })).not.toHaveAttribute(
    'data-wirkung',
  )
})
