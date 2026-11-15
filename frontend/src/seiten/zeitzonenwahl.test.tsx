// **Die Zeitzone ist eine Auswahl mit Suche** — UXR-F1
// (#2007),
// R7 aus `docs/UIREVIEW.md`: *„Zeitzone als Freitext"*.
//
// # Was geprüft wird
//
// Die Einstellung hat **keine Werteliste** in Go — `domain.PruefeZeitzone`
// prüft gegen die Zonendatenbank und nennt nichts. Das Fenster bietet deshalb
// seine eigene Liste an (`sprache/zeitzonen.json`), und dass Go jede Zeile
// annimmt, hält `internal/werkzeuge/zeitzonenliste_test.go`. Hier steht die
// andere Hälfte: **was das Fenster mit der Liste tut**.
//
//   - die Einstellung ist ein Kombinationsfeld, kein Textfeld;
//   - leer heißt „wie das System", und es ist die erste Zeile;
//   - ein Wert, der nicht in der Liste steht, **bleibt sichtbar gewählt**;
//   - gewählt wird über die rohe Eingabe, und der Schreibweg bekommt
//     Schlüssel und Namen;
//   - die übrigen Textfelder ohne Werteliste bleiben Textfelder.
import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { raeumeApi } from '../test/api'
import { verstoesse } from '../test/axe'
import { zeigeEinstellungen } from '../test/einstellungsseite'
import zonen from '../sprache/zeitzonen.json'
import { zeilengrenze } from '../bausteine/suchwahl'

import type { EinstellungDTO } from '../api'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.einstellungen
const system = kataloge.de.bausteine.zeitzonenwahl.system

function zeitzone(wert: string): EinstellungDTO {
  return {
    schluessel: 'zeitzone',
    wert,
    wirksam: wert,
    werte: [],
    bereich: 'anwendung',
  }
}

const darstellung: EinstellungDTO = {
  schluessel: 'darstellung',
  wert: '',
  wirksam: '',
  werte: [],
  bereich: 'anwendung',
}

const feld = () => screen.getByRole('combobox', { name: eigene.namen.zeitzone })

