// **Bedarf und Zuordnung eines Decks** — Werkstück U9 von Phase 5.1
// (#1420).
//
// # Die eine Zusage, die diese Fläche halten muss
//
// **Bedarf ist nicht Zuordnung.** Zwei Decks mit denselben drei Blitzen sind
// beide vollständig — die Deckliste reserviert nichts. Erst wer zuordnet,
// nimmt die Stücke aus dem Freien, und dann sagt die Zeile beides: was da ist
// und was noch **frei** ist. Zwei Zahlen nebeneinander, keine davon falsch.
//
// # Ein Satz oben, die Zeilen dahinter
//
// **UXR-11** (#1919).
// Hier stand zu jedem Druck ein Formular aus Exemplar, Menge und Knopf — bei
// einem Commanderdeck 76 Zeilen, 152 Felder und 76 Knöpfe in einer Spalte von
// 320 Punkten (R3, R4 und R6 in `docs/UIREVIEW.md`). Das Artboard *Decks*
// zeigt an dieser Stelle einen Satz und den Weg zum Tausch. Jetzt steht der
// Satz oben samt seinen Wegen, die Zeilen dahinter **eingeklappt**, und das
// Formular einer Zeile erst, wenn jemand an ihr zuordnen will — eines
// zugleich.
import { useCallback, useId, useState } from 'react'

import { Aufklappkasten } from './Aufklappkasten'
import { Knopf } from './Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { Vorlaeufig } from './Vorlaeufig'
import { Zuordnungswahl } from './Zuordnungswahl'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { Deckstand } from '../haken/useDecks'
import type { DeckDTO, ZuordnungszeileDTO } from '../apidecks'
import type { ExemplarDTO } from '../api'
import { Tauschwunschknopf } from './Tauschwunschknopf'
import { Wunschlistenknopf } from './Wunschlistenknopf'

