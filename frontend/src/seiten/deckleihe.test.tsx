// **Ein Deck verleihen oder leihen** — Werkstück G6 von Phase 5.5
// (#1463).
//
// # Wogegen hier geprüft wird
//
//   - Die **drei Sätze** stehen da: Nutzungsfreigabe, kein Bestandswechsel,
//     kein Kopierschutz. Das Issue verlangt sie wörtlich.
//   - Der Wunsch geht **so** hinaus, wie er gewählt wurde — Entleiher und
//     Deck aus Listen, nicht getippt (UX-08.AK-04, #1680).
//   - Eine **verliehene** Leihe wird widerrufen, eine **geliehene**
//     zurückgegeben — zwei Knöpfe für zwei Rollen.
import { screen, waitFor } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { Vorgabestunden } from './Deckleihe'
import { antwortAbwarten } from '../test/api'
import { einDeck } from './partieprobe'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'
import {
  eigene,
  eineLeihe,
  gewaehlt,
  leer,
  leiheApi,
  oeffneAnnahme,
  umschlagGewaehlt,
  waehleBertUndDeck,
  waehleUmschlag,
  waehleZiel,
  zeige,
} from '../test/leiheprobe'

// **Benannt wird die Leihe von ihrer Schublade** — seit UXR-13 (#1921)
// steht sie hinter „Verleihen oder leihen …" auf der Spielenseite, und die
// Schublade trägt ihren Namen (`spielenschubladen.test.tsx`). Hier bleiben die
// drei Sätze, die das Issue wörtlich verlangt.
it('zeigt die drei Grenzen', () => {
  leiheApi({ Deckleihen: vi.fn() })

  zeige()

  for (const satz of [
    eigene.leiheErklaerung,
    eigene.leiheKeinEigentum,
    eigene.leiheKeinKopierschutz,
    eigene.leiheRuecknahmeGrenze,
  ]) {
    expect(screen.getByText(kurzform(satz))).toBeInTheDocument()
  }
})

it('holt die Leihen nicht von selbst', () => {
  const geholt = vi.fn()
  leiheApi({ Deckleihen: geholt })

  zeige()

  expect(screen.queryByText(eigene.leiheKeine)).not.toBeInTheDocument()
  expect(geholt).not.toHaveBeenCalled()
})

it('sagt, wenn es keine Leihe gibt', async () => {
  leiheApi({ Deckleihen: vi.fn().mockResolvedValue(leer) })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.leiheAnsehen }))

  expect(await screen.findByText(eigene.leiheKeine)).toBeInTheDocument()
})

// **Ohne Gegenüber, Deck und Ziel geht nichts hinaus.**
it('hält den Knopf an, solange etwas fehlt', () => {
  leiheApi({ DeckleiheAnbieten: vi.fn() })

  zeige()

  expect(
    screen.getByRole('button', { name: eigene.leiheAnbieten }),
  ).toBeDisabled()
})

it('schickt den Wunsch mit gewähltem Entleiher und Deck', async () => {
  const angeboten = vi.fn().mockResolvedValue({
    leihen: [eineLeihe()],
    pfad: '/tmp/leihe.msepackage',
    fehler: null,
  })

  leiheApi({
    DeckleiheAnbieten: angeboten,
    DeckleiheZielWaehlen: vi
      .fn()
      .mockResolvedValue(gewaehlt('/tmp/leihe.msepackage')),
  })

  const hand = zeige()

  await waehleBertUndDeck(hand)
  await hand.click(screen.getByRole('button', { name: eigene.leiheFachleute }))
  await hand.type(screen.getByLabelText(eigene.leiheWoran), 'p1')
  await waehleZiel(hand, '/tmp/leihe.msepackage')

  await hand.click(screen.getByRole('button', { name: eigene.leiheAnbieten }))

  await waitFor(() => {
    expect(angeboten).toHaveBeenCalledWith(
      {
        entleiher: '12D3KooWBert',
        deckID: '7',
        bindungsart: 'partie',
        woran: 'p1',
        rechte: ['spielen'],
        stunden: Vorgabestunden,
        modus: 'zuende',
      },
      '/tmp/leihe.msepackage',
    )
  })

  // **Der abgelegte Umschlag wird genannt** — neben dem gewählten Ziel, das
  // dort weiter steht.
  expect(await screen.findByRole('status')).toHaveTextContent(
    `${eigene.leiheAbgelegt} /tmp/leihe.msepackage`,
  )
})

