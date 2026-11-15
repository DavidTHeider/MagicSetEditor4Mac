// **Die Sammlung ausgeben und zurücklesen** — Werkstück 4 von Phase 4.5
// (#1292).
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Sammlungsausfuhr } from './Sammlungsausfuhr'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { ohneDublette } from '../test/einfuhr'

afterEach(() => {
  raeumeApi()
})

const texte = {
  csv: 'Als CSV',
  liste: 'Als Liste',
  zurueck: 'CSV zurücklesen',
  ausgabe: 'Ausgabe',
}

// probeCSV ist, was die Naht zurückgibt.
const probeCSV = 'printing,menge\nprobe-druck,4\n'

// eingelesen ist der Rückruf, den der Kasten nach dem Schreiben ruft.
function eingelesen(): void {
  // absichtlich leer
}

// zeige stellt den Kasten mit einer belegten Fassade auf.
function zeige(zusatz = {}) {
  belegeApi({
    SammlungCSVVorschauen: ohneDublette,
    SammlungAusgeben: (form: string) =>
      Promise.resolve({ form, text: probeCSV, fehler: null }),
    ...zusatz,
  })

  render(
    <Sprachraum sprache="de">
      <Sammlungsausfuhr aufEingelesen={eingelesen} />
    </Sprachraum>,
  )
}

// **Die Ausgabe steht danach im Feld.**
it('schreibt die Ausgabe ins Feld', async () => {
  zeige()

  await userEvent.click(screen.getByRole('button', { name: texte.csv }))

  await waitFor(() => {
    expect(screen.getByLabelText(texte.ausgabe)).toHaveValue(probeCSV)
  })
})

// **Die Form geht mit** — und zwar die, auf die geklickt wurde.
it('gibt die gewählte Form aus', async () => {
  const ausgeben = vi.fn((form: string) =>
    Promise.resolve({ form, text: probeCSV, fehler: null }),
  )

  zeige({ SammlungAusgeben: ausgeben })

  await userEvent.click(screen.getByRole('button', { name: texte.liste }))

  await waitFor(() => {
    expect(ausgeben).toHaveBeenCalledWith('liste')
  })
})

// **Zurücklesen gibt es erst, wenn etwas dasteht.**
//
// Ein Knopf, der eine leere Eingabe schickte, bekäme einen Fehler, den niemand
// verursacht hat.
it('bietet das Zurücklesen erst mit Inhalt an', async () => {
  zeige()

  expect(
    screen.queryByRole('button', { name: texte.zurueck }),
  ).not.toBeInTheDocument()

  await userEvent.click(screen.getByRole('button', { name: texte.csv }))

  expect(
    await screen.findByRole('button', { name: texte.zurueck }),
  ).toBeInTheDocument()
})

// **Was zurückgelesen wird, kommt als Bericht an.**
it('sagt, was zurückgelesen wurde', async () => {
  const einlesen = vi.fn(() =>
    Promise.resolve({
      bericht: { angelegt: [1], erhoeht: 2, stueck: 7, befunde: [] },
      fehler: null,
    }),
  )

  zeige({ SammlungCSVEinlesen: einlesen })

  await userEvent.click(screen.getByRole('button', { name: texte.csv }))
  await userEvent.click(
    await screen.findByRole('button', { name: texte.zurueck }),
  )

  await waitFor(() => {
    // **Zwei Argumente und nicht eines** (Befund F03, [#1747]). Der Vertrag
    // erklärte eine Stelligkeit, die Bindung nahm eine andere, und diese
    // Zusicherung hielt die falsche fest: sie war grün, während die
    // ausgelieferte Anwendung `received 1 arguments … expected 2` meldete.
    expect(einlesen).toHaveBeenCalledWith(probeCSV, '', 'addieren')
  })

  expect(
    await screen.findByText('1 neu, 2 erhöht, 7 Stück aufgenommen.'),
  ).toBeInTheDocument()
})