export function Deckzuordnung({
  stand,
  deck,
}: {
  stand: Deckstand
  deck: DeckDTO
}) {
  const eigene = useTexte().seiten.decks
  const [fehlkarten, setzeFehlkarten] = useState<string | null>(null)

  // offeneZeile ist die eine Zeile, deren Formular gerade steht.
  const [offeneZeile, setzeOffeneZeile] = useState<string>()
  const formularID = useId()

  const schluessel = `zuordnung:${deck.id}`

  // **Über `useLaden` und keinen eigenen Effekt** (§7, Regel 2): der Haken
  // führt den Stand mit, verwirft eine überholte Antwort und räumt beim
  // Deckwechsel. Die Sammlung kommt **im selben Zug** — welche Exemplare zu
  // einer Zeile passen, entscheidet die Anzeige; **ob** eines zugeordnet
  // werden darf, entscheidet der Dienst (§2).
  const holen = useCallback(async () => {
    const antwort = await api().DeckZuordnungen(deck.id)
    const stand = auspacken(antwort, () => antwort)

    const sammlung = await api().Sammlung()

    return {
      stand,
      exemplare: auspacken(sammlung, () => sammlung.exemplare),
    }
  }, [deck.id])

  const geladen = useLaden(holen, [deck.id])
  const zuordnungen = geladen.daten?.stand ?? null
  const exemplare = geladen.daten?.exemplare ?? []

  const laden = () => {
    geladen.nachladen()
  }

  const zuordnen = async (
    platzID: string,
    exemplarID: string,
    menge: number,
  ) => {
    await stand.fuehre(schluessel, async () => {
      const antwort = await api().DeckZuordnen(
        deck.id,
        platzID,
        exemplarID,
        menge,
      )

      auspacken(antwort, () => antwort.zuordnung)

      // **Geschrieben ist geschrieben** (§7, Regel 3): nachgeladen wird nur
      // nach einem geglückten Schreiben — `auspacken` wirft vorher.
      laden()
    })
  }

  const loesen = async (zuordnungID: string) => {
    await stand.fuehre(schluessel, async () => {
      const antwort = await api().DeckZuordnungLoesen(deck.id, zuordnungID)

      auspacken(antwort, () => null)
      laden()
    })
  }

  const holeFehlkarten = async () => {
    const antwort = await api().DeckFehlkarten(deck.id)

    setzeFehlkarten(auspacken(antwort, () => antwort.liste))
  }

  if (geladen.fehler !== undefined) {
    return <p role="alert">{geladen.fehler}</p>
  }

  if (zuordnungen === null) {
    return <p>{eigene.zuordnungenLaden}</p>
  }

  return (
    <div className="flex flex-col gap-3">
      <p className="font-medium">
        {zuordnungen.fehlend === 0
          ? eigene.bestandVollstaendig
          : einsetzen(eigene.bestandFehlt, {
              zahl: String(zuordnungen.fehlend),
            })}
      </p>

      {/* **„Die 7 eintauschen"** — UX-03.AK-01 (#1675), wie im Artboard
       *Decks*: was fehlt, reist als Wunsch zum Handel. */}
      <div className="flex flex-wrap items-start gap-3">
        <Tauschwunschknopf zeilen={zuordnungen.zeilen} />
        <Wunschlistenknopf
          stand={stand}
          deck={deck}
          fehlend={zuordnungen.fehlend}
        />
      </div>

      {zuordnungen.zeilen.length === 0 ? (
        <p>{eigene.zuordnungenLeer}</p>
      ) : (
        <Aufklappkasten titel={eigene.zuordnungen}>
          <div className="flex flex-col gap-3">
            <Erklaerungsabsatz
              text={eigene.zuordnungHinweis}
              klasse="text-sm text-leise"
            />

            <ul aria-label={eigene.zuordnungen} className="flex flex-col gap-3">
              {zuordnungen.zeilen.map((zeile, stelle) => {
                const kennung = zeile.printing + zeile.druckID + zeile.grund
                const offen = offeneZeile === kennung
                const passend = passende(exemplare, zeile)

                return (
                  <li key={kennung} className="flex flex-col gap-1">
                    <span className="flex items-center gap-2 font-medium">
                      {zeile.grund === '' ? zeile.bezeichnung : zeile.grund}
                      <Vorlaeufig vorlaeufig={zeile.vorlaeufig} />
                    </span>

                    <span className="text-sm text-leise tabular-nums">
                      {einsetzen(eigene.zuordnungszahlen, {
                        vorhanden: String(zeile.vorhanden),
                        gebraucht: String(zeile.gebraucht),
                        frei: String(zeile.frei),
                        zugeordnet: String(zeile.zugeordnet),
                        reserviert: String(zeile.reserviert),
                      })}
                    </span>

                    {zeile.zuordnungen.length > 0 && (
                      <ul className="flex flex-col gap-1 pl-4">
                        {zeile.zuordnungen.map((zuordnung) => (
                          <li
                            key={zuordnung.id}
                            className="flex items-center gap-2"
                          >
                            <span className="tabular-nums">
                              {zuordnung.menge} ×
                            </span>
                            <span>{zuordnung.bezeichnung}</span>

                            {zuordnung.abweichend && (
                              <span
                                className="rounded-sm border border-kante px-1 text-xs text-leise"
                                title={eigene.abweichendErklaerung}
                              >
                                {eigene.abweichend}
                              </span>
                            )}

                            <Knopf
                              bezeichnung={`${eigene.zuordnungLoesen}: ${zuordnung.bezeichnung}`}
                              angehalten={stand.vorgang.laeuft(schluessel)}
                              aufKlick={() => {
                                void loesen(zuordnung.id)
                              }}
                            >
                              {eigene.zuordnungLoesen}
                            </Knopf>
                          </li>
                        ))}
                      </ul>
                    )}

                    {/* **Nur, wo es etwas zuzuordnen gibt.** Ohne ein
                        passendes Exemplar stand hier eine Auswahl mit nichts
                        darin und ein Knopf, der nie ging. */}
                    {passend.length > 0 && (
                      <div>
                        <Knopf
                          klein
                          gewicht="leise"
                          bezeichnung={`${eigene.zuordnungExemplarWaehlen}: ${zeile.bezeichnung}`}
                          ausgeklappt={offen}
                          steuert={`${formularID}-${String(stelle)}`}
                          aufKlick={() => {
                            setzeOffeneZeile(offen ? undefined : kennung)
                          }}
                        >
                          {eigene.zuordnungExemplarWaehlen}
                        </Knopf>
                      </div>
                    )}

                    {offen && (
                      <div id={`${formularID}-${String(stelle)}`}>
                        <Zuordnungswahl
                          deck={deck}
                          zeile={zeile}
                          exemplare={passend}
                          angehalten={stand.vorgang.laeuft(schluessel)}
                          aufZuordnen={(platzID, exemplarID, menge) => {
                            void zuordnen(platzID, exemplarID, menge)
                          }}
                        />
                      </div>
                    )}
                  </li>
                )
              })}
            </ul>
          </div>
        </Aufklappkasten>
      )}

      <div className="flex flex-wrap items-start gap-3">
        <Knopf
          bezeichnung={eigene.fehlkarten}
          aufKlick={() => {
            void holeFehlkarten()
          }}
        >
          {eigene.fehlkarten}
        </Knopf>

        {fehlkarten !== null && (
          <pre
            role="region"
            aria-label={eigene.fehlkartenliste}
            tabIndex={0}
            className="max-h-40 overflow-auto rounded-md border border-kante p-2 text-sm whitespace-pre-wrap focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            {fehlkarten === '' ? eigene.fehlkartenKeine : fehlkarten}
          </pre>
        )}
      </div>
    </div>
  )
}

// passende gibt die Exemplare, die zu einer Zeile gehören.
//
// **Der Druck entscheidet, nicht der Name** (A2, #1441): über den
// Scryfall-Alias, sonst über die Druckzeile. Wünsche stehen nicht dabei — sie
// sind kein Besitz ([0057](../../docs/entscheidungen/0057-ein-wunsch-ist-kein-besitz.md)),
// und der Dienst wiese sie ohnehin ab.
function passende(
  exemplare: readonly ExemplarDTO[],
  zeile: ZuordnungszeileDTO,
): readonly ExemplarDTO[] {
  return exemplare.filter(
    (exemplar) =>
      !exemplar.wunsch &&
      (zeile.printing === ''
        ? exemplar.druckID !== '' && exemplar.druckID === zeile.druckID
        : exemplar.printing === zeile.printing),
  )
}
