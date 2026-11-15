// Die Stilauswahl eines Sets (#41).
//
// **Was sie tut, steht hier; wie sie aussieht, in den Klassen** (§2).
//
// # Warum zwei Schritte
//
// Ein Wechsel kann etwas kosten — Stiloptionen gehen verloren, und bei einem
// Spielwechsel bleiben Felder ungedeutet. Was er kostet, steht **vor** ihm:
// erst die Folgen, dann die Zustimmung. Dieselbe Form wie beim Paketbezug, und
// aus demselben Grund.
import { useCallback, useState } from 'react'

import { api, auspacken, type FolgenDTO, type StilDTO } from '../api'
import { useLaden } from '../haken/useLaden'
import { useLauf } from '../haken/useLauf'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { Knopf } from '../bausteine/Knopf'
import { Stilauswahl } from '../bausteine/Stilauswahl'
import { Stiloptionen } from './Stiloptionen'
import { Stilvorschau } from './Stilvorschau'

interface Eigenschaften {
  setID: string

  // geaendert meldet dem Aufrufer, dass die Karten neu zu zeichnen sind.
  //
  // **Ein Rückruf und kein Nachladen hier**: die Vorschau gehört der Kartenliste,
  // und zwei Ladewege auf dieselben Daten liefen auseinander (§7, Regel 2).
  geaendert: () => void

  // zuDenPaketen führt ohne installierten Stil dorthin, wo einer
  // herkommt (G02, #2249).
  zuDenPaketen?: () => void
}

