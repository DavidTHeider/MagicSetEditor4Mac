// **Was Aufmerksamkeit braucht** — Werkstück U11 von Phase 5.1
// (#1422).
//
// # Aus Befunden, nicht aus einer Vermutung
//
// Ein Hinweis steht hier, weil der Bestand ihn hergibt: ein Set mit
// Vorbehalten aus dem Lesen, ein Deck, dem Karten in der Sammlung fehlen, ein
// Paket mit einem Skriptbefund. Gezählt wird in Go (`internal/app/
// aufmerksamkeit.go`); diese Seite zeigt die Zahl und führt hin.
//
// # Der leere Zustand sagt, was geprüft wurde
//
// „Nichts" allein wäre eine Behauptung. Deshalb steht daneben, wie viele Sets,
// Decks und Pakete durchgesehen wurden — und was **nicht** durchgesehen werden
// konnte, steht als Befund darunter statt still zu fehlen (§8).
//
// # Fünf, und der Rest auf Wunsch
//
// UX-15 (#1771),
// gemessen am 2026-09-24 in WebKit mit dem echten Bestand: 536 Pakete mit
// einem Skriptbefund machten die Startseite 24.827 Punkte hoch. Das Artboard
// *Main* zeigt drei Hinweise und neben dem Titel ihre Zahl; hier stehen die
// ersten fünf, und keiner geht verloren — der Knopf darunter zeigt alle.
//
// # Geladen am Bestandsstand — eine Ebene höher
//
// Wie `Weiterarbeiten`: `useLaden` mit `stand`, damit ein anderer Prozess,
// der schreibt, die Hinweise erneuert (§7, Regel 2 über den Haken). **Seit
// AP-8, Teil 2b fragt `Start`** ([useAufmerksamkeit]): die Kachel „Decks“
// zählt die Decks mit fehlenden Karten aus derselben Antwort.
//
// # Der Grund in der Sprache des Fensters
//
// #1963,
// gefunden an der Hüllenprobe `leisteenglisch`: hier stand `hinweis.grund`,
// der deutsche Satz aus Go — auch unter „Needs attention". Jetzt reist der
// Schlüssel mit, und der Satz kommt aus dem Katalog.
import { useState } from 'react'

import { api } from '../api'
import { auspacken } from '../umschlag'
import { Flaeche } from '../bausteine/Flaeche'
import { Ladefehler, Wirdgeladen } from '../bausteine/Ladehinweise'
import { useLaden, type Ladezustand } from '../haken/useLaden'
import { useBetriebsband } from '../haken/useBetriebsband'
import { Betriebshinweise } from './Betriebshinweise'
import { aufmerksamkeitszeilen, grundsatz } from './aufmerksamkeitszeilen'
import type { Setbefund } from './problemlage'
import { Paketsammelzeile } from './Paketsammelzeile'
import { useAnschluss } from '../anschlussraum'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'
import type { AufmerksamkeitAntwort, HinweisDTO } from '../apiaufmerksamkeit'

interface AufmerksamkeitEigenschaften {
  waehleSet: (setID: string, karteID?: string) => void

  // uebersicht ist die geladene Antwort — **von `Start` gefragt und nicht
  // von hier**, weil die Kachel „Decks“ dieselbe liest (AP-8, Teil 2b):
  // einmal gefragt statt zweimal gerechnet ([useAufmerksamkeit]).
  uebersicht: Ladezustand<AufmerksamkeitAntwort>

  // setbefunde sind die aktuell geprüften Sets mit Karten mit Befund — aus
  // den Prüfständen der Kacheln (AP-8, Teil 2b).
  setbefunde?: readonly Setbefund[] | undefined
}

// Gezeigt ist die Zahl der Hinweise, die ohne Klick dastehen.
const Gezeigt = 5

// zeilenknopf ist ein Hinweis der Liste — eine ganze Zeile, die hinführt.
const zeilenknopf =
  'w-full rounded-sm px-2 py-1 text-left text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

async function holen(): Promise<AufmerksamkeitAntwort> {
  const antwort = await api().Aufmerksamkeit()

  return auspacken(antwort, () => antwort)
}

// useAufmerksamkeit fragt die Übersicht — neu, wenn `stand` steigt.
export function useAufmerksamkeit(
  stand: number,
): Ladezustand<AufmerksamkeitAntwort> {
  return useLaden(holen, [], false, stand)
}

