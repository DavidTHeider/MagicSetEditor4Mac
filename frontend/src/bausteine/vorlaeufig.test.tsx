// **Die Marke für einen Druck ohne Alias** — Werkstück A2 von Phase 5.3 (#1441).
import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Vorlaeufig } from './Vorlaeufig'

const texte = kataloge.de.bausteine.druck

describe('die Marke »vorläufig«', () => {
  it('steht mit ihrer Erklärung da, wenn der Druck keinen Alias hat', () => {
    render(
      <Sprachraum sprache="de">
        <Vorlaeufig vorlaeufig />
      </Sprachraum>,
    )

    expect(screen.getByText(texte.vorlaeufig)).toHaveAttribute(
      'title',
      texte.vorlaeufigErklaerung,
    )
  })

  it('bleibt aus, wenn der Druck einen Alias hat', () => {
    render(
      <Sprachraum sprache="de">
        <Vorlaeufig vorlaeufig={false} />
      </Sprachraum>,
    )

    expect(screen.queryByText(texte.vorlaeufig)).not.toBeInTheDocument()
  })
})
