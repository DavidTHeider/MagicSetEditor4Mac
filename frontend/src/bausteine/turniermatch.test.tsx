// **Eine Paarung, an eine Partie gebunden** — Werkstück G7 von Phase 5.5
// (#1464).
//
// # Wogegen hier geprüft wird
//
//   - **Die Grenze steht da**: der Nachweis belegt denselben Verlauf, nicht
//     richtiges Spiel — und eine private Runde ist keine Zulassung.
//   - Der Weg: **binden → melden → übernehmen**.
//   - Ohne Bindung stehen die Felder da, mit Bindung die Sitze.
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Turniermatch } from './Turniermatch'
import { antwortAbwarten, belegeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'

import type { MatchAntwort } from '../apiturniere'

const eigene = kataloge.de.seiten.turniere

// einStand gibt eine Antwort mit den genannten Abweichungen.
function einStand(abweichend: Partial<MatchAntwort> = {}): MatchAntwort {
  return {
    gebunden: false,
    paarungID: '7',
    partieID: '',
    fassung: '3',
    sitze: [],
    meldungen: [],
    uebernommen: null,
    fehler: null,
    ...abweichend,
  }
}

// gebunden gibt eine Antwort mit zwei Sitzen.
function gebunden(abweichend: Partial<MatchAntwort> = {}): MatchAntwort {
  return einStand({
    gebunden: true,
    partieID: 'p1',
    sitze: [
      { teilnehmerID: '10', sitz: 'anna', peer: '', deckabdruck: 'abdruckA' },
      { teilnehmerID: '11', sitz: 'bert', peer: '', deckabdruck: 'abdruckB' },
    ],
    ...abweichend,
  })
}

// zeige zeichnet das Bauteil und gibt die Hand.
function zeige() {
  render(
    <Sprachraum sprache="de">
      <Turniermatch turnierID="1" paarungID="7" teilnehmer={['10', '11']} />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

it('nennt den Abschnitt und zeigt die Grenze', () => {
  belegeApi({ TurnierMatch: vi.fn() })

  zeige()

  expect(screen.getByRole('region', { name: eigene.match })).toBeInTheDocument()

  expect(screen.getByText(eigene.matchGrenze)).toBeInTheDocument()
  expect(screen.getByText(kurzform(eigene.matchErklaerung))).toBeInTheDocument()
})

it('holt den Stand nicht von selbst', () => {
  const geholt = vi.fn()
  belegeApi({ TurnierMatch: geholt })

  zeige()

  expect(screen.queryByText(eigene.matchNichtGebunden)).not.toBeInTheDocument()
  expect(geholt).not.toHaveBeenCalled()
})

it('sagt, wenn keine Partie daran hängt', async () => {
  belegeApi({ TurnierMatch: vi.fn().mockResolvedValue(einStand()) })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.matchAnsehen }))

  expect(await screen.findByText(eigene.matchNichtGebunden)).toBeInTheDocument()
})

it('bindet die Paarung an eine Partie', async () => {
  const gebundenRuf = vi.fn().mockResolvedValue(gebunden())

  belegeApi({
    TurnierMatch: vi.fn().mockResolvedValue(einStand()),
    TurnierMatchBinden: gebundenRuf,
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.matchAnsehen }))

  await hand.type(await screen.findByLabelText(eigene.matchPartie), 'p1')
  await hand.type(screen.getByLabelText(eigene.matchSitzA), 'anna')
  await hand.type(screen.getByLabelText(eigene.matchSitzB), 'bert')

  await hand.click(screen.getByRole('button', { name: eigene.matchBinden }))

  await antwortAbwarten()

  expect(gebundenRuf).toHaveBeenCalledWith('1', '7', 'p1', [
    { teilnehmerID: '10', sitz: 'anna', peer: '', deckabdruck: '' },
    { teilnehmerID: '11', sitz: 'bert', peer: '', deckabdruck: '' },
  ])

  expect(screen.getByText(/abdruckA/)).toBeInTheDocument()
})