// **Ein eingetipptes CSV geht genauso zurück.**
//
// Das Feld ist beschreibbar, weil hier auch ein CSV hereinkommt, das jemand
// anderswo abgelegt hat — nicht nur die eigene Ausgabe.
it('liest ein eingetipptes CSV zurück', async () => {
  const einlesen = vi.fn(() =>
    Promise.resolve({
      bericht: { angelegt: [1], erhoeht: 0, stueck: 1, befunde: [] },
      fehler: null,
    }),
  )

  zeige({ SammlungCSVEinlesen: einlesen })

  await userEvent.type(screen.getByLabelText(texte.ausgabe), 'printing,menge')
  await userEvent.click(
    await screen.findByRole('button', { name: texte.zurueck }),
  )

  await waitFor(() => {
    expect(einlesen).toHaveBeenCalledWith('printing,menge', '', 'addieren')
  })
})

// **Ein Fehler steht am Kasten** (§6).
it('zeigt einen Fehler beim Ausgeben', async () => {
  belegeApi({
    SammlungCSVVorschauen: ohneDublette,
    SammlungAusgeben: () =>
      Promise.resolve({
        form: '',
        text: '',
        fehler: {
          code: 'validierung' as const,
          schluessel: '',
          meldung: 'Das ist keine Ausfuhrform',
          feld: '',
        },
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Sammlungsausfuhr aufEingelesen={eingelesen} />
    </Sprachraum>,
  )

  await userEvent.click(screen.getByRole('button', { name: texte.csv }))

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Das ist keine Ausfuhrform',
  )
})

// **Und einer beim Zurücklesen ebenso.**
it('zeigt einen Fehler beim Zurücklesen', async () => {
  zeige({
    SammlungCSVEinlesen: () =>
      Promise.resolve({
        bericht: null,
        fehler: {
          code: 'validierung' as const,
          schluessel: '',
          meldung: 'Die Spalte printing fehlt',
          feld: '',
        },
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: texte.csv }))
  await userEvent.click(
    await screen.findByRole('button', { name: texte.zurueck }),
  )

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'Die Spalte printing fehlt',
  )
})

// **Ein Bericht, der `null` ist, gibt keine Zeile** — und keinen Absturz.
//
// Der Zweig, den die Naht offen lässt: `bericht` ist `Einfuhrbericht | null`.
it('kommt mit einem leeren Bericht zurecht', async () => {
  zeige({
    SammlungCSVEinlesen: () => Promise.resolve({ bericht: null, fehler: null }),
  })

  await userEvent.click(screen.getByRole('button', { name: texte.csv }))
  await userEvent.click(
    await screen.findByRole('button', { name: texte.zurueck }),
  )

  await waitFor(() => {
    expect(screen.queryByText(/aufgenommen/)).not.toBeInTheDocument()
  })
})

// **Was nicht durchkam, steht da** — Befund 4 des Zweitblicks über Phase 4.5.
it('zeigt die Zeilen, die beim Zurücklesen nicht durchkamen', async () => {
  zeige({
    SammlungCSVEinlesen: () =>
      Promise.resolve({
        bericht: {
          angelegt: [1],
          erhoeht: 0,
          stueck: 1,
          befunde: ['Zeile 3: "keineZahl" ist keine Stückzahl'],
        },
        fehler: null,
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: texte.csv }))
  await userEvent.click(
    await screen.findByRole('button', { name: texte.zurueck }),
  )

  expect(
    await screen.findByText('Zeile 3: "keineZahl" ist keine Stückzahl'),
  ).toBeInTheDocument()
})

// **Nach »Als Liste« gibt es kein Zurücklesen** — Befund 9.
//
// Die Komponente verwarf die Ausgabeform. Nach »Als Liste« stand derselbe
// Knopf da und schickte den **verlustbehafteten** Listentext an den CSV-Leser.
it('bietet nach einer Listenausgabe kein Zurücklesen an', async () => {
  zeige()

  await userEvent.click(screen.getByRole('button', { name: texte.liste }))

  await waitFor(() => {
    expect(screen.getByLabelText(texte.ausgabe)).toHaveValue(probeCSV)
  })

  expect(
    screen.queryByRole('button', { name: texte.zurueck }),
  ).not.toBeInTheDocument()
})

// **Getippter Text hat keine Form** — und darf zurückgelesen werden.
//
// Die Gegenprobe zu Befund 9: wer selbst ein CSV einfügt, soll es einlesen
// können. Ohne diesen Lauf wäre eine Sperre, die **jeden** getippten Text
// abwiese, grün.
it('bietet für getippten Text wieder ein Zurücklesen an', async () => {
  zeige()

  await userEvent.click(screen.getByRole('button', { name: texte.liste }))
  await waitFor(() => {
    expect(screen.getByLabelText(texte.ausgabe)).toHaveValue(probeCSV)
  })

  await userEvent.type(screen.getByLabelText(texte.ausgabe), 'x')

  expect(
    await screen.findByRole('button', { name: texte.zurueck }),
  ).toBeInTheDocument()
})

// **Und ein Listentext, den jemand bearbeitet hat, wird laut abgewiesen.**
//
// # Der Befund, den dieser Lauf belegt
//
// Der Nachlauf über Phase 4.5 (Befund 9) sah die Klasse offen: schon ein
// angehängtes Zeichen bietet den **verlustbehafteten** Listentext wieder als
// CSV an. Nachgemessen: `sammelliste.kopfLesen` weist ihn ab, und der Satz
// nennt die fehlende Spalte. Es geht nichts verloren — es kostet einen Klick.
//
// **Der Satz stand als Begründung im Quelltext und war ungeprüft.** Das war
// der eigentliche Mangel, und dieser Lauf behebt ihn.
it('weist einen bearbeiteten Listentext beim Zurücklesen laut ab', async () => {
  zeige({
    SammlungAusgeben: (form: string) =>
      Promise.resolve({
        form,
        text: '4 Lightning Bolt (M11) 149\n',
        fehler: null,
      }),
    SammlungCSVEinlesen: () =>
      Promise.resolve({
        bericht: null,
        fehler: {
          code: 'validierung' as const,
          schluessel: '',
          meldung:
            'die Spalte "printing" fehlt — ohne sie ist eine Zeile kein Exemplar',
        },
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: texte.liste }))
  await screen.findByDisplayValue(/4 Lightning Bolt/)

  // **Erst nach der Bearbeitung erscheint der Knopf** — das ist der Fall, den
  // der Nachlauf beschrieb.
  await userEvent.type(screen.getByLabelText(texte.ausgabe), 'x')
  await userEvent.click(screen.getByRole('button', { name: texte.zurueck }))

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'die Spalte "printing" fehlt',
  )
})

// **Und beim Zurücklesen ebenso** — Nachlauf über Phase 4.5, Befund 3.
//
// Dieselbe Naht schickt Bericht und Fehler zugleich
// (`App.SammlungCSVEinlesen`), und dieselbe Hülle warf hier den Bericht weg.
it('zeigt den Teilbericht neben dem Fehler', async () => {
  zeige({
    SammlungCSVEinlesen: () =>
      Promise.resolve({
        bericht: { angelegt: [1, 2], erhoeht: 1, stueck: 7, befunde: [] },
        fehler: {
          code: 'unbekannt' as const,
          schluessel: '',
          meldung: 'Zeile 9 ließ sich nicht schreiben',
        },
      }),
  })

  await userEvent.click(screen.getByRole('button', { name: texte.csv }))
  await screen.findByDisplayValue(/printing/)
  await userEvent.click(screen.getByRole('button', { name: texte.zurueck }))

  expect(
    await screen.findByText(/Zeile 9 ließ sich nicht schreiben/),
  ).toBeInTheDocument()

  expect(screen.getByText(/2 neu, 1 erhöht, 7 Stück/)).toBeInTheDocument()
})
