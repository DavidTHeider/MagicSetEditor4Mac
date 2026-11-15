// **Der Weg vom Editor bis ins Feld** —
// #897.
//
// # Warum eine eigene Datei
//
// Der Schnitt folgt der Frage (§2): `karteneditor.test.tsx` prüft, was der
// Editor **zeigt** und **schreibt**; hier steht, dass er die Autoersetzung
// **durchreicht**. Und `karteneditor.test.tsx` ging mit ihr über die
// Zeilengrenze.
//
// # Was hier geprüft wird
//
// Die **Naht**, nicht die Rechnung. Dass `Feldeingabe` ersetzt, wenn man ihr
// einen Ersetzer gibt, steht in `autoersetzung.test.tsx`; dass die Regeln an
// der Schreibmarke greifen, in `internal/service`. Hier steht, dass der Weg
// über zwei Ebenen — Karteneditor, Feldzeile, Feldeingabe — überhaupt gegangen
// wird. Genau der hat gefehlt (§10).
import { fireEvent, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { KarteDTO, VorschauAntwort } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Karteneditor } from './Karteneditor'

function probevorschau(): VorschauAntwort {
  return {
    vorschau: {
      stil: '',
      gezeichnet: '',
      bild: '/vorschau/AAAA.png',
      breite: 10,
      hoehe: 14,
      dpi: 150,
      befunde: [],
      flaechen: [],
      nennungen: [],
      vorgaben: [],
    },
  }
}

const karte: KarteDTO = {
  id: '5',
  nummer: 0,
  werte: [{ feld: 'name', wert: 'Feuerball' }],
}

describe('die Autoersetzung erreicht das Feld', () => {
  afterEach(raeumeApi)

  it('reicht das offene Set und die Schreibmarke bis nach Go durch', async () => {
    const AutoErsetzen = vi.fn(() =>
      Promise.resolve({ text: '©', stelle: 1, fehler: null }),
    )

    belegeApi({
      Vorschau: () => Promise.resolve(probevorschau()),
      AutoErsetzen,
    })

    render(
      <Sprachraum sprache="de">
        <Karteneditor
          setID="1"
          karte={karte}
          formen={[]}
          geaendert={() => undefined}
        />
      </Sprachraum>,
    )

    fireEvent.input(screen.getByLabelText('name'), {
      target: { value: '(C)' },
      inputType: 'insertText',
    })

    await vi.waitFor(() => {
      expect(AutoErsetzen).toHaveBeenCalledWith('1', '(C)', 3)
    })
  })
})
