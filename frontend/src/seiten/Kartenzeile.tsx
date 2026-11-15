// **Eine Zeile der Kartenliste** —
// #1290.
//
// **Eigene Datei, weil sie eine eigene Frage beantwortet** (§2): `Kartenliste`
// sagt, **wie die Liste sich verhält** — Sichtfenster, Tastenweg, wandernde
// Tabstelle —, hier steht, **wie eine Zeile aussieht**. Zusammen gingen sie
// über die Zeilengrenze, als die Liste ein Gitter wurde.

import type { KeyboardEvent, MouseEvent, RefCallback } from 'react'

import type { KartenzeileDTO } from '../api'
import { useTexte } from '../sprache/kontext'
import { gefuellte, type Listenzusatz } from './listenzusatz'

interface ZeilenEigenschaften {
  karte: KartenzeileDTO
  stelle: number
  ist: boolean
  name: string
  eigene: readonly string[]
  spalten: readonly Listenzusatz[]
  waehlen: (karteID: string) => void
  angehakt: readonly string[]
  anhaken: (karteID: string, an: boolean) => void

  // mitKaestchen sagt, ob die Zeile ihr Kästchen für die Massenänderung
  // trägt — nur im Auswahlmodus (UXR-05, #1913; der Grund steht in
  // `Kartenspalte.tsx`).
  mitKaestchen: boolean
  bild: string | undefined

  // tabstelle sagt, ob **diese** Zeile die eine Tabstelle der Liste ist.
  tabstelle: boolean

  zeilenmass: RefCallback<HTMLElement> | undefined
  aufTaste: (ereignis: KeyboardEvent<HTMLElement>) => void

  // aufRechtsklick kommt vor dem Kontextmenü um die Liste — AP-5 (#2362):
  // auf einer anderen als der offenen Karte wählt er sie zuerst.
  aufRechtsklick: (ereignis: MouseEvent<HTMLElement>) => void
}

