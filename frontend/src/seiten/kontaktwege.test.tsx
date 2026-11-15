// **Kontakte aus Apple Kontakte** — ID-04, Nachgang zum Zweitblick
// (#1703).
import { screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KontaktDTO } from '../apikontakte'
import { einsetzen } from '../sprache/einsetzen'
import { raeumeApi } from '../test/api'
import { belege, karten, kontakt, texte, zeige } from '../test/kontakte'

describe('Kontakte aus Apple Kontakte, nach dem Zweitblick', () => {
  afterEach(raeumeApi)

  // **Während eine Übernahme läuft, wird nicht neu eingelesen** — Befund 3
  // des Zweitblicks über ID-04: sonst räumte die späte Antwort der
  // Übernahme eine Vorschau ab, die zu ihr nicht gehört (§7, Regel 2).
  it('liest nicht ein, solange eine Übernahme läuft', async () => {
    let weiter: (() => void) | undefined
    const eingelesen = vi.fn(() =>
      Promise.resolve({
        karten: karten(),
        befunde: [],
        abgebrochen: false,
        fehler: null,
      }),
    )
    belege({
      KontakteEinlesen: eingelesen,
      KontakteUebernehmen: () =>
        new Promise((los) => {
          weiter = () => {
            los({ angelegt: 1, ersetzt: 1, fehler: null })
          }
        }),
    })
    await zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.waehlen }))
    await userEvent.click(
      await screen.findByRole('button', {
        name: einsetzen(texte.uebernehmen, { anzahl: '2' }),
      }),
    )

    expect(screen.getByRole('button', { name: texte.waehlen })).toBeDisabled()
    expect(eingelesen).toHaveBeenCalledTimes(1)

    weiter?.()
    await waitFor(() => {
      expect(screen.getByRole('button', { name: texte.waehlen })).toBeEnabled()
    })
  })

  // **Eine Vermutung ersetzt, wenn man sie wählt** — Befund 1: neu angelegt
  // wäre sie genau die Dublette, die AK-02 ausschließt.
  it('übernimmt eine Vermutung als Ersatz des vorhandenen Kontakts', async () => {
    const uebernommen = vi.fn<
      (auswahl: KontaktDTO[]) => Promise<{
        angelegt: number
        ersetzt: number
        fehler: null
      }>
    >(() => Promise.resolve({ angelegt: 0, ersetzt: 1, fehler: null }))
    belege({
      KontakteEinlesen: () =>
        Promise.resolve({
          karten: [
            {
              kontakt: kontakt({ ursprung: '' }),
              abgleich: 'vermutlich',
              vorhanden: kontakt({ id: 5, fassung: 7 }),
            },
          ],
          befunde: [],
          abgebrochen: false,
          fehler: null,
        }),
      KontakteUebernehmen: uebernommen,
    })
    await zeige()

    await userEvent.click(screen.getByRole('button', { name: texte.waehlen }))
    const vorschau = within(
      await screen.findByRole('region', { name: texte.vorschauTitel }),
    )
    await userEvent.click(
      vorschau.getByRole('checkbox', { name: /Ada Lovelace/ }),
    )
    await userEvent.click(
      vorschau.getByRole('button', {
        name: einsetzen(texte.uebernehmen, { anzahl: '1' }),
      }),
    )

    await waitFor(() => {
      expect(uebernommen).toHaveBeenCalled()
    })
    expect(uebernommen.mock.calls[0]?.[0][0]).toMatchObject({
      id: 5,
      fassung: 7,
    })
  })
})
