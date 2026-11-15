// **Ein Setwechsel räumt den Ausgabeassistenten** —
// #749.
//
// # Der Befund
//
// `SetVorlagenausfuhr` und `MTGJSONSichern` blieben beim Wechsel von `setID`
// **dieselbe** React-Komponente. Vorlagen, gewählte Vorlage, Optionen und die
// Erfolgsmeldung von Set A standen danach unter Set B — und eine laufende
// Ausfuhr trug ihre Antwort in die Seite ein, die inzwischen ein anderes Set
// zeigt.
//
// Damit galt weder Regel 1 aus §7 („ein Wechsel des Bezugsobjekts räumt den
// offenen Kontext") noch Regel 2 („eine Antwort gehört zu dem Stand, aus dem
// sie stammt"). Angestrichen vom Zweitblick durch Codex am 2026-08-21.
//
// # Warum ein Schlüssel und keine Aufräumwirkung
//
// **Weil er die Klasse schließt und nicht den Fall.** Ein `useEffect`, der
// sieben Zustände zurücksetzt, ist am Tag richtig, an dem er geschrieben wird —
// und am Tag falsch, an dem jemand den achten dazustellt. Ein `key` sagt
// React, dass ein anderes Set eine andere Sache ist; was danach dazukommt,
// wird mitgeräumt, ohne dass jemand daran denkt.
//
// **Anders als in `Karten.tsx`**, und der Unterschied hat einen Grund: dort
// muss die Komponente stehen bleiben — die Liste, die Auswahl, der offene
// Editor —, und geräumt wird über `useAktuellesSet` je Rückruf. Hier ist ein
// frisches Set eine frische Seite.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Ausgeben } from './Ausgeben'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.sets

// satz gibt einen Setkopf, wie die Fassade ihn liefert.
function satz(id: string, name: string) {
  return {
    id,
    name,
    spiel: 'magic',
    stil: 'm15',
    herkunft: `/pfad/${name}.mse-set`,
    karten: 2,
    herkunftFehlt: false,
  }
}

// Die zwei Vorlagen, die jedes Set hier anbietet.
const vorlagen = [
  {
    paket: 'magic-spoiler.mse-export-template',
    name: 'Spoiler',
    dateityp: '*.html|*.html',
    optionen: [],
  },
  {
    paket: 'magic-text.mse-export-template',
    name: 'Text',
    dateityp: '*.txt|*.txt',
    optionen: [],
  },
]

function zeige(teile: Parameters<typeof belegeApi>[0], setID = '4') {
  belegeApi({
    Sets: () =>
      Promise.resolve({ sets: [satz('4', 'Erstes'), satz('9', 'Zweites')] }),
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    Exportvorlagen: () => Promise.resolve({ vorlagen, befunde: [] }),
    ...teile,
  })

  return render(
    <Sprachraum sprache="de">
      <Ausgeben setID={setID} seite="ausgeben" wechseln={vi.fn()} />
    </Sprachraum>,
  )
}

// **Seit UXR-09 ein Assistent** (#1917): erst wird das Ziel gewählt, und die
// Exportvorlagen laden dann ohne weiteren Knopf. Ein frisches Set lädt sie
// darum **neu** — die Zusage ist nicht mehr „die Auswahl verschwindet",
// sondern „die Wahl von Set A steht nicht unter Set B".
describe('Ein Setwechsel im Ausgabeassistenten', () => {
  it('räumt die gewählte Vorlage', async () => {
    const { rerender } = zeige({})

    await userEvent.click(
      await screen.findByRole('radio', { name: texte.vorlageWaehlen }),
    )
    await userEvent.selectOptions(
      await screen.findByRole('combobox', { name: texte.vorlageWaehlen }),
      'magic-text.mse-export-template',
    )

    rerender(
      <Sprachraum sprache="de">
        <Ausgeben setID="9" seite="ausgeben" wechseln={vi.fn()} />
      </Sprachraum>,
    )

    // **Die erste Vorlage, wie beim ersten Laden** — und nicht die, die
    // unter Set A gewählt war.
    await waitFor(() => {
      expect(
        screen.getByRole('combobox', { name: texte.vorlageWaehlen }),
      ).toHaveValue('magic-spoiler.mse-export-template')
    })
  })

  // **Und die MTG-JSON-Meldung daneben** — dieselbe Klasse, derselbe Schlüssel.
  //
  // Ohne sie hielte dieser Wächter nur die halbe Zusage: die Flächen stehen
  // nebeneinander im Assistenten, und ein `key` an einer von ihnen sähe
  // genauso aus (§10).
  it('räumt die Meldung des MTG-JSON-Ausgangs', async () => {
    const { rerender } = zeige({
      MTGJSONSchreiben: () =>
        Promise.resolve({
          karten: 2,
          pfad: '/pfad/erstes.json',
          befunde: [],
          abgebrochen: false,
        }),
    })

    await userEvent.click(
      await screen.findByRole('radio', { name: texte.mtgjsonUeberschrift }),
    )
    await userEvent.click(
      await screen.findByRole('button', {
        name: `${texte.mtgjsonSichern} — Erstes`,
      }),
    )

    // **Zahl und Wort kommen aus `zaehlwort`** (#1219): der Satz lautet „2
    // Karten gesichert nach …", und bei genau einer „1 Karte".
    expect(
      await screen.findByText(
        `2 ${kataloge.de.allgemein.zaehlwoerter.karten.viele} ` +
          `${texte.mtgjsonGesichert} /pfad/erstes.json`,
      ),
    ).toBeInTheDocument()

    rerender(
      <Sprachraum sprache="de">
        <Ausgeben setID="9" seite="ausgeben" wechseln={vi.fn()} />
      </Sprachraum>,
    )

    // **Belegt, dass die Fläche des neuen Sets steht** — ihr Knopf nennt es —,
    // und erst dann die Verneinung (§3).
    expect(
      await screen.findByRole('button', {
        name: `${texte.mtgjsonSichern} — Zweites`,
      }),
    ).toBeInTheDocument()
    expect(screen.queryByText(/\/pfad\/erstes\.json/)).not.toBeInTheDocument()
  })

  // **Die Bilder hängen am Schlüssel und an nichts sonst.** Exportvorlage und
  // MTG JSON warten beim Setwechsel auf ihr Set und entstehen danach ohnehin
  // neu; die Bildausgabe braucht nur die Kennung — ohne `key` stünde der
  // Ordner von Set A unter Set B. **Gefunden über die Gegenprobe**: ohne den
  // Schlüssel blieben die beiden Fälle darüber grün.
  it('räumt den gewählten Bildordner', async () => {
    const { rerender } = zeige({
      EinstellungSetzen: () => Promise.resolve({}),
      Bildordner: () =>
        Promise.resolve({ ordner: '/tmp/erstes', abgebrochen: false }),
    })

    await userEvent.click(
      await screen.findByRole('radio', {
        name: kataloge.de.seiten.bilder.titel,
      }),
    )
    await userEvent.click(
      screen.getByRole('button', {
        name: kataloge.de.seiten.bilder.ordnerWaehlen,
      }),
    )
    expect(await screen.findByText('/tmp/erstes')).toBeInTheDocument()

    rerender(
      <Sprachraum sprache="de">
        <Ausgeben setID="9" seite="ausgeben" wechseln={vi.fn()} />
      </Sprachraum>,
    )

    expect(
      await screen.findByText(kataloge.de.seiten.bilder.keinOrdner),
    ).toBeInTheDocument()
    expect(screen.queryByText('/tmp/erstes')).not.toBeInTheDocument()
  })
})
