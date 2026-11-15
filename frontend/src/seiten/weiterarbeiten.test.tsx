// **Weiterarbeiten auf der Startseite** — Werkstück U11 von Phase 5.1 (#1422).
//
// # Was UXR-03 daran geändert hat
//
// #1911:
// hier standen **vier Listen** — angeheftet, zuletzt geöffnet, eingeführt,
// ausgegeben —, jede mit eigener Überschrift und eigenem Leerzustand. Ein Set,
// das jemand geöffnet und ausgegeben hat, stand zweimal da, angeheftet
// dreimal; mit dem, was die Ablage hergibt (sechs je Vorgang, `service.
// Hoechstens`), waren es bis zu **achtzehn Zeilen und sechsunddreißig
// Knöpfe**. Das Artboard *Main* zeigt **eine Reihe**: Angeheftetes oben, die
// Reihe darunter nach der Zeit.
//
// **Die Tests, die die vier Listen festschrieben** — „zeigt vier Listen", „sagt
// es je Liste, wenn dort nichts liegt" —, sind deshalb ersetzt: was sie
// zusagten, ist jetzt eine Zeile in der Reihe und ein Satz darunter.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it, vi } from 'vitest'
import { Weiterarbeiten } from './Weiterarbeiten'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi } from '../test/api'
import type {
  ArbeitsobjektDTO,
  ArbeitsschrittDTO,
  WeiterarbeitenAntwort,
} from '../apiweiterarbeiten'

const texte = kataloge.de.seiten.start.weiterarbeiten
const mehr = kataloge.de.seiten.start.aufmerksamkeit

function schritt(
  art: ArbeitsschrittDTO['objekt']['art'],
  kennung: string,
  name: string,
  vorgang: ArbeitsschrittDTO['vorgang'] = 'geoeffnet',
  zeitpunkt = '2026-09-09T10:00:00Z',
): ArbeitsschrittDTO {
  return {
    objekt: { art, kennung, name, angeheftet: false },
    vorgang,
    zeitpunkt,
  }
}

function angeheftet(
  art: ArbeitsobjektDTO['art'],
  kennung: string,
  name: string,
): ArbeitsobjektDTO {
  return { art, kennung, name, angeheftet: true }
}

// **Zeitpunkte, die sich unterscheiden** — die Reihe folgt der Zeit, und ein
// Test mit lauter gleichen sagte nichts über die Ordnung.
const gefuellt: WeiterarbeitenAntwort = {
  zuletzt: schritt(
    'set',
    '7',
    'Probeset',
    'ausgegeben',
    '2026-09-09T12:00:00Z',
  ),
  geoeffnet: [
    schritt('set', '7', 'Probeset', 'geoeffnet', '2026-09-09T11:00:00Z'),
    schritt('deck', '3', 'Mono Rot', 'geoeffnet', '2026-09-09T09:00:00Z'),
  ],
  eingefuehrt: [
    schritt('set', '9', 'Import', 'eingefuehrt', '2026-09-09T10:00:00Z'),
  ],
  ausgegeben: [
    schritt('set', '7', 'Probeset', 'ausgegeben', '2026-09-09T12:00:00Z'),
  ],
  angeheftet: [angeheftet('paket', 'magic-m15', 'M15')],
  fehler: null,
}

const leer: WeiterarbeitenAntwort = {
  zuletzt: null,
  geoeffnet: [],
  eingefuehrt: [],
  ausgegeben: [],
  angeheftet: [],
  fehler: null,
}

// **So viel, wie die Ablage hergibt** — sechs je Vorgang (`service.
// Hoechstens`), dazu drei angeheftete. Die Ausgangslage aus
// `docs/UIREVIEW.md` ist mit einem einzigen Probeset gemessen; dort sieht jede
// Fassung dieser Fläche ordentlich aus.
function realistisch(): WeiterarbeitenAntwort {
  const zeit = (stunde: number) =>
    `2026-09-09T${String(stunde).padStart(2, '0')}:00:00Z`
  const reihe = (vorgang: ArbeitsschrittDTO['vorgang'], ab: number) =>
    Array.from({ length: 6 }, (_, stelle) =>
      schritt(
        'set',
        `${vorgang}-${String(stelle)}`,
        `${vorgang} ${String(stelle)}`,
        vorgang,
        zeit(ab - stelle),
      ),
    )

  return {
    zuletzt: schritt(
      'set',
      'geoeffnet-0',
      'geoeffnet 0',
      'geoeffnet',
      zeit(23),
    ),
    geoeffnet: reihe('geoeffnet', 23),
    eingefuehrt: reihe('eingefuehrt', 17),
    ausgegeben: reihe('ausgegeben', 11),
    angeheftet: [
      angeheftet('set', 'a', 'Erstes'),
      angeheftet('deck', 'b', 'Zweites'),
      angeheftet('paket', 'c', 'Drittes'),
    ],
    fehler: null,
  }
}

