// **Was ein Auftrag draußen kosten darf** — Werkstück O12 von Phase 5.2
// (#1439).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Kosten } from './Kosten'
import { einsetzen } from '../sprache/einsetzen'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'

import type { HaushaltsstandAntwort, KostenpostenDTO } from '../apikosten'

const worte = kataloge.de.seiten.kosten

afterEach(() => {
  raeumeApi()
})

function zeige() {
  return render(
    <Sprachraum sprache="de">
      <Kosten />
    </Sprachraum>,
  )
}

function euro(wert: number) {
  return { wert, waehrung: 'EUR' }
}

function stand(
  teil: Partial<HaushaltsstandAntwort> = {},
): HaushaltsstandAntwort {
  return {
    konto: 'anbieter',
    zeitraum: 'tag',
    grenze: euro(50_000),
    reserviert: euro(3_000),
    verbraucht: euro(12_000),
    unklar: euro(0),
    verfuegbar: euro(35_000),
    fehler: null,
    ...teil,
  }
}

function posten(teil: Partial<KostenpostenDTO> = {}): KostenpostenDTO {
  return {
    kennung: 'eins',
    konto: 'anbieter',
    aufgabe: 'karte.text',
    hoechstbetrag: euro(20_000),
    betrag: euro(12_000),
    stand: 'abgerechnet',
    grund: '',
    begonnen: '2026-09-15T12:00:00Z',
    beendet: '2026-09-15T12:00:05Z',
    ...teil,
  }
}

