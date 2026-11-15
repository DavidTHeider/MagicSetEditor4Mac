// **Suchen und Ersetzen über ein Set** — Werkstück 2 von Phase 3.7
// (#1168).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { ErsetzungsAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Ersetzen } from './Ersetzen'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.ersetzen

// antwort baut eine Antwort mit einer Stelle.
function antwort(
  abweichung: Partial<ErsetzungsAntwort> = {},
): ErsetzungsAntwort {
  return {
    stellen: [
      {
        karteID: '7',
        nummer: 1,
        feld: 'rule text',
        vorher: 'Fliegend. Sehr schnell.',
        nachher: 'Rennend. Sehr schnell.',
        treffer: 1,
      },
    ],
    karten: 1,
    treffer: 1,
    weitere: false,
    stand: 'abc123',
    fehler: null,
    ...abweichung,
  }
}

function zeige(geaendert = vi.fn()) {
  render(
    <Sprachraum sprache="de">
      <Ersetzen setID="1" geaendert={geaendert} />
    </Sprachraum>,
  )

  return geaendert
}

// tippe füllt Suchtext und Ersatz.
async function tippe(was = 'Flieg', wodurch = 'Renn') {
  await userEvent.type(screen.getByLabelText(texte.suchtext), was)
  await userEvent.type(screen.getByLabelText(texte.ersatz), wodurch)
}

