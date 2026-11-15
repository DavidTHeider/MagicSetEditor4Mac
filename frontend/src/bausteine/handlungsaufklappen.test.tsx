// **Die Felder eines Befehls stehen erst auf Zuruf da** — UX-08
// (#1680),
// „dynamische Interaktionsmuster statt Formularwänden", und das Artboard
// *Tisch*: „Mit dieser Karte" zeigt Knöpfe und keine Formulare.
//
// # Wogegen hier geprüft wird
//
// **Dass die Spalte neben dem Tisch aus Knöpfen besteht.** Im Sandkasten ist
// fast jeder Befehl erlaubt, und bis zum 2026-09-24 stand jeder mit seinen
// Feldern da — zwanzig Formulare untereinander, und die gewählte Karte
// darüber rutschte aus dem Blick.
//
// **Und dass das Aufklappen nichts verliert und nichts stehen lässt**: was
// eingegeben war, bleibt beim Zuklappen; was zu einer anderen Karte gehörte,
// geht mit ihr (§7, Regel 1).
import { render, screen, within } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'
import nutzerereignis from '@testing-library/user-event'

import { Handlungsleiste } from './Handlungsleiste'
import { Sprachraum } from '../sprache/kontext'
import { einePartie } from '../seiten/partieprobe'
import {
  eigene,
  eineKarte,
  form,
  name,
  oeffne,
  stelle,
  zeige,
} from '../test/handlungsprobe'

import type { Spielobjekt } from '../apipartie'

describe('ein Befehl mit Feldern', () => {
  test('zeigt seine Felder erst auf Zuruf', async () => {
    const { hand } = zeige({ erlaubt: ['spieler.leben'] })

    const knopf = screen.getByRole('button', { name: name('spieler.leben') })

    expect(knopf).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByLabelText(eigene.handlungsfeld.um)).toBeNull()
    expect(
      screen.queryByRole('button', { name: eigene.handlungStellen }),
    ).toBeNull()

    const felder = await oeffne(hand, 'spieler.leben')

    expect(knopf).toHaveAttribute('aria-expanded', 'true')
    expect(knopf).toHaveAttribute('aria-controls', felder.id)
    expect(
      within(felder).getByLabelText(eigene.handlungsfeld.um),
    ).toBeInTheDocument()
  })

  test('klappt nach dem Stellen wieder zu', async () => {
    const { hand, befehlen } = zeige({ erlaubt: ['spieler.leben'] })

    await stelle(hand, 'spieler.leben')

    expect(befehlen).toHaveBeenCalledWith('spieler.leben', { um: -1 })
    expect(form('spieler.leben')).toBeNull()

    // **Und der Fokus steht wieder auf dem Knopf** (UX-08.AK-03): der
    // gedrückte ist mit den Feldern verschwunden.
    const knopf = screen.getByRole('button', { name: name('spieler.leben') })

    expect(knopf).toHaveAttribute('aria-expanded', 'false')
    expect(knopf).toHaveFocus()
  })

  // **Einer zur Zeit je Gruppe** — zwei offene Formulare wären wieder die
  // Wand, die hier verschwinden soll.
  test('schließt den, der vorher offen war', async () => {
    const { hand } = zeige({
      erlaubt: ['spieler.leben', 'spieler.vergiftung'],
    })

    await oeffne(hand, 'spieler.leben')
    await oeffne(hand, 'spieler.vergiftung')

    expect(form('spieler.leben')).toBeNull()
    expect(form('spieler.vergiftung')).toBeInTheDocument()
  })

  test('behält beim Zuklappen, was eingegeben war', async () => {
    const { hand } = zeige({ erlaubt: ['spieler.vergiftung'] })

    const felder = await oeffne(hand, 'spieler.vergiftung')
    const feld = within(felder).getByLabelText(eigene.handlungsfeld.um)

    await hand.clear(feld)
    await hand.type(feld, '3')

    // Zu und wieder auf, über denselben Knopf.
    await hand.click(
      screen.getByRole('button', { name: name('spieler.vergiftung') }),
    )
    expect(form('spieler.vergiftung')).toBeNull()

    const wieder = await oeffne(hand, 'spieler.vergiftung')

    expect(within(wieder).getByLabelText(eigene.handlungsfeld.um)).toHaveValue(
      3,
    )
  })
})

describe('ein Befehl ohne Felder', () => {
  // **Er klappt nichts auf**, und ein `aria-expanded` an ihm sagte einem
  // Vorleseprogramm, es gebe etwas zu öffnen (#697).
  test('ist ein Knopf, der sofort stellt', () => {
    zeige({ erlaubt: ['objekt.umdrehen'], objekt: eineKarte() })

    expect(
      screen.getByRole('button', { name: name('objekt.umdrehen') }),
    ).not.toHaveAttribute('aria-expanded')
  })
})

// **Eine andere Karte räumt die offene Form** (§7, Regel 1): die Anzahl der
// Marken gehörte dem Wald, und am Berg stünde sie als Angebot von vorhin.
test('eine andere Karte räumt die offenen Felder', async () => {
  const befehlen = vi.fn()

  const leiste = (objekt: Spielobjekt) => (
    <Sprachraum sprache="de">
      <Handlungsleiste
        partie={einePartie()}
        erlaubt={['objekt.zaehler']}
        objekt={objekt}
        befehlen={befehlen}
        angehalten={false}
      />
    </Sprachraum>
  )

  const { rerender } = render(leiste(eineKarte()))
  const hand = nutzerereignis.setup()

  const felder = await oeffne(hand, 'objekt.zaehler')
  const menge = within(felder).getByLabelText(eigene.handlungsfeld.menge)

  await hand.clear(menge)
  await hand.type(menge, '4')

  rerender(leiste(eineKarte({ kennung: 'sitz1/k2/0', name: 'Berg' })))

  expect(form('objekt.zaehler')).toBeNull()

  // **Die Gegenprobe** (§10): wieder aufgeklappt, steht die Vorgabe da und
  // nicht die Vier vom Wald.
  const neu = await oeffne(hand, 'objekt.zaehler')

  expect(within(neu).getByLabelText(eigene.handlungsfeld.menge)).toHaveValue(1)
})
