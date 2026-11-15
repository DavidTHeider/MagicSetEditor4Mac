// **Der Filter, den die Statistik über die Kartenliste legt** — UXR-F6
// (#2013).
//
// Der Raum ist die Naht zwischen zwei Flächen, und was er zusagt, ist
// schlicht: **eine Wahl, zwei Leser**, und die Wege hin und zurück gehen über
// die Seitenwechsel des Sets. Daneben steht, was gilt, wo es keinen Raum gibt
// — die Kartenliste arbeitet auch ohne Statistik nach Plan.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import {
  Statistikfilterraum,
  useStatistikfilter,
  type Gruppenwahl,
} from './statistikfilter'

const rare: Gruppenwahl = {
  wahl: 'feld:rarity',
  quelle: 'rarity',
  gruppe: 'rare',
  kartenIDs: ['3', '7'],
}

// Leser zeigt die Wahl und hat für jeden Griff einen Knopf — so sieht der
// Test, was ein Leser im Raum bekommt.
function Leser() {
  const filter = useStatistikfilter()

  return (
    <>
      <p data-testid="wahl">
        {filter.wahl === undefined
          ? 'keine'
          : `${filter.wahl.gruppe}: ${filter.wahl.kartenIDs.join(',')}`}
      </p>
      <button
        type="button"
        onClick={() => {
          filter.waehlen(rare)
        }}
      >
        wählen
      </button>
      <button type="button" onClick={filter.aufheben}>
        aufheben
      </button>
      <button type="button" onClick={filter.zurStatistik}>
        zur Statistik
      </button>
      <button type="button" onClick={filter.zurKartenliste}>
        zur Kartenliste
      </button>
    </>
  )
}

describe('Der Raum des Statistikfilters', () => {
  it('hält die Wahl, bis jemand sie aufhebt', async () => {
    render(
      <Statistikfilterraum wechseln={vi.fn()}>
        <Leser />
      </Statistikfilterraum>,
    )

    expect(screen.getByTestId('wahl')).toHaveTextContent('keine')

    await userEvent.click(screen.getByRole('button', { name: 'wählen' }))

    expect(screen.getByTestId('wahl')).toHaveTextContent('rare: 3,7')

    await userEvent.click(screen.getByRole('button', { name: 'aufheben' }))

    expect(screen.getByTestId('wahl')).toHaveTextContent('keine')
  })

  // **Hin und zurück sind Seitenwechsel**, und der Raum weiß nicht, wo die
  // Seiten stehen: er sagt dem Arbeitsplatz, wohin.
  it('führt zur Statistik und zur Kartenliste über den Seitenwechsel', async () => {
    const wechseln = vi.fn()

    render(
      <Statistikfilterraum wechseln={wechseln}>
        <Leser />
      </Statistikfilterraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'zur Statistik' }))
    expect(wechseln).toHaveBeenLastCalledWith('statistik')

    await userEvent.click(
      screen.getByRole('button', { name: 'zur Kartenliste' }),
    )
    expect(wechseln).toHaveBeenLastCalledWith('karten')
  })

  // **Ein anderer Raum ist eine andere Wahl** (§7, Regel 1): `Setbereich`
  // baut ihn mit dem Set neu, und die Wahl des einen Sets erreicht das andere
  // nicht.
  it('trägt in einem neuen Raum keine Wahl des alten', async () => {
    const { rerender } = render(
      <Statistikfilterraum key="a" wechseln={vi.fn()}>
        <Leser />
      </Statistikfilterraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: 'wählen' }))
    expect(screen.getByTestId('wahl')).toHaveTextContent('rare')

    rerender(
      <Statistikfilterraum key="b" wechseln={vi.fn()}>
        <Leser />
      </Statistikfilterraum>,
    )

    expect(screen.getByTestId('wahl')).toHaveTextContent('keine')
  })

  // **Ohne Raum gilt kein Filter**, und jeder Griff tut nichts — die
  // Kartenliste und die Statistik arbeiten auch dort, wo sie niemand verbindet.
  it('tut ohne Raum nichts und wirft nichts', async () => {
    render(<Leser />)

    for (const name of [
      'wählen',
      'aufheben',
      'zur Statistik',
      'zur Kartenliste',
    ]) {
      await userEvent.click(screen.getByRole('button', { name }))
    }

    expect(screen.getByTestId('wahl')).toHaveTextContent('keine')
  })
})
