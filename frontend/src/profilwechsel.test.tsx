// **Ein Profilwechsel räumt den offenen Kontext** — Werkstück 3 von Phase
// 4.12 (#1351).
//
// # Warum dieser Test über der ganzen Anwendung steht
//
// **Weil die Zusage über der ganzen Anwendung gilt** (§7, Regel 1). Ein Test
// an `Profilwahl` belegt, dass gemeldet wird; dass die **Anwendung** danach
// von vorn beginnt, belegt nur ein Lauf, der sie ganz aufbaut — und eine
// Abnahme ist beobachtbares Verhalten (§10).
import {
  render,
  renderHook,
  screen,
  waitFor,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { belegeEinstellungen } from './test/fensterbelegung'
import { belegeAnmeldung } from './test/anmeldebelegung'

import type { Api, NutzerDTO } from './api'
import { probeprofil } from './test/profil'
import { kataloge } from './sprache/kataloge'
import { leisteneintrag } from './test/leistenweg'
import { merke as merkeBefehl, useBefehlsverlauf } from './befehlsverlauf'
import { merke as merkeSkript, useSkriptverlauf } from './skriptverlauf'

const zwei: NutzerDTO[] = [
  {
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
      probeprofil({
        id: '10',
        nutzerID: '1',
        name: 'Mein Profil',
        fassung: '1',
      }),
      probeprofil({ id: '11', nutzerID: '1', name: 'Turnier', fassung: '1' }),
    ],
  },
]

// arbeitsprofil ist das Profil, das der Bestand als das gezeigte nennt —
// der Wechsel stellt es um, so wie es die Anwendung über den Zeiger täte.
let arbeitsprofil = '10'

// belegeAlles belegt, was die Anwendung beim Wechseln braucht — und was
// ein Lauf dazu ändert. **Ein Ruf und nicht zwei**: `belegeApi` baut die
// Fassade jedes Mal neu.
function belegeAlles(zusatz: Partial<Api> = {}) {
  belegeApi({
    ...belegeAnmeldung(),
    ...belegeFenster('de'),
    ...belegeAnmeldung(),
    Nutzer: vi.fn(() => Promise.resolve({ nutzer: zwei })),
    GezeigtesProfil: vi.fn(() =>
      Promise.resolve({ profil: zwei[0]?.profile[0] ?? null }),
    ),
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
    Formate: () => Promise.resolve({ formate: [], fehler: null }),
    Bauarten: () => Promise.resolve({ bauarten: [], fehler: null }),
    ProfilWechseln: vi.fn(() => {
      arbeitsprofil = '11'

      return Promise.resolve({ profil: zwei[0]?.profile[1] ?? null })
    }),

    // **Die Kontextleiste liest den Arbeitskontext** — Werkstück U4 von
    // Phase 5.1 (#1415); nach dem Wechsel nennt er das neue Profil.
    Arbeitskontext: vi.fn(() =>
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
    ),
    ...zusatz,
  })
}

