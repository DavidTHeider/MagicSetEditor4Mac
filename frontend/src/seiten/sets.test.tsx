// Die Setliste.
//
// Hieß `seiten.test.tsx`, solange sie die Karten mitführte. Der Name sagt
// jetzt, was drinsteht.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { SetDTO, SetsAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { stillerSpeicherlauf } from '../test/speicherlauf'
import { belegeApi, raeumeApi } from '../test/api'
import { Sets } from './Sets'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function probeSet(teile: Partial<SetDTO> = {}): SetDTO {
  return {
    id: '1',
    name: 'Mein Probeset',
    spiel: 'magic',
    stil: 'm15',
    herkunft: '/pfad/probeset.mse-set',
    karten: 2,
    herkunftFehlt: false,
    ...teile,
  }
}

// saubereVorschau ist eine Datei ohne Befund und ohne fehlendes Paket — sie
// wird ohne weiteren Klick eingelesen (FMT-03, #1692).
function saubereVorschau(pfad = '/tmp/probe.mse-set') {
  return () =>
    Promise.resolve({
      abgebrochen: false,
      vorschau: {
        pfad,
        name: 'Probe',
        spiel: 'magic',
        stil: 'm15',
        karten: 1,
        befunde: [],
        schonOffen: '',
        bedarf: [],
      },
    })
}

function zeigeSets(waehlen = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Sets
        wechseln={vi.fn()}
        waehlen={waehlen}
        speichern={stillerSpeicherlauf()}
      />
    </Sprachraum>,
  )

  return waehlen
}