describe('Die Zeitzone in den Einstellungen', () => {
  it('ist eine Auswahl mit Suche und kein Textfeld', () => {
    zeigeEinstellungen({ eintraege: [zeitzone('')] })

    expect(feld()).toHaveAttribute('data-einstellung', 'zeitzone')
    expect(
      screen.queryByRole('textbox', { name: eigene.namen.zeitzone }),
    ).toBeNull()
  })

  // **Leer heißt „wie das System"** — mit einem Namen im Feld, nicht leer.
  it('nennt den leeren Wert „wie das System"', () => {
    zeigeEinstellungen({ eintraege: [zeitzone('')] })

    expect(feld()).toHaveValue(system)
  })

  it('zeigt die gewählte Zone', () => {
    zeigeEinstellungen({ eintraege: [zeitzone('Europe/Berlin')] })

    expect(feld()).toHaveValue('Europe/Berlin')
  })

  // **Ein Name, den Go annimmt und die Liste nicht kennt** — `US/Pacific` ist
  // ein alter Name, den die Zonendatenbank noch führt. Er steht im Feld, wie
  // er gespeichert ist.
  it('lässt einen Wert, der nicht in der Liste steht, sichtbar gewählt', () => {
    zeigeEinstellungen({ eintraege: [zeitzone('US/Pacific')] })

    expect(zonen).not.toContain('US/Pacific')
    expect(feld()).toHaveValue('US/Pacific')
  })

  // **Zeichnen tut sie die ersten zweihundert** (`zeilengrenze`) — die Liste
  // hat 420 Zeilen, und jede ist ein Knoten. Gesucht wird in allen.
  it('bietet „wie das System" und die ersten Zonen an und sagt, dass es mehr gibt', async () => {
    zeigeEinstellungen({ eintraege: [zeitzone('')] })

    await userEvent.click(feld())

    const zeilen = screen
      .getAllByRole('option')
      .map((zeile) => zeile.textContent)

    expect(zeilen).toHaveLength(zeilengrenze)
    expect(zeilen[0]).toBe(system)
    expect(zeilen.slice(1)).toEqual(zonen.slice(0, zeilengrenze - 1))
    expect(
      screen.getAllByText(
        kataloge.de.bausteine.suchauswahl.begrenzt
          .replace('{gezeigt}', String(zeilengrenze))
          .replace('{anzahl}', String(zonen.length + 1)),
      ),
    ).toHaveLength(2)
  })

  // **Die letzte Zone steht hinter der Grenze** und lässt sich trotzdem
  // finden und schreiben.
  it('findet eine Zone hinter der Grenze', async () => {
    const setze = zeigeEinstellungen({ eintraege: [zeitzone('')] })

    expect(zonen.indexOf('UTC')).toBeGreaterThan(zeilengrenze)

    await userEvent.click(feld())
    await userEvent.keyboard('utc{Enter}')

    expect(setze).toHaveBeenCalledExactlyOnceWith('zeitzone', 'UTC')
  })

  it('findet eine Zone über ein Teilwort, auch mit Leerzeichen', async () => {
    zeigeEinstellungen({ eintraege: [zeitzone('')] })

    await userEvent.click(feld())
    await userEvent.keyboard('new york')

    expect(
      screen.getAllByRole('option').map((zeile) => zeile.textContent),
    ).toEqual(['America/New_York'])
  })

  it('schreibt die gewählte Zone unter ihrem Schlüssel', async () => {
    const setze = zeigeEinstellungen({ eintraege: [zeitzone('')] })

    await userEvent.click(feld())
    await userEvent.keyboard('berlin{Enter}')

    expect(setze).toHaveBeenCalledExactlyOnceWith('zeitzone', 'Europe/Berlin')
  })

  // **Der Weg zurück**: ohne die Zeile „wie das System" gäbe es für eine
  // gesetzte Zone keinen.
  it('schreibt den leeren Wert, wenn „wie das System" gewählt wird', async () => {
    const setze = zeigeEinstellungen({ eintraege: [zeitzone('Europe/Berlin')] })

    await userEvent.click(feld())
    await userEvent.click(screen.getByRole('option', { name: system }))

    expect(setze).toHaveBeenCalledExactlyOnceWith('zeitzone', '')
  })

  it('schreibt nichts, wenn die Eingabe verworfen wird', async () => {
    const setze = zeigeEinstellungen({ eintraege: [zeitzone('Europe/Berlin')] })

    await userEvent.click(feld())
    await userEvent.keyboard('paris{Escape}')

    expect(setze).not.toHaveBeenCalled()
    expect(feld()).toHaveValue('Europe/Berlin')
  })

  // **Ein Vorgang je Datensatz** (§7, Regel 4): läuft das Schreiben, ist das
  // Feld gesperrt.
  it('ist gesperrt, solange das Schreiben läuft', () => {
    zeigeEinstellungen({
      eintraege: [zeitzone('')],
      laeuft: (schluessel) => schluessel === 'zeitzone',
    })

    expect(feld()).toBeDisabled()
  })

  it('lässt die übrigen Einstellungen ohne Werteliste Textfelder bleiben', () => {
    zeigeEinstellungen({ eintraege: [zeitzone(''), darstellung] })

    expect(
      screen.getByRole('textbox', { name: eigene.namen.darstellung }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('combobox', { name: eigene.namen.darstellung }),
    ).toBeNull()
  })

  it('hält die Regeln der Barrierefreiheit, auch geöffnet', async () => {
    zeigeEinstellungen({ eintraege: [zeitzone('Europe/Berlin')] })

    expect(await verstoesse(document.body)).toEqual([])

    await userEvent.click(feld())

    expect(await verstoesse(document.body)).toEqual([])
  })
})
