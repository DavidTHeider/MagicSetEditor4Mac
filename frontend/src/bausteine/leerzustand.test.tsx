// **Ein Leerzustand ist für eine Probe erkennbar** — UXR-23
// (#1931).
//
// Die Hüllenprobe `uxr23` geht jede Seite mit einem frischen Bestand durch
// und zählt, wo ein Leerzustand steht und wo statt seiner ein Fehler. Sie
// erkennt ihn am Merkmal `data-leerzustand` und nicht an seinen Klassen: die
// sind Gestalt und dürfen sich ändern.
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { Leerzustand } from './Leerzustand'

describe('der Leerzustand', () => {
  it('trägt das Merkmal, an dem eine Probe ihn findet', () => {
    render(<Leerzustand titel="Noch nichts" erklaerung="Lege etwas an." />)

    expect(
      screen.getByText('Noch nichts').closest('[data-leerzustand]'),
    ).not.toBeNull()
  })
})
