// **Der zweite Faktor am Fenster** — Werkstück 7 von Phase 4.12
// (#1351).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Nutzerseite } from './Nutzerseite'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { belegeAnmeldung } from '../test/anmeldebelegung'

import type { NutzerDTO } from '../api'
import { probeprofil } from '../test/profil'

afterEach(() => {
  raeumeApi()
})

const einer: NutzerDTO = {
  id: '1',
  name: 'Ich',
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
    probeprofil({ id: '10', nutzerID: '1', name: 'Mein Profil', fassung: '1' }),
  ],
}

const gezeigtesProfil = () =>
  Promise.resolve({ profil: einer.profile[0] ?? null })

function belege(teile: Parameters<typeof belegeApi>[0] = {}) {
  belegeApi({
    ...belegeAnmeldung(),
    GezeigtesProfil: vi.fn(gezeigtesProfil),
    Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
    ...teile,
  })

  render(
    <Sprachraum sprache="de">
      <Nutzerseite />
    </Sprachraum>,
  )
}

// zumKonto öffnet die Lasche *Konto*.
async function zumKonto() {
  await userEvent.click(await screen.findByRole('tab', { name: 'Konto' }))
}

// derAbschnitt gibt den Bereich des zweiten Faktors.
async function derAbschnitt() {
  return within(await screen.findByRole('region', { name: 'Zweiter Faktor' }))
}

describe('der zweite Faktor', () => {
  // **Ohne Sensor ruht der Knopf — und der Grund steht daneben** (§6).
  //
  // **Ein ausgegrauter Knopf ohne Auskunft ist eine Ablehnung, die niemand
  // versteht.**
  it('sagt, warum Biometrie hier nicht geht', async () => {
    belege()
    await zumKonto()

    const abschnitt = await derAbschnitt()

    expect(
      abschnitt.getByRole('button', {
        name: 'Touch ID oder Face ID einrichten',
      }),
    ).toBeDisabled()

    expect(
      abschnitt.getByText('in diesem Lauf gibt es keinen Sensor'),
    ).toBeInTheDocument()
  })

  // **Und wo es geht, geht es** — mit einem Ruf an den Sensor.
  it('richtet Biometrie ein, wo es sie gibt', async () => {
    const einrichten = vi.fn().mockResolvedValue({})

    belege({
      Faktoren: vi.fn(() =>
        Promise.resolve({
          faktoren: [],
          biometrie: true,
          biometrieart: 'Touch ID',
          biometriegrund: '',
        }),
      ),
      BiometrieEinrichten: einrichten,
    })
    await zumKonto()

    const abschnitt = await derAbschnitt()

    // **Der Rechner nennt sein Merkmal beim Namen** — „Touch ID" und nicht
    // „Biometrie".
    expect(abschnitt.getByText('Touch ID')).toBeInTheDocument()

    await userEvent.click(
      abschnitt.getByRole('button', {
        name: 'Touch ID oder Face ID einrichten',
      }),
    )

    expect(einrichten).toHaveBeenCalledWith('1')
  })

  // **Das Geheimnis und die Codes stehen einmal da — und die Seite sagt es.**
  it('zeigt Geheimnis und Codes genau einmal', async () => {
    const bestaetigen = vi.fn().mockResolvedValue({})

    belege({
      FaktorEinrichten: vi.fn(() =>
        Promise.resolve({
          adresse: 'otpauth://totp/MagicSetEditor4Mac:Nutzer%201?secret=ABC',
          geheimnis: 'ABCDEFGHIJKLMNOP',
          codes: ['aaaa-bbbb-cccc', 'dddd-eeee-ffff'],
        }),
      ),
      FaktorBestaetigen: bestaetigen,
    })
    await zumKonto()

    const abschnitt = await derAbschnitt()

    await userEvent.click(
      abschnitt.getByRole('button', {
        name: 'Einen Code aus einer App einrichten',
      }),
    )

    expect(await screen.findByText('ABCDEFGHIJKLMNOP')).toBeInTheDocument()
    expect(screen.getByText('aaaa-bbbb-cccc')).toBeInTheDocument()
    expect(
      screen.getByText(/2 Wiederherstellungscodes, jeder gilt einmal/),
    ).toBeInTheDocument()

    // **Und erst ein Code aus der App trägt ihn ein.**
    await userEvent.type(screen.getByLabelText('Code'), '123456')
    await userEvent.click(screen.getByRole('button', { name: 'Bestätigen' }))

    expect(bestaetigen).toHaveBeenCalledWith('1', '123456')
  })

  // **Ein eingerichteter Faktor lässt sich entfernen.**
  it('entfernt einen eingerichteten Faktor', async () => {
    const entfernen = vi.fn().mockResolvedValue({})

    belege({
      Faktoren: vi.fn(() =>
        Promise.resolve({
          faktoren: [{ art: 'totp', bezeichnung: '' }],
          biometrie: false,
          biometrieart: '',
          biometriegrund: 'kein Sensor',
        }),
      ),
      FaktorEntfernen: entfernen,
    })
    await zumKonto()

    const abschnitt = await derAbschnitt()

    await userEvent.click(
      abschnitt.getByRole('button', {
        name: 'Den Code aus einer App entfernen',
      }),
    )

    // **Erst nach der Frage** (UX-10.AK-04, #1682) — und Abbrechen entfernt
    // nichts.
    await userEvent.click(abschnitt.getByRole('button', { name: 'Abbrechen' }))
    expect(entfernen).not.toHaveBeenCalled()

    await userEvent.click(
      abschnitt.getByRole('button', {
        name: 'Den Code aus einer App entfernen',
      }),
    )
    await userEvent.click(
      within(
        abschnitt.getByRole('group', {
          name: 'Den Code aus einer App entfernen?',
        }),
      ).getByRole('button', { name: 'Entfernen' }),
    )

    expect(entfernen).toHaveBeenCalledWith('1', 'totp')
  })

  // **Und Biometrie ebenso.**
  it('entfernt einen biometrischen Faktor', async () => {
    const entfernen = vi.fn().mockResolvedValue({})

    belege({
      Faktoren: vi.fn(() =>
        Promise.resolve({
          faktoren: [{ art: 'biometrie', bezeichnung: '' }],
          biometrie: true,
          biometrieart: 'Touch ID',
          biometriegrund: '',
        }),
      ),
      FaktorEntfernen: entfernen,
    })
    await zumKonto()

    const abschnitt = await derAbschnitt()

    await userEvent.click(
      abschnitt.getByRole('button', {
        name: 'Touch ID oder Face ID entfernen',
      }),
    )
    await userEvent.click(
      within(
        abschnitt.getByRole('group', {
          name: 'Touch ID oder Face ID entfernen?',
        }),
      ).getByRole('button', { name: 'Entfernen' }),
    )

    expect(entfernen).toHaveBeenCalledWith('1', 'biometrie')
  })

  // **Ein Fehler beim Einrichten steht da, statt still zu verschwinden** (§8).
  it('meldet einen Fehler beim Einrichten', async () => {
    belege({
      FaktorEinrichten: vi.fn(() =>
        Promise.resolve({
          adresse: '',
          geheimnis: '',
          codes: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'unbekannt',
            meldung: 'Der Bund antwortet nicht.',
          },
        }),
      ),
    })
    await zumKonto()

    const abschnitt = await derAbschnitt()

    await userEvent.click(
      abschnitt.getByRole('button', {
        name: 'Einen Code aus einer App einrichten',
      }),
    )

    // **Der Satz kommt aus dem Katalog und nicht aus Go** (§9): ein
    // Schlüssel, den das Fenster nicht kennt, gibt den Rückfallsatz. Geprüft
    // wird, dass der Fehler die Oberfläche **erreicht** — nicht, wie er dort
    // heißt.
    expect(await screen.findByRole('alert')).toHaveTextContent(
      /unerwarteter Fehler/,
    )
  })

  // **Und ein Fehler beim Laden ebenso.**
  it('meldet einen Fehler beim Laden', async () => {
    belege({
      Faktoren: vi.fn(() =>
        Promise.resolve({
          faktoren: [],
          biometrie: false,
          biometrieart: '',
          biometriegrund: '',
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'unbekannt',
            meldung: 'Der Bestand antwortet nicht.',
          },
        }),
      ),
    })
    await zumKonto()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      /unerwarteter Fehler/,
    )
  })
})

