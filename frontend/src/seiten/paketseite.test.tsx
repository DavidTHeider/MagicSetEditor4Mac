// **Die Gestalt der Paketseite** — UXR-17
// (#1925),
// nach dem Artboard `Pakete`: der Arbeitsplatz füllt die Seite, *Paket
// beziehen …* führt im Kopf und öffnet eine Schublade, und das Seltene steht
// eingeklappt darunter.
//
// Gemessen am 2026-10-02 an der Hüllenprobe `uirevue`: Rolltiefe **3,3** bei
// 1280 × 800 (R4 in `docs/UIREVIEW.md` lässt 2,5 zu) — unter dem Arbeitsplatz
// standen sechs Abschnitte mit elf Feldern offen untereinander.
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { paketDTO } from '../test/vorlagenprobe'
import { raeumeApi } from '../test/api'
import {
  belegeSeite,
  oeffneBezug,
  oeffneKasten,
  texte,
  zeige,
} from '../test/paketseite'

afterEach(() => {
  raeumeApi()
})

const t = texte.seiten.pakete

describe('Die Paketseite', () => {
  it('klappt das Seltene unter dem Arbeitsplatz ein', async () => {
    belegeSeite({})
    zeige()

    await screen.findByText(t.katalogLeer)

    for (const titel of [
      t.anlegen,
      t.abhaengigUeberschrift,
      t.sammeldaten.ueberschrift,
    ]) {
      expect(screen.getByRole('button', { name: titel })).toHaveAttribute(
        'aria-expanded',
        'false',
      )
    }

    // **Zu heißt: nicht gezeichnet** — kein Feld der Anlage, kein Knopf der
    // Sammeldaten, bis jemand aufklappt.
    expect(screen.queryByLabelText(t.anlegenVollname)).toBeNull()
    expect(
      screen.queryByRole('button', { name: t.sammeldaten.holen }),
    ).toBeNull()

    await oeffneKasten(t.anlegen)

    expect(screen.getByLabelText(t.anlegenVollname)).toBeInTheDocument()
  })

  it('bezieht über eine Schublade aus dem Kopf', async () => {
    belegeSeite({})
    zeige()

    // **Das Formular steht nicht auf der Seite**, sondern hinter dem
    // führenden Knopf im Kopf — wie im Artboard.
    expect(screen.queryByLabelText(t.bezugAdresse)).toBeNull()

    const schublade = await oeffneBezug()

    expect(within(schublade).getByLabelText(t.bezugAdresse)).toBeInTheDocument()

    // **Übernahme und Umschlag liegen eingeklappt darunter** — eine Gruppe
    // offen, nicht drei (R3).
    for (const titel of [t.mse2Ueberschrift, t.umschlagTitel]) {
      expect(
        within(schublade).getByRole('button', { name: titel }),
      ).toHaveAttribute('aria-expanded', 'false')
    }
  })

  // **Schließen verwirft nichts** — S11.NAT-03 sagt es für das Aufklappen,
  // und eine Schublade, die eine halb getippte Adresse vergisst, wäre
  // schlimmer als ein Kasten, der es tut.
  it('behält eine Eingabe über das Schließen der Schublade hinweg', async () => {
    belegeSeite({})
    zeige()

    const schublade = await oeffneBezug()
    await userEvent.type(
      within(schublade).getByLabelText(t.bezugAdresse),
      'https://example.org/pack',
    )

    await userEvent.keyboard('{Escape}')

    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull()
    })
    expect(screen.getByRole('button', { name: t.plan.beziehen })).toHaveFocus()

    const wieder = await oeffneBezug()

    expect(within(wieder).getByLabelText(t.bezugAdresse)).toHaveValue(
      'https://example.org/pack',
    )
  })

  // **Unter der Schublade ist die Seite gesperrt** — `inert` wie bei den
  // Schubladen des Sets (`Setbereich.tsx`).
  it('sperrt die Seite, solange die Schublade steht', async () => {
    belegeSeite({})
    zeige()

    const knopf = await screen.findByRole('button', { name: t.plan.beziehen })
    expect(knopf.closest('[inert]')).toBeNull()

    await oeffneBezug()

    expect(knopf.closest('[inert]')).not.toBeNull()
  })
})

// **Eine Übernahme frischt die Liste daneben auf** (#55) — seit UXR-17 aus
// der Schublade heraus: die Liste darunter liest neu, auch wenn sie gerade
// verdeckt ist.
it('lädt den Katalog neu, nachdem Vorlagen übernommen wurden', async () => {
  const Pakete = vi
    .fn()
    .mockResolvedValueOnce({ pakete: [], befunde: [] })
    .mockResolvedValue({
      pakete: [
        { ...paketDTO('magic.mse-game'), name: 'magic', art: 'mse-game' },
      ],
      befunde: [],
    })

  belegeSeite({
    Pakete,
    VorlagenUebernehmen: () =>
      Promise.resolve({
        uebernommen: [],
        uebergangen: [],
        befunde: [],
        abgebrochen: false,
      }),
  })
  zeige()

  await screen.findByText(t.katalogLeer)
  await oeffneBezug()
  await oeffneKasten(t.mse2Ueberschrift)
  await userEvent.click(screen.getByRole('button', { name: t.mse2Waehlen }))

  expect(await screen.findByText('magic')).toBeInTheDocument()
})

// **Eine angenommene Vorlage lädt die Liste nach** —
// #1639.
//
// **Über die Seite und nicht über den Kasten allein**: gemessen wird, dass
// die Liste daneben den neuen Stand holt. Ein Lauf am Kasten prüfte nur, dass
// ein Rückruf läuft — nicht, dass er etwas bewirkt.
it('lädt die Paketliste nach einer angenommenen Vorlage nach', async () => {
  const gelesen = vi.fn(() => Promise.resolve({ pakete: [], befunde: [] }))

  belegeSeite({
    Pakete: gelesen,
    UmschlagVorlage: () =>
      Promise.resolve({
        paket: {
          ...paketDTO('angenommen.mse-game'),
          art: 'mse-game',
          name: 'angenommen',
          anzeigename: 'angenommen',
          version: '2026-09-19',
          pfad: '/pakete/angenommen.mse-game',
          artrang: 0,
          hauptdatei: 'game',
        },
        abgebrochen: false,
        fehler: null,
      }),
  })
  zeige()

  await screen.findByText(t.katalogLeer)
  const vorher = gelesen.mock.calls.length

  await oeffneBezug()
  await oeffneKasten(t.umschlagTitel)

  const hand = userEvent.setup()
  await hand.click(screen.getByLabelText(t.umschlagName))
  await hand.paste('angenommen.mse-game')
  await hand.click(screen.getByRole('button', { name: t.umschlagAnnehmen }))

  expect(await screen.findByText(t.umschlagAngenommen)).toBeInTheDocument()
  expect(gelesen.mock.calls.length).toBeGreaterThan(vorher)
})
