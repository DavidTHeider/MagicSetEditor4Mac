// **Was man sich zu einem Gegenüber notiert** — Werkstück T3 von Phase 5.6
// (#1468).
//
// # Was hier gemessen wird
//
// **Dass die Notiz an jedem Kontakt steht — auch an einem gesperrten.** Eine
// Sperre nimmt einem Gegenüber, was es darf; sie nimmt nicht, was man über es
// denkt.
//
// **Und dass das Löschen eine Angabe ist**: ein leeres Feld und „Notiz
// merken" schickt eine leere Notiz.
import { screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { einKontakt, eineListe, zeige } from './peerkontaktprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.technik

// **Hinter dem Aufklapper „Adresse und Notiz"** (UXR-16, #1924): offen trug
// jede Zeile beide Felder samt „Merken" und Hilfe; jetzt öffnet der Lauf den
// Kasten, wie ein Mensch es tut.
const aufklapper = eigene.gegenueber.adresseUndNotiz

describe('Die Notiz zu einem Gegenüber', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('zeigt die eingetragene Notiz', async () => {
    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(eineListe([einKontakt({ notiz: 'tauscht fair' })])),
    })

    await zeige(aufklapper)

    expect(await screen.findByLabelText(eigene.kontakteNotiz)).toHaveValue(
      'tauscht fair',
    )
  })

  // **Auch an einem gesperrten** — das ist die Entscheidung.
  it('steht auch an einem gesperrten Gegenüber', async () => {
    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(eineListe([einKontakt({ stand: 'gesperrt' })])),
    })

    await zeige(aufklapper)

    expect(
      await screen.findByRole('button', { name: eigene.kontakteNotizMerken }),
    ).toBeInTheDocument()
  })

  it('schickt die getippte Notiz', async () => {
    const notiert = vi.fn(() => Promise.resolve(eineListe()))

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe()),
      PeerkontaktNotiz: notiert,
    })

    const hand = await zeige(aufklapper)

    await hand.type(
      await screen.findByLabelText(eigene.kontakteNotiz),
      'antwortet spät',
    )

    await hand.click(
      screen.getByRole('button', { name: eigene.kontakteNotizMerken }),
    )

    expect(notiert).toHaveBeenCalledWith('bgegenueber', 'antwortet spät')
  })

  // **Ein leeres Feld löscht** — und der Knopf bleibt dafür bedienbar.
  it('schickt eine leere Notiz zum Löschen', async () => {
    const notiert = vi.fn(() => Promise.resolve(eineListe()))

    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(eineListe([einKontakt({ notiz: 'alt' })])),
      PeerkontaktNotiz: notiert,
    })

    const hand = await zeige(aufklapper)

    await hand.clear(await screen.findByLabelText(eigene.kontakteNotiz))

    await hand.click(
      screen.getByRole('button', { name: eigene.kontakteNotizMerken }),
    )

    expect(notiert).toHaveBeenCalledWith('bgegenueber', '')
  })

  // **Zwei Geschwister dürfen nicht denselben Schlüssel tragen** — T4, #1469.
  //
  // **Gefunden an einem Kontakt ohne Adresse und ohne Notiz**: beide Bauteile
  // hängen an einem `key`, der ihren Wert trägt, und leer war er derselbe.
  // React sagt dazu, das Verhalten sei **nicht unterstützt** — Kinder dürfen
  // dabei verdoppelt oder ausgelassen werden. Was hier auf dem Spiel steht,
  // ist §7 Regel 1 in ihrer unsichtbaren Gestalt: der `key` räumt das Bauteil
  // beim Wechsel, und zwei gleiche Schlüssel räumen möglicherweise das falsche.
  it('gibt Adresse und Notiz verschiedene Schlüssel', async () => {
    const klagen = vi
      .spyOn(console, 'error')
      .mockImplementation(() => undefined)

    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(
          eineListe([
            einKontakt({ stand: 'bestaetigt', adresse: '', notiz: '' }),
          ]),
        ),
      Sammlungsfreigaben: () =>
        Promise.resolve({ freigaben: [], umfaenge: [], fehler: null }),
    })

    await zeige(aufklapper)

    expect(
      await screen.findByLabelText(eigene.kontakteNotiz),
    ).toBeInTheDocument()

    const doppelt = klagen.mock.calls.filter((ruf) =>
      String(ruf[0]).includes('same key'),
    )

    expect(doppelt).toHaveLength(0)
  })
})
