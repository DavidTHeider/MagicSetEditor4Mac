// **Mitspieler an diesen Tisch lassen** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass die Sitze vom Wirt kommen und nicht vom Gast** — und dass der eigene
// nicht dabei ist. Wer ihn weggäbe, ließe sich selbst vom Tisch schieben.
//
// **Und dass die Erklärung dasteht, bevor geklickt wird** (§6): erreichbar
// sein und mitspielen lassen sind zwei Entscheidungen, und die zweite sieht
// aus wie ein Knopf.
import { render, screen, waitFor } from '@testing-library/react'
import { expect, it, vi } from 'vitest'
import nutzerereignis from '@testing-library/user-event'

import { Partiefreigabe } from './Partiefreigabe'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi } from '../test/api'
import { einePartie, einSitz } from './partieprobe'
import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'

const eigene = kataloge.de.seiten.spielen

// zu ist die Antwort für einen Tisch, der nicht offensteht.
const zu = { laeuft: false, partie: '', wirt: '', vergeben: [], fehler: null }

// zeige zeichnet die Freigabe und gibt die Hand.
function zeige() {
  render(
    <Sprachraum sprache="de">
      <Partiefreigabe partie={einePartie()} />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// **Die Erklärung steht da, bevor jemand klickt.**
it('erklärt, was ein offener Tisch bedeutet', () => {
  belegeApi({ PartieFreigeben: vi.fn() })

  zeige()

  expect(
    screen.getByText(kurzform(eigene.freigabeErklaerung)),
  ).toBeInTheDocument()
  expect(screen.getByText(eigene.freigabeZu)).toBeInTheDocument()
})

// **Freigegeben werden die fremden Sitze — der eigene nicht.**
it('gibt jeden Sitz außer dem eigenen frei', async () => {
  const freigegeben = vi.fn().mockResolvedValue({
    laeuft: true,
    partie: 'p1',
    wirt: 'sitz1',
    vergeben: [],
    fehler: null,
  })

  belegeApi({ PartieFreigeben: freigegeben })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.freigabeStarten }))

  await waitFor(() => {
    expect(screen.getByText(eigene.freigabeOffen)).toBeInTheDocument()
  })

  // **`einePartie` hat zwei Sitze, und `fuer` ist der erste.**
  expect(freigegeben).toHaveBeenCalledWith('p1', 'sitz1', ['sitz2'])
})

// **Steht der Tisch offen, schließt derselbe Knopf ihn wieder.**
it('schließt den Tisch mit demselben Knopf', async () => {
  const geschlossen = vi.fn().mockResolvedValue(zu)

  belegeApi({
    PartieFreigeben: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      wirt: 'sitz1',
      vergeben: [],
      fehler: null,
    }),
    PartieFreigabeBeenden: geschlossen,
  })

  const hand = zeige()

  await hand.click(screen.getByRole('button', { name: eigene.freigabeStarten }))

  await hand.click(
    await screen.findByRole('button', { name: eigene.freigabeBeenden }),
  )

  await waitFor(() => {
    expect(geschlossen).toHaveBeenCalled()
  })
})

// **Wer sitzt, steht da — mit seinem Fingerabdruck.**
//
// **Der Abdruck ist der Punkt**: ein Sitzname sagt nicht, **wer** ihn hat, und
// zwei Menschen können nur den Abdruck miteinander vergleichen.
it('zeigt, wer schon sitzt', async () => {
  belegeApi({
    PartieFreigabeStand: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      wirt: 'sitz1',
      vergeben: [{ sitz: 'sitz2', fingerabdruck: 'abcd-ef01-2345-6789' }],
      fehler: null,
    }),
  })

  const hand = zeige()

  expect(screen.getByText(eigene.freigabeNiemand)).toBeInTheDocument()

  await hand.click(
    screen.getByRole('button', { name: eigene.freigabeNachsehen }),
  )

  expect(await screen.findByText('abcd-ef01-2345-6789')).toBeInTheDocument()
  expect(screen.getByText(eigene.freigabeBesetzt)).toBeInTheDocument()
})

// **Ein Fehler steht am Bauteil und nimmt nichts zurück** (§7, Regel 3).
it('meldet einen Fehler, ohne den Stand wegzuwerfen', async () => {
  belegeApi({
    PartieFreigabeStand: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      wirt: 'sitz1',
      vergeben: [],
      fehler: null,
    }),
    PartieFreigabeBeenden: vi.fn().mockResolvedValue({
      laeuft: false,
      partie: '',
      wirt: '',
      vergeben: [],
      fehler: { code: 'validierung', meldung: 'Keine Rolle.', feld: '' },
    }),
  })

  const hand = zeige()

  await hand.click(
    screen.getByRole('button', { name: eigene.freigabeNachsehen }),
  )

  await screen.findByText(eigene.freigabeOffen)

  await hand.click(screen.getByRole('button', { name: eigene.freigabeBeenden }))

  expect(await screen.findByRole('alert')).toHaveTextContent('Keine Rolle.')

  // **Und der offene Tisch steht noch da.**
  expect(screen.getByText(eigene.freigabeOffen)).toBeInTheDocument()
})

