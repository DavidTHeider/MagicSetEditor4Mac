// **Für andere erreichbar sein** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Was hier gemessen wird
//
// **Dass nichts von selbst horcht** — der Kasten fragt beim Aufbau nach dem
// Stand und tut sonst nichts. Und **dass die wirkliche Adresse dasteht**: wer
// `:0` eingibt, bekommt einen Port vom Betriebssystem, und ohne sie in der
// Antwort wüsste niemand, welchen er dem Gegenüber nennen soll.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Horchposten } from './Horchposten'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { Horchantwort } from '../apipeerkontakt'

const texte = kataloge.de.seiten.technik

// einStand baut eine Antwort.
function einStand(
  laeuft: boolean,
  adresse: string,
  fehler: Horchantwort['fehler'] = null,
  sichtbar = false,
  erreichbarAdressen: string[] = [],
  erreichbarNamen: string[] = [],
): Horchantwort {
  return {
    stand: {
      laeuft,
      adresse,
      sichtbar,
      erreichbarNamen,
      erreichbarAdressen,
    },
    fehler,
  }
}

// zeige stellt den Kasten auf.
function zeige() {
  render(
    <Sprachraum sprache="de">
      <Horchposten />
    </Sprachraum>,
  )

  return userEvent.setup()
}

describe('Für andere erreichbar sein', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('sagt, dass gerade nicht gehorcht wird', async () => {
    belegeApi({ HorchStand: () => Promise.resolve(einStand(false, '')) })

    zeige()

    expect(await screen.findByText(texte.horchenStill)).toBeInTheDocument()
  })

  // **Der Lauf, für den es diesen Kasten gibt.**
  it('nennt die wirkliche Adresse, nicht die eingetippte', async () => {
    // **Die Adresse wird mitgeschrieben** — ein Spion ohne Argument
    // belegte nur, dass gerufen wurde, und nicht, womit.
    const gesehen: string[] = []

    const gerufen = vi.fn((adresse: string) => {
      gesehen.push(adresse)

      return Promise.resolve(einStand(true, '127.0.0.1:53123'))
    })

    let laeuft = false

    belegeApi({
      HorchStand: () =>
        Promise.resolve(
          laeuft ? einStand(true, '127.0.0.1:53123') : einStand(false, ''),
        ),
      HorchStarten: (adresse: string) => {
        laeuft = true

        return gerufen(adresse)
      },
    })

    const hand = zeige()

    const feld = await screen.findByLabelText(texte.horchenAdresse)
    await hand.clear(feld)
    await hand.type(feld, '127.0.0.1:0')
    await hand.click(screen.getByRole('button', { name: texte.horchenStarten }))

    expect(gesehen).toEqual(['127.0.0.1:0'])

    expect(await screen.findByText(/127\.0\.0\.1:53123/)).toBeInTheDocument()
  })

  it('hört wieder auf', async () => {
    let laeuft = true

    const beendet = vi.fn(() => {
      laeuft = false

      return Promise.resolve(einStand(false, ''))
    })

    belegeApi({
      HorchStand: () =>
        Promise.resolve(
          laeuft ? einStand(true, ':41234') : einStand(false, ''),
        ),
      HorchBeenden: beendet,
    })

    const hand = zeige()

    await hand.click(
      await screen.findByRole('button', { name: texte.horchenBeenden }),
    )

    expect(beendet).toHaveBeenCalled()
    expect(await screen.findByText(texte.horchenStill)).toBeInTheDocument()
  })

  // **Ohne Adresse geht nichts hinaus** — der Knopf bleibt zu.
  it('hält den Knopf an, solange keine Adresse dasteht', async () => {
    belegeApi({ HorchStand: () => Promise.resolve(einStand(false, '')) })

    const hand = zeige()

    await hand.clear(await screen.findByLabelText(texte.horchenAdresse))

    expect(
      screen.getByRole('button', { name: texte.horchenStarten }),
    ).toBeDisabled()
  })

  it('zeigt einen Fehler des Dienstes', async () => {
    belegeApi({
      HorchStand: () => Promise.resolve(einStand(false, '')),
      HorchStarten: () =>
        Promise.resolve(
          einStand(false, '', {
            code: 'validierung',
            schluessel: 'fehler.validierung',
            meldung: 'Diese Adresse ist keine.',
          }),
        ),
    })

    const hand = zeige()

    const feld = await screen.findByLabelText(texte.horchenAdresse)
    await hand.clear(feld)
    await hand.type(feld, 'krumm')
    await hand.click(screen.getByRole('button', { name: texte.horchenStarten }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Diese Adresse ist keine.',
    )
  })

  it('zeigt einen Sturz der Naht', async () => {
    belegeApi({
      HorchStand: () => Promise.resolve(einStand(false, '')),
      HorchStarten: () => Promise.reject(new Error('kaputt')),
    })

    const hand = zeige()

    await hand.type(await screen.findByLabelText(texte.horchenAdresse), ':1')
    await hand.click(screen.getByRole('button', { name: texte.horchenStarten }))

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })

  it('zeigt einen Fehler beim Lesen des Standes', async () => {
    belegeApi({
      HorchStand: () =>
        Promise.resolve(
          einStand(false, '', {
            code: 'validierung',
            schluessel: 'fehler.validierung',
            meldung: 'Kein Profil.',
          }),
        ),
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent('Kein Profil.')
  })
})

describe('In der Nähe gefunden werden', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  // **Ohne Port keine Ankündigung** — und das steht da, statt dass ein Knopf
  // ins Leere führte.
  it('verlangt erst das Erreichbarsein', async () => {
    belegeApi({ HorchStand: () => Promise.resolve(einStand(false, '')) })

    zeige()

    expect(
      await screen.findByText(texte.sichtbarBrauchtHorchen),
    ).toBeInTheDocument()

    expect(
      screen.queryByRole('button', { name: texte.sichtbarWerden }),
    ).not.toBeInTheDocument()
  })

  it('kündigt unter dem getippten Namen an', async () => {
    const gesehen: string[] = []

    let sichtbar = false

    belegeApi({
      HorchStand: () =>
        Promise.resolve(einStand(true, ':41234', null, sichtbar)),
      SichtbarWerden: (name: string) => {
        gesehen.push(name)
        sichtbar = true

        return Promise.resolve(einStand(true, ':41234', null, true))
      },
    })

    const hand = zeige()

    await hand.type(
      await screen.findByLabelText(texte.sichtbarName),
      'Davids Rechner',
    )
    await hand.click(screen.getByRole('button', { name: texte.sichtbarWerden }))

    expect(await screen.findByText(texte.sichtbarLaeuft)).toBeInTheDocument()
    expect(gesehen).toEqual(['Davids Rechner'])
  })

  it('hört wieder auf, sich anzukündigen', async () => {
    let sichtbar = true

    const beendet = vi.fn(() => {
      sichtbar = false

      return Promise.resolve(einStand(true, ':41234', null, false))
    })

    belegeApi({
      HorchStand: () =>
        Promise.resolve(einStand(true, ':41234', null, sichtbar)),
      SichtbarkeitBeenden: beendet,
    })

    const hand = zeige()

    await hand.click(
      await screen.findByRole('button', { name: texte.sichtbarBeenden }),
    )

    expect(beendet).toHaveBeenCalled()
    expect(await screen.findByText(texte.sichtbarStill)).toBeInTheDocument()
  })

  it('zeigt einen Fehler der Ankündigung', async () => {
    belegeApi({
      HorchStand: () => Promise.resolve(einStand(true, ':41234')),
      SichtbarWerden: () =>
        Promise.resolve(
          einStand(true, ':41234', {
            code: 'konflikt',
            schluessel: 'fehler.konflikt',
            meldung: 'Dieses Profil kündigt sich schon an.',
          }),
        ),
    })

    const hand = zeige()

    await hand.click(
      await screen.findByRole('button', { name: texte.sichtbarWerden }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Dieses Profil kündigt sich schon an.',
    )
  })
})

describe('Wohin ein Gegenüber wählt', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  // **Der Befund aus der Sichtprüfung vom 2026-09-14.**
  //
  // Der Kasten sagte „Erreichbar unter `[::]:41234`" — wahr über die Bindung
  // und unbrauchbar für den, der wählen soll. Gemessen am selben Tag war der
  // Rechner unter `192.168.178.51:41234` erreichbar, und diese Zeichenkette
  // stand an keiner Fläche.
  it('nennt die Adresse, die ein Gegenüber eintippen kann', async () => {
    belegeApi({
      HorchStand: () =>
        Promise.resolve(
          einStand(
            true,
            '[::]:41234',
            null,
            false,
            [
              '192.168.178.51:41234',
              '[2001:db8:470f:cd00:1ce7:a492:8132:2b75]:41234',
            ],
            ['MacBook-Air-von-Ada.fritz.box:41234'],
          ),
        ),
    })

    zeige()

    expect(await screen.findByText('192.168.178.51:41234')).toBeInTheDocument()
    expect(
      screen.getByText('[2001:db8:470f:cd00:1ce7:a492:8132:2b75]:41234'),
    ).toBeInTheDocument()

    // **Und der Name steht dabei** — er überlebt einen Leasewechsel, die
    // Adresse nicht.
    expect(
      screen.getByText('MacBook-Air-von-Ada.fritz.box:41234'),
    ).toBeInTheDocument()
    expect(screen.getByText(texte.horchenNamen)).toBeInTheDocument()
  })

  // **Ein Rechner ohne auflösbaren Namen zeigt keine leere Überschrift.**
  it('lässt die Namenszeile weg, wenn es keinen gibt', async () => {
    belegeApi({
      HorchStand: () =>
        Promise.resolve(
          einStand(true, '[::]:41234', null, false, ['192.168.178.51:41234']),
        ),
    })

    zeige()

    expect(await screen.findByText('192.168.178.51:41234')).toBeInTheDocument()
    expect(screen.queryByText(texte.horchenNamen)).not.toBeInTheDocument()
  })

  // **Und wer nur auf der Rückschleife horcht, erfährt es.**
  //
  // **Eine leere Liste ist eine Auskunft und kein fehlender Wert**: es ist
  // der Unterschied zwischen „niemand kommt hier an" und „wir wissen es
  // nicht". Ohne diesen Satz stünde da eine Überschrift ohne Inhalt.
  it('sagt, wenn von außen niemand ankommt', async () => {
    belegeApi({
      HorchStand: () => Promise.resolve(einStand(true, '127.0.0.1:41234')),
    })

    zeige()

    expect(await screen.findByText(texte.horchenNurHier)).toBeInTheDocument()
  })
})
