// Die Paketliste **als Verwaltung** —
// #244.
//
// Eine eigene Datei neben `paketliste.test.tsx`, weil beide sonst über 400
// Zeilen gingen (§2). Der Schnitt folgt der Frage: dort steht, dass der
// Katalog lädt, installiert und entfernt; hier, dass man in ihm **etwas
// findet**.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { AbhaengigkeitDTO, PaketDTO } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { PaketlisteMitBefunden } from '../test/abhaengigkeiten'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

// **Die Anlage steht auf derselben Seite** (#1350) und liest beim Aufbau,
// welche Arten sich anlegen lassen. Ohne diese Belegung scheiterte die
// Vorrichtung mit „Die Methode PaketArten ist nicht belegt" — und die
// Ablehnung stünde als unbehandelt im Lauf.
const anlegbar = () =>
  Promise.resolve([
    { art: 'mse-game', brauchtSpiel: false },
    { art: 'mse-style', brauchtSpiel: true },
  ])

const texte = kataloge.de

function paket(teile: Partial<PaketDTO> & { name: string }): PaketDTO {
  return {
    art: 'mse-style',
    anzeigename: teile.name,
    version: '',
    pfad: '',
    herkunft: '',
    pruefsumme: '',
    braucht: [],
    gruppe: '',
    rang: 1000000,
    artrang: 30,
    herkunftsart: 'mitgeliefert' as const,
    hauptdatei: 'style',
    geaendert: false,
    gebrauchtVon: [],
    benutzt: 0,
    ...teile,
  }
}

const bestand = [
  paket({ name: 'magic', art: 'mse-game', artrang: 20, gruppe: 'magic' }),
  paket({ name: 'magic-m15', gruppe: 'magic/m15 style' }),
  paket({ name: 'magic-m15-planeswalker', gruppe: 'magic/m15 style' }),
  paket({ name: 'vtes-standard', gruppe: 'vtes' }),
]

function zeige(pakete: PaketDTO[] = bestand, befunde: AbhaengigkeitDTO[] = []) {
  belegeApi({
    PaketArten: anlegbar,
    Pakete: vi.fn(() => Promise.resolve({ pakete, befunde: [] })),
    Abhaengigkeiten: vi.fn(() => Promise.resolve({ befunde })),
  })

  render(
    <Sprachraum sprache="de">
      <PaketlisteMitBefunden />
    </Sprachraum>,
  )
}

