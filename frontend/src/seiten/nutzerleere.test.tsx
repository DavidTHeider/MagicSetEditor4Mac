// **Die leeren Zustände der Nutzerseite** — Phase 4.12
// (#1351).
//
// **Eine eigene Datei**, seit `nutzerseite.test.tsx` mit Werkstück 8 über die
// Zeilengrenze aus §2 lief. Der Schnitt folgt der Frage: dort steht, was die
// Seite **tut**, hier, was sie sagt, wenn es nichts zu zeigen gibt.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Nutzerseite } from './Nutzerseite'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { belegeAnmeldung } from '../test/anmeldebelegung'
import { effekteNachholen } from '../test/effekte'

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

// **Das gezeigte Profil gehört zu jeder Belegung** — die Seite holt beides in
// einem Zug (§7, Regel 2), und eine Vorrichtung, die nur die Liste belegt,
// scheitert mit dem Namen der fehlenden Methode statt mit `undefined`.
const gezeigtesProfil = () =>
  Promise.resolve({ profil: einer.profile[0] ?? null })

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Nutzerseite />
    </Sprachraum>,
  )
}

describe('die leeren Zustände der Nutzerseite', () => {
  // nächsten Schritt).
  it('sagt, wenn ein Nutzer noch kein Profil hat', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() =>
        Promise.resolve({ nutzer: [{ ...einer, profile: [] }] }),
      ),
    })

    zeige()

    expect(await screen.findByText(/noch kein Profil/)).toBeInTheDocument()
  })

  // **Und eine leere Liste erklärt, warum sie leer ist.**
  it('erklärt die leere Liste', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [] })),
    })
    zeige()

    expect(
      await screen.findByText(/Noch kein weiterer Nutzer/),
    ).toBeInTheDocument()
  })

  describe('Profile', () => {
    // **Die Schublade wird geöffnet** — sonst prüft dieser Lauf ein Bauteil,
    // das niemand sieht. **Seit UXR-20** (#1928) steht das Anlegen hinter
    // *+ Profil anlegen …* im Kopf und nicht mehr am Fuß der Lasche
    // *Profile*; der Name `zumProfil` blieb, der Weg ist der neue.
    async function zumProfil() {
      await userEvent.click(
        await screen.findByRole('button', { name: 'Profil anlegen …' }),
      )
      await screen.findByRole('dialog', { name: 'Weiteres Profil anlegen' })
    }

    it('legt ein Profil für den gewählten Nutzer an', async () => {
      const anlegen = vi.fn().mockResolvedValue({
        profil: probeprofil({
          id: '20',
          nutzerID: '1',
          name: 'Turnier',
          fassung: '1',
        }),
      })

      belegeApi({
        ...belegeAnmeldung(),
        GezeigtesProfil: vi.fn(gezeigtesProfil),
        Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
        ProfilAnlegen: anlegen,
      })

      zeige()
      await effekteNachholen()
      await zumProfil()

      await userEvent.type(screen.getByLabelText(/Name des Profils/), 'Turnier')
      await userEvent.click(
        screen.getByRole('button', { name: 'Profil anlegen' }),
      )

      await screen.findByText(/Angelegt: Turnier/)

      // **Der erste Nutzer ist vorgewählt** — es gibt immer einen.
      expect(anlegen).toHaveBeenCalledWith('1', 'Turnier')
    })

    it('nimmt den gewählten Nutzer, nicht den ersten', async () => {
      const zweiter = { ...einer, id: '2', name: 'Zweite', profile: [] }

      const anlegen = vi.fn().mockResolvedValue({
        profil: probeprofil({
          id: '21',
          nutzerID: '2',
          name: 'Arbeit',
          fassung: '1',
        }),
      })

      belegeApi({
        ...belegeAnmeldung(),
        GezeigtesProfil: vi.fn(gezeigtesProfil),
        Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer, zweiter] })),
        ProfilAnlegen: anlegen,
      })

      zeige()
      await effekteNachholen()
      await zumProfil()

      await userEvent.selectOptions(
        screen.getByLabelText(/Für welchen Nutzer/),
        '2',
      )
      await userEvent.type(screen.getByLabelText(/Name des Profils/), 'Arbeit')
      await userEvent.click(
        screen.getByRole('button', { name: 'Profil anlegen' }),
      )

      await screen.findByText(/Angelegt: Arbeit/)

      expect(anlegen).toHaveBeenCalledWith('2', 'Arbeit')
    })

    // **Ohne Nutzer ruht der Knopf.** Es gibt zwar immer einen (Wanderung
    // 0034), aber ein Formular, das ins Leere schriebe, wäre ein Knopf, der
    // nichts tut — und die leere Wahl ist der Zustand, in dem er es täte.
    it('ruht, solange es keinen Nutzer gibt', async () => {
      belegeApi({
        ...belegeAnmeldung(),
        GezeigtesProfil: vi.fn(gezeigtesProfil),
        Nutzer: vi.fn(() => Promise.resolve({ nutzer: [] })),
      })

      zeige()
      await effekteNachholen()
      await zumProfil()

      await userEvent.type(screen.getByLabelText(/Name des Profils/), 'Waise')

      expect(
        screen.getByRole('button', { name: 'Profil anlegen' }),
      ).toBeDisabled()
    })

    it('zeigt einen Feldfehler am Namensfeld', async () => {
      belegeApi({
        ...belegeAnmeldung(),
        GezeigtesProfil: vi.fn(gezeigtesProfil),
        Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
        ProfilAnlegen: () =>
          Promise.resolve({
            profil: null,
            fehler: {
              code: 'validierung' as const,
              schluessel: 'validierung',
              meldung: 'Ein Name trägt keine Leerzeichen am Rand.',
              feld: 'name',
            },
          }),
      })

      zeige()
      await effekteNachholen()
      await zumProfil()

      const feld = screen.getByLabelText(/Name des Profils/)

      await userEvent.type(feld, 'Turnier')
      await userEvent.click(
        screen.getByRole('button', { name: 'Profil anlegen' }),
      )

      const meldung = await screen.findByRole('alert')

      // **Enthält, nicht gleich** — seit `Feld` (#1413) nennt die Beschreibung
      // Befund und Hinweis zusammen.
      expect(feld.getAttribute('aria-describedby')).toContain(meldung.id)
    })

    // **Ein Fehler ohne Feldbezug steht über dem Formular** — die Gegenprobe.
    it('zeigt einen Fehler ohne Feldbezug über dem Formular', async () => {
      belegeApi({
        ...belegeAnmeldung(),
        GezeigtesProfil: vi.fn(gezeigtesProfil),
        Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
        ProfilAnlegen: () =>
          Promise.resolve({
            profil: null,
            fehler: {
              code: 'unbekannt' as const,
              schluessel: 'unbekannt',
              meldung: 'Etwas ging schief.',
            },
          }),
      })

      zeige()
      await effekteNachholen()
      await zumProfil()

      const feld = screen.getByLabelText(/Name des Profils/)

      await userEvent.type(feld, 'Turnier')
      await userEvent.click(
        screen.getByRole('button', { name: 'Profil anlegen' }),
      )

      await screen.findByRole('alert')

      expect(feld).not.toHaveAttribute('aria-invalid', 'true')
    })
  })
})
