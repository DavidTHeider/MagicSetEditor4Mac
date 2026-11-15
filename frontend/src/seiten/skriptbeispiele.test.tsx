// **Die Beispiele der Skriptkonsole** — Werkstück 1 von Phase 3.3
// (#1164).
//
// **Ob sie laufen, prüft Go** — `internal/service/beispiele_test.go` wertet
// jeden Ausdruck dieser Datei aus. Hier geht es um die Bedienung: kommt der
// Ausdruck ins Feld, und steht dabei, welche Gruppe eine Vorlage braucht?
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { Skriptbeispiele } from './Skriptbeispiele'
import { Sprachraum } from '../sprache/kontext'
import { immerLaufende, mitVorlage } from './beispiele'
import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'

const texte = kataloge.de.seiten.karten.konsole

describe('Die Beispiele', () => {
  it('setzen einen Ausdruck per Klick ein', async () => {
    const eingesetzt = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Skriptbeispiele einsetzen={eingesetzt} />
      </Sprachraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: /1 \+ 1/ }))

    expect(eingesetzt).toHaveBeenCalledWith('1 + 1')
  })

  // **Beide Gruppen stehen da, und die zweite sagt, was sie braucht.**
  //
  // Ohne diesen Satz führe der erste Klick eines Neulings auf `card.name` in
  // eine Fehlermeldung, die nichts erklärt.
  it('nennt beide Gruppen und ihre Bedingung', () => {
    render(
      <Sprachraum sprache="de">
        <Skriptbeispiele einsetzen={vi.fn()} />
      </Sprachraum>,
    )

    expect(screen.getByText(texte.beispieleImmer)).toBeInTheDocument()
    expect(screen.getByText(texte.beispieleVorlage)).toBeInTheDocument()
    expect(
      screen.getByText(kurzform(texte.beispieleVorlageErklaerung)),
    ).toBeInTheDocument()
  })

  // **Jedes Beispiel steht mit seinem Satz da.**
  //
  // Der Knopf trägt beides im Namen — wer die Sprache nicht kennt, hört den
  // Satz und nicht nur `to_upper("magic")`.
  it('nennt zu jedem Ausdruck seinen Satz', () => {
    render(
      <Sprachraum sprache="de">
        <Skriptbeispiele einsetzen={vi.fn()} />
      </Sprachraum>,
    )

    for (const beispiel of [...immerLaufende, ...mitVorlage]) {
      const satz = texte.beispielListe[beispiel.schluessel]

      expect(
        screen.getByRole('button', {
          name: `${satz} — ${beispiel.ausdruck}`,
        }),
      ).toBeInTheDocument()
    }
  })
})