describe('ein Profilwechsel', () => {
  beforeEach(() => {
    arbeitsprofil = '10'
    belegeAlles()
  })

  afterEach(() => {
    raeumeApi()
  })

  // **Die Anwendung beginnt danach von vorn.**
  //
  // Geprüft an der **Startseite**: wer über ⌘' auf die Nutzerseite geht und
  // wechselt, steht danach wieder dort, wo die Anwendung anfängt. Das ist die
  // sichtbare Seite von „der offene Kontext ist geräumt" — Formular,
  // Rückfrage und geladene Nebenlisten gehen mit, weil das Bauteil, das sie
  // hielt, nicht mehr steht.
  it('baut die Anwendung neu auf', async () => {
    render(<App />)

    // **Vor dem Wechsel nennt die Leiste das alte Profil.**
    const leiste = () => screen.findByRole('region', { name: 'Arbeitskontext' })
    await within(await leiste()).findByText('Ich · Mein Profil')

    await userEvent.click(
      await screen.findByRole('button', { name: /Nutzer und Profile/ }),
    )
    await userEvent.click(await screen.findByRole('tab', { name: 'Profile' }))

    await userEvent.click(
      await screen.findByRole('button', { name: 'Zu Ich · Turnier wechseln' }),
    )

    // **Die Nutzerseite ist weg** — die Anwendung steht wieder am Anfang.
    expect(
      await screen.findByRole('heading', { level: 1, name: /MagicSetEditor/ }),
    ).toBeInTheDocument()

    // **Und die Leiste nennt das neue Profil** — nicht den Menschen von
    // vorhin im Profil von jetzt (#1415).
    await within(await leiste()).findByText('Ich · Turnier')
    expect(await leiste()).not.toHaveTextContent('Mein Profil')
  })

  // **Was eine verlassene Seite aufhob, gehört dem alten Profil** (UX-03.AK-04,
  // #1675): seit die Seiten stehen bleiben, wäre ein halb getippter Deckname
  // sonst im neuen Profil wieder da — fremde Daten.
  it('vergisst, was die verlassenen Seiten aufhoben', async () => {
    render(<App />)
    const decks = kataloge.de.seiten.decks

    await userEvent.click(await leisteneintrag('decks'))
    await userEvent.click(
      await screen.findByRole('button', { name: decks.neu }),
    )
    await userEvent.type(
      await screen.findByRole('textbox', { name: decks.name }),
      'Aus dem alten Profil',
    )

    await userEvent.click(
      await screen.findByRole('button', { name: /Nutzer und Profile/ }),
    )
    await userEvent.click(await screen.findByRole('tab', { name: 'Profile' }))
    await userEvent.click(
      await screen.findByRole('button', { name: 'Zu Ich · Turnier wechseln' }),
    )
    await screen.findByRole('heading', { level: 1, name: /MagicSetEditor/ })

    await userEvent.click(await leisteneintrag('decks'))
    await userEvent.click(
      await screen.findByRole('button', { name: decks.neu }),
    )

    expect(
      await screen.findByRole('textbox', { name: decks.name }),
    ).toHaveValue('')
  })

  // **Die Einstellungen des neuen Profils gelten** — angestrichen vom
  // Zweitblick über Phase 6.3 (S5): gelesen wurde nur beim Start und nach
  // einem ersetzten Bestand, und Profil B trug Farbwelt, eigene Farben und
  // Fensterbild von A. Belegt an der Farbwelt am Wurzelelement.
  it('liest die Einstellungen des neuen Profils', async () => {
    const vorgabe = belegeEinstellungen('de')
    belegeAlles({
      Einstellungen: async () => {
        const antwort = await vorgabe()
        const farbwelt = arbeitsprofil === '11' ? 'schiefer' : 'pergament'

        return {
          einstellungen: [
            ...antwort.einstellungen,
            {
              schluessel: 'farbwelt',
              wert: farbwelt,
              wirksam: farbwelt,
              werte: ['pergament', 'klar', 'schiefer', 'moos'],
              bereich: 'arbeitsumgebung',
            },
          ],
        }
      },
    })
    render(<App />)
    await screen.findByRole('region', { name: 'Arbeitskontext' })
    expect(document.documentElement.getAttribute('data-farbwelt')).toBeNull()

    await userEvent.click(
      await screen.findByRole('button', { name: /Nutzer und Profile/ }),
    )
    await userEvent.click(await screen.findByRole('tab', { name: 'Profile' }))
    await userEvent.click(
      await screen.findByRole('button', { name: 'Zu Ich · Turnier wechseln' }),
    )

    await waitFor(() => {
      expect(document.documentElement.getAttribute('data-farbwelt')).toBe(
        'schiefer',
      )
    })
  })

  // **Was Kommandozentrale und Skriptkonsole gemerkt haben, gehört dem
  // Menschen von vorhin** — samt Argumenten (Nachlauf über Phase 6.3, N6).
  it('vergisst die Verläufe von Befehlen und Skripten', async () => {
    merkeBefehl({
      name: 'set-neu',
      argumente: { name: 'Aus dem alten Profil' },
      ausgang: 'gelungen',
      dauerMs: 1,
      zeitpunkt: '2026-09-26T04:00:00Z',
    })
    merkeSkript({
      ausdruck: '1 + 1',
      ausgang: 'gelungen',
      dauerMs: 1,
      zeitpunkt: '2026-09-26T04:00:00Z',
    })
    render(<App />)

    await userEvent.click(
      await screen.findByRole('button', { name: /Nutzer und Profile/ }),
    )
    await userEvent.click(await screen.findByRole('tab', { name: 'Profile' }))
    await userEvent.click(
      await screen.findByRole('button', { name: 'Zu Ich · Turnier wechseln' }),
    )
    await screen.findByRole('heading', { level: 1, name: /MagicSetEditor/ })

    expect(renderHook(() => useBefehlsverlauf()).result.current).toEqual([])
    expect(renderHook(() => useSkriptverlauf()).result.current).toEqual([])
  })
})
