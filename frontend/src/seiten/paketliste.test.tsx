// Der Katalog der installierten Pakete (#54) — seit DOM-01 ein Arbeitsplatz,
// und jede Änderung beginnt mit einem Plan
// (#1696).
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { anlegbar, katalog, plan, texte, zeige } from '../test/paketliste'

afterEach(() => {
  raeumeApi()
})

describe('Installierte Pakete', () => {
  it('zeigt Name, Art, Fassung und Herkunft — und das Übrige im Inspector', async () => {
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn(() => Promise.resolve(katalog())),
    })
    zeige()

    expect(await screen.findByText('probe-schlicht')).toBeInTheDocument()

    // **Über die Zelle und nicht über den Text.** `mse-style` steht seit #244
    // zweimal auf der Seite: in der Zeile und im Artfilter.
    expect(screen.getByRole('cell', { name: 'mse-style' })).toBeInTheDocument()
    expect(screen.getByRole('cell', { name: '2026-08-15' })).toBeInTheDocument()

    // **Der Stand steht als Chip, die Herkunftsart in seinem `title`** —
    // ganz steht sie im Inspector (#56, §11, DOM-01).
    expect(
      screen.getByRole('cell', {
        name: texte.seiten.pakete.plan.standEingebunden,
      }),
    ).toHaveAttribute(
      'title',
      texte.seiten.pakete.plan.herkunftsarten.mitgeliefert,
    )

    await userEvent.click(screen.getByRole('row', { name: 'probe-schlicht' }))

    expect(screen.getByRole('row', { name: 'probe-schlicht' })).toHaveAttribute(
      'aria-selected',
      'true',
    )
    expect(
      screen.getByRole('heading', { name: 'probe-schlicht' }),
    ).toBeInTheDocument()

    // **Und die Prüfsumme dazu** — angezeigt *und* vermerkt (#56).
    expect(screen.getByText('sha256:abc')).toBeInTheDocument()
  })

  it('erklärt den leeren Zustand, statt „keine Daten“ zu zeigen', async () => {
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn(() => Promise.resolve({ pakete: [], befunde: [] })),
    })
    zeige()

    expect(
      await screen.findByText(texte.seiten.pakete.katalogLeer),
    ).toBeInTheDocument()
  })

  it('nennt die Befunde des Abgleichs', async () => {
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn(() =>
        Promise.resolve({
          pakete: [],
          befunde: ['kaputt.mse-style: nicht lesbar'],
        }),
      ),
    })
    zeige()

    expect(
      await screen.findByText('kaputt.mse-style: nicht lesbar'),
    ).toBeInTheDocument()
  })

  it('meldet einen Fehler beim Laden am Ort', async () => {
    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn(() =>
        Promise.resolve({
          pakete: [],
          befunde: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'fehler.unbekannt',
            meldung: 'Der Bestand schweigt.',
          },
        }),
      ),
    })
    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Bestand schweigt.',
    )
  })

  // **Ein Wechsel des Pakets räumt den Plan** (§7, Regel 1) — und der
  // Inspector zeigt danach das gewählte.
  it('räumt den Plan, wenn ein anderes Paket gewählt wird', async () => {
    const zwei = katalog()
    zwei.pakete.push(
      ...katalog().pakete.map((paket) => ({
        ...paket,
        name: 'probe-bild',
        anzeigename: 'probe-bild',
        pfad: '/pakete/probe-bild.mse-style',
      })),
    )

    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn().mockResolvedValue(zwei),
      PaketPlan: () => Promise.resolve(plan('entfernen')),
    })
    zeige()

    await screen.findByText('probe-bild')
    // **Erst wählen** — seit UXR-17 (#1925) steht das Menü nur an der
    // aktiven Zeile.
    await userEvent.click(screen.getByRole('row', { name: 'probe-schlicht' }))
    await userEvent.click(
      screen.getByRole('button', { name: 'Handlungen — probe-schlicht' }),
    )
    await userEvent.click(
      screen.getByRole('menuitem', { name: texte.seiten.pakete.plan.zurueck }),
    )
    expect(await screen.findByText('Erstes')).toBeInTheDocument()

    // **Und dasselbe Paket noch einmal gewählt lässt ihn stehen.**
    await userEvent.click(screen.getByRole('row', { name: 'probe-schlicht' }))
    expect(screen.getByText('Erstes')).toBeInTheDocument()

    screen.getByRole('row', { name: 'probe-bild' }).focus()
    await userEvent.keyboard('{Enter}')

    expect(screen.queryByText('Erstes')).toBeNull()
    expect(
      screen.getByRole('heading', { name: 'probe-bild' }),
    ).toBeInTheDocument()
  })

  // **Der Stand steht an der Zeile** — wie im Artboard ein Chip; eine
  // lokale Änderung ist dort zu sehen, ohne das Paket zu wählen (AK-01).
  it('zeigt eine lokale Änderung an der Zeile', async () => {
    const geaendert = katalog()
    geaendert.pakete = geaendert.pakete.map((paket) => ({
      ...paket,
      geaendert: true,
    }))

    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn().mockResolvedValue(geaendert),
    })
    zeige()

    expect(
      await screen.findByRole('cell', {
        name: texte.seiten.pakete.plan.standGeaendert,
      }),
    ).toBeInTheDocument()
  })

  it('plant Löschen und Zurück auch aus dem Inspector', async () => {
    const PaketPlan = vi.fn(() => Promise.resolve(plan('entfernen')))

    belegeApi({
      PaketArten: anlegbar,
      Pakete: vi.fn().mockResolvedValue(katalog()),
      PaketPlan,
    })
    zeige()

    await userEvent.click(
      await screen.findByRole('row', { name: 'probe-schlicht' }),
    )
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.pakete.plan.zurueck }),
    )
    expect(PaketPlan).toHaveBeenLastCalledWith(
      'zurueck',
      'mse-style',
      'probe-schlicht',
      '',
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.pakete.plan.verwerfen,
      }),
    )
    await userEvent.click(
      screen.getByRole('button', {
        name: new RegExp(texte.seiten.pakete.plan.loeschen),
      }),
    )
    expect(PaketPlan).toHaveBeenLastCalledWith(
      'entfernen',
      'mse-style',
      'probe-schlicht',
      '',
    )
  })
})
