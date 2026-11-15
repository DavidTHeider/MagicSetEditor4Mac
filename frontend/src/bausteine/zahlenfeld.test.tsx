// **Ein Zahlenfeld sagt am Feld, wenn der Wert nicht passt** — Werkstück F
// von Phase 4.10
// (#1325).
//
// # Warum es das braucht
//
// Vier Zahlenfelder trugen `min` und `max` — und **meldeten nichts**.
// Gemessen am 2026-09-03 im laufenden Fenster: nach `99999` in „Breite in
// Bildpunkten" gab `checkValidity()` `false` und `validationMessage` den Satz
// „Wert muss kleiner als oder gleich 10000 sein" — im Fenster stand **kein**
// `aria-invalid`, **kein** `role=alert` und kein sichtbarer Hinweis. Der Wert
// blieb im Feld, und der Lauf hätte damit gerechnet.
//
// **`min` und `max` ohne Formular sind eine Zusage, die niemand einlöst**
// (§10). Der Browser meldet sie beim Absenden; hier wird nichts abgesendet.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Zahlenfeld } from './Zahlenfeld'
import { einsetzen } from '../sprache/einsetzen'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

const texte = kataloge.de.allgemein

function zeige(zusatz: Record<string, unknown> = {}) {
  const geaendert = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Zahlenfeld
        beschriftung="Breite"
        wert={750}
        kleinstens={1}
        groesstens={10000}
        geaendert={geaendert}
        {...zusatz}
      />
    </Sprachraum>,
  )

  return geaendert
}

it('nimmt einen Wert im Bereich ohne Hinweis', async () => {
  const geaendert = zeige()

  const feld = screen.getByLabelText('Breite')
  await userEvent.clear(feld)
  await userEvent.type(feld, '900')

  expect(feld).toHaveAttribute('aria-invalid', 'false')
  expect(screen.queryByRole('alert')).toBeNull()
  expect(geaendert).toHaveBeenLastCalledWith(900)
})

it('sagt am Feld, wenn der Wert zu groß ist', async () => {
  const geaendert = zeige()

  const feld = screen.getByLabelText('Breite')
  await userEvent.clear(feld)
  await userEvent.type(feld, '99999')

  expect(feld).toHaveAttribute('aria-invalid', 'true')

  const hinweis = screen.getByRole('alert')
  expect(hinweis).toHaveTextContent('1')
  expect(hinweis).toHaveTextContent('10000')
  expect(hinweis.textContent).toContain(
    texte.zahlenbereich.slice(0, texte.zahlenbereich.indexOf('{')).trim(),
  )

  // **Der Hinweis gehört dem Feld** (§6): `aria-describedby` zeigt darauf,
  // damit ein Vorleseprogramm ihn beim Feld nennt und nicht irgendwo.
  // **Das Ziel liegt im Alarm, nicht auf ihm** (#1989): zeigt
  // `aria-describedby` auf ein Element mit `role="alert"`, liest VoiceOver es
  // nicht vor.
  expect(feld.getAttribute('aria-describedby')).toBe(
    hinweis.firstElementChild?.id,
  )

  // **Und der falsche Wert wird nicht gemeldet.** Ein Aufrufer, der ihn
  // bekäme, rechnete damit — §7, Regel 3: was nicht gilt, wird nicht
  // geschrieben.
  expect(geaendert).not.toHaveBeenCalledWith(99999)
})

it('sagt es auch, wenn der Wert zu klein ist', async () => {
  zeige()

  const feld = screen.getByLabelText('Breite')
  await userEvent.clear(feld)
  await userEvent.type(feld, '0')

  expect(feld).toHaveAttribute('aria-invalid', 'true')
  expect(screen.getByRole('alert')).toBeInTheDocument()
})

// **Ein leeres Feld ist kein Fehler, sondern ein halber Gedanke.**
//
// Wer eine Zahl ändert, löscht sie zuerst. Ein Hinweis nach dem ersten
// Rücktaster wäre eine Meldung über etwas, das gerade geschieht.
it('meldet ein leeres Feld nicht', async () => {
  zeige()

  const feld = screen.getByLabelText('Breite')
  await userEvent.clear(feld)

  expect(feld).toHaveAttribute('aria-invalid', 'false')
  expect(screen.queryByRole('alert')).toBeNull()
})

// **Nicht jedes Feld hat eine Obergrenze** — Werkstück U3 von Phase 5.1
// (#1414).
//
// `domain.Format.zahlenPruefen` prüft die vier Zahlen eines Formats **nur**
// gegen `< 0`, und `service.Sammlungssichten.Dubletten` weist `behalten < 0`
// ab. Eine hier erfundene Obergrenze wiese einen gültigen Wert ab (§11).
it('kennt eine Untergrenze ohne Obergrenze', async () => {
  const geaendert = zeige({ wert: 60, kleinstens: 0, groesstens: undefined })

  const feld = screen.getByLabelText('Breite')
  await userEvent.clear(feld)
  await userEvent.type(feld, '99999')

  expect(feld).toHaveAttribute('aria-invalid', 'false')
  expect(geaendert).toHaveBeenLastCalledWith(99999)
})

