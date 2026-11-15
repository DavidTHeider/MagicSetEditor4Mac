// **Wo ein Gegenüber erreichbar ist** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Was hier gemessen wird
//
// **Dass die Adresse an jedem Kontakt steht — auch an einem unbestätigten.**
// Sie erlaubt nichts; wer sie an die Bestätigung bände, könnte ein frisch
// eingetragenes Gegenüber nicht einmal anrufen, um seinen Abdruck zu
// vergleichen.
//
// **Und dass das Löschen eine Angabe ist.** Ein leeres Feld und „Merken"
// schickt eine leere Adresse; bliebe die alte stehen, ginge der nächste
// Versuch an einen Rechner, der inzwischen jemand anderem gehört.
import { screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { einKontakt, eineListe, zeige } from './peerkontaktprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.technik

// **Hinter dem Aufklapper „Adresse und Notiz"** (UXR-16, #1924): offen trug
// jede Zeile beide Felder samt „Merken" und Hilfe; jetzt öffnet der Lauf den
// Kasten, wie ein Mensch es tut.
const aufklapper = eigene.gegenueber.adresseUndNotiz

describe('Wo ein Gegenüber erreichbar ist', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('zeigt die eingetragene Adresse', async () => {
    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(
          eineListe([einKontakt({ adresse: '192.168.1.7:41234' })]),
        ),
    })

    await zeige(aufklapper)

    expect(await screen.findByLabelText(eigene.kontakteAdresse)).toHaveValue(
      '192.168.1.7:41234',
    )
  })

  // **Auch an einem unbestätigten** — das ist die Entscheidung.
  it('steht auch an einem unbestätigten Gegenüber', async () => {
    belegeApi({ Peerkontakte: () => Promise.resolve(eineListe()) })

    await zeige(aufklapper)

    expect(
      await screen.findByText(eigene.kontakteStandUnbestaetigt),
    ).toBeInTheDocument()

    expect(
      screen.getByRole('button', { name: eigene.kontakteAdresseMerken }),
    ).toBeInTheDocument()
  })

  it('hält den Knopf an, solange sich nichts geändert hat', async () => {
    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(eineListe([einKontakt({ adresse: 'beispiel:443' })])),
    })

    await zeige(aufklapper)

    expect(
      await screen.findByRole('button', {
        name: eigene.kontakteAdresseMerken,
      }),
    ).toBeDisabled()
  })

  it('schickt die getippte Adresse', async () => {
    const gerufen = vi.fn(() => Promise.resolve(eineListe()))

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe()),
      PeerkontaktAdresse: gerufen,
    })

    const hand = await zeige(aufklapper)

    await hand.type(
      await screen.findByLabelText(eigene.kontakteAdresse),
      '192.168.1.7:41234',
    )
    await hand.click(
      screen.getByRole('button', { name: eigene.kontakteAdresseMerken }),
    )

    expect(gerufen).toHaveBeenCalledWith('bgegenueber', '192.168.1.7:41234')
  })

  // **Der Lauf, für den der leere Wert eine Angabe ist.**
  it('schickt eine leere Adresse, wenn das Feld geräumt wird', async () => {
    const gerufen = vi.fn(() => Promise.resolve(eineListe()))

    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(eineListe([einKontakt({ adresse: 'beispiel:443' })])),
      PeerkontaktAdresse: gerufen,
    })

    const hand = await zeige(aufklapper)

    await hand.clear(await screen.findByLabelText(eigene.kontakteAdresse))
    await hand.click(
      screen.getByRole('button', { name: eigene.kontakteAdresseMerken }),
    )

    expect(gerufen).toHaveBeenCalledWith('bgegenueber', '')
  })

  // **Und der neue Stand von außen räumt den Entwurf** (§7, Regel 1) — über
  // den `key` und nicht über einen Effekt.
  it('übernimmt den Stand, der nach dem Schreiben zurückkommt', async () => {
    let gesetzt = ''

    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(eineListe([einKontakt({ adresse: gesetzt })])),
      PeerkontaktAdresse: (_kennung: string, adresse: string) => {
        // **Das Backend ändert sie ab** — so, wie es eine Normalisierung
        // täte. Was danach dasteht, muss der Stand sein und nicht der
        // Entwurf.
        gesetzt = adresse.toLowerCase()

        return Promise.resolve(eineListe([einKontakt({ adresse: gesetzt })]))
      },
    })

    const hand = await zeige(aufklapper)

    await hand.type(
      await screen.findByLabelText(eigene.kontakteAdresse),
      'MEIN-MAC.local:443',
    )
    await hand.click(
      screen.getByRole('button', { name: eigene.kontakteAdresseMerken }),
    )

    // **Gewartet wird auf den Wert** und nicht auf das Feld (§3): seit NET-01
    // bleibt die Liste beim Nachladen stehen, und das Feld mit dem Entwurf
    // ist sofort zu finden — erst der `key` tauscht es gegen den Stand.
    await waitFor(() => {
      expect(screen.getByLabelText(eigene.kontakteAdresse)).toHaveValue(
        'mein-mac.local:443',
      )
    })
  })
})
