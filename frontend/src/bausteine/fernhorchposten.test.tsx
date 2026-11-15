// **Über fremde Netze erreichbar sein** — Werkstück N3 von Phase 5.4
// (#1456).
//
// # Was hier gemessen wird
//
// **Dass nichts von selbst eine Einladung ausgibt** — der Kasten fragt beim
// Aufbau nach dem Stand und tut sonst nichts. Und **dass die Einladung
// dasteht**, sobald es eine gibt: sie ist der ganze Zweck dieses Kastens.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Fernhorchposten } from './Fernhorchposten'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { Fernantwort } from '../apipeerkontakt'

const texte = kataloge.de.seiten.technik

function einStand(laeuft: boolean, einladung: string): Fernantwort {
  return { stand: { laeuft, einladung, wege: [], angebunden: 0 }, fehler: null }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Fernhorchposten />
    </Sprachraum>,
  )

  return userEvent.setup()
}

describe('Über fremde Netze erreichbar sein', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('sagt, dass es keine Einladung gibt', async () => {
    belegeApi({ FernHorchStand: () => Promise.resolve(einStand(false, '')) })

    zeige()

    expect(await screen.findByText(texte.fernStill)).toBeInTheDocument()
  })

  // **Der Lauf, für den es diesen Kasten gibt.**
  it('zeigt die Einladung, sobald es eine gibt', async () => {
    belegeApi({
      FernHorchStand: () =>
        Promise.resolve(einStand(true, 'mse4mac:1:tc1abcdef')),
    })

    zeige()

    expect(await screen.findByText('mse4mac:1:tc1abcdef')).toBeInTheDocument()

    // **Und der Satz, der sagt, wie lange sie gilt** — eine Einladung ohne
    // diese Auskunft sähe dauerhaft aus.
    expect(screen.getByText(texte.fernStirbt)).toBeInTheDocument()
  })

  it('erzeugt eine Einladung auf Knopfdruck', async () => {
    const gerufen = vi.fn(() =>
      Promise.resolve(einStand(true, 'mse4mac:1:tc1neu')),
    )

    let laeuft = false

    belegeApi({
      FernHorchStand: () =>
        Promise.resolve(einStand(laeuft, laeuft ? 'mse4mac:1:tc1neu' : '')),
      FernHorchStarten: () => {
        laeuft = true

        return gerufen()
      },
    })

    const nutzer = zeige()

    await nutzer.click(
      await screen.findByRole('button', { name: texte.fernStarten }),
    )

    expect(await screen.findByText('mse4mac:1:tc1neu')).toBeInTheDocument()
    expect(gerufen).toHaveBeenCalledTimes(1)
  })

  it('beendet die Erreichbarkeit auf Knopfdruck', async () => {
    const gerufen = vi.fn(() => Promise.resolve(einStand(false, '')))

    let laeuft = true

    belegeApi({
      FernHorchStand: () =>
        Promise.resolve(einStand(laeuft, laeuft ? 'mse4mac:1:tc1alt' : '')),
      FernHorchBeenden: () => {
        laeuft = false

        return gerufen()
      },
    })

    const nutzer = zeige()

    await nutzer.click(
      await screen.findByRole('button', { name: texte.fernBeenden }),
    )

    expect(await screen.findByText(texte.fernStill)).toBeInTheDocument()
    expect(gerufen).toHaveBeenCalledTimes(1)
  })

  it('meldet einen Fehler am Kasten', async () => {
    belegeApi({
      FernHorchStand: () => Promise.resolve(einStand(false, '')),
      FernHorchStarten: () =>
        Promise.resolve({
          stand: { laeuft: false, einladung: '', wege: [], angebunden: 0 },
          fehler: {
            code: 'validierung',
            schluessel: 'unbekannt',
            meldung: 'kein DERP erreichbar',
            feld: '',
            werte: {},
            vorgang: '',
          },
        } as Fernantwort),
    })

    const nutzer = zeige()

    await nutzer.click(
      await screen.findByRole('button', { name: texte.fernStarten }),
    )

    // **Der Satz kommt aus dem Katalog, nicht aus Go** (§9): die Meldung
    // trägt einen Schlüssel, und das Fenster schlägt ihn nach. Mein erster
    // Anlauf erwartete den rohen Satz — und prüfte damit, was die
    // Oberfläche gerade **nicht** tun soll.
    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.unbekannt,
    )
  })

  // **Die Einladung lässt sich kopieren.**
  //
  // **Erst `setup`, dann die Zwischenablage** — `userEvent.setup()` legt eine
  // eigene an und überschreibt, was vorher dastand. Der Grund steht
  // ausführlich in `agentenkonfiguration.test.tsx`.
  it('kopiert die Einladung', async () => {
    belegeApi({
      FernHorchStand: () =>
        Promise.resolve(einStand(true, 'mse4mac:1:tc1abcdef')),
    })

    const hand = zeige()

    await screen.findByText('mse4mac:1:tc1abcdef')

    const geschrieben = vi.fn(() => Promise.resolve())
    setzeZwischenablage(geschrieben)

    await hand.click(screen.getByRole('button', { name: texte.fernKopieren }))

    expect(await screen.findByText(texte.fernKopiert)).toBeInTheDocument()
    expect(geschrieben).toHaveBeenCalledWith('mse4mac:1:tc1abcdef')
  })

  // **Und ohne Erlaubnis steht da, warum nicht** — statt still nichts zu tun.
  it('meldet, wenn das Kopieren nicht geht', async () => {
    belegeApi({
      FernHorchStand: () =>
        Promise.resolve(einStand(true, 'mse4mac:1:tc1abcdef')),
    })

    const hand = zeige()

    await screen.findByText('mse4mac:1:tc1abcdef')

    setzeZwischenablage(
      vi.fn(() => Promise.reject(new Error('keine Erlaubnis'))),
    )

    await hand.click(screen.getByRole('button', { name: texte.fernKopieren }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      texte.kopierfehler,
    )
  })

  // **Ein Ladefehler steht am Kasten** — und nicht nur einer, der beim
  // Schalten entsteht. Ohne diesen Lauf bliebe die Frage offen, was ein
  // Mensch sieht, wenn schon das Nachfragen scheitert.
  it('meldet einen Fehler beim Nachfragen', async () => {
    belegeApi({
      FernHorchStand: () =>
        Promise.resolve({
          stand: { laeuft: false, einladung: '', wege: [], angebunden: 0 },
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            meldung: 'der Bestand ist zu',
            feld: '',
            werte: {},
            vorgang: '',
          },
        } as Fernantwort),
    })

    zeige()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.unbekannt,
    )
  })
})

// **Über `defineProperty` und nicht `Object.assign`**: nach
// `userEvent.setup()` ist `navigator.clipboard` ein **Getter**, und eine
// Zuweisung darauf wirft. Der Grund steht ausführlich in
// `agentenkonfiguration.test.tsx`.
function setzeZwischenablage(schreiben: () => Promise<void>) {
  Object.defineProperty(navigator, 'clipboard', {
    configurable: true,
    value: { writeText: schreiben },
  })
}