describe('Kostenjournal', () => {
  it('zeigt die vier Zahlen und die Grenze', async () => {
    belegeApi({ Kostenstand: () => Promise.resolve(stand()) })

    zeige()

    expect(await screen.findByText('500,00 €')).toBeInTheDocument()
    expect(screen.getByText('350,00 €')).toBeInTheDocument()

    // **Ohne ungeklärte Menge steht der Hinweis dazu nicht da** — er erklärt
    // etwas, das es gerade nicht gibt.
    expect(screen.queryByText(kurzform(worte.unklarHinweis))).toBeNull()
  })

  it('erklärt die ungeklärte Menge, sobald es eine gibt', async () => {
    belegeApi({
      Kostenstand: () => Promise.resolve(stand({ unklar: euro(500) })),
    })

    zeige()

    expect(
      await screen.findByText(kurzform(worte.unklarHinweis)),
    ).toBeInTheDocument()
  })

  it('sagt es, wenn nichts freigegeben ist', async () => {
    belegeApi({
      Kostenstand: () =>
        Promise.resolve(stand({ grenze: euro(0), verfuegbar: euro(0) })),
    })

    zeige()

    expect(await screen.findByText(worte.ohneGrenze)).toBeInTheDocument()
  })

  it('setzt eine Grenze und lädt danach nach', async () => {
    const setzen = vi.fn().mockResolvedValue({ fehler: null })
    const lesen = vi.fn().mockResolvedValue(stand())

    belegeApi({ Kostenstand: lesen, KostenbudgetSetzen: setzen })

    const nutzer = userEvent.setup()

    zeige()

    await screen.findByText('500,00 €')

    await nutzer.type(
      screen.getByRole('textbox', { name: worte.betragTippen }),
      '70000',
    )
    await nutzer.click(screen.getByRole('button', { name: worte.setzen }))

    await waitFor(() => {
      expect(setzen).toHaveBeenCalledWith('anbieter', 'tag', 70_000, 'EUR')
    })

    // **§7, Regel 3**: nach dem Schreiben wird nachgeladen.
    await waitFor(() => {
      expect(lesen).toHaveBeenCalledTimes(2)
    })
  })

  it('nimmt eine andere Währung entgegen', async () => {
    const setzen = vi.fn().mockResolvedValue({ fehler: null })

    belegeApi({
      Kostenstand: () => Promise.resolve(stand()),
      KostenbudgetSetzen: setzen,
    })

    const nutzer = userEvent.setup()

    zeige()

    await screen.findByText('500,00 €')

    await nutzer.type(
      screen.getByRole('textbox', { name: worte.betragTippen }),
      '1000',
    )
    // **Ein Feld mit Vorschlägen ist eine Combobox** (HTML-AAM: `input` mit
    // `list`) — UX-06 (#1678).
    await nutzer.clear(
      screen.getByRole('combobox', { name: worte.waehrungTippen }),
    )
    await nutzer.type(
      screen.getByRole('combobox', { name: worte.waehrungTippen }),
      'USD',
    )
    await nutzer.click(screen.getByRole('button', { name: worte.setzen }))

    await waitFor(() => {
      expect(setzen).toHaveBeenCalledWith('anbieter', 'tag', 1000, 'USD')
    })
  })

  // **Klein geschrieben wird erklärt und nicht hinübergeschickt** — UX-06
  // (#1678): `Betrag.Pruefe` wiese `usd` ab, und das Feld sagt vorher, wie
  // der Code heißt. Leer geht ebenso wenig hinüber.
  it('lässt eine Währung ohne Gestalt nicht setzen und sagt es am Feld', async () => {
    const setzen = vi.fn()

    belegeApi({
      Kostenstand: () => Promise.resolve(stand()),
      KostenbudgetSetzen: setzen,
    })

    const nutzer = userEvent.setup()

    zeige()

    await screen.findByText('500,00 €')

    await nutzer.type(
      screen.getByRole('textbox', { name: worte.betragTippen }),
      '1000',
    )

    const waehrung = screen.getByRole('combobox', {
      name: worte.waehrungTippen,
    })

    await nutzer.clear(waehrung)
    expect(screen.getByRole('button', { name: worte.setzen })).toBeDisabled()

    await nutzer.type(waehrung, 'usd')
    await nutzer.tab()

    expect(screen.getByRole('button', { name: worte.setzen })).toBeDisabled()
    expect(waehrung).toHaveAttribute('aria-invalid', 'true')
    expect(screen.getByRole('alert')).toHaveTextContent(
      einsetzen(kataloge.de.bausteine.waehrung.grossSchreiben, {
        wert: 'usd',
        gross: 'USD',
      }),
    )
    expect(setzen).not.toHaveBeenCalled()
  })

  // **»1.000« ist kein Betrag in der kleinsten Einheit** — für `Number`
  // wäre es eins, und die Grenze stünde auf einem Cent (Nachlauf über Phase
  // 6.3, N5). Das Feld sagt, warum der Knopf aus bleibt.
  it('lässt nichts setzen, was keine ganze Zahl ist, und sagt es am Feld', async () => {
    belegeApi({ Kostenstand: () => Promise.resolve(stand()) })

    const nutzer = userEvent.setup()

    zeige()

    await screen.findByText('500,00 €')

    const knopf = screen.getByRole('button', { name: worte.setzen })
    const feld = screen.getByRole('textbox', { name: worte.betragTippen })
    expect(knopf).toBeDisabled()

    for (const eingabe of ['12,50', '1.000']) {
      await nutzer.clear(feld)
      await nutzer.type(feld, eingabe)

      expect(knopf).toBeDisabled()
      expect(feld).toHaveAccessibleDescription(kataloge.de.allgemein.nurZiffern)
    }
  })

  it('zeigt die letzten Aufträge mit ihrem Stand', async () => {
    belegeApi({
      Kostenstand: () => Promise.resolve(stand()),
      Kostenposten: () =>
        Promise.resolve({
          posten: [
            posten(),
            posten({ kennung: 'zwei', stand: 'reserviert', betrag: euro(0) }),
          ],
          fehler: null,
        }),
    })

    zeige()

    expect(
      await screen.findByText(/karte\.text · abgerechnet · 120,00 €/),
    ).toBeInTheDocument()

    // **Ein angemeldeter Auftrag zeigt seinen Höchstbetrag** — sein Betrag
    // steht noch nicht fest.
    expect(
      screen.getByText(/karte\.text · angemeldet · 200,00 €/),
    ).toBeInTheDocument()
  })

  it('zeigt einen unbekannten Stand als Marke, statt ihn zu verschlucken', async () => {
    // **Der Stand kommt von der Go-Seite** — wächst die Aufzählung, bevor der
    // Katalog nachzieht, steht die Marke da. Eine leere Stelle wäre die
    // gefährlichere Anzeige.
    belegeApi({
      Kostenstand: () => Promise.resolve(stand()),
      Kostenposten: () =>
        Promise.resolve({
          posten: [posten({ stand: 'verhandelt' })],
          fehler: null,
        }),
    })

    zeige()

    expect(
      await screen.findByText(/karte\.text · verhandelt · 200,00 €/),
    ).toBeInTheDocument()
  })

  it('sagt es, wenn noch nichts gebucht ist', async () => {
    belegeApi({ Kostenstand: () => Promise.resolve(stand()) })

    zeige()

    expect(await screen.findByText(worte.keinePosten)).toBeInTheDocument()
  })

  it('holt für ein anderes Konto neue Zahlen', async () => {
    const lesen = vi.fn().mockResolvedValue(stand())

    belegeApi({ Kostenstand: lesen })

    const nutzer = userEvent.setup()

    zeige()

    await screen.findByText('500,00 €')

    await nutzer.clear(
      screen.getByRole('textbox', { name: worte.kontoWaehlen }),
    )
    await nutzer.type(
      screen.getByRole('textbox', { name: worte.kontoWaehlen }),
      'zweiter',
    )

    await waitFor(() => {
      expect(lesen).toHaveBeenLastCalledWith('zweiter', 'tag')
    })
  })

  it('holt für einen anderen Zeitraum neue Zahlen', async () => {
    const lesen = vi.fn().mockResolvedValue(stand())

    belegeApi({ Kostenstand: lesen })

    const nutzer = userEvent.setup()

    zeige()

    await screen.findByText('500,00 €')

    await nutzer.selectOptions(
      screen.getByRole('combobox', { name: worte.zeitraumWaehlen }),
      'monat',
    )

    await waitFor(() => {
      expect(lesen).toHaveBeenLastCalledWith('anbieter', 'monat')
    })
  })

  it('meldet einen Fehler beim Setzen und lässt die Zahlen stehen', async () => {
    belegeApi({
      Kostenstand: () => Promise.resolve(stand()),
      KostenbudgetSetzen: () => Promise.reject(new Error('das ging nicht')),
    })

    const nutzer = userEvent.setup()

    zeige()

    await screen.findByText('500,00 €')

    await nutzer.type(
      screen.getByRole('textbox', { name: worte.betragTippen }),
      '70000',
    )
    await nutzer.click(screen.getByRole('button', { name: worte.setzen }))

    expect(await screen.findByRole('alert')).toHaveTextContent('das ging nicht')
    expect(screen.getByText('500,00 €')).toBeInTheDocument()
  })

  it('meldet einen Fehler beim Lesen', async () => {
    belegeApi({
      Kostenstand: () => Promise.reject(new Error('das Journal ist zu')),
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'das Journal ist zu',
    )
  })

  it('schreibt einen Betrag ohne Währung als blanke Zahl', async () => {
    belegeApi({
      Kostenstand: () =>
        Promise.resolve(
          stand({
            grenze: { wert: 0, waehrung: '' },
            verfuegbar: { wert: 0, waehrung: '' },
            reserviert: { wert: 0, waehrung: '' },
            verbraucht: { wert: 0, waehrung: '' },
            unklar: { wert: 0, waehrung: '' },
          }),
        ),
    })

    zeige()

    // **Ohne Freigabe gibt es keine Währung** — und eine erfundene wäre
    // schlimmer als eine blanke Null.
    await waitFor(() => {
      expect(screen.getAllByText('0')).toHaveLength(5)
    })
  })
})
