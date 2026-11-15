// **Die Zugangsdaten auf der Einstellungsseite** — Werkstück 8 von Phase 4.8
// (#1321).
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Zugaenge } from '../bausteine/Zugaenge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import type { ZugangsstandDTO } from '../apizugaenge'

const texte = kataloge.de.seiten.einstellungen

const staende: ZugangsstandDTO[] = [
  { dienst: 'scryfall', angaben: [], hinterlegt: [], vollstaendig: true },
  {
    dienst: 'mtgjson',
    angaben: ['marke'],
    hinterlegt: [],
    vollstaendig: false,
  },
  {
    dienst: 'scrydex',
    angaben: ['schluessel', 'team'],
    hinterlegt: ['schluessel'],
    vollstaendig: false,
  },
  // **CardTrader benutzt dieselbe Angabeart wie MTGJSON** — `marke` —, und
  // genau daran hing der Befund: eine Beschriftung je Angabeart kann nicht
  // für beide stimmen (#1534).
  {
    dienst: 'cardtrader',
    angaben: ['marke'],
    hinterlegt: [],
    vollstaendig: false,
  },
]

afterEach(() => {
  raeumeApi()
})

function hinstellen(weitere: Record<string, unknown> = {}) {
  belegeApi({
    Zugaenge: () => Promise.resolve({ zugaenge: staende, fehler: null }),
    ...weitere,
  })

  render(
    <Sprachraum sprache="de">
      <Zugaenge />
    </Sprachraum>,
  )
}