describe('Ersetzen', () => {
  // **Die Vorschau zeigt jede Stelle — und schreibt nichts.**
  it('zeigt die Vorschau, ohne zu schreiben', async () => {
    const vorschau = vi.fn(() => Promise.resolve(antwort()))
    const ausfuehren = vi.fn(() => Promise.resolve(antwort()))
    const geaendert = vi.fn()

    belegeApi({ ErsetzungVorschau: vorschau, ErsetzungAusfuehren: ausfuehren })
    zeige(geaendert)

    await tippe()
    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauen }),
    )

    // **Gewartet wird auf die Wirkung, nicht auf den Ruf** (§3, #1321).
    expect(
      await screen.findByText('Rennend. Sehr schnell.'),
    ).toBeInTheDocument()

    expect(vorschau).toHaveBeenCalledWith('1', {
      text: 'Flieg',
      ersatz: 'Renn',
      feld: '',
      grossKleinBeachten: false,
      ganzesWort: false,
      stand: '',
    })
    expect(ausfuehren).not.toHaveBeenCalled()
    expect(geaendert).not.toHaveBeenCalled()
  })

  // **Ersetzen erscheint erst nach der Vorschau** — der Schritt davor ist die
  // Zusage des Werkstücks.
  it('bietet das Ersetzen erst an, wenn eine Vorschau dasteht', async () => {
    belegeApi({
      ErsetzungVorschau: () => Promise.resolve(antwort()),
      ErsetzungAusfuehren: () => Promise.resolve(antwort()),
    })
    zeige()

    await tippe()

    expect(
      screen.queryByRole('button', {
        name: texte.ersetzen.replace('{treffer}', '1'),
      }),
    ).not.toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauen }),
    )

    expect(
      await screen.findByRole('button', {
        name: texte.ersetzen.replace('{treffer}', '1'),
      }),
    ).toBeInTheDocument()
  })

  // **Ausführen meldet den geänderten Bestand — und nennt die Rücknahme.**
  it('ersetzt und sagt, dass ⌘Z den ganzen Zug zurücknimmt', async () => {
    const ausfuehren = vi.fn(() => Promise.resolve(antwort()))
    const geaendert = vi.fn()

    belegeApi({
      ErsetzungVorschau: () => Promise.resolve(antwort()),
      ErsetzungAusfuehren: ausfuehren,
    })
    zeige(geaendert)

    await tippe()
    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauen }),
    )
    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.ersetzen.replace('{treffer}', '1'),
      }),
    )

    // **Gewartet wird auf die Wirkung, nicht auf den Ruf** (§3, #1321).
    // `toHaveBeenCalled` trifft zu, sobald gerufen wurde — die Antwort ist
    // dann noch unterwegs, und was sie bewirkt, steht noch nicht da.
    expect(await screen.findByText(texte.ruecknahme)).toBeInTheDocument()

    expect(ausfuehren).toHaveBeenCalled()
    expect(geaendert).toHaveBeenCalled()

    // **Und danach steht kein Ersetzenknopf mehr da** — der Zug ist getan.
    expect(
      screen.queryByRole('button', {
        name: texte.ersetzen.replace('{treffer}', '1'),
      }),
    ).not.toBeInTheDocument()
  })

  // **Die beiden Schalter gehen mit der Frage hinaus.**
  it('reicht »groß und klein« und »ganzes Wort« durch', async () => {
    const vorschau = vi.fn(() => Promise.resolve(antwort()))

    belegeApi({ ErsetzungVorschau: vorschau })
    zeige()

    await tippe()
    await userEvent.click(screen.getByLabelText(texte.grossKlein))
    await userEvent.click(screen.getByLabelText(texte.ganzesWort))
    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauen }),
    )

    await waitFor(() => {
      expect(vorschau).toHaveBeenCalledWith('1', {
        text: 'Flieg',
        ersatz: 'Renn',
        feld: '',
        grossKleinBeachten: true,
        ganzesWort: true,
        stand: '',
      })
    })
  })

  // **Ohne Suchtext ist der Knopf angehalten** — eine leere Frage träfe alles.
  it('hält den Knopf ohne Suchtext an', async () => {
    belegeApi({ ErsetzungVorschau: () => Promise.resolve(antwort()) })
    zeige()

    expect(
      screen.getByRole('button', { name: texte.vorschauen }),
    ).toBeDisabled()

    await userEvent.type(screen.getByLabelText(texte.suchtext), 'a')

    expect(
      screen.getByRole('button', { name: texte.vorschauen }),
    ).not.toBeDisabled()
  })

  // **Nichts gefunden sagt es** — statt einer leeren Liste (§6).
  it('erklärt, wenn nichts gefunden wurde', async () => {
    belegeApi({
      ErsetzungVorschau: () =>
        Promise.resolve(antwort({ stellen: [], karten: 0, treffer: 0 })),
    })
    zeige()

    await tippe()
    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauen }),
    )

    expect(await screen.findByText(texte.nichtsGefunden)).toBeInTheDocument()
  })

  // **Die Grenze der Vorschau steht dabei** — und sagt, dass trotzdem alles
  // ersetzt wird.
  it('nennt die Grenze der Vorschau', async () => {
    belegeApi({
      ErsetzungVorschau: () =>
        Promise.resolve(antwort({ weitere: true, karten: 300, treffer: 900 })),
    })
    zeige()

    await tippe()
    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauen }),
    )

    expect(await screen.findByText(texte.weitere)).toBeInTheDocument()
  })

  // **Ein Fehler steht als Meldung da, und der Bestand wird nicht gemeldet**
  // (§7, Regel 3).
  it('meldet einen Fehler und meldet den Bestand nicht als geändert', async () => {
    const geaendert = vi.fn()

    belegeApi({
      ErsetzungAusfuehren: () =>
        Promise.resolve(
          antwort({
            fehler: {
              code: 'konflikt',
              meldung: 'Der Bestand ist belegt.',
              schluessel: '',
              feld: '',
            },
          }),
        ),
      ErsetzungVorschau: () => Promise.resolve(antwort()),
    })
    zeige(geaendert)

    await tippe()
    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauen }),
    )
    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.ersetzen.replace('{treffer}', '1'),
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Bestand ist belegt.',
    )

    expect(geaendert).not.toHaveBeenCalled()

    // **Und die Vorschau steht noch da** — sie ist die Angabe, an der sich
    // ablesen lässt, was schiefging.
    expect(screen.getByText('Rennend. Sehr schnell.')).toBeInTheDocument()
  })
})