export function Kartenzeile({
  karte,
  stelle,
  ist,
  name,
  eigene,
  spalten,
  waehlen,
  angehakt,
  anhaken,
  mitKaestchen,
  bild,
  tabstelle,
  zeilenmass,
  aufTaste,
  aufRechtsklick,
}: ZeilenEigenschaften) {
  const texte = useTexte()

  return (
    <div
      role="row"
      // **Die wahre Stelle** — sonst hörte eine Vorlesehilfe „Eintrag 3
      // von 28", während es 3.788 sind. Die wahre Zahl steht am Gitter.
      aria-rowindex={stelle + 1}
      aria-selected={ist}
      aria-label={name}
      // **Die wandernde Tabstelle.** Genau eine Zeile trägt `0`; alle
      // anderen `-1`. Wer die Liste verlässt und wiederkommt, landet auf
      // der aufgeschlagenen Karte und nicht wieder oben.
      tabIndex={tabstelle ? 0 : -1}
      data-karte={karte.id}
      onKeyDown={aufTaste}
      onContextMenu={aufRechtsklick}
      ref={zeilenmass}
      // **Der Grund hebt sich unter dem Zeiger** — AP-4 (#2361), B-11.
      className="flex items-center gap-2 py-1 transition-colors duration-(--dauer-kurz) hover:bg-betont focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
    >
      {/* **Das Kästchen steht nur im Auswahlmodus** (UXR-05, #1913) — ohne
          es führt → von der Zeile gleich auf ihren Knopf. */}
      {mitKaestchen && (
        <span role="gridcell">
          <input
            type="checkbox"
            // **Innerhalb des Gitters erreichbar, nicht über den Tabulator**
            // — mit → und ←, wie in jedem Gitter.
            tabIndex={-1}
            checked={angehakt.includes(karte.id)}
            aria-label={`${name} — ${texte.seiten.karten.massen.auswaehlen}`}
            onChange={(ereignis) => {
              anhaken(karte.id, ereignis.target.checked)
            }}
            className="shrink-0 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
        </span>
      )}

      {/* **Die Stelle in der Liste, nicht der Schlüssel** (#666).
                `karte.nummer` ist `cards.reihenfolge` — ein Sortierschlüssel,
                der nach einem Löschen eine Lücke behält, damit die Rücknahme
                dieselbe Stelle wiederbekommt
                (`internal/storage/karteloeschen.go`). Angezeigt zählte er
                mit, und ein Set mit zwei Karten stand als 2 und 3 da.

                MSE2 hat gar keine gespeicherte Nummer: `Set::cards` ist ein
                `vector`, und die Anzeige ist sein Index. */}
      <span
        role="gridcell"
        className="w-8 shrink-0 text-right text-sm tabular-nums text-leise"
      >
        {stelle + 1}
      </span>

      {/* **Das Bild, wenn es eines gibt** (#1166).
                
                `aria-hidden`, weil daneben der Name steht: eine Vorlesehilfe
                soll nicht „Bild von Lightning Bolt, Lightning Bolt" sagen. Ein
                Wechselstext wäre die zweite Aussage über dieselbe Karte.
                
                **Der Platz bleibt auch ohne Bild.** Sonst sprängen die Namen
                beim Rollen hin und her, während die Bilder nachkommen — und
                genau das tun sie, denn sie kommen später als die Zeile. */}
      <span
        role="gridcell"
        aria-hidden="true"
        // **40 Punkte hoch, wie das Bild aus `service.Vorschauhoehe`.**
        //
        // Der erste Anlauf gab ihm 24, und im laufenden Fenster
        // nachgesehen war das Ergebnis eine Briefmarke: erkennbar war,
        // *dass* da ein Bild ist, nicht *welches*. Für ein Werkzeug,
        // mit dem man Karten gestaltet, ist das die halbe Antwort.
        //
        // **Was es kostet**: die Zeile wächst von 41 auf 48 Punkte, und
        // damit passen sechs Karten weniger ins Fenster. Der Tausch ist
        // Absicht — wer eine Liste aus Textzeilen wollte, hatte sie.
        className="flex h-10 w-14 shrink-0 items-center justify-center overflow-hidden rounded-sm bg-betont"
      >
        {bild !== undefined && (
          <img
            src={bild}
            alt=""
            className="max-h-full max-w-full object-contain"
          />
        )}
      </span>

      {/* **Die ganze Zeile ist der Knopf**, nicht nur der Name: ein
                Ziel von 12 × 60 Punkten trifft man, eines von 12 × 12 nicht.
                
                **`tabIndex={-1}`, und der Knopf bleibt trotzdem** (#1290): die
                Zeile darüber trägt jetzt die Tabstelle, aber mit der Maus wird
                weiter der Knopf geklickt, und eine Vorlesehilfe findet ihn
                über seine Rolle. Ihn wegzunehmen hieße, die Zeile selbst zum
                Knopf zu machen — und ein `role="row"`, das wie ein Knopf
                klingt, ist eine dritte Ansage über dieselbe Sache. */}
      <span role="gridcell" className="flex min-w-0 flex-1">
        <button
          type="button"
          tabIndex={-1}
          aria-current={ist ? 'true' : undefined}
          aria-label={`${name} — ${texte.seiten.karten.bearbeiten}`}
          onClick={() => {
            waehlen(karte.id)
          }}
          className="flex min-w-0 flex-1 flex-col rounded-sm px-2 py-1 text-left hover:bg-betont focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-[current=true]:bg-betont aria-[current=true]:font-medium dark:aria-[current=true]:bg-betont"
        >
          {/* **Der Name bleibt einzeilig** — mit eigenem `truncate`, denn der
                  Knopf ist eine Spalte; ohne liefe »Akroma, Angel of Wrath
                  Avatar« in der laufenden Anwendung über vier Zeilen. */}
          <span className="truncate">
            {name}

            {karte.felder === 0 && (
              <span className="ml-2 text-sm text-leise">
                {texte.seiten.karten.ohneFelder}
              </span>
            )}
          </span>

          {/* **Die gewählten Spalten stehen unter dem Namen** (#1168).
                  
                  **Als eine Zeile und nicht als Tabelle**, und das ist eine
                  Abweichung mit Grund: die Liste ist hier eine schmale Spalte
                  neben dem Editor (14 rem), nicht die breite Tabelle eines
                  eigenen Fensters.

                  **Und in einer Zeile, nicht umbrechend.** Der erste Anlauf
                  ließ sie umbrechen — in der laufenden Anwendung wurde daraus
                  eine Wand: `magic.mse-game` schlägt **elf** sichtbare Spalten
                  vor, und jede Kartenzeile war elf Zeilen hoch. Was nicht
                  hinpasst, wird abgeschnitten; wer mehr sehen will, blendet
                  Spalten aus.

                  **Ein leerer Wert steht nicht da.** »Name:« ohne etwas
                  dahinter ist keine Auskunft, sondern Lärm — und der Vorschlag
                  des Packs enthält Felder wie `full name`, die eine Karte oft
                  nicht führt.

                  **Und was abgeschnitten wird, endet auf »…«** (#1271).
                  Gemessen am 2026-08-28: die Spalte ist **302 px** breit, der
                  Streifen trägt **993 px** — sichtbar sind zwei von elf
                  Feldern. Ohne Auslassungszeichen endete er mitten im Wort und
                  las sich wie ein Fehler; mit ihm liest er sich als das, was
                  er ist — gekürzt.

                  **`block space-x-3` und nicht `flex gap-x-3`**, und das ist
                  der Grund für den Wechsel: `text-overflow` gilt für
                  **Blockcontainer mit Inline-Inhalt**, nicht für einen
                  Flexcontainer. Der erste Anlauf setzte nur `text-ellipsis`
                  daneben — im laufenden Fenster nachgemessen war
                  `text-overflow: ellipsis` gesetzt und wirkungslos, weil
                  `display: flex` danebenstand (§3).
                  
                  **Sie stehen im Knopf**, damit die ganze Zeile weiter
                  anklickbar bleibt — und ihre Aufschrift steht dabei, weil
                  ein nackter Wert nicht sagt, was er ist. */}
          {gefuellte(spalten, karte.id, name).length > 0 && (
            <span className="mt-0.5 block space-x-3 overflow-hidden text-sm text-ellipsis whitespace-nowrap text-leise">
              {gefuellte(spalten, karte.id, name).map((spalte) => (
                <span key={spalte.feld}>
                  <span className="opacity-70">
                    {`${spalte.aufschrift !== '' ? spalte.aufschrift : spalte.feld}: `}
                  </span>
                  {spalte.werte[karte.id]}
                </span>
              ))}
            </span>
          )}

          {/* **Die Marken stehen unter den Spalten** (#1208) — kleiner und
                  eingefasst, damit sie sich von einem Feldwert unterscheiden:
                  eine Marke ist eine Aussage über die **Arbeit**, kein Inhalt
                  der Karte. */}
          {eigene.length > 0 && (
            <span className="mt-0.5 flex gap-1 overflow-hidden whitespace-nowrap">
              {eigene.map((marke) => (
                <span
                  key={marke}
                  className="rounded-full border border-kante px-1.5 text-xs text-leise"
                >
                  {marke}
                </span>
              ))}
            </span>
          )}
        </button>
      </span>
    </div>
  )
}
