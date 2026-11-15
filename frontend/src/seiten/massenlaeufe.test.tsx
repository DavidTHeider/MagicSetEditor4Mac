// **Was die Massenänderung bei nebenläufigen Handgriffen tut** — die vier
// Regeln aus §7.
//
// **Eine eigene Datei**, damit `massenaenderung.test.tsx` unter der
// Zeilengrenze aus §2 bleibt. Der Schnitt folgt der Frage: dort steht, was ein
// einzelner Handgriff zusagt, hier, was zwei einander antun.
//
// Getrennt am 2026-09-02.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { MassenvorschauAntwort } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { antwortAbwarten, belegeApi, raeumeApi } from '../test/api'
import { Massenaenderung } from './Massenaenderung'
import {
  feldwerte,
  fuelleAus,
  karten,
  mitVorschau,
  texte,
  vorschauAntwort,
  waehleFeld,
  zeige,
} from './massenprobe'

describe('Massenänderung, nebenläufig', () => {
  afterEach(() => {
    raeumeApi()
  })

  // **Ein Vorgang je Datensatz** (§7, Regel 4): der zweite Klick geht nicht
  // gegen einen Stand, den das Backend noch nicht kennt.
  it('lässt einen zweiten Anlauf nicht los, solange der erste läuft', async () => {
    let loesen = (): void => undefined

    const FelderSetzen = vi.fn(
      () =>
        new Promise<{ betroffen: string[] }>((erfuelle) => {
          loesen = () => {
            erfuelle({ betroffen: [] })
          }
        }),
    )

    belegeApi({ ...mitVorschau, FelderSetzenMitErwartung: FelderSetzen })
    zeige()

    await fuelleAus()

    // Die sichtbare Hälfte der Zusage: die Schaltfläche reagiert nicht.
    const knopf = screen.getByRole('button', { name: texte.anwenden })
    await waitFor(() => {
      expect(knopf).toBeDisabled()
    })

    // Und die unsichtbare: auch ein Aufruf, der an der Schaltfläche vorbeikäme,
    // wird angehalten — dafür steht `useVorgang`.
    await userEvent.click(knopf)

    loesen()

    // **Nach dem Anwenden ist die Vorschau weg**, und damit gibt es nichts
    // mehr anzuwenden — frei ist der Vorschauknopf.
    await waitFor(() => {
      expect(screen.getByRole('button', { name: texte.vorschau })).toBeEnabled()
    })

    expect(FelderSetzen).toHaveBeenCalledOnce()
  })

  // **Eine überholte Antwort trägt sich nicht ein** (§7, Regel 2).
  //
  // Wer die Auswahl wechselt, während ein Schreiben läuft, bekam vorher den
  // Fehler der **alten** Auswahl an die neue gehängt — und die vorweggenommene
  // Anzeige gleich mit.
  // **Die Vorschau einer überholten Auswahl fällt** — geglückt wie
  // gescheitert (§7, Regel 2): weder Tafel noch Meldung gehören zu einer
  // Auswahl, die es nicht mehr gibt.
  it('lässt eine geglückte Vorschau einer überholten Auswahl fallen', async () => {
    let antworten: (antwort: MassenvorschauAntwort) => void = () => undefined
    belegeApi({
      ...feldwerte,
      FelderVorschau: () =>
        new Promise<MassenvorschauAntwort>((los) => {
          antworten = los
        }),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Massenaenderung setID="1" karten={karten} nachladen={vi.fn()} />
      </Sprachraum>,
    )

    await waehleFeld('typ')
    await userEvent.type(screen.getByLabelText(texte.wert), 'Artefakt')
    await userEvent.click(screen.getByRole('button', { name: texte.vorschau }))
    rerender(
      <Sprachraum sprache="de">
        <Massenaenderung
          setID="1"
          karten={karten.slice(0, 1)}
          nachladen={vi.fn()}
        />
      </Sprachraum>,
    )

    antworten(vorschauAntwort())

    await waitFor(() => {
      expect(screen.getAllByRole('listitem')).toHaveLength(1)
    })
    await antwortAbwarten()

    expect(screen.queryByRole('table')).not.toBeInTheDocument()
  })

  it('lässt eine gescheiterte Vorschau einer überholten Auswahl fallen', async () => {
    let ablehnen: (fehler: unknown) => void = () => undefined
    belegeApi({
      ...feldwerte,
      FelderVorschau: () =>
        new Promise<MassenvorschauAntwort>((_, ab) => {
          ablehnen = ab
        }),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Massenaenderung setID="1" karten={karten} nachladen={vi.fn()} />
      </Sprachraum>,
    )

    await waehleFeld('typ')
    await userEvent.type(screen.getByLabelText(texte.wert), 'Artefakt')
    await userEvent.click(screen.getByRole('button', { name: texte.vorschau }))
    rerender(
      <Sprachraum sprache="de">
        <Massenaenderung
          setID="1"
          karten={karten.slice(0, 1)}
          nachladen={vi.fn()}
        />
      </Sprachraum>,
    )

    ablehnen(new Error('Die alte Auswahl ist gescheitert.'))

    await waitFor(() => {
      expect(screen.getAllByRole('listitem')).toHaveLength(1)
    })
    await antwortAbwarten()

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('lässt die Antwort einer überholten Auswahl fallen', async () => {
    let ablehnen: (fehler: unknown) => void = () => undefined

    belegeApi({
      ...mitVorschau,
      FelderSetzenMitErwartung: () =>
        new Promise((_, ab) => {
          ablehnen = ab
        }),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Massenaenderung setID="1" karten={karten} nachladen={vi.fn()} />
      </Sprachraum>,
    )

    await fuelleAus()

    // Die Auswahl wechselt, während der erste Aufruf noch läuft.
    rerender(
      <Sprachraum sprache="de">
        <Massenaenderung
          setID="1"
          karten={karten.slice(0, 1)}
          nachladen={vi.fn()}
        />
      </Sprachraum>,
    )

    ablehnen(new Error('Die alte Auswahl ist gescheitert.'))

    await waitFor(() => {
      expect(screen.getAllByRole('listitem')).toHaveLength(1)
    })

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  // Auch ein **geglücktes** Schreiben der alten Auswahl lädt nicht nach.
  //
  // Nachladen heißt: die Liste, die gerade dasteht, neu holen. Für eine
  // Auswahl, die niemand mehr offen hat, ist das keine Bestenfalls-Arbeit,
  // sondern eine Bewegung ohne Anlass.
  it('lädt nach einer überholten, geglückten Antwort nicht nach', async () => {
    let erfuellen: (antwort: { betroffen: string[] }) => void = () => undefined

    belegeApi({
      ...mitVorschau,
      FelderSetzenMitErwartung: () =>
        new Promise((erfuelle) => {
          erfuellen = erfuelle
        }),
    })

    const nachladen = vi.fn()

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Massenaenderung setID="1" karten={karten} nachladen={nachladen} />
      </Sprachraum>,
    )

    await fuelleAus()

    rerender(
      <Sprachraum sprache="de">
        <Massenaenderung
          setID="1"
          karten={karten.slice(0, 1)}
          nachladen={nachladen}
        />
      </Sprachraum>,
    )

    erfuellen({ betroffen: [] })

    await waitFor(() => {
      expect(screen.getAllByRole('listitem')).toHaveLength(1)
    })

    expect(nachladen).not.toHaveBeenCalled()
  })

  // **Der ABA-Fall**: weg von der Auswahl und zurück.
  //
  // Der Schlüssel ist dann wieder derselbe — und ohne laufende Nummer trüge
  // sich die Antwort des ersten, längst überholten Laufs in den neuen ein. Der
  // Nachlauf des Zweitblicks hat genau das angestrichen.
  it('lässt eine Antwort fallen, wenn dieselbe Auswahl inzwischen neu ist', async () => {
    let ablehnen: (fehler: unknown) => void = () => undefined

    belegeApi({
      ...mitVorschau,
      FelderSetzenMitErwartung: () =>
        new Promise((_, ab) => {
          ablehnen = ab
        }),
    })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Massenaenderung setID="1" karten={karten} nachladen={vi.fn()} />
      </Sprachraum>,
    )

    await fuelleAus()

    // Weg von der Auswahl …
    rerender(
      <Sprachraum sprache="de">
        <Massenaenderung
          setID="1"
          karten={karten.slice(0, 1)}
          nachladen={vi.fn()}
        />
      </Sprachraum>,
    )

    // … und wieder zurück: derselbe Schlüssel, ein anderer Stand.
    rerender(
      <Sprachraum sprache="de">
        <Massenaenderung setID="1" karten={karten} nachladen={vi.fn()} />
      </Sprachraum>,
    )

    ablehnen(new Error('Die alte Auswahl ist gescheitert.'))

    await waitFor(() => {
      expect(screen.getAllByRole('listitem')).toHaveLength(2)
    })

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})
