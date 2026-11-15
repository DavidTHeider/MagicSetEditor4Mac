// **Eine abgewiesene Handlung wird am Tisch erklärt** — PLAY-04.AK-04
// (#1712).
//
// # Der Befund dahinter
//
// **Die Meldung stand am Kopf der Seite.** `Spielen.tsx` zeichnet sie über
// allem, was darunter kommt — und darunter kommt der Tisch, der den ganzen
// Bildschirm füllt. Wer eine Karte ablegt, die der Kern abweist, sieht die
// Begründung dort, wo er gerade nicht hinsieht.
//
// # Und was „visuell zurückgenommen" hier heißt
//
// **Nichts** — und das ist eine Entscheidung, kein Versäumnis: der Tisch
// zeichnet **nicht** optimistisch
// ([0126](../../../docs/entscheidungen/0126-der-tisch-zeichnet-nicht-optimistisch.md)).
// Was nie vorweggenommen wurde, muss nicht zurückgenommen werden.
import { screen, waitFor, within } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { antwortAbwarten, raeumeApi } from '../test/api'
import { einLauf, einZug, eineAntwort, element } from './partieprobe'
import { mitDecks, steigeEin, zeige as zeigeSeite } from './spielenprobe'
import { kataloge } from '../sprache/kataloge'
import { zeige } from './partietischprobe'

const eigene = kataloge.de.seiten.spielen

afterEach(() => {
  raeumeApi()
})

// **Am Tisch und nicht am Kopf der Seite.**
it('erklärt eine abgewiesene Handlung am Tisch', () => {
  zeige(einLauf({ fehler: 'Das darfst du gerade nicht.' }))

  expect(screen.getByRole('alert')).toHaveTextContent(
    'Das darfst du gerade nicht.',
  )
})

// **Und nur einmal.** Zwei gleiche Meldungen lassen jemanden zwei Fehler
// suchen — derselbe Satz steht schon über `Spielen.tsx` (§6).
it('sagt es nicht zweimal, solange ein Tisch offensteht', async () => {
  mitDecks({
    Partien: () =>
      Promise.resolve({
        partien: [
          {
            id: 'p1',
            fassung: '2',
            spiel: 'magic',
            stand: 'laeuft',
            sitze: [{ kennung: 'sitz1', name: 'Anna' }],
            objekte: 12,
            zug: 2,
            eigenerSitz: '',
            dran: false,
          },
        ],
        fehler: null,
      }),
    // **Mit einem Zug, denn ohne ihn gibt es keine Zugleiste** — ein
    // Sandkastentisch aus G1 hat keinen, und dann stünde der Knopf nicht da.
    Partie: () =>
      Promise.resolve(
        eineAntwort(
          {
            zug: einZug({
              nummer: 1,
              aktiv: 'sitz1',
              starter: 'sitz1',
              phase: 'haupt',
              schritt: '',
              prioritaet: 'sitz1',
            }),
          },
          // **Passen muss der Kern anbieten**, sonst steht der Knopf still
          // da (Zugleiste, Befund 2 des Zweitblicks über Phase 6.9) — hier
          // weist er ihn danach ab.
          ['zug.passen'],
        ),
      ),
    PartieBefehl: () =>
      Promise.resolve({
        partie: eineAntwort().partie,
        erlaubt: [],
        fehler: {
          code: 'konflikt',
          schluessel: '',
          meldung: 'Das darfst du gerade nicht.',
          feld: '',
        },
      }),
  })

  const hand = zeigeSeite()

  await steigeEin(hand, 'Anna', 'Anna')

  // **Passen ist der kürzeste Befehl** — er braucht keine Karte und keine
  // Zonenliste, und abgewiesen wird er wie jeder andere.
  await hand.click(
    await screen.findByRole('button', { name: eigene.zugPassen }),
  )

  await antwortAbwarten()

  // **Genau einmal.** Zwei gleiche Meldungen lassen jemanden zwei Fehler
  // suchen — derselbe Satz steht schon über `Spielen.tsx` (§6).
  await waitFor(() => {
    expect(
      screen
        .getAllByRole('alert')
        .filter((eine) =>
          eine.textContent.includes('Das darfst du gerade nicht.'),
        ),
    ).toHaveLength(1)
  })

  // **Und sie steht beim Tisch** — nicht irgendwo auf der Seite.
  expect(
    within(element(screen.getByRole('region', { name: /Am Tisch/ })))
      .getAllByRole('alert')
      .map((eine) => eine.textContent),
  ).toContain('Das darfst du gerade nicht.')
})
