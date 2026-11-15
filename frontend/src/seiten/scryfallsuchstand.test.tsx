// **Was die Suche über ihren Stand sagt** — die Befunde 4 und 5 des
// Zweitblicks über Phase 3.6
// (#1167).
//
// **Eine eigene Datei**, weil `scryfallsuche.test.tsx` mit ihr über die
// Zeilengrenze aus §2 ginge — und weil der Schnitt einer Frage folgt: dort
// steht, **was** die Suche findet, hier, **was sie über ihren Fund sagt**.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { SuchtrefferDTO } from '../apiscryfall'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Scryfallsuche } from './Scryfallsuche'

const texte = kataloge.de.seiten.karten.scryfallsuche

function treffer(teile: Partial<SuchtrefferDTO> = {}): SuchtrefferDTO {
  return {
    kennung: 'abc',
    name: 'Lightning Bolt',
    setcode: 'lea',
    setname: 'Limited Edition Alpha',
    nummer: '161',
    seltenheit: 'common',
    manakosten: '{R}',
    typzeile: 'Instant',
    bild: '',
    ...teile,
  }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Scryfallsuche waehlen={vi.fn()} />
    </Sprachraum>,
  )
}

afterEach(() => {
  raeumeApi()
})

// **Der Stand sagt, wie alt er ist** — Befund 4 des Zweitblicks über
// Phase 3.6.
//
// Alter und Befunde reisen durch `service` und `app` bis an die Fassade und
// wurden im Fenster weggeworfen: eine Antwort aus dem Zwischenspeicher sah aus
// wie eine frische. §16, Regel 5.
describe('was die Suche über ihren Stand sagt', () => {
  it('nennt das Alter einer Antwort aus dem Zwischenspeicher', async () => {
    belegeApi({
      ScryfallSuche: vi.fn().mockResolvedValue({
        treffer: [treffer()],
        weitere: false,
        alterStunden: 9,
        befunde: [],
        fehler: null,
      }),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    expect(await screen.findByText(/9 Stunden alt/)).toBeInTheDocument()
  })

  it('sagt bei einer frischen Antwort, dass sie frisch ist', async () => {
    belegeApi({
      ScryfallSuche: vi.fn().mockResolvedValue({
        treffer: [treffer()],
        weitere: false,
        alterStunden: 0,
        befunde: [],
        fehler: null,
      }),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    expect(await screen.findByText(texte.frisch)).toBeInTheDocument()
  })

  // **Und die Befunde reisen bis hierher** (§8).
  it('zeigt die Befunde der Antwort', async () => {
    belegeApi({
      ScryfallSuche: vi.fn().mockResolvedValue({
        treffer: [treffer()],
        weitere: false,
        alterStunden: 0,
        befunde: ['layout »zerplatzt« ist unbekannt'],
        fehler: null,
      }),
    })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    expect(await screen.findByText(/zerplatzt/)).toBeInTheDocument()
  })

  // **Eine gescheiterte Suche räumt die alten Treffer** — Befund 5.
  //
  // Sie stehen zu lassen hieße, eine gescheiterte Suche mit dem Ergebnis der
  // vorigen zu beantworten; wer den Fehler übersieht, klickt einen Treffer an,
  // den er gar nicht gesucht hat.
  it('räumt die Treffer, wenn die nächste Suche scheitert', async () => {
    const gesucht = vi
      .fn()
      .mockResolvedValueOnce({
        treffer: [treffer()],
        weitere: false,
        alterStunden: 0,
        befunde: [],
        fehler: null,
      })
      .mockResolvedValue({
        treffer: [],
        weitere: false,
        alterStunden: 0,
        befunde: [],
        fehler: {
          code: 'unbekannt',
          schluessel: 'unbekannt',
          meldung: 'die Gegenstelle antwortet nicht',
        },
      })

    belegeApi({ ScryfallSuche: gesucht })

    zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    expect(
      await screen.findByRole('button', { name: /Lightning Bolt/ }),
    ).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    expect(await screen.findByRole('alert')).toBeInTheDocument()

    expect(
      screen.queryByRole('button', { name: /Lightning Bolt/ }),
    ).not.toBeInTheDocument()
  })
})

// **Eine gescheiterte Suche nimmt auch die gehobene Wahl zurück** — Befund 1
// des Nachlaufs über Phase 3.6.
//
// Der erste Durchgang räumte nur die **eigene** Trefferliste. Der Treffer, den
// die Suche vorher nach oben gegeben hatte, blieb dort stehen — und der
// Verweiskasten bot ihn weiter an, obwohl die Liste, aus der er stammt, nicht
// mehr existiert.
describe('was eine gescheiterte Suche nach oben meldet', () => {
  it('nimmt die gehobene Wahl zurück', async () => {
    const gesucht = vi
      .fn()
      .mockResolvedValueOnce({
        treffer: [treffer()],
        weitere: false,
        alterStunden: 0,
        befunde: [],
        fehler: null,
      })
      .mockResolvedValue({
        treffer: [],
        weitere: false,
        alterStunden: 0,
        befunde: [],
        fehler: {
          code: 'unbekannt',
          schluessel: 'unbekannt',
          meldung: 'die Gegenstelle antwortet nicht',
        },
      })

    belegeApi({ ScryfallSuche: gesucht })

    const waehlen = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Scryfallsuche waehlen={waehlen} />
      </Sprachraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    await userEvent.click(
      await screen.findByRole('button', { name: /Lightning Bolt/ }),
    )

    expect(waehlen).toHaveBeenLastCalledWith(
      expect.objectContaining({ name: 'Lightning Bolt' }),
    )

    await userEvent.click(screen.getByRole('button', { name: texte.suchen }))

    await waitFor(() => {
      expect(waehlen).toHaveBeenLastCalledWith(undefined)
    })
  })
})
