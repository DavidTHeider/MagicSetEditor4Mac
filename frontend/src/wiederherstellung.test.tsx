// **Das Fenster, wenn der Bestand nicht aufging** —
// #962.
//
// # Warum das ein Lauf über die ganze Anwendung ist
//
// Geprüft wird nicht, dass eine Komponente etwas zeichnet, sondern **dass der
// Weg erreichbar ist**: `App` fragt beim Start nach, entscheidet, und die
// Seite dahinter ruft dieselben Bindungen wie die gewöhnliche
// Bestandspflege. Ein Lauf gegen `Wiederherstellung` allein sagte darüber
// nichts — genau das war der Befund.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { verstoesse } from './test/axe'
import { einsetzen } from './sprache/einsetzen'
import { kataloge } from './sprache/kataloge'

afterEach(() => {
  raeumeApi()
})

// **Die englischen Texte, und das hat einen Grund.**
//
// Ohne offenen Bestand gibt es keine gewählte Sprache — sie steht **in** ihm.
// Die Wiederherstellung spricht deshalb die Sprache des Systems, und die ist
// unter jsdom Englisch. Das ist keine Eigenheit des Laufs, sondern die
// Auskunft der Seite: wer sein Fenster auf Deutsch gestellt hat, bekommt sie
// hier auf Systemsprache, weil die Einstellung nicht lesbar ist (#962).
const worte = kataloge.en.seiten.bestandspflege

// rueckwegProbe ist die Sicherung, die diese Fassung öffnen kann.
const rueckwegProbe = {
  datei: 'bestand-vor-0.4.0.db',
  schemastand: '0035_platz.sql',
  sets: 12,
  karten: 840,
}

// grundlage belegt, was jeder Lauf hier braucht.
function grundlage(befund: {
  offen: boolean
  grund: string
  schluessel?: string
  rueckweg?: {
    datei: string
    schemastand: string
    sets: number
    karten: number
  } | null
}) {
  belegeApi({
    ...belegeFenster('de'),
    // **Wie im Betrieb: ohne offenen Bestand antwortet die Einstellung
    // nicht** — sie steht **in** ihm. Hier antwortete sie mit „de", und
    // welche Sprache die Seite sprach, entschied, ob diese Antwort vor dem
    // Startbefund ankam: allein lief der Lauf deutsch, unter der Last der
    // Pipeline englisch (gefunden am 2026-09-24, PR zu #1674). Jetzt spricht
    // die Seite die Sprache des Systems, wie sie es im Fenster tut.
    ...(befund.offen
      ? {}
      : { Einstellungen: () => new Promise<never>(() => undefined) }),
    Startbefund: vi.fn().mockResolvedValue(befund),
    Datenstand: vi.fn().mockResolvedValue({ stand: 0, fehler: null }),
  })
}

