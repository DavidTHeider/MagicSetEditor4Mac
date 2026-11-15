// **Exportvorlagen, wie sie wirklich installiert sind** — UXR-09
// (#1917),
// R7 und R3 aus `docs/UIREVIEW.md`.
//
// # Der Befund
//
// Am Probebestand stand **eine** Vorlage. Gezählt am 2026-10-02 in
// `full-magic-pack`: **achtzehn** für Magic (`magic-*.mse-export-template`),
// und die reichsten zwei — `magic-simple-exporter` und
// `magic-symbol-extractor` — erklären je **zwölf** Optionen. Die Wahl stand
// als bloßes Auswahlfeld da (R7 verlangt ab fünfzehn Werten eine Suche), und
// zwölf Optionen mit der Wahl der Vorlage und der Kartenwahl darüber sind
// fünfzehn Felder (R3 erlaubt zwölf, danach gruppiert und eingeklappt).
//
// **Die Gestalt kommt aus UXR-07**: ein natives Auswahlfeld, eingegrenzt
// durch ein Suchfeld davor — und die Optionen wie die Stiloptionen: einige
// vorn, der Rest hinter einem Knopf, der sagt, wie viele es sind.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { belegeApi, raeumeApi } from '../test/api'
import { SetVorlagenausfuhr } from './SetVorlagenausfuhr'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten

const probeset = {
  id: '7',
  name: 'Mein Probeset',
  spiel: 'magic',
  stil: 'magic-m15',
  herkunft: '/pfad/probe.mse-set',
  karten: 2,
  herkunftFehlt: false,
}

// Die Namen aus dem Pack — gezählt, nicht erfunden.
const namen = [
  'cockatrice-v2',
  'field-test',
  'forum-booster',
  'forum',
  'lackeybot',
  'lackeyccg',
  'mws',
  'namecheck',
  'planesculptors',
  'simple-exporter',
  'skeleton-extractor',
  'spoiler',
  'symbol-extractor',
  'text',
  'toexp',
  'tts',
  'wagic',
  'xml',
]

// option gibt eine Ja-Nein-Wahl, wie `magic-simple-exporter` sie elfmal hat.
function option(stelle: number) {
  return {
    art: 'boolean',
    name: `option${String(stelle)}`,
    beschreibung: `Option ${String(stelle)}`,
    vorgabe: 'no',
    wert: 'no',
    auswahl: [],
  }
}

function vorlagen(anzahl: number, optionen = 0) {
  return namen.slice(0, anzahl).map((name) => ({
    paket: `magic-${name}.mse-export-template`,
    name: `Magic ${name}`,
    dateityp: '*.txt|*.txt',
    optionen: Array.from({ length: optionen }, (_, stelle) =>
      option(stelle + 1),
    ),
  }))
}

function zeige(liste: ReturnType<typeof vorlagen>) {
  belegeApi({
    Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
    Exportvorlagen: () => Promise.resolve({ vorlagen: liste, befunde: [] }),
  })

  render(
    <Sprachraum sprache="de">
      <SetVorlagenausfuhr satz={probeset} schliessen={vi.fn()} />
    </Sprachraum>,
  )
}

describe('Achtzehn Vorlagen', () => {
  it('lassen sich eingrenzen', async () => {
    zeige(vorlagen(18))

    const wahl = await screen.findByRole('combobox', {
      name: texte.sets.vorlageWaehlen,
    })
    expect(within(wahl).getAllByRole('option')).toHaveLength(18)

    await userEvent.type(
      screen.getByRole('searchbox', {
        name: texte.ausgeben.vorlagenEingrenzen,
      }),
      'forum',
    )

    // **Die gewählte bleibt angeboten** — die erste, `cockatrice-v2` —, dazu
    // die beiden, die passen. Eine Auswahl ohne ihren Wert fiele still auf
    // den ersten Eintrag.
    expect(
      within(wahl)
        .getAllByRole('option')
        .map((eintrag) => eintrag.textContent),
    ).toEqual(['Magic cockatrice-v2', 'Magic forum-booster', 'Magic forum'])
  })

  it('sagen, wenn keine passt', async () => {
    zeige(vorlagen(18))

    await userEvent.type(
      await screen.findByRole('searchbox', {
        name: texte.ausgeben.vorlagenEingrenzen,
      }),
      'apprentice',
    )

    expect(
      screen.getByText(
        einsetzen(kataloge.de.bausteine.listenansicht.nichtsPasst, {
          suche: 'apprentice',
        }),
      ),
    ).toBeInTheDocument()
  })

  // **Unter fünfzehn kein Suchfeld** — die Grenze aus R7. Das Auswahlfeld
  // daneben belegt, dass die Vorlagen geladen sind.
  it('aber vierzehn brauchen keine Suche', async () => {
    zeige(vorlagen(14))

    expect(
      await screen.findByRole('combobox', { name: texte.sets.vorlageWaehlen }),
    ).toBeInTheDocument()
    expect(screen.queryByRole('searchbox')).not.toBeInTheDocument()
  })
})

// **Drei vorn und nicht vier** — im Ausgabeassistenten steht die Fläche
// neben fünf Zielen, und mit der Kartenwahl (zwei), der Suche und der Wahl
// der Vorlage sind drei Optionen genau die zwölf Felder, die R3 erlaubt.
// Vier waren es zuerst, nach den Stiloptionen (UXR-07) — gezählt am
// Assistenten waren das dreizehn.
describe('Zwölf Optionen', () => {
  it('zeigen drei und klappen neun ein', async () => {
    zeige(vorlagen(1, 12))

    expect(await screen.findAllByRole('checkbox')).toHaveLength(3)

    await userEvent.click(
      screen.getByRole('button', {
        name: einsetzen(texte.ausgeben.weitereOptionen, { zahl: '9' }),
      }),
    )

    expect(screen.getAllByRole('checkbox')).toHaveLength(12)
    expect(
      screen.getByRole('checkbox', { name: 'Option 12' }),
    ).toBeInTheDocument()
  })

  // **Drei und weniger stehen ganz da** — ein Knopf für null weitere wäre
  // eine Tür ins Leere.
  it('und drei stehen ganz da', async () => {
    zeige(vorlagen(1, 3))

    expect(await screen.findAllByRole('checkbox')).toHaveLength(3)
    expect(
      screen.queryByRole('button', { name: /Weitere Optionen/ }),
    ).not.toBeInTheDocument()
  })

  // **Und schon die vierte steht dahinter** — die Grenze, an der die Zahl
  // vorn entschieden wird.
  it('und von vieren steht eine dahinter', async () => {
    zeige(vorlagen(1, 4))

    expect(await screen.findAllByRole('checkbox')).toHaveLength(3)
    expect(
      screen.getByRole('button', {
        name: einsetzen(texte.ausgeben.weitereOptionen, { zahl: '1' }),
      }),
    ).toBeInTheDocument()
  })
})