// **Und wer das ist** — Werkstück T7 von Phase 5.6
// (#1472).
//
// **Ein bekanntes Gegenüber wird wiedererkannt**, ohne zweite Paarung — und
// ein unbekanntes ist kein Fehler, sondern eine Auskunft.
it('nennt das bekannte Gegenüber auf einem Sitz', async () => {
  belegeApi({
    PartieFreigabeStand: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      wirt: 'sitz1',
      vergeben: [{ sitz: 'sitz2', fingerabdruck: 'abcd-ef01-2345-6789' }],
      fehler: null,
    }),
    Handelspartner: vi.fn().mockResolvedValue({
      sitz: 'sitz2',
      kontakt: {
        kennung: 'bgegenueber',
        fingerabdruck: 'abcd-ef01-2345-6789',
        // **Ohne eigenen Namen steht der selbstgewählte da** — er belegt
        // nichts, aber er ist besser als eine leere Zeile.
        selbstgewaehlterName: 'Turfer',
        eigenerName: '',
        stand: 'bestaetigt',
        rechte: [],
        adresse: '',
        notiz: '',
      },
      bekannt: true,
      fehler: null,
    }),
  })

  const hand = zeige()

  await hand.click(
    screen.getByRole('button', { name: eigene.freigabeNachsehen }),
  )

  expect(await screen.findByText(/Turfer/)).toBeInTheDocument()
})

// **Ein unbekanntes Gegenüber sagt es und klagt nicht.**
it('sagt es, wenn das Gegenüber nicht in den Kontakten steht', async () => {
  belegeApi({
    PartieFreigabeStand: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      wirt: 'sitz1',
      vergeben: [{ sitz: 'sitz2', fingerabdruck: 'abcd-ef01-2345-6789' }],
      fehler: null,
    }),
    Handelspartner: vi.fn().mockResolvedValue({
      sitz: 'sitz2',
      kontakt: {
        kennung: '',
        fingerabdruck: '',
        selbstgewaehlterName: '',
        eigenerName: '',
        stand: '',
        rechte: [],
        adresse: '',
        notiz: '',
      },
      bekannt: false,
      fehler: null,
    }),
  })

  const hand = zeige()

  await hand.click(
    screen.getByRole('button', { name: eigene.freigabeNachsehen }),
  )

  expect(
    await screen.findByText(eigene.freigabePartnerFremd),
  ).toBeInTheDocument()
})

// **Und ein Fehler beim Nachschlagen nimmt dem Sitz nichts.**
it('zeigt den Sitz auch, wenn das Nachschlagen klagt', async () => {
  belegeApi({
    PartieFreigabeStand: vi.fn().mockResolvedValue({
      laeuft: true,
      partie: 'p1',
      wirt: 'sitz1',
      vergeben: [{ sitz: 'sitz2', fingerabdruck: 'abcd-ef01-2345-6789' }],
      fehler: null,
    }),
    Handelspartner: vi.fn().mockResolvedValue({
      sitz: 'sitz2',
      kontakt: {
        kennung: '',
        fingerabdruck: '',
        selbstgewaehlterName: '',
        eigenerName: '',
        stand: '',
        rechte: [],
        adresse: '',
        notiz: '',
      },
      bekannt: false,
      fehler: {
        code: 'nichtGefunden' as const,
        schluessel: 'sitz',
        meldung: 'da sitzt niemand',
      },
    }),
  })

  const hand = zeige()

  await hand.click(
    screen.getByRole('button', { name: eigene.freigabeNachsehen }),
  )

  // **Gewartet wird auf die Wirkung des Nachschlagens und nicht auf den
  // Sitz** (§3): der Fingerabdruck steht schon da, bevor `Handelspartner`
  // geantwortet hat — eine Zusicherung darauf träfe zu, ohne etwas geprüft
  // zu haben, und der Griff daneben ginge ins Leere.
  expect(
    await screen.findByText(eigene.freigabePartnerFremd),
  ).toBeInTheDocument()

  expect(screen.getByText('abcd-ef01-2345-6789')).toBeInTheDocument()
})

// **Wer erwartet wird, steht da** — P9c (#2377): der Sitz wartet auf das
// Deck des Gastes, und der Wirt liest, auf wen.
it('nennt, wer erwartet wird', () => {
  belegeApi({ PartieFreigabeStand: () => Promise.resolve(zu) })

  const partie = einePartie({
    sitze: [
      einSitz({ kennung: 'sitz1', name: 'Anna', deckabdruck: 'abc' }),
      einSitz({ kennung: 'sitz2', name: 'Bo', wartet: true }),
    ],
  })

  render(
    <Sprachraum sprache="de">
      <Partiefreigabe partie={partie} />
    </Sprachraum>,
  )

  expect(
    screen.getByText(eigene.freigabeWartet.replace('{namen}', 'Bo')),
  ).toBeInTheDocument()
})