describe('die Zugangsdaten', () => {
  // **Dienste ohne Zugangsdaten stehen nicht da.** Scryfall braucht keine;
  // eine leere Zeile dafür wäre eine Frage ohne Antwort.
  it('zeigt nur die Dienste, die Zugangsdaten brauchen', async () => {
    hinstellen()

    expect(await screen.findByText('MTGJSON')).toBeInTheDocument()
    expect(screen.getByText('Scrydex')).toBeInTheDocument()
    expect(screen.queryByText('Scryfall')).toBeNull()
  })

  // **Die letzte Reihe**: eine Angabe, die im Katalog keinen Namen hat, zeigt
  // ihren rohen Bezeichner — statt einer leeren Zeile.
  //
  // **Dass es nie so weit kommt**, hält
  // `internal/werkzeuge/einstellungstexte_test.go` gegen die Go-Seite. Hier
  // steht, was geschieht, wenn dieser Wächter einmal umgangen wird.
  it('zeigt den rohen Bezeichner, wenn er keinen Namen hat', async () => {
    belegeApi({
      Zugaenge: () =>
        Promise.resolve({
          zugaenge: [
            {
              dienst: 'mtgjson',
              angaben: ['unbekannt'],
              hinterlegt: [],
              vollstaendig: false,
            },
          ],
          fehler: null,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Zugaenge />
      </Sprachraum>,
    )

    expect(
      await screen.findByText(`unbekannt: ${texte.zugangFehlt}`),
    ).toBeInTheDocument()
  })

  // **Befund 3 und 5 des Zweitblicks über Phase 4.8.**
  //
  // Die Go-Seite gibt Schlüsselbundfehler in `SchreibAntwort.Fehler` zurück
  // und prüft diesen Weg ausdrücklich (`zugangbindungen_test.go`). Das Fenster
  // hat die Antwort **gar nicht angesehen**: es leerte das Feld und lud nach,
  // gleich was zurückkam.
  //
  // Damit war zweierlei verletzt: der fachliche Fehler verschwand (§8), und
  // nach einem **gescheiterten** Schreiben wurde nachgeladen (§7, Regel 3).
  it('meldet einen gescheiterten Eintrag am Feld und lädt nicht nach', async () => {
    const Zugaenge = vi.fn(() =>
      Promise.resolve({ zugaenge: staende, fehler: null }),
    )

    hinstellen({
      Zugaenge,
      ZugangSetzen: () =>
        Promise.resolve({
          fehler: {
            code: 'validierung',
            schluessel: 'fehler.validierung',
            meldung: 'Der Schlüsselbund ist verschlossen.',
          },
        }),
    })

    const scrydex = await screen.findByRole('group', { name: 'Scrydex' })

    await userEvent.type(
      within(scrydex).getByLabelText(`Scrydex: ${texte.angaben.team}`),
      'mannschaft',
    )

    const gelesen = Zugaenge.mock.calls.length

    await userEvent.click(
      within(scrydex).getByRole('button', {
        name: `${texte.zugangSetzen}: Scrydex ${texte.angaben.team}`,
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'Der Schlüsselbund ist verschlossen.',
    )

    // **Nicht nachgeladen** (§7, Regel 3): ein Lesefehler überschriebe sonst
    // die Meldung, die sagt, was zu tun ist.
    expect(Zugaenge.mock.calls.length).toBe(gelesen)

    // **Und das Eingetippte steht noch da** — wer es verlöre, müsste seinen
    // Schlüssel ein zweites Mal abtippen.
    expect(
      within(scrydex).getByLabelText(`Scrydex: ${texte.angaben.team}`),
    ).toHaveValue('mannschaft')
  })

  // **Ein Vorgang je Angabe** (§7, Regel 4).
  //
  // `docs/SCHREIBWEGE.md` behauptete, zwei Aufrufe für dieselbe Angabe seien
  // nicht möglich: „das Feld ist nach dem ersten leer". Das Feld wird aber
  // erst **nach** der Antwort geleert, und der Knopf steht die ganze Zeit da.
  it('lässt keinen zweiten Eintrag los, solange der erste läuft', async () => {
    let loesen = (): void => undefined

    const ZugangSetzen = vi.fn(
      () =>
        new Promise<{ fehler: null }>((fertig) => {
          loesen = () => {
            fertig({ fehler: null })
          }
        }),
    )

    hinstellen({ ZugangSetzen })

    const scrydex = await screen.findByRole('group', { name: 'Scrydex' })

    await userEvent.type(
      within(scrydex).getByLabelText(`Scrydex: ${texte.angaben.team}`),
      'mannschaft',
    )

    const knopf = within(scrydex).getByRole('button', {
      name: `${texte.zugangSetzen}: Scrydex ${texte.angaben.team}`,
    })

    await userEvent.click(knopf)
    await userEvent.click(knopf)

    expect(ZugangSetzen).toHaveBeenCalledTimes(1)

    loesen()
  })

  // **Was liegt, wird als »liegt« gezeigt — und nie als Wert.**
  it('sagt, welche Angabe hinterlegt ist', async () => {
    hinstellen()

    const scrydex = await screen.findByRole('group', { name: 'Scrydex' })

    expect(
      within(scrydex).getByText(
        `${texte.angaben.schluessel}: ${texte.zugangLiegt}`,
      ),
    ).toBeInTheDocument()

    expect(
      within(scrydex).getByText(`${texte.angaben.team}: ${texte.zugangFehlt}`),
    ).toBeInTheDocument()
  })

  it('trägt eine Angabe ein', async () => {
    const ZugangSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

    hinstellen({ ZugangSetzen })

    const scrydex = await screen.findByRole('group', { name: 'Scrydex' })

    await userEvent.type(
      within(scrydex).getByLabelText(`Scrydex: ${texte.angaben.team}`),
      'mannschaft',
    )

    await userEvent.click(
      within(scrydex).getByRole('button', {
        name: `${texte.zugangSetzen}: Scrydex ${texte.angaben.team}`,
      }),
    )

    await waitFor(() => {
      expect(ZugangSetzen).toHaveBeenCalledWith('scrydex', 'team', 'mannschaft')
    })
  })

  // **Erst nach einer Frage** — UX-10.AK-04 (#1682): der Wert ist danach aus
  // dem Schlüsselbund weg, und zurück gibt ihn nur der Dienst. Bis zum
  // 2026-09-25 genügte ein Klick.
  it('nimmt eine Angabe heraus, wenn die Frage bejaht ist', async () => {
    const ZugangLoeschen = vi.fn(() => Promise.resolve({ fehler: null }))

    hinstellen({ ZugangLoeschen })

    const scrydex = await screen.findByRole('group', { name: 'Scrydex' })
    const knopf = within(scrydex).getByRole('button', {
      name: `${texte.zugangLoeschen}: Scrydex ${texte.angaben.schluessel}`,
    })

    // **Abbrechen nimmt nichts heraus**, und der Fokus steht wieder am Knopf.
    await userEvent.click(knopf)
    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.bausteine.aktionen.abbrechen,
      }),
    )
    expect(ZugangLoeschen).not.toHaveBeenCalled()
    expect(knopf).toHaveFocus()

    await userEvent.click(knopf)

    const frage = screen.getByRole('group', {
      name: `${texte.angaben.schluessel} für Scrydex löschen?`,
    })
    await userEvent.click(
      within(frage).getByRole('button', { name: texte.zugangLoeschenJa }),
    )

    await waitFor(() => {
      expect(ZugangLoeschen).toHaveBeenCalledWith('scrydex', 'schluessel')
    })
  })

  // **Das Feld ist ein Passwortfeld.** Wer den Schlüssel eintippt, soll ihn
  // nicht auf dem Bildschirm stehen haben.
  it('verbirgt die Eingabe', async () => {
    hinstellen()

    // **Der Name je Dienst und nicht der Rückfall** (#1534): MTGJSON hat
    // einen eigenen, und ihn über `angaben.marke` zu suchen hieße, die
    // Auflösung zu umgehen, die dieser Befund gerade eingeführt hat.
    const feld = await screen.findByLabelText(
      `MTGJSON: ${texte.dienstangaben['mtgjson.marke']}`,
    )

    expect(feld).toHaveAttribute('type', 'password')
  })

  // **Ein unbekannter Dienst fällt auf seinen Schlüssel zurück.**
  //
  // Er kann nur entstehen, wenn Go einen fünften Dienst kennt und das Fenster
  // noch nicht — dann steht dort `neuerdienst` statt einer leeren Zeile.
  it('nennt einen unbekannten Dienst bei seinem Schlüssel', async () => {
    belegeApi({
      Zugaenge: () =>
        Promise.resolve({
          zugaenge: [
            {
              dienst: 'neuerdienst',
              angaben: ['marke'],
              hinterlegt: [],
              vollstaendig: false,
            },
          ],
          fehler: null,
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Zugaenge />
      </Sprachraum>,
    )

    expect(
      await screen.findByRole('group', { name: 'neuerdienst' }),
    ).toBeInTheDocument()
  })

  // **Dieselbe Angabeart, zwei Dienste, zwei Namen** (#1534).
  //
  // # Der Befund, für den es diesen Lauf gibt
  //
  // **Bei CardTrader stand „Patreon-Marke".** `domain.ZugangMarke` bedient
  // zwei Dienste — MTGJSON und CardTrader —, und der Katalog beschriftete
  // nach der **Angabeart**. Eine Zeichenkette kann nicht für beide stimmen:
  // MTGJSONs Marke gibt es nur für Patreon-Unterstützer
  // ([0056](../../docs/entscheidungen/0056-drei-weitere-kartendienste.md)),
  // CardTraders kommt aus den Profileinstellungen seines eigenen Kontos
  // ([0095](../../docs/entscheidungen/0095-was-cardtrader-sagt-und-was-davon-gilt.md)).
  //
  // **Kein Wächter konnte es sehen**, und er sagte das auch: er hält, dass
  // ein Name dasteht, nicht dass er den richtigen Anbieter trifft.
  it('beschriftet dieselbe Angabeart je Dienst verschieden', async () => {
    hinstellen()

    const cardtrader = await screen.findByRole('group', { name: 'CardTrader' })
    const mtgjson = await screen.findByRole('group', { name: 'MTGJSON' })

    // **Patreon gehört zu MTGJSON und zu sonst niemandem.**
    expect(within(mtgjson).getByText(/Patreon/)).toBeInTheDocument()
    expect(within(cardtrader).queryByText(/Patreon/)).toBeNull()

    // **Und CardTrader nennt, woher seine Marke kommt** — aus dem Konto,
    // nicht aus einem fremden Unterstützerprogramm.
    expect(within(cardtrader).getByText(/Profil/)).toBeInTheDocument()
  })

  // **Ein Dienst ohne eigenen Namen fällt auf die Angabeart zurück.**
  //
  // **Das ist die Hälfte, die den neuen Block tragbar macht**: ein Dienst,
  // den Go kennt und der Katalog noch nicht, steht sonst ohne Namen da — und
  // der rohe Bezeichner war der Zustand, den Werkstück 8 von Phase 4.8
  // abgeschafft hat.
  it('fällt auf die Angabeart zurück, wo kein eigener Name steht', async () => {
    hinstellen()

    const scrydex = await screen.findByRole('group', { name: 'Scrydex' })

    expect(
      within(scrydex).getByText(`${texte.angaben.team}: ${texte.zugangFehlt}`),
    ).toBeInTheDocument()
  })

  it('meldet einen Fehler', async () => {
    hinstellen({
      Zugaenge: () =>
        Promise.resolve({
          zugaenge: [],
          fehler: { code: 'unbekannt', meldung: 'kaputt', feld: '' },
        }),
    })

    expect(await screen.findByText(/kaputt/)).toBeInTheDocument()
  })
})
