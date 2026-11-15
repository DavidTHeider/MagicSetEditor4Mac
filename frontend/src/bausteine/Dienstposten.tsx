// **Einen der zwei Dienste schalten** — Werkstück N6 von Phase 5.4
// (#1572).
//
// # Ein Bauteil für beide
//
// **Weil es dieselbe Frage ist**: läuft er, wo, und wie kommt man hin. REST
// und MCP unterscheiden sich in ihrem Namen und in dem, was sie bedienen —
// nicht darin, wie man sie schaltet. Zwei Kästen wären zwei Stellen, an denen
// dieselbe Regel steht.
//
// # Und warum die Adresse ein Feld ist und keine Wahl
//
// **Weil `0.0.0.0` nicht die einzige Antwort ist.** Wer eine bestimmte
// Schnittstelle meint, nennt ihre Adresse — und eine Liste, die sie aufzählte,
// wäre eine Liste, die dieser Rechner beim nächsten WLAN anders hätte.
import { useId, useState } from 'react'

import { Aktionsknopf } from './Aktionsknopf'
import { Feld } from './Feld'
import { Knopf } from './Knopf'
import { api } from '../fassade'
import { ganzzahlAus } from '../ganzzahl'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { useDarstellung, useTexte, useZeitzone } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { alsZeitpunkt } from '../zeitpunkt'
import { inDieZwischenablage } from '../zwischenablage'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { DienstAntwort, DienstZustand } from '../apidienste'
import { kennungseingabe } from './kennungseingabe'

interface DienstpostenEigenschaften {
  // art ist „rest" oder „mcp" — sie wählt die Bindungen und die Texte.
  art: 'rest' | 'mcp'
}

// hoechsterPort ist die Obergrenze, die `domain.PruefeBindung` zieht
// (`b.Port > 65535`). **Die Null gehört dazu**: sie heißt „einen freien
// suchen" und ist die Vorbelegung des Feldes.
const hoechsterPort = 65535

// portAus liest einen getippten Port — oder gibt `undefined`, wenn der Text
// keiner sein kann.
//
// **Nur die Gestalt, nicht die Lage** (F03, #2269): ob ein Port frei ist,
// weiß erst der Dienst, und dessen Absage bleibt eine allgemeine Meldung.
// Streng gelesen über `ganzzahlAus` — `parseInt` machte aus „8080x" den Port
// 8080 (#2245) —, und ein Wert über der Grenze ist kein Port, den man erst
// hinausschicken müsste, um das zu erfahren.
function portAus(roh: string): number | undefined {
  const zahl = ganzzahlAus(roh)

  return zahl !== undefined && zahl <= hoechsterPort ? zahl : undefined
}

