// **Einstellungen mitnehmen und einlesen** — Werkstück O3 von Phase 5.2
// (#1427).
//
// # Was hier geprüft wird
//
// **Der zweite Schritt** (§16, Regel 6): dass eine Datei erst **gezeigt** und
// dann übernommen wird, dass eine abgewählte Zeile nicht mitgeht, und dass die
// **Revision** aus der Vorschau zurückgeht — ohne sie ginge die Übernahme
// gegen einen Stand, den niemand gesehen hat (§7, Regel 2).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Einstellungsaustausch } from './Einstellungsaustausch'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'

import type { Einstellungssicht } from '../api'

afterEach(() => {
  raeumeApi()
})

const texte = {
  sichern: 'Einstellungen sichern',
  einlesen: 'Datei einlesen',
  uebernehmen: 'Gewählte übernehmen',
  verwerfen: 'Verwerfen',
  layout: 'Fensterlayout mitnehmen (hängt am Bildschirm)',
  zeileSprache: 'Die Zeile profil:sprache übernehmen',
  zeileKontrast: 'Die Zeile nutzer:barrierefreiheit.kontrast übernehmen',
}

// zweizeilig ist eine Vorschau mit zwei Zeilen, einem Befund und einer
// unveränderten Zeile.
const zweizeilig: Einstellungssicht = {
  inhalt: '{"fassung":1,"bereiche":{}}',
  fassung: 1,
  revision: 'abc123',
  zeilen: [
    {
      ebene: 'nutzer',
      schluessel: 'barrierefreiheit.kontrast',
      vorher: 'system',
      nachher: 'hoch',
      herkunft: 'vorgabe',
      wirkung: 'sofort',
    },
    {
      ebene: 'profil',
      schluessel: 'sprache',
      vorher: 'de',
      nachher: 'en',
      herkunft: 'profil',
      wirkung: 'sofort',
    },
  ],
  befunde: [
    {
      ebene: 'profil',
      schluessel: 'vorgabe.hochladepfad',
      grund: 'Dieser Schlüssel beschreibt einen Rechner; er reist nicht mit.',
    },
  ],
  unveraendert: 3,
}

// leereSicht ist eine Vorschau, die nichts ändert.
const leereSicht: Einstellungssicht = {
  inhalt: '{"fassung":1,"bereiche":{}}',
  fassung: 1,
  revision: 'leer1',
  zeilen: [],
  befunde: [],
  unveraendert: 0,
}

// zeige stellt den Kasten mit einer belegten Fassade auf.
function zeige(teile: Parameters<typeof belegeApi>[0]) {
  belegeApi(teile)

  render(
    <Sprachraum sprache="de">
      <Einstellungsaustausch />
    </Sprachraum>,
  )
}

// einlesenMit belegt das Einlesen mit einer Sicht.
function einlesenMit(sicht: Einstellungssicht) {
  return () => Promise.resolve({ sicht, fehler: null })
}

// **Die Vorschau zeigt jede Zeile mit vorher, nachher und Herkunft.**
it('zeigt den Unterschied, bevor etwas geschieht', async () => {
  const uebernehmen = vi.fn()

  zeige({
    EinstellungenEinlesen: einlesenMit(zweizeilig),
    EinstellungenUebernehmen: uebernehmen,
  })

  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))

  expect(await screen.findByText('profil:sprache')).toBeInTheDocument()
  expect(screen.getByText('diesem Profil')).toBeInTheDocument()
  expect(screen.getByText('3 Zeilen stehen schon so da.')).toBeInTheDocument()

  // **Der Befund steht da und wird nicht verschluckt** (§8).
  expect(
    screen.getByText(/Dieser Schlüssel beschreibt einen Rechner/),
  ).toBeInTheDocument()

  // **Und geschrieben wurde nichts.**
  expect(uebernehmen).not.toHaveBeenCalled()
})

// **Die Übernahme nimmt die Revision aus der Vorschau mit.**
it('übernimmt gegen den Stand, den die Vorschau zeigte', async () => {
  const uebernehmen = vi.fn(() =>
    Promise.resolve({
      bericht: {
        uebernommen: 2,
        revision: 'neu456',
        zeilen: [],
        befunde: [],
      },
      fehler: null,
    }),
  )

  zeige({
    EinstellungenEinlesen: einlesenMit(zweizeilig),
    EinstellungenUebernehmen: uebernehmen,
  })

  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))
  await screen.findByText('profil:sprache')

  await userEvent.click(screen.getByRole('button', { name: texte.uebernehmen }))

  await waitFor(() => {
    expect(uebernehmen).toHaveBeenCalledWith(zweizeilig.inhalt, 'abc123', [
      'nutzer:barrierefreiheit.kontrast',
      'profil:sprache',
    ])
  })

  // **Und die Vorschau ist danach weg** (§7, Regel 1): sie beschreibt einen
  // Stand, den es nicht mehr gibt.
  expect(await screen.findByText('2 Zeilen übernommen.')).toBeInTheDocument()
  expect(screen.queryByText('profil:sprache')).not.toBeInTheDocument()
})

// **Eine abgewählte Zeile geht nicht mit.**
it('lässt eine abgewählte Zeile stehen', async () => {
  const uebernehmen = vi.fn(() =>
    Promise.resolve({
      bericht: { uebernommen: 1, revision: 'neu', zeilen: [], befunde: [] },
      fehler: null,
    }),
  )

  zeige({
    EinstellungenEinlesen: einlesenMit(zweizeilig),
    EinstellungenUebernehmen: uebernehmen,
  })

  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))

  await userEvent.click(
    await screen.findByRole('checkbox', { name: texte.zeileSprache }),
  )

  await userEvent.click(screen.getByRole('button', { name: texte.uebernehmen }))

  await waitFor(() => {
    expect(uebernehmen).toHaveBeenCalledWith(zweizeilig.inhalt, 'abc123', [
      'nutzer:barrierefreiheit.kontrast',
    ])
  })
})

