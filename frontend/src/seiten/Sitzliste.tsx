// **Wer sitzt am Tisch** — die erste Tafel des Artboards *Eine Partie
// anfangen* (PLAY-01, S11.NAT-01).
//
// # Warum ein eigenes Bauteil
//
// **Weil `Partieanfang.tsx` an der Zeilengrenze aus §2 stand**, und der
// Schnitt folgt der Frage: hier, *wer* an welchem Sitz mit welchem Deck
// sitzt; dort, *wie* aus den Sitzen eine Partie wird. Den Zustand hält der
// Aufrufer — er schickt die Sitze, und zwei Wahrheiten über dieselben Sitze
// liefen auseinander.
import { useId } from 'react'

import {
  Beschriftetesfeld,
  eingabeklasse,
} from '../bausteine/Beschriftetesfeld'
import { Auswahlfeld } from '../bausteine/Auswahlfeld'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Flaeche } from '../bausteine/Flaeche'
import { Knopf } from '../bausteine/Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import {
  deckwert,
  einladewert,
  gewaehlterSitz,
  leihwert,
} from '../spiel/sitzwahl'

import type { DeckDTO } from '../apidecks'
import type { Partiekontakt } from '../haken/usePartiekontakte'
import type { Leihe } from '../apileihe'
import type { ProfilDTO } from '../apinutzer'

// Sitzwunsch ist ein Sitz, wie das Formular ihn führt.
export interface Sitzwunsch {
  name: string
  deckID: string

  // profilID bindet diesen Sitz an ein Profil **desselben Nutzers** (G8,
  // #1465).
  // **Leer heißt: keines**, und das bleibt der Normalfall.
  profilID: string

  // leiheID setzt eine angenommene Leihe an die Stelle des Decks — P1b
  // (#2370).
  // **Entweder oder**: mit ihr ist `deckID` leer.
  leiheID: string

  // eingeladen ist die Peerkennung eines Gastes, der sein Deck mitbringt —
  // P9c (#2377).
  // Mit ihr sind Deck und Leihe leer.
  eingeladen?: string
}

// sitzkennung gibt die Kennung zur Stelle — `sitz1`, `sitz2`, …
//
// **Aus der Stelle und nicht aus dem Namen**: zwei Spieler dürfen gleich
// heißen, zwei Sitze nicht dieselbe Kennung tragen.
export function sitzkennung(stelle: number): string {
  return `sitz${String(stelle + 1)}`
}

// zeichen ist, was im Kreis vor einem Sitz steht — der erste Buchstabe des
// Namens, und die Nummer, solange keiner dasteht.
function zeichen(name: string, stelle: number): string {
  const erster = name.trim().charAt(0)

  return erster === '' ? String(stelle + 1) : erster.toUpperCase()
}

