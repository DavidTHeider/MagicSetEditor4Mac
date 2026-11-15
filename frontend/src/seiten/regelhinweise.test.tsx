// **Was vor dem Anfangen sichtbar bleiben muss** — die Hindernisse einer
// Absage und die Widersprüche einer Abwahl (G4, PLAY-02).
//
// **Eine eigene Datei**, seit die langen Aufzählungen der Regelschau
// eingeklappt stehen (das Artboard *Eine Partie anfangen*, Befund F13):
// was den Anfang anhält oder eine Abwahl erklärt, darf nicht mit ihnen
// verschwinden. Die Läufe standen bis zum 2026-09-24 in
// `regelschau.test.tsx`.
import { render, screen, within } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Regelhinweise } from './Regelhinweise'
import { Sprachraum } from '../sprache/kontext'
import { eineFreigabe } from '../test/regelprobe'
import { kataloge } from '../sprache/kataloge'

import type { RegelfreigabeDTO } from '../apiregeln'

const eigene = kataloge.de.seiten.spielen

function zeige(freigabe: RegelfreigabeDTO) {
  render(
    <Sprachraum sprache="de">
      <Regelhinweise freigabe={freigabe} />
    </Sprachraum>,
  )
}

// **Und eine Absage erscheint als Warnung.**
//
// **`alert` und nicht `status`**: hier geht es nicht weiter, und das soll
// eine Vorlesehilfe sofort sagen.
it('zeigt die Hindernisse einer Absage', () => {
  zeige(
    eineFreigabe({
      erlaubt: false,
      hindernisse: ['das Regelwerk trägt die Strenge »streng« nicht'],
    }),
  )

  expect(screen.getByRole('alert')).toHaveTextContent(
    'das Regelwerk trägt die Strenge »streng« nicht',
  )
  // **Und bleibt eine Liste** (P4, #2372, axe in WebKit): eine `ul` mit der
  // Rolle `alert` ist keine mehr.
  expect(
    within(screen.getByRole('alert')).getByRole('list'),
  ).toBeInTheDocument()
})

// **Ein Widerspruch wird erklärt — mit beiden Namen und seiner Quelle.**
//
// **Erklärt und nicht gesperrt** (AK-02): die Partie darf trotzdem
// anfangen, und wer die Priorität selbst übernimmt, soll wissen, dass der
// Stapel dann auf ihn wartet.
it('erklärt, was an einer Abwahl noch hängt', () => {
  zeige(
    eineFreigabe({
      widersprueche: [
        {
          bereich: 'prioritaet',
          haengt: 'stapel',
          quelle: 'CR 608 (Fassung 2026-08-19)',
        },
      ],
    }),
  )

  const erklaerung = screen.getByText(/CR 608/)

  expect(erklaerung).toHaveTextContent(eigene.regelbereichPrioritaet)
  expect(erklaerung).toHaveTextContent(eigene.regelbereichStapel)
})

// **Ohne Widerspruch steht die Überschrift nicht da** — eine leere
// Abteilung wäre eine Frage, die niemand gestellt hat.
it('zeigt die Überschrift nur, wo es etwas zu erklären gibt', () => {
  zeige(eineFreigabe())

  expect(screen.queryByText(eigene.regelWidersprueche)).not.toBeInTheDocument()
})