// **Eine Stundenzahl, die keine ganze ab eins ist, hält das Anbieten an** —
// F03 (#2245).
// Bis dahin wurden „0", „-2" und „0.5" still zur Vorgabe und „1.5" zu einer
// Stunde: eine Leihe, die niemand so getippt hatte. **Und nach oben**
// (J05, #2269): 5.124.096 Stunden liefen in Go zu 25 Minuten um.
it.each(['0', '-2', '0.5', '1.5', '2562048', '5124096'])(
  'hält das Anbieten bei %j am Feld an',
  async (roh) => {
    const angeboten = vi.fn().mockResolvedValue(leer)
    leiheApi({
      DeckleiheAnbieten: angeboten,
      DeckleiheZielWaehlen: vi
        .fn()
        .mockResolvedValue(gewaehlt('/tmp/l.msepackage')),
    })

    const hand = zeige()

    await waehleBertUndDeck(hand)
    await waehleZiel(hand, '/tmp/l.msepackage')

    const feld = screen.getByLabelText(eigene.leiheStunden)
    await hand.clear(feld)
    await hand.type(feld, roh)

    expect(feld).toHaveAttribute('aria-invalid', 'true')
    expect(feld).toHaveAccessibleDescription(
      einsetzen(kataloge.de.allgemein.ganzeZahlBereich, {
        kleinstens: '1',
        groesstens: '2562047',
      }),
    )
    expect(
      screen.getByRole('button', { name: eigene.leiheAnbieten }),
    ).toBeDisabled()
    expect(angeboten).not.toHaveBeenCalled()
  },
)

// **Ein leeres Feld ist keine Stundenzahl** — der Knopf wartet, bis eine
// dasteht, statt still die Vorgabe zu nehmen. Ein Fehler ist es nicht: wer
// eine Zahl ändert, löscht sie zuerst.
it('hält das Anbieten an, solange das Feld leer ist', async () => {
  leiheApi({
    DeckleiheZielWaehlen: vi
      .fn()
      .mockResolvedValue(gewaehlt('/tmp/l.msepackage')),
  })

  const hand = zeige()

  await waehleBertUndDeck(hand)
  await waehleZiel(hand, '/tmp/l.msepackage')

  const feld = screen.getByLabelText(eigene.leiheStunden)
  await hand.clear(feld)

  expect(feld).not.toHaveAttribute('aria-invalid', 'true')
  expect(
    screen.getByRole('button', { name: eigene.leiheAnbieten }),
  ).toBeDisabled()
})

// **Und eine richtige geht durch** — die Gegenprobe zu den beiden darüber.
it('bietet mit der getippten Stundenzahl an', async () => {
  const angeboten = vi.fn().mockResolvedValue(leer)
  leiheApi({
    DeckleiheAnbieten: angeboten,
    DeckleiheZielWaehlen: vi
      .fn()
      .mockResolvedValue(gewaehlt('/tmp/l.msepackage')),
  })

  const hand = zeige()

  await waehleBertUndDeck(hand)
  await waehleZiel(hand, '/tmp/l.msepackage')

  const feld = screen.getByLabelText(eigene.leiheStunden)
  await hand.clear(feld)
  await hand.type(feld, '48')

  await hand.click(screen.getByRole('button', { name: eigene.leiheAnbieten }))

  await waitFor(() => {
    expect(angeboten).toHaveBeenCalledWith(
      expect.objectContaining({ stunden: 48 }),
      '/tmp/l.msepackage',
    )
  })
})

it('nimmt einen Umschlag an', async () => {
  const genommen = vi.fn().mockResolvedValue(leer)
  leiheApi({
    DeckleiheAnnehmen: genommen,
    UmschlagWaehlen: vi
      .fn()
      .mockResolvedValue(umschlagGewaehlt('/tmp/fremd.msepackage')),
  })

  const hand = zeige()

  await oeffneAnnahme(hand)
  await waehleUmschlag(hand, '/tmp/fremd.msepackage')

  await hand.click(screen.getByRole('button', { name: eigene.leiheAnnehmen }))

  await waitFor(() => {
    expect(genommen).toHaveBeenCalledWith('/tmp/fremd.msepackage')
  })
})

