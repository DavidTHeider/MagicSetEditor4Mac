// **Nur was angehakt ist** — Werkstück A11 von Phase 5.3
// (#1450).
//
// # Zwei Zusagen des Issues, und beide sind sichtbar
//
// „**Nur ausgewählte gültige Felder werden geändert.**" Und: „**Nutzende
// sehen konkrete Änderungen.**"
//
// **Eine eigene Datei**, weil `kivorschlag.test.tsx` mit ihr über die
// Zeilengrenze aus §2 ginge — und weil der Schnitt der Frage folgt: dort
// steht, **dass** ein Vorschlag entsteht, hier, **was davon geschrieben wird**.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { KIVorschlag } from './KIVorschlag'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { KIVorschlagAntwort } from '../antworten'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.einstellungen

// zweiFelder gibt einen Vorschlag über zwei Felder.
function zweiFelder(): KIVorschlagAntwort {
  return {
    kennung: 'ki-4-1',
    anbieter: 'ollama',
    modell: 'gemma4:8b',
    felder: [
      { feld: 'name', werte: ['Blitzschlag'], bisher: 'Flammenball' },
      { feld: 'text', werte: ['3 Schaden'], bisher: '' },
    ],
    befunde: [],
    ablauf: '2026-09-12T12:00:00Z',
    uebernommen: 0,
    belege: [],
    korpusstaende: [],
    einschraenkungen: [],
    weg: 'gewaehlt',
    ausgang: '',
    rueckfaelle: null,
    fehler: null,
  }
}

// aufschlagenUndVorschlagen führt bis zum fertigen Vorschlag.
async function aufschlagenUndVorschlagen() {
  const hand = userEvent.setup()

  await hand.click(screen.getByText(eigene.kiVorschlagTitel))

  await hand.selectOptions(
    await screen.findByRole('combobox', { name: eigene.kiVorschlagAnbieter }),
    'ollama/gemma4:8b',
  )

  await hand.type(
    screen.getByRole('textbox', { name: eigene.kiVorschlagAuftrag }),
    'Schreibe den Regeltext.',
  )

  await hand.click(screen.getByRole('button', { name: eigene.kiVorschlagen }))

  return hand
}

// welt stellt den Kasten mit einem zweifeldrigen Vorschlag hin.
function welt(uebernahme = vi.fn()) {
  belegeApi({
    KIModelle: () =>
      Promise.resolve({
        nurLokal: false,
        vorrang: [],
        vorrangbefunde: [],
        einrichtung: 'bereit',
        anbieter: [],
        cloud: { stand: 'ungewiss', beleg: 'keine Serverdatei' },
        befunde: [],
        modelle: [
          {
            anbieter: 'ollama',
            alias: 'gemma4:8b',
            laufzeit: 'ollama 0.12.3',
            ort: 'lokal',
            faehigkeiten: [],
            herkunft: [],
            bindung: 'ollama/gemma4:8b',
            kontextfenster: 0,
            kontextbeleg: '',
          },
        ],
        fehler: null,
      }),
    KIVorschlagen: () => Promise.resolve(zweiFelder()),
    KIVorschlagUebernehmen: uebernahme,
  })

  render(
    <Sprachraum sprache="de">
      <KIVorschlag karteID="4" geaendert={vi.fn()} />
    </Sprachraum>,
  )

  return uebernahme
}

// **Abgehakt heißt: nicht übernommen.**
it('ein abgehaktes Feld geht nicht mit', async () => {
  const uebernahme = welt(
    vi.fn().mockResolvedValue({ felder: 1, fehler: null }),
  )

  const hand = await aufschlagenUndVorschlagen()

  // **Auf die Wirkung warten und nicht auf den Ruf** (§3): der Haken steht
  // erst da, wenn der Vorschlag gezeichnet ist.
  const zweiter = await screen.findByRole('checkbox', {
    name: 'Übernehmen: text',
  })

  await hand.click(zweiter)

  await hand.click(
    screen.getByRole('button', { name: eigene.kiVorschlagUebernehmen }),
  )

  expect(uebernahme).toHaveBeenCalledWith('ki-4-1', ['name'])
})

// **Und wer alles abhakt, schreibt nichts.**
//
// **Der Knopf ist dann aus, und daneben steht, warum** — ein Knopf, der
// nichts tut, ist schlimmer als keiner (§6).
it('ohne Haken lässt sich nichts übernehmen', async () => {
  const uebernahme = welt()

  const hand = await aufschlagenUndVorschlagen()

  for (const feld of ['name', 'text']) {
    await hand.click(
      await screen.findByRole('checkbox', { name: `Übernehmen: ${feld}` }),
    )
  }

  expect(
    screen.getByRole('button', { name: eigene.kiVorschlagUebernehmen }),
  ).toBeDisabled()

  expect(screen.getByText(eigene.kiVorschlagNichtsGewaehlt)).toBeInTheDocument()

  expect(uebernahme).not.toHaveBeenCalled()
})

// **Was heute im Feld steht, steht daneben.**
//
// # Warum beides und nicht nur das Neue
//
// **Das Issue verlangt „Feld-Diff"**: ein Vorschlag ohne den alten Wert
// verlangt von jemandem, dass er ihn im Kopf hat — bei acht Feldern hat ihn
// niemand.
it('der bisherige Wert steht am Vorschlag', async () => {
  welt()

  await aufschlagenUndVorschlagen()

  expect(await screen.findByText('Bisher: Flammenball')).toBeInTheDocument()

  // **Und ein leeres Feld sagt, dass es leer ist** — statt gar nichts zu
  // zeigen, was aussähe wie ein vergessener Wert.
  expect(screen.getByText(eigene.kiVorschlagBisherLeer)).toBeInTheDocument()
})

// **Ein Haken lässt sich wieder setzen.**
//
// **Der Weg zurück ist der halbe Weg** (§10): wer nur das Abhaken prüft,
// prüft die Hälfte — und ein Zweig, der die Liste nur kürzt, sähe genauso
// aus.
it('ein abgehaktes Feld lässt sich wieder anhaken', async () => {
  const uebernahme = welt(
    vi.fn().mockResolvedValue({ felder: 2, fehler: null }),
  )

  const hand = await aufschlagenUndVorschlagen()

  const zweiter = await screen.findByRole('checkbox', {
    name: 'Übernehmen: text',
  })

  await hand.click(zweiter)
  await hand.click(zweiter)

  await hand.click(
    screen.getByRole('button', { name: eigene.kiVorschlagUebernehmen }),
  )

  expect(uebernahme).toHaveBeenCalledWith('ki-4-1', ['name', 'text'])
})
