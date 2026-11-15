// **Eine Auswahl reist mit** — UX-03.AK-01
// (#1675).
//
// # Was noch fehlte
//
// Vier der sechs Wege aus den Artboards standen seit #1761; offen waren die
// zwei mit einer **Mehrfachauswahl**: „3 Exemplare → Deck" im Navigator der
// Sammlung und „Die 7 eintauschen" im Inspector der Decks. Der Kommentar zum
// Issue nannte den Grund: die Kartenwahl legt einen Druck je Ruf, eine
// Handelszeile nennt einen Druck — eine Auswahl brauchte einen eigenen Weg.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum, useAnschluss } from './anschlussraum'
import { Auswahlband } from './bausteine/Auswahlband'
import { Deckstapelaufnahme } from './bausteine/Deckstapelaufnahme'
import { Tauschwunschknopf } from './bausteine/Tauschwunschknopf'
import { Tauschwunschaufnahme } from './bausteine/Tauschwunschaufnahme'
import { Sprachraum } from './sprache/kontext'
import { kataloge } from './sprache/kataloge'
import { einsetzen } from './sprache/einsetzen'
import { raeumeApi } from './test/api'
import { Bindungsfehler } from './bindungsfehler'
import { Anschlussbuehne, uebergabeknopf } from './test/anschlussbuehne'

import type { Anschluss } from './anschluss'
import type { ExemplarDTO } from './api'
import type { ZuordnungszeileDTO } from './apidecks'

const texte = kataloge.de

afterEach(() => {
  raeumeApi()
})

// Offen zeigt, was im Raum unterwegs ist — die Probe für eine Übergabe.
function Offen() {
  const anschluss = useAnschluss()

  return (
    <pre data-testid="offen">{JSON.stringify(anschluss.offen ?? null)}</pre>
  )
}

function offen(): Anschluss | null {
  return JSON.parse(screen.getByTestId('offen').textContent) as Anschluss | null
}

const exemplar = (teil: Partial<ExemplarDTO>): ExemplarDTO =>
  ({
    id: 1,
    printing: 'druck-a',
    name: 'Aschenwächter',
    bezeichnung: 'Aschenwächter',
    menge: 3,
    ...teil,
  }) as ExemplarDTO

describe('das Band in der Sammlung', () => {
  it('trägt die gewählten Exemplare gesammelt zu den Decks', async () => {
    const wechseln = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Anschlussraum aktuell="sammlung" wechseln={wechseln}>
          <Auswahlband
            exemplare={[
              exemplar({}),
              exemplar({ id: 2, printing: 'druck-b', name: 'Glut', menge: 1 }),
              exemplar({ id: 3, printing: 'druck-a', menge: 1 }),
            ]}
          />
          <Offen />
        </Anschlussraum>
      </Sprachraum>,
    )

    // **Die Zahl zählt Stücke**, nicht Zeilen: drei, eins und eins.
    await userEvent.click(
      screen.getByRole('button', {
        name: einsetzen(texte.seiten.sammlung.auswahlInsDeck, {
          anzahl: '5 Exemplare',
        }),
      }),
    )

    expect(wechseln).toHaveBeenCalledWith('decks')
    expect(offen()).toEqual({
      art: 'deckstapel',
      name: '5 Exemplare',
      stuecke: [
        { printing: 'druck-a', name: 'Aschenwächter', menge: 4 },
        { printing: 'druck-b', name: 'Glut', menge: 1 },
      ],
    })
  })

  // **Ein Exemplar ist eines** — die WebKit-Probe las „1 Exemplare → Deck".
  it('zählt in der Einzahl, wenn es eines ist', () => {
    render(
      <Sprachraum sprache="de">
        <Auswahlband exemplare={[exemplar({ menge: 1 })]} />
      </Sprachraum>,
    )

    expect(
      screen.getByRole('button', { name: '1 Exemplar → Deck' }),
    ).toBeInTheDocument()
  })

  it('steht ohne Auswahl nicht da', () => {
    render(
      <Sprachraum sprache="de">
        <Auswahlband exemplare={[]} />
      </Sprachraum>,
    )

    expect(screen.queryByRole('button')).toBeNull()
  })
})

