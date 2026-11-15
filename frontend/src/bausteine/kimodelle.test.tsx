// **Was über KI-Modelle feststeht** — Werkstück A6 von Phase 5.3
// (#1445).
import { render, screen } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { KIModelle } from './KIModelle'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

import type { KIKatalogantwort } from '../antworten'

afterEach(() => {
  raeumeApi()
})

const eigene = kataloge.de.seiten.einstellungen

// katalog baut eine Antwort.
//
// **Die Vorgabe ist »bereit«**, damit die Läufe darunter von der Einrichtung
// nichts wissen müssen; wer sie prüft, setzt sie ausdrücklich.
function katalog(teil = {}): KIKatalogantwort {
  return {
    nurLokal: true,
    vorrang: [],
    vorrangbefunde: [],
    einrichtung: 'bereit',
    anbieter: [
      {
        name: 'ollama',
        einrichtung: 'bereit',
        laufzeit: 'ollama 0.12.3',
        adresse: '',
        ort: 'lokal',
        befunde: [],
        messbar: [],
        gemessen: [],
        sitzungAus: false,
      },
    ],
    cloud: { stand: 'ungewiss', beleg: 'keine Serverdatei' },
    befunde: [],
    modelle: [],
    fehler: null,
    ...teil,
  }
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <KIModelle />
    </Sprachraum>,
  )
}

// **Eine leere Liste ist die Auskunft und kein Mangel.**
it('sagt es, wenn kein Modell feststeht', async () => {
  belegeApi({ KIModelle: () => Promise.resolve(katalog()) })
  zeige()

  expect(await screen.findByText(eigene.kimodelleKeine)).toBeInTheDocument()
  expect(screen.getByText(eigene.kiNurLokal)).toBeInTheDocument()
})

// **Die Regel steht vor der Liste** — sie sagt, ob überhaupt etwas hinausgeht.
it('zeigt die Regel, wenn auch Cloudmodelle erlaubt sind', async () => {
  belegeApi({
    KIModelle: () => Promise.resolve(katalog({ nurLokal: false })),
  })
  zeige()

  expect(await screen.findByText(eigene.kiAuchCloud)).toBeInTheDocument()
})

it('zeigt ein Modell samt Beleg und Herkunft', async () => {
  belegeApi({
    KIModelle: () =>
      Promise.resolve(
        katalog({
          modelle: [
            {
              anbieter: 'ollama',
              alias: 'beispiel:3b',
              laufzeit: 'ollama 0.12.3',
              ort: 'lokal',
              faehigkeiten: [
                {
                  faehigkeit: 'einbettung',
                  datum: '2026-09-11',
                  beleg: 'eigener Lauf',
                },
              ],
              herkunft: [
                {
                  art: 'artefakt',
                  kennung: 'ausgefuehrt',
                  aus: '',
                  quelle: 'ollama:beispiel',
                  revision: 'sha256:dddd',
                  lizenz: 'Apache-2.0',
                },
              ],
              bindung: 'abcdef',
              kontextfenster: 0,
              kontextbeleg: '',
            },
          ],
        }),
      ),
  })
  zeige()

  expect(await screen.findByText('ollama/beispiel:3b')).toBeInTheDocument()
  expect(screen.getByText(/lokal · ollama 0\.12\.3/)).toBeInTheDocument()
  expect(
    screen.getByText(
      einsetzen(eigene.kiBeleg, {
        faehigkeit: 'einbettung',
        datum: '2026-09-11',
        beleg: 'eigener Lauf',
      }),
    ),
  ).toBeInTheDocument()

  // **Bei bekannter Herkunft steht der Hinweis nicht da.**
  expect(screen.queryByText(eigene.kiHerkunftUnbekannt)).not.toBeInTheDocument()
})

