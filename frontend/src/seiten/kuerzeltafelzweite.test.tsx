// **Die Kürzeltafel zeigt auch die zweite Reihe** —
// #1291,
// Werkstück E.
//
// Ein Kürzel, das wirkt und nirgends steht, kennt nur, wer es kennt — das war
// der Befund, aus dem die Tafel entstanden ist
// (#245). Mit
// ⌥⌘1…⌥⌘8 sind acht dazugekommen.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { seiten } from '../navigation'
import { Kuerzeltafel } from './Kuerzeltafel'
import { Sprachraum } from '../sprache/kontext'

describe('die Kürzeltafel', () => {
  it('nennt jedes Zweitkürzel mit ⌥⌘', () => {
    render(
      <Sprachraum sprache="de">
        <Kuerzeltafel />
      </Sprachraum>,
    )

    const zweite = seiten.filter((seite) => seite.zweitkuerzel !== '')

    expect(zweite.length).toBeGreaterThan(0)

    for (const seite of zweite) {
      expect(screen.getByText(`⌥⌘${seite.zweitkuerzel}`)).toBeInTheDocument()
    }
  })

  // **Und sie sagt, welche Taste das ist** —
  // #1345.
  //
  // `⌥⌘1` ist eine Beschriftung, gedeutet wird über `event.code`. Der benennt
  // die **Position**: auf AZERTY liegt an `Digit1` ein `&`, und die Tafel
  // schrieb trotzdem „⌥⌘1". Eine Beschriftung je Belegung ginge nur über
  // `navigator.keyboard.getLayoutMap()` — Chromium hat sie, das WKWebView
  // dieses Fensters nicht. Also steht der Satz da.
  // **Der Satz steht da, die Belegung dahinter** — UX-15 (#1771).
  it('sagt, dass die Ziffern die Tasten der obersten Reihe sind', async () => {
    render(
      <Sprachraum sprache="de">
        <Kuerzeltafel />
      </Sprachraum>,
    )

    expect(screen.getByText(/obersten Reihe/)).toBeInTheDocument()

    await userEvent.hover(
      screen.getByRole('button', { name: /Mehr zu „Die acht Kürzel/ }),
    )
    expect(await screen.findByRole('tooltip')).toHaveTextContent('AZERTY')
  })

  it('nennt jede Seite genau einmal', () => {
    render(
      <Sprachraum sprache="de">
        <Kuerzeltafel />
      </Sprachraum>,
    )

    // **Keine Seite steht mit zwei Kürzeln da.** Ein Eintrag trägt entweder
    // eine Ziffer der ersten Reihe oder eine der zweiten.
    const mitKuerzel = seiten.filter(
      (seite) => seite.kuerzel !== '' || seite.zweitkuerzel !== '',
    )

    for (const seite of mitKuerzel) {
      expect(seite.kuerzel === '' || seite.zweitkuerzel === '').toBe(true)
    }
  })
})