export function Aufmerksamkeit({
  setbefunde = [],
  waehleSet,
  uebersicht,
}: AufmerksamkeitEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.start.aufmerksamkeit
  const woerter = texte.allgemein.zaehlwoerter
  const [alle, setzeAlle] = useState(false)

  // **Das Betriebsband steht mit in dieser Liste** (OPS02, #2348, Artboard
  // `Start09`) — es fragt selbst, alle 30 Sekunden, und auch dann, wenn die
  // Übersicht nicht zu laden war.
  const betrieb = useBetriebsband()

  // **Am Objekt, nicht auf der Seite** — AP-8 (#2365), B-14: ein Set geht
  // auf, ein Deck mit seiner Prüfung. Pakete kommen hier nicht an — sie
  // teilen eine Zeile, und die führt zur Paketliste, gefiltert auf genau sie.
  const anschluss = useAnschluss()
  const hin = (hinweis: HinweisDTO) => {
    if (hinweis.objekt.art === 'set') {
      waehleSet(hinweis.objekt.kennung)
    } else {
      anschluss.uebergib({
        art: 'deckansicht',
        name: hinweis.objekt.name,
        deckID: hinweis.objekt.kennung,
        lasche: 'pruefung',
      })
    }
  }

  // **Laden und Scheitern sagen sich, mit Rückweg** — G11 (#2277). Der
  // Vorsatz bleibt: auf der Startseite stehen zwei Abschnitte, und er sagt,
  // dass es ums Laden ging.
  if (uebersicht.fehler !== undefined) {
    return (
      <div className="flex flex-col gap-2">
        <Betriebshinweise hinweise={betrieb} />
        <Ladefehler
          fehler={einsetzen(texte.bausteine.zustaende.fehlerMitGrund, {
            grund: uebersicht.fehler,
          })}
          wiederholen={uebersicht.nachladen}
        />
      </div>
    )
  }

  const daten = uebersicht.daten
  if (daten === undefined) {
    return uebersicht.laedt ? <Wirdgeladen /> : null
  }

  const zeilen = aufmerksamkeitszeilen(daten.hinweise, setbefunde)
  const anzahl = daten.hinweise.length + setbefunde.length

  const gezaehlt = {
    sets: zaehlwort(daten.durchgesehen.sets, woerter.sets),
    decks: zaehlwort(daten.durchgesehen.decks, woerter.decks),
    pakete: zaehlwort(daten.durchgesehen.pakete, woerter.eigenePakete),
  }

  return (
    // **Ohne Abstand nach oben** (UXR-03): die Startseite stellt diesen Bereich
    // neben *Weiterarbeiten* in ein Raster, und das Raster hält den Abstand.
    <section aria-label={eigene.titel} className="flex flex-col gap-4">
      <h2 className="flex items-center gap-2 text-lg font-semibold">
        {eigene.titel}{' '}
        {anzahl + betrieb.length > 0 && (
          <span className="rounded-full border border-kante px-2 py-0.5 font-text text-xs font-medium text-leise">
            {anzahl + betrieb.length}
          </span>
        )}
      </h2>

      <Flaeche>
        <Betriebshinweise hinweise={betrieb} />
        {anzahl === 0 && betrieb.length === 0 ? (
          <p className="text-sm text-leise">
            {einsetzen(eigene.nichts, gezaehlt)}
          </p>
        ) : (
          <>
            {zeilen.length > 0 && (
              <ul className="flex flex-col gap-1">
                {(alle ? zeilen : zeilen.slice(0, Gezeigt)).map((zeile) => {
                  if (zeile.art === 'setbefund') {
                    const { befund } = zeile
                    const satz = einsetzen(eigene.setBefunde, {
                      karten: zaehlwort(befund.zahl, woerter.karten),
                    })

                    return (
                      <li key={`befund:${befund.setID}`}>
                        <button
                          type="button"
                          className={zeilenknopf}
                          onClick={() => {
                            // **Das Set auf, die Liste gefiltert** — wie
                            // `Start09` es zeichnet.
                            waehleSet(befund.setID, befund.karteIDs[0])
                            anschluss.uebergib({
                              art: 'kartenauswahl',
                              name: satz,
                              setID: befund.setID,
                              karteIDs: befund.karteIDs,
                            })
                          }}
                        >
                          <span className="font-medium">{befund.name}</span>{' '}
                          <span className="text-leise">
                            {texte.seiten.start.weiterarbeiten.arten.set}
                            {' · '}
                            {satz}
                          </span>
                        </button>
                      </li>
                    )
                  }

                  if (zeile.art === 'pakete') {
                    return (
                      <li key="pakete">
                        <Paketsammelzeile
                          hinweise={zeile.hinweise}
                          klasse={zeilenknopf}
                        />
                      </li>
                    )
                  }

                  const { hinweis } = zeile

                  return (
                    <li key={`${hinweis.objekt.art}:${hinweis.objekt.kennung}`}>
                      <button
                        type="button"
                        className={zeilenknopf}
                        onClick={() => {
                          hin(hinweis)
                        }}
                      >
                        <span className="font-medium">
                          {hinweis.objekt.name}
                        </span>{' '}
                        <span className="text-leise">
                          {
                            texte.seiten.start.weiterarbeiten.arten[
                              hinweis.objekt.art
                            ]
                          }
                          {' · '}
                          {einsetzen(eigene.arten[hinweis.art], {
                            anzahl: String(hinweis.anzahl),
                          })}
                        </span>
                        {hinweis.grund !== '' && (
                          <span className="block text-xs text-leise">
                            {grundsatz(hinweis, texte)}
                          </span>
                        )}
                      </button>
                    </li>
                  )
                })}
              </ul>
            )}
            {zeilen.length > Gezeigt && (
              <button
                type="button"
                aria-expanded={alle}
                className="mt-2 self-start rounded-sm px-2 py-1 text-sm underline focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                onClick={() => {
                  setzeAlle((bisher) => !bisher)
                }}
              >
                {alle
                  ? eigene.weniger
                  : einsetzen(eigene.alle, {
                      anzahl: String(zeilen.length),
                    })}
              </button>
            )}
            <p className="mt-3 text-sm text-leise">
              {einsetzen(eigene.durchgesehen, gezaehlt)}
            </p>
          </>
        )}

        {daten.befunde.length > 0 && (
          <div className="mt-3 text-sm text-leise">
            <p>{eigene.befunde}</p>
            <ul className="list-disc pl-5">
              {daten.befunde.map((befund) => (
                <li key={befund}>{befund}</li>
              ))}
            </ul>
          </div>
        )}
      </Flaeche>
    </section>
  )
}
