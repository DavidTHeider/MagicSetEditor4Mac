// **Die gewählte Abwahl geht mit hinaus** — P1a, Befund C-01 aus dem Audit
// zur Produktionsreife 0.9.0
// (#2335).
//
// # Wogegen hier geprüft wird
//
// **Gegen eine Vorschau, die etwas anderes sagte als die Partie.** Bis zum
// 2026-10-13 zeigte die Regeltafel nach „Zugstruktur selbst übernehmen“
// den Bereich als manuell, und der Anfang schickte trotzdem
// `abgewaehlt: []` — die eingefrorene Partie trug die Wahl nicht, und ihre
// Matrix nannte die Zugstruktur automatisch. Nachgestellt am nativen
// Fenster im Zweitblick zu 0.9.0 (`docs/audits/0.9.0/ZWEITBLICK-CODEX.md`,
// Abzug 04).
//
// **Der Lauf beginnt an der rohen Eingabe** (§3): Modus umstellen, Gruppe
// aufklappen, Kästchen anhaken — und nicht an einem schon umgerechneten
// Zustand.
import { screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { einsetzen } from '../sprache/einsetzen'
import { eineAntwort, element } from './partieprobe'
import { eineFreigabe, mitDecks, zeige } from './spielenprobe'
import { kataloge } from '../sprache/kataloge'
import { antwortAbwarten, raeumeApi } from '../test/api'
import { regelbereichwort, regelgruppenwort } from './regelwoerter'

import type { Partieentwurf } from '../apipartie'

const eigene = kataloge.de.seiten.spielen

afterEach(() => {
  raeumeApi()
})

// mitAblauf belegt die Fassade so, dass die Gruppe *Ablauf* die
// Zugstruktur trägt — wie das Grundprofil (`domain.Regelbaum`).
function mitAblauf(angelegt: (entwurf: Partieentwurf) => unknown) {
  const freigabe = eineFreigabe()

  mitDecks({
    PartieAnlegen: angelegt,
    Regelfreigabe: () =>
      Promise.resolve({
        ...freigabe,
        freigabe: {
          ...freigabe.freigabe,
          baum: [{ kennung: 'ablauf', bereiche: ['zonen', 'zugstruktur'] }],
        },
      }),
  })
}

// einAnlegen ist ein Spion auf `PartieAnlegen`, der den Entwurf festhält.
function einAnlegen() {
  return vi.fn<(entwurf: Partieentwurf) => Promise<unknown>>(() =>
    Promise.resolve(eineAntwort()),
  )
}

// sitzeBelegen gibt beiden Sitzen Namen und Deck — erst dann ist der
// Anfang vollständig.
async function sitzeBelegen(hand: ReturnType<typeof zeige>) {
  await hand.type(await screen.findByLabelText(`${eigene.sitzName} 1`), 'Ada')
  await hand.type(screen.getByLabelText(`${eigene.sitzName} 2`), 'Bo')

  const decks = screen.getAllByLabelText(eigene.sitzDeck)

  await hand.selectOptions(element(decks[0]), '1')
  await hand.selectOptions(element(decks[1]), '2')
}

// zugstrukturAbwaehlen stellt auf den freien Modus und hakt die
// Zugstruktur an — der Weg, den ein Mensch an der Tafel geht.
async function zugstrukturAbwaehlen(hand: ReturnType<typeof zeige>) {
  await hand.click(await screen.findByLabelText(eigene.modusFrei))

  await hand.click(
    await screen.findByRole('button', {
      name: regelgruppenwort(eigene, 'ablauf'),
    }),
  )

  await hand.click(
    screen.getByLabelText(
      einsetzen(eigene.regelAbwaehlen, {
        bereich: regelbereichwort(eigene, 'zugstruktur'),
      }),
    ),
  )
}

it('schickt die an der Tafel gewählte Abwahl mit dem Anfang', async () => {
  const angelegt = einAnlegen()

  mitAblauf(angelegt)

  const hand = zeige()

  await sitzeBelegen(hand)
  await zugstrukturAbwaehlen(hand)

  await hand.click(screen.getByRole('button', { name: eigene.anfangen }))

  await waitFor(() => {
    expect(angelegt.mock.lastCall?.[0].regelwerk.abgewaehlt).toEqual([
      'zugstruktur',
    ])
  })
})

// **Die Gegenprobe**: wer zurück in den Anfängermodus stellt, räumt die
// Abwahl mit (§7, Regel 1) — und der Anfang schickt dann keine.
it('schickt keine Abwahl, wenn der Anfängermodus sie geräumt hat', async () => {
  const angelegt = einAnlegen()

  mitAblauf(angelegt)

  const hand = zeige()

  await sitzeBelegen(hand)
  await zugstrukturAbwaehlen(hand)

  await hand.click(screen.getByLabelText(eigene.modusAnfaenger))
  await hand.click(screen.getByRole('button', { name: eigene.anfangen }))

  await waitFor(() => {
    expect(angelegt.mock.lastCall?.[0].regelwerk.abgewaehlt).toEqual([])
  })
})

// **Das Anhaken lässt die Tafel stehen** — gefunden an der Hüllenprobe
// `p1a` am 2026-10-13: die neue Regelfreigabe nach der Wahl nahm die alte
// Antwort weg, die Regeltafel verschwand für einen Augenblick, und danach
// war die Gruppe wieder zu und der Fokus fort. Mit der Tastatur hieß das:
// nach jedem Kästchen die Gruppe neu suchen.
it('lässt nach dem Anhaken die Gruppe offen und den Fokus am Kästchen', async () => {
  mitAblauf(einAnlegen())

  const hand = zeige()

  await sitzeBelegen(hand)
  await zugstrukturAbwaehlen(hand)

  await antwortAbwarten()

  const kaestchen = screen.getByLabelText(
    einsetzen(eigene.regelAbwaehlen, {
      bereich: regelbereichwort(eigene, 'zugstruktur'),
    }),
  )

  expect(kaestchen).toBeChecked()
  expect(kaestchen).toHaveFocus()
  expect(
    screen.getByRole('button', { name: regelgruppenwort(eigene, 'ablauf') }),
  ).toHaveAttribute('aria-expanded', 'true')
})
