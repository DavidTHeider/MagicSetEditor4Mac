// **Die Vorgabe für den Kartenzustand** — Werkstück 9 von Phase 4.12
// (#1351),
// Punkt 3.
//
// **Eine eigene Datei**, weil `sammlungsformular.test.tsx` mit ihr über 400
// Zeilen ginge (§2). Der Schnitt folgt der Frage: dort steht, **was das
// Formular kann**, hier, **womit es beginnt**.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sammlungsformular } from './Sammlungsformular'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { probeeinstellung } from '../test/einstellung'

const texte = kataloge.de.seiten.sammlung

describe('Die Vorgabe für den Kartenzustand', () => {
  afterEach(raeumeApi)

  // **„ohne" ist ein Wert im Bestand und kein Kartenzustand.**
  //
  // `vorgabe.zustand` kennt ihn, seit die Profilvorgaben da sind (Werkstück 9,
  // #1351): er heißt „ich lege nichts fest". Wer ihn durchreichte, trüge
  // „ohne" als Zustand in jedes neue Exemplar ein.
  //
  // **Der Lauf steht hier, weil der Zweig sonst unbetreten bleibt.** Ohne ihn
  // meldete die Abdeckung 83 % an einer Zeile, die genau diesen Fehler
  // verhindert — und niemand sähe, welcher Fall fehlt (§3).
  it('trägt „ohne" nicht als Zustand ein', async () => {
    belegeApi({
      Einstellungen: () =>
        Promise.resolve({
          einstellungen: [
            probeeinstellung({
              schluessel: 'vorgabe.zustand',
              wert: 'ohne',
              werte: ['ohne', 'mint'],
              bereich: 'vorgaben',
            }),
          ],
          fehler: null,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Sammlungsformular aufAufnehmen={vi.fn()} laeuft={false} />
      </Sprachraum>,
    )

    // **Auf die Antwort gewartet und nicht auf die Zusicherung** (§3): „das
    // Feld ist leer" trifft **vor** der Antwort ohnehin zu, und der Lauf war
    // grün, ohne den Zweig je zu betreten. Aufgefallen an der Abdeckung —
    // sie meldete 87,5 % an einer Zeile, die genau diesen Fall behandelt.
    await antwortAbwarten()

    expect(screen.getByLabelText(texte.zustand)).toHaveValue('')
  })

  // **Und die Gegenrichtung**: ein echter Zustand wird übernommen.
  //
  // Ohne ihn stünde der Lauf darüber auch dann grün da, wenn die Vorgabe
  // überhaupt nicht ankäme (§10).
  it('übernimmt eine echte Vorgabe', async () => {
    belegeApi({
      Einstellungen: () =>
        Promise.resolve({
          einstellungen: [
            probeeinstellung({
              schluessel: 'vorgabe.zustand',
              wert: 'mint',
              werte: ['ohne', 'mint'],
              bereich: 'vorgaben',
            }),
          ],
          fehler: null,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Sammlungsformular aufAufnehmen={vi.fn()} laeuft={false} />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(screen.getByLabelText(texte.zustand)).toHaveValue('mint')
    })
  })

  // **Und die Währung** — LOC-03
  // (#1689):
  // sie ist eine eigene Einstellung und steht vor, statt zu rechnen (AK-02).
  // **Die Sprache ist dabei Englisch** — gerade damit sichtbar wird, dass
  // aus ihr keine Währung folgt.
  it('beginnt mit der Vorgabewährung, gleich in welcher Sprache', async () => {
    belegeApi({
      Einstellungen: () =>
        Promise.resolve({
          einstellungen: [
            probeeinstellung({
              schluessel: 'vorgabe.waehrung',
              wert: 'JPY',
              werte: [],
              bereich: 'vorgaben',
            }),
          ],
          fehler: null,
        }),
    })

    render(
      <Sprachraum sprache="en">
        <Sammlungsformular aufAufnehmen={vi.fn()} laeuft={false} />
      </Sprachraum>,
    )

    // **Die Währung steht seit UXR-10 (#1918) unter „Weitere Angaben"** —
    // aufgeklappt wird, wenn die Vorgabe da ist: mit ihr entsteht das
    // Formular neu (`key`), und ein Kasten davor stünde wieder zu.
    await antwortAbwarten()
    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.en.seiten.sammlung.weitereAngaben,
      }),
    )

    await waitFor(() => {
      expect(
        screen.getByLabelText(kataloge.en.seiten.sammlung.erwerbswaehrung),
      ).toHaveValue('JPY')
    })
  })

  // **Ohne Vorgabe bleibt das Feld leer** — wie es immer war.
  it('lässt die Währung ohne Vorgabe leer', async () => {
    belegeApi({
      Einstellungen: () => Promise.resolve({ einstellungen: [], fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Sammlungsformular aufAufnehmen={vi.fn()} laeuft={false} />
      </Sprachraum>,
    )

    await antwortAbwarten()
    await userEvent.click(
      screen.getByRole('button', { name: texte.weitereAngaben }),
    )

    expect(screen.getByLabelText(texte.erwerbswaehrung)).toHaveValue('')
  })
})