describe('Ein Bestand, der nicht aufging', () => {
  it('zeigt den Grund und den Weg zur Sicherung', async () => {
    grundlage({
      offen: false,
      grund: 'migrieren: doppelte Spalte urheber',
    })

    render(<App />)

    expect(
      await screen.findByRole('heading', {
        name: worte.wiederherstellungTitel,
      }),
    ).toBeInTheDocument()

    expect(
      screen.getByText('migrieren: doppelte Spalte urheber'),
    ).toBeInTheDocument()

    expect(
      screen.getByRole('button', { name: worte.waehlen }),
    ).toBeInTheDocument()

    // **Und kein Sicherungsknopf.** Es ist nichts offen, was sich sichern
    // ließe; ein Knopf, der nur scheitern kann, ist ein falsches Versprechen.
    expect(
      screen.queryByRole('button', { name: worte.sichern }),
    ).not.toBeInTheDocument()
  })

  // **Die Menüleiste spricht dieselbe Sprache wie die Seite** — Befund 2 des
  // Zweitblicks über Phase 6.2. Go baut sie beim Start deutsch; die Meldung
  // der Sprache saß allein in `Anwendung`, und die steht auf diesem Weg
  // nicht. Das Fenster war englisch, seine Leiste deutsch (#255).
  it('meldet die Sprache auch der Menüleiste', async () => {
    const SpracheSetzen = vi.fn(() => Promise.resolve({ fehler: null }))

    belegeApi({
      ...belegeFenster('de'),
      Startbefund: vi.fn().mockResolvedValue({ offen: false, grund: 'kaputt' }),
      Datenstand: vi.fn().mockResolvedValue({ stand: 0, fehler: null }),
      SpracheSetzen,
    })

    render(<App />)

    expect(
      await screen.findByRole('heading', {
        name: worte.wiederherstellungTitel,
      }),
    ).toBeInTheDocument()

    await waitFor(() => {
      expect(SpracheSetzen).toHaveBeenCalledWith('en')
    })
  })

  it('spielt eine Sicherung über dieselben Bindungen zurück', async () => {
    const SicherungWaehlen = vi.fn().mockResolvedValue({
      abgebrochen: false,
      pfad: '/tmp/sicherung.db',
      angelegtAm: '2026-08-25',
      schemastand: '0016_packarten.sql',
      sets: 3,
      karten: 42,
      verschluesselt: false,
      paketeEnthalten: [],
      paketeFehlen: [],
      fehler: null,
    })

    const SicherungZurueckspielen = vi.fn().mockResolvedValue({
      ersetzt: true,
      beiseite: '/tmp/bestand-vor-rueckspielung.db',
      unveraendert: false,
      fehler: null,
    })

    const Startbefund = vi
      .fn()
      .mockResolvedValueOnce({ offen: false, grund: 'kaputt' })
      .mockResolvedValue({ offen: true, grund: '' })

    belegeApi({
      ...belegeFenster('de'),
      Startbefund,
      Datenstand: vi.fn().mockResolvedValue({ stand: 0, fehler: null }),
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
      SicherungWaehlen,
      SicherungZurueckspielen,
    })

    render(<App />)

    await userEvent.click(
      await screen.findByRole('button', { name: worte.waehlen }),
    )

    await userEvent.click(
      await screen.findByRole('button', { name: worte.bestaetigen }),
    )

    await waitFor(() => {
      expect(SicherungZurueckspielen).toHaveBeenCalledWith(
        '/tmp/sicherung.db',
        '',
      )
    })

    // **Und danach ist die Anwendung da** — der zweite Startbefund sagt
    // „offen", und die gewöhnliche Oberfläche kommt.
    await waitFor(() => {
      expect(
        screen.queryByRole('heading', { name: worte.wiederherstellungTitel }),
      ).not.toBeInTheDocument()
    })
  })

  // **Der Rat nennt die Datei** — Werkstück 5 von Phase 4.13
  // (#1352).
  //
  // „Spiele eine Sicherung von vorher zurück" ist kein Rat, solange neben dem
  // Bestand dreißig Dateien mit Fassungsnummern im Namen liegen und nur eine
  // davon aufgeht. Welche das ist, hat Go nachgesehen; hier steht sie.
  it('nennt die Sicherung, die diese Fassung öffnen kann', async () => {
    grundlage({
      offen: false,
      grund: 'migrieren: 9999_aus_der_zukunft.sql kennt diese Fassung nicht',
      schluessel: 'bestandNeuereFassung',
      rueckweg: rueckwegProbe,
    })

    render(<App />)

    expect(
      await screen.findByRole('heading', { name: worte.wegZurueckTitel }),
    ).toBeInTheDocument()

    expect(
      screen.getByText(
        einsetzen(worte.wegZurueck, {
          datei: 'bestand-vor-0.4.0.db',
          stand: '0035_platz.sql',
          sets: '12',
          karten: '840',
        }),
      ),
    ).toBeInTheDocument()
  })

  // **Und keine, wenn keine dalag.**
  //
  // Die Gegenprobe zur Zusage (§10): ein Satz, der immer eine Datei nennt,
  // nennt irgendwann eine, die es nicht gibt — und danach sucht jemand.
  it('nennt keine, wenn keine brauchbare dalag', async () => {
    grundlage({
      offen: false,
      grund: 'migrieren: 9999_aus_der_zukunft.sql kennt diese Fassung nicht',
      schluessel: 'bestandNeuereFassung',
      rueckweg: null,
    })

    render(<App />)

    expect(
      await screen.findByRole('heading', {
        name: worte.wiederherstellungTitel,
      }),
    ).toBeInTheDocument()

    expect(
      screen.queryByRole('heading', { name: worte.wegZurueckTitel }),
    ).not.toBeInTheDocument()
  })

  // **Und sie hält WCAG 2.2 AA ein** — nachgetragen am 2026-09-07.
  //
  // # Warum das hier steht und nicht in `barrierefreiheit.test.tsx`
  //
  // **Weil sie nicht in `navigation.ts` steht.** Jener Lauf nimmt seine
  // Seitenliste von dort, und das ist richtig — eine Kopie liefe auseinander.
  // Die Wiederherstellung ist aber keine Station der Navigation: man landet
  // dort, wenn der Bestand nicht aufgeht, und dann gibt es keine Navigation.
  //
  // Damit war sie **nie** geprüft, obwohl der Kopf dort „jede Seite" sagt.
  // Geprüft wird über **dieselbe** Abfrage (`test/axe.ts`); eine zweite
  // Fassung der Markenliste wäre die schwächere, und sie sähe genauso grün aus.
  //
  // **Mit und ohne Weg zurück**, weil das zwei verschiedene Bäume sind: der
  // zusätzliche Block trägt eine eigene Überschrift, und eine Überschrift an
  // der falschen Ebene ist genau das, was axe findet.
  it.each([
    ['mit einem Weg zurück', rueckwegProbe],
    ['ohne einen', null],
  ])('hält WCAG 2.2 AA ein — %s', async (_name, rueckweg) => {
    grundlage({
      offen: false,
      grund: 'migrieren: 9999_aus_der_zukunft.sql kennt diese Fassung nicht',
      schluessel: 'bestandNeuereFassung',
      rueckweg,
    })

    const { container } = render(<App />)

    await screen.findByRole('heading', { name: worte.wiederherstellungTitel })

    expect(await verstoesse(container)).toEqual([])
  })

  it('bleibt fort, wenn der Bestand offen ist', async () => {
    grundlage({ offen: true, grund: '' })

    belegeApi({
      ...belegeFenster('de'),
      Startbefund: vi.fn().mockResolvedValue({ offen: true, grund: '' }),
      Datenstand: vi.fn().mockResolvedValue({ stand: 0, fehler: null }),
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
    })

    render(<App />)

    await waitFor(() => {
      expect(
        screen.queryByRole('heading', { name: worte.wiederherstellungTitel }),
      ).not.toBeInTheDocument()
    })
  })

  // **Scheitert die Frage selbst, gilt der Bestand als offen.**
  //
  // Sie geht über dieselbe Naht wie alles andere; kommt sie nicht durch, ist
  // das Fenster ohnehin nicht benutzbar, und eine Wiederherstellungsseite,
  // die auch nicht arbeiten kann, verdeckte nur die eigentliche Meldung.
  it('verdeckt eine kaputte Naht nicht', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Startbefund: vi.fn().mockRejectedValue(new Error('keine Naht')),
      Datenstand: vi.fn().mockResolvedValue({ stand: 0, fehler: null }),
      Sets: vi.fn().mockResolvedValue({ sets: [], fehler: null }),
    })

    render(<App />)

    await waitFor(() => {
      expect(
        screen.queryByRole('heading', { name: worte.wiederherstellungTitel }),
      ).not.toBeInTheDocument()
    })
  })
})

