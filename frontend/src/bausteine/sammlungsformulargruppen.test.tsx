// **Das Formular in zwei Gruppen** — UXR-10
// (#1918),
// R3 in `docs/UIREVIEW.md`: höchstens zwölf Felder, sonst gruppiert und
// eingeklappt bis auf eine Gruppe.
//
// **Vierzehn Felder standen offen da**, gemessen am 2026-10-02 am nativen
// Fenster. Offen bleibt, was ein Exemplar ausmacht — Druck, Menge, Zustand,
// Sprache, Ausführung, Lagerort, Absicht —, eingeklappt der lesbare Bezug,
// der Erwerb und die Notiz.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sammlungsformular } from './Sammlungsformular'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { druckfeld } from '../test/druckfeld'
import { felder } from '../test/sammlungsseite'

const texte = kataloge.de.seiten.sammlung

afterEach(raeumeApi)

function zeige(aufAufnehmen = vi.fn(() => Promise.resolve(true))) {
  render(
    <Sprachraum sprache="de">
      <Sammlungsformular aufAufnehmen={aufAufnehmen} laeuft={false} />
    </Sprachraum>,
  )

  return aufAufnehmen
}

// **Seit AP-10 führt der Name**, und die Kennung steht eingeklappt
// (`Wege09`).
const offen = [
  texte.drucksuche.karte,
  texte.menge,
  texte.zustand,
  texte.sprache,
  texte.ausfuehrung,
  texte.lagerort,
  texte.wunsch,
]

const eingeklappt = [
  texte.druck,
  texte.name,
  texte.setcode,
  texte.nummer,
  texte.erwerbspreis,
  texte.erwerbswaehrung,
  texte.erwerbsdatum,
  texte.notiz,
]

describe('Das Formular in zwei Gruppen', () => {
  it('zeigt, was ein Exemplar ausmacht, und klappt das Übrige ein', async () => {
    zeige()

    for (const feld of offen) {
      expect(screen.getByLabelText(feld)).toBeInTheDocument()
    }
    for (const feld of eingeklappt) {
      expect(screen.queryByLabelText(feld)).toBeNull()
    }
    expect(felder()).toBe(offen.length)

    const knopf = screen.getByRole('button', { name: texte.weitereAngaben })
    expect(knopf).toHaveAttribute('aria-expanded', 'false')

    await userEvent.click(knopf)

    for (const feld of eingeklappt) {
      expect(screen.getByLabelText(feld)).toBeInTheDocument()
    }
  })

  // **Ein Befund in einer zugeklappten Gruppe sähe niemand** — und der
  // Knopf ruhte oder schriebe nicht, ohne dass einer sagt, warum.
  it('öffnet die weiteren Angaben, wenn der Preis zwei Lesarten hat', async () => {
    const aufgenommen = zeige()

    await userEvent.type(await druckfeld(), 'abc')

    await userEvent.type(screen.getByLabelText(texte.erwerbswaehrung), 'BHD')
    await userEvent.type(screen.getByLabelText(texte.erwerbspreis), '1.234')

    // **Wieder zugeklappt**, bevor jemand aufnimmt — der Fall, für den es
    // die Zusage gibt.
    await userEvent.click(
      screen.getByRole('button', { name: texte.weitereAngaben }),
    )
    expect(screen.getByLabelText(texte.erwerbspreis)).not.toBeVisible()

    await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))

    expect(await screen.findByText(/1\.234,000/)).toBeVisible()
    expect(
      screen.getByRole('button', { name: texte.weitereAngaben }),
    ).toHaveAttribute('aria-expanded', 'true')
    expect(aufgenommen).not.toHaveBeenCalled()

    // **Und sie bleibt offen, während der Preis berichtigt wird** — sonst
    // klappte sie unter den Fingern zu, sobald der Befund geht.
    await userEvent.type(screen.getByLabelText(texte.erwerbspreis), '5')
    expect(screen.getByLabelText(texte.erwerbspreis)).toBeVisible()
  })
})
