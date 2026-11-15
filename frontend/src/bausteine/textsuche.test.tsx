// **Suchen und Ersetzen im Feld** — Werkstück 1 von Phase 3.3
// (#1164).
//
// # Warum im Feld und nicht über die Seitensuche
//
// ⌘F führt in dieser Anwendung zur Volltextsuche über die Karten (§6). In
// einem Textfeld heißt dieselbe Taste seit je „suche **hier**" — in jedem
// Editor, den es gibt. Das ist keine gebrochene Gewohnheit, sondern die
// stärkere: wer in einer 400-Zeilen-Vorlage steht, meint nicht die Karten.
//
// # Warum Groß- und Kleinschreibung zählt
//
// In MSE-Script sind `card` und `Card` zweierlei. Ein Ersetzen, das darüber
// hinwegsähe, änderte still etwas anderes als gemeint.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Quelltextfeld } from './Quelltextfeld'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { merkeEscape } from '../test/escape'

const texte = kataloge.de.bausteine.suche

// zeige stellt ein Feld mit einem Text und einer Merkstelle für Änderungen.
function zeige(wert: string) {
  belegeApi({ Skriptfarben: () => Promise.resolve({ merkmale: [] }) })

  const gesetzt = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Quelltextfeld
        bezeichnung="Inhalt"
        wert={wert}
        setzeWert={gesetzt}
        sprache="keine"
      />
    </Sprachraum>,
  )

  return gesetzt
}

// oeffneSuche schlägt die Suche über ⌘F auf.
async function oeffneSuche() {
  await userEvent.click(screen.getByLabelText('Inhalt'))
  await userEvent.keyboard('{Meta>}f{/Meta}')

  return screen.findByLabelText(texte.suchen)
}

