// **Die Bedienfläche der Symmetrie**
// (#1092).
//
// MSE hat sie in der Werkzeugleiste des Symboleditors
// (`src/gui/symbol/symmetry_editor.cpp`, Fassung f9d9356 vom 2021-07-06):
// zwei Knöpfe für die Art, ein Drehfeld für die Kopienzahl, und Mitte und
// Griff als Anfasser auf der Zeichenfläche.
//
// **Hier stehen Mitte und Griff zusätzlich als Zahlenfelder.** Sie sind
// Anfasser, die man zieht — und §6 verlangt, dass jede Handlung auch mit der
// Tastatur erreichbar ist. Ein Anfasser, den man nur ziehen kann, wäre für
// jemanden ohne Maus nicht da.

import type { SymbolteilDTO, SymmetrieweiseDTO } from '../antworten'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { useTexte } from '../sprache/kontext'
import { Zahlenfeld } from '../bausteine/Zahlenfeld'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

// hoechsteKopien und niedrigsteKopien sind die Spanne, die MSEs Drehfeld
// anbietet — 2 bis 10. Dieselben Zahlen stehen in `app.HoechsteKopien`.
const niedrigsteKopien = 2
const hoechsteKopien = 10

// zahlenfeld ist die Gestalt eines Eingabefeldes — **mit Fokusring** (§6).
//
// **Dieselben Klassen wie im Zuschnittfenster.** Zwei Fassungen desselben
// Feldes liefen auseinander, und ein Wächter zählt die Ringe nach
// (`internal/werkzeuge/fokusring_test.go`).
const zahlenfeld =
  'w-24 rounded-sm border border-kante bg-transparent px-2 py-1 ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'

interface SymmetriefeldEigenschaften {
  teil: SymbolteilDTO | undefined
  anlegen: () => void
  aufloesen: () => void
  setzen: (weise: Omit<SymmetrieweiseDTO, 'teil'>) => void
}

export function Symmetriefeld({
  teil,
  anlegen,
  aufloesen,
  setzen,
}: SymmetriefeldEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.symbol

  if (teil === undefined) {
    return (
      <section className="flex flex-col gap-2">
        <h3 className="text-lg font-medium">{eigene.symmetrie}</h3>
        <Erklaerungsabsatz
          text={eigene.symmetrieHinweis}
          klasse="text-sm text-leise"
        />
      </section>
    )
  }

  if (teil.art !== 'Symmetrie') {
    return (
      <section className="flex flex-col gap-2">
        <h3 className="text-lg font-medium">{eigene.symmetrie}</h3>
        <Erklaerungsabsatz
          text={eigene.symmetrieErklaerung}
          klasse="text-sm text-leise"
        />
        <Aktionsknopf aktion="anlegen" gewicht="leise" aufKlick={anlegen}>
          {eigene.symmetrieAnlegen}
        </Aktionsknopf>
      </section>
    )
  }

  // **Leer heißt „nicht setzen".** Die vier Zahlen reisen immer mit; welche
  // gelten, sagen `mitteSetzen` und `griffSetzen` — die 0 ist eine gültige
  // Stelle, und ein Aufruf, der nur die Art ändert, dürfte die Mitte nicht
  // dorthin ziehen.
  const unveraendert = {
    spiegelart: '',
    kopien: 0,
    mitteX: teil.mitteX,
    mitteY: teil.mitteY,
    griffX: teil.griffX,
    griffY: teil.griffY,
    mitteSetzen: false,
    griffSetzen: false,
  } as const

  return (
    <section className="flex min-w-64 flex-col gap-3">
      <h3 className="text-lg font-medium">{eigene.symmetrie}</h3>

      <div
        className="flex flex-wrap gap-2"
        role="group"
        aria-label={eigene.symmetrieart}
      >
        {(
          [
            ['rotation', eigene.drehung],
            ['reflection', eigene.achse],
          ] as const
        ).map(([art, beschriftung]) => (
          <Knopf
            key={art}
            gewicht="leise"
            gedrueckt={teil.spiegelart === art}
            aufKlick={() => {
              setzen({ ...unveraendert, spiegelart: art })
            }}
          >
            {beschriftung}
          </Knopf>
        ))}
      </div>

      {/* **Ein Zahlenfeld, das seinen Bereich am Feld sagt** — Werkstück F
          von Phase 4.10. `min` und `max` allein melden nichts, solange kein
          Formular abgesendet wird. */}
      <Zahlenfeld
        beschriftung={eigene.kopien}
        wert={teil.kopien}
        kleinstens={niedrigsteKopien}
        groesstens={hoechsteKopien}
        geaendert={(kopien) => {
          setzen({ ...unveraendert, kopien })
        }}
      />

      {teil.spiegelart === 'reflection' && (
        <Erklaerungsabsatz
          text={eigene.kopienHinweis}
          klasse="text-xs text-leise"
        />
      )}

      <Zahlenpaar
        beschriftungX={eigene.mitteX}
        beschriftungY={eigene.mitteY}
        x={teil.mitteX}
        y={teil.mitteY}
        setzen={(x, y) => {
          setzen({ ...unveraendert, mitteX: x, mitteY: y, mitteSetzen: true })
        }}
      />

      <Zahlenpaar
        beschriftungX={eigene.griffX}
        beschriftungY={eigene.griffY}
        x={teil.griffX}
        y={teil.griffY}
        setzen={(x, y) => {
          setzen({ ...unveraendert, griffX: x, griffY: y, griffSetzen: true })
        }}
      />

      <Knopf gewicht="leise" aufKlick={aufloesen}>
        {eigene.symmetrieAufloesen}
      </Knopf>
    </section>
  )
}

