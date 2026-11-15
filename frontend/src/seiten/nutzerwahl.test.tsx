// **Den Menschen wechseln, der arbeitet** — Werkstück 17 von Phase 4.12
// (#1400).
//
// # Was hier gefragt wird
//
// **Nicht, ob der Wechsel geht** — das steht im Dienst und im Flächenlauf.
// Hier steht, ob die Oberfläche die beiden Gesten auseinanderhält: „ich sehe
// da hinein" und „ich bin jetzt jemand anderes".
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Nutzerseite } from './Nutzerseite'
import { Profilraum } from '../profil'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { belegeAnmeldung } from '../test/anmeldebelegung'
import { probeprofil } from '../test/profil'

import type { NutzerDTO } from '../api'

afterEach(() => {
  raeumeApi()
})

function nutzer(id: string, name: string): NutzerDTO {
  return {
    id,
    name,
    bild: '',
    email: '',
    telefon: '',
    anonym: false,
    anrede: '',
    namenszusatz: '',
    vorname: '',
    weiterenamen: '',
    nachname: '',
    slogan: '',
    kurzebio: '',
    langebio: '',
    fassung: '1',
    profile: [
      probeprofil({
        id: id + '0',
        nutzerID: id,
        name: name + 's Profil',
        fassung: '1',
      }),
    ],
  }
}

const zwei = [nutzer('1', 'Ich'), nutzer('2', 'Andere')]

function zeige(teile: Record<string, unknown> = {}) {
  const gewechselt = vi.fn()

  belegeApi({
    ...belegeAnmeldung(),
    Nutzer: vi.fn(() => Promise.resolve({ nutzer: zwei })),
    GezeigtesProfil: vi.fn(() =>
      Promise.resolve({ profil: zwei[0]?.profile[0] ?? null }),
    ),
    WerArbeitet: vi.fn(() => Promise.resolve({ nutzerID: '1' })),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Profilraum gewechselt={gewechselt}>
        <Nutzerseite />
      </Profilraum>
    </Sprachraum>,
  )

  return gewechselt
}

describe('den arbeitenden Menschen wechseln', () => {
  // **Wer man ist, trägt keinen Knopf** (§6): eine Handlung, die nichts tut,
  // ist keine.
  it('nennt, wer arbeitet, und bietet nur die anderen an', async () => {
    zeige()

    expect(await screen.findByText('Das bist du gerade')).toBeInTheDocument()

    expect(
      await screen.findAllByRole('button', { name: /werden$/ }),
    ).toHaveLength(1)
  })

  it('wechselt und meldet es nach oben', async () => {
    const wechseln = vi
      .fn()
      .mockResolvedValue({ profil: zwei[1]?.profile[0] ?? null })

    const gewechselt = zeige({ NutzerWechseln: wechseln })

    await userEvent.click(
      await screen.findByRole('button', { name: 'Andere werden' }),
    )

    // **Gewartet wird auf die Wirkung und nicht auf den Ruf** (§3).
    await vi.waitFor(() => {
      expect(gewechselt).toHaveBeenCalled()
    })

    expect(wechseln).toHaveBeenCalledWith('2', '')
  })

  // **Ein fehlender zweiter Faktor klappt das Feld auf, statt zu klagen**
  // (§6) — dieselbe Form wie bei der Profilwahl daneben.
  it('fragt nach dem Ausweis, wenn der Mensch einen führt', async () => {
    const wechseln = vi
      .fn()
      .mockResolvedValueOnce({
        profil: null,
        fehler: {
          code: 'validierung',
          schluessel: 'zweiterFaktor',
          meldung: 'Dieser Nutzer verlangt einen zweiten Faktor.',
        },
      })
      .mockResolvedValue({ profil: zwei[1]?.profile[0] ?? null })

    const gewechselt = zeige({ NutzerWechseln: wechseln })

    await userEvent.click(
      await screen.findByRole('button', { name: 'Andere werden' }),
    )

    const feld = await screen.findByRole('textbox', { name: /Code/ })

    await userEvent.type(feld, '123456')
    await userEvent.keyboard('{Enter}')

    await vi.waitFor(() => {
      expect(gewechselt).toHaveBeenCalled()
    })

    expect(wechseln).toHaveBeenLastCalledWith('2', '123456')
  })

  // **Ein gescheiterter Wechsel meldet und baut nichts neu auf.**
  it('meldet einen Fehlschlag und baut nichts neu auf', async () => {
    const gewechselt = zeige({
      NutzerWechseln: vi.fn().mockResolvedValue({
        profil: null,
        fehler: {
          code: 'nichtGefunden',
          schluessel: 'nichtGefunden',
          meldung: 'Den gibt es nicht.',
        },
      }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: 'Andere werden' }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der gesuchte Eintrag wurde nicht gefunden.',
    )

    expect(gewechselt).not.toHaveBeenCalled()
  })
})