describe('Die Paketverwaltung', () => {
  // **Aufgeklappt ist der Anfangszustand.**
  //
  //	package_list->expandAll();
  //
  // `src/gui/packages_window.cpp`, Zeile 222, Fassung f9d9356 vom 2021-07-06.
  //
  // Der erste Anlauf hier machte es umgekehrt und begründete es mit der Länge
  // — über die 324 Pakete des installierten `M15-Magic-Pack` wären es rund 480
  // Zeilen. Das ist wahr und war trotzdem der falsche Schluss: eine
  // gruppierte, eingerückte, geordnete Liste ist **nicht** dieselbe wie die
  // flache von vorher, und wer sie zuklappt, verlangt von jedem Suchenden
  // erst einen Klick.
  it('zeigt anfangs jede Zeile', async () => {
    zeige()

    expect(
      await screen.findByRole('row', { name: 'Gruppe magic' }),
    ).toBeVisible()
    expect(screen.getByText('magic-m15')).toBeInTheDocument()
    expect(screen.getByText('vtes-standard')).toBeInTheDocument()
  })

  // **Die Zeile ist der Schalter, kein Knopf darin** — UXR-17 (#1925): ein
  // Knopf je Gruppe zählte die Hüllenprobe als Bedienelement, bei Hunderten
  // Gruppen (R6, `docs/UIREVIEW.md`). Klick und `Enter` bleiben dieselben.
  it('klappt eine Gruppe zu und wieder auf', async () => {
    zeige()

    // **Der Name ist „Gruppe magic", nicht „▾ magic".** Der Pfeil trägt
    // `aria-hidden` und steht deshalb nicht im zugänglichen Namen — was ein
    // Vorleseprogramm hört, ist genau das hier.
    const gruppe = await screen.findByRole('row', { name: 'Gruppe magic' })
    expect(gruppe).toHaveAttribute('aria-expanded', 'true')
    expect(screen.queryByRole('button', { name: 'Gruppe magic' })).toBeNull()

    await userEvent.click(gruppe)

    expect(screen.getByRole('row', { name: 'Gruppe magic' })).toHaveAttribute(
      'aria-expanded',
      'false',
    )

    // **Alles darunter ist fort — auch die Enkel.** `vtes` daneben bleibt.
    expect(screen.queryByText('magic-m15')).not.toBeInTheDocument()
    expect(screen.getByText('vtes-standard')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('row', { name: 'Gruppe magic' }))

    expect(screen.getByText('magic-m15')).toBeInTheDocument()
  })

  // **Mit der Tastatur allein** (§6, und ein Abnahmekasten aus #244).
  //
  // Der Weg ist: Suchfeld → Art → Ordnung → erste Gruppe. Dann `Enter`.
  //
  // **Und dieser Test steht hier, weil die Sichtprüfung ihn nicht führen
  // konnte**: der synthetische Tastendruck des Browserwerkzeugs erreicht den
  // Knopf (`keydown` kommt an), löst ihn aber nicht aus — ein `click` folgt
  // nicht. Was dort zu sehen war, ist der Fokusring; dass `Enter` schaltet,
  // steht hier.
  // **Eine Tabstelle, und die Pfeiltasten bewegen darin** (#1290).
  //
  // Gemessen mit `scripts/serienmass.mjs` bei 1280 × 800: **413** Tabstellen
  // auf dieser Seite, bei 331 Paketen. Die Begründung in `Paketzeilen.tsx`
  // sagte „Tab läuft die Knöpfe ab" — das stimmt und war bei 331 Zeilen eine
  // Sperre.
  it('bewegt den Fokus mit den Pfeiltasten zwischen den Zeilen', async () => {
    zeige()

    await screen.findByRole('row', { name: 'Gruppe magic' })

    // **Nur die Zeilen des Rumpfes.** `getAllByRole('row')` nimmt die
    // Kopfzeile mit, und die ist keine, die man anspringt.
    const zeilen = screen
      .getAllByRole('row')
      .filter((z) => z.closest('tbody') !== null)
    const tabstellen = zeilen.filter((z) => z.getAttribute('tabindex') === '0')

    expect(tabstellen).toHaveLength(1)

    tabstellen[0]?.focus()
    await userEvent.keyboard('{ArrowDown}')

    expect(document.activeElement).toBe(zeilen[1])

    await userEvent.keyboard('{ArrowUp}')
    expect(document.activeElement).toBe(zeilen[0])

    // **Am Rand bleibt er stehen** — es gibt keinen Umlauf.
    await userEvent.keyboard('{ArrowUp}')
    expect(document.activeElement).toBe(zeilen[0])

    await userEvent.keyboard('{End}')
    expect(document.activeElement).toBe(zeilen[zeilen.length - 1])
  })

  // **Und → führt in die Zeile.**
  it('erreicht den Knopf einer Paketzeile mit →', async () => {
    zeige()

    await screen.findByRole('row', { name: 'Gruppe magic' })

    const paketzeile = screen
      .getAllByRole('row')
      .find((z) => z.getAttribute('aria-label') === 'magic-m15')

    paketzeile?.focus()
    await userEvent.keyboard('{ArrowRight}')

    expect(document.activeElement?.tagName).toBe('BUTTON')
  })

  it('klappt eine Gruppe mit der Tastatur zu', async () => {
    zeige()

    await screen.findByRole('row', { name: 'Gruppe magic' })

    screen
      .getByRole('searchbox', { name: texte.seiten.pakete.katalogSuche })
      .focus()

    // **Durch die Filter des Navigators bis zur ersten Zeile** — seit
    // DOM-01 (#1696) stehen dort auch die Herkunftsknöpfe; gezählt wird
    // deshalb nicht, sondern gegangen, bis der Fokus in einer Zeile ist.
    for (
      let schritt = 0;
      schritt < 15 && document.activeElement?.tagName !== 'TR';
      schritt += 1
    ) {
      await userEvent.tab()
    }

    // **Der Tabulator landet auf der Zeile**, seit die Tabelle ein Gitter ist
    // (#1290): eine Tabstelle statt 413 auf dieser Seite. Die Gruppe schaltet
    // dann `Enter` auf der Zeile um — seit UXR-17 ist sie selbst der Schalter.
    expect(document.activeElement).toHaveAttribute('aria-label', 'Gruppe magic')

    await userEvent.keyboard('{Enter}')

    expect(document.activeElement).toHaveAttribute('aria-expanded', 'false')
  })

  // **Der eine Abnahmekasten aus #244**: ein Paket unter 329 lässt sich in
  // einem Zug finden. „In einem Zug" heißt: tippen, sehen — ohne vorher zu
  // wissen, in welcher Gruppe es steckt.
  it('findet ein Paket über die Suche, ohne dass jemand aufklappt', async () => {
    zeige()

    await screen.findByRole('row', { name: 'Gruppe magic' })
    await userEvent.type(
      screen.getByRole('searchbox', { name: texte.seiten.pakete.katalogSuche }),
      'planeswalker',
    )

    expect(screen.getByText('magic-m15-planeswalker')).toBeInTheDocument()
    expect(screen.queryByText('vtes-standard')).not.toBeInTheDocument()
  })

  // **Beim Suchen ist die Gruppe kein Knopf mehr.** Der erste Anlauf ließ ihn
  // stehen: der Pfeil zeigte „zugeklappt", die Kinder standen darunter, und
  // ein Klick änderte sichtbar nichts. Gesehen hat das die Sichtprüfung, nicht
  // ein Test — deshalb steht der Test jetzt hier.
  it('macht die Gruppenzeile beim Suchen zur Beschriftung', async () => {
    zeige()

    await screen.findByRole('row', { name: 'Gruppe magic' })
    await userEvent.type(
      screen.getByRole('searchbox', { name: texte.seiten.pakete.katalogSuche }),
      'planeswalker',
    )

    expect(screen.getByText('magic-m15-planeswalker')).toBeInTheDocument()
    expect(
      screen.getByRole('row', { name: 'Gruppe magic' }),
    ).not.toHaveAttribute('aria-expanded')
  })

  it('nennt, wie viele von wie vielen gefunden sind', async () => {
    zeige()

    await screen.findByRole('row', { name: 'Gruppe magic' })
    await userEvent.type(
      screen.getByRole('searchbox', { name: texte.seiten.pakete.katalogSuche }),
      'planeswalker',
    )

    expect(screen.getByRole('status')).toHaveTextContent('1 von 4')
  })

  it('erklärt eine leere Suche, statt nichts zu zeigen', async () => {
    zeige()

    await screen.findByRole('row', { name: 'Gruppe magic' })
    await userEvent.type(
      screen.getByRole('searchbox', { name: texte.seiten.pakete.katalogSuche }),
      'gibtesnicht',
    )

    expect(
      screen.getByText(texte.seiten.pakete.katalogNichtsGefunden),
    ).toBeInTheDocument()
  })

  it('filtert nach der Art', async () => {
    zeige()

    await screen.findByRole('row', { name: 'Gruppe magic' })
    await userEvent.selectOptions(
      screen.getByRole('combobox', { name: texte.seiten.pakete.katalogArt }),
      'mse-game',
    )

    expect(screen.getByRole('status')).toHaveTextContent('1 von 4')
  })

  // **Flach heißt: keine Gruppen, alles sichtbar.** Wer nach Version ordnet,
  // fragt „was ist alt" — darauf gibt ein Baum keine Antwort, und ein Baum,
  // der dabei zugeklappt bliebe, gäbe gar keine.
  it('zeigt in der flachen Ordnung jedes Paket ohne Aufklappen', async () => {
    zeige()

    await screen.findByRole('row', { name: 'Gruppe magic' })
    await userEvent.selectOptions(
      screen.getByRole('combobox', {
        name: texte.seiten.pakete.katalogOrdnung,
      }),
      'name',
    )

    expect(screen.getAllByRole('row')).toHaveLength(5)
    expect(screen.getByText('vtes-standard')).toBeInTheDocument()
  })

  // **Der Zustand steht an der Zeile** — #244 nennt es ausdrücklich: heute
  // stand er in einem eigenen Abschnitt daneben (#57), und wer die Liste las,
  // sah nicht, welche Zeile gemeint war.
  it('zeigt an der Zeile, dass ein Paket Vorbehalte hat', async () => {
    zeige(
      [paket({ name: 'magic-m15' })],
      [
        {
          paket: 'magic-m15.mse-style',
          braucht: 'magic.mse-game',
          art: 'fehlt',
          verlangt: '',
          da: '',
        },
      ],
    )

    const zeile = await screen.findByRole('row', { name: /magic-m15/ })

    expect(
      within(zeile).getByText(texte.seiten.pakete.katalogEinVorbehalt),
    ).toBeInTheDocument()
  })

  // **Einer und mehrere sind zwei Sätze**, nicht einer mit „(e)". Deutsch
  // bildet den Plural, und „1 Vorbehalt(e)" ist die Form, an der man sieht,
  // dass jemand ihn nicht bilden wollte.
  it('zählt mehrere Vorbehalte in der Mehrzahl', async () => {
    zeige(
      [paket({ name: 'magic-m15' })],
      [
        {
          paket: 'magic-m15.mse-style',
          braucht: 'magic.mse-game',
          art: 'fehlt',
          verlangt: '',
          da: '',
        },
        {
          paket: 'magic-m15.mse-style',
          braucht: 'magic-mana.mse-symbol-font',
          art: 'zualt',
          verlangt: '2021-01-01',
          da: '2009-07-23',
        },
      ],
    )

    const zeile = await screen.findByRole('row', { name: /magic-m15/ })

    expect(within(zeile).getByText('2 Vorbehalte')).toBeInTheDocument()
  })

  it('lässt eine Zeile ohne Vorbehalt in Ruhe', async () => {
    zeige([paket({ name: 'magic-m15', gruppe: 'magic' })], [])

    const zeile = await screen.findByRole('row', { name: /magic-m15/ })

    expect(within(zeile).queryByText(/Vorbehalt/)).not.toBeInTheDocument()
  })

  // **Eine leere Angabe wird benannt, nicht weggelassen.** Eine leere Zelle
  // sagt „hier steht nichts"; sie sagt nicht, ob das Paket keine Version hat
  // oder ob die Liste kaputt ist.
  it('benennt eine fehlende Version und die Herkunft', async () => {
    zeige([paket({ name: 'schlicht', gruppe: 'g' })])

    expect(
      await screen.findByRole('cell', {
        name: texte.seiten.pakete.katalogOhneVersion,
      }),
    ).toBeInTheDocument()
    // **Die Herkunft ist seit DOM-01 eine Art** und nie leer (#1696) — sie
    // steht im `title` des Stands.
    expect(
      screen.getByRole('cell', {
        name: texte.seiten.pakete.plan.standEingebunden,
      }),
    ).toHaveAttribute(
      'title',
      texte.seiten.pakete.plan.herkunftsarten.mitgeliefert,
    )
  })
})
