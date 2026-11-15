// **Der Kernweg `deck` in sechs Handlungen** — AP-6
// (#2363),
// Befund B-22: ⌘4, „Deck anlegen“, Name, „Anlegen“, Karte, „Ins Deck
// legen“. Gemessen standen zehn da (`belege/huellenprobe/ausgangswerte.md`),
// und drei davon nennt `docs/UX-ARCHITEKTUR.md` beim Namen: der Klick ins
// Namensfeld, die Wahl des Spiels und die Wahl des Sets. **Das offene Set
// beantwortet beide Wahlen** — es ist das, woran gerade gearbeitet wird.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { Api, SetDTO } from '../api'
import { Deckkartenwahl } from '../bausteine/Deckkartenwahl'
import { OffenesSet } from '../haken/offenesset'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Decks } from './Decks'
import { bauartprobe, formatprobe } from './deckprobe'

const texte = kataloge.de.seiten.decks

afterEach(() => {
  raeumeApi()
})

const satz = (id: string, spiel: string): SetDTO => ({
  id,
  name: `Satz ${id}`,
  spiel,
  stil: 'm15',
  herkunft: '',
  herkunftFehlt: false,
  karten: 2,
})

function belege(sets: readonly SetDTO[], teile: Partial<Api> = {}) {
  belegeApi({
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Formate: () => Promise.resolve({ formate: [formatprobe()], fehler: null }),
    Bauarten: () =>
      Promise.resolve({ bauarten: [bauartprobe()], fehler: null }),
    Sets: () => Promise.resolve({ sets: [...sets], fehler: null }),
    Kartenzeilen: () =>
      Promise.resolve({
        zeilen: [{ id: '11', nummer: 1, name: 'Lightning Bolt', felder: 4 }],
        fehler: null,
      }),
    ...teile,
  })
}

describe('die Anlage eines Decks', () => {
  it('setzt den Fokus ins Namensfeld, sobald der Kasten aufgeht', async () => {
    belege([satz('1', 'magic')])
    render(
      <Sprachraum sprache="de">
        <Decks wechseln={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.neu }),
    )

    await waitFor(() => {
      expect(screen.getByLabelText(new RegExp(`^${texte.name}`))).toHaveFocus()
    })
  })

  it('belegt das Spiel mit dem des offenen Sets vor', async () => {
    const sets = [satz('1', 'vs'), satz('2', 'probe')]
    belege(sets)
    render(
      <Sprachraum sprache="de">
        <OffenesSet wert={sets[1]}>
          <Decks wechseln={vi.fn()} />
        </OffenesSet>
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.neu }),
    )

    const wahl = await screen.findByLabelText(texte.spiel)
    await waitFor(() => {
      expect((wahl as HTMLSelectElement).value).toBe('probe')
    })
  })
})

describe('die Kartenwahl im offenen Deck', () => {
  function zeige(spiel: string, offen: SetDTO | undefined) {
    render(
      <Sprachraum sprache="de">
        <OffenesSet wert={offen}>
          <Deckkartenwahl spiel={spiel} angehalten={false} legen={vi.fn()} />
        </OffenesSet>
      </Sprachraum>,
    )

    return screen.findByLabelText(texte.setWaehlen)
  }

  it('belegt das offene Set vor, wenn es zum Spiel des Decks gehört', async () => {
    const sets = [satz('1', 'magic'), satz('2', 'probe')]
    belege(sets)

    const wahl = await zeige('probe', sets[1])

    await waitFor(() => {
      expect((wahl as HTMLSelectElement).value).toBe('2')
    })
    // **Und die Karten des Sets stehen schon zur Wahl** — die nächste
    // Handlung ist die Karte, nicht das Set.
    expect(
      await screen.findByRole('option', { name: 'Lightning Bolt' }),
    ).toBeInTheDocument()
  })

  it('nimmt das einzige Set des Spiels, wenn keines offen ist', async () => {
    belege([satz('1', 'magic'), satz('2', 'probe')])

    const wahl = await zeige('probe', undefined)

    await waitFor(() => {
      expect((wahl as HTMLSelectElement).value).toBe('2')
    })
  })

  it('belegt nichts vor, wenn das offene Set zu einem anderen Spiel gehört und mehrere passen', async () => {
    const sets = [satz('1', 'magic'), satz('2', 'probe'), satz('3', 'magic')]
    belege(sets)

    const wahl = await zeige('magic', sets[1])

    // **Erst die Sets laden lassen** — sonst prüfte der Test die leere Wahl
    // vor der Antwort und nicht die Entscheidung danach.
    expect(
      await screen.findByRole('option', { name: 'Satz 3' }),
    ).toBeInTheDocument()
    expect((wahl as HTMLSelectElement).value).toBe('')
  })

  it('lässt eine Wahl stehen, auch wenn das offene Set danach ankommt', async () => {
    const sets = [satz('1', 'magic'), satz('2', 'magic')]
    belege(sets)

    const wahl = await zeige('magic', undefined)
    await screen.findByRole('option', { name: 'Satz 2' })
    await userEvent.selectOptions(wahl, '1')

    expect((wahl as HTMLSelectElement).value).toBe('1')
  })
})