it('meldet ohne Obergrenze nur die Untergrenze', async () => {
  const geaendert = zeige({ wert: 60, kleinstens: 0, groesstens: undefined })

  const feld = screen.getByLabelText('Breite')
  await userEvent.clear(feld)
  await userEvent.type(feld, '-3')

  expect(feld).toHaveAttribute('aria-invalid', 'true')
  expect(screen.getByRole('alert')).toHaveTextContent(
    einsetzen(texte.zahlenAb, { kleinstens: '0' }),
  )
  expect(geaendert).not.toHaveBeenCalledWith(-3)
})

// **„Ganz" ist keine Verzierung**: die Felder dahinter sind in Go `int`. Eine
// Kommazahl käme dort als Umwandlungsfehler an — also als Satz über JSON und
// nicht als Befund am Feld.
it('weist eine Kommazahl ab, wo nur ganze Zahlen hingehören', async () => {
  const geaendert = zeige({
    wert: 60,
    kleinstens: 0,
    groesstens: undefined,
    ganzzahlig: true,
  })

  const feld = screen.getByLabelText('Breite')
  await userEvent.clear(feld)
  await userEvent.type(feld, '2.5')

  expect(feld).toHaveAttribute('aria-invalid', 'true')
  expect(screen.getByRole('alert')).toHaveTextContent(
    einsetzen(texte.ganzeZahlAb, { kleinstens: '0' }),
  )
  expect(geaendert).not.toHaveBeenCalledWith(2.5)
})

// **Eine ganze Zahl jenseits der sicheren Grenze ist keine** — Befund 1 des
// nachgeholten Zweitblicks über F03 (J05, #2269): `Number('9007199254740993')`
// ist `9007199254740992`, und `Number.isInteger` hält auch die für ganz. Die
// Menge, die weiterginge, hätte niemand getippt.
it('weist eine ganze Zahl ab, die sich nicht verlustfrei darstellen lässt', async () => {
  const geaendert = zeige({
    wert: 1,
    kleinstens: 1,
    groesstens: undefined,
    ganzzahlig: true,
  })

  const feld = screen.getByLabelText('Breite')
  await userEvent.clear(feld)
  await userEvent.type(feld, '9007199254740993')

  expect(feld).toHaveAttribute('aria-invalid', 'true')
  expect(screen.getByRole('alert')).toHaveTextContent(
    einsetzen(texte.ganzeZahlAb, { kleinstens: '1' }),
  )
  expect(geaendert).not.toHaveBeenCalledWith(9007199254740992)

  // Die Gegenprobe am Rand: die größte sichere Zahl geht durch.
  await userEvent.clear(feld)
  await userEvent.type(feld, '9007199254740991')

  expect(feld).not.toHaveAttribute('aria-invalid', 'true')
  expect(geaendert).toHaveBeenLastCalledWith(9007199254740991)
})

it('nennt bei zwei Grenzen und ganzen Zahlen beide Grenzen', async () => {
  zeige({ kleinstens: 1, groesstens: 15, ganzzahlig: true })

  const feld = screen.getByLabelText('Breite')
  await userEvent.clear(feld)
  await userEvent.type(feld, '2.5')

  expect(screen.getByRole('alert')).toHaveTextContent(
    einsetzen(texte.ganzeZahlBereich, { kleinstens: '1', groesstens: '15' }),
  )
})

// **Der Befund der Anwendung schlägt das eigene Urteil** — dieselbe Regel wie
// in `Pflichtfeld`: wer etwas eingegeben hat, das der Dienst abweist, soll das
// lesen und nicht „eine Zahl zwischen 1 und 10000".
it('zeigt den Befund der Anwendung am Feld', () => {
  zeige({ befund: 'Die Obergrenze 40 liegt unter der Untergrenze 60.' })

  const feld = screen.getByLabelText('Breite')
  expect(feld).toHaveAttribute('aria-invalid', 'true')
  expect(screen.getByRole('alert')).toHaveTextContent(
    'Die Obergrenze 40 liegt unter der Untergrenze 60.',
  )
})

// **Ein Hilfesatz der Seite bleibt am Feld, auch wenn ein Befund dasteht.**
// Der Kopf dieses Bausteins sagte das seit dem 2026-09-03 zu — und der Rumpf
// **ersetzte** die Kennung, statt sie stehen zu lassen (§10).
it('behält den Hilfesatz der Seite neben dem Befund', async () => {
  zeige({ beschreibung: 'breite-hilfe' })

  const feld = screen.getByLabelText('Breite')
  expect(feld.getAttribute('aria-describedby')).toContain('breite-hilfe')

  await userEvent.clear(feld)
  await userEvent.type(feld, '99999')

  const beschrieben = feld.getAttribute('aria-describedby') ?? ''
  expect(beschrieben).toContain('breite-hilfe')
  expect(beschrieben).toContain(screen.getByRole('alert').id)
})