// **Zwei Knöpfe für zwei Rollen.**
it('widerruft eine verliehene und gibt eine geliehene zurück', async () => {
  // **Jede Antwort bringt den ganzen Stand mit** — so wie die Go-Seite:
  // sie gibt nach jeder Änderung die Liste zurück. Eine Probe, die leer
  // antwortete, prüfte eine Naht, die es nicht gibt (§3).
  const beide = {
    leihen: [
      eineLeihe(),
      eineLeihe({ kennung: 'l2', verliehen: false, modus: 'abbruch' }),
    ],
    pfad: '',
    fehler: null,
  }

  const widerrufen = vi.fn().mockResolvedValue(beide)
  const zurueck = vi.fn().mockResolvedValue(beide)

  leiheApi({
    Deckleihen: vi.fn().mockResolvedValue(beide),
    DeckleiheWiderrufen: widerrufen,
    DeckleiheZurueckgeben: zurueck,
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.leiheAnsehen }))

  await hand.click(
    await screen.findByRole('button', { name: eigene.leiheWiderrufen }),
  )

  await waitFor(() => {
    expect(widerrufen).toHaveBeenCalledWith('l1')
  })

  // **Auf den bedienbaren Knopf gewartet, nicht auf den Ruf** (#1809):
  // solange der Widerruf läuft, ruht die Zeile.
  await waitFor(() => {
    expect(
      screen.getByRole('button', { name: eigene.leiheZurueckgeben }),
    ).toBeEnabled()
  })

  await hand.click(
    screen.getByRole('button', { name: eigene.leiheZurueckgeben }),
  )

  // **Auf die Antwort und nicht auf den Ruf**: die Zusicherung darunter
  // greift an den DOM, und der steht erst, wenn die Antwort durch ist (§3).
  await antwortAbwarten()

  expect(zurueck).toHaveBeenCalledWith('l2')
  expect(screen.getByText(eigene.leiheModusAbbruch)).toBeInTheDocument()
})

// **Eine zurückgenommene Leihe hat keinen Knopf mehr** — und sagt es.
it('zeigt eine zurückgenommene Leihe ohne Knopf', async () => {
  leiheApi({
    Deckleihen: vi.fn().mockResolvedValue({
      leihen: [eineLeihe({ widerrufen: true, verbraucht: 'p1' })],
      pfad: '',
      fehler: null,
    }),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.leiheAnsehen }))

  expect(
    await screen.findByText(eigene.leiheZurueckgenommen),
  ).toBeInTheDocument()

  expect(screen.getByText(`${eigene.leiheVerbraucht} p1`)).toBeInTheDocument()

  expect(
    screen.queryByRole('button', { name: eigene.leiheWiderrufen }),
  ).not.toBeInTheDocument()
})

it('meldet einen Fehler', async () => {
  leiheApi({
    Deckleihen: vi.fn().mockResolvedValue({
      leihen: [],
      pfad: '',
      fehler: {
        code: 'konflikt',
        schluessel: 'konflikt',
        meldung: 'ging nicht',
      },
    }),
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.leiheAnsehen }))

  expect(await screen.findByRole('alert')).toBeInTheDocument()
})

// **Ohne Deck sagt die Seite, wo eines entsteht** — statt einer leeren Wahl.
it('erklärt, wenn es noch kein Deck gibt', async () => {
  leiheApi({
    DeckleiheAnbieten: vi.fn(),
    Decks: () => Promise.resolve({ decks: [], fehler: null }),
  })

  zeige()

  expect(await screen.findByText(eigene.leiheKeinDeck)).toBeInTheDocument()
})

// **Kommen die Decks nicht, steht der nächste Versuch da.**
it('bietet Wiederholen, wenn die Decks nicht kommen', async () => {
  const Decks = vi
    .fn()
    .mockRejectedValueOnce(new Error('Bestand gesperrt'))
    .mockResolvedValue({ decks: [einDeck('7', 'Mono Rot')], fehler: null })

  leiheApi({ DeckleiheAnbieten: vi.fn(), Decks })

  const hand = zeige()

  await hand.click(
    await screen.findByRole('button', {
      name: kataloge.de.bausteine.zustaende.wiederholen,
    }),
  )

  expect(
    await screen.findByRole('option', { name: 'Mono Rot' }),
  ).toBeInTheDocument()
})
