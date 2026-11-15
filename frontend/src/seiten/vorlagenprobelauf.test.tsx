// „Wie es wird" vor der Exportvorlage — UXR-F7
// (#2015).
//
// Die Vorlage wird ausgewertet, bevor sie schreibt: Name und Größe der Datei
// stehen da, **bevor** jemand ausführt, und der Probelauf fragt mit denselben
// Wahlen wie der Knopf — Vorlage, Optionen, Karten. Wer tippt, löst nicht je
// Anschlag eine Auswertung über das ganze Set aus.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { VorlagenprobelaufAntwort } from '../antwortenprobelauf'
import { probelauffrist } from '../haken/useVorlagenprobelauf'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { stillerVorlagenprobelauf } from '../test/apiprobelauf'
import { SetVorlagenausfuhr } from './SetVorlagenausfuhr'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de
const eigene = texte.seiten.ausgeben.probelauf

const probeset = {
  id: '7',
  name: 'Mein Probeset',
  spiel: 'magic',
  stil: 'magic-m15',
  herkunft: '/pfad/probe.mse-set',
  karten: 2,
  herkunftFehlt: false,
}

const vorlage = {
  paket: 'magic-spoiler.mse-export-template',
  name: 'Spoiler',
  dateityp: '*.html|*.html',
  optionen: [
    {
      art: 'text',
      name: 'kuerzel',
      beschreibung: 'Kürzel',
      vorgabe: 'PRB',
      auswahl: [],
      wert: 'PRB',
    },
  ],
}

const andere = {
  paket: 'magic-liste.mse-export-template',
  name: 'Liste',
  dateityp: '*.txt|*.txt',
  optionen: [],
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <SetVorlagenausfuhr satz={probeset} imAssistenten />
    </Sprachraum>,
  )
}

function probe(teile: Partial<VorlagenprobelaufAntwort> = {}) {
  return Promise.resolve({ ...stillerVorlagenprobelauf, ...teile })
}

function belege(
  VorlageProbelauf: (...werte: unknown[]) => unknown,
  zusatz: Record<string, unknown> = {},
) {
  belegeApi({
    Kartenzeilen: () =>
      Promise.resolve({
        zeilen: [
          { id: '11', nummer: 0, name: 'Feuerball', felder: 2 },
          { id: '12', nummer: 1, name: 'Blitz', felder: 2 },
        ],
        fehler: null,
      }),
    Exportvorlagen: () =>
      Promise.resolve({ vorlagen: [vorlage, andere], befunde: [] }),
    VorlageProbelauf,
    ...zusatz,
  } as Parameters<typeof belegeApi>[0])
}

async function kasten() {
  return screen.findByRole('region', { name: eigene.ueberschrift })
}