export function Sitzliste({
  wuensche,
  setzeWunsch,
  hinzufuegen,
  wegnehmen,
  decks,
  leihen,
  kontakte = [],
  profile,
  auskunft,
}: {
  wuensche: readonly Sitzwunsch[]
  setzeWunsch: (stelle: number, neu: Sitzwunsch) => void
  hinzufuegen: () => void
  wegnehmen: (stelle: number) => void
  decks: readonly DeckDTO[]

  // leihen sind die angenommenen Leihen, die sich spielen lassen — P1b.
  leihen: readonly Leihe[]

  // kontakte sind die, die sich an einen Sitz einladen lassen — P9c.
  kontakte?: readonly Partiekontakt[]
  profile: readonly ProfilDTO[]

  // auskunft ist die Zeile über Spiel, Spielart und Format — leer, solange
  // kein Deck gewählt ist.
  auskunft: string
}) {
  const eigene = useTexte().seiten.spielen
  const basis = useId()

  const zahl =
    wuensche.length === 1
      ? eigene.sitzeEiner
      : einsetzen(eigene.sitzeZahl, { zahl: String(wuensche.length) })

  return (
    <Flaeche ueberschrift={eigene.sitzeTitel} marke={zahl}>
      <ol className="flex flex-col gap-3">
        {wuensche.map((wunsch, stelle) => {
          const deck = decks.find((eines) => eines.id === wunsch.deckID)

          return (
            <li
              key={sitzkennung(stelle)}
              // **Die Beschriftungen stehen oben auf einer Linie**: WebKit
              // zeichnet ein Auswahlfeld niedriger als ein Textfeld, und am
              // unteren Rand ausgerichtet sprangen die Beschriftungen um
              // eine halbe Zeile — gesehen in WebKit am 2026-09-24.
              //
              // **Und ein Sitz ist eine Gruppe** (P01, #2165): Zeichen, Name,
              // Deck und Profil brechen gemeinsam um; „Wegnehmen" steht
              // daneben oben rechts und nicht als letztes Glied der Reihe —
              // gemessen am nativen Fenster stand es sonst schon bei 1280 px
              // allein in einer eigenen Zeile (`politur`).
              className="flex items-start gap-3 rounded-md border border-kante p-3"
            >
              <div
                role="group"
                aria-label={einsetzen(eigene.sitzGruppe, {
                  nummer: String(stelle + 1),
                })}
                className="flex min-w-0 flex-1 flex-wrap items-start gap-3"
              >
                {/* **Zeichen und Name haben ihre eigene Zeile** — ein Name
                    hat keine Längengrenze, und neben einem langen Deck blieb
                    ihm sonst die Mindestbreite (gemessen: abgeschnitten bei
                    1280 px). Jeder Sitz bricht damit gleich um. */}
                <div className="flex w-full items-start gap-3">
                  <span
                    aria-hidden="true"
                    className="mt-6 flex size-8 shrink-0 items-center justify-center rounded-full border border-kante text-sm font-semibold"
                  >
                    {zeichen(wunsch.name, stelle)}
                  </span>

                  <Beschriftetesfeld
                    wachsen
                    bezeichnung={`${eigene.sitzName} ${String(stelle + 1)}`}
                    wert={wunsch.name}
                    setzeWert={(name) => {
                      setzeWunsch(stelle, { ...wunsch, name })
                    }}
                  />
                </div>

                {/* **Das Deck ist ein Auswahlfeld** — UXR-F1
                  (#2007),
                  R7: wer fünfzehn Decks hat, tippt, statt zu rollen. Die
                  Beschriftung steht neben dem Feld und nicht um es. */}
                <div className="flex flex-col gap-1">
                  <label
                    htmlFor={`${basis}-${sitzkennung(stelle)}`}
                    className="text-sm font-medium"
                  >
                    {eigene.sitzDeck}
                  </label>
                  <Auswahlfeld
                    id={`${basis}-${sitzkennung(stelle)}`}
                    bezeichnung={eigene.sitzDeck}
                    eintraege={[
                      { wert: '', name: eigene.deckWaehlen },
                      ...decks.map((eines) => ({
                        wert: eines.id,
                        name: eines.name,
                      })),
                      // **Eine angenommene Leihe sitzt wie ein Deck** —
                      // P1b (#2370), C-12. Der Name nennt, von wem sie
                      // kommt: die letzten Zeichen der Peerkennung.
                      ...leihen.map((eine) => ({
                        wert: leihwert(eine.kennung),
                        name: einsetzen(eigene.sitzLeihe, {
                          von: eine.verleiher.slice(-8),
                          karten: String(eine.karten),
                        }),
                      })),
                      // **Ein Gast bringt sein Deck mit** — P9c (#2377),
                      // Vermerk 0199. Nie am ersten Sitz: dort sitzt, wer
                      // die Partie anlegt.
                      ...(stelle === 0 ? [] : kontakte).map((einer) => ({
                        wert: einladewert(einer.kennung),
                        name: einsetzen(eigene.sitzEinladen, {
                          name: einer.name,
                        }),
                      })),
                    ]}
                    wert={deckwert(wunsch)}
                    aendern={(wert) => {
                      setzeWunsch(
                        stelle,
                        gewaehlterSitz(wunsch, wert, kontakte),
                      )
                    }}
                    klasse={eingabeklasse}
                  />
                </div>

                {/* **Die Größe des Decks steht beim Sitz** — „60 + 15" wie im
                  Artboard, für ein Vorleseprogramm in Worten. */}
                {deck !== undefined && (
                  <span className="mt-6 rounded-full border border-kante px-2 text-xs text-leise">
                    <span aria-hidden="true">
                      {deck.hauptdeck} + {deck.sideboard}
                    </span>
                    <span className="sr-only">
                      {einsetzen(eigene.deckGroesse, {
                        haupt: String(deck.hauptdeck),
                        seite: String(deck.sideboard),
                      })}
                    </span>
                  </span>
                )}

                {/* **Das Profil ist freiwillig** (G8, #1465) und steht nur da,
                  wenn es etwas zu wählen gibt: bei einem einzigen Profil wäre
                  es ein Feld mit einer Wahl. */}
                {profile.length > 1 && (
                  <label className="flex flex-col gap-1">
                    <span className="text-sm font-medium">
                      {eigene.sitzprofil}
                    </span>
                    <select
                      value={wunsch.profilID}
                      onChange={(ereignis) => {
                        setzeWunsch(stelle, {
                          ...wunsch,
                          profilID: ereignis.target.value,
                        })
                      }}
                      className={eingabeklasse}
                    >
                      <option value="">{eigene.sitzprofilKeines}</option>
                      {profile.map((profil) => (
                        <option key={profil.id} value={profil.id}>
                          {profil.name}
                        </option>
                      ))}
                    </select>
                  </label>
                )}

                {/* **Ab zwei Sitzen lässt sich einer wegnehmen** — bis auf
                  einen herunter, denn allein üben ist eine Spielart (CR
                  100.1: ein Aufbau, kein Spiel) und kein Fehler. */}
              </div>

              {wuensche.length > 1 && (
                <span className="mt-5 shrink-0">
                  <Knopf
                    gewicht="leise"
                    aufKlick={() => {
                      wegnehmen(stelle)
                    }}
                  >
                    {eigene.sitzWeg}
                  </Knopf>
                </span>
              )}
            </li>
          )
        })}
      </ol>

      <div className="mt-3 flex flex-wrap items-center gap-3">
        <Knopf gewicht="leise" schreibt aufKlick={hinzufuegen}>
          {eigene.sitzHinzu}
        </Knopf>

        {auskunft !== '' && <p className="text-sm text-leise">{auskunft}</p>}
      </div>

      {/* **Was ein Profil am Sitz bedeutet, steht dabei** — und nur dort,
          wo es eines zu wählen gibt. */}
      {profile.length > 1 && (
        <Erklaerungsabsatz
          text={eigene.sitzprofilErklaerung}
          klasse="mt-3 max-w-prose text-sm text-leise"
        />
      )}
    </Flaeche>
  )
}
