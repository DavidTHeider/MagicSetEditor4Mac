// **Ausgeben ist ein Assistent: erst wohin, dann wie** — UXR-09
// (#1917),
// nach dem Artboard *Ausgeben*.
//
// # Der Befund
//
// Die Lasche war ein Wegweiser: fünf Flächen mit sieben Überschriften, zwei
// davon führten auf eine **andere** Lasche, eine war nur ein Satz. „Wohin?"
// beantwortete man, indem man die Lasche wechselte — und die Einstellungen
// des Druckbogens standen auf einer eigenen Lasche, ob man drucken wollte
// oder nicht.
//
// # Was das Issue prüft
//
// *„… ob jeder Schritt eine Frage stellt und ob Optionen des Druckbogens nur
// erscheinen, wenn der Druckbogen gewählt ist."* Beides steht hier.
import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import type { Ausgabeseite } from '../ausgabe'
import { Ausgeben } from './Ausgeben'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten

// **Die Namen der Ziele sind die, unter denen man sie schon kennt** — die
// Seiten *Bilder* und *Druckbogen*, die Exportvorlage und MTG JSON, wie sie
// auf der Setliste heißen. Neu ist nur die *Setdatei*.
const ziele = {
  setdatei: texte.ausgeben.ziele.setdatei.name,
  bilder: texte.bilder.titel,
  druckbogen: texte.druckbogen.titel,
  vorlage: texte.sets.vorlageWaehlen,
  mtgjson: texte.sets.mtgjsonUeberschrift,
}

function belege() {
  belegeApi({
    // **Der Assistent merkt sich das Ziel** (AP-10, #2367).
    EinstellungSetzen: () => Promise.resolve({}),
    Sets: () =>
      Promise.resolve({
        sets: [
          {
            id: '4',
            name: 'Mein Probeset',
            spiel: 'magic',
            stil: 'm15',
            herkunft: '/pfad/probeset.mse-set',
            karten: 2,
            herkunftFehlt: false,
          },
        ],
      }),
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    Exportvorlagen: () =>
      Promise.resolve({
        vorlagen: [
          {
            paket: 'magic-spoiler.mse-export-template',
            name: 'Spoiler',
            dateityp: '*.html|*.html',
            optionen: [],
          },
        ],
        befunde: [],
      }),
  })
}

function zeige(seite: Ausgabeseite = 'ausgeben') {
  const wechseln = vi.fn()

  belege()

  const { rerender } = render(
    <Sprachraum sprache="de">
      <Ausgeben setID="4" seite={seite} wechseln={wechseln} />
    </Sprachraum>,
  )

  const neu = (naechste: Ausgabeseite) => {
    rerender(
      <Sprachraum sprache="de">
        <Ausgeben setID="4" seite={naechste} wechseln={wechseln} />
      </Sprachraum>,
    )
  }

  return { wechseln, neu }
}

async function waehle(ziel: string) {
  await userEvent.click(await screen.findByRole('radio', { name: ziel }))
}

describe('Der erste Schritt fragt, wohin', () => {
  it('bietet fünf Ziele an und wählt keines vor', async () => {
    zeige()

    const wohin = await screen.findByRole('group', {
      name: texte.ausgeben.wohin,
    })
    const knoepfe = within(wohin).getAllByRole('radio')

    expect(knoepfe.map((knopf) => knopf.getAttribute('value'))).toEqual([
      'setdatei',
      'bilder',
      'druckbogen',
      'vorlage',
      'mtgjson',
    ])
    expect(
      knoepfe.filter((knopf) => (knopf as HTMLInputElement).checked),
    ).toHaveLength(0)

    // **Und daneben steht, was zu tun ist** — ein leerer Zustand, der den
    // nächsten Schritt nennt (§6).
    expect(screen.getByText(texte.ausgeben.keinZiel)).toBeInTheDocument()
  })

  // **Jedes Ziel sagt in einem Satz, wofür es ist** — der erste Satz dessen,
  // was bisher als Fläche dastand.
  it('sagt bei jedem Ziel, wofür es ist', async () => {
    zeige()

    expect(
      await screen.findByRole('radio', { name: ziele.druckbogen }),
    ).toHaveAccessibleDescription(
      'Die Karten nebeneinander auf Papier, mit Schnittmarken.',
    )
  })
})

