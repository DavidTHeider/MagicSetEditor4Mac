// **Was ein KI-Auftrag kostet, wird vorher freigegeben** —
// #1539.
//
// # Der Befund
//
// **Am laufenden Fenster ging kein bezahlter Auftrag hinaus**: die
// Kostenwache verlangt die ausdrückliche Freigabe eines Menschen, und es gab
// keinen Weg, sie zu geben. Jeder Ruf endete mit einem Satz, auf den niemand
// reagieren konnte.
//
// **Kein Lauf sah es**, weil jeder ein Modell auf diesem Rechner wählte.
import { screen, waitFor } from '@testing-library/react'
import { afterEach, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import {
  ausfuellenUndVorschlagen,
  einKatalog,
  einVorschlag,
  vorschlagstexte as eigene,
  zeigeVorschlagskasten as zeige,
} from '../test/kivorschlag'

afterEach(() => {
  raeumeApi()
})

// **Ein Auftrag an ein Cloudmodell geht erst nach einer Freigabe hinaus.**
//
// # Der Befund
//
// **Am laufenden Fenster ging keiner** (#1539):
// die Kostenwache verlangt die ausdrückliche Freigabe eines Menschen, und es
// gab keinen Weg, sie zu geben. Jeder Ruf endete mit einem Satz, auf den
// niemand reagieren konnte.
//
// **Kein Lauf sah es**, weil jeder ein Modell auf diesem Rechner wählte.
it('fragt vor einem bezahlten Auftrag nach der Freigabe', async () => {
  const vorschlagen = vi.fn(() => Promise.resolve(einVorschlag()))

  belegeApi({
    KIModelle: () =>
      Promise.resolve(
        einKatalog({
          anbieter: 'anthropic',
          alias: 'claude-opus-5',
          ort: 'cloud',
        }),
      ),
    Kostenstand: () => Promise.resolve(einBudget()),
    KIFreigabeErteilen: () =>
      Promise.resolve({
        kennung: 'n-1',
        laeuftab: '2026-09-12T12:05:00Z',
        fehler: null,
      }),
    KIVorschlagen: vorschlagen,
  })

  zeige()

  const hand = await ausfuellenUndVorschlagen('anthropic/claude-opus-5')

  // **Erst der Beleg, dass der Kasten steht** — die Verneinung darunter
  // träfe sonst auch zu, wenn der Knopf gar nichts ausgelöst hätte (§3).
  const feld = await screen.findByLabelText(eigene.kiFreigabeBetrag)

  expect(vorschlagen).not.toHaveBeenCalled()

  await hand.type(feld, '1200')
  await hand.click(
    screen.getByRole('button', { name: eigene.kiFreigabeErteilen }),
  )

  expect(await screen.findByText('Blitzschlag')).toBeInTheDocument()

  expect(vorschlagen).toHaveBeenCalledWith('4', {
    anbieter: 'anthropic',
    alias: 'claude-opus-5',
    auftrag: 'Schreibe.',
    nachweis: 'n-1',

    // **Ohne Belegfrage und ohne Freigabe** (A13.4, #1452) — dieser Lauf
    // stellt keine, und dann wird der Suchindex gar nicht gefragt.
    belegfrage: '',
    freigegeben: [],
  })
})

// **Und ein Modell auf diesem Rechner fragt nicht.**
//
// **Weil eine Freigabe, die überall verlangt wird, keine Auskunft ist**: wer
// sie für ein kostenloses Modell geben müsste, gewöhnte sich das Wegklicken
// an — und dann gälte die Frage auch dort nichts mehr, wo sie zählt.
it('fragt nicht, wo nichts kostet', async () => {
  const erteilen = vi.fn()

  belegeApi({
    KIModelle: () =>
      Promise.resolve(einKatalog({ anbieter: 'ollama', alias: 'gemma4:8b' })),
    KIVorschlagen: () => Promise.resolve(einVorschlag()),
    KIFreigabeErteilen: erteilen,
  })

  zeige()

  await ausfuellenUndVorschlagen()

  expect(await screen.findByText('Blitzschlag')).toBeInTheDocument()
  expect(erteilen).not.toHaveBeenCalled()
})

// **Wer nicht freigibt, bekommt keinen Auftrag — und keine Meldung.**
//
// **Abgebrochen ist kein Fehler**: wer es sich anders überlegt, hat keinen
// Fehlschlag erlebt. Ein roter Hinweis daneben behauptete etwas anderes.
it('schickt nichts hinaus, wenn niemand freigibt', async () => {
  const vorschlagen = vi.fn()

  belegeApi({
    KIModelle: () =>
      Promise.resolve(
        einKatalog({
          anbieter: 'anthropic',
          alias: 'claude-opus-5',
          ort: 'cloud',
        }),
      ),
    Kostenstand: () => Promise.resolve(einBudget()),
    KIVorschlagen: vorschlagen,
  })

  zeige()

  const hand = await ausfuellenUndVorschlagen('anthropic/claude-opus-5')

  await hand.click(
    await screen.findByRole('button', { name: eigene.kiFreigabeVerzichten }),
  )

  // **Erst der Beleg, dass der Kasten weg ist** — die Verneinung darunter
  // träfe sonst auch zu, wenn der Knopf nichts ausgelöst hätte (§3).
  await waitFor(() => {
    expect(screen.queryByText(eigene.kiFreigabeTitel)).not.toBeInTheDocument()
  })

  expect(vorschlagen).not.toHaveBeenCalled()
  expect(screen.queryByRole('alert')).not.toBeInTheDocument()
})

// einBudget ist die Auskunft, mit der eine Freigabe möglich ist.
function einBudget() {
  const betrag = (wert: number) => ({ wert, waehrung: 'EUR' })

  return {
    konto: 'anthropic',
    zeitraum: 'tag',
    grenze: betrag(50_000),
    reserviert: betrag(0),
    verbraucht: betrag(0),
    unklar: betrag(0),
    verfuegbar: betrag(50_000),
    fehler: null,
  }
}