async function hinstellen(antwort: WeiterarbeitenAntwort = gefuellt) {
  const Weiterarbeitenruf = vi.fn().mockResolvedValue(antwort)
  const Anheften = vi.fn().mockResolvedValue({ fehler: null })
  const Loesen = vi.fn().mockResolvedValue({ fehler: null })
  belegeApi({ Weiterarbeiten: Weiterarbeitenruf, Anheften, Loesen })
  const wechseln = vi.fn()
  const waehleSet = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Weiterarbeiten wechseln={wechseln} waehleSet={waehleSet} stand={0} />
    </Sprachraum>,
  )
  const bereich = await screen.findByRole('region', { name: texte.titel })

  return { Weiterarbeitenruf, Anheften, Loesen, wechseln, waehleSet, bereich }
}

// zeilen sind die Zeilen der Reihe, je mit ihrem Text.
function zeilen(bereich: HTMLElement): string[] {
  return within(within(bereich).getByRole('list'))
    .getAllByRole('listitem')
    .map((zeile) => zeile.textContent)
}

describe('Weiterarbeiten', () => {
  it('führt zum jüngsten Objekt', async () => {
    const { waehleSet } = await hinstellen()

    await userEvent.click(
      screen.getByRole('button', { name: 'Weiter an Set „Probeset“' }),
    )

    expect(waehleSet).toHaveBeenCalledWith('7')
    expect(screen.getByText(/Zuletzt ausgegeben am/)).toBeInTheDocument()
  })

  // **Eine Reihe, eine Überschrift** — vorher fünf: der Bereich und je Liste
  // eine.
  it('zeigt eine Reihe: Angeheftetes oben, dann nach der Zeit, jedes Objekt einmal', async () => {
    const { bereich } = await hinstellen()

    expect(within(bereich).getAllByRole('heading')).toHaveLength(1)

    const reihe = zeilen(bereich)

    expect(reihe).toHaveLength(4)
    expect(reihe[0]).toMatch(/^M15 Paket · angeheftet/)
    // **Der jüngste Schritt eines Objekts gilt**: das Probeset wurde um elf
    // geöffnet und um zwölf ausgegeben — es steht einmal da, als ausgegeben.
    expect(reihe[1]).toMatch(/^Probeset Set · ausgegeben · /)
    expect(reihe[2]).toMatch(/^Import Set · eingeführt · /)
    expect(reihe[3]).toMatch(/^Mono Rot Deck · geöffnet · /)
  })

  it('führt aus der Reihe zum Set, zum Deck und zum Paket', async () => {
    const { wechseln, waehleSet, bereich } = await hinstellen()

    await userEvent.click(
      within(bereich).getByRole('button', { name: /^Mono Rot/ }),
    )
    expect(wechseln).toHaveBeenCalledWith('decks')

    await userEvent.click(within(bereich).getByRole('button', { name: /^M15/ }))
    expect(wechseln).toHaveBeenCalledWith('vorlagen')

    await userEvent.click(
      within(bereich).getByRole('button', { name: /^Import/ }),
    )
    expect(waehleSet).toHaveBeenCalledWith('9')
  })

  it('heftet an und löst — und lädt danach nach', async () => {
    const { Anheften, Loesen, Weiterarbeitenruf, bereich } = await hinstellen()

    await userEvent.click(
      within(bereich).getByRole('button', { name: '„Probeset“ anheften' }),
    )
    expect(Anheften).toHaveBeenCalledWith('set', '7')

    await userEvent.click(
      within(bereich).getByRole('button', { name: '„M15“ lösen' }),
    )
    expect(Loesen).toHaveBeenCalledWith('paket', 'magic-m15')

    // **Nachgeladen wird nach dem Schreiben** — zwei Vorgänge, zwei Nachladen.
    await vi.waitFor(() => {
      expect(Weiterarbeitenruf).toHaveBeenCalledTimes(3)
    })
  })

  // **Vier, und der Rest auf Wunsch** — wie im Artboard *Main*, das vier
  // Zeilen zeigt. Mit dem, was die Ablage hergibt, standen vorher 21 Zeilen
  // und 42 Knöpfe untereinander. **Vier und nicht fünf wie die Aufmerksamkeit
  // daneben**: jede Zeile trägt hier zwei Knöpfe, und mit fünf stünden bei
  // vollem Bestand 22 Bedienelemente im Bild — R2 aus `docs/UIREVIEW.md`
  // erlaubt 20.
  it('zeigt vier Zeilen und die übrigen auf Wunsch', async () => {
    const { bereich } = await hinstellen(realistisch())

    expect(zeilen(bereich)).toHaveLength(4)
    // **Das Angeheftete steht in den vieren** — es ist das, was jemand oben
    // halten wollte.
    expect(zeilen(bereich).slice(0, 3).join(' ')).toMatch(
      /Erstes.*Zweites.*Drittes/,
    )

    await userEvent.click(
      within(bereich).getByRole('button', { name: 'Alle 21 zeigen' }),
    )
    expect(zeilen(bereich)).toHaveLength(21)

    await userEvent.click(
      within(bereich).getByRole('button', { name: mehr.weniger }),
    )
    expect(zeilen(bereich)).toHaveLength(4)
  })

  // **Eine führende Handlung** — R1 aus `docs/UIREVIEW.md`. Hier stand der
  // Knopf „Lösen" als gedrückter Umschalter, und ein gedrückter `Knopf` trägt
  // das Hauptgewicht (`bg-marke`). Mit drei angehefteten Objekten standen
  // drei Hauptknöpfe in einem Bereich, und „Weiter an …" war einer von vier.
  // **Die Messung zählt genau diese Klasse**; der Probebestand hatte nichts
  // angeheftet und sah es deshalb nicht.
  it('führt mit einem Hauptknopf, auch wenn vieles angeheftet ist', async () => {
    const { bereich } = await hinstellen(realistisch())

    const haupt = within(bereich)
      .getAllByRole('button')
      .filter((knopf) => knopf.classList.contains('bg-marke'))

    expect(haupt.map((knopf) => knopf.textContent)).toEqual([
      'Weiter an Set „geoeffnet 0“',
    ])
  })

  it('erklärt bei leerem Bestand den nächsten Schritt und bietet ihn an', async () => {
    const { wechseln } = await hinstellen(leer)

    expect(screen.getByText(texte.leer)).toBeInTheDocument()

    // **Und er trägt die Marke eines Leerzustands** (G02, #2249) — die
    // Zustandsprobe erkennt ihn daran; ohne sie hieß Start dort „ohne
    // Leerzustand", obwohl hier zwei Knöpfe stehen.
    expect(
      screen.getByText(texte.leer).closest('[data-leerzustand]'),
    ).not.toBeNull()
    await userEvent.click(
      screen.getByRole('button', { name: texte.handlungSets }),
    )
    expect(wechseln).toHaveBeenCalledWith('sets')
    await userEvent.click(
      screen.getByRole('button', { name: texte.handlungPakete }),
    )
    expect(wechseln).toHaveBeenCalledWith('pakete')
  })

  // **Wie man etwas oben hält, sagt ein Satz** — solange nichts oben steht.
  // Vorher hatte jede der vier Listen ihren eigenen Leerzustand.
  it('sagt, wie man anheftet, solange nichts angeheftet ist', async () => {
    const { bereich } = await hinstellen({
      ...leer,
      zuletzt: schritt('set', '7', 'Probeset'),
      geoeffnet: [schritt('set', '7', 'Probeset')],
    })

    expect(within(bereich).getByText(texte.angeheftetLeer)).toBeVisible()
    expect(within(bereich).queryByText(texte.leer)).not.toBeInTheDocument()
  })

  // **Angeheftetes ohne jüngsten Schritt** — die Ablage kann es: der Verlauf
  // ist leer, das Angeheftete bleibt. Dann gibt es keinen Schritt, an dem
  // „weiter" ansetzt, und auch keinen Satz, wie man anheftet.
  it('zeigt Angeheftetes, auch wenn kein Schritt vermerkt ist', async () => {
    const { bereich, wechseln } = await hinstellen({
      ...leer,
      angeheftet: [angeheftet('deck', '3', 'Mono Rot')],
    })

    expect(zeilen(bereich)).toEqual([expect.stringMatching(/^Mono Rot Deck/)])
    expect(
      within(bereich).queryByRole('button', { name: /^Weiter an/ }),
    ).not.toBeInTheDocument()
    expect(
      within(bereich).queryByText(texte.angeheftetLeer),
    ).not.toBeInTheDocument()

    await userEvent.click(
      within(bereich).getByRole('button', { name: /^Mono Rot/ }),
    )
    expect(wechseln).toHaveBeenCalledWith('decks')
  })

  it('meldet einen Fehler der Bindung', async () => {
    belegeApi({
      Weiterarbeiten: () =>
        Promise.resolve({
          ...leer,
          fehler: {
            code: 'unbekannt',
            schluessel: '',
            meldung: 'die Ablage schweigt',
          },
        }),
    })

    render(
      <Sprachraum sprache="de">
        <Weiterarbeiten wechseln={vi.fn()} waehleSet={vi.fn()} stand={0} />
      </Sprachraum>,
    )

    expect(await screen.findByRole('alert')).toHaveTextContent('schweigt')
  })

  // **Zweimal abgebrochen, ohne vorherigen Stand, zeigt sie nichts** — G11
  // (#2277): kein Ladehinweis, der nie endet, und keine Meldung über einen
  // Abbruch, der die Absicht war (`useLaden` holt genau einmal nach).
  it('steht nach zwei Abbrüchen ohne Stand still da', async () => {
    const abgebrochen = {
      code: 'abgebrochen' as const,
      schluessel: '',
      meldung: 'abgebrochen',
    }
    const ruf = vi.fn(() => Promise.resolve({ ...leer, fehler: abgebrochen }))
    belegeApi({ Weiterarbeiten: ruf })

    render(
      <Sprachraum sprache="de">
        <Weiterarbeiten wechseln={vi.fn()} waehleSet={vi.fn()} stand={0} />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(ruf).toHaveBeenCalledTimes(2)
    })
    await waitFor(() => {
      expect(screen.queryByRole('status')).toBeNull()
    })
    expect(screen.queryByRole('alert')).toBeNull()
  })
})
