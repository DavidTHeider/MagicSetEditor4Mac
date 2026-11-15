// **Den Ausschnitt eines Bildes wählen** —
// #576.
//
// # Was MSE zeigt
//
// `ImageSliceWindow` (`src/gui/image_slice_window.cpp`, Zeilen 68–360, Fassung
// f9d9356 vom 2021-07-06): links das Ursprungsbild mit einem verschiebbaren
// Rahmen, rechts die Vorschau in Feldgröße, dazu vier Zahlenfelder — `left`,
// `top`, `width`, `height` — und eine Ankreuzfläche für das Seitenverhältnis.
//
// **Die vier Zahlenfelder sind kein Zusatz, sondern die Tastaturbedienung**
// (§6). MSE hat sie, und mit ihnen ist jede Bewegung ohne Maus erreichbar.
//
// # Was hier fehlt und warum
//
// MSEs Fenster führt zusätzlich vier Größenarten (`original size`, `size to
// fit`, `force to fit`, `custom size`), zwei Zoomfelder und einen
// Schärfefilter. Die Größenarten sind Abkürzungen für Werte, die sich hier
// unmittelbar setzen lassen; `force to fit` steht als Ankreuzfläche da. Zoom
// ist dieselbe Angabe wie die Breite, nur in Prozent. Der Schärfefilter ist
// ein Filter und kein Ausschnitt — MSE hat ihn in der Vorgabe **aus**.
//
// # Kein Dialog über einem Dialog (§6)
//
// Der Ausschnitt steht **im Editor**, an der Stelle der Vorschau, und nicht in
// einem Fenster darüber.
import { useState } from 'react'

import { api, type AuswahlDTO, type ZuschnittAntwort } from '../api'
import { useLauf } from '../haken/useLauf'
import { einsetzen as einsetzenText } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Zuschnittrahmen } from './Zuschnittrahmen'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface ZuschnittfensterEigenschaften {
  zuschnitt: ZuschnittAntwort
  laeuft: boolean
  uebernehmen: (auswahl: AuswahlDTO) => void
  abbrechen: () => void
}

const knopf =
  'rounded-sm border border-kante px-3 py-1 ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60 '

const zahlenfeld =
  'w-20 rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'