export function Stilwahl({ setID, geaendert, zuDenPaketen }: Eigenschaften) {
  const texte = useTexte()
  const vorgang = useVorgang()

  // **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (§7, Regel 2).
  //
  // Ohne diesen Zähler baute eine Rückfrage für Set A ihre Zustimmung am Set B
  // auf, sobald sie verspätet zurückkam — und `StilWechseln` liefe mit der
  // Kennung von B. Der Zweitblick über Phase 7 hat es angestrichen.
  const lauf = useLauf()

  const holen = useCallback(async () => {
    const antwort = await api().Stile(setID)

    return auspacken(antwort, () => antwort.stile)
  }, [setID])
  const { daten, laedt, fehler, nachladen } = useLaden<StilDTO[]>(holen, [
    setID,
  ])

  const [geplant, setzeGeplant] = useState<string>()
  const [folgen, setzeFolgen] = useState<FolgenDTO>()
  const [handlungsfehler, setzeHandlungsfehler] = useState<string>()

  // stilstand steigt bei jeder Änderung am Stil — daran hängt die Vorschau
  // (#643).
  const [stilstand, setzeStilstand] = useState(0)

  // stilfassung steigt **nur** beim Stilwechsel, nicht bei jeder
  // Optionsänderung — aus dem Zweitblick über Phase L (2026-08-22).
  //
  // **Ein eigener Zähler und nicht `stilstand`.** Der steigt bei jeder
  // geschriebenen Option; als Schlüssel benutzt, baute er die Optionsliste bei
  // jeder Änderung neu auf und nähme dem Feld den Fokus. Was hier gebraucht
  // wird, ist der Wechsel des **Bezugsobjekts**: das Set in seinem Stil (§7,
  // Regel 1).
  const [stilfassung, setzeStilfassung] = useState(0)

  // getippt ist der Stand eines Textfeldes, während jemand darin tippt —
  // `undefined`, sobald geschrieben ist
  // (#590).
  //
  // **Hier und nicht in [Stiloptionen]**: die Vorschau steht neben der Liste,
  // nicht darin.
  const [getippt, setzeGetippt] = useState<{ name: string; wert: string }>()

  // **Ein Wechsel des Bezugsobjekts räumt den offenen Kontext** (§7, Regel 1):
  // wer ein anderes Set wählt, hat keine offene Rückfrage mehr über das alte.
  const [zuSet, setzeZuSet] = useState(setID)

  if (zuSet !== setID) {
    setzeZuSet(setID)
    setzeGeplant(undefined)
    setzeFolgen(undefined)
    setzeHandlungsfehler(undefined)

    // **Und der Lauf zählt weiter**, damit eine unterwegs befindliche Antwort
    // die geräumte Rückfrage nicht wieder aufbaut (§7, Regel 2).
    lauf.naechster()
  }

  const fragen = async (paket: string) => {
    const stand = lauf.naechster()

    await vorgang.starte('folgen', async () => {
      setzeHandlungsfehler(undefined)

      try {
        const antwort = await api().Stilfolgen(setID, paket)

        auspacken(antwort, () => undefined)

        // **Erfolg und Fehler gleichermaßen** (§7, Regel 2): eine verspätete
        // Rückfrage gehört zu dem Set, aus dem sie stammt.
        if (!lauf.istAktuell(stand)) {
          return
        }

        setzeGeplant(paket)
        setzeFolgen(antwort.folgen ?? undefined)
      } catch (ursache: unknown) {
        if (lauf.istAktuell(stand)) {
          setzeHandlungsfehler(meldungAus(ursache, texte))
        }
      }
    })
  }

  const wechseln = async (paket: string) => {
    await vorgang.starte('wechsel', async () => {
      setzeHandlungsfehler(undefined)

      try {
        const antwort = await api().StilWechseln(setID, paket)

        auspacken(antwort, () => undefined)

        setzeGeplant(undefined)
        setzeFolgen(undefined)

        // **Der neue Stil hat andere Optionen** — bis zum 2026-08-22 blieb die
        // Liste des alten stehen, und die Vorschau daneben ebenso. Der
        // Zweitblick über Phase L hat es gefunden.
        setzeStilfassung((bisher) => bisher + 1)

        // **Geschrieben ist geschrieben** (§7, Regel 3): der Stil steht, auch
        // wenn das Nachladen danach scheitert.
        nachladen()
        geaendert()
      } catch (ursache: unknown) {
        setzeHandlungsfehler(meldungAus(ursache, texte))
      }
    })
  }

  return (
    // **Ohne Trennlinie oben** — Werkstück C von Phase 4.10.
    //
    // Sie stammt aus der Zeit, in der die Stilwahl **unter** der Kartenliste
    // stand (#288) und sich von ihr absetzen musste. Seit sie eine eigene
    // Lasche ist, trennt sie nichts von nichts: im Abzug vom 2026-09-03 stand
    // sie als leere Linie über der Überschrift und kostete 45 Punkte.
    <section>
      <h2 className="text-lg font-semibold">{texte.seiten.karten.stilTitel}</h2>

      {laedt && <p className="mt-2">{texte.bausteine.zustaende.laedt}</p>}

      {fehler !== undefined && (
        <p role="alert" className="mt-2 text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      {handlungsfehler !== undefined && (
        <p role="alert" className="mt-2 text-fehler">
          {texte.bausteine.zustaende.fehler} {handlungsfehler}
        </p>
      )}

      {daten?.length === 0 && (
        <div className="mt-2 flex max-w-prose flex-col items-start gap-2">
          <p className="text-leise">{texte.seiten.karten.stilLeer}</p>

          {zuDenPaketen !== undefined && (
            <Knopf aufKlick={zuDenPaketen}>
              {texte.allgemein.arbeitsplatz.zuDenPaketen}
            </Knopf>
          )}
        </div>
      )}

      {/* **Dieselbe Auswahl wie beim neuen Set** — Name und Kennung (UX-05,
          S07.NAT-02, #1677) —, **ein Feld mit Suche statt eines Optionsfelds
          je Stil**: das waren in dieser Schublade gemessen 382 Felder (UXR-07,
          #1915; R3 und R7 aus `docs/UIREVIEW.md`), seit UXR-F1 (#2007) auf der
          `Suchauswahl`. Gewählt ist, was gefragt wird, sonst, was das Set
          führt. **Führt es keinen installierten**, sagt es der Satz darüber,
          und das Feld steht leer.

          **Die Eingabe in der Auswahl räumt der Setwechsel mit** (§7, Regel 1):
          `useLaden` räumt `daten`, die Auswahl wird abgehängt und mit ihr ihr
          Entwurf. Bis UXR-F1 hielt diese Seite die Suche selbst und räumte
          sie in dem Block oben; `stilwahl.test.tsx` hält die Wirkung. */}
      {daten !== undefined && daten.length > 0 && (
        <div className="mt-2 flex flex-col gap-2">
          {!daten.some((stil) => stil.gewaehlt) && (
            <p className="text-sm text-leise">
              {texte.seiten.karten.stilKeiner}
            </p>
          )}

          <Stilauswahl
            beschriftung={texte.seiten.karten.stilWaehlen}
            eintraege={daten.map((stil) => ({
              kennung: stil.paket,
              anzeigename: stil.anzeigename,
              ...(stil.eigenesSpiel ? {} : { zusatz: stil.spiel }),
            }))}
            gewaehlt={
              geplant ?? daten.find((stil) => stil.gewaehlt)?.paket ?? ''
            }
            waehlen={(paket) => {
              void fragen(paket)
            }}
          />
        </div>
      )}

      {geplant !== undefined && folgen !== undefined && (
        <div className="mt-3 max-w-prose rounded-sm border border-kante p-3">
          <p className="text-text">{texte.seiten.karten.stilFrage}</p>

          {folgen.spielwechsel && (
            <p className="mt-2 text-leise">
              {texte.seiten.karten.stilSpielwechsel}
            </p>
          )}

          {folgen.ungedeuteteFelder.length > 0 && (
            <>
              <p className="mt-2 text-leise">
                {texte.seiten.karten.stilUngedeutet}
              </p>
              <ul className="mt-1 font-mono text-sm">
                {folgen.ungedeuteteFelder.map((feld) => (
                  <li key={feld}>{feld}</li>
                ))}
              </ul>
            </>
          )}

          {folgen.verloreneOptionen.length > 0 && (
            <>
              <p className="mt-2 text-leise">
                {texte.seiten.karten.stilVerloreneOptionen}
              </p>
              <ul className="mt-1 font-mono text-sm">
                {folgen.verloreneOptionen.map((feld) => (
                  <li key={feld}>{feld}</li>
                ))}
              </ul>
            </>
          )}

          {/* **Das Übernehmen führt** — eine führende Handlung je Region
              (R1, `docs/UIREVIEW.md`), wie im Artboard *Eigenschaften*; das
              Behalten steht daneben (UXR-07, #1915). */}
          <div className="mt-3 flex gap-3">
            <Knopf
              gewicht="haupt"
              klein
              angehalten={vorgang.laeuft('wechsel')}
              aufKlick={() => {
                void wechseln(geplant)
              }}
            >
              {texte.seiten.karten.stilUebernehmen}
            </Knopf>

            <Knopf
              klein
              aufKlick={() => {
                setzeGeplant(undefined)
                setzeFolgen(undefined)
              }}
            >
              {texte.seiten.karten.stilAbbrechen}
            </Knopf>
          </div>
        </div>
      )}

      {/* **Die Vorschau steht neben den Optionen** — Werkstück C von Phase
          4.10 (#1325).

          # Was vorher war

          Sie stand **unter** ihnen. Die Lasche ist 5,6 Bildschirme hoch und
          trägt 98 Tabstellen; wer eine Option ändert, sah die Wirkung erst
          nach dem Scrollen. §6 verlangt, dass eine Änderung in der Vorschau
          zu sehen ist, ohne dass jemand etwas bestätigt — sie **war** zu
          sehen, nur nicht von dort, wo man tippt.

          # Warum die Vorschau im Baum zuerst steht

          Damit sie auch dann oben landet, wenn das Gitter bei wenig Platz
          auf **eine** Spalte umbricht — und damit ein Vorleseprogramm sie
          zuerst nennt. Bei viel Platz schiebt `@3xl:order-2` sie nach
          rechts.

          **Und sie bleibt stehen** (`@3xl:sticky`): wer durch achtundneunzig
          Bedienelemente rollt, soll die Karte dabei nicht verlieren.

          # Der Platz und nicht das Fenster

          **Die Spalten fragen ihren Behälter, nicht das Fenster** (UX-08,
          #1680). Die Seite steht in der Schublade *Eigenschaften*, 460
          Punkte breit; `lg:` fragte das Fenster, und bei 1.280 Punkten
          stünden die Optionen neben einer 22rem breiten Vorschau in gut
          fünfzig Punkten. `@3xl:` fragt den nächsten `@container` — das
          Feld der Schublade. */}
      <div className="mt-4 grid gap-6 @3xl:grid-cols-[1fr_22rem]">
        <div className="@3xl:order-2 @3xl:sticky @3xl:top-4 @3xl:self-start">
          <Stilvorschau
            setID={setID}
            stand={stilstand + stilfassung}
            probe={getippt}
          />
        </div>

        <div className="@3xl:order-1">
          <Stiloptionen
            key={stilfassung}
            setID={setID}
            getippt={setzeGetippt}
            geaendert={() => {
              // **Beides**: der Aufrufer zeichnet seine Kartenliste neu, und
              // die Vorschau daneben rechnet mit (§6 — ohne Bestätigung).
              setzeStilstand((bisher) => bisher + 1)
              geaendert()
            }}
          />
        </div>
      </div>
    </section>
  )
}
