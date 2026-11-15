// **Die Seitenleiste klappt ein und aus** — Werkstück U14 von Phase 5.1
// (#1490): mit dem Knopf, mit ⌥⌘S, und eingeklappt bleibt jede Fläche über
// ihre Taste erreichbar. Der Fokus fällt nie ins Leere.
import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Rahmen } from './Rahmen'
import { belegeApi, raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { leisteneintrag } from '../test/leistenweg'

const navigation = kataloge.de.allgemein.navigation

function zeige(offen: boolean, umschalten = vi.fn(), wechseln = vi.fn()) {
  belegeApi({})

  const ansicht = render(
    <Sprachraum sprache="de">
      <Rahmen
        aktuell="start"
        wechseln={wechseln}
        seitenleiste={{ offen, umschalten }}
      >
        <p>{kataloge.de.seiten.start.was}</p>
      </Rahmen>
    </Sprachraum>,
  )

  return { ...ansicht, umschalten, wechseln }
}

afterEach(() => {
  raeumeApi()
})

describe('Die Seitenleiste', () => {
  it('trägt einen beschrifteten Knopf, der ihren Zustand sagt', () => {
    zeige(true)

    const knopf = screen.getByRole('button', {
      name: navigation.seitenleisteAusblenden,
    })
    expect(knopf).toHaveAttribute('aria-expanded', 'true')
    expect(knopf).toHaveAttribute('aria-controls', 'seitenleiste')
    expect(screen.getByRole('navigation')).toBeInTheDocument()
  })

  it('ist eingeklappt nicht mehr da — nur der Knopf', () => {
    zeige(false)

    const knopf = screen.getByRole('button', {
      name: navigation.seitenleisteEinblenden,
    })
    expect(knopf).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByRole('navigation')).not.toBeInTheDocument()
  })

  it('schaltet mit dem Knopf und mit ⌥⌘S um', async () => {
    const { umschalten } = zeige(true)

    await userEvent.click(
      screen.getByRole('button', { name: navigation.seitenleisteAusblenden }),
    )
    expect(umschalten).toHaveBeenCalledTimes(1)

    fireEvent.keyDown(window, {
      key: 's',
      code: 'KeyS',
      metaKey: true,
      altKey: true,
    })
    expect(umschalten).toHaveBeenCalledTimes(2)

    // ⌘S allein bleibt das Sichern — nicht die Leiste.
    fireEvent.keyDown(window, { key: 's', code: 'KeyS', metaKey: true })
    expect(umschalten).toHaveBeenCalledTimes(2)
  })

  // **Eingeklappt bleibt jede Fläche erreichbar** — über ihre Taste und ⌘K.
  it('lässt eingeklappt ⌘2 und ⌘K durch', () => {
    const { wechseln } = zeige(false)

    fireEvent.keyDown(window, { key: '2', metaKey: true })
    expect(wechseln).toHaveBeenCalledWith('sets')

    fireEvent.keyDown(window, { key: 'k', metaKey: true })
    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })

  // **Der Fokus fällt nicht ins Leere**: lag er in der Leiste, steht er danach
  // auf dem Knopf.
  it('holt den Fokus aus der eingeklappten Leiste auf den Knopf', async () => {
    const { rerender } = zeige(true)

    const sets = await leisteneintrag('sets')
    sets.focus()
    expect(sets).toHaveFocus()

    rerender(
      <Sprachraum sprache="de">
        <Rahmen
          aktuell="start"
          wechseln={vi.fn()}
          seitenleiste={{ offen: false, umschalten: vi.fn() }}
        >
          <p>{kataloge.de.seiten.start.was}</p>
        </Rahmen>
      </Sprachraum>,
    )

    expect(
      screen.getByRole('button', { name: navigation.seitenleisteEinblenden }),
    ).toHaveFocus()
  })

  // **Ohne Angabe steht die Leiste**, und Knopf wie ⌥⌘S tun nichts — so
  // rendern die Tests der übrigen Seiten den Rahmen.
  it('steht ohne Angabe und lässt sich dann nicht umschalten', async () => {
    belegeApi({})
    render(
      <Sprachraum sprache="de">
        <Rahmen aktuell="start" wechseln={vi.fn()}>
          <p>{kataloge.de.seiten.start.was}</p>
        </Rahmen>
      </Sprachraum>,
    )

    await userEvent.click(
      screen.getByRole('button', { name: navigation.seitenleisteAusblenden }),
    )
    fireEvent.keyDown(window, {
      key: 's',
      code: 'KeyS',
      metaKey: true,
      altKey: true,
    })

    expect(screen.getByRole('navigation')).toBeInTheDocument()
  })

  it('hält die Regeln der Barrierefreiheit in beiden Zuständen', async () => {
    const { container, rerender } = zeige(true)
    expect(await verstoesse(container)).toEqual([])

    rerender(
      <Sprachraum sprache="de">
        <Rahmen
          aktuell="start"
          wechseln={vi.fn()}
          seitenleiste={{ offen: false, umschalten: vi.fn() }}
        >
          <p>{kataloge.de.seiten.start.was}</p>
        </Rahmen>
      </Sprachraum>,
    )
    expect(await verstoesse(container)).toEqual([])
  })
})
