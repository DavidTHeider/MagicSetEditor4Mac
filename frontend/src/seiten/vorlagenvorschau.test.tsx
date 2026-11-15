// **Die Vorschau einer ungespeicherten Vorlage** — Werkstück 3 von Phase 3.3
// (#1164).
import { render, screen, waitFor } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { Vorlagenvorschau } from './Vorlagenvorschau'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { Feldapi } from '../apifeld'
import type { VorschauAntwort, VorschauDTO } from '../api'

const texte = kataloge.de.seiten.vorlagen

// bild ist eine Vorschau, wie die Bindung sie gibt.
const bild: VorschauDTO = {
  stil: '',
  gezeichnet: '',
  bild: '/vorschau/AAAA.png',
  breite: 375,
  hoehe: 523,
  dpi: 150,
  befunde: [],
  flaechen: [],
  nennungen: [],
  vorgaben: [],
}

// zeichne belegt die Bindung mit einer festen Antwort.
function zeichne(antwort: VorschauAntwort): Feldapi['Vorlagenvorschau'] {
  return () => Promise.resolve(antwort)
}

describe('Die Vorlagenvorschau', () => {
  afterEach(raeumeApi)

  it('zeichnet mit dem Text, der im Feld steht', async () => {
    const gezeichnet = vi.fn<Feldapi['Vorlagenvorschau']>(
      zeichne({ vorschau: bild, fehler: null }),
    )

    belegeApi({ Vorlagenvorschau: gezeichnet })

    render(
      <Sprachraum sprache="de">
        <Vorlagenvorschau
          paket="/pakete/magic-m15.mse-style"
          datei="style"
          inhalt="mse version: 2.0.0"
          karteID=""
        />
      </Sprachraum>,
    )

    expect(await screen.findByAltText(texte.vorschau)).toBeInTheDocument()

    // **Der ungespeicherte Text geht mit** — das ist die ganze Zusage.
    expect(gezeichnet).toHaveBeenCalledWith(
      '/pakete/magic-m15.mse-style',
      'style',
      'mse version: 2.0.0',
      1,
      '',
    )
  })

  // **Eine Änderung zeichnet neu, ohne dass jemand bestätigt** (§6).
  it('zeichnet nach einer Änderung neu', async () => {
    const gezeichnet = vi.fn<Feldapi['Vorlagenvorschau']>(
      zeichne({ vorschau: bild, fehler: null }),
    )

    belegeApi({ Vorlagenvorschau: gezeichnet })

    const { rerender } = render(
      <Sprachraum sprache="de">
        <Vorlagenvorschau paket="p" datei="style" inhalt="eins" karteID="" />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(gezeichnet).toHaveBeenCalledWith('p', 'style', 'eins', 1, '')
    })

    rerender(
      <Sprachraum sprache="de">
        <Vorlagenvorschau paket="p" datei="style" inhalt="zwei" karteID="" />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(gezeichnet).toHaveBeenCalledWith('p', 'style', 'zwei', 1, '')
    })
  })

  // **Und `nichtGefunden` allein ist noch keine Auskunft über den Bestand.**
  //
  // # Warum der Schlüssel entscheidet und nicht der Code
  //
  // `nichtGefunden` steht für jedes `ErrNotFound` — eine fehlende Karte, ein
  // fehlendes Paket, ein fehlender Stil. Wer daran „kein offenes Set benutzt
  // diese Vorlage" festmacht, schickt bei jedem anderen davon jemanden ein Set
  // öffnen, das nicht das Problem ist. Angestrichen vom Nachlauf des
  // Zweitblicks über Phase 3.3.
  it('zeigt einen anderen nichtGefunden-Fehler als Fehler', async () => {
    belegeApi({
      Vorlagenvorschau: zeichne({
        vorschau: null,
        fehler: {
          code: 'nichtGefunden',
          schluessel: 'nichtGefunden',
          meldung: 'die Karte gibt es nicht mehr',
        },
      }),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagenvorschau paket="p" datei="style" inhalt="x" karteID="" />
      </Sprachraum>,
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
    expect(screen.queryByText(texte.vorschauKeineKarte)).not.toBeInTheDocument()
  })

  // **Ein anderer Fehler ist kein leerer Zustand.**
  //
  // # Warum das die eigentliche Zusage ist
  //
  // „Kein offenes Set benutzt diese Vorlage" sagt, was zu tun ist. Ein
  // Lesefehler im Bestand, ein geschlossener Bestand, ein Fehler beim
  // Zeichnen sagen etwas **anderes** — und wer sie in denselben Satz
  // übersetzt, schickt jemanden ein Set öffnen, das längst offen ist.
  //
  // Der erste Anlauf zeigte den Satz bei **jedem** Fehler; nur der Lauf
  // darunter prüfte ihn, und der schickt `nichtGefunden`. Angestrichen vom
  // Zweitblick durch Codex über Phase 3.3.
  it('zeigt einen anderen Fehler als Fehler', async () => {
    belegeApi({
      Vorlagenvorschau: zeichne({
        vorschau: null,
        fehler: {
          code: 'konflikt',
          schluessel: 'bestandGeschlossen',
          meldung: 'der Bestand ist geschlossen',
        },
      }),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagenvorschau paket="p" datei="style" inhalt="x" karteID="" />
      </Sprachraum>,
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      'der Bestand ist geschlossen',
    )
    expect(screen.queryByText(texte.vorschauKeineKarte)).not.toBeInTheDocument()
  })

  // **Ohne passendes Set steht dort, was zu tun ist** (§6: leere Zustände
  // erklären den nächsten Schritt) — und keine Fehlermeldung.
  it('sagt es, wenn kein Set diese Vorlage benutzt', async () => {
    belegeApi({
      Vorlagenvorschau: zeichne({
        vorschau: null,
        fehler: {
          code: 'nichtGefunden',
          schluessel: 'keinSetMitStil',
          meldung: 'kein Set',
        },
      }),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagenvorschau paket="p" datei="style" inhalt="x" karteID="" />
      </Sprachraum>,
    )

    expect(
      await screen.findByText(texte.vorschauKeineKarte),
    ).toBeInTheDocument()
    expect(screen.queryByAltText(texte.vorschau)).not.toBeInTheDocument()
  })

  // **Die Befunde stehen neben dem Bild** (§8): eine Vorlage, die halb aufgeht,
  // zeigt eine Karte und sagt, was fehlte.
  it('zeigt die Befunde neben dem Bild', async () => {
    belegeApi({
      Vorlagenvorschau: zeichne({
        vorschau: { ...bild, befunde: ['die Schrift fehlt'] },
        fehler: null,
      }),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagenvorschau paket="p" datei="style" inhalt="x" karteID="" />
      </Sprachraum>,
    )

    expect(await screen.findByText('die Schrift fehlt')).toBeInTheDocument()
  })
})

describe('Die Vorlagenvorschau in der Werkstatt', () => {
  afterEach(raeumeApi)

  it('reicht die gewählte Testkarte an die Bindung', async () => {
    const gezeichnet = vi.fn<Feldapi['Vorlagenvorschau']>(
      zeichne({
        vorschau: {
          stil: '',
          gezeichnet: '',
          bild: 'data:image/png;base64,',
          breite: 1,
          hoehe: 1,
          dpi: 150,
          befunde: [],
          flaechen: [],
          nennungen: [],
          vorgaben: [],
        },
        fehler: null,
      }),
    )
    belegeApi({ Vorlagenvorschau: gezeichnet })

    render(
      <Sprachraum sprache="de">
        <Vorlagenvorschau paket="p" datei="style" inhalt="x" karteID="7" />
      </Sprachraum>,
    )

    await screen.findByAltText(texte.vorschau)

    expect(gezeichnet).toHaveBeenCalledWith('p', 'style', 'x', 1, '7')
  })

  it('markiert die letzte gültige Vorschau als veraltet', async () => {
    belegeApi({
      Vorlagenvorschau: zeichne({
        vorschau: {
          stil: '',
          gezeichnet: '',
          bild: 'data:image/png;base64,',
          breite: 1,
          hoehe: 1,
          dpi: 150,
          befunde: [],
          flaechen: [],
          nennungen: [],
          vorgaben: [],
        },
        fehler: null,
      }),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagenvorschau
          paket="p"
          datei="style"
          inhalt="x"
          karteID=""
          veraltet={{ zeile: 12, spalte: 3 }}
        />
      </Sprachraum>,
    )

    expect(await screen.findByAltText(texte.vorschau)).toBeInTheDocument()
    expect(
      screen.getByText(
        'Veraltet — die Datei hat einen Befund in Zeile 12; gezeigt wird der letzte gültige Stand.',
      ),
    ).toBeInTheDocument()
  })
})