// **Eine unbekannte Herkunft steht da und fehlt nicht.**
it('nennt eine unbekannte Herkunft', async () => {
  belegeApi({
    KIModelle: () =>
      Promise.resolve(
        katalog({
          modelle: [
            {
              anbieter: 'ollama',
              alias: 'beispiel:3b',
              laufzeit: 'ollama 0.12.3',
              ort: 'lokal',
              faehigkeiten: [],
              herkunft: [],
              bindung: 'abcdef',
              kontextfenster: 0,
              kontextbeleg: '',
            },
          ],
        }),
      ),
  })
  zeige()

  expect(
    await screen.findByText(eigene.kiHerkunftUnbekannt),
  ).toBeInTheDocument()
})

it('zeigt einen Fehler am Ort', async () => {
  belegeApi({
    KIModelle: () =>
      Promise.resolve(
        katalog({
          fehler: { code: 'validierung', meldung: 'ein Modell ohne Alias' },
        }),
      ),
  })
  zeige()

  expect(await screen.findByRole('alert')).toHaveTextContent(
    'ein Modell ohne Alias',
  )
})

// **Die Einrichtung sagt, warum die Liste leer ist** — A7, #1446.
//
// **Das ist der Unterschied zwischen einer Auskunft und einem leeren
// Bildschirm** (§6): „es steht keines fest" beantwortet nicht, ob jemand eine
// Laufzeit einschalten, Ollama starten oder ein Modell laden muss.
//
// **Seit A8 (#1447) steht sie je Anbieter und nicht einmal darüber.** Der
// Katalog trägt den besten Stand über alle weiter — die Flächen ohne Fenster
// lesen ihn —, und am Fenster stand er als eigener Absatz **über** der Liste:
// bei zwei abgeschalteten Anbietern dreimal derselbe Satz. Angestrichen hat
// es die Sichtprüfung am laufenden Fenster.
it.each([
  ['aus', eigene.kiEinrichtungAus],
  ['unerreichbar', eigene.kiEinrichtungUnerreichbar],
  ['ohnemodell', eigene.kiEinrichtungOhneModell],
  ['ohnefaehigkeit', eigene.kiEinrichtungOhneFaehigkeit],
  ['bereit', eigene.kiEinrichtungBereit],
])('erklärt den Einrichtungsstand %s', async (stand, satz) => {
  belegeApi({
    KIModelle: () =>
      Promise.resolve(
        katalog({
          anbieter: [
            {
              name: 'ollama',
              einrichtung: stand,
              laufzeit: '',
              adresse: '',
              ort: 'lokal',
              befunde: [],
              messbar: [],
              gemessen: [],
              sitzungAus: false,
            },
          ],
        }),
      ),
  })
  zeige()

  expect(
    await screen.findByText(
      einsetzen(eigene.kiAnbieterZeile, { anbieter: 'ollama', stand: satz }),
    ),
  ).toBeInTheDocument()
})

// **Der Cloudstand hat drei Werte, und der dritte ist der ehrliche.**
it.each([
  ['aus', eigene.kiCloudAus],
  ['an', eigene.kiCloudAn],
  ['ungewiss', eigene.kiCloudUngewiss],
])('nennt den Cloudstand %s samt Beleg', async (stand, satz) => {
  belegeApi({
    KIModelle: () =>
      Promise.resolve(
        katalog({ cloud: { stand, beleg: 'server.json sagt nichts' } }),
      ),
  })
  zeige()

  expect(
    await screen.findByText(new RegExp(satz.slice(0, 20))),
  ).toBeInTheDocument()
  expect(screen.getByText(/server\.json sagt nichts/)).toBeInTheDocument()
})

// **Befunde stehen bei der Liste und nicht in einem Protokoll.**
it('zeigt die Befunde der Laufzeit', async () => {
  belegeApi({
    KIModelle: () =>
      Promise.resolve(
        katalog({ befunde: ['einbetter:300m ist nicht vermessen'] }),
      ),
  })
  zeige()

  expect(
    await screen.findByText('einbetter:300m ist nicht vermessen'),
  ).toBeInTheDocument()
})