describe('Setliste', () => {
  // **Eine Fläche je Set und keine Tabellenzeile** seit
  // #288: die
  // Tabelle hatte sieben Spalten, vier davon Knöpfe, und die brachen über fünf
  // Zeilen um. Eine Tabelle ist zum Vergleichen da; hier gibt es etwas zu tun.
  it('zeigt die Sets mit ihren Angaben', async () => {
    belegeApi({ Sets: () => Promise.resolve({ sets: [probeSet()] }) })
    zeigeSets()

    const eintrag = await screen.findByRole('listitem')

    // **Spiel, Kartenzahl und Stil in einer Zeile** — die Kennzahlen, nach
    // denen jemand ein Set wiedererkennt.
    expect(within(eintrag).getByText(/^magic · 2 /)).toBeInTheDocument()
    expect(
      within(eintrag).getByText('/pfad/probeset.mse-set'),
    ).toBeInTheDocument()
  })

  // **Der Name ist die Überschrift des Eintrags.** Wer mit einem
  // Vorleseprogramm über die Überschriften springt, kommt so von Set zu Set.
  it('macht den Namen zur Überschrift', async () => {
    belegeApi({ Sets: () => Promise.resolve({ sets: [probeSet()] }) })
    zeigeSets()

    // **Der Name steht im Knopf, und der Knopf trägt die Absicht.** Die
    // Überschrift heißt deshalb „Mein Probeset — Karten dieses Sets zeigen":
    // wer sie hört, weiß auch, was ein Druck bewirkt.
    const ueberschrift = await screen.findByRole('heading', {
      level: 2,
      name: /Mein Probeset/,
    })

    expect(ueberschrift).toHaveTextContent('Mein Probeset')
  })

  // **Der leere Zustand erklärt den nächsten Schritt** (§6), statt „Keine
  // Daten" zu zeigen.
  it('erklärt den nächsten Schritt, wenn kein Set da ist', async () => {
    belegeApi({ Sets: () => Promise.resolve({ sets: [] }) })
    zeigeSets()

    await waitFor(() => {
      expect(screen.getByText(texte.seiten.sets.leer)).toBeInTheDocument()
    })

    expect(screen.queryByRole('listitem')).not.toBeInTheDocument()
  })

  it('meldet einen Fehler, statt leer zu bleiben', async () => {
    belegeApi({ Sets: () => Promise.reject(new Error('Bestand nicht lesbar')) })
    zeigeSets()

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(
        'Bestand nicht lesbar',
      )
    })
  })

  it('sagt, dass geladen wird', () => {
    belegeApi({ Sets: () => new Promise<SetsAntwort>(() => undefined) })
    zeigeSets()

    // **Die Liste sagt es** (UX-05, #1677). Das Anlegen steht, solange sie
    // lädt, noch nicht da — ob offen oder in einer Schublade, entscheidet
    // erst die Liste (AP-10, `Setanlage`).
    expect(
      screen.queryByRole('region', { name: texte.seiten.sets.neuTitel }),
    ).toBeNull()
    const liste = screen.getAllByRole('status')

    expect(liste).toHaveLength(1)
    expect(liste[0]).toHaveTextContent(texte.bausteine.zustaende.laedt)
  })

  it('gibt die Kennung weiter, wenn jemand ein Set wählt', async () => {
    belegeApi({
      Sets: () => Promise.resolve({ sets: [probeSet({ id: '42' })] }),
    })
    const waehlen = zeigeSets()

    // **Genau der Namensknopf**, nicht irgendeiner, der den Setnamen führt:
    // die Zeile hat seit der Ausfuhr eine zweite Schaltfläche, deren
    // Beschriftung ihn ebenfalls nennt.
    await userEvent.click(
      await screen.findByRole('button', {
        name: `Mein Probeset — ${texte.seiten.sets.oeffnen}`,
      }),
    )

    expect(waehlen).toHaveBeenCalledWith('42')
  })

  // **Ein Set lässt sich öffnen** — über den nativen Dialog, der in Go liegt.
  //
  // Die Seite kennt keinen Pfad: sie ruft, und was gewählt wird, entscheidet
  // das Fenster. Ein getipptes Feld wäre die schlechtere Hälfte dessen, was
  // macOS ohnehin kann (§6).
  it('öffnet ein Set und lädt die Liste nach', async () => {
    // **Nicht `Sets` nennen.** Der erste Anlauf hieß so — und `<Sets wechseln={vi.fn()} …/>`
    // zeichnete dann die Attrappe statt der Komponente. Der Bildschirm blieb
    // leer, und der Test suchte eine Schaltfläche, die es nie gab.
    const abfrage = vi
      .fn()
      .mockResolvedValueOnce({ sets: [] })
      .mockResolvedValue({ sets: [probeSet({ name: 'Frisch geöffnet' })] })

    const SetAusDatei = vi.fn().mockResolvedValue({
      setID: '1',
      karten: 312,
      befunde: [],
      schonOffen: false,
      angelegt: [],
      unvollstaendig: false,
      abgebrochen: false,
    })

    belegeApi({
      Sets: abfrage,
      SetVorschauWaehlen: saubereVorschau('/tmp/frisch.mse-set'),
      SetAusDatei,
    })

    render(
      <Sprachraum sprache="de">
        <Sets
          wechseln={vi.fn()}
          waehlen={vi.fn()}
          speichern={stillerSpeicherlauf()}
        />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.seiten.sets.setOeffnen }),
    )

    // **Eine saubere Datei wird gleich eingelesen** — ohne Vorschau dazwischen.
    expect(SetAusDatei).toHaveBeenCalledWith('/tmp/frisch.mse-set')

    // Die Liste zeigt danach das neue Set — nachgeladen, nicht behauptet.
    expect(await screen.findByText('Frisch geöffnet')).toBeInTheDocument()
  })

  // **Die Vorbehalte stehen beim Erfolg** (§8).
  it('nennt beim Öffnen, was nicht gedeutet wurde', async () => {
    belegeApi({
      Sets: () => Promise.resolve({ sets: [] }),
      SetVorschauWaehlen: saubereVorschau(),
      SetAusDatei: () =>
        Promise.resolve({
          setID: '1',
          karten: 300,
          befunde: ['set:12: unlesbare Karte'],
          schonOffen: false,
          angelegt: [],
          unvollstaendig: false,
          abgebrochen: false,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Sets
          wechseln={vi.fn()}
          waehlen={vi.fn()}
          speichern={stillerSpeicherlauf()}
        />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.seiten.sets.setOeffnen }),
    )

    expect(
      await screen.findByText(texte.seiten.sets.beimOeffnenBefunde),
    ).toBeInTheDocument()
    expect(screen.getByText('set:12: unlesbare Karte')).toBeInTheDocument()
  })

  // Dieselbe Datei ein zweites Mal: gesagt, nicht doppelt angelegt.
  it('sagt, wenn die Datei schon offen ist', async () => {
    belegeApi({
      Sets: () => Promise.resolve({ sets: [probeSet()] }),
      SetVorschauWaehlen: saubereVorschau(),
      SetAusDatei: () =>
        Promise.resolve({
          setID: '1',
          karten: 12,
          befunde: [],
          schonOffen: true,
          angelegt: [],
          unvollstaendig: false,
          abgebrochen: false,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Sets
          wechseln={vi.fn()}
          waehlen={vi.fn()}
          speichern={stillerSpeicherlauf()}
        />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.seiten.sets.setOeffnen }),
    )

    expect(
      await screen.findByText(texte.seiten.sets.schonOffen),
    ).toBeInTheDocument()
  })

  // **Ein Abbruch ist keine Meldung wert** — und schon gar keine Fehlermeldung.
  it('sagt nichts, wenn niemand eine Datei wählt', async () => {
    belegeApi({
      Sets: () => Promise.resolve({ sets: [] }),
      SetVorschauWaehlen: () =>
        Promise.resolve({ abgebrochen: true, vorschau: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Sets
          wechseln={vi.fn()}
          waehlen={vi.fn()}
          speichern={stillerSpeicherlauf()}
        />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.seiten.sets.setOeffnen }),
    )

    await waitFor(() => {
      expect(screen.getByText(texte.seiten.sets.leer)).toBeInTheDocument()
    })

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('meldet einen Fehler beim Öffnen', async () => {
    belegeApi({
      Sets: () => Promise.resolve({ sets: [] }),
      // **Die Vorschau weist ab, was das Einlesen abwiese** — mit derselben
      // Meldung (FMT-03).
      SetVorschauWaehlen: () =>
        Promise.resolve({
          vorschau: null,
          abgebrochen: false,
          fehler: {
            code: 'validierung' as const,
            schluessel: 'validierung',
            meldung: 'Das ist kein Set.',
          },
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Sets
          wechseln={vi.fn()}
          waehlen={vi.fn()}
          speichern={stillerSpeicherlauf()}
        />
      </Sprachraum>,
    )

    await userEvent.click(
      await screen.findByRole('button', { name: texte.seiten.sets.setOeffnen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      // **Der Satz aus dem Katalog** (#133): übersetzt wird über den
      // Schlüssel, nicht über den Satz aus Go.
      kataloge.de.fehler.validierung,
    )
  })
})
