// **Ein Mensch gibt einen bezahlten Auftrag frei** —
// #1539.
//
// # Der Befund
//
// **Am laufenden Fenster ging kein einziger hinaus.** Die Kostenwache
// verlangt die ausdrückliche Freigabe eines Menschen; ausgestellt hat sie
// keine Bindung, und die Oberfläche bot nichts dafür an. Jeder Auftrag an
// ein Cloudmodell endete mit einem Satz, auf den niemand reagieren konnte.
//
// # Was hier geprüft wird
//
// **Der Weg vom Knopf bis zum Auftrag** — dass vorher nichts hinausgeht, dass
// die Kennung am Auftrag ankommt und dass ein Konto ohne Budget den nächsten
// Schritt nennt, statt den Ruf scheitern zu lassen.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { KIKostenfreigabe } from './KIKostenfreigabe'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { einsetzen } from '../sprache/einsetzen'

import type { HaushaltsstandAntwort } from '../apikosten'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.einstellungen

// einStand baut die Auskunft des Kostenjournals.
function einStand(grenze: number, verfuegbar: number): HaushaltsstandAntwort {
  const betrag = (wert: number) => ({ wert, waehrung: 'EUR' })

  return {
    konto: 'anthropic',
    zeitraum: 'tag',
    grenze: betrag(grenze),
    reserviert: betrag(grenze - verfuegbar),
    verbraucht: betrag(0),
    unklar: betrag(0),
    verfuegbar: betrag(verfuegbar),
    fehler: null,
  }
}

// zeige rendert den Kasten und gibt die beiden Spione.
function zeige(anbieter = 'anthropic') {
  const freigegeben = vi.fn()
  const verzichtet = vi.fn()

  render(
    <Sprachraum sprache="de">
      <KIKostenfreigabe
        anbieter={anbieter}
        freigegeben={freigegeben}
        verzichtet={verzichtet}
      />
    </Sprachraum>,
  )

  return { freigegeben, verzichtet }
}

// gibFrei tippt den Höchstbetrag und drückt.
//
// **Gewartet wird auf das Feld und nicht auf die Überschrift** (§3). Die
// steht sofort da — der Kasten zeichnet sich, bevor das Kostenjournal
// geantwortet hat —, und ein `findByText` darauf ist eine Zusicherung, die
// schon vor der Antwort zutrifft. Das Feld gibt es erst, wenn ein Budget
// bekannt ist; der erste Anlauf lief genau in diese Falle.
async function gibFrei(hand: ReturnType<typeof userEvent.setup>, wert: string) {
  await hand.type(await screen.findByLabelText(eigene.kiFreigabeBetrag), wert)
  await hand.click(
    screen.getByRole('button', { name: eigene.kiFreigabeErteilen }),
  )
}

// **Eine erteilte Freigabe reicht ihre Kennung weiter.**
//
// **Das ist die Naht, die gefehlt hat**: erteilen und weitergeben, in einem
// Lauf. Eine Freigabe, die niemand bekommt, ist keine.
it('erteilt die Freigabe und reicht ihre Kennung weiter', async () => {
  const erteilen = vi.fn(() =>
    Promise.resolve({
      kennung: 'n-1',
      laeuftab: '2026-09-12T12:05:00Z',
      fehler: null,
    }),
  )

  belegeApi({
    Kostenstand: () => Promise.resolve(einStand(50_000, 30_000)),
    KIFreigabeErteilen: erteilen,
  })

  const { freigegeben } = zeige()
  const hand = userEvent.setup()

  await gibFrei(hand, '1200')

  await vi.waitFor(() => {
    expect(freigegeben).toHaveBeenCalledWith('n-1')
  })

  expect(erteilen).toHaveBeenCalledWith('anthropic', 1200, 'EUR')
})

// **»1.000« ist kein Höchstbetrag** — für `Number` wäre es ein Cent
// (Nachlauf über Phase 6.3, N5).
it('erteilt nichts, was keine Folge von Ziffern ist, und sagt es am Feld', async () => {
  const erteilen = vi.fn()

  belegeApi({
    Kostenstand: () => Promise.resolve(einStand(50_000, 30_000)),
    KIFreigabeErteilen: erteilen,
  })

  zeige()
  const hand = userEvent.setup()
  const feld = await screen.findByLabelText(eigene.kiFreigabeBetrag)

  await hand.type(feld, '1.000')

  expect(
    screen.getByRole('button', { name: eigene.kiFreigabeErteilen }),
  ).toBeDisabled()
  expect(feld).toHaveAccessibleDescription(kataloge.de.allgemein.nurZiffern)
  expect(erteilen).not.toHaveBeenCalled()
})

// **Der verfügbare Rest steht dabei — in Geld und nicht in Cent.**
it('nennt, was heute noch verfügbar ist', async () => {
  belegeApi({ Kostenstand: () => Promise.resolve(einStand(50_000, 30_000)) })
  zeige()

  expect(
    await screen.findByText(
      einsetzen(eigene.kiFreigabeVerfuegbar, { betrag: '300,00 €' }),
    ),
  ).toBeInTheDocument()
})

