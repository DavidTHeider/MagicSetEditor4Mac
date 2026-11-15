// **Die Verwaltung einer Partie als Ganzem** — PLAY-03.AK-05
// (#1711).
//
// # Wogegen hier geprüft wird
//
//   - **Jeder Eintrag ruft seinen Weg mit seiner Partie** — ein Aufruf
//     allein wäre auch grün, wenn alle drei dasselbe täten (§10). Seit UXR-13
//     (#1921) stehen die Wege in einem Menü der Zeile und nicht mehr als drei
//     Knöpfe: der Name des Menüs nennt die Partie, die Einträge heißen, was
//     sie tun.
//   - **Die Meldung steht an der Zeile, die sie ausgelöst hat** — und an
//     keiner anderen (§6).
//   - **Ein laufender Vorgang hält die Einträge an** (§7, Regel 4).
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Partienverwaltung } from './Partienverwaltung'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { stilleVerwaltung } from '../test/verwaltung'

import type { Verwaltungsstand } from '../haken/usePartieverwaltung'

const eigene = kataloge.de.seiten.spielen

// zeige zeichnet das Menü einer Zeile und gibt die Hand.
function zeige(
  lauf: Verwaltungsstand,
  kennung = 'p1',
  marken: () => void = () => undefined,
) {
  render(
    <Sprachraum sprache="de">
      <Partienverwaltung
        kennung={kennung}
        name="Anna, Bert"
        lauf={lauf}
        marken={marken}
      />
    </Sprachraum>,
  )

  return nutzerereignis.setup()
}

// oeffne schlägt das Menü der Partie auf — **über den Namen, den ein
// Vorleseprogramm hört**: er trägt die Partie (WCAG 2.4.6).
async function oeffne(hand: ReturnType<typeof nutzerereignis.setup>) {
  await hand.click(
    screen.getByRole('button', {
      name: einsetzen(eigene.partienHandlungen, { name: 'Anna, Bert' }),
    }),
  )
}

// **Jeder Eintrag ruft seinen Weg — mit dieser Partie.**
it('ruft jeden Weg mit der Kennung seiner Zeile', async () => {
  const loeschen = vi.fn()
  const kopieren = vi.fn()
  const ausfuehren = vi.fn()
  const marken = vi.fn()

  const hand = zeige(
    stilleVerwaltung({ loeschen, kopieren, ausfuehren }),
    'p7',
    marken,
  )

  for (const [eintrag, gerufen] of [
    [eigene.partienDuplizieren, kopieren],
    [eigene.partienAusfuhr, ausfuehren],
    [eigene.partienLoeschen, loeschen],
  ] as const) {
    await oeffne(hand)
    await hand.click(screen.getByRole('menuitem', { name: eintrag }))

    expect(gerufen).toHaveBeenCalledWith('p7')
  }

  // **Und die Marken** — sie stehen seit UXR-13 im selben Menü und nicht
  // mehr als eigener Knopf an der Zeile.
  await oeffne(hand)
  await hand.click(
    screen.getByRole('menuitem', {
      name: kataloge.de.bausteine.objektmarken.ueberschrift,
    }),
  )

  expect(marken).toHaveBeenCalled()
})

// **Der Name des Menüs nennt die Partie** — WCAG 2.4.6: „Löschen" allein
// sagte bei neunzehn Zeilen nicht, welche. Ein Vorleseprogramm hört den
// Namen des Menüs, bevor es seine Einträge hört.
it('nennt die Partie im Namen des Menüs', async () => {
  const hand = zeige(stilleVerwaltung())

  await oeffne(hand)

  expect(
    screen.getByRole('menu', {
      name: einsetzen(eigene.partienHandlungen, { name: 'Anna, Bert' }),
    }),
  ).toBeInTheDocument()
})

// **Ein laufender Vorgang hält die Einträge an** — §7, Regel 4. Die Marken
// gehören nicht dazu: sie schreiben über einen eigenen Vorgang.
it('hält die Einträge an, solange etwas läuft', async () => {
  const hand = zeige(stilleVerwaltung({ laeuft: () => true }))

  await oeffne(hand)

  for (const eintrag of [
    eigene.partienDuplizieren,
    eigene.partienAusfuhr,
    eigene.partienLoeschen,
  ]) {
    expect(screen.getByRole('menuitem', { name: eintrag })).toHaveAttribute(
      'aria-disabled',
      'true',
    )
  }
})

// **Die Meldung steht an ihrer Zeile — und an keiner anderen.**
it('zeigt die Meldung nur an der Zeile, die sie ausgelöst hat', () => {
  const { unmount } = render(
    <Sprachraum sprache="de">
      <Partienverwaltung
        kennung="p1"
        name="Anna"
        lauf={stilleVerwaltung({
          meldung: { kennung: 'p2', satz: 'Gelöscht.', istFehler: false },
        })}
        marken={() => undefined}
      />
    </Sprachraum>,
  )

  expect(screen.queryByText('Gelöscht.')).not.toBeInTheDocument()

  unmount()

  zeige(
    stilleVerwaltung({
      meldung: { kennung: 'p1', satz: 'Gelöscht.', istFehler: false },
    }),
  )

  expect(screen.getByRole('status')).toHaveTextContent('Gelöscht.')
})

// **Ein Fehler ist eine Warnung und keine Statuszeile** — §6: er wird
// angesagt, und zwar als das, was er ist.
it('zeigt einen Fehler als Warnung', () => {
  zeige(
    stilleVerwaltung({
      meldung: { kennung: 'p1', satz: 'Geht nicht.', istFehler: true },
    }),
  )

  expect(screen.getByRole('alert')).toHaveTextContent('Geht nicht.')
  expect(screen.queryByRole('status')).not.toBeInTheDocument()
})
