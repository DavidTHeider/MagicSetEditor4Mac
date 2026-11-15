// **Was ein Gegenüber von der Sammlung sieht** — Werkstück T3 von Phase 5.6
// (#1468).
//
// # Was hier gemessen wird
//
// **Dass der Kasten erst bei einem bestätigten Gegenüber erscheint.** Eine
// Freigabe an einen unbestätigten tut nichts — die Domäne fragt zuerst den
// Kontakt. Ein Formular, das etwas erteilt, das nicht wirkt, wäre eine Zusage
// ohne Einlösung.
//
// **Dass `paarung` nicht zur Wahl steht.** Sie ist das Kennenlernen, und die
// Domäne weist eine Freigabe darauf ab; eine Wahl, die zu einem Nein führt,
// ist keine Wahl.
//
// **Und dass die Vorschau nach einer Änderung verschwindet** (§7, Regel 1):
// sie zeigte sonst, was **vorher** galt — die gefährlichste Auskunft, die
// dieser Kasten geben kann.
import { screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { einKontakt, eineListe, zeige } from './peerkontaktprobe'
import { kataloge } from '../sprache/kataloge'

import type {
  Sichtfreigabe,
  Sichtfreigabenantwort,
  Sichtvorschauantwort,
} from '../apisichtfreigabe'

const eigene = kataloge.de.seiten.technik

// **Hinter dem Aufklapper „Was dieses Gegenüber sieht"** (UXR-16, #1924):
// offen fragte jede bestätigte Zeile schon beim Zeichnen den Bestand; jetzt
// öffnet der Lauf den Kasten, wie ein Mensch es tut.
const aufklapper = eigene.freigabenTitel

// dieUmfaenge sind die fünf aus `domain`.
const dieUmfaenge = [
  'auswahl',
  'dubletten',
  'wuensche',
  'sammlung',
  'digitales',
]

// eineFreigabe baut eine erteilte Sicht.
function eineFreigabe(anders: Partial<Sichtfreigabe> = {}): Sichtfreigabe {
  return {
    empfaenger: 'bgegenueber',
    umfang: 'sammlung',
    zweck: 'handel',
    auswahl: [],
    behalten: 0,
    ablauf: '',
    erteilt: '2026-09-18T12:00:00Z',
    ...anders,
  }
}

// freigabenliste baut die Antwort.
function freigabenliste(
  freigaben: Sichtfreigabe[] = [],
): Sichtfreigabenantwort {
  return { freigaben, umfaenge: dieUmfaenge, fehler: null }
}

// eineVorschau baut die Sicht eines Gegenübers.
function eineVorschau(): Sichtvorschauantwort {
  return {
    zeilen: [
      {
        referenz: 'ABCdefGHI',
        printing: 'a1b2',
        menge: 3,
        zustand: 'played',
        sprache: 'de',
        finish: 'etched',
        bezeichnung: 'Lightning Bolt',
      },
    ],
    fehler: null,
  }
}

// bestaetigt ist ein Kontakt, an dem der Kasten erscheint.
function bestaetigt() {
  return einKontakt({ stand: 'bestaetigt', rechte: ['handel'] })
}

describe('Was ein Gegenüber von der Sammlung sieht', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  it('erscheint nicht an einem unbestätigten Gegenüber', async () => {
    belegeApi({ Peerkontakte: () => Promise.resolve(eineListe()) })

    await zeige()

    expect(
      await screen.findByText(eigene.kontakteStandUnbestaetigt),
    ).toBeInTheDocument()

    expect(screen.queryByText(eigene.freigabenTitel)).not.toBeInTheDocument()
  })

  it('nennt die erteilten Sichten und sagt, wenn es keine gibt', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Sammlungsfreigaben: () => Promise.resolve(freigabenliste()),
    })

    await zeige(aufklapper)

    expect(await screen.findByText(eigene.freigabenLeer)).toBeInTheDocument()
  })

  it('erteilt die gewählte Sicht', async () => {
    const erteilt = vi.fn(() => Promise.resolve(freigabenliste()))

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Sammlungsfreigaben: () => Promise.resolve(freigabenliste()),
      SammlungFreigeben: erteilt,
    })

    const hand = await zeige(aufklapper)

    await hand.selectOptions(
      await screen.findByLabelText(eigene.freigabenUmfang),
      'dubletten',
    )

    await hand.click(
      screen.getByRole('button', { name: eigene.freigabenErteilen }),
    )

    expect(erteilt).toHaveBeenCalledWith(
      expect.objectContaining({
        empfaenger: 'bgegenueber',
        umfang: 'dubletten',
        zweck: 'handel',
      }),
    )
  })

  it('widerruft eine erteilte Sicht', async () => {
    const widerrufen = vi.fn(() => Promise.resolve(freigabenliste()))

    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Sammlungsfreigaben: () =>
        Promise.resolve(freigabenliste([eineFreigabe()])),
      SammlungsfreigabeWiderrufen: widerrufen,
    })

    const hand = await zeige(aufklapper)

    await hand.click(
      await screen.findByRole('button', {
        name: eigene.freigabenWiderrufen,
      }),
    )

    expect(widerrufen).toHaveBeenCalledWith('bgegenueber', 'handel', 'sammlung')
  })

  it('zeigt, was das Gegenüber sähe — und räumt es nach einer Änderung', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Sammlungsfreigaben: () =>
        Promise.resolve(freigabenliste([eineFreigabe()])),
      Freigabesicht: () => Promise.resolve(eineVorschau()),
      SammlungsfreigabeWiderrufen: () => Promise.resolve(freigabenliste()),
    })

    const hand = await zeige(aufklapper)

    await hand.click(
      await screen.findByRole('button', { name: eigene.freigabenVorschau }),
    )

    expect(await screen.findByText(/Lightning Bolt/)).toBeInTheDocument()

    // **Und nichts Privates**: die Vorschau zeigt sechs Felder.
    expect(screen.queryByText(/4,17|Kiste/)).toBeNull()

    await hand.click(
      screen.getByRole('button', { name: eigene.freigabenWiderrufen }),
    )

    expect(screen.queryByText(/Lightning Bolt/)).toBeNull()
  })

  // **Ohne erlaubten Zweck steht kein Formular da** — es erteilte etwas,
  // das nie gilt.
  it('bietet nichts an, solange das Gegenüber nichts darf', async () => {
    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(
          eineListe([einKontakt({ stand: 'bestaetigt', rechte: [] })]),
        ),
      Sammlungsfreigaben: () => Promise.resolve(freigabenliste()),
    })

    await zeige(aufklapper)

    expect(
      await screen.findByText(eigene.freigabenOhneZweck),
    ).toBeInTheDocument()

    expect(
      screen.queryByRole('button', { name: eigene.freigabenErteilen }),
    ).toBeNull()
  })

  // **Der Zweck ist wählbar, und er geht mit** — sonst erteilte jede Wahl
  // dieselbe Sicht.
  it('erteilt für den gewählten Zweck', async () => {
    const erteilt = vi.fn(() => Promise.resolve(freigabenliste()))

    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(
          eineListe([
            einKontakt({ stand: 'bestaetigt', rechte: ['teilen', 'handel'] }),
          ]),
        ),
      Sammlungsfreigaben: () => Promise.resolve(freigabenliste()),
      SammlungFreigeben: erteilt,
    })

    const hand = await zeige(aufklapper)

    await hand.selectOptions(
      await screen.findByLabelText(eigene.freigabenZweck),
      'handel',
    )

    // **Ein Umfang ohne Angabe** — „Ausgewählte Drucke" verlangt Drucke
    // (UXR-F18, #2033); dieser Lauf misst den Zweck, nicht die Wahl.
    await hand.selectOptions(
      screen.getByLabelText(eigene.freigabenUmfang),
      'sammlung',
    )

    await hand.click(
      screen.getByRole('button', { name: eigene.freigabenErteilen }),
    )

    expect(erteilt).toHaveBeenCalledWith(
      expect.objectContaining({ zweck: 'handel' }),
    )
  })

  // **Ein gescheitertes Erteilen sagt es** — und nimmt die Liste nicht weg
  // (§7, Regel 3).
  it('meldet ein gescheitertes Erteilen', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Sammlungsfreigaben: () => Promise.resolve(freigabenliste()),
      SammlungFreigeben: () =>
        Promise.resolve({
          freigaben: [],
          umfaenge: dieUmfaenge,
          // **Ein Schlüssel, den kein Katalog kennt** — dann steht die
          // Meldung aus Go da, und der Lauf misst, dass sie ankommt, statt
          // einen übersetzten Satz wiederzuerkennen.
          fehler: {
            code: 'validierung' as const,
            schluessel: 'gibtesnicht',
            meldung: 'so nicht',
          },
        }),
    })

    const hand = await zeige(aufklapper)

    // **Ein Umfang ohne Angabe** — sonst bliebe der Ruf aus (UXR-F18, #2033).
    await hand.selectOptions(
      await screen.findByLabelText(eigene.freigabenUmfang),
      'sammlung',
    )

    await hand.click(
      await screen.findByRole('button', { name: eigene.freigabenErteilen }),
    )

    expect(await screen.findByText('so nicht')).toBeInTheDocument()
  })

  // **Und ohne Umfänge ebenso** — sie kommen aus Go, und eine leere Liste
  // hieße, dass diese Fassung keinen kennt.
  it('bietet nichts an, solange es keine Umfänge gibt', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Sammlungsfreigaben: () =>
        Promise.resolve({ freigaben: [], umfaenge: [], fehler: null }),
    })

    await zeige(aufklapper)

    expect(
      await screen.findByText(eigene.freigabenOhneZweck),
    ).toBeInTheDocument()

    expect(
      screen.queryByRole('button', { name: eigene.freigabenErteilen }),
    ).toBeNull()
  })

  it('sagt es, wenn das Gegenüber nichts sähe', async () => {
    belegeApi({
      Peerkontakte: () => Promise.resolve(eineListe([bestaetigt()])),
      Sammlungsfreigaben: () =>
        Promise.resolve(freigabenliste([eineFreigabe()])),
      Freigabesicht: () => Promise.resolve({ zeilen: [], fehler: null }),
    })

    const hand = await zeige(aufklapper)

    await hand.click(
      await screen.findByRole('button', { name: eigene.freigabenVorschau }),
    )

    expect(
      await screen.findByText(eigene.freigabenVorschauLeer),
    ).toBeInTheDocument()
  })
  // **Was gewählt war, bleibt gewählt** — sonst beantwortet die Vorschau
  // gleich darauf eine andere Frage als die, die eben erteilt wurde.
  //
  // **Gemessen an der Vorschau und nicht am Auswahlfeld**: dass dort wieder
  // der erste Zweck stünde, wäre eine Unschönheit; dass „Ansehen, was es
  // sieht" danach die **fremde** Sicht zeigt, ist die falsche Auskunft, für
  // die es diesen Kasten gibt (§10).
  it('zeigt nach dem Erteilen die Sicht des gewählten Zwecks', async () => {
    const gezeigt = vi.fn(() => Promise.resolve({ zeilen: [], fehler: null }))

    let stand: Sichtfreigabe[] = []

    belegeApi({
      Peerkontakte: () =>
        Promise.resolve(
          eineListe([
            einKontakt({ stand: 'bestaetigt', rechte: ['handel', 'teilen'] }),
          ]),
        ),
      Sammlungsfreigaben: () => Promise.resolve(freigabenliste(stand)),
      SammlungFreigeben: (freigabe: Sichtfreigabe) => {
        stand = [freigabe]

        return Promise.resolve(freigabenliste(stand))
      },
      Freigabesicht: gezeigt,
    })

    const hand = await zeige(aufklapper)

    await hand.selectOptions(
      await screen.findByLabelText(eigene.freigabenZweck),
      'teilen',
    )

    await hand.selectOptions(
      screen.getByLabelText(eigene.freigabenUmfang),
      'wuensche',
    )

    await hand.click(
      screen.getByRole('button', { name: eigene.freigabenErteilen }),
    )

    // **Auf die neue Zeile warten und nicht auf den Ruf** (§3): erst wenn sie
    // dasteht, ist nachgeladen und der Kasten fertig gezeichnet.
    expect(
      await screen.findByText(
        `${eigene.zweckTeilen} · ${eigene.gegenueber.umfaenge.wuensche}`,
      ),
    ).toBeInTheDocument()

    await hand.click(
      screen.getByRole('button', { name: eigene.freigabenVorschau }),
    )

    expect(gezeigt).toHaveBeenCalledWith('bgegenueber', 'teilen')
  })
})