describe('Der zweite fragt, wie', () => {
  // **Die Optionen des Druckbogens stehen nur, wenn er gewählt ist** — die
  // Frage aus dem Issue, wörtlich.
  it('zeigt die Optionen des Druckbogens nur für den Druckbogen', async () => {
    zeige()

    await waehle(ziele.druckbogen)

    const flaeche = screen.getByRole('region', { name: ziele.druckbogen })
    expect(
      within(flaeche).getByRole('combobox', { name: texte.druckbogen.papier }),
    ).toBeInTheDocument()

    await waehle(ziele.bilder)

    expect(
      screen.queryByRole('combobox', { name: texte.druckbogen.papier }),
    ).not.toBeInTheDocument()
    expect(
      screen.getByRole('textbox', { name: texte.bilder.vorlage }),
    ).toBeInTheDocument()
  })

  // **Eine führende Handlung je Region** (R1): der Knopf, der schreibt —
  // und im ersten Schritt keiner.
  it('führt mit genau einem Hauptknopf, und der schreibt', async () => {
    zeige()

    await waehle(ziele.druckbogen)

    const haupt = within(screen.getByRole('region', { name: ziele.druckbogen }))
      .getAllByRole('button')
      .filter((knopf) => knopf.classList.contains('bg-marke'))

    expect(haupt.map((knopf) => knopf.textContent)).toEqual([
      texte.druckbogen.legen,
    ])
    expect(
      within(
        screen.getByRole('group', { name: texte.ausgeben.wohin }),
      ).queryAllByRole('button'),
    ).toHaveLength(0)
  })

  // **Die Vorlagen stehen ohne weiteren Klick da.** Auf der Setliste lädt
  // erst ein Knopf sie — dort stünde sonst je Zeile ein Paketlauf. Hier hat
  // die Wahl „Exportvorlage" schon gesagt, was gewünscht ist.
  it('lädt die Exportvorlagen, sobald sie gewählt sind', async () => {
    zeige()

    await waehle(ziele.vorlage)

    expect(
      await screen.findByRole('combobox', { name: texte.sets.vorlageWaehlen }),
    ).toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: /Exportvorlage …/ }),
    ).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: texte.sets.vorlageSchliessen }),
    ).not.toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: texte.sets.vorlageAusfuehren }),
    ).toHaveClass('bg-marke')
  })

  // **Die Setdatei schreibt ⌘S** — hier steht, was sie enthält und was
  // nicht; ein zweiter Sichernknopf neben dem der Setliste wäre ein zweiter
  // Vorgang für dasselbe Set (§7, Regel 4).
  it('sagt bei der Setdatei, wie sie entsteht', async () => {
    zeige()

    await waehle(ziele.setdatei)

    expect(
      within(screen.getByRole('region', { name: ziele.setdatei })).getByText(
        /Die Taste dafür ist ⌘S/,
      ),
    ).toBeInTheDocument()
  })
})

describe('Was gewählt ist, bleibt', () => {
  // **Zurück zu einem Ziel heißt: zu dem, was man dort eingestellt hat**
  // (S11.NAT-03: „Aufklappen verändert keine Eingaben").
  it('behält die Einstellungen eines Ziels beim Wechsel', async () => {
    zeige()

    await waehle(ziele.druckbogen)
    await userEvent.selectOptions(
      screen.getByRole('combobox', { name: texte.druckbogen.papier }),
      'letter',
    )

    await waehle(ziele.bilder)
    await waehle(ziele.druckbogen)

    expect(
      screen.getByRole('combobox', { name: texte.druckbogen.papier }),
    ).toHaveValue('letter')
  })

  // **⌥⌘0 und „Ablage › Drucken" führen auf den Druckbogen** — die Seite
  // `druckbogen` ist der Wunsch, und er wählt das Ziel vor.
  it('wählt das Ziel vor, das die Seite verlangt', async () => {
    zeige('druckbogen')

    expect(
      await screen.findByRole('radio', { name: ziele.druckbogen }),
    ).toBeChecked()
  })

  // **Und ein neuer Wunsch zählt, ein alter nicht.** Wer auf der Seite
  // `druckbogen` die Exportvorlage wählt, kommt auf `ausgeben` zurück — dort
  // heißt die Fläche, wie sie ist. Kommt danach ⌥⌘1, gilt die Bilderwahl.
  it('nimmt einen neuen Wunsch an und überstimmt keine Wahl mit einem alten', async () => {
    const { wechseln, neu } = zeige('druckbogen')

    await waehle(ziele.vorlage)

    expect(wechseln).toHaveBeenCalledWith('ausgeben')
    expect(screen.getByRole('radio', { name: ziele.vorlage })).toBeChecked()

    act(() => {
      neu('ausgeben')
    })
    expect(screen.getByRole('radio', { name: ziele.vorlage })).toBeChecked()

    act(() => {
      neu('bilder')
    })
    expect(screen.getByRole('radio', { name: ziele.bilder })).toBeChecked()
  })

  // **Auf `ausgeben` selbst wechselt eine Wahl keine Seite** — sonst stünde
  // jeder Klick in der Seitengeschichte. Die Bilderfläche danach belegt, dass
  // die Wahl angekommen ist.
  it('wechselt auf Ausgeben selbst keine Seite', async () => {
    const { wechseln } = zeige()

    await waehle(ziele.bilder)

    expect(
      screen.getByRole('region', { name: ziele.bilder }),
    ).toBeInTheDocument()
    expect(wechseln).not.toHaveBeenCalled()
  })
})