// **Wer einen zweiten Faktor hat, wird beim Wechsel danach gefragt.**
//
// # Erst versuchen, dann fragen
//
// **Weil die meisten keinen haben** (§6). Die Zeile klappt ihr Feld erst auf,
// wenn die Anwendung einen verlangt — und nicht vorsorglich neben jeder.
describe('der zweite Faktor beim Wechseln', () => {
  it('fragt erst, wenn die Anwendung ihn verlangt', async () => {
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
      .mockResolvedValueOnce({ profil: einer.profile[0] })

    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(() => Promise.resolve({ profil: null })),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
      ProfilWechseln: wechseln,
    })

    render(
      <Sprachraum sprache="de">
        <Nutzerseite />
      </Sprachraum>,
    )

    await userEvent.click(await screen.findByRole('tab', { name: 'Profile' }))

    // **Kein Feld, bevor gefragt wurde.**
    expect(screen.queryByLabelText('Code')).not.toBeInTheDocument()

    await userEvent.click(
      await screen.findByRole('button', {
        name: 'Zu Ich · Mein Profil wechseln',
      }),
    )

    await userEvent.type(await screen.findByLabelText('Code'), '123456{Enter}')

    expect(wechseln).toHaveBeenNthCalledWith(1, '10', '')
    expect(wechseln).toHaveBeenNthCalledWith(2, '10', '123456')
  })

  // **Und ein anderer Fehler bleibt eine Meldung.**
  //
  // **Die Gegenprobe zur Zeile darüber**: ohne sie klappte das Feld auch bei
  // einem Fehler auf, mit dem es nichts zu tun hat.
  it('klappt bei einem anderen Fehler kein Feld auf', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(() => Promise.resolve({ profil: null })),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
      ProfilWechseln: vi.fn(() =>
        Promise.resolve({
          profil: null,
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'unbekannt',
            meldung: 'Der Bestand antwortet nicht.',
          },
        }),
      ),
    })

    render(
      <Sprachraum sprache="de">
        <Nutzerseite />
      </Sprachraum>,
    )

    await userEvent.click(await screen.findByRole('tab', { name: 'Profile' }))
    await userEvent.click(
      await screen.findByRole('button', {
        name: 'Zu Ich · Mein Profil wechseln',
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      /unerwarteter Fehler/,
    )

    expect(screen.queryByLabelText('Code')).not.toBeInTheDocument()
  })
})