describe('die Aufnahme im Deck', () => {
  const stapel: Anschluss = {
    art: 'deckstapel',
    name: '5 Exemplare',
    stuecke: [
      { printing: 'druck-a', name: 'Aschenwächter', menge: 4 },
      { printing: 'druck-b', name: 'Glut', menge: 1 },
    ],
  }

  it('legt alle mitgereisten Drucke in das offene Deck und verbraucht den Stapel', async () => {
    const legeAlle = vi.fn(() =>
      Promise.resolve({ gelegt: 2, abgewiesen: [], offen: [] }),
    )

    render(
      <Sprachraum sprache="de">
        <Anschlussbuehne anschluss={stapel}>
          <Deckstapelaufnahme legeAlle={legeAlle} angehalten={false} />
          <Offen />
        </Anschlussbuehne>
      </Sprachraum>,
    )

    expect(screen.queryByRole('button', { name: /ins Deck/ })).toBeNull()

    await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))

    expect(screen.getByText(/Aschenwächter × 4, Glut × 1/)).toBeInTheDocument()

    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.decks.stapelLegen }),
    )

    expect(legeAlle).toHaveBeenCalledWith(stapel.stuecke)
    await waitFor(() => {
      expect(offen()).toBeNull()
    })
  })

  // **Teilerfolg benennt beides** (S08.NAT-03): was liegt, und warum der
  // Rest nicht liegt — **am Stapel**, denn dort wurde gelegt (§6). Die
  // WebKit-Probe fand die erste Fassung: sie verwies auf „die Meldung
  // darüber", und dort stand keine.
  it('sagt, wie viele liegen und warum der Rest nicht', async () => {
    const legeAlle = vi.fn(() =>
      Promise.resolve({
        gelegt: 1,
        abgewiesen: [],
        // **Die Gestalt, die `auspacken` wirft** — kein nachgebauter Verbund.
        ursache: new Bindungsfehler({
          code: 'nichtGefunden',
          meldung: 'nicht gefunden',
          feld: '',
          schluessel: 'nichtGefunden',
        }),
        offen: [{ printing: 'druck-b', name: 'Glut', menge: 1 }],
      }),
    )

    render(
      <Sprachraum sprache="de">
        <Anschlussbuehne anschluss={stapel}>
          <Deckstapelaufnahme legeAlle={legeAlle} angehalten={false} />
          <Offen />
        </Anschlussbuehne>
      </Sprachraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.decks.stapelLegen }),
    )

    expect(
      await screen.findByText(
        einsetzen(texte.seiten.decks.stapelTeilweise, {
          gelegt: '1',
          gesamt: '2',
        }),
      ),
    ).toBeInTheDocument()
    expect(
      within(
        screen.getByRole('group', { name: texte.seiten.decks.stapelTitel }),
      ).getByRole('alert'),
    ).toHaveTextContent(texte.fehler.nichtGefunden)

    // **Der Stapel bleibt, aber nur sein Rest** — ein zweiter Klick legte
    // sonst die schon gelegten Drucke noch einmal.
    expect(offen()).not.toBeNull()
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.decks.stapelLegen }),
    )
    expect(legeAlle).toHaveBeenLastCalledWith([
      { printing: 'druck-b', name: 'Glut', menge: 1 },
    ])
  })
})

