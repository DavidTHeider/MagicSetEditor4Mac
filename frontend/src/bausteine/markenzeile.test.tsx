// **Marken in jeder Liste** — UX-12.AK-01
// (#1684).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Markenzeile } from './Markenzeile'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.bausteine.objektmarken

afterEach(() => {
  raeumeApi()
})

describe('Die Markenzeile', () => {
  it('zeigt die Marken und öffnet den Editor auf Wunsch', async () => {
    const setzen = vi.fn().mockResolvedValue({ fehler: null })
    const nachladen = vi.fn()
    belegeApi({ ObjektmarkeSetzen: setzen })

    render(
      <Sprachraum sprache="de">
        <Markenzeile
          art="deck"
          kennung="11"
          name="Rot"
          marken={['turnier']}
          nachladen={nachladen}
        />
      </Sprachraum>,
    )

    expect(screen.getByText('turnier')).toBeInTheDocument()
    expect(screen.queryByRole('textbox')).toBeNull()

    await userEvent.click(
      screen.getByRole('button', { name: 'Marken von Rot' }),
    )
    await userEvent.type(
      screen.getByRole('textbox', { name: texte.setzenFuer }),
      'Blau',
    )
    await userEvent.click(screen.getByRole('button', { name: texte.setzen }))

    await waitFor(() => {
      expect(nachladen).toHaveBeenCalledTimes(1)
    })
    await antwortAbwarten()
    expect(setzen).toHaveBeenCalledWith('deck', '11', 'Blau')

    await userEvent.click(
      screen.getByRole('button', { name: texte.zeileSchliessen }),
    )
    expect(screen.queryByRole('textbox')).toBeNull()
  })

  it('zeigt ohne Marken nur den Knopf', () => {
    render(
      <Sprachraum sprache="de">
        <Markenzeile
          art="set"
          kennung="1"
          name="Probe"
          marken={[]}
          nachladen={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(
      screen.getByRole('button', { name: 'Marken von Probe' }),
    ).toBeInTheDocument()
    expect(screen.queryByRole('list')).toBeNull()
  })

  // **Leise in einer langen Liste** — UXR-24 (#1932): „+ Marke" an jedem
  // Eintrag hob die Sets über R2. Leise steht der Knopf nur unter dem Zeiger
  // oder mit dem Fokus in der Zeile da (`group/eintrag`, gesetzt von der
  // Liste) und bleibt im Baum — für die Tastatur und für VoiceOver. Offen ist
  // er immer zu sehen, denn dann schließt er den Editor.
  it('ist leise nur unter dem Zeiger oder mit dem Fokus zu sehen', async () => {
    render(
      <Sprachraum sprache="de">
        <Markenzeile
          art="set"
          kennung="3"
          name="Werkstatt"
          marken={[]}
          nachladen={vi.fn()}
          leise
        />
      </Sprachraum>,
    )

    const knopf = screen.getByRole('button', { name: 'Marken von Werkstatt' })

    // **Nur wo es einen Zeiger mit Hover gibt** — N1 des Nachlaufs über 6.13
    // (#1906): ohne Hover (Berührung, Stift) gäbe es keinen Weg, ihn zu
    // sehen, und ein Set ohne Marken zeigte keine Spur der Handlung.
    expect(knopf.className).toContain('[@media(hover:hover)]:opacity-0')
    expect(knopf.className).not.toMatch(/(^|\s)opacity-0(\s|$)/)
    expect(knopf.className).toContain('group-hover/eintrag:opacity-100')
    expect(knopf.className).toContain('group-focus-within/eintrag:opacity-100')

    await userEvent.click(knopf)
    expect(
      screen.getByRole('button', { name: texte.zeileSchliessen }).className,
    ).not.toContain('opacity-0')
  })

  it('ist ohne leise immer zu sehen', () => {
    render(
      <Sprachraum sprache="de">
        <Markenzeile
          art="deck"
          kennung="11"
          name="Rot"
          marken={[]}
          nachladen={vi.fn()}
        />
      </Sprachraum>,
    )

    expect(
      screen.getByRole('button', { name: 'Marken von Rot' }).className,
    ).not.toContain('opacity-0')
  })
})