export function Zuschnittfenster({
  zuschnitt,
  laeuft,
  uebernehmen,
  abbrechen,
}: ZuschnittfensterEigenschaften) {
  const texte = useTexte()
  const lauf = useLauf()

  const [auswahl, setzeAuswahl] = useState<AuswahlDTO>(zuschnitt.auswahl)
  const [frei, setzeFrei] = useState(false)

  // **Die Antwort aus Go setzt den Rahmen, nicht die Eingabe.**
  //
  // Die Regel ist portiert (`ImageSlice::constrain`); eine zweite Fassung hier
  // wäre eine zweite Antwort auf dieselbe Frage (§15). Der Preis ist eine
  // Bindungsfahrt je Bewegung — und §7, Regel 2: eine überholte Antwort wird
  // verworfen, sonst spränge der Rahmen auf einen Stand von vorhin zurück.
  const zeige = async (holen: () => Promise<AuswahlDTO>) => {
    const meiner = lauf.naechster()
    const gegrenzt = await holen()

    if (lauf.istAktuell(meiner)) {
      setzeAuswahl(gegrenzt)
    }
  }

  const grenze = (naechste: AuswahlDTO, bevorzugt: string, freiJetzt = frei) =>
    zeige(() =>
      api().Zuschnittgrenze(
        zuschnitt.quelleBreite,
        zuschnitt.quelleHoehe,
        zuschnitt.zielBreite,
        zuschnitt.zielHoehe,
        naechste,
        freiJetzt,
        bevorzugt,
      ),
    )

  const ziehe = (x: number, y: number, kanteX: number, kanteY: number) =>
    zeige(() =>
      api().Zuschnittziehen(
        zuschnitt.quelleBreite,
        zuschnitt.quelleHoehe,
        zuschnitt.zielBreite,
        zuschnitt.zielHoehe,
        auswahl,
        frei,
        x,
        y,
        kanteX,
        kanteY,
      ),
    )

  // **Die Vorschau in Feldgröße als Hintergrund** — die Quelle im Maßstab des
  // Feldes, verschoben um den Ursprung des Ausschnitts.
  //
  // **Oberhalb des JSX zusammengesetzt**: `keintextimjsx` weist eine
  // Zeichenkette mit Platzhalter im Ausdruck ab, und sie ist von einem
  // Oberflächentext nicht zu unterscheiden.
  const wachstumB = zuschnitt.zielBreite / auswahl.breite
  const wachstumH = zuschnitt.zielHoehe / auswahl.hoehe
  const masse = `${String(zuschnitt.quelleBreite * wachstumB)}px ${String(zuschnitt.quelleHoehe * wachstumH)}px`
  const ursprung = `${String(-auswahl.x * wachstumB)}px ${String(-auswahl.y * wachstumH)}px`

  const zahl = (
    beschriftung: string,
    wert: number,
    setzen: (neu: number) => AuswahlDTO,
    bevorzugt: string,
  ) => (
    <label className="flex items-center gap-2 text-sm">
      <span className="text-leise">{beschriftung}</span>
      {/* **Was keine Zahl ist, löst keinen Zug aus** — Werkstück U3 von
          Phase 5.1
          (#1414).
          `Number('')` ist `0`: wer das Feld leerte, um eine neue Zahl zu
          tippen, schickte damit einen Zuschnitt bei null los, und der Rahmen
          sprang. **Ein Bereich steht hier trotzdem nicht** — die Regel ist
          `render.Zuschnittwahl.Begrenze`, und sie **berichtigt** statt
          abzuweisen; eine zweite Fassung wäre eine zweite Antwort auf
          dieselbe Frage (§15). */}
      <input
        type="number"
        value={wert}
        className={zahlenfeld}
        onChange={(ereignis) => {
          const gelesen = alsZahl(ereignis.target.value)

          if (gelesen !== undefined) {
            void grenze(setzen(gelesen), bevorzugt)
          }
        }}
      />
    </label>
  )

  return (
    <div className="rounded-sm border border-kante p-3">
      <h3 className="font-medium">
        {texte.seiten.karten.zuschnitt.ueberschrift}
      </h3>
      <Erklaerungsabsatz
        text={texte.seiten.karten.zuschnitt.erklaerung}
        klasse="mt-1 max-w-prose text-sm text-leise"
      />

      <div className="mt-3 flex flex-wrap items-start gap-4">
        <Zuschnittrahmen
          zuschnitt={zuschnitt}
          auswahl={auswahl}
          frei={frei}
          ziehen={(x, y, kanteX, kanteY) => {
            void ziehe(x, y, kanteX, kanteY)
          }}
        />

        {/* **Die Vorschau in Feldgröße** — MSEs `ImageSlicePreview`.
            Der Ausschnitt wird über den Hintergrund gezeigt: `background-size`
            ist die Quelle im Maßstab des Feldes, die Verschiebung sein
            Ursprung. */}
        <div>
          <div
            role="img"
            aria-label={texte.seiten.karten.zuschnitt.vorschau}
            className="border border-kante bg-betont"
            style={{
              width: zuschnitt.zielBreite,
              height: zuschnitt.zielHoehe,
              backgroundImage: `url(${zuschnitt.quelle})`,
              backgroundSize: masse,
              backgroundPosition: ursprung,
            }}
          />
          <p className="mt-1 text-sm text-leise">
            {einsetzenText(texte.seiten.karten.zuschnitt.masse, {
              breite: String(zuschnitt.zielBreite),
              hoehe: String(zuschnitt.zielHoehe),
            })}
          </p>
        </div>
      </div>

      <div className="mt-3 flex flex-wrap items-center gap-3">
        {zahl(
          texte.seiten.karten.zuschnitt.links,
          auswahl.x,
          (neu) => ({ ...auswahl, x: neu }),
          '',
        )}
        {zahl(
          texte.seiten.karten.zuschnitt.oben,
          auswahl.y,
          (neu) => ({ ...auswahl, y: neu }),
          '',
        )}
        {zahl(
          texte.seiten.karten.zuschnitt.breite,
          auswahl.breite,
          (neu) => ({ ...auswahl, breite: neu }),
          'breite',
        )}
        {zahl(
          texte.seiten.karten.zuschnitt.hoehe,
          auswahl.hoehe,
          (neu) => ({ ...auswahl, hoehe: neu }),
          'hoehe',
        )}

        <label className="flex items-center gap-2 text-sm">
          <input
            type="checkbox"
            checked={!frei}
            className="size-4 accent-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            onChange={(ereignis) => {
              const naechste = !ereignis.target.checked

              setzeFrei(naechste)
              void grenze(auswahl, '', naechste)
            }}
          />
          <span>{texte.seiten.karten.zuschnitt.seitenverhaeltnis}</span>
        </label>
      </div>

      <div className="mt-3 flex gap-2">
        <button
          type="button"
          disabled={laeuft}
          className={knopf}
          onClick={() => {
            uebernehmen(auswahl)
          }}
        >
          {texte.seiten.karten.zuschnitt.uebernehmen}
        </button>
        <button type="button" className={knopf} onClick={abbrechen}>
          {texte.seiten.karten.zuschnitt.abbrechen}
        </button>
      </div>
    </div>
  )
}

// alsZahl gibt die Zahl eines Feldes — `undefined`, wenn keine dasteht.
//
// **`Infinity` ist keine**: ein `type="number"` nimmt `1e999` an, und daraus
// wird eine Fläche, mit der die Grenzrechnung nichts anfangen kann.
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