// **Eine Vorschau gilt für die Frage, aus der sie stammt** — Befund 5 des
// Zweitblicks über Phase 3.7.
describe('Eine Vorschau und danach eine andere Frage', () => {
  // # Warum das ein Fehler war und keine Kleinigkeit
  //
  // Der Knopf „Jetzt ersetzen" erscheint erst nach der Vorschau — das ist die
  // Zusage des Werkstücks. Er blieb aber stehen, wenn danach jemand den
  // Suchtext änderte, und führte dann die **neue**, nie gezeigte Frage aus:
  // unter der Trefferzahl der alten. Der Bestätigungsschritt bestätigte etwas
  // anderes als das, was geschah.
  it('nimmt den Ersetzenknopf weg, sobald sich die Frage ändert', async () => {
    belegeApi({
      ErsetzungVorschau: () => Promise.resolve(antwort()),
      ErsetzungAusfuehren: () => Promise.resolve(antwort()),
    })
    zeige()

    await tippe()

    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauen }),
    )

    const knopf = await screen.findByRole('button', {
      name: texte.ersetzen.replace('{treffer}', '1'),
    })

    expect(knopf).toBeInTheDocument()

    // **Ein einziges Zeichen genügt** — die Frage ist eine andere.
    await userEvent.type(screen.getByLabelText(texte.suchtext), 'e')

    expect(
      screen.queryByRole('button', {
        name: texte.ersetzen.replace('{treffer}', '1'),
      }),
    ).not.toBeInTheDocument()
  })

  // **Und ein Kästchen zählt genauso.** Groß/klein und „ganzes Wort" ändern,
  // was getroffen wird; ein Stand, der nur am Text hinge, gälte weiter.
  it('nimmt ihn auch weg, wenn ein Kästchen umschaltet', async () => {
    belegeApi({
      ErsetzungVorschau: () => Promise.resolve(antwort()),
      ErsetzungAusfuehren: () => Promise.resolve(antwort()),
    })
    zeige()

    await tippe()

    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauen }),
    )

    await screen.findByRole('button', {
      name: texte.ersetzen.replace('{treffer}', '1'),
    })

    await userEvent.click(screen.getByLabelText(texte.ganzesWort))

    expect(
      screen.queryByRole('button', {
        name: texte.ersetzen.replace('{treffer}', '1'),
      }),
    ).not.toBeInTheDocument()
  })
})

// **Eine bestätigte Vorschau geht mit ihrem Stand hinaus** —
// #1241.
//
// # Der Befund
//
// Die Seite band ihre Vorschau an die **Frage**, aus der sie stammt. Was sie
// nicht band, ist der **Datenstand**: zwischen Vorschau und Bestätigung kann
// eine Karte durch den Editor, eine Rücknahme, den Dienst oder den
// unabhängigen MCP-Prozess geändert werden — und der Knopf bestätigte dann
// andere Treffer als die angezeigten.
describe('Der Stand der Vorschau', () => {
  it('geht beim Bestätigen mit hinaus', async () => {
    const vorschau = vi.fn(() =>
      Promise.resolve(antwort({ stand: 'stand-von-vorhin' })),
    )
    const ausfuehren = vi.fn(() => Promise.resolve(antwort()))

    belegeApi({
      ErsetzungVorschau: vorschau,
      ErsetzungAusfuehren: ausfuehren,
    })
    zeige()

    await tippe()
    await userEvent.click(
      screen.getByRole('button', { name: texte.vorschauen }),
    )

    await screen.findByText('Rennend. Sehr schnell.')

    await userEvent.click(
      screen.getByRole('button', { name: /Jetzt ersetzen/ }),
    )

    await waitFor(() => {
      expect(ausfuehren).toHaveBeenCalledWith(
        '1',
        expect.objectContaining({ stand: 'stand-von-vorhin' }),
      )
    })
  })

  // **Ohne Vorschau geht kein Stand hinaus.** Es gibt nichts zu binden.
  it('bleibt leer, wenn keine Vorschau dasteht', async () => {
    const ausfuehren = vi.fn(() => Promise.resolve(antwort()))

    belegeApi({ ErsetzungAusfuehren: ausfuehren })
    zeige()

    await tippe()

    // **Der Knopf steht ohne Vorschau nicht da** — geprüft wird deshalb, dass
    // er nicht da ist: das ist die Bindung an die Frage, die schon steht.
    expect(
      screen.queryByRole('button', { name: /Jetzt ersetzen/ }),
    ).not.toBeInTheDocument()
  })
})
