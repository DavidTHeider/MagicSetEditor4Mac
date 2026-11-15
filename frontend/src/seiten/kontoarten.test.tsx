// **Die vier Kontoarten am Fenster** — Werkstück 6 von Phase 4.12
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

// zumKonto öffnet die Lasche *Konto* — und seit UXR-20 (#1928) die beiden
// Kästen, um die es hier geht: das eigene Konto und das ohne Angaben stehen
// dort eingeklappt hinter Sitzung und zweitem Faktor.
async function zumKonto() {
  await userEvent.click(await screen.findByRole('tab', { name: 'Konto' }))
  await userEvent.click(
    await screen.findByRole('button', { name: 'Ein eigenes Konto' }),
  )
  await userEvent.click(screen.getByRole('button', { name: 'Ohne Angaben' }))
}

// fuelleAus trägt Kennung und Kennwort ein.
async function fuelleAus(kennung: string, kennwort: string) {
  await userEvent.type(
    await screen.findByLabelText('Adresse oder Rufnummer'),
    kennung,
  )
  await userEvent.type(screen.getByLabelText('Kennwort'), kennwort)
}

describe('die vier Kontoarten', () => {
  // **Eine Adresse ist hier eine Kennung** — und die Seite sagt es, bevor
  // jemand tippt (§6). Das Programm verschickt nichts und kann nicht wissen,
  // ob es das Postfach gibt.
  it('sagt, dass die Adresse nicht nachgewiesen wird', async () => {
    belege()
    await zumKonto()

    expect(
      await screen.findByText(/Ob es das Postfach oder die Nummer wirklich/),
    ).toBeInTheDocument()
  })

  // **Die Kennung korrigiert macOS nicht** — eine Adresse ist kein Wort.
  it('schaltet die Korrektur an der Kennung ab', async () => {
    belege()
    await zumKonto()

    expect(
      await screen.findByLabelText('Adresse oder Rufnummer'),
    ).toHaveAttribute('spellcheck', 'false')
  })

  // **Ein Konto entsteht mit Kennung und Kennwort.**
  it('trägt ein Konto ein', async () => {
    const eintragen = vi.fn().mockResolvedValue({ nutzer: einer })

    belege({ KontoEintragen: eintragen })
    await zumKonto()
    await fuelleAus('wer@example.org', 'einlangesgeheimnis')

    await userEvent.click(screen.getByRole('button', { name: 'Konto anlegen' }))

    expect(eintragen).toHaveBeenCalledWith(
      'wer@example.org',
      'einlangesgeheimnis',
      '',
    )

    expect(await screen.findByText('Angelegt: Ich')).toBeInTheDocument()
  })

  // **Und derselbe Kasten meldet an.**
  it('meldet mit derselben Kennung an', async () => {
    const anmelden = vi.fn().mockResolvedValue({ nutzer: einer })

    belege({ KontoAnmelden: anmelden })
    await zumKonto()
    await fuelleAus('+4915112345678', 'einlangesgeheimnis')

    const abschnitt = within(
      screen.getByRole('region', { name: 'Ein eigenes Konto' }),
    )

    await userEvent.click(abschnitt.getByRole('button', { name: 'Anmelden' }))

    expect(anmelden).toHaveBeenCalledWith(
      '+4915112345678',
      'einlangesgeheimnis',
      '',
    )

    expect(await screen.findByText('Angemeldet als Ich')).toBeInTheDocument()
  })

  // **Ein Fehler zur Kennung steht am Feld** (§6) — und nicht als anonymer
  // Hinweis über dem Formular.
  it('setzt einen Befund an sein Feld', async () => {
    belege({
      KontoEintragen: vi.fn().mockResolvedValue({
        nutzer: null,
        fehler: {
          code: 'validierung',
          meldung: 'Unter dieser Kennung gibt es schon ein Konto.',
          feld: 'kennung',
        },
      }),
    })
    await zumKonto()
    await fuelleAus('wer@example.org', 'einlangesgeheimnis')

    await userEvent.click(screen.getByRole('button', { name: 'Konto anlegen' }))

    expect(
      await screen.findByLabelText('Adresse oder Rufnummer'),
    ).toHaveAccessibleDescription(/schon ein Konto/)
  })

  // **Ein Fehler ohne Feldbezug steht über dem Formular.**
  it('meldet einen Fehler ohne Feld über dem Formular', async () => {
    belege({
      KontoEintragen: vi.fn().mockResolvedValue({
        nutzer: null,
        fehler: { code: 'unbekannt', meldung: 'Der Bestand antwortet nicht.' },
      }),
    })
    await zumKonto()
    await fuelleAus('wer@example.org', 'einlangesgeheimnis')

    await userEvent.click(screen.getByRole('button', { name: 'Konto anlegen' }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      /Der Bestand antwortet nicht/,
    )
  })

  // **Ein anonymes Konto braucht kein Feld.**
  it('legt ein anonymes Konto an', async () => {
    const anonym = vi.fn().mockResolvedValue({
      nutzer: { ...einer, name: 'anonym-abcdefghijklmnop', anonym: true },
    })

    belege({ KontoAnonym: anonym })
    await zumKonto()

    await userEvent.click(
      await screen.findByRole('button', { name: 'Anonymes Konto anlegen' }),
    )

    expect(anonym).toHaveBeenCalledWith()

    expect(
      await screen.findByText('Angelegt: anonym-abcdefghijklmnop'),
    ).toBeInTheDocument()
  })

  // **Und ein Fehler dabei steht da, statt still zu verschwinden.**
  it('meldet einen Fehler beim anonymen Konto', async () => {
    belege({
      KontoAnonym: vi.fn().mockResolvedValue({
        nutzer: null,
        fehler: { code: 'unbekannt', meldung: 'Der Bestand antwortet nicht.' },
      }),
    })
    await zumKonto()

    await userEvent.click(
      await screen.findByRole('button', { name: 'Anonymes Konto anlegen' }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      /Der Bestand antwortet nicht/,
    )
  })

  // **Das Kennwort ist verdeckt — und lässt sich zeigen.**
  //
  // **Ohne diesen Weg tippt man blind**: wer sein Kennwort nicht sehen kann,
  // vertippt sich und erfährt es erst bei der Anmeldung.
  it('zeigt das Kennwort auf Wunsch und verdeckt es wieder', async () => {
    belege()
    await zumKonto()

    // **Im eigenen Konto** — seit ID-02 (#1701) steht auf der Lasche ein
    // zweites Kennwortfeld, das der Sitzung.
    const konto = within(
      await screen.findByRole('region', { name: 'Ein eigenes Konto' }),
    )
    const feld = konto.getByLabelText('Kennwort')

    expect(feld).toHaveAttribute('type', 'password')

    await userEvent.click(
      konto.getByRole('button', { name: 'Das Kennwort im Klartext zeigen' }),
    )

    expect(konto.getByLabelText('Kennwort')).toHaveAttribute('type', 'text')

    await userEvent.click(
      konto.getByRole('button', {
        name: 'Das Kennwort wieder verdecken',
      }),
    )

    expect(konto.getByLabelText('Kennwort')).toHaveAttribute('type', 'password')
  })

  // **Eine Antwort ohne Nutzer und ohne Fehler ist ein Fehler** (§8).
  //
  // **Sie kann nicht vorkommen — und steht deshalb hier**: eine halbe
  // Antwort, die wie eine ganze aussieht, ließe die Seite „angelegt" sagen,
  // ohne dass etwas entstanden wäre.
  it('meldet eine leere Antwort, statt Erfolg zu behaupten', async () => {
    belege({
      KontoEintragen: vi.fn().mockResolvedValue({ nutzer: null }),
      KontoAnonym: vi.fn().mockResolvedValue({ nutzer: null }),
    })
    await zumKonto()
    await fuelleAus('wer@example.org', 'einlangesgeheimnis')

    await userEvent.click(screen.getByRole('button', { name: 'Konto anlegen' }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      /ohne Nutzer und ohne Fehler/,
    )

    await userEvent.click(
      screen.getByRole('button', { name: 'Anonymes Konto anlegen' }),
    )

    const meldungen = await screen.findAllByRole('alert')

    expect(meldungen.at(-1)).toHaveTextContent(/ohne Nutzer und ohne Fehler/)
  })

  // **Ohne Kennung und Kennwort ruht der Knopf** — und der Ruf bleibt aus.
  it('ruft nichts, solange die Felder leer sind', async () => {
    const eintragen = vi.fn()

    belege({ KontoEintragen: eintragen })
    await zumKonto()

    expect(
      await screen.findByRole('button', { name: 'Konto anlegen' }),
    ).toBeDisabled()

    expect(eintragen).not.toHaveBeenCalled()
  })
})

// **Wer einen zweiten Faktor hat, wird beim Anmelden danach gefragt.**
//
// **Erst versuchen, dann fragen** (§6) — dieselbe Form wie beim Profilwechsel.
describe('der zweite Faktor beim Anmelden', () => {
  it('fragt erst, wenn die Anwendung ihn verlangt', async () => {
    const anmelden = vi
      .fn()
      .mockResolvedValueOnce({
        nutzer: null,
        fehler: {
          code: 'validierung',
          schluessel: 'zweiterFaktor',
          meldung: 'Dieser Nutzer verlangt einen zweiten Faktor.',
        },
      })
      .mockResolvedValueOnce({ nutzer: einer })

    belege({ KontoAnmelden: anmelden })
    await zumKonto()
    await fuelleAus('wer@example.org', 'einlangesgeheimnis')

    expect(screen.queryByLabelText('Code')).not.toBeInTheDocument()

    const abschnitt = within(
      screen.getByRole('region', { name: 'Ein eigenes Konto' }),
    )

    await userEvent.click(abschnitt.getByRole('button', { name: 'Anmelden' }))

    await userEvent.type(await screen.findByLabelText('Code'), '123456{Enter}')

    expect(anmelden).toHaveBeenNthCalledWith(
      1,
      'wer@example.org',
      'einlangesgeheimnis',
      '',
    )

    expect(anmelden).toHaveBeenNthCalledWith(
      2,
      'wer@example.org',
      'einlangesgeheimnis',
      '123456',
    )
  })
})
