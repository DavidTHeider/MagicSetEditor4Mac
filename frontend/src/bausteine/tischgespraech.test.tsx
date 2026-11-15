// **Das Gespräch am Tisch** — Werkstück G5 von Phase 5.5
// (#1462).
//
// # Wogegen hier geprüft wird
//
// **Dass beides dasteht und beides bedienbar ist.** Ein Chat, den man nur mit
// der Maus abschicken kann, ist keiner (§6) — und eine Präsenz, die nur als
// gezeichneter Zeiger existierte, wäre für ein Vorleseprogramm nichts.
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Satzlaenge, Tischgespraech } from './Tischgespraech'
import { verstoesse } from '../test/axe'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type { Nebenlauf } from '../haken/usePartienebenbei'

const eigene = kataloge.de.seiten.spielen

// einLauf baut einen Nebenlauf mit den genannten Abweichungen.
function einLauf(abweichend: Partial<Nebenlauf> = {}): Nebenlauf {
  return {
    gespraech: [],
    zeiger: [],
    sagen: vi.fn(),
    zeigen: vi.fn(),
    fehler: undefined,
    ...abweichend,
  }
}

// zeige zeichnet die Leiste und gibt die Hand.
function zeige(lauf: Nebenlauf) {
  render(
    <Sprachraum sprache="de">
      <Tischgespraech lauf={lauf} />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// **Ein leerer Tisch erklärt, was fehlt, statt nichts zu zeigen** (§6).
it('sagt, dass noch niemand geredet und niemand gezeigt hat', () => {
  zeige(einLauf())

  expect(screen.getByText(eigene.gespraechLeer)).toBeInTheDocument()
  expect(screen.getByText(eigene.gespraechNiemandZeigt)).toBeInTheDocument()
})

// **Gesagtes steht mit seinem Sprecher da.**
it('zeigt jeden Satz mit seinem Sprecher', () => {
  zeige(
    einLauf({
      gespraech: [
        { nummer: 1, von: 'anna', text: 'gut gezogen' },
        { nummer: 2, von: 'bert', text: 'danke' },
      ],
    }),
  )

  expect(screen.getByText('gut gezogen')).toBeInTheDocument()
  expect(screen.getByText('bert')).toBeInTheDocument()
  // **Das Gespräch rollt, also erreicht es die Tastatur** (G10).
  expect(
    screen.getByRole('region', { name: eigene.gespraechVerlauf }),
  ).toHaveAttribute('tabindex', '0')
  expect(screen.queryByText(eigene.gespraechLeer)).not.toBeInTheDocument()
})

// **Die Liste bleibt eine Liste** — eine Region auf der `<ul>` nahm ihr die
// Rolle, und jedes `<li>` stand ohne gültige Liste da (Nachlauf über Phase
// 6.3, N7).
it('hält mit Zeilen axe stand', async () => {
  zeige(
    einLauf({
      gespraech: [
        { nummer: 1, von: 'anna', text: 'gut gezogen' },
        { nummer: 2, von: 'bert', text: 'danke' },
      ],
    }),
  )

  expect(await verstoesse(document.body)).toEqual([])
})

// **Ein Zeiger steht als Satz da — mit und ohne Objekt.**
it('sagt, wer gerade was tut', () => {
  zeige(
    einLauf({
      zeiger: [
        { von: 'anna', art: 'greift', objekt: 'k7', x: 0, y: 0, seit: 0 },
        { von: 'bert', art: 'zeiger', objekt: '', x: 0, y: 0, seit: 0 },
      ],
    }),
  )

  expect(
    screen.getByText(`anna: ${eigene.zeigerGreift} k7`),
  ).toBeInTheDocument()
  expect(screen.getByText(`bert: ${eigene.zeigerZeiger}`)).toBeInTheDocument()
})

// **Eine Art, die das Fenster nicht kennt, zeigt ihre Kennung** (§8).
//
// **Besser als eine leere Zeile**: der Kern darf eine fünfte Präsenzart
// bekommen, und bis das Fenster sie kennt, sagt die Kennung wenigstens, dass
// etwas geschieht.
it('zeigt eine unbekannte Zeigerart als Kennung', () => {
  zeige(
    einLauf({
      zeiger: [{ von: 'anna', art: 'winken', objekt: '', x: 0, y: 0, seit: 0 }],
    }),
  )

  expect(screen.getByText('anna: winken')).toBeInTheDocument()
})

// **Die Eingabetaste schickt, und das Feld ist danach leer.**
it('schickt einen Satz mit der Eingabetaste', async () => {
  const sagen = vi.fn()
  const hand = zeige(einLauf({ sagen }))

  const feld = screen.getByLabelText(eigene.gespraechEingabe)

  await hand.type(feld, 'gut gezogen{Enter}')

  expect(sagen).toHaveBeenCalledWith('gut gezogen')
  expect(feld).toHaveValue('')
})

// **Der Knopf tut dasselbe** — §6 verlangt beide Wege.
it('schickt einen Satz über den Knopf', async () => {
  const sagen = vi.fn()
  const hand = zeige(einLauf({ sagen }))

  await hand.type(screen.getByLabelText(eigene.gespraechEingabe), '  hallo  ')
  await hand.click(screen.getByRole('button', { name: eigene.gespraechSenden }))

  // **Ohne die Leerzeichen** — was hinausgeht, ist der Satz und nicht das
  // Tippen drumherum.
  expect(sagen).toHaveBeenCalledWith('hallo')
})

// **Ein leerer Satz geht nicht hinaus, auf keinem der beiden Wege.**
it('schickt nichts, solange nichts dasteht', async () => {
  const sagen = vi.fn()
  const hand = zeige(einLauf({ sagen }))

  expect(
    screen.getByRole('button', { name: eigene.gespraechSenden }),
  ).toBeDisabled()

  const feld = screen.getByLabelText(eigene.gespraechEingabe)

  await hand.type(feld, '   ')
  await hand.type(feld, '{Enter}')

  expect(sagen).not.toHaveBeenCalled()

  // **Der Beleg, dass jemand zuhört** (§3): derselbe Weg mit Inhalt wirkt.
  await hand.type(feld, 'doch{Enter}')

  expect(sagen).toHaveBeenCalledWith('doch')
})

// **Eine andere Taste schickt nichts.**
it('lässt eine andere Taste durch, ohne zu schicken', async () => {
  const sagen = vi.fn()
  const hand = zeige(einLauf({ sagen }))

  const feld = screen.getByLabelText(eigene.gespraechEingabe)

  await hand.type(feld, 'hallo')
  await hand.type(feld, '{Escape}')

  expect(sagen).not.toHaveBeenCalled()
  expect(feld).toHaveValue('hallo')
})

// **Das Feld hält die Grenze ein, die auch Go hält.**
it('lässt nicht mehr Zeichen zu, als ein Satz fasst', () => {
  zeige(einLauf())

  expect(screen.getByLabelText(eigene.gespraechEingabe)).toHaveAttribute(
    'maxlength',
    String(Satzlaenge),
  )
})

// **Ein Fehler steht am Gespräch und nicht irgendwo.**
it('zeigt einen Fehler als Meldung', () => {
  zeige(einLauf({ fehler: 'kein Tisch' }))

  expect(screen.getByRole('alert')).toHaveTextContent('kein Tisch')
})