it('meldet für einen Sitz und übernimmt', async () => {
  const gemeldet = vi.fn().mockResolvedValue(
    gebunden({
      meldungen: [
        {
          von: 'anna',
          ergebnis: { spieleA: 2, spieleB: 1, unentschieden: 0 },
          kette: 'glied',
        },
      ],
    }),
  )

  const uebernommen = vi
    .fn()
    .mockResolvedValue(
      gebunden({ uebernommen: { spieleA: 2, spieleB: 1, unentschieden: 0 } }),
    )

  belegeApi({
    TurnierMatch: vi.fn().mockResolvedValue(gebunden()),
    TurnierMatchMelden: gemeldet,
    TurnierMatchUebernehmen: uebernommen,
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.matchAnsehen }))

  expect(await screen.findByText(eigene.matchKeineMeldung)).toBeInTheDocument()

  await hand.type(screen.getByLabelText(eigene.matchNachweis), 'glied')
  await hand.type(screen.getByLabelText(eigene.spieleA), '2')
  await hand.type(screen.getByLabelText(eigene.spieleB), '1')

  await hand.click(screen.getByRole('button', { name: eigene.matchMelden }))

  await antwortAbwarten()

  expect(gemeldet).toHaveBeenCalledWith('7', 'anna', 'glied', {
    spieleA: 2,
    spieleB: 1,
    unentschieden: 0,
  })

  expect(
    screen.getByText(`2:1:0 ${eigene.matchMeldungVon} anna`),
  ).toBeInTheDocument()

  await hand.click(
    screen.getByRole('button', { name: eigene.matchUebernehmen }),
  )

  expect(await screen.findByText(eigene.matchUebernommen)).toBeInTheDocument()

  expect(uebernommen).toHaveBeenCalledWith('1', '7')
})

// **Eine unsinnige Zahl bleibt am Feld stehen** — F03
// (#2245).
// Bis dahin wurde sie zu null und ging als Ergebnis hinaus: „viele" gegen
// „2" meldete ein 0:2, das niemand getippt hatte, und „2x" ein 2.
it('weist eine unsinnige Zahl am Feld ab und meldet nichts', async () => {
  const gemeldet = vi.fn().mockResolvedValue(gebunden())

  belegeApi({
    TurnierMatch: vi.fn().mockResolvedValue(gebunden()),
    TurnierMatchMelden: gemeldet,
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.matchAnsehen }))

  await hand.type(await screen.findByLabelText(eigene.matchNachweis), 'glied')
  await hand.type(screen.getByLabelText(eigene.spieleA), 'viele')
  await hand.type(screen.getByLabelText(eigene.spieleB), '2')

  await hand.click(screen.getByRole('button', { name: eigene.matchMelden }))

  const feld = screen.getByLabelText(eigene.spieleA)

  expect(feld).toHaveAttribute('aria-invalid', 'true')
  expect(feld).toHaveAccessibleDescription(
    einsetzen(kataloge.de.allgemein.ganzeZahlAb, { kleinstens: '0' }),
  )
  expect(screen.getByLabelText(eigene.spieleB)).not.toHaveAttribute(
    'aria-invalid',
    'true',
  )
  expect(gemeldet).not.toHaveBeenCalled()
})

// **Ein leeres Feld ist null** — dabei bleibt es: wer kein Unentschieden
// hatte, tippt keines.
it('nimmt ein leeres Feld als null', async () => {
  const gemeldet = vi.fn().mockResolvedValue(gebunden())

  belegeApi({
    TurnierMatch: vi.fn().mockResolvedValue(gebunden()),
    TurnierMatchMelden: gemeldet,
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.matchAnsehen }))

  await hand.type(await screen.findByLabelText(eigene.matchNachweis), 'glied')
  await hand.type(screen.getByLabelText(eigene.spieleA), ' 2 ')

  await hand.click(screen.getByRole('button', { name: eigene.matchMelden }))

  await antwortAbwarten()

  expect(gemeldet).toHaveBeenCalledWith('7', 'anna', 'glied', {
    spieleA: 2,
    spieleB: 0,
    unentschieden: 0,
  })
})

it('meldet einen Fehler', async () => {
  belegeApi({
    TurnierMatch: vi.fn().mockResolvedValue(
      einStand({
        fehler: {
          code: 'konflikt',
          schluessel: 'konflikt',
          meldung: 'ging nicht',
        },
      }),
    ),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.matchAnsehen }))

  expect(await screen.findByRole('alert')).toBeInTheDocument()
})

