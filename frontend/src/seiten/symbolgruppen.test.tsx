// **Mehrfachwahl, Gruppen und Reihenfolge** — Werkstück 5 von Phase 3.7
// (#1168,
// Befund Z5-02).
//
// # Was hier geprüft wird
//
// **Die Naht und die Bedienung** — nicht die Regeln. Wo die Gruppe landet und
// in welcher Reihenfolge ihre Kinder stehen, entscheidet die Go-Seite
// (`internal/service/symbolteile.go`); geprüft ist es dort. Hier steht: kommt
// die Wahl an, kommt die Antwort an, und stehen die Knöpfe offen, wenn die
// Handlung möglich ist.
//
// **Die Malreihenfolge steht nebenan** in `symbolflaechenordnung.test.tsx` —
// zusammen ginge die Datei über die Zeilengrenze (§2).
import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { raeumeApi } from '../test/api'
import {
  beideWaehlen,
  mitSymbol,
  teil,
  teilknopf,
  zeige,
} from './symbolgruppenprobe'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.karten.symbol

describe('Die Mehrfachwahl in der Teilliste', () => {
  // **Ohne zweites gewähltes Teil steht der Knopf still.**
  //
  // Dieselbe Auskunft wie MSEs ausgegrautem Menüeintrag
  // (`select_editor.cpp`, Zeile 156, Fassung `f9d9356` vom 2021-07-06).
  it('hält den Gruppenknopf an, solange nur eines gewählt ist', async () => {
    mitSymbol()
    zeige()

    expect(
      await screen.findByRole('button', { name: texte.gruppieren }),
    ).toBeDisabled()
  })

  // **Mit ⌘-Klick kommt das zweite hinzu, und der Knopf öffnet.**
  it('nimmt mit gedrückter Befehlstaste hinzu', async () => {
    mitSymbol()
    zeige()

    await beideWaehlen()

    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: texte.gruppieren }),
      ).toBeEnabled()
    })
  })

  // **Und der Klick schickt beide Stellen an die Go-Seite.**
  it('gruppiert die gewählten Stellen', async () => {
    const TeileGruppieren = vi.fn().mockResolvedValue({
      teile: [teil('Gruppe', 'Gruppe', [teil('Dreieck'), teil('Rechteck')])],
      name: '',
      fehler: null,
    })

    mitSymbol({ TeileGruppieren })
    zeige()

    await beideWaehlen()

    await userEvent.click(
      screen.getByRole('button', { name: texte.gruppieren }),
    )

    await waitFor(() => {
      expect(TeileGruppieren).toHaveBeenCalledWith(expect.anything(), [0, 1])
    })

    // **Die Gruppe steht in der Liste, mit der Zahl ihrer Kinder.**
    expect(
      await screen.findByRole('button', { name: 'Gruppe — 2 Teile' }),
    ).toBeInTheDocument()

    // **Und ihre Kinder stehen eingerückt darunter — nicht anklickbar.**
    expect(screen.getByText('Dreieck, in einer Gruppe')).toBeInTheDocument()
  })

  // **Eine Gruppe lässt sich auflösen, eine Form nicht.**
  it('öffnet den Auflösenknopf nur über einer Gruppe', async () => {
    mitSymbol({
      Symbol: vi.fn().mockResolvedValue({
        teile: [teil('Gruppe', 'Gruppe', [teil('Dreieck')]), teil('Rechteck')],
        name: 'symbol1.mse-symbol',
        fehler: null,
      }),
      GruppeAufloesen: vi.fn().mockResolvedValue({
        teile: [teil('Dreieck'), teil('Rechteck')],
        name: '',
        fehler: null,
      }),
    })

    zeige()

    // Die Gruppe steht an Stelle 0 und ist von Anfang an gewählt.
    expect(
      await screen.findByRole('button', { name: texte.aufloesen }),
    ).toBeEnabled()

    await userEvent.click(teilknopf('Rechteck'))

    await waitFor(() => {
      expect(
        screen.getByRole('button', { name: texte.aufloesen }),
      ).toBeDisabled()
    })
  })

  // **Und der Klick gibt die Kinder frei.**
  it('löst die gewählte Gruppe auf', async () => {
    const GruppeAufloesen = vi.fn().mockResolvedValue({
      teile: [teil('Dreieck'), teil('Rechteck')],
      name: '',
      fehler: null,
    })

    mitSymbol({
      Symbol: vi.fn().mockResolvedValue({
        teile: [teil('Gruppe', 'Gruppe', [teil('Dreieck')]), teil('Rechteck')],
        name: 'symbol1.mse-symbol',
        fehler: null,
      }),
      GruppeAufloesen,
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: texte.aufloesen }),
    )

    await waitFor(() => {
      expect(GruppeAufloesen).toHaveBeenCalledWith(expect.anything(), [0])
    })

    expect(
      await screen.findByRole('button', { name: 'Dreieck — 3 Punkte' }),
    ).toBeInTheDocument()
  })

  // **Nach oben und nach unten schicken die Stellen an die Go-Seite.**
  //
  // **Die Rechnung steht dort** — die Zielstelle zählt ohne das bewegte Teil,
  // und das ist MSEs Umrechnung (`part_list.cpp`, Zeile 158, Fassung
  // `f9d9356` vom 2021-07-06). Hier steht nur, welche zwei Zahlen hinausgehen.
  it('schickt eine Verschiebung mit beiden Stellen', async () => {
    const TeilVerschieben = vi.fn().mockResolvedValue({
      teile: [teil('Rechteck'), teil('Dreieck')],
      name: '',
      fehler: null,
    })

    mitSymbol({ TeilVerschieben })
    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: 'Dreieck nach unten' }),
    )

    await waitFor(() => {
      expect(TeilVerschieben).toHaveBeenCalledWith(expect.anything(), 0, 1)
    })

    // **Und nach oben zählt rückwärts** — dieselbe Naht, andere Richtung.
    //
    // Die Antwort hat die beiden getauscht: das Dreieck steht jetzt an
    // Stelle 1, und sein Knopf schickt (1, 0).
    await userEvent.click(
      await screen.findByRole('button', { name: 'Dreieck nach oben' }),
    )

    await waitFor(() => {
      expect(TeilVerschieben).toHaveBeenCalledWith(expect.anything(), 1, 0)
    })
  })

  // **Das erste Teil kann nicht weiter nach oben, das letzte nicht weiter
  // nach unten.**
  it('hält die Enden der Liste an', async () => {
    mitSymbol()
    zeige()

    expect(
      await screen.findByRole('button', { name: 'Dreieck nach oben' }),
    ).toBeDisabled()

    expect(
      screen.getByRole('button', { name: 'Rechteck nach oben' }),
    ).toBeEnabled()

    expect(
      screen.getByRole('button', { name: 'Dreieck nach unten' }),
    ).toBeEnabled()

    expect(
      screen.getByRole('button', { name: 'Rechteck nach unten' }),
    ).toBeDisabled()
  })

  // **Ein Fehler räumt die Zeichnung nicht** (§7, Regel 3).
  it('meldet einen Fehler und behält die Teile', async () => {
    mitSymbol({
      TeileGruppieren: vi.fn().mockResolvedValue({
        teile: [teil('Dreieck'), teil('Rechteck')],
        name: '',
        fehler: { code: 'validation', meldung: 'geht nicht', feld: '' },
      }),
    })

    zeige()

    await beideWaehlen()

    await userEvent.click(
      screen.getByRole('button', { name: texte.gruppieren }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent('geht nicht')
    expect(teilknopf('Dreieck')).toBeInTheDocument()
    expect(teilknopf('Rechteck', 4)).toBeInTheDocument()
  })
})

