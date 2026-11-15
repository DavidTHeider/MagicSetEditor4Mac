// **Was zu welchem Umfang einer Sicht gehört** — UXR-F18
// (#2033).
//
// # Der Befund
//
// Das Fenster bot „Ausgewählte Drucke" an, ohne Drucke wählen zu lassen — die
// Domäne weist eine Auswahl ohne Drucke ab (`Sammlungsfreigabe.Validate`),
// jede Freigabe auf diesem Umfang scheiterte also. Und bei „Dubletten"
// schickte es die Behaltemenge 0, ohne dass sie sich setzen ließe.
//
// # Was hier gemessen wird
//
// **Dass die Felder nur beim passenden Umfang stehen** — und dass ein Umfang
// nichts schickt, was ihn nicht angeht: die Domäne weist eine Auswahl bei
// einem anderen Umfang ebenso ab wie eine Behaltemenge außerhalb der
// Dubletten.
//
// **Dass ein Fehler am Feld steht** (§6, §8), misst
// `sichtfreigabenbefund.test.tsx`.
import { screen, within } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { raeumeApi } from '../test/api'
import {
  aufstellen,
  eigene,
  eineFreigabe,
  erteilen,
  freigabenliste,
  umfangWaehlen,
  worte,
} from '../test/sichtfreigabenprobe'

afterEach(() => {
  raeumeApi()
})

describe('Die Felder je Umfang', () => {
  it('lässt bei „Ausgewählte Drucke" Drucke wählen und fragt keine Behaltemenge', async () => {
    await aufstellen()

    const gruppe = await screen.findByRole('group', { name: worte.drucke })

    expect(within(gruppe).getAllByRole('checkbox')).toHaveLength(2)
    expect(screen.queryByLabelText(worte.behalten)).toBeNull()
  })

  it('fragt bei „Dubletten" die Behaltemenge — mit der Vorgabe 1 — und lässt keine Drucke wählen', async () => {
    const hand = await aufstellen()

    await umfangWaehlen(hand, 'dubletten')

    // **1 und nicht 0**: wie `mse dubletten` und die Ordnung der Sammlung.
    // Mit 0 wäre jedes Stück eine Dublette, und „Dubletten" gäbe die ganze
    // Sammlung frei.
    expect(screen.getByLabelText(worte.behalten)).toHaveValue(1)
    expect(screen.queryByRole('group', { name: worte.drucke })).toBeNull()
  })

  it.each(['wuensche', 'sammlung', 'digitales'])(
    'zeigt bei %s weder Drucke noch Behaltemenge',
    async (umfang) => {
      const hand = await aufstellen()

      await umfangWaehlen(hand, umfang)

      expect(screen.queryByRole('group', { name: worte.drucke })).toBeNull()
      expect(screen.queryByLabelText(worte.behalten)).toBeNull()
    },
  )
})

describe('Was erteilt wird', () => {
  it('erteilt die gewählten Drucke', async () => {
    const erteilt = vi.fn(() => Promise.resolve(freigabenliste()))
    const hand = await aufstellen({ SammlungFreigeben: erteilt })

    await hand.click(
      await screen.findByRole('checkbox', { name: /Riesenwuchs/ }),
    )
    await hand.click(screen.getByRole('checkbox', { name: /Blitzschlag/ }))
    await hand.click(erteilen())

    expect(erteilt).toHaveBeenCalledWith({
      empfaenger: 'bgegenueber',
      umfang: 'auswahl',
      zweck: 'handel',
      // **In der Reihenfolge der Wahl.**
      auswahl: ['giant', 'bolt'],
      behalten: 0,
      ablauf: '',
      erteilt: '',
    })
  })

  it('erteilt Dubletten mit der Vorgabe, wenn niemand sie ändert', async () => {
    const erteilt = vi.fn(() => Promise.resolve(freigabenliste()))
    const hand = await aufstellen({ SammlungFreigeben: erteilt })

    await umfangWaehlen(hand, 'dubletten')
    await hand.click(erteilen())

    expect(erteilt).toHaveBeenCalledWith(
      expect.objectContaining({
        umfang: 'dubletten',
        behalten: 1,
        auswahl: [],
      }),
    )
  })

  it('erteilt Dubletten mit der gesetzten Behaltemenge', async () => {
    const erteilt = vi.fn(() => Promise.resolve(freigabenliste()))
    const hand = await aufstellen({ SammlungFreigeben: erteilt })

    await umfangWaehlen(hand, 'dubletten')

    const feld = screen.getByLabelText(worte.behalten)
    await hand.clear(feld)
    await hand.type(feld, '4')
    await hand.click(erteilen())

    expect(erteilt).toHaveBeenCalledWith(
      expect.objectContaining({
        umfang: 'dubletten',
        behalten: 4,
        auswahl: [],
      }),
    )
  })

  // **Die Domäne weist beides ab**, wo es nicht zum Umfang gehört — wer eine
  // Wahl trifft und den Umfang dann wechselt, darf sie nicht mitschicken.
  it('schickt bei einem anderen Umfang weder Drucke noch Behaltemenge', async () => {
    const erteilt = vi.fn(() => Promise.resolve(freigabenliste()))
    const hand = await aufstellen({ SammlungFreigeben: erteilt })

    await hand.click(
      await screen.findByRole('checkbox', { name: /Blitzschlag/ }),
    )
    await umfangWaehlen(hand, 'dubletten')

    const feld = screen.getByLabelText(worte.behalten)
    await hand.clear(feld)
    await hand.type(feld, '5')

    await umfangWaehlen(hand, 'sammlung')
    await hand.click(erteilen())

    expect(erteilt).toHaveBeenCalledWith(
      expect.objectContaining({ umfang: 'sammlung', auswahl: [], behalten: 0 }),
    )
  })

  it('behält die Wahl, wenn der Umfang hin- und zurückgewechselt wird', async () => {
    const hand = await aufstellen()

    await hand.click(
      await screen.findByRole('checkbox', { name: /Blitzschlag/ }),
    )
    await umfangWaehlen(hand, 'wuensche')
    await umfangWaehlen(hand, 'auswahl')

    expect(
      await screen.findByRole('checkbox', { name: /Blitzschlag/ }),
    ).toBeChecked()
  })
})

describe('Was die Liste der erteilten Sichten sagt', () => {
  it('nennt die Zahl der Drucke und die Behaltemenge', async () => {
    await aufstellen(
      {},
      freigabenliste([
        eineFreigabe({ umfang: 'auswahl', auswahl: ['bolt', 'giant'] }),
        eineFreigabe({ umfang: 'dubletten', behalten: 2, zweck: 'teilen' }),
        eineFreigabe({ umfang: 'sammlung', zweck: 'partie' }),
      ]),
    )

    const wahl = eigene.gegenueber.umfaenge
    const zeilen = (await screen.findAllByRole('listitem')).map(
      (zeile) => zeile.textContent,
    )

    expect(zeilen).toContainEqual(
      expect.stringContaining(`${wahl.auswahl} · 2 Drucke`),
    )
    expect(zeilen).toContainEqual(
      expect.stringContaining(`${wahl.dubletten} · du behältst 2`),
    )

    // **Die ganze Sammlung hat nichts zu ergänzen** — und eine Lücke an dieser
    // Stelle läse sich wie ein Fehler.
    expect(zeilen).toContainEqual(
      expect.stringMatching(new RegExp(`${wahl.sammlung}\\s*Widerrufen$`)),
    )
  })
})
