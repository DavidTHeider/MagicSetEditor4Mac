// **Die Wahl, wie viel das Programm übernimmt** — PLAY-02.AK-01 und AK-02
// (#1710).
//
// # Wogegen hier geprüft wird
//
//   - **Im Anfängermodus steht kein Kästchen da** — dreizehn Kästchen vor
//     dem ersten Spiel sind eine Prüfung und kein Anfang.
//   - **Der Wechsel zurück räumt die Abwahl mit** (§7, Regel 1).
//   - **Die Bereiche tragen ihre Namen** und nicht ihre Schlüssel.
//   - **Im freien Modus stehen die Gruppen eingeklappt da** — UXR-13
//     (#1921): dreizehn Kästchen neben den Feldern der Sitze waren die
//     Formularwand aus R3 (`docs/UIREVIEW.md`). Wer wählt, klappt die Gruppe
//     auf, in der er wählt.
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Anfaengermodus, Freiermodus, Regelwahlfeld } from './Regelwahlfeld'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.spielen

// einBaum ist der, den der Kern schickt — zwei Äste, vier Bereiche.
const einBaum = [
  { kennung: 'ablauf', bereiche: ['zonen', 'prioritaet'] },
  { kennung: 'mechanik', bereiche: ['marken', 'mana'] },
]

// zeige zeichnet das Feld und gibt die Hand samt Spionen.
function zeige(modus: string, abgewaehlt: string[] = []) {
  const setzeModus = vi.fn()
  const setzeAbgewaehlt = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Regelwahlfeld
        modus={modus}
        setzeModus={setzeModus}
        abgewaehlt={abgewaehlt}
        setzeAbgewaehlt={setzeAbgewaehlt}
        baum={einBaum}
      />
    </Sprachraum>,
  )

  return { hand: nutzerereignis.setup(), setzeModus, setzeAbgewaehlt }
}

// **Im Anfängermodus steht kein Kästchen da.**
it('zeigt im Anfängermodus keine Bereiche', () => {
  zeige(Anfaengermodus)

  expect(screen.queryByRole('checkbox')).not.toBeInTheDocument()

  // **Und beide Modi stehen zur Wahl** — „so viel wie möglich" ist einer
  // von zweien und keine Vorgabe ohne Alternative (AK-01).
  expect(screen.getAllByRole('radio')).toHaveLength(2)

  expect(
    screen.getByRole('radio', { name: eigene.modusAnfaenger }),
  ).toBeChecked()
})

// **Im freien Modus steht je Bereich ein Kästchen — mit seinem Namen, und
// eingeklappt in seiner Gruppe.**
it('zeigt im freien Modus die Gruppen eingeklappt, darin jeden Bereich', async () => {
  const { hand } = zeige(Freiermodus)

  expect(screen.queryByRole('checkbox')).not.toBeInTheDocument()

  // **Die Gruppen tragen ihre Überschrift** — auf dem Knopf, der sie
  // aufklappt.
  await hand.click(
    screen.getByRole('button', { name: eigene.regelgruppeAblauf }),
  )

  expect(screen.getAllByRole('checkbox')).toHaveLength(2)
  expect(
    screen.getByRole('checkbox', {
      name: einsetzen(eigene.regelAbwaehlen, {
        bereich: eigene.regelbereichPrioritaet,
      }),
    }),
  ).toBeInTheDocument()

  await hand.click(
    screen.getByRole('button', { name: eigene.regelgruppeMechanik }),
  )

  expect(screen.getAllByRole('checkbox')).toHaveLength(4)
})

// **Ein Klick wählt ab, der nächste wieder zu.**
it('trägt eine Abwahl ein und nimmt sie zurück', async () => {
  const name = einsetzen(eigene.regelAbwaehlen, {
    bereich: eigene.regelbereichMana,
  })

  const { hand, setzeAbgewaehlt } = zeige(Freiermodus)

  await hand.click(
    screen.getByRole('button', { name: eigene.regelgruppeMechanik }),
  )
  await hand.click(screen.getByRole('checkbox', { name }))

  expect(setzeAbgewaehlt).toHaveBeenCalledWith(['mana'])

  const zurueck = zeige(Freiermodus, ['mana', 'marken'])

  await zurueck.hand.click(
    screen.getAllByRole('button', { name: eigene.regelgruppeMechanik })[1] ??
      screen.getByRole('button', { name: eigene.regelgruppeMechanik }),
  )
  await zurueck.hand.click(
    screen.getAllByRole('checkbox', { name })[1] ??
      screen.getByRole('checkbox', { name }),
  )

  expect(zurueck.setzeAbgewaehlt).toHaveBeenCalledWith(['marken'])
})

// **Der Wechsel in den Anfängermodus räumt die Abwahl mit** — §7, Regel 1:
// sie gehört zu einem Stand, den es nicht mehr gibt.
it('räumt die Abwahl beim Wechsel in den Anfängermodus', async () => {
  const { hand, setzeModus, setzeAbgewaehlt } = zeige(Freiermodus, ['mana'])

  await hand.click(screen.getByRole('radio', { name: eigene.modusAnfaenger }))

  expect(setzeModus).toHaveBeenCalledWith(Anfaengermodus)
  expect(setzeAbgewaehlt).toHaveBeenCalledWith([])
})

// **Und der Wechsel in den freien Modus räumt nichts** — dort fängt die
// Wahl erst an.
it('räumt beim Wechsel in den freien Modus nichts', async () => {
  const { hand, setzeModus, setzeAbgewaehlt } = zeige(Anfaengermodus)

  await hand.click(screen.getByRole('radio', { name: eigene.modusFrei }))

  expect(setzeModus).toHaveBeenCalledWith(Freiermodus)
  expect(setzeAbgewaehlt).not.toHaveBeenCalled()
})