// **Ohne Budget sagt der Kasten, wo die Grenze gesetzt wird.**
//
// **Und bietet keine Freigabe an** (§6): sie ginge hinaus und scheiterte am
// Journal — mit einer Meldung, die von einem Budget spricht, während jemand
// auf einen Knopf gedrückt hat, der Freigabe hieß.
it('nennt ohne Budget den nächsten Schritt statt einen Knopf', async () => {
  belegeApi({ Kostenstand: () => Promise.resolve(einStand(0, 0)) })
  zeige()

  expect(
    await screen.findByText(
      einsetzen(eigene.kiFreigabeOhneBudget, { anbieter: 'Anthropic' }),
    ),
  ).toBeInTheDocument()

  expect(
    screen.queryByRole('button', { name: eigene.kiFreigabeErteilen }),
  ).not.toBeInTheDocument()
})

// **Der Anbieter steht mit seinem Namen da.**
//
// **Im Fenster stand „rechnet bei openai"** — die Kennung, mit der das
// Programm rechnet. Gefunden in der Sichtprüfung am 2026-09-13.
it('nennt den Anbieter mit seinem Namen', async () => {
  belegeApi({ Kostenstand: () => Promise.resolve(einStand(50_000, 30_000)) })
  zeige()

  expect(
    await screen.findByText(
      einsetzen(eigene.kiFreigabeSatz, { anbieter: 'Anthropic' }),
    ),
  ).toBeInTheDocument()
})

// **Wer nicht freigibt, kommt ohne Ruf davon.**
it('meldet den Verzicht und ruft nichts', async () => {
  const erteilen = vi.fn()

  belegeApi({
    Kostenstand: () => Promise.resolve(einStand(50_000, 30_000)),
    KIFreigabeErteilen: erteilen,
  })

  const { verzichtet } = zeige()
  const hand = userEvent.setup()

  await hand.click(
    await screen.findByRole('button', {
      name: eigene.kiFreigabeVerzichten,
    }),
  )

  expect(verzichtet).toHaveBeenCalled()
  expect(erteilen).not.toHaveBeenCalled()
})

// **Und ein Fehler beim Ausstellen steht da, statt still zu verschwinden.**
//
// **Mit dem Satz aus dem Katalog** (§9): was Go schickt, ist ein Schlüssel;
// was jemand liest, steht in `sprache/`. Ein Lauf, der die deutsche Meldung
// aus der Antwort erwartete, prüfte die Vorrichtung statt den Weg.
it('zeigt, woran das Ausstellen scheiterte', async () => {
  belegeApi({
    Kostenstand: () => Promise.resolve(einStand(50_000, 30_000)),
    KIFreigabeErteilen: () =>
      Promise.resolve({
        kennung: '',
        laeuftab: '',
        fehler: {
          code: 'erlaubnis',
          schluessel: 'erlaubnis',
          meldung: 'einen Nachweis stellt aus, wer mindestens mitarbeitet',
          feld: '',
          werte: {},
          vorgang: 'v-1',
        },
      }),
  })

  const { freigegeben } = zeige()

  await gibFrei(userEvent.setup(), '1200')

  expect(await screen.findByRole('alert')).toHaveTextContent(
    kataloge.de.fehler.erlaubnis,
  )

  expect(freigegeben).not.toHaveBeenCalled()
})

// **Und ein Fehler beim Lesen des Budgets steht ebenso da.**
//
// **Zwei Fehler, eine Stelle**: das Ausstellen und das Lesen können beide
// scheitern, und wer nur den ersten zeigte, ließe jemanden vor einem Kasten
// ohne Knopf und ohne Grund stehen (§6).
it('zeigt, woran das Lesen des Budgets scheiterte', async () => {
  belegeApi({
    Kostenstand: () => Promise.reject(new Error('kein Bestand offen')),
  })

  zeige()

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'kein Bestand offen',
  )

  expect(
    screen.queryByRole('button', { name: eigene.kiFreigabeErteilen }),
  ).not.toBeInTheDocument()
})

// **Ein Anbieter ohne Eigennamen steht mit seiner Kennung da.**
//
// **Die letzte Reihe und nicht der Normalfall** (§6): lieber `neuling` lesen
// als einen Satz mit einer Lücke. Dass jeder eingebaute Anbieter einen Namen
// hat, hält `internal/werkzeuge/einstellungstexte_test.go`.
it('fällt ohne Eigennamen auf die Kennung zurück', async () => {
  belegeApi({ Kostenstand: () => Promise.resolve(einStand(0, 0)) })
  zeige('neuling')

  expect(
    await screen.findByText(
      einsetzen(eigene.kiFreigabeSatz, { anbieter: 'neuling' }),
    ),
  ).toBeInTheDocument()
})
