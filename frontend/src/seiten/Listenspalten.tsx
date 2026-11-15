// **Welche Spalten die Kartenliste zeigt und wonach sie ordnet** —
// Werkstück 3 von Phase 3.7
// (#1168).
//
// # Die Wahl gehört dem Spiel
//
// MSE2 merkt sie in `GameSettings` (`src/data/settings.hpp`, Zeilen 78–80,
// Fassung f9d9356 vom 2021-07-06) — wer sich seine Spalten für Magic
// einrichtet, findet sie in **jedem** Magic-Set wieder. Der Satz steht auch in
// der Oberfläche, weil sonst niemand damit rechnet.
//
// # Warum sie zugeklappt beginnt
//
// **Weil MSE2 sie hinter einen Dialog legt** — `CardListColumnSelectDialog`,
// aufgerufen aus `CardListBase::selectColumns` (`src/gui/control/card_list.cpp`,
// Zeilen 290–305, Fassung f9d9356 vom 2021-07-06). Sie ist eine Einrichtung
// und keine Handlung, die man oft tut. §6 verbietet den Dialog über dem
// Dialog; ein Aufklappknopf ist die Gestalt, die hierher passt — und er hält
// die schmale Spalte frei.
//
// # Warum eine Liste mit Ankreuzfeldern und kein Ziehen
//
// **Weil Ziehen mit der Tastatur nicht geht** (§6). Die Reihenfolge ändert
// sich über zwei Knöpfe je Zeile — dasselbe Muster wie beim Umordnen der
// Karten (`useKartenordnen.ts`), und dort ist es aus demselben Grund so.
import { useId, useState } from 'react'