interface ZahlenpaarEigenschaften {
  beschriftungX: string
  beschriftungY: string
  x: number
  y: number
  setzen: (x: number, y: number) => void
}

// Zahlenpaar sind zwei Felder für einen Punkt im Einheitsquadrat.
//
// **Ein Hundertstel je Schritt** — dieselbe Feinheit wie die Pfeiltasten auf
// der Zeichenfläche. Ein Feld, das in Zehnteln springt, träfe keine Mitte.
//
// # Warum hier kein Bereich steht
//
// Weil es keinen gibt. `app.symmetrieAendern` setzt Mitte und Griff ohne
// jeden Vergleich (`internal/app/symbolsymmetrie.go`), und MSE2 tut dasselbe:
// `SymmetryMoveAction::move` (`src/data/action/symbol_part.cpp`, Zeilen
// 461–476, Fassung f9d9356 vom 2021-07-06) rastet auf ein Gitter und auf
// Vielfache von 2π/24, **begrenzt aber nichts**. Eine Mitte außerhalb des
// Einheitsquadrats ist ein zulässiger Drehpunkt; ein `min`/`max` hier wäre
// eine erfundene Regel (§11).
//
// **Ungeprüft heißt nicht ungefiltert.** `Number('')` ist `0` — wer das Feld
// leert, um eine neue Zahl zu tippen, schöbe damit die Mitte auf null. Was
// keine Zahl ist, löst deshalb keinen Ruf aus (Werkstück U3 von Phase 5.1,
// #1414).
function Zahlenpaar({
  beschriftungX,
  beschriftungY,
  x,
  y,
  setzen,
}: ZahlenpaarEigenschaften) {
  return (
    <div className="flex gap-2">
      <label className="flex flex-col gap-1 text-sm">
        {beschriftungX}
        <input
          type="number"
          step={0.01}
          value={x}
          onChange={(ereignis) => {
            const gelesen = alsZahl(ereignis.target.value)

            if (gelesen !== undefined) {
              setzen(gelesen, y)
            }
          }}
          className={zahlenfeld}
        />
      </label>

      <label className="flex flex-col gap-1 text-sm">
        {beschriftungY}
        <input
          type="number"
          step={0.01}
          value={y}
          onChange={(ereignis) => {
            const gelesen = alsZahl(ereignis.target.value)

            if (gelesen !== undefined) {
              setzen(x, gelesen)
            }
          }}
          className={zahlenfeld}
        />
      </label>
    </div>
  )
}

// alsZahl gibt die Zahl eines Feldes — `undefined`, wenn keine dasteht.
function alsZahl(text: string): number | undefined {
  if (text.trim() === '') {
    return undefined
  }

  const zahl = Number(text)

  /* v8 ignore next -- **unter jsdom unerreichbar, im Browser nicht.** Am
     2026-09-08 nachgemessen: jsdom setzt den Wert eines `type="number"` bei
     `1e999` auf die leere Zeichenkette, und dann greift schon der Zweig
     darüber. Ein Browser behält die Schreibweise, `Number` gibt `Infinity`,
     und damit erreichte die Fläche Go als ungültiges JSON. */
  return Number.isFinite(zahl) ? zahl : undefined
}
