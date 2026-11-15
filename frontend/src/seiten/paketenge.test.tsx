// **Bei Enge geht der Inspector der Pakete auf, wenn jemand wählt** —
// UXR-17 (#1925),
// Befund 1 aus der Hüllenprobe `fmt02e`: bei 1280 × 820 Punkten weicht der
// Inspector, und weder eine gewählte Zeile noch *Löschen …* im Menü der
// Zeile zeigte etwas — der Plan stand in einer Spalte, die niemand sah.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { effekteNachholen } from '../test/effekte'
import { anlegbar, katalog, plan, texte, zeige } from '../test/paketliste'
import { gibJsdomEineBreite } from '../test/sichtbreite'

let raeumeBreite: () => void = () => undefined

beforeEach(() => {
  raeumeBreite = gibJsdomEineBreite(1000)
})

afterEach(() => {
  raeumeBreite()
  raeumeApi()
})

const inspector = () =>
  screen.queryByRole('complementary', {
    name: texte.seiten.pakete.plan.inspector,
  })

// aufgegangen wartet auf die Spalte — der Plan kommt eine Antwort später.
const aufgegangen = () =>
  screen.findByRole('complementary', {
    name: texte.seiten.pakete.plan.inspector,
  })

describe('Der Paketinspector bei Enge', () => {
  it('geht über dem Zentrum auf, wenn eine Zeile gewählt wird', async () => {
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn().mockResolvedValue(katalog()),
    })
    zeige()
    await effekteNachholen()

    const zeile = await screen.findByRole('row', { name: 'probe-schlicht' })
    expect(inspector()).toBeNull()

    await userEvent.click(zeile)

    expect(inspector()).toHaveAttribute('data-ueber', 'zentrum')
    expect(
      within(await aufgegangen()).getByRole('heading', {
        name: 'probe-schlicht',
      }),
    ).toBeInTheDocument()
  })

  it('zeigt den Plan, den das Menü der Zeile öffnet', async () => {
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn().mockResolvedValue(katalog()),
      PaketPlan: () => Promise.resolve(plan('entfernen')),
    })
    zeige()
    await effekteNachholen()

    // **Mit der Tastatur**, damit die Wahl nicht schon über den Klick auf
    // die Zeile kommt: ↓ auf das Paket, → auf sein Menü.
    const zeilen = (await screen.findAllByRole('row')).filter(
      (eine) => eine.closest('tbody') !== null,
    )
    zeilen[0]?.focus()
    await userEvent.keyboard('{ArrowDown}{ArrowRight}{Enter}')
    await userEvent.click(
      screen.getByRole('menuitem', { name: texte.seiten.pakete.plan.loeschen }),
    )

    expect(
      await within(await aufgegangen()).findByText('Erstes'),
    ).toBeInTheDocument()
  })

  it('zeigt den Plan zum Einbinden', async () => {
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn().mockResolvedValue(katalog()),
      PaketPlanWaehlen: () => Promise.resolve(plan('einbinden', '/neu')),
    })
    zeige()
    await effekteNachholen()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.pakete.plan.einbinden,
      }),
    )

    expect(
      await within(await aufgegangen()).findByText('Erstes'),
    ).toBeInTheDocument()
  })

  // **Ein abgebrochener Dialog ist kein Anlass** — die Spalte ginge sonst
  // leer auf, mit dem Satz „Wähle ein Paket“ über der Liste.
  it('bleibt zu, wenn das Einbinden abgebrochen wird', async () => {
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn().mockResolvedValue(katalog()),
      PaketPlanWaehlen: () =>
        Promise.resolve({ plan: null, abgebrochen: true }),
    })
    zeige()
    await effekteNachholen()

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.pakete.plan.einbinden,
      }),
    )
    await effekteNachholen()

    expect(inspector()).toBeNull()
  })
})