// **Eine gemischte Wahl** — Befund 7 des Zweitblicks über Phase 3.7.
describe('Eine Gruppe neben einem anderen Teil', () => {
  // **Eine Gruppe in der Wahl genügt.**
  //
  // MSE schaltet den Eintrag frei, sobald **mindestens eine** gewöhnliche
  // Gruppe gewählt ist (`select_editor.cpp`, Zeilen 157–164, Fassung
  // `f9d9356` vom 2021-07-06); beim Ausführen bleibt stehen, was keine ist.
  // Die Go-Seite trägt das seit Werkstück 5 — die Oberfläche verlangte
  // `every` und machte den geprüften Weg **unerreichbar**.
  it('lässt sich auflösen und nimmt die Form mit', async () => {
    const GruppeAufloesen = vi.fn().mockResolvedValue({
      teile: [teil('Dreieck'), teil('Rechteck', 'Form', [], 4)],
      name: '',
      fehler: null,
    })

    mitSymbol({
      Symbol: vi.fn().mockResolvedValue({
        teile: [
          teil('Gruppe', 'Gruppe', [teil('Dreieck')]),
          teil('Rechteck', 'Form', [], 4),
        ],
        name: 'symbol1.mse-symbol',
        fehler: null,
      }),
      GruppeAufloesen,
    })

    zeige()

    const nutzer = userEvent.setup()

    await screen.findByRole('button', { name: 'Gruppe — 1 Teile' })

    await nutzer.keyboard('{Meta>}')
    await nutzer.click(teilknopf('Rechteck', 4))
    await nutzer.keyboard('{/Meta}')

    const knopf = screen.getByRole('button', { name: texte.aufloesen })
    expect(knopf).toBeEnabled()

    await nutzer.click(knopf)

    await waitFor(() => {
      expect(GruppeAufloesen).toHaveBeenCalledWith(expect.anything(), [0, 1])
    })
  })
})
