// **Eingefügt, in fremder Schrift, mehrdeutig** — UX-06.AK-04
// (#1678):
// Grenzwerte, Unicode, Leerwerte, Einfügen, Dezimalzeichen, ungültige
// Währungscodes und mehrdeutige Eingaben werden getestet.
//
// **Die meisten stehen bei ihren Bauteilen** — Grenzen, leere Felder und
// Kommazahlen in `zahlenfeld.test.tsx`, `€`, `ÄUR` und der Zahlencode `978`
// in `waehrungsgestalt.test.ts`, das Trennzeichen je Sprache in
// `geld.test.ts`. Hier stehen die drei, die dort fehlten: **das Einfügen**,
// **Zeichen, die wie lateinische Buchstaben aussehen**, und **eine Zahl, die
// zweierlei heißen kann**.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { expect, it, vi } from 'vitest'

import { Waehrungsfeld } from './Waehrungsfeld'
import { Zahlenfeld } from './Zahlenfeld'
import { einsetzen } from '../sprache/einsetzen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

const waehrung = kataloge.de.bausteine.waehrung
const allgemein = kataloge.de.allgemein

function Waehrung() {
  const [wert, setzeWert] = useState('')

  return (
    <Waehrungsfeld
      beschriftung="Währung"
      wert={wert}
      setzeWert={setzeWert}
      kleinErlaubt
    />
  )
}

async function fuegeEin(text: string) {
  render(
    <Sprachraum sprache="de">
      <Waehrung />
    </Sprachraum>,
  )

  const feld = screen.getByLabelText('Währung')
  await userEvent.click(feld)
  await userEvent.paste(text)
  await userEvent.tab()

  return feld
}

// **Wie Go es nimmt**: `domain.NormalisiereWaehrung` glättet und hebt an —
// das Fenster urteilt nach derselben Regel.
it('nimmt einen eingefügten Code samt Leerraum, wie Go ihn nimmt', async () => {
  const feld = await fuegeEin(' eur ')

  expect(screen.queryByRole('alert')).toBeNull()

  // **Der Rohtext bleibt stehen** — geglättet wird beim Urteil, nicht im
  // Feld (AK-03).
  expect(feld).toHaveValue(' eur ')
})

// **Vollbreite Buchstaben sehen aus wie lateinische und sind keine** — Go
// zählt nur `A` bis `Z` (`istWaehrungscode`).
it('weist Buchstaben in voller Breite ab und lässt sie stehen', async () => {
  const feld = await fuegeEin('ＥＵＲ')

  expect(screen.getByRole('alert')).toHaveTextContent(
    einsetzen(waehrung.keineGestalt, { wert: 'ＥＵＲ' }),
  )
  expect(feld).toHaveValue('ＥＵＲ')
})

// **„1.500" heißt auf Deutsch eintausendfünfhundert und für das Zahlenfeld
// anderthalb.** Keines von beiden wird still genommen: an einem Feld für
// ganze Zahlen steht der Befund, und der Wert geht nicht weiter.
it('nimmt „1.500" an einem Feld für ganze Zahlen nicht still an', async () => {
  const geaendert = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Zahlenfeld
        beschriftung="Menge"
        wert={1}
        kleinstens={1}
        ganzzahlig
        geaendert={geaendert}
      />
    </Sprachraum>,
  )

  const feld = screen.getByLabelText('Menge')
  await userEvent.clear(feld)
  await userEvent.click(feld)
  await userEvent.paste('1.500')

  expect(feld).toHaveAttribute('aria-invalid', 'true')
  expect(screen.getByRole('alert')).toHaveTextContent(
    einsetzen(allgemein.ganzeZahlAb, { kleinstens: '1' }),
  )
  expect(geaendert).not.toHaveBeenCalledWith(1.5)
  expect(geaendert).not.toHaveBeenCalledWith(1500)
})
