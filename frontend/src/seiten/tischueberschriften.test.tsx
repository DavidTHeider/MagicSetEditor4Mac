// **Die Überschriften am Tisch springen nicht** —
// #1992.
//
// Gemessen am 2026-10-01 in der Abnahme-VM mit VoiceOver: „At the table" war
// eine h1, „What lies where" eine h3 — wer mit VO-⌘H durch die Seite geht,
// hört eine Stufe, die es nicht gibt. Der Tisch steht an zwei Orten, und an
// beiden gilt: **keine Überschrift ist mehr als eine Stufe tiefer als die
// vor ihr**.
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Partietisch } from './Partietisch'
import { Sprachraum } from '../sprache/kontext'
import { einLauf, einePartie } from './partieprobe'

function stufen(): number[] {
  return screen
    .getAllByRole('heading')
    .map((ueberschrift) => Number(ueberschrift.tagName.slice(1)))
}

describe('die Überschriften am Tisch', () => {
  it.each([1, 2] as const)('folgen unter einer h%s ohne Sprung', (ebene) => {
    render(
      <Sprachraum sprache="de">
        <Partietisch lauf={einLauf()} partie={einePartie()} ebene={ebene} />
      </Sprachraum>,
    )

    const gefunden = stufen()

    expect(gefunden[0]).toBe(ebene)
    expect(gefunden.length).toBeGreaterThan(3)

    for (let i = 1; i < gefunden.length; i++) {
      expect(
        gefunden[i] ?? 0,
        `Überschrift ${String(i + 1)} von ${String(gefunden.length)}`,
      ).toBeLessThanOrEqual((gefunden[i - 1] ?? 0) + 1)
    }
  })
})