// **Verworfen heißt weg — und nichts geschrieben.**
it('räumt die Vorschau, ohne zu schreiben', async () => {
  const uebernehmen = vi.fn()

  zeige({
    EinstellungenEinlesen: einlesenMit(zweizeilig),
    EinstellungenUebernehmen: uebernehmen,
  })

  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))
  await screen.findByText('profil:sprache')

  await userEvent.click(screen.getByRole('button', { name: texte.verwerfen }))

  expect(screen.queryByText('profil:sprache')).not.toBeInTheDocument()
  expect(uebernehmen).not.toHaveBeenCalled()
})

// **Eine Datei, die nichts ändert, sagt es — und der Knopf ist aus.**
it('sagt es, wenn eine Datei nichts ändert', async () => {
  zeige({ EinstellungenEinlesen: einlesenMit(leereSicht) })

  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))

  expect(
    await screen.findByText(
      'Diese Datei ändert nichts an deinen Einstellungen.',
    ),
  ).toBeInTheDocument()

  expect(screen.getByRole('button', { name: texte.uebernehmen })).toBeDisabled()
})

// **Ein Fehler steht da, und die Vorschau bleibt stehen** (§7, Regel 3).
it('zeigt einen gescheiterten Schreibvorgang und räumt nicht', async () => {
  zeige({
    EinstellungenEinlesen: einlesenMit(zweizeilig),
    EinstellungenUebernehmen: () =>
      Promise.resolve({
        bericht: {
          uebernommen: 0,
          revision: '',
          zeilen: [],
          befunde: [],
        },
        fehler: {
          code: 'konflikt',
          schluessel: '',
          meldung: 'Die Einstellungen haben sich geändert.',
        },
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))
  await screen.findByText('profil:sprache')

  await userEvent.click(screen.getByRole('button', { name: texte.uebernehmen }))

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Die Einstellungen haben sich geändert.',
  )

  // **Die Vorschau steht noch da** — sonst müsste jemand die Datei erneut
  // suchen, nur weil der Stand sich geändert hat.
  expect(screen.getByText('profil:sprache')).toBeInTheDocument()
})

// **Ein gescheitertes Einlesen zeigt seinen Fehler.**
it('zeigt einen gescheiterten Dialog', async () => {
  zeige({
    EinstellungenEinlesen: () =>
      Promise.resolve({
        sicht: {
          inhalt: '',
          fassung: 0,
          revision: '',
          zeilen: [],
          befunde: [],
          unveraendert: 0,
        },
        fehler: {
          code: 'validierung',
          schluessel: '',
          meldung: 'Die Datei fehlt.',
        },
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: texte.einlesen }))

  expect(await screen.findByRole('alert')).toHaveTextContent('Die Datei fehlt.')
})

// **Das Sichern reicht die gewählten Abschnitte weiter.**
it('nimmt den Arbeitsplatz nur auf Wunsch mit', async () => {
  const sichern = vi.fn(() => Promise.resolve({ fehler: null }))

  zeige({ EinstellungenSichern: sichern })

  await userEvent.click(screen.getByRole('button', { name: texte.sichern }))
  await waitFor(() => {
    expect(sichern).toHaveBeenCalledWith([])
  })

  // **Auf das Ende des ersten Sicherns gewartet, nicht auf den Ruf** (#1809):
  // solange es läuft, ruhen Kästchen und Knopf, und ein Klick ginge verloren.
  await waitFor(() => {
    expect(screen.getByRole('button', { name: texte.sichern })).toBeEnabled()
  })

  await userEvent.click(screen.getByRole('checkbox', { name: texte.layout }))
  await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

  await waitFor(() => {
    expect(sichern).toHaveBeenLastCalledWith([
      'nutzer',
      'profil',
      'arbeitsplatz',
    ])
  })
})

// **Ein gescheitertes Sichern steht da.**
it('zeigt einen gescheiterten Speicherdialog', async () => {
  zeige({
    EinstellungenSichern: () =>
      Promise.resolve({
        fehler: {
          code: 'validierung',
          schluessel: '',
          meldung: 'Kein Ziel gewählt.',
        },
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: texte.sichern }))

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Kein Ziel gewählt.',
  )
})

// **Eine überholte Antwort wird nicht eingetragen** (§7, Regel 2).
//
// **Der Beleg ist die zweite Datei**: das erste Einlesen antwortet langsam,
// das zweite schnell. Ohne den Laufzähler stünde am Ende die **erste** Sicht
// da — die zu einer Datei gehört, die niemand mehr offen hat.
it('lässt eine überholte Vorschau fallen', async () => {
  let welche = 0

  zeige({
    EinstellungenEinlesen: () => {
      welche += 1

      const sicht = welche === 1 ? zweizeilig : leereSicht

      return new Promise((fertig) => {
        setTimeout(
          () => {
            fertig({ sicht, fehler: null })
          },
          welche === 1 ? 30 : 0,
        )
      })
    },
  })

  const knopf = screen.getByRole('button', { name: texte.einlesen })

  await userEvent.click(knopf)
  await screen.findByText('profil:sprache')

  await userEvent.click(knopf)

  expect(
    await screen.findByText(
      'Diese Datei ändert nichts an deinen Einstellungen.',
    ),
  ).toBeInTheDocument()

  // **Und die erste Antwort kommt danach und ändert nichts mehr.**
  await antwortAbwarten()

  expect(screen.queryByText('profil:sprache')).not.toBeInTheDocument()
})
