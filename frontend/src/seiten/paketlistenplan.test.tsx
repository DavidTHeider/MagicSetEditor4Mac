// **Einbinden und Löschen beginnen mit einem Plan** — DOM-01.AK-02
// (#1696):
// der Plan steht, bevor etwas geschieht, und ausgeführt wird genau er.
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import {
  anlegbar,
  entferne,
  fehler,
  katalog,
  plan,
  texte,
  zeige,
} from '../test/paketliste'

afterEach(() => {
  raeumeApi()
})

describe('Paketänderungen nach dem Plan', () => {
  it('bindet erst nach dem Plan ein und lädt die Liste nach', async () => {
    const Pakete = vi
      .fn()
      .mockResolvedValueOnce({ pakete: [], befunde: [] })
      .mockResolvedValue(katalog())
    const PaketPlanWaehlen = vi.fn(() =>
      Promise.resolve(plan('einbinden', '/q/probe-schlicht.mse-style')),
    )
    const PaketEinbindenVon = vi.fn(() =>
      Promise.resolve({ paket: null, abgebrochen: false }),
    )

    belegeApi({
      PaketArten: anlegbar,
      Pakete,
      PaketPlanWaehlen,
      PaketEinbindenVon,
    })
    zeige()

    await screen.findByText(texte.seiten.pakete.katalogLeer)
    await userEvent.click(kopfknopf())

    // **Der Plan steht, bevor etwas geschieht.**
    expect(await screen.findByText('Erstes')).toBeInTheDocument()
    expect(PaketEinbindenVon).not.toHaveBeenCalled()

    await userEvent.click(
      screen.getByRole('button', {
        name: texte.seiten.pakete.plan.bestaetigen.einbinden,
      }),
    )

    expect(PaketEinbindenVon).toHaveBeenCalledWith(
      '/q/probe-schlicht.mse-style',
    )
    expect(await screen.findByText('probe-schlicht')).toBeInTheDocument()
  })

  it('meldet einen Fehler beim Einbinden am Plan', async () => {
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn(() => Promise.resolve({ pakete: [], befunde: [] })),
      PaketPlanWaehlen: () => Promise.resolve(plan('einbinden', '/q')),
      PaketEinbindenVon: vi.fn(() =>
        Promise.resolve({
          paket: null,
          abgebrochen: false,
          fehler: fehler('Liegt schon da.'),
        }),
      ),
    })
    zeige()

    await screen.findByText(texte.seiten.pakete.katalogLeer)
    await userEvent.click(kopfknopf())
    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.pakete.plan.bestaetigen.einbinden,
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Liegt schon da.',
    )
  })

  it('löscht ein Paket nach dem Plan und lädt nach', async () => {
    const Pakete = vi
      .fn()
      .mockResolvedValueOnce(katalog())
      .mockResolvedValue({ pakete: [], befunde: [] })
    const PaketPlan = vi.fn(() => Promise.resolve(plan('entfernen')))
    const PaketEntfernen = vi.fn(() => Promise.resolve({}))

    belegeApi({ PaketArten: anlegbar, Pakete, PaketPlan, PaketEntfernen })
    zeige()

    await screen.findByText('probe-schlicht')
    await entferne('probe-schlicht')

    expect(PaketPlan).toHaveBeenCalledWith(
      'entfernen',
      'mse-style',
      'probe-schlicht',
      '',
    )
    expect(PaketEntfernen).toHaveBeenCalledWith('mse-style', 'probe-schlicht')
    await waitFor(() => {
      expect(
        screen.getByText(texte.seiten.pakete.katalogLeer),
      ).toBeInTheDocument()
    })
  })

  // **Nach einem Fehler wird trotzdem nachgeladen** (§7, Regel 3) — Entfernen
  // löscht **vor** dem Ausbuchen; scheitert der Katalog danach, hat sich die
  // Platte trotzdem geändert (Zweitblick über Phase 7).
  it('meldet einen Fehler beim Löschen, lässt die Liste stehen und lädt nach', async () => {
    const Pakete = vi.fn().mockResolvedValue(katalog())

    belegeApi({
      PaketArten: anlegbar,
      Pakete,
      PaketPlan: () => Promise.resolve(plan('entfernen')),
      PaketEntfernen: () =>
        Promise.resolve({ fehler: fehler('Der Katalog schweigt.') }),
    })
    zeige()

    await screen.findByText('probe-schlicht')
    expect(Pakete).toHaveBeenCalledTimes(1)

    await entferne('probe-schlicht')

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Katalog schweigt.',
    )
    expect(screen.getByText('probe-schlicht')).toBeInTheDocument()
    await waitFor(() => {
      expect(Pakete).toHaveBeenCalledTimes(2)
    })
  })

  // **Ein abgebrochener Dialog plant nichts und lädt nicht nach.**
  it('lädt nicht nach, wenn der Dialog abgebrochen wurde', async () => {
    const Pakete = vi.fn().mockResolvedValue(katalog())
    const PaketPlanWaehlen = vi.fn(() =>
      Promise.resolve({ plan: null, abgebrochen: true }),
    )

    belegeApi({ PaketArten: anlegbar, Pakete, PaketPlanWaehlen })
    zeige()

    await screen.findByText('probe-schlicht')
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.pakete.plan.einbinden }),
    )

    await waitFor(() => {
      expect(PaketPlanWaehlen).toHaveBeenCalledTimes(1)
    })
    expect(Pakete).toHaveBeenCalledTimes(1)
    expect(screen.queryByRole('group', { name: /einbinden/ })).toBeNull()
  })
})

// **Ein Abbruch löscht nichts.**
it('lässt ein Paket stehen, wenn der Plan verworfen wird', async () => {
  const PaketEntfernen = vi.fn(() => Promise.resolve({}))

  belegeApi({
    PaketArten: anlegbar,
    Pakete: vi.fn().mockResolvedValue(katalog()),
    PaketPlan: () => Promise.resolve(plan('entfernen')),
    PaketEntfernen,
  })
  zeige()

  await screen.findByText('probe-schlicht')
  // **Erst wählen** — seit UXR-17 (#1925) steht das Menü nur an der
  // aktiven Zeile.
  await userEvent.click(screen.getByRole('row', { name: 'probe-schlicht' }))
  await userEvent.click(
    screen.getByRole('button', { name: 'Handlungen — probe-schlicht' }),
  )
  await userEvent.click(
    screen.getByRole('menuitem', { name: texte.seiten.pakete.plan.loeschen }),
  )
  await userEvent.click(
    await screen.findByRole('button', {
      name: texte.seiten.pakete.plan.verwerfen,
    }),
  )

  expect(PaketEntfernen).not.toHaveBeenCalled()
  expect(screen.queryByText('Erstes')).toBeNull()
})

// kopfknopf ist „Paket einbinden …" im Kopf der Liste — im leeren Katalog
// steht derselbe Knopf noch im Leerzustand (G02, #2249).
function kopfknopf(): HTMLElement {
  const knopf = screen
    .getAllByRole('button', { name: texte.seiten.pakete.plan.einbinden })
    .find((einer) => einer.closest('[data-leerzustand]') === null)

  if (knopf === undefined) {
    throw new Error('im Kopf steht kein „Paket einbinden …"')
  }

  return knopf
}