describe('ein Druck ohne Scryfall-Kennung im Stapel', () => {
  // **Der Grund nennt den Druck** — die WebKit-Probe las „Die Eingabe ist
  // nicht gültig. Der Hinweis am Feld sagt, woran es liegt": am Stapel gibt
  // es kein Feld, und der Satz der Domäne stand im Protokoll.
  it('sagt, welcher Druck keine Kennung hat', async () => {
    const glut = { printing: 'bbb-2', name: 'Glut', menge: 1 }
    const legeAlle = vi.fn(() =>
      Promise.resolve({
        gelegt: 0,
        abgewiesen: [
          {
            stueck: glut,
            ursache: new Bindungsfehler({
              code: 'validierung',
              meldung: '"bbb-2" ist keine Druckkennung.',
              feld: 'printing',
              schluessel: 'validierung',
            }),
          },
        ],
        offen: [glut],
      }),
    )

    render(
      <Sprachraum sprache="de">
        <Anschlussbuehne
          anschluss={{
            art: 'deckstapel',
            name: '1 Exemplar',
            stuecke: [{ printing: 'bbb-2', name: 'Glut', menge: 1 }],
          }}
        >
          <Deckstapelaufnahme legeAlle={legeAlle} angehalten={false} />
        </Anschlussbuehne>
      </Sprachraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))
    await userEvent.click(
      screen.getByRole('button', { name: texte.seiten.decks.stapelLegen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      einsetzen(texte.seiten.decks.stapelOhneDruck, { name: 'Glut' }),
    )
  })
})

describe('„Die N eintauschen“ im Deck', () => {
  const zeile = (teil: Partial<ZuordnungszeileDTO>): ZuordnungszeileDTO => ({
    printing: 'druck-a',
    druckID: '',
    vorlaeufig: false,
    bezeichnung: 'Aschenwächter',
    gebraucht: 4,
    vorhanden: 1,
    fehlend: 3,
    grund: '',
    reserviert: 0,
    zugeordnet: 0,
    frei: 0,
    zuordnungen: [],
    ...teil,
  })

  it('trägt, was fehlt, als Wunsch zum Handel', async () => {
    const wechseln = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Anschlussraum aktuell="decks" wechseln={wechseln}>
          <Tauschwunschknopf
            zeilen={[
              zeile({}),
              zeile({ printing: 'druck-b', bezeichnung: 'Glut', fehlend: 0 }),
              // **Ohne Druck kein Wunsch** — eine eigene Karte lässt sich
              // nicht bei einem Gegenüber holen.
              zeile({ printing: '', bezeichnung: 'Entwurf', fehlend: 2 }),
            ]}
          />
          <Offen />
        </Anschlussraum>
      </Sprachraum>,
    )

    await userEvent.click(
      screen.getByRole('button', {
        name: einsetzen(texte.seiten.decks.eintauschen, { anzahl: '3' }),
      }),
    )

    expect(wechseln).toHaveBeenCalledWith('handeln')
    expect(offen()).toEqual({
      art: 'tauschwunsch',
      name: '3 Karten',
      stuecke: [{ printing: 'druck-a', name: 'Aschenwächter', menge: 3 }],
    })
  })

  it('steht nicht da, wenn nichts fehlt', () => {
    render(
      <Sprachraum sprache="de">
        <Tauschwunschknopf zeilen={[zeile({ fehlend: 0 })]} />
      </Sprachraum>,
    )

    expect(screen.queryByRole('button')).toBeNull()
  })
})

describe('die Aufnahme im Handel', () => {
  const wunsch: Anschluss = {
    art: 'tauschwunsch',
    name: '3 Karten',
    stuecke: [{ printing: 'druck-a', name: 'Aschenwächter', menge: 3 }],
  }

  it('trägt die Wünsche als Positionen vom Gegenüber ein', async () => {
    const eintragen = vi.fn()

    render(
      <Sprachraum sprache="de">
        <Anschlussbuehne anschluss={wunsch}>
          <Tauschwunschaufnahme eintragen={eintragen} />
          <Offen />
        </Anschlussbuehne>
      </Sprachraum>,
    )

    await userEvent.click(screen.getByRole('button', { name: uebergabeknopf }))

    const kasten = screen.getByRole('group', {
      name: texte.seiten.handeln.wunschTitel,
    })
    expect(kasten).toHaveTextContent('Aschenwächter × 3')

    await userEvent.click(
      within(kasten).getByRole('button', {
        name: texte.seiten.handeln.wunschEintragen,
      }),
    )

    expect(eintragen).toHaveBeenCalledWith(wunsch.stuecke)
    expect(offen()).toBeNull()
  })
})
