// **Nutzer und Profile im Fenster** — Phase 4.12
// (#1351).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Nutzerseite } from './Nutzerseite'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'
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

// zurAnlage klappt *Neuen Nutzer anlegen* auf — **seit UXR-20 (#1928)**
// steht das Anlegen eines Menschen eingeklappt unter den eigenen Angaben.
async function zurAnlage() {
  await userEvent.click(
    await screen.findByRole('button', {
      name: kataloge.de.seiten.nutzer.neuerNutzer,
    }),
  )
}

describe('Nutzer und Profile', () => {
  // **Vom Vorbehalt über der Seite steht ein Satz da, der Rest hinter dem
  // Hilfesymbol** — UX-15 (#1771). In der Sichtprüfung vom 2026-09-24
  // standen drei Zeilen über den Laschen, an keinem Bauteil vorbei, das die
  // Grenze von 140 Zeichen kennt.
  it('zeigt vom Vorbehalt über der Seite nur den ersten Satz', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
    })
    zeige()

    const vorbehalt = kataloge.de.seiten.nutzer.nochNicht

    expect(await screen.findByText(kurzform(vorbehalt))).toBeInTheDocument()
    expect(screen.queryByText(vorbehalt)).toBeNull()
  })

  it('zeigt jeden Nutzer mit seinen Profilen', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
    })
    zeige()

    // **Zweimal je Name**: die Liste zeigt sie, die Pflege bietet sie an.
    // Ein `findByText` fände zwei und klagte — `findAllByText` ist die Frage,
    // die zum Bild passt.
    expect(await screen.findAllByText('Ich')).not.toHaveLength(0)
    expect(screen.getAllByText('Mein Profil')).not.toHaveLength(0)
  })

  // **Die drei Laschen sind die Gliederung, die der Auftraggeber verlangt
  // hat** — Allgemeines, Konto, Profile.
  it('gliedert in drei Laschen', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
    })
    zeige()

    for (const name of ['Allgemeines', 'Konto', 'Profile']) {
      expect(await screen.findByRole('tab', { name })).toBeInTheDocument()
    }
  })

  it('legt einen Nutzer an und lädt die Liste neu', async () => {
    const lesen = vi
      .fn()
      .mockResolvedValueOnce({ nutzer: [einer] })
      .mockResolvedValue({
        nutzer: [einer, { ...einer, id: '2', name: 'Zweite', profile: [] }],
      })

    const anlegen = vi.fn().mockResolvedValue({
      nutzer: { ...einer, id: '2', name: 'Zweite', profile: [] },
    })

    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: lesen,
      NutzerAnlegen: anlegen,
    })
    zeige()
    await effekteNachholen()
    await zurAnlage()

    await userEvent.type(
      await screen.findByLabelText(kataloge.de.seiten.nutzer.neuerNutzerName),
      'Zweite',
    )
    await userEvent.click(
      screen.getByRole('button', { name: 'Nutzer anlegen' }),
    )

    // **Die Liste wird neu gelesen** — geschrieben ist geschrieben (§7,
    // Regel 3), und das Nachladen ist die Bestenfalls-Arbeit danach.
    expect(await screen.findAllByText('Zweite')).not.toHaveLength(0)
    expect(anlegen).toHaveBeenCalledWith('Zweite', '')
  })

  // **Ein Feldfehler steht am Feld** (§6) — die Lehre aus dem Nachlauf über
  // Phase 4.10.
  it('zeigt einen Feldfehler am Namensfeld', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
      NutzerAnlegen: () =>
        Promise.resolve({
          nutzer: null,
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
    await zurAnlage()

    const feld = await screen.findByLabelText(
      kataloge.de.seiten.nutzer.neuerNutzerName,
    )

    await userEvent.type(feld, 'David')
    await userEvent.click(
      screen.getByRole('button', { name: 'Nutzer anlegen' }),
    )

    const meldung = await screen.findByRole('alert')

    expect(feld).toHaveAttribute('aria-invalid', 'true')
    // **Enthält, nicht gleich** — seit `Feld` (#1413) nennt die Beschreibung
    // Befund und Hinweis zusammen.
    expect(feld.getAttribute('aria-describedby')).toContain(meldung.id)
  })

  it('meldet einen Fehler beim Laden', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: () =>
        Promise.resolve({
          nutzer: [],
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'unbekannt',
            meldung: 'Die Platte schweigt.',
          },
        }),
    })

    zeige()

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })

  // **Und die Seite sagt, was sie noch nicht kann** (§6).
  it('nennt, was in dieser Phase noch fehlt', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
    })
    zeige()

    expect(
      await screen.findByText(/Die Daten sind zwischen Profilen getrennt/),
    ).toBeInTheDocument()
  })

  // **Ein Fehler ohne Feldbezug steht über dem Formular** — die Gegenprobe
  // zum Lauf darüber: ohne sie wäre eine Fassung, die **jeden** Fehler ans
  // Namensfeld hängt, ebenso grün.
  it('zeigt einen Nutzerfehler ohne Feldbezug über dem Formular', async () => {
    belegeApi({
      ...belegeAnmeldung(),
      GezeigtesProfil: vi.fn(gezeigtesProfil),
      Nutzer: vi.fn(() => Promise.resolve({ nutzer: [einer] })),
      NutzerAnlegen: () =>
        Promise.resolve({
          nutzer: null,
          fehler: {
            code: 'unbekannt' as const,
            schluessel: 'unbekannt',
            meldung: 'Etwas ging schief.',
          },
        }),
    })

    zeige()
    await effekteNachholen()
    await zurAnlage()

    const feld = await screen.findByLabelText(
      kataloge.de.seiten.nutzer.neuerNutzerName,
    )

    await userEvent.type(feld, 'Zweite')
    await userEvent.click(
      screen.getByRole('button', { name: 'Nutzer anlegen' }),
    )

    await screen.findByRole('alert')

    expect(feld).not.toHaveAttribute('aria-invalid', 'true')
  })
})