import { api } from '../fassade'
import { Knopf } from '../bausteine/Knopf'
import type { ListenspalteDTO } from '../api'
import type { Vorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { auspacken, type Umschlag } from '../umschlag'

interface ListenspaltenEigenschaften {
  setID: string

  spalten: readonly ListenspalteDTO[]

  // sortierfeld ist die wirksame Ordnung — leer heißt: die Reihenfolge des
  // Sets.
  sortierfeld: string
  aufsteigend: boolean

  // fehler ist die Meldung des Ladens — sie steht hier, weil hier fehlt, was
  // sie erklärt.
  fehler: string | undefined

  // ohneVorlage sagt, dass die Spalten aus den Karten stammen.
  ohneVorlage: boolean

  vorgang: Vorgang

  // geaendert lädt die Liste neu — die Wahl gilt dem Spiel, und die Antwort
  // darauf ist ein neuer Aufbau.
  geaendert: () => void
}

// faltgrenze ist R3 aus `docs/UIREVIEW.md`: höchstens zwölf Felder offen.
const faltgrenze = 12

export function Listenspalten({
  setID,
  spalten,
  sortierfeld,
  aufsteigend,
  fehler,
  ohneVorlage,
  vorgang,
  geaendert,
}: ListenspaltenEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.listenspalten

  const [offen, setzeOffen] = useState(false)
  const kastenID = useId()
  const filterID = useId()

  // **Ein Filter, weil es 243 sind.** Gemessen in der laufenden Anwendung am
  // 2026-08-27: `magic.mse-game` bietet **243** Spalten an, elf davon
  // sichtbar. Aufgeklappt war der Kasten 10.361 Punkte hoch in einer Spalte
  // von 504 — wer »rarity« sucht, rollt an 200 Namen vorbei.
  //
  // **MSE2 braucht ihn nicht**, und das ist kein Widerspruch: dort steht die
  // Wahl in einem eigenen Fenster, das so hoch sein darf, wie der Bildschirm
  // ist (`CardListColumnSelectDialog`). Hier ist sie eine schmale Spalte
  // neben dem Editor.
  const [filter, setzeFilter] = useState('')

  // **Die Schreibmeldung steht hier und nicht oben** (§6): sie gehört an die
  // Handlung, die sie verursacht hat. Dieser Kasten überlebt den Kartenwechsel
  // — er hängt am **Set**, nicht an der Karte —, und deshalb ist eine eigene
  // Meldung hier keine, die mit ihrem Bauteil verschwindet.
  const [schreibfehler, setzeSchreibfehler] = useState<string>()

  const umschaltenOffen = () => {
    setzeOffen((war) => !war)
  }

  // **Ein Vorgang für Spalten und Ordnung** (§7, Regel 4). Beide schreiben
  // dieselbe Wahl eines Spiels; zwei Schlüssel ließen ein Umsortieren und ein
  // Ausblenden nebeneinander laufen.
  const laeuft = vorgang.laeuft('listenwahl')

  const schreibe = (arbeit: () => Promise<Umschlag>) => {
    void vorgang.starte('listenwahl', async () => {
      try {
        const antwort = await arbeit()

        auspacken(antwort, () => antwort)
        setzeSchreibfehler(undefined)

        // **Regel 3**: geschrieben ist geschrieben. Das Nachladen ist
        // Bestenfalls-Arbeit.
        geaendert()
      } catch (ursache: unknown) {
        setzeSchreibfehler(meldungAus(ursache, texte))
      }
    })
  }

  // passend sind die Spalten, die der Filter durchlässt.
  //
  // **Über die Aufschrift **und** den Feldnamen**: das Paket beschriftet
  // `casting cost` mit »Cost«, und wer das eine sucht, soll das andere finden.
  const gesucht = filter.trim().toLowerCase()

  const passend =
    gesucht === ''
      ? spalten
      : spalten.filter(
          (spalte) =>
            spalte.feld.toLowerCase().includes(gesucht) ||
            spalte.aufschrift.toLowerCase().includes(gesucht),
        )

  // **Eine lange Wahl wird gefaltet** — UXR-05
  // (#1913).
  //
  // `magic.mse-game` bietet 243 Spalten an, elf davon sichtbar. Aufgeklappt
  // standen bis dahin alle da: 243 Kästchen und 486 Pfeile — R6 aus
  // `docs/UIREVIEW.md` lässt 200 Bedienelemente zu, R3 zwölf Felder.
  //
  // **Ohne Suche stehen die gezeigten Spalten da**, die übrigen hinter dem
  // Suchfeld; **mit Suche höchstens zwölf Treffer**, und ein Satz sagt, dass
  // es mehr gibt. Ein Spiel mit höchstens zwölf Spalten sieht alles wie
  // vorher — gefaltet wird erst, wo die Menge es verlangt.
  const gefaltet = gesucht === '' && spalten.length > faltgrenze
  const gezeigt = gefaltet
    ? spalten.filter((spalte) => spalte.sichtbar)
    : passend.slice(0, faltgrenze)

  const setzeSpalten = (neue: ListenspalteDTO[]) => {
    schreibe(() => api().ListenspaltenSetzen(setID, neue))
  }

  const umschalten = (feld: string, sichtbar: boolean) => {
    setzeSpalten(
      spalten.map((spalte) =>
        spalte.feld === feld ? { ...spalte, sichtbar } : spalte,
      ),
    )
  }

  // stelleVon gibt die Stelle einer Spalte in der **ganzen** Liste.
  const stelleVon = (spalte: ListenspalteDTO) =>
    spalten.findIndex((eine) => eine.feld === spalte.feld)

  // schiebe tauscht eine Spalte mit ihrer Nachbarin.
  //
  // **Über den Index und nicht über eine Stellenzahl**: die Reihenfolge der
  // Liste **ist** die Reihenfolge der Spalten, und eine zweite Angabe daneben
  // liefe damit auseinander.
  const schiebe = (
    spalte: ListenspalteDTO,
    stelle: number,
    richtung: -1 | 1,
  ) => {
    // **Herausnehmen und wieder einsetzen** statt zweier Zuweisungen über
    // Indizes. Die Knöpfe an den Rändern sind angehalten, ein Nachbar fehlt
    // also nie — aber ein Zugriff über einen Index gäbe `undefined` her, und
    // die Wache dagegen wäre ein Zweig, den kein Lauf betreten kann (§3).
    const uebrige = spalten.filter((_, stand) => stand !== stelle)

    setzeSpalten([
      ...uebrige.slice(0, stelle + richtung),
      spalte,
      ...uebrige.slice(stelle + richtung),
    ])
  }

  return (
    <section aria-label={eigene.ueberschrift} className="flex flex-col gap-2">
      <Knopf aufKlick={umschaltenOffen} ausgeklappt={offen} steuert={kastenID}>
        {eigene.ueberschrift}
      </Knopf>

      {!offen ? null : (
        <div id={kastenID} className="flex flex-col gap-2">
          {/* **Der Satz steht da, weil sonst niemand damit rechnet.** Eine Wahl,
              die stillschweigend auch andere Sets ändert, ist eine Überraschung. */}
          <p className="text-sm text-leise">{eigene.jeSpiel}</p>

          {/* **Ohne Spielpaket steht es dabei** (§8). Sonst sähe die Wahl
              aus, als hätte das Spiel keine Aufschriften — und niemand käme
              auf die Idee, das fehlende Paket zu beziehen. */}
          {ohneVorlage && (
            <p className="text-sm text-leise">{eigene.ohneVorlage}</p>
          )}

          {(fehler ?? schreibfehler) !== undefined && (
            <p role="alert" className="text-sm text-fehler">
              {fehler ?? schreibfehler}
            </p>
          )}

          <label className="flex flex-col gap-1 text-sm" htmlFor={filterID}>
            {eigene.filtern}
            <input
              id={filterID}
              type="search"
              value={filter}
              onChange={(ereignis) => {
                setzeFilter(ereignis.target.value)
              }}
              className="rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            />
          </label>

          {/* **Der Kasten rollt für sich.** Ohne die Höhengrenze wäre er bei
              `magic.mse-game` über zehntausend Punkte hoch, und die Spalte
              daneben käme nie wieder in Sicht. */}
          <ul className="flex max-h-64 flex-col gap-1 overflow-y-auto">
            {gezeigt.map((spalte) => (
              <li key={spalte.feld} className="flex items-center gap-2 text-sm">
                <label className="flex min-w-0 flex-1 items-center gap-2">
                  <input
                    type="checkbox"
                    checked={spalte.sichtbar}
                    disabled={laeuft}
                    onChange={(ereignis) => {
                      umschalten(spalte.feld, ereignis.target.checked)
                    }}
                    className="shrink-0 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                  />
                  <span className="truncate">
                    {spalte.aufschrift !== '' ? spalte.aufschrift : spalte.feld}
                  </span>
                </label>

                {/* **Die Stelle kommt aus der ganzen Liste und nicht aus der
                    gefilterten** — sonst schöbe ein Pfeil die Spalte an den
                    Platz ihres Nachbarn **im Filter**, und das wäre irgendwo. */}
                <Knopf
                  angehalten={laeuft || stelleVon(spalte) === 0}
                  aufKlick={() => {
                    schiebe(spalte, stelleVon(spalte), -1)
                  }}
                  bezeichnung={`${spalte.feld} — ${eigene.nachOben}`}
                >
                  {'\u2191'}
                </Knopf>

                <Knopf
                  angehalten={
                    laeuft || stelleVon(spalte) === spalten.length - 1
                  }
                  aufKlick={() => {
                    schiebe(spalte, stelleVon(spalte), 1)
                  }}
                  bezeichnung={`${spalte.feld} — ${eigene.nachUnten}`}
                >
                  {'\u2193'}
                </Knopf>
              </li>
            ))}
          </ul>

          {/* **Wie viele es sind, steht dabei** — sonst sieht ein Filter mit
              zwei Treffern aus wie ein Spiel mit zwei Feldern. */}
          <p className="text-sm text-leise">
            {einsetzen(eigene.gezeigt, {
              gezeigt: String(gezeigt.length),
              alle: String(spalten.length),
            })}
          </p>

          {/* **Und wo der Rest steht** (UXR-05): gefaltet ist nicht fort. */}
          {gefaltet && (
            <p className="text-sm text-leise">
              {einsetzen(eigene.ausgeblendet, {
                anzahl: String(spalten.length - gezeigt.length),
              })}
            </p>
          )}

          {!gefaltet && passend.length > gezeigt.length && (
            <p className="text-sm text-leise">{eigene.genauer}</p>
          )}

          <Ordnungswahl
            setID={setID}
            spalten={spalten}
            sortierfeld={sortierfeld}
            aufsteigend={aufsteigend}
            laeuft={laeuft}
            schreibe={schreibe}
          />
        </div>
      )}
    </section>
  )
}

interface OrdnungsEigenschaften {
  setID: string
  spalten: readonly ListenspalteDTO[]
  sortierfeld: string
  aufsteigend: boolean
  laeuft: boolean
  schreibe: (arbeit: () => Promise<Umschlag>) => void
}

// Ordnungswahl setzt, wonach die Liste ordnet.
//
// **Ein Auswahlfeld und keine Klicks auf Spaltenköpfe**: die Liste ist hier
// schmal und trägt keine Köpfe — anders als MSE2, wo der Kopf der Ort dafür
// ist. Wer die Ordnung ändert, tut es dort, wo er auch die Spalten wählt.
function Ordnungswahl({
  setID,
  spalten,
  sortierfeld,
  aufsteigend,
  laeuft,
  schreibe,
}: OrdnungsEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.listenspalten

  const setze = (feld: string, auf: boolean) => {
    schreibe(() => api().ListensortierungSetzen(setID, feld, auf))
  }

  return (
    <div className="flex flex-wrap items-end gap-2">
      <label className="flex flex-col gap-1 text-sm">
        {eigene.sortieren}
        <select
          value={sortierfeld}
          disabled={laeuft}
          onChange={(ereignis) => {
            setze(ereignis.target.value, aufsteigend)
          }}
          className="rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        >
          {/* **„Gar nicht" ist eine Wahl und kein leerer Eintrag** — es ist
              MSEs `sort_by_column = -1`, und dann gilt die Reihenfolge des
              Sets. */}
          <option value="">{eigene.garNicht}</option>

          {/* **Angeboten wird, was die Liste zeigt** — UXR-05 (#1913). In MSE2
              zeigt die Ordnung auf eine **sichtbare** Spalte: ein gemerktes
              Feld, das keine ist, ergibt `sort_by_column = -1`
              (`CardListBase::rebuild`, `src/gui/control/card_list.cpp`,
              Zeilen 231–243, Fassung f9d9356 vom 2021-07-06). Bei `magic`
              standen hier 243 Werte in einer Auswahl ohne Suche (R7). **Eine
              schon gewählte bleibt stehen**, auch ausgeblendet — sonst zeigte
              die Auswahl eine Ordnung an, die nicht gilt. */}
          {spalten
            .filter((spalte) => spalte.sichtbar || spalte.feld === sortierfeld)
            .map((spalte) => (
              <option key={spalte.feld} value={spalte.feld}>
                {spalte.aufschrift !== '' ? spalte.aufschrift : spalte.feld}
              </option>
            ))}
        </select>
      </label>

      {sortierfeld !== '' && (
        <Knopf
          angehalten={laeuft}
          aufKlick={() => {
            setze(sortierfeld, !aufsteigend)
          }}
        >
          {aufsteigend ? eigene.aufsteigend : eigene.absteigend}
        </Knopf>
      )}
    </div>
  )
}
