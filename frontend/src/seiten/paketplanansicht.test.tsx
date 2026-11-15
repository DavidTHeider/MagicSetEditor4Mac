// **Der Plan vor einer Paketänderung** — DOM-01.AK-02 und AK-03
// (#1696):
// betroffene Sets, gebrochene Abhängigkeiten und eine lokale Änderung stehen
// da, **bevor** jemand bestätigt.
import { cleanup, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { PaketplanDTO } from '../api'
import type { Paketplanung } from '../haken/usePaketplan'
import { kataloge } from '../sprache/kataloge'
import { Paketplanansicht } from './Paketplanansicht'

afterEach(cleanup)

const t = kataloge.de.seiten.pakete.plan

function planung(teil: Partial<PaketplanDTO> | undefined, fehler?: string) {
  const plan: PaketplanDTO | undefined =
    teil === undefined
      ? undefined
      : {
          vorgang: 'entfernen',
          paket: 'probe.mse-game',
          quelle: '',
          von: '1.0',
          zu: '',
          geaendert: false,
          moeglich: true,
          sets: [],
          konflikte: [],
          ...teil,
        }

  return {
    plan,
    fehler,
    laeuft: false,
    planen: vi.fn(),
    waehlen: vi.fn(),
    ausfuehren: vi.fn(() => Promise.resolve()),
    verwerfen: vi.fn(),
  } satisfies Paketplanung
}

describe('Paketplanansicht', () => {
  it('zeigt ohne Plan nichts — und einen Fehler beim Planen', () => {
    const { container } = render(
      <Paketplanansicht planung={planung(undefined)} />,
    )
    expect(container).toBeEmptyDOMElement()

    render(<Paketplanansicht planung={planung(undefined, 'Klemmt.')} />)
    expect(screen.getByRole('alert')).toHaveTextContent('Klemmt.')
  })

  it('nennt vor dem Löschen Sets und gebrochene Abhängigkeiten', async () => {
    const stand = planung({
      sets: ['Erstes', 'Zweites'],
      konflikte: [
        {
          paket: 'probe-schlicht.mse-style',
          braucht: 'probe.mse-game',
          art: 'fehlt',
          verlangt: '',
          da: '',
        },
      ],
    })
    render(<Paketplanansicht planung={stand} />)

    expect(
      screen.getByRole('group', { name: 'probe.mse-game löschen' }),
    ).toBeInTheDocument()
    expect(screen.getByText('Erstes')).toBeInTheDocument()
    expect(
      screen.getByText(/probe-schlicht\.mse-style braucht/),
    ).toBeInTheDocument()

    // **Der Fokus steht auf dem Abbrechen** — die Rückfrage vor dem Löschen.
    expect(screen.getByRole('button', { name: t.verwerfen })).toHaveFocus()

    await userEvent.click(screen.getByRole('button', { name: /Löschen/ }))
    expect(stand.ausfuehren).toHaveBeenCalledWith(false)

    await userEvent.click(screen.getByRole('button', { name: t.verwerfen }))
    expect(stand.verwerfen).toHaveBeenCalledTimes(1)
  })

  it('ersetzt ein geändertes Paket nur mit dem Haken — und zeigt den Fehler dort', async () => {
    const stand = planung(
      { vorgang: 'ersetzen', von: '1.0', zu: '2.0', geaendert: true },
      'Lokal geändert.',
    )
    render(<Paketplanansicht planung={stand} />)

    expect(screen.getByText('Fassung 1.0 → 2.0')).toBeInTheDocument()
    expect(screen.getByText(t.planGeaendert)).toBeInTheDocument()
    expect(screen.getByRole('alert')).toHaveTextContent('Lokal geändert.')

    await userEvent.click(
      screen.getByRole('button', { name: t.bestaetigen.ersetzen }),
    )
    expect(stand.ausfuehren).toHaveBeenLastCalledWith(false)

    await userEvent.click(
      screen.getByRole('checkbox', { name: t.ueberschreiben }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: t.bestaetigen.ersetzen }),
    )
    expect(stand.ausfuehren).toHaveBeenLastCalledWith(true)

    // **Und wieder aus.**
    await userEvent.click(
      screen.getByRole('checkbox', { name: t.ueberschreiben }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: t.bestaetigen.ersetzen }),
    )
    expect(stand.ausfuehren).toHaveBeenLastCalledWith(false)
  })

  it('bietet nichts an, was nicht geht, und nennt, was nicht bricht', async () => {
    const stand = planung({ vorgang: 'zurueck', moeglich: false })
    render(<Paketplanansicht planung={stand} />)

    expect(screen.getByText(t.nichtMoeglich.zurueck)).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: t.bestaetigen.zurueck }),
    ).not.toBeInTheDocument()
    expect(screen.getByText(t.keineSets)).toBeInTheDocument()
    expect(screen.getByText(t.keineKonflikte)).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: t.verwerfen }))
    expect(stand.verwerfen).toHaveBeenCalledTimes(1)
  })

  it('nennt beim Einbinden die neue Fassung, auch ohne alte', () => {
    render(
      <Paketplanansicht
        planung={planung({ vorgang: 'einbinden', von: '', zu: '' })}
      />,
    )

    expect(
      screen.getByText(`Fassung ${t.ohneFassung} → ${t.ohneFassung}`),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: t.bestaetigen.einbinden }),
    ).toBeInTheDocument()
  })
})