export function Dienstposten({ art }: DienstpostenEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.technik
  const darstellung = useDarstellung()
  const zeitzone = useZeitzone()
  const vorgang = useVorgang()
  const ueberschrift = useId()

  const [geaendert, setzeGeaendert] = useState(0)
  const [fehler, setzeFehler] = useState<string>()
  // **Leer heißt die Lauschwahl** (G08, #2272) — eine Adresse hier gilt
  // nur für diesen Start.
  const [adresse, setzeAdresse] = useState('')
  const [port, setzePort] = useState('0')

  // portAbgewiesen merkt, dass ein Klick den Port als unmöglich abgewiesen
  // hat — **für den Fall, den die Eingabe allein nicht zeigt**: ein leeres
  // Feld ist mitten im Ändern kein Fehler, nach dem Absenden aber einer.
  // Jede neue Eingabe nimmt es zurück.
  const [portAbgewiesen, setzePortAbgewiesen] = useState(false)
  const [kopiert, setzeKopiert] = useState<'marke' | 'abdruck'>()

  const stand = useLaden<DienstZustand>(async () => {
    const antwort =
      art === 'rest' ? await api().RESTStand() : await api().MCPStand()

    return auspacken(antwort, () => antwort.stand)
  }, [geaendert, art])

  const laeuft = vorgang.laeuft(`dienst-${art}`)

  const schalte = (fahre: () => Promise<DienstAntwort>) => {
    void vorgang.starte(`dienst-${art}`, async () => {
      setzeFehler(undefined)

      try {
        const antwort = await fahre()

        auspacken(antwort, () => undefined)

        setzeGeaendert((zahl) => zahl + 1)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  const posten = stand.daten

  // **Der Befund steht am Feld** (§6, F03, #2269) und nicht als allgemeine
  // Meldung darüber: wer „8080x" tippt, soll dort lesen, was nicht stimmt,
  // wo er es berichtigt. Schon beim Tippen, wenn etwas dasteht, das kein Port
  // sein kann; für ein leeres Feld erst nach dem Klick.
  const gewuenscht = portAus(port)
  const portbefund =
    gewuenscht === undefined && (port.trim() !== '' || portAbgewiesen)
      ? einsetzen(texte.allgemein.ganzeZahlBereich, {
          kleinstens: '0',
          groesstens: String(hoechsterPort),
        })
      : undefined

  // **Einmal gerechnet und nicht je Knopf.** Zwei Knöpfe, dieselbe Regel:
  // gemerkt wird nur, was wirklich in der Zwischenablage landete — sonst
  // stünde „Kopiert" über einer leeren. Der erste Anlauf hatte sie zweimal,
  // und die Abdeckung fand prompt den Zweig, den kein Lauf betrat.
  const kopiere = (was: 'marke' | 'abdruck', text: string) => () => {
    void inDieZwischenablage(text).then((geklappt) => {
      setzeKopiert(geklappt ? was : undefined)
    })
  }

  return (
    <section
      aria-labelledby={ueberschrift}
      className="rounded-lg border border-kante bg-flaeche p-4"
    >
      <h3 id={ueberschrift} className="text-lg font-semibold">
        {art === 'rest' ? eigene.dienstRESTTitel : eigene.dienstMCPTitel}
      </h3>

      <p className="mt-1 max-w-prose text-leise">
        {art === 'rest'
          ? eigene.dienstRESTErklaerung
          : eigene.dienstMCPErklaerung}
      </p>

      {fehler !== undefined && (
        <p role="alert" className="mt-3 text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      {posten?.laeuft ? (
        <div className="mt-3">
          <p>
            <span className="text-leise">{`${art === 'rest' ? eigene.dienstRESTAdresse : eigene.dienstMCPAdresse}: `}</span>
            <span className="break-all font-mono">{posten.adresse}</span>
          </p>

          {/* **Jede Adresse der Wahl und jede, die fehlt** (G08, #2272). */}
          {posten.lauscher.length > 1 && (
            <p className="mt-1">
              <span className="text-leise">{`${eigene.schnittstellen.lauschwahl.weitere}: `}</span>
              <span className="break-all font-mono">
                {posten.lauscher.slice(1).join(', ')}
              </span>
            </p>
          )}
          {posten.fehlende.map((eine) => (
            <p key={eine.adresse} className="mt-1 text-fehler">
              {einsetzen(eigene.schnittstellen.lauschwahl.ohne, {
                adresse: eine.adresse,
                grund: eine.grund,
              })}
            </p>
          ))}

          <p className="mt-1">
            <span className="text-leise">{`${eigene.dienstMarke}: `}</span>
            <span className="break-all font-mono">{posten.marke}</span>
          </p>

          {/* **Der Fingerabdruck steht da, wo TLS gilt** — jenseits der
              Schleife ist er der einzige Nachweis, den ein Gegenüber
              vergleichen kann (0115). Ohne ihn stünde daneben ein Satz, der
              zu etwas auffordert, das nirgends steht. */}
          {posten.fingerabdruck !== '' && (
            <p className="mt-1">
              <span className="text-leise">
                {`${eigene.dienstFingerabdruck}: `}
              </span>
              <span className="break-all font-mono">
                {posten.fingerabdruck}
              </span>
            </p>
          )}

          {/* **Und welches Zeugnis es ist** (NET-03, #1707): ein eigenes
              nennt seine Stelle und sein Ende; das der Identität keines. */}
          {posten.tls && (
            <p className="mt-1 max-w-prose text-sm text-leise">
              {posten.zeugnisAussteller === ''
                ? eigene.dienstzeugnis.identitaet
                : einsetzen(eigene.dienstzeugnis.eigenes, {
                    aussteller: posten.zeugnisAussteller,
                    bis: alsZeitpunkt(posten.zeugnisBis, darstellung, zeitzone),
                  })}
            </p>
          )}

          {!posten.tls && (
            <p className="mt-1 max-w-prose text-sm text-leise">
              {eigene.dienstOhneTLS}
            </p>
          )}

          {posten.tls && (
            <p className="mt-1 max-w-prose text-sm text-leise">
              {eigene.dienstMitTLS}
            </p>
          )}

          {posten.adressen.length > 0 && (
            <>
              <p className="mt-3 text-sm text-leise">
                {eigene.dienstErreichbar}
              </p>

              <ul className="mt-1 space-y-1 font-mono text-sm">
                {posten.namen.concat(posten.adressen).map((eintrag) => (
                  <li key={eintrag} className="break-all">
                    {eintrag}
                  </li>
                ))}
              </ul>
            </>
          )}

          <div className="mt-3 flex flex-wrap gap-2">
            <Knopf
              angehalten={laeuft}
              aufKlick={() => {
                schalte(() =>
                  art === 'rest' ? api().RESTBeenden() : api().MCPBeenden(),
                )
              }}
            >
              {eigene.dienstBeenden}
            </Knopf>

            <Aktionsknopf
              aktion="kopieren"
              angehalten={laeuft}
              aufKlick={kopiere('marke', posten.marke)}
            >
              {kopiert === 'marke'
                ? eigene.dienstKopiert
                : eigene.dienstMarkeKopieren}
            </Aktionsknopf>

            {posten.fingerabdruck !== '' && (
              <Aktionsknopf
                aktion="kopieren"
                angehalten={laeuft}
                aufKlick={kopiere('abdruck', posten.fingerabdruck)}
              >
                {kopiert === 'abdruck'
                  ? eigene.dienstKopiert
                  : eigene.dienstFingerabdruckKopieren}
              </Aktionsknopf>
            )}
          </div>
        </div>
      ) : (
        <div className="mt-3 flex flex-wrap items-end gap-3">
          <Feld
            beschriftung={
              art === 'rest'
                ? eigene.dienstRESTAdresse
                : eigene.dienstMCPAdresse
            }
          >
            {(verbindung) => (
              <input
                {...kennungseingabe}
                {...verbindung}
                value={adresse}
                placeholder={eigene.schnittstellen.lauschwahl.platzhalter}
                onChange={(ereignis) => {
                  setzeAdresse(ereignis.target.value)
                }}
                className="w-44 rounded-sm border border-kante bg-transparent px-2 py-1 font-mono focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              />
            )}
          </Feld>

          {/* **Über `Feld` und nicht mehr als bloße Beschriftung** (F03,
              #2269): der Befund steht unter dem Feld, `aria-invalid` und die
              Verbindung zum Satz setzt `Feld`. Die Adresse geht denselben
              Weg, damit beide Beschriftungen gleich aussehen. */}
          <Feld
            beschriftung={
              art === 'rest' ? eigene.dienstRESTPort : eigene.dienstMCPPort
            }
            {...(portbefund === undefined ? {} : { befund: portbefund })}
          >
            {(verbindung) => (
              <input
                {...kennungseingabe}
                {...verbindung}
                value={port}
                onChange={(ereignis) => {
                  setzePort(ereignis.target.value)
                  setzePortAbgewiesen(false)
                }}
                className="w-24 rounded-sm border border-kante bg-transparent px-2 py-1 font-mono focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring aria-[invalid=true]:border-fehler"
              />
            )}
          </Feld>

          <Knopf
            angehalten={laeuft}
            aufKlick={() => {
              // **Was kein Port sein kann, geht nicht hinaus** (F03,
              // #2269). Bis dahin wurde daraus `-1`, abgeschickt und von
              // `domain.PruefeBindung` als allgemeine Meldung über dem
              // Formular beantwortet — am Feld stand nichts. Der Knopf
              // bleibt bedienbar und sagt, was fehlt (§6).
              //
              // **Null ist ein Port** („einen freien suchen") und ein
              // leeres Feld keiner: auch das lehnt `portAus` ab, statt es
              // still umzudeuten.
              if (gewuenscht === undefined) {
                setzePortAbgewiesen(true)

                return
              }

              schalte(() =>
                art === 'rest'
                  ? api().RESTStarten(adresse, gewuenscht)
                  : api().MCPStarten(adresse, gewuenscht),
              )
            }}
          >
            {eigene.dienstStarten}
          </Knopf>
        </div>
      )}

      <Erklaerungsabsatz
        text={eigene.dienstHinweis}
        klasse="mt-3 max-w-prose text-sm text-leise"
      />
    </section>
  )
}