// **Wechselt die Vorbelegung, geht die gewählte Karte mit** — Zweitblick
// über Phase 6, zweiter Teil, Befund 2 (§7, Regel 1 und 2): die Karte
// gehörte dem vorbelegten Set A; öffnet jemand Set B, folgt die Vorbelegung,
// und „Ins Deck legen“ schriebe sonst die Karte aus A.
it('räumt die Karte, wenn die Vorbelegung des Sets wechselt', async () => {
  const sets = [satz('1', 'probe'), satz('2', 'probe')]
  belege(sets)
  const legen = vi.fn()
  const { rerender } = render(
    <Sprachraum sprache="de">
      <OffenesSet wert={sets[0]}>
        <Deckkartenwahl spiel="probe" angehalten={false} legen={legen} />
      </OffenesSet>
    </Sprachraum>,
  )
  const karte = await screen.findByLabelText(texte.karteWaehlen)
  await screen.findByRole('option', { name: 'Lightning Bolt' })
  await userEvent.selectOptions(karte, '11')
  expect(screen.getByRole('button', { name: texte.legen })).toBeEnabled()

  rerender(
    <Sprachraum sprache="de">
      <OffenesSet wert={sets[1]}>
        <Deckkartenwahl spiel="probe" angehalten={false} legen={legen} />
      </OffenesSet>
    </Sprachraum>,
  )

  await waitFor(() => {
    expect(
      screen.getByLabelText<HTMLSelectElement>(texte.setWaehlen).value,
    ).toBe('2')
  })
  expect(
    screen.getByLabelText<HTMLSelectElement>(texte.karteWaehlen).value,
  ).toBe('')
  expect(screen.getByRole('button', { name: texte.legen })).toBeDisabled()
})

// **Zurück nach A holt die alte Karte nicht wieder** — Nachlauf zum
// Zweitblick über Phase 6, zweiter Teil, R2: die Wahl war nur ausgeblendet,
// solange B galt; zurück in A stand sie sofort wieder, während As Karten
// noch luden, und „Ins Deck legen“ schrieb die alte Kennung.
it('verwirft die Karte bei jedem Setwechsel, auch zurück ins alte Set', async () => {
  const sets = [satz('1', 'probe'), satz('2', 'probe')]
  let zurueck: (wert: Awaited<ReturnType<Api['Kartenzeilen']>>) => void = () =>
    undefined
  const zeilen = vi
    .fn<Api['Kartenzeilen']>()
    .mockResolvedValueOnce({
      zeilen: [{ id: '11', nummer: 1, name: 'Lightning Bolt', felder: 4 }],
      fehler: null,
    })
    .mockResolvedValueOnce({
      zeilen: [{ id: '21', nummer: 1, name: 'Shock', felder: 4 }],
      fehler: null,
    })
    .mockImplementationOnce(
      () =>
        new Promise((fertig) => {
          zurueck = fertig
        }),
    )
  belege(sets, { Kartenzeilen: zeilen })
  const legen = vi.fn()
  const zeige = (offen: SetDTO | undefined) => (
    <Sprachraum sprache="de">
      <OffenesSet wert={offen}>
        <Deckkartenwahl spiel="probe" angehalten={false} legen={legen} />
      </OffenesSet>
    </Sprachraum>
  )
  const { rerender } = render(zeige(sets[0]))
  await screen.findByRole('option', { name: 'Lightning Bolt' })
  await userEvent.selectOptions(screen.getByLabelText(texte.karteWaehlen), '11')

  rerender(zeige(sets[1]))
  await screen.findByRole('option', { name: 'Shock' })

  rerender(zeige(sets[0]))
  await waitFor(() => {
    expect(
      screen.getByLabelText<HTMLSelectElement>(texte.setWaehlen).value,
    ).toBe('1')
  })
  const knopf = screen.getByRole('button', { name: texte.legen })
  expect(knopf).toBeDisabled()

  zurueck({
    zeilen: [{ id: '11', nummer: 1, name: 'Lightning Bolt', felder: 4 }],
    fehler: null,
  })
  await screen.findByRole('option', { name: 'Lightning Bolt' })
  expect(
    screen.getByLabelText<HTMLSelectElement>(texte.karteWaehlen).value,
  ).toBe('')
  expect(knopf).toBeDisabled()
  expect(legen).not.toHaveBeenCalled()
})