describe('Die Suche im Quelltextfeld', () => {
  afterEach(raeumeApi)

  // **⌃⌘F gehört dem Vollbild** —
  // #1994.
  // ⌘ und ⌃ zusammen sind keines unserer Kürzel; erst danach öffnet ⌘F —
  // der Beleg, dass das Feld zuhört.
  it('lässt ⌃⌘F dem System', async () => {
    zeige('card.name')

    await userEvent.click(screen.getByLabelText('Inhalt'))
    await userEvent.keyboard('{Control>}{Meta>}f{/Meta}{/Control}')

    expect(screen.queryByLabelText(texte.suchen)).not.toBeInTheDocument()

    await userEvent.keyboard('{Meta>}f{/Meta}')

    expect(await screen.findByLabelText(texte.suchen)).toBeInTheDocument()
  })

  it('geht über ⌘F auf und über Esc wieder zu', async () => {
    zeige('card.name')

    const feld = await oeffneSuche()

    expect(feld).toBeInTheDocument()

    await userEvent.keyboard('{Escape}')

    await waitFor(() => {
      expect(screen.queryByLabelText(texte.suchen)).not.toBeInTheDocument()
    })
  })

  // **Das Escape, das die Suche schließt, ist verbraucht** — in der Konsole,
  // die in der Schublade *Werkzeuge* steht, schlösse es sonst die Schublade
  // gleich mit (UX-08, #1680).
  it('verbraucht das Escape, mit dem sie schließt', async () => {
    zeige('card.name')

    await oeffneSuche()

    const verbraucht = merkeEscape()
    await userEvent.keyboard('{Escape}')

    expect(verbraucht()).toBe(true)
  })

  // **⌘F und ⌃F tun dasselbe.**
  //
  // Unter macOS ist ⌘ die Taste; ⌃ steht daneben, weil dieselbe Anwendung
  // unter Windows und Linux übersetzt wird (§12) und dort ⌃F gilt. Ohne diese
  // Zusicherung wäre der zweite Zweig unbetreten — und niemand merkte es hier.
  it('geht auch über Strg+F auf', async () => {
    zeige('card.name')

    await userEvent.click(screen.getByLabelText('Inhalt'))
    await userEvent.keyboard('{Control>}f{/Control}')

    expect(await screen.findByLabelText(texte.suchen)).toBeInTheDocument()
  })

  // **⌘F kommt nicht am Fenster an.**
  //
  // # Warum das eine eigene Zusicherung braucht
  //
  // Der Rahmen horcht auf ⌘F am **Fenster** und schickt zur Volltextsuche über
  // die Karten (§6). Der erste Anlauf rief nur `preventDefault` — das hält den
  // Browser auf, nicht den zweiten Horcher. Bei der Sichtprüfung sprang die
  // Seite weg, während der Schreibzeiger in der Vorlage stand.
  //
  // **Die Läufe darüber sahen das nicht** und konnten es nicht: sie stellen
  // das Feld allein auf, ohne den Rahmen darum.
  it('lässt ⌘F nicht ans Fenster durch', async () => {
    zeige('card.name')

    // **Nur das `f` zählt.** `{Meta>}f{/Meta}` schickt drei Ereignisse; das
    // Meta selbst kommt am Fenster an und soll es auch — abgefangen wird die
    // Tastenkombination, nicht jede Taste.
    const amFenster = vi.fn()
    const nurF = (ereignis: KeyboardEvent) => {
      if (ereignis.key === 'f' || ereignis.key === 'k') {
        amFenster(ereignis.key)
      }
    }

    window.addEventListener('keydown', nurF)

    try {
      await userEvent.click(screen.getByLabelText('Inhalt'))
      await userEvent.keyboard('{Meta>}f{/Meta}')

      expect(await screen.findByLabelText(texte.suchen)).toBeInTheDocument()

      // **Die Gegenprobe steht daneben**: eine andere Taste **kommt** an.
      // Ohne sie sagte diese Zusicherung auch dann ja, wenn der Horcher gar
      // nicht hinge (§3).
      expect(amFenster).not.toHaveBeenCalled()

      await userEvent.keyboard('{Meta>}k{/Meta}')

      expect(amFenster).toHaveBeenCalledWith('k')
    } finally {
      window.removeEventListener('keydown', nurF)
    }
  })

  it('zählt die Treffer', async () => {
    zeige('card.name + card.type + card.name')

    await userEvent.type(await oeffneSuche(), 'card')

    expect(await screen.findByText('1 von 3')).toBeInTheDocument()
  })

  // **Groß und klein sind zweierlei.**
  //
  // Die Gegenprobe zum Zählen: ohne sie sagte der Lauf darüber auch dann ja,
  // wenn die Suche jede Schreibweise fände — und ein Ersetzen änderte dann
  // still etwas anderes als gemeint.
  it('unterscheidet groß und klein', async () => {
    zeige('card.name + Card.name')

    await userEvent.type(await oeffneSuche(), 'card')

    expect(await screen.findByText('1 von 1')).toBeInTheDocument()
  })

  // **Und angesagt** — UX-10.AK-02 (#1682): wer tippt, hört die Zahl, ohne
  // mit dem Blick das Feld zu verlassen.
  it('sagt die Zahl der Treffer als Status an', async () => {
    zeige('card.name')

    await userEvent.type(await oeffneSuche(), 'card')

    expect(await screen.findByRole('status')).toHaveTextContent('1 von 1')
  })

  it('sagt es, wenn nichts passt', async () => {
    zeige('card.name')

    await userEvent.type(await oeffneSuche(), 'zzz')

    expect(await screen.findByText(texte.keinTreffer)).toBeInTheDocument()
  })

  it('geht zum nächsten und zum vorigen Treffer', async () => {
    zeige('a a a')

    await userEvent.type(await oeffneSuche(), 'a')

    await userEvent.click(screen.getByRole('button', { name: texte.weiter }))

    expect(await screen.findByText('2 von 3')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.zurueck }))

    expect(await screen.findByText('1 von 3')).toBeInTheDocument()
  })

  // **Nach dem letzten kommt der erste.**
  //
  // Ein Suchlauf, der am Ende stehen bliebe, zwänge zum Zurückblättern — und
  // in einer Vorlage mit vierzig Treffern ist das die Bedienung selbst.
  it('läuft vom letzten Treffer wieder zum ersten', async () => {
    zeige('a a')

    await userEvent.type(await oeffneSuche(), 'a')
    await userEvent.click(screen.getByRole('button', { name: texte.weiter }))

    expect(await screen.findByText('2 von 2')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: texte.weiter }))

    expect(await screen.findByText('1 von 2')).toBeInTheDocument()
  })

  it('ersetzt den aktuellen Treffer', async () => {
    const gesetzt = zeige('a b a')

    await userEvent.type(await oeffneSuche(), 'a')
    await userEvent.type(screen.getByLabelText(texte.ersetzen), 'x')
    await userEvent.click(screen.getByRole('button', { name: texte.einsetzen }))

    expect(gesetzt).toHaveBeenCalledWith('x b a')
  })

  it('ersetzt alle Treffer', async () => {
    const gesetzt = zeige('a b a')

    await userEvent.type(await oeffneSuche(), 'a')
    await userEvent.type(screen.getByLabelText(texte.ersetzen), 'x')
    await userEvent.click(
      screen.getByRole('button', { name: texte.alleEinsetzen }),
    )

    expect(gesetzt).toHaveBeenCalledWith('x b x')
  })

  // **Ohne Treffer wird nichts ersetzt.**
  //
  // Die Gegenprobe: ohne sie sagten die zwei Läufe darüber auch dann ja, wenn
  // ein Ersetzen den ganzen Text überschriebe.
  it('ersetzt nichts, wenn nichts passt', async () => {
    const gesetzt = zeige('a b a')

    await userEvent.type(await oeffneSuche(), 'zzz')
    await userEvent.type(screen.getByLabelText(texte.ersetzen), 'x')
    await userEvent.click(screen.getByRole('button', { name: texte.einsetzen }))
    await userEvent.click(
      screen.getByRole('button', { name: texte.alleEinsetzen }),
    )

    expect(gesetzt).not.toHaveBeenCalled()
  })

  // **Der Treffer wird in der Anzeige hervorgehoben, nicht ausgewählt.**
  //
  // # Warum nicht ausgewählt
  //
  // Der erste Anlauf setzte die Auswahl in der Eingabe und holte dafür den
  // Fokus zurück — damit ging **jeder zweite Tastendruck ins Textfeld statt
  // ins Suchfeld**. Diese Zusicherung hat es gefunden: sie fand 14 statt 5,
  // weil nach dem ersten Buchstaben niemand mehr im Suchfeld stand.
  it('hebt den Treffer in der Anzeige hervor', async () => {
    zeige('eins zwei drei')

    await userEvent.type(await oeffneSuche(), 'zwei')

    const marke = await screen.findByTestId('fundmarke')

    expect(marke.textContent).toBe('zwei')
  })

  // **Und der Fokus bleibt im Suchfeld** — die Gegenprobe zum Lauf darüber.
  it('lässt den Fokus im Suchfeld', async () => {
    zeige('eins zwei drei')

    const suchfeld = await oeffneSuche()

    await userEvent.type(suchfeld, 'zwei')

    expect(suchfeld).toHaveFocus()
    expect(suchfeld).toHaveValue('zwei')
  })

  // **Beim Schließen geht die Auswahl in die Eingabe.** Dann ist der Fokus
  // dort richtig aufgehoben, und wer weitertippt, steht am Fund.
  it('setzt die Auswahl beim Schließen', async () => {
    zeige('eins zwei drei')

    await userEvent.type(await oeffneSuche(), 'zwei')
    await userEvent.click(
      screen.getByRole('button', { name: texte.schliessen }),
    )

    const feld = screen.getByLabelText<HTMLTextAreaElement>('Inhalt')

    await waitFor(() => {
      expect(feld.selectionStart).toBe(5)
    })

    expect(feld.selectionEnd).toBe(9)
  })

  // **Ein Treffer in der zweiten Zeile wird auch dort markiert.**
  //
  // Die Umrechnung von Zeichenstelle auf Zeile und Spalte hat ein `+ 1` für
  // den Umbruch, den `split` weggenommen hat; ohne es wandert jede Zeile um
  // eins nach links.
  it('markiert einen Treffer in der zweiten Zeile', async () => {
    zeige('eins\nzwei drei')

    await userEvent.type(await oeffneSuche(), 'drei')

    const marke = await screen.findByTestId('fundmarke')

    expect(marke.textContent).toBe('drei')
    expect(marke.parentElement?.textContent).toContain('zwei')
  })
})
