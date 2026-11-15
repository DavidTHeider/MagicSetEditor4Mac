// **Die Leihe ohne Formularwand** — UXR-13
// (#1921).
//
// # Wogegen hier geprüft wird
//
// **Gegen die Leihe, wie sie bis UXR-13 auf der Spielenseite stand**: drei
// Erklärungen, Entleiher, Deck, Stunden, Ziel, „Verleihen und ablegen",
// „Leihen ansehen", der Umschlag samt „Leihe annehmen" und die Liste — alles
// offen untereinander, kein Knopf führte (R1, R3 in `docs/UIREVIEW.md`).
//
//   - **Offen ist eine Gruppe**: das Verleihen, mit seinem Knopf als
//     Hauptknopf. Das Annehmen eines Umschlags ist eingeklappt.
//   - **Die Stunden sind eine Zahl** (R7): ein Textfeld nahm „viele" an und
//     verlieh dann stillschweigend für die Vorgabe.
import { screen } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { Vorgabestunden } from './Deckleihe'
import { eigene, leiheApi, oeffneAnnahme, zeige } from '../test/leiheprobe'

it('zeigt offen nur das Verleihen und klappt die Annahme ein', async () => {
  leiheApi({ Deckleihen: vi.fn() })

  const hand = zeige()

  expect(
    screen.getByRole('heading', { name: eigene.leiheVerleihenTitel }),
  ).toBeInTheDocument()
  expect(screen.queryByLabelText(eigene.leiheUmschlag)).not.toBeInTheDocument()

  // **Das Verleihen führt** — ein Hauptknopf, und nur einer.
  expect(
    screen.getByRole('button', { name: eigene.leiheAnbieten }),
  ).toHaveClass('bg-marke')

  await oeffneAnnahme(hand)

  expect(screen.getByLabelText(eigene.leiheUmschlag)).toBeInTheDocument()
  expect(
    screen.getByRole('button', { name: eigene.leiheAnnehmen }),
  ).not.toHaveClass('bg-marke')
})

it('fragt die Stunden als Zahl', () => {
  leiheApi({ Deckleihen: vi.fn() })

  zeige()

  expect(
    screen.getByRole('spinbutton', { name: eigene.leiheStunden }),
  ).toHaveValue(Vorgabestunden)
})

it('nennt die Leihen unter eigener Überschrift', () => {
  leiheApi({ Deckleihen: vi.fn() })

  zeige()

  expect(
    screen.getByRole('heading', { name: eigene.leiheListeTitel }),
  ).toBeInTheDocument()
  expect(
    screen.getByRole('button', { name: eigene.leiheAnsehen }),
  ).toBeInTheDocument()
})
