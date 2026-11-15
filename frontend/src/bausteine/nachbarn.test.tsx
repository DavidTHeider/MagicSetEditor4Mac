// **Wer in der Nähe ist** — Werkstück N2 von Phase 5.4
// (#1455).
//
// # Was hier gemessen wird
//
// **Dass nichts von selbst ins Netz ruft** — die Liste kommt auf Nachfrage.
// **Dass ein Fund kein Kontakt ist** — wer schon in der Liste steht, bekommt
// keinen Knopf. **Und dass beim Aufnehmen die Adresse mitgeht**: sie ist der
// Grund, warum dieser Weg mehr ist als ein Abtippen der Kennung.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Nachbarn } from './Nachbarn'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'

import type { Nachbar, Nachbarnantwort } from '../apipeerkontakt'

const eigene = kataloge.de.seiten.technik

// einNachbar baut einen Fund.
function einNachbar(anders: Partial<Nachbar> = {}): Nachbar {
  return {
    kennung: 'bnachbar',
    fingerabdruck: '1111-2222-3333-4444-5555-6666-7777-8888',
    name: 'Annas Mac',
    adresse: '192.168.1.7:41234',
    merkadresse: 'annas-mac.local:41234',
    bekannt: false,
    stand: '',
    gemerkteAdresse: '',
    adresseGeaendert: false,
    ...anders,
  }
}

// eineAntwort baut die Liste.
function eineAntwort(
  nachbarn: Nachbar[] = [einNachbar()],
  verworfen = 0,
  fehler: Nachbarnantwort['fehler'] = null,
): Nachbarnantwort {
  return { nachbarn, verworfen, fehler }
}

// zeige stellt den Kasten auf.
function zeige() {
  render(
    <Sprachraum sprache="de">
      <Nachbarn />
    </Sprachraum>,
  )

  return userEvent.setup()
}

