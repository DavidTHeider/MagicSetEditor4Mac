// **Die Kacheln nach einem Profilwechsel** — AP-8, Teil 2b
// (#2365),
// Abnahme „Profilwechsel während Nachladen“. Prüfstände und Portfolio kommen
// **für sich** (PERF02) — also auch für sich zu spät. Hier halten beide ihre
// Antwort aus dem alten Profil zurück, die Anwendung wechselt wirklich, und
// erst danach kommen die alten Antworten an.
//
// **Was schützt**, steht in `alteantwort.test.tsx`: der Abbau des Baums
// unter `key={profilstand}`. Dieser Test hält, dass die zwei neuen Wege
// darunter liegen und keinen Zwischenspeicher haben, der ihn überlebt.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import type { NutzerDTO, Portfolioantwort } from './api'
import type { SetproblemeAntwort } from './apiprobleme'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { belegeAnmeldung } from './test/anmeldebelegung'
import { effekteNachholen } from './test/effekte'
import { probeprofil } from './test/profil'

afterEach(raeumeApi)

const texte = kataloge.de.seiten.start

const ich: NutzerDTO = {
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
    probeprofil({ id: '11', nutzerID: '1', name: 'Turnier', fassung: '1' }),
  ],
}

function portfolio(wert: string): Portfolioantwort {
  return {
    zeilen: [
      {
        waehrung: 'EUR',
        wert,
        exemplare: 1,
        aeltesterPreis: '2026-10-14T08:00:00Z',
        alterStunden: 72,
        quelle: 'Scryfall',
      },
    ],
    ohnePreis: 0,
    fehler: null,
  }
}

function probleme(zahl: number): SetproblemeAntwort {
  return {
    staende: [
      {
        setID: '1',
        name: 'Tischprobe',
        urteil: 'aktuell',
        zahl,
        sicher: true,
        geprueft: 5,
        gesamt: 5,
        beendet: '2026-10-17T03:00:00Z',
        alter: 240,
        karten: [],
      },
    ],
    fehler: null,
  }
}

function belege() {
  const gehalten: (() => void)[] = []
  let gewechselt = false
  let arbeitsprofil = '10'

  // halte gibt im neuen Profil sofort, im alten erst auf Freigabe.
  function halte<T>(alt: T, neu: T): () => Promise<T> {
    return () =>
      gewechselt
        ? Promise.resolve(neu)
        : new Promise<T>((fertig) => {
            gehalten.push(() => {
              fertig(alt)
            })
          })
  }

  belegeApi({
    ...belegeAnmeldung(),
    ...belegeFenster('de'),
    Sets: () =>
      Promise.resolve({
        sets: [
          {
            id: '1',
            name: 'Tischprobe',
            spiel: 'magic',
            stil: 'magic-m15',
            karten: 5,
            herkunft: '',
            herkunftFehlt: false,
          },
        ],
        fehler: null,
      }),
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Sammlungsfenster: () =>
      Promise.resolve({
        exemplare: [],
        gesamt: 0,
        insgesamt: 1,
        stueck: 1,
        fehler: null,
      }),
    Setprobleme: halte(probleme(4), probleme(1)),
    Portfolio: halte(portfolio('999.00'), portfolio('3.00')),
    Nutzer: () => Promise.resolve({ nutzer: [ich] }),
    GezeigtesProfil: () =>
      Promise.resolve({
        profil: ich.profile.find((p) => p.id === arbeitsprofil) ?? null,
      }),
    ProfilWechseln: () => {
      arbeitsprofil = '11'
      gewechselt = true

      return Promise.resolve({ profil: ich.profile[1] ?? null })
    },
    Arbeitskontext: () =>
      Promise.resolve({
        nutzerID: '1',
        profilID: arbeitsprofil,
        eignerID: '1',
        rolle: 'verwalter',
        nurLesend: false,
        modus: 'fenster',
        revision: 1,
        fehler: null,
      }),
  })

  return { gehalten }
}

const ueberblick = () =>
  within(screen.getByRole('list', { name: texte.ueberblick.titel }))

describe('die Kacheln über einen Profilwechsel', () => {
  it('zeigen nur, was das neue Profil sagt', async () => {
    const { gehalten } = belege()
    render(<App />)

    // Die Startseite des alten Profils fragt — und bekommt noch nichts.
    await vi.waitFor(() => {
      expect(gehalten.length).toBe(2)
    })

    await userEvent.click(
      await screen.findByRole('button', { name: /Nutzer und Profile/ }),
    )
    await userEvent.click(await screen.findByRole('tab', { name: 'Profile' }))
    await userEvent.click(
      await screen.findByRole('button', { name: 'Zu Ich · Turnier wechseln' }),
    )

    // **Beleg, dass jemand zuhört**: das neue Profil steht da, die alten
    // Rufe hängen noch.
    // **Die Darstellung des Geldes ist die des Fensters**, nicht die der
    // Sprache — gemessen wird deshalb an den Ziffern.
    const neu = /^1 Exemplar \S*3[.,]00\S* · ältester Preis vor 3 Tagen$/
    expect(
      await ueberblick().findByRole('button', {
        name: neu,
      }),
    ).toBeInTheDocument()
    expect(
      ueberblick().getByRole('button', { name: /1 Set 1 mit Problemen/ }),
    ).toBeInTheDocument()

    for (const freigeben of gehalten) freigeben()
    await effekteNachholen()

    expect(screen.queryByText(/999/)).toBeNull()
    expect(
      ueberblick().getByRole('button', {
        name: neu,
      }),
    ).toBeInTheDocument()
    expect(
      ueberblick().getByRole('button', { name: /5 Karten 1 mit Befund/ }),
    ).toBeInTheDocument()
  })
})