// **Eine Bindung ohne Sitze bricht nichts** — §8: fremde und alte Daten sind
// der Normalfall.
//
// **Strukturell gibt es sie nicht** (eine Bindung hat zwei Sitze, und Go weist
// alles andere ab). Eine Oberfläche, die daran zerbräche, zerbräche an einer
// älteren Fassung oder an einer Zeile, die jemand von Hand geändert hat.
it('kommt mit einer Bindung ohne Sitze zurecht', async () => {
  belegeApi({
    TurnierMatch: vi.fn().mockResolvedValue(gebunden({ sitze: [] })),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.matchAnsehen }))

  expect(await screen.findByText(eigene.matchKeineMeldung)).toBeInTheDocument()

  expect(
    screen.getByRole('button', { name: eigene.matchMelden }),
  ).toBeDisabled()
})

// **Und der Weg ohne Tippen** — M17 (#1618).
//
// Die dritte Abnahme des Issues: „Das Ergebnis erreicht die Turnierbindung
// aus G7, **ohne dass jemand es eintippt**." Weder Spielstand noch Nachweis
// gehen hier über die Tastatur — allein der Sitz sagt, wer meldet, und den
// füllt die Bindung vor.
it('meldet das Ergebnis aus der Partie, ohne dass jemand es tippt', async () => {
  const ausDerPartie = vi.fn().mockResolvedValue(
    gebunden({
      meldungen: [
        {
          von: 'anna',
          ergebnis: { spieleA: 1, spieleB: 0, unentschieden: 0 },
          kette: 'glied',
        },
      ],
    }),
  )

  belegeApi({
    TurnierMatch: vi.fn().mockResolvedValue(gebunden()),
    TurnierMatchAusDerPartie: ausDerPartie,
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.matchAnsehen }))

  expect(await screen.findByText(eigene.matchKeineMeldung)).toBeInTheDocument()

  await hand.click(
    screen.getByRole('button', { name: eigene.matchAusDerPartie }),
  )

  await antwortAbwarten()

  expect(ausDerPartie).toHaveBeenCalledWith('7', 'anna')

  expect(
    screen.getByText(`1:0:0 ${eigene.matchMeldungVon} anna`),
  ).toBeInTheDocument()
})

// **Ohne Sitz meldet auch dieser Weg nicht.**
//
// Die Gegenprobe zum Lauf darüber: der Nachweis fehlt hier absichtlich, und
// trotzdem ist der Knopf bedienbar — was ihn anhält, ist allein der leere
// Sitz.
it('meldet ohne Sitz auch aus der Partie nicht', async () => {
  belegeApi({
    TurnierMatch: vi.fn().mockResolvedValue(gebunden({ sitze: [] })),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.matchAnsehen }))

  expect(await screen.findByText(eigene.matchKeineMeldung)).toBeInTheDocument()

  expect(
    screen.getByRole('button', { name: eigene.matchAusDerPartie }),
  ).toBeDisabled()
})

// **Gebunden ist gebunden** (§7, Regel 3; #1970): kam der Stand danach
// nicht, sagt das Bauteil das an seiner Stelle — und nicht „keine Partie
// daran", was die leere Antwort behauptete.
it('nimmt ein Binden mit Nachladefehler als geschehen', async () => {
  belegeApi({
    TurnierMatch: vi.fn().mockResolvedValue(einStand()),
    TurnierMatchBinden: vi.fn().mockResolvedValue(
      einStand({
        nachladefehler: {
          code: 'validierung',
          schluessel: '',
          meldung: 'Der Stand klemmt',
        },
      }),
    ),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.matchAnsehen }))
  await hand.type(await screen.findByLabelText(eigene.matchPartie), 'p1')
  await hand.type(screen.getByLabelText(eigene.matchSitzA), 'anna')
  await hand.type(screen.getByLabelText(eigene.matchSitzB), 'bert')
  await hand.click(screen.getByRole('button', { name: eigene.matchBinden }))

  expect(await screen.findByRole('alert')).toHaveTextContent('Der Stand klemmt')
  expect(screen.queryByText(eigene.matchNichtGebunden)).toBeNull()
})
