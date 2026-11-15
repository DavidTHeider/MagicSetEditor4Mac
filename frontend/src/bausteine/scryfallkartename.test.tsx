// **Der vorbelegte Kartenname** — Werkstück 1 von Phase 3.6
// (#1167).
//
// **Eine eigene Datei**, weil `seiten/scryfallkarte.test.tsx` den Weg über die
// Seite prüft; hier steht die eine Eigenschaft des Bausteins, die von außen
// gesetzt wird.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { Scryfallkarte } from './Scryfallkarte'

const texte = kataloge.de.seiten.karten.scryfallkarte

describe('der vorbelegte Kartenname', () => {
  // **Ohne Vorbelegung ist der Kasten leer.**
  it('ist leer, wenn niemand einen Namen setzt', () => {
    render(
      <Sprachraum sprache="de">
        <Scryfallkarte anlegen={vi.fn()} laeuft={false} />
      </Sprachraum>,
    )

    expect(screen.getByLabelText(texte.name)).toHaveValue('')
  })

  // **Mit Vorbelegung steht er da — und bleibt tippbar.**
  //
  // Ein Kasten, der sich nicht mehr ändern ließe, machte aus der Hilfe eine
  // Bevormundung: wer den Namen danach anpassen will, soll es können.
  it('zeigt den gesetzten Namen und lässt ihn ändern', async () => {
    render(
      <Sprachraum sprache="de">
        <Scryfallkarte anlegen={vi.fn()} laeuft={false} name="Lightning Bolt" />
      </Sprachraum>,
    )

    const kasten = screen.getByLabelText(texte.name)

    expect(kasten).toHaveValue('Lightning Bolt')

    await userEvent.type(kasten, '!')

    expect(kasten).toHaveValue('Lightning Bolt!')
  })
})
