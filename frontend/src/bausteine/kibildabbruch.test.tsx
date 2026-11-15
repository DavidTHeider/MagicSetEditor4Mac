// **Der Abbruch eines laufenden Bildauftrags** — Werkstück A17 von Phase 5.3
// (#1532).
//
// **Eine eigene Datei**, seit `kibild.test.tsx` mit dem Fall aus Befund 3 des
// Zweitblicks über Phase 5.3 über die Zeilengrenze aus §2 ging. Der Schnitt
// folgt der Frage: dort steht, wie ein Bild **entsteht** und übernommen wird,
// hier, wie man es **anhält**.
import { screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import {
  aufschlagenUndErzeugen,
  bildtexte as eigene,
  einKatalog,
  einVorschlag,
  mitFreigabe,
  ohneNachweis,
  zeigeBildkasten as zeige,
} from '../test/kibild'

afterEach(() => {
  raeumeApi()
})

// **Während ein Bild entsteht, steht da, worauf gewartet wird — und ein
// Abbruch daneben** (A17,
// #1532).
//
// # Der Befund
//
// **Es gab keinen Abbruch.** Der Kasten kannte einen Zustand `laeuft`, der
// Knöpfe abschaltet — und nichts, was den Lauf beendet. Ein Bildauftrag
// dauert Sekunden bis Minuten; wer ihn nicht wollte, wartete ihn ab.
//
// **Und „es läuft" sagte nicht, wen man bezahlt.** Anbieter und Modell
// stehen jetzt darin.
it('nennt beim Warten Anbieter und Modell — und bietet den Abbruch an', async () => {
  let loesen = (): void => undefined

  const abbrechen = vi.fn(() => Promise.resolve(true))

  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildAbbrechen: abbrechen,
      KIBildVorschlagen: () =>
        new Promise((fertig) => {
          loesen = () => {
            fertig(einVorschlag())
          }
        }),
    }),
  )

  const { hand } = zeige()

  await aufschlagenUndErzeugen(hand)

  // **Der ganze Satz und nicht ein Ausdruck**: derselbe Anbieter steht auch
  // in der Auswahlliste, und ein loser Ausdruck fände beide.
  expect(
    await screen.findByText(
      einsetzen(eigene.kiBildLaeuft, {
        anbieter: 'openai',
        modell: 'gpt-image-2.5-flare',
      }),
    ),
  ).toBeInTheDocument()

  await hand.click(screen.getByRole('button', { name: eigene.kiBildAbbrechen }))

  expect(abbrechen).toHaveBeenCalledWith(4)

  loesen()

  await screen.findByText(eigene.kiBildAbgebrochen)
})

// **Und ohne laufenden Auftrag gibt es keinen Abbruch.**
//
// **Die Gegenprobe** (§10): ein Knopf, der immer dasteht, hielte den Lauf
// darüber, ohne dass „während es läuft" etwas entschiede.
it('bietet keinen Abbruch an, solange nichts läuft', async () => {
  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
    }),
  )

  const { hand } = zeige()

  await hand.click(screen.getByText(eigene.kiBildTitel))

  await screen.findByLabelText(eigene.kiVorschlagAnbieter)

  expect(
    screen.queryByRole('button', { name: eigene.kiBildAbbrechen }),
  ).not.toBeInTheDocument()
})

// **Und „abgebrochen" steht nur da, wenn wirklich etwas abgebrochen wurde** —
// Befund 3 des Zweitblicks über Phase 5.3.
//
// # Der Befund
//
// **Der Knopf setzte den Satz, ohne die Antwort zu lesen.** `KIBildAbbrechen`
// gibt zurück, **ob** ein Auftrag lief; das Fenster warf den Wert weg und
// schrieb „Abgebrochen. Es wurde nichts übernommen …" — auch dann, wenn der
// Auftrag längst durch war und sein Ergebnis gleich danach ankam.
//
// **Auf der Go-Seite war es derselbe Befund**: das Ende eines Laufs nahm dem
// zweiten seinen Abbruchweg, und `KIBildAbbrechen` meldete dann `false`,
// während der Auftrag lief. Wer das `false` wegwirft, sieht auch das nicht.
it('sagt nicht „abgebrochen", wenn nichts abzubrechen war', async () => {
  let loesen = (): void => undefined

  const abbrechen = vi.fn(() => Promise.resolve(false))

  belegeApi(
    mitFreigabe({
      KIModelle: () => Promise.resolve(einKatalog()),
      KIBildHerkunft: () => Promise.resolve(ohneNachweis()),
      KIBildAbbrechen: abbrechen,
      KIBildVorschlagen: () =>
        new Promise((fertig) => {
          loesen = () => {
            fertig(einVorschlag())
          }
        }),
    }),
  )

  const { hand } = zeige()

  await aufschlagenUndErzeugen(hand)

  await screen.findByRole('button', { name: eigene.kiBildAbbrechen })
  await hand.click(screen.getByRole('button', { name: eigene.kiBildAbbrechen }))

  // **Auf die Antwort warten und nicht auf den Ruf** — was danach dasteht,
  // steht erst nach ihr fest.
  await waitFor(() => {
    expect(abbrechen).toHaveBeenCalledWith(4)
  })

  loesen()

  // **Der Vorschlag kommt an** — er war schon fertig, als geklickt wurde.
  expect(await screen.findByAltText(/Vorschlag für/)).toBeInTheDocument()

  expect(screen.queryByText(eigene.kiBildAbgebrochen)).not.toBeInTheDocument()
})