// **Der Start sagt, was zu tun ist — nicht, was SQLite meldet** — Werkstück 3
// von Phase 4.13
// (#1352).
//
// **Am 2026-09-07 gemessen**: eine Datei, die kein Bestand ist, gab
// „journal_mode: sqlite3: file is not a database" — und das stand als einzige
// Auskunft im Fenster.
describe('der Grund eines gescheiterten Starts', () => {
  it('zeigt den Satz zum Schlüssel und die Einzelheit darunter', async () => {
    grundlage({
      offen: false,
      grund: 'journal_mode: sqlite3: file is not a database',
      schluessel: 'bestandUnbrauchbar',
    })

    render(<App />)

    expect(
      await screen.findByText(kataloge.en.fehler.bestandUnbrauchbar),
    ).toBeInTheDocument()

    // **Und der rohe Grund bleibt als Spur stehen** — wer hier steht, sucht
    // einen Fehler.
    expect(
      screen.getByText('journal_mode: sqlite3: file is not a database'),
    ).toBeInTheDocument()
  })

  // **Ein Schlüssel, den der Katalog nicht kennt, lässt den Grund stehen** —
  // eine unübersetzte Meldung ist besser als eine leere (§9).
  it('kommt ohne Satz aus, wenn der Schlüssel unbekannt ist', async () => {
    grundlage({ offen: false, grund: 'irgendwas', schluessel: 'gibtsnicht' })

    render(<App />)

    expect(await screen.findByText('irgendwas')).toBeInTheDocument()
  })
})