describe('„Wie es wird" vor der Exportvorlage', () => {
  it('nennt Datei, Größe und Karten, bevor jemand ausführt', async () => {
    const VorlageAusfuehren = vi.fn()

    belege(
      () => probe({ dateiname: 'ausfuhr.html', bytes: 12582, karten: 2 }),
      {
        VorlageAusfuehren,
      },
    )

    zeige()

    const rahmen = await kasten()

    expect(
      await within(rahmen).findByText('Datei: ausfuhr.html (12,3 KB)'),
    ).toBeVisible()
    expect(within(rahmen).getByText('Ausgewertet über 2 Karten.')).toBeVisible()
    expect(within(rahmen).getByText(eigene.hinweis)).toBeVisible()
    expect(within(rahmen).queryByText(/Dateien daneben/)).toBeNull()

    // **Geschrieben ist noch nichts.**
    expect(VorlageAusfuehren).not.toHaveBeenCalled()
  })

  it('nennt die Dateien daneben, mit „mindestens", wo die Größe offen ist', async () => {
    belege(() =>
      probe({
        beigaben: 3,
        beigabenBytes: 2048,
        verzeichnis: 'ausfuhr-files',
        geschaetzt: true,
      }),
    )

    zeige()

    const rahmen = await kasten()

    expect(
      await within(rahmen).findByText(
        'Dateien daneben: 3 (Ordner „ausfuhr-files“, mindestens 2,0 KB)',
      ),
    ).toBeVisible()
  })

  it('nennt die Größe der Dateien daneben ohne Vorbehalt, wo sie feststeht', async () => {
    belege(() =>
      probe({
        beigaben: 1,
        beigabenBytes: 2048,
        verzeichnis: 'ausfuhr-files',
        geschaetzt: false,
      }),
    )

    zeige()

    expect(
      await within(await kasten()).findByText(
        'Dateien daneben: 1 (Ordner „ausfuhr-files“, 2,0 KB)',
      ),
    ).toBeVisible()
  })

  it('nennt, was bei der Auswertung aufgefallen ist', async () => {
    belege(() =>
      probe({
        befunde: ['die Vorlage führt keine Wahlmöglichkeit dieses Namens'],
      }),
    )

    zeige()

    const rahmen = await kasten()
    expect(
      await within(rahmen).findByText(eigene.vorlage.befunde),
    ).toBeVisible()
    expect(
      within(rahmen).getByText(
        'die Vorlage führt keine Wahlmöglichkeit dieses Namens',
      ),
    ).toBeVisible()
  })

  // **Ein Weg, zwei Ausgänge** — der Probelauf fragt mit den Wahlen des Knopfes.
  it('fragt mit derselben Vorlage, Wahl und Auswahl, mit denen der Knopf ausführt', async () => {
    const VorlageProbelauf = vi.fn(() => probe())
    const VorlageAusfuehren = vi.fn(() =>
      Promise.resolve({
        pfad: '/tmp/ausfuhr.html',
        befunde: [],
        abgebrochen: false,
      }),
    )

    belege(VorlageProbelauf, { VorlageAusfuehren })

    zeige()

    await within(await kasten()).findByText('Datei: ausfuhr.txt (120 B)')
    expect(VorlageProbelauf).toHaveBeenLastCalledWith(
      '7',
      vorlage.paket,
      { kuerzel: 'PRB' },
      [],
    )

    // Eine Option ändern und eine Teilmenge wählen.
    const feld = screen.getByRole('textbox', { name: 'Kürzel' })
    await userEvent.clear(feld)
    await userEvent.type(feld, 'XYZ')

    await userEvent.click(
      screen.getByRole('radio', { name: /Nur ausgewählte/ }),
    )

    await vi.waitFor(() => {
      expect(VorlageProbelauf).toHaveBeenLastCalledWith(
        '7',
        vorlage.paket,
        { kuerzel: 'XYZ' },
        ['11'],
      )
    })

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.sets.vorlageAusfuehren }),
    )

    // Die Wahl des Knopfes: Set, Paket, Dateityp, Wahl, Karten.
    expect(VorlageAusfuehren).toHaveBeenCalledWith(
      '7',
      vorlage.paket,
      vorlage.dateityp,
      { kuerzel: 'XYZ' },
      ['11'],
    )

    const [setID, paket, wahl, auswahl] = VorlageProbelauf.mock
      .lastCall as unknown as [string, string, Record<string, string>, string[]]
    const [ausSet, ausPaket, , ausWahl, ausAuswahl] = VorlageAusfuehren.mock
      .lastCall as unknown as [
      string,
      string,
      string,
      Record<string, string>,
      string[],
    ]

    expect([setID, paket, wahl, auswahl]).toEqual([
      ausSet,
      ausPaket,
      ausWahl,
      ausAuswahl,
    ])
  })

  it('fragt nach der anderen Vorlage, wenn sie gewählt wird', async () => {
    const VorlageProbelauf = vi.fn(() => probe())

    belege(VorlageProbelauf)

    zeige()

    await within(await kasten()).findByText('Datei: ausfuhr.txt (120 B)')

    await userEvent.selectOptions(
      screen.getByRole('combobox', { name: texte.seiten.sets.vorlageWaehlen }),
      andere.paket,
    )

    await vi.waitFor(() => {
      expect(VorlageProbelauf).toHaveBeenLastCalledWith(
        '7',
        andere.paket,
        {},
        [],
      )
    })
  })

  // **Wer tippt, löst nicht je Anschlag eine Auswertung aus.**
  it('wertet erst aus, wenn die Wahl zur Ruhe gekommen ist', async () => {
    const VorlageProbelauf = vi.fn(() => probe())

    belege(VorlageProbelauf)

    zeige()

    await within(await kasten()).findByText('Datei: ausfuhr.txt (120 B)')
    const davor = VorlageProbelauf.mock.calls.length

    const feld = screen.getByRole('textbox', { name: 'Kürzel' })
    await userEvent.clear(feld)
    await userEvent.type(feld, 'ABC')

    expect(VorlageProbelauf.mock.calls.length).toBe(davor)

    await vi.waitFor(() => {
      expect(VorlageProbelauf.mock.calls.length).toBe(davor + 1)
    })

    // Und es bleibt bei einer.
    await new Promise((fertig) => setTimeout(fertig, probelauffrist + 100))
    expect(VorlageProbelauf.mock.calls.length).toBe(davor + 1)
  })

  // **Bis die neue Antwort da ist, gehören die Zahlen zu einer früheren Wahl.**
  it('sagt neben den alten Zahlen, dass gerechnet wird, sobald die Wahl sich geändert hat', async () => {
    belege(() => probe())

    zeige()

    const rahmen = await kasten()
    await within(rahmen).findByText('Datei: ausfuhr.txt (120 B)')

    const lebendig = rahmen.querySelector('[aria-live]')
    expect(lebendig).toHaveAttribute('aria-busy', 'false')

    await userEvent.type(screen.getByRole('textbox', { name: 'Kürzel' }), 'X')

    // Schon in der Ruhefrist, vor der Frage.
    expect(lebendig).toHaveAttribute('aria-busy', 'true')
    expect(within(rahmen).getByText('Datei: ausfuhr.txt (120 B)')).toBeVisible()
    expect(within(rahmen).getByText(eigene.rechnet)).toBeVisible()

    await vi.waitFor(() => {
      expect(lebendig).toHaveAttribute('aria-busy', 'false')
    })
    expect(within(rahmen).queryByText(eigene.rechnet)).toBeNull()
  })

  it('verwirft die Antwort einer überholten Frage', async () => {
    let erste: (antwort: VorlagenprobelaufAntwort) => void = () => undefined

    const VorlageProbelauf = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise<VorlagenprobelaufAntwort>((fertig) => {
            erste = fertig
          }),
      )
      .mockImplementation(() => probe({ dateiname: 'zweite.txt', bytes: 5 }))

    belege(VorlageProbelauf)

    zeige()

    await userEvent.type(
      await screen.findByRole('textbox', { name: 'Kürzel' }),
      'X',
    )

    expect(await screen.findByText('Datei: zweite.txt (5 B)')).toBeVisible()

    erste({ ...stillerVorlagenprobelauf, dateiname: 'erste.txt', bytes: 9 })
    await new Promise((fertig) => setTimeout(fertig, 0))

    expect(screen.queryByText('Datei: erste.txt (9 B)')).toBeNull()
    expect(screen.getByText('Datei: zweite.txt (5 B)')).toBeVisible()
  })

  it('sagt, dass gerechnet wird, solange nichts da ist', async () => {
    belege(() => new Promise(() => undefined))

    zeige()

    expect(
      await within(await kasten()).findByText(eigene.rechnet),
    ).toBeVisible()
  })

  // **Derselbe Satz, den der Knopf gäbe** — am verursachenden Ort (§6).
  it('zeigt den Fehler im Rahmen und lässt den Knopf stehen', async () => {
    belege(() =>
      Promise.resolve({
        ...stillerVorlagenprobelauf,
        fehler: {
          code: 'validierung',
          schluessel: 'validierung',
          meldung: 'die Vorlage lief nicht',
          feld: '',
          werte: {},
        },
      }),
    )

    zeige()

    const rahmen = await kasten()
    expect(await within(rahmen).findByRole('alert')).toHaveTextContent(
      /nicht gültig/,
    )
    expect(within(rahmen).queryByText(/ausfuhr/)).toBeNull()

    expect(
      screen.getByRole('button', { name: texte.seiten.sets.vorlageAusfuehren }),
    ).toBeEnabled()
  })
})
