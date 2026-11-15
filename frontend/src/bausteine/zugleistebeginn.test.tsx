// **Mitten im Beginn bietet die Zugleiste nichts an, was der Kern abweist** —
// Befund 2 des Zweitblicks über Phase 6.9, CR 103.8 (Fassung 2026-09-25).
//
// Der Startspieler hat die Priorität des ersten Zugs schon, während ein
// anderer Sitz noch seine Starthand hält. Der Kern nimmt dann nur den Beginn
// an (`service.beginnHaelt`); drei Knöpfe, die er abwiese, wären eine
// Einladung, die man zurücknimmt — und grau, ohne ein Wort, ließen sie raten
// (§6).
import { render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Zugleiste } from './Zugleiste'
import { einZug, einePartie } from '../seiten/partieprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.spielen

const amZug = einePartie({
  fuer: 'sitz1',
  zug: einZug({
    nummer: 1,
    aktiv: 'sitz1',
    phase: 'anfang',
    schritt: 'versorgung',
    prioritaet: 'sitz1',
  }),
})

function zeige(erlaubt: readonly string[]) {
  render(
    <Sprachraum sprache="de">
      <Zugleiste
        partie={amZug}
        erlaubt={erlaubt}
        laeuft={false}
        passen={() => undefined}
        abkuerzen={() => undefined}
        zugende={() => undefined}
      />
    </Sprachraum>,
  )
}

it('hält die drei Knöpfe an und sagt, worauf der Zug wartet', () => {
  zeige(['spieler.starthand', 'zug.aufgeben'])

  for (const name of [
    eigene.zugPassen,
    eigene.zugAbkuerzen,
    eigene.zugBeenden,
  ]) {
    expect(screen.getByRole('button', { name })).toBeDisabled()
  }

  expect(screen.getByText(eigene.zugWartetAufBeginn)).toBeInTheDocument()
})

it('gibt sie frei, sobald der Kern sie annimmt', () => {
  zeige(['zug.passen', 'zug.abkuerzen', 'zug.beenden'])

  for (const name of [
    eigene.zugPassen,
    eigene.zugAbkuerzen,
    eigene.zugBeenden,
  ]) {
    expect(screen.getByRole('button', { name })).toBeEnabled()
  }

  expect(screen.getByText(eigene.zugDuBistDran)).toBeInTheDocument()
})

// **Eine beendete Partie wartet auf nichts mehr** — P4
// (#2372):
// an ihr ist nichts erlaubt, und „dran, aber nicht passen dürfen“ las sich
// als „der Beginn läuft“. Das Ergebnis steht oben (`Partieausgang`); die
// Leiste zeichnet dann nichts.
it('zeichnet an einer beendeten Partie keine Aufforderung', () => {
  render(
    <Sprachraum sprache="de">
      <Zugleiste
        partie={{ ...amZug, stand: 'beendet' }}
        erlaubt={[]}
        laeuft={false}
        passen={() => undefined}
        abkuerzen={() => undefined}
        zugende={() => undefined}
      />
    </Sprachraum>,
  )

  expect(screen.queryByText(eigene.zugWartetAufBeginn)).toBeNull()
  expect(screen.queryByRole('region', { name: eigene.zugLeiste })).toBeNull()
})