describe('Wer in der Nähe ist', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  // **Der Lauf, für den der Knopf da ist.**
  it('ruft erst auf Nachfrage ins Netz', async () => {
    const gerufen = vi.fn(() => Promise.resolve(eineAntwort()))

    belegeApi({ Nachbarn: gerufen })

    const hand = zeige()

    expect(gerufen).not.toHaveBeenCalled()

    await hand.click(
      screen.getByRole('button', { name: eigene.nachbarnSuchen }),
    )

    expect(await screen.findByText(/Annas Mac/)).toBeInTheDocument()
    expect(gerufen).toHaveBeenCalledOnce()
  })

  it('sagt, wenn sich niemand meldet', async () => {
    belegeApi({ Nachbarn: () => Promise.resolve(eineAntwort([])) })

    const hand = zeige()

    await hand.click(
      screen.getByRole('button', { name: eigene.nachbarnSuchen }),
    )

    expect(await screen.findByText(eigene.nachbarnLeer)).toBeInTheDocument()
  })

  // **Der Fund, der keiner ist, steht dabei.**
  it('nennt die Antworten, die keine Peerkennung trugen', async () => {
    belegeApi({ Nachbarn: () => Promise.resolve(eineAntwort([], 3)) })

    const hand = zeige()

    await hand.click(
      screen.getByRole('button', { name: eigene.nachbarnSuchen }),
    )

    expect(await screen.findByText(/3 Antworten/)).toBeInTheDocument()
  })

  it('zeigt einen unbekannten Fund als unbekannt', async () => {
    belegeApi({ Nachbarn: () => Promise.resolve(eineAntwort()) })

    const hand = zeige()

    await hand.click(
      screen.getByRole('button', { name: eigene.nachbarnSuchen }),
    )

    expect(
      await screen.findByText(eigene.nachbarnUnbekannt),
    ).toBeInTheDocument()

    expect(
      screen.getByRole('button', { name: eigene.nachbarnAufnehmen }),
    ).toBeInTheDocument()
  })

  // **Wer schon dasteht, bekommt keinen Knopf.**
  it('bietet einem bekannten Fund kein zweites Eintragen an', async () => {
    belegeApi({
      Nachbarn: () =>
        Promise.resolve(
          eineAntwort([einNachbar({ bekannt: true, stand: 'bestaetigt' })]),
        ),
    })

    const hand = zeige()

    await hand.click(
      screen.getByRole('button', { name: eigene.nachbarnSuchen }),
    )

    expect(
      await screen.findByText(eigene.nachbarnSchonBekannt),
    ).toBeInTheDocument()

    expect(
      screen.queryByRole('button', { name: eigene.nachbarnAufnehmen }),
    ).not.toBeInTheDocument()

    expect(screen.getByText(eigene.kontakteStandBestaetigt)).toBeInTheDocument()
  })

  // **Der Lauf, für den es das Aufnehmen gibt.**
  //
  // **Eingetragen wird die Merkadresse und nicht die, über die gewählt
  // wird** (#1633):
  // eine IP von heute ist nach dem nächsten DHCP-Wechsel falsch und zeigte
  // dann auf einen Rechner, der inzwischen jemand anderem gehört.
  it('trägt beim Aufnehmen den Namen ein, nicht die IP', async () => {
    const aufgenommen: string[] = []
    const adressen: string[] = []

    belegeApi({
      Nachbarn: () => Promise.resolve(eineAntwort()),
      PeerkontaktAufnehmen: (kennung: string, name: string) => {
        aufgenommen.push(kennung + '/' + name)

        return Promise.resolve({ kontakte: [], zwecke: [], fehler: null })
      },
      PeerkontaktAdresse: (kennung: string, adresse: string) => {
        adressen.push(kennung + '/' + adresse)

        return Promise.resolve({ kontakte: [], zwecke: [], fehler: null })
      },
    })

    const hand = zeige()

    await hand.click(
      screen.getByRole('button', { name: eigene.nachbarnSuchen }),
    )

    await hand.click(
      await screen.findByRole('button', { name: eigene.nachbarnAufnehmen }),
    )

    await screen.findByText(eigene.nachbarnUnbekannt)

    expect(aufgenommen).toEqual(['bnachbar/Annas Mac'])
    expect(adressen).toEqual(['bnachbar/annas-mac.local:41234'])
  })

  // **Und wo es keinen Namen gibt, bleibt die Adresse** (#1633).
  //
  // **Die Gegenprobe zum Lauf darüber**: ohne sie sagte er nur, dass
  // irgendein Feld eingetragen wird — nicht, dass der Fall ohne Namen
  // seinen Weg behält.
  it('trägt ohne Namen die Adresse ein', async () => {
    const adressen: string[] = []

    belegeApi({
      Nachbarn: () =>
        Promise.resolve(
          eineAntwort([einNachbar({ merkadresse: '192.168.1.7:41234' })]),
        ),
      PeerkontaktAufnehmen: () =>
        Promise.resolve({ kontakte: [], zwecke: [], fehler: null }),
      PeerkontaktAdresse: (kennung: string, adresse: string) => {
        adressen.push(kennung + '/' + adresse)

        return Promise.resolve({ kontakte: [], zwecke: [], fehler: null })
      },
    })

    const hand = zeige()

    await hand.click(
      screen.getByRole('button', { name: eigene.nachbarnSuchen }),
    )

    await hand.click(
      await screen.findByRole('button', { name: eigene.nachbarnAufnehmen }),
    )

    await screen.findByText(eigene.nachbarnUnbekannt)

    expect(adressen).toEqual(['bnachbar/192.168.1.7:41234'])

    // **Und die Zeile »Gemerkt wird« steht dann gar nicht da** — eine
    // Wiederholung erklärt nichts.
    expect(screen.queryByText(eigene.nachbarnGemerkt)).not.toBeInTheDocument()
  })

  it('zeigt einen Fehler des Dienstes', async () => {
    belegeApi({
      Nachbarn: () =>
        Promise.resolve(
          eineAntwort([], 0, {
            code: 'validierung',
            schluessel: 'fehler.validierung',
            meldung: 'Das Netz nimmt niemanden auf.',
          }),
        ),
    })

    const hand = zeige()

    await hand.click(
      screen.getByRole('button', { name: eigene.nachbarnSuchen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Das Netz nimmt niemanden auf.',
    )
  })

  it('zeigt einen Sturz der Naht', async () => {
    belegeApi({ Nachbarn: () => Promise.reject(new Error('kaputt')) })

    const hand = zeige()

    await hand.click(
      screen.getByRole('button', { name: eigene.nachbarnSuchen }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })

  // **Ein Fund ohne Namen zeigt keine leere Zeile.**
  it('zeigt keine Namenszeile, wo keiner dasteht', async () => {
    belegeApi({
      Nachbarn: () => Promise.resolve(eineAntwort([einNachbar({ name: '' })])),
    })

    const hand = zeige()

    await hand.click(
      screen.getByRole('button', { name: eigene.nachbarnSuchen }),
    )

    await screen.findByText(eigene.nachbarnUnbekannt)

    expect(
      screen.queryByText(kurzform(eigene.kontakteNenntHinweis)),
    ).not.toBeInTheDocument()
  })
  // **Ein vierter Stand ist immer noch eine Auskunft.**
  //
  // **Der Rückfall ist der Stand selbst** — dieselbe Überlegung wie bei den
  // Kontakten nebenan: ein Stand, den dieses Fenster noch nicht kennt, soll
  // dastehen und nicht verschwinden.
  it('zeigt einen unbekannten Stand, statt ihn zu verschlucken', async () => {
    belegeApi({
      Nachbarn: () =>
        Promise.resolve(
          eineAntwort([einNachbar({ bekannt: true, stand: 'ganzneu' })]),
        ),
    })

    const hand = zeige()

    await hand.click(
      screen.getByRole('button', { name: eigene.nachbarnSuchen }),
    )

    expect(await screen.findByText('ganzneu')).toBeInTheDocument()
  })
})
