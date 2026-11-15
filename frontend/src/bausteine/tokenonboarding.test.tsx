// **So richtet man einen MCP-Client ein** — Werkstück O2 von Phase 5.2
// (#1426).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Tokenonboarding } from './Tokenonboarding'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.nutzer

describe('Tokenonboarding', () => {
  it('zeigt nach dem Aufklappen vier Schritte und das Beispiel', async () => {
    render(
      <Sprachraum sprache="de">
        <Tokenonboarding />
      </Sprachraum>,
    )

    await userEvent.click(
      screen.getByRole('button', { name: eigene.tokenOnboarding }),
    )

    expect(screen.getAllByRole('listitem')).toHaveLength(4)
    expect(screen.getByText(eigene.tokenOnboardingSchritt3)).toBeInTheDocument()
    expect(screen.getByText(eigene.tokenOnboardingBeispiel)).toBeInTheDocument()
  })
})
