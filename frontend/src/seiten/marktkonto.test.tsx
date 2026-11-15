// **Das eigene Konto bei einem Marktplatz** — Werkstück T11 von Phase 5.6
// (#1476).
//
// # Was hier gemessen wird
//
// **Dass beim Zeichnen nichts hinausgeht.** Ein Abruf ohne Nachfrage ginge
// zu einem fremden Dienst, ohne dass jemand danach gefragt hat (§16, Regel 5).
//
// **Dass an jeder Bestellung steht, ob sie Bestand bewegt hat** — und dass
// das eine Auskunft bleibt: gebucht wird von hier nichts.
//
// **Und dass eine ausgeschaltete Kontoschau als Meldung ankommt** statt als
// leere Liste, die aussähe wie ein leeres Konto.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Marktkonto } from '../bausteine/Marktkonto'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.handeln

// einStand ist eine Antwort mit einem Angebot und zwei Bestellungen.
function einStand() {
  return {
    stand: {
      name: 'Test App',
      nutzerID: 51242,
      angebote: [
        {
          id: 1,
          blueprintID: 7480,
          name: 'Tranquil Cove',
          menge: 2,
          preisCent: 250,
          waehrung: 'EUR',
          eigennotiz: 'Kiste 4',
          beschreibung: '',
        },
      ],
      bestellungen: [
        {
          id: 9,
          code: 'AB12',
          zustand: 'hub_pending',
          ueberNullkonto: true,
          rolle: 'seller',
          stueck: 3,
          bezahlt: '',
          bewegtBestand: true,
        },
        {
          id: 10,
          code: 'CD34',
          zustand: 'paid',
          ueberNullkonto: true,
          rolle: 'seller',
          stueck: 3,
          bezahlt: '',
          bewegtBestand: false,
        },
      ],
      befunde: ['CardTrader nennt den Zustand „verschollen"'],
    },
    fehler: null,
  }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Marktkonto />
    </Sprachraum>,
  )

  return userEvent.setup()
}

describe('Das eigene Konto bei einem Marktplatz', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('fragt beim Zeichnen niemanden', () => {
    const gefragt = vi.fn()

    belegeApi({ Marktkonto: gefragt })

    zeige()

    expect(gefragt).not.toHaveBeenCalled()
    expect(
      screen.getByRole('button', { name: eigene.kontoNachsehen }),
    ).toBeInTheDocument()
  })

  it('zeigt Angebote, Bestellungen und ihre Wirkung auf den Bestand', async () => {
    const gefragt = vi.fn().mockResolvedValue(einStand())

    belegeApi({ Marktkonto: gefragt })

    const hand = zeige()

    await hand.click(
      screen.getByRole('button', { name: eigene.kontoNachsehen }),
    )

    expect(await screen.findByRole('status')).toHaveTextContent('Test App')

    // **Die Blätterangaben gehen als Null hinaus** — das heißt „nimm die
    // Vorgabe des Anbieters" und nicht „nimm alles".
    expect(gefragt).toHaveBeenCalledWith('cardtrader', 0, 0)

    expect(screen.getByText(/2 × Tranquil Cove/)).toBeInTheDocument()

    // **Beide Bestellungen stehen da, und sie sagen Verschiedenes** — eine
    // Anzeige, die beide gleich beschriftete, verdeckte genau die Regel, für
    // die es dieses Werkstück gibt.
    expect(
      screen.getByText(new RegExp(`AB12.*${eigene.kontoBewegt}`)),
    ).toBeInTheDocument()

    expect(
      screen.getByText(new RegExp(`CD34.*${eigene.kontoBewegtNicht}`)),
    ).toBeInTheDocument()

    // **Und die Befunde stehen dabei** — ein Zustand, den dieses Programm
    // nicht kennt, soll nicht still durchgehen.
    expect(screen.getByText(/verschollen/)).toBeInTheDocument()
  })

  // **Ein Konto mit echten Mengen** — UXR-15 (#1923): ein Händlerkonto hat
  // Hunderte Angebote, und jedes stand als eigene Zeile untereinander (R4 in
  // `docs/UIREVIEW.md`). Jetzt zehn je Liste und der Rest stückweise, und
  // jede Liste trägt ihren Namen.
  it('zeigt zehn Angebote und zehn Bestellungen und den Rest stückweise', async () => {
    const voll = einStand()
    const [angebot] = voll.stand.angebote
    const [bestellung] = voll.stand.bestellungen
    if (angebot === undefined || bestellung === undefined) {
      throw new Error('die Probe hat kein Angebot')
    }
    voll.stand.angebote = Array.from({ length: 45 }, (_, stelle) => ({
      ...angebot,
      id: stelle + 1,
      name: `Mappenkarte ${String(stelle + 1)}`,
    }))
    voll.stand.bestellungen = Array.from({ length: 25 }, (_, stelle) => ({
      ...bestellung,
      id: stelle + 1,
      code: `B${String(stelle + 1)}`,
    }))
    belegeApi({ Marktkonto: () => Promise.resolve(voll) })

    const hand = zeige()
    await hand.click(
      screen.getByRole('button', { name: eigene.kontoNachsehen }),
    )

    const angebote = await screen.findByRole('list', {
      name: eigene.kontoAngebote,
    })
    expect(within(angebote).getAllByRole('listitem')).toHaveLength(10)
    expect(
      within(
        screen.getByRole('list', { name: eigene.kontoBestellungen }),
      ).getAllByRole('listitem'),
    ).toHaveLength(10)
    expect(screen.getByText('10 von 45')).toBeInTheDocument()
    expect(screen.getByText('10 von 25')).toBeInTheDocument()

    // **Jede Liste ihr Knopf**: der der Angebote zeigt die nächsten zwanzig,
    // danach fehlen dort wie bei den Bestellungen fünfzehn.
    await hand.click(screen.getByRole('button', { name: 'Weitere 20 zeigen' }))
    expect(within(angebote).getAllByRole('listitem')).toHaveLength(30)
    expect(
      screen.getAllByRole('button', { name: 'Weitere 15 zeigen' }),
    ).toHaveLength(2)
  })

  it('meldet, wenn die Kontoschau aus ist', async () => {
    belegeApi({
      Marktkonto: () => Promise.reject(new Error('die Kontoschau ist aus')),
    })

    const hand = zeige()

    await hand.click(
      screen.getByRole('button', { name: eigene.kontoNachsehen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'die Kontoschau ist aus',
    )

    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })
})