// **Ein Befund eines Anbieters steht bei diesem Anbieter.**
//
// **Und nicht in der gemeinsamen Liste darunter** (A8, #1447): »der
// Schlüssel fehlt« gilt für einen von zweien, und wer ihn allgemein zeigte,
// ließe raten, welcher gemeint ist.
it('zeigt einen Befund bei seinem Anbieter', async () => {
  belegeApi({
    KIModelle: () =>
      Promise.resolve(
        katalog({
          anbieter: [
            {
              name: 'anthropic',
              einrichtung: 'aus',
              laufzeit: '',
              adresse: '',
              ort: 'lokal',
              befunde: ['Anthropic ist abgeschaltet.'],
              messbar: [],
              gemessen: [],
              sitzungAus: false,
            },
          ],
        }),
      ),
  })
  zeige()

  expect(
    await screen.findByText('Anthropic ist abgeschaltet.'),
  ).toBeInTheDocument()

  // **Und ohne Fassung steht keine Fassungszeile da** — eine leere wäre
  // eine Behauptung über etwas, das nicht geantwortet hat.
  expect(
    screen.queryByText(einsetzen(eigene.kiLaufzeitFassung, { fassung: '' })),
  ).not.toBeInTheDocument()
})

// **Ohne eigene Adresse steht die Vorgabe da** — und nicht eine Lücke.
it('nennt die Vorgabeadresse, wo keine gesetzt ist', async () => {
  belegeApi({ KIModelle: () => Promise.resolve(katalog()) })
  zeige()

  expect(
    await screen.findByText(
      einsetzen(eigene.kiAdresse, { adresse: eigene.kiAdresseVorgabe }),
    ),
  ).toBeInTheDocument()
})

// **Und eine gesetzte Adresse steht so da, wie sie gesetzt ist.**
it('nennt eine gesetzte Adresse', async () => {
  belegeApi({
    KIModelle: () =>
      Promise.resolve(
        katalog({
          anbieter: [
            {
              name: 'ollama',
              einrichtung: 'bereit',
              laufzeit: 'ollama 0.12.3',
              adresse: 'http://ki.example.org:11434',
              befunde: [],
              messbar: [],
              gemessen: [],
              sitzungAus: false,
            },
          ],
        }),
      ),
  })
  zeige()

  expect(
    await screen.findByText(
      einsetzen(eigene.kiAdresse, { adresse: 'http://ki.example.org:11434' }),
    ),
  ).toBeInTheDocument()
})

// **Ein Stand, den dieses Fenster nicht kennt, bleibt nicht leer.**
//
// # Warum das geprüft wird
//
// **Weil die Go-Seite einen sechsten Stand bekommen könnte.** Die Abbildung
// hier fiele dann auf einen Rückfall zurück statt auf einen leeren Absatz —
// und ein leerer Absatz wäre genau das, wogegen die Einrichtung gebaut ist
// (§6). Derselbe Fall für den Cloudstand: dort ist `ungewiss` der Rückfall,
// und das ist die zurückhaltendste Auskunft.
it('fällt bei einem unbekannten Stand auf eine Auskunft zurück', async () => {
  belegeApi({
    KIModelle: () =>
      Promise.resolve(
        katalog({
          anbieter: [
            {
              name: 'ollama',
              einrichtung: 'etwasneues',
              laufzeit: '',
              adresse: '',
              ort: 'lokal',
              befunde: [],
              messbar: [],
              gemessen: [],
              sitzungAus: false,
            },
          ],
          cloud: { stand: 'etwasneues', beleg: 'server.json sagt nichts' },
        }),
      ),
  })
  zeige()

  expect(
    await screen.findByText(
      einsetzen(eigene.kiAnbieterZeile, {
        anbieter: 'ollama',
        stand: eigene.kiEinrichtungAus,
      }),
    ),
  ).toBeInTheDocument()

  expect(
    screen.getByText(new RegExp(eigene.kiCloudUngewiss.slice(0, 20))),
  ).toBeInTheDocument()
})
