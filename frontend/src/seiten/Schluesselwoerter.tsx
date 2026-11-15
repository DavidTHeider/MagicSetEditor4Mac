// Die **Schlüsselwörter eines Sets**
// (#1062,
// #2009).
//
// # Was MSE hier tut
//
// `src/gui/set/keywords_panel.cpp` (394 Zeilen, Fassung f9d9356 vom
// 2021-07-06): links eine Liste (`KeywordList`), rechts der Editor für das
// **gewählte** Wort, und der Unterschied zwischen **festen** (aus dem Spiel)
// und **eigenen** (aus dem Set) — die festen zeigt es grau und lässt sie nicht
// ändern.
//
// **Die Wahl**, wie MSE sie hält (`KeywordList::onAction`,
// `src/gui/control/keyword_list.cpp`, Zeilen 61–74, und
// `ItemList::selectItemPos`, `src/gui/control/item_list.cpp`, Zeilen 87–100):
//
// - beim Öffnen ist nichts gewählt,
// - *Hinzufügen* wählt das neue Wort,
// - *Entfernen* wählt das Wort an derselben Stelle — das folgende, beim
//   letzten das davor, beim einzigen keines.
//
// # Warum es hier an zwei Stellen anders aussieht — `adapted`
//
// **Die festen stehen nicht in der Liste.** MSE mischt sie unter die eigenen;
// hier sind es 334 gegen eine Handvoll, und sie bleiben die eingeklappte
// Liste darunter (`Schluesselwortfeste.tsx`). Wer das letzte eigene Wort
// entfernt, hat deshalb kein „folgendes aus dem Spiel" gewählt, sondern
// nichts.
//
// **Und das neue Wort nimmt den Fokus im Namen.** MSE lässt ihn in der Liste;
// hier ist der Name das Erste, was man an einem leeren Wort ausfüllt, und der
// Knopf, der den Fokus hatte, steht danach nicht mehr in der Reihenfolge.
//
// # Bis UXR-F3 stand jedes Wort als eigene Karte da
//
// Mit allen Feldern zugleich — ab drei Wörtern über R3 aus `docs/UIREVIEW.md`
// (höchstens zwölf Felder). Jetzt ist es **eine** Liste und **ein** Editor:
// fünf Wörter zeichnen fünf Felder, nicht fünfundzwanzig.
import { useCallback, useEffect, useRef, useState } from 'react'

import { api } from '../fassade'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Ladefeld } from './Ladefeld'
import { useLaden } from '../haken/useLaden'
import { useVorgang } from '../haken/useVorgang'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { nachrueckendeKarte } from './kartenauswahl'
import { Schluesselwortfeste } from './Schluesselwortfeste'
import { Schluesselworteditor } from './Schluesselworteditor'
import { Schluesselwortliste } from './Schluesselwortliste'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface SchluesselwoerterEigenschaften {
  setID: string

  // verlaufsstand steigt mit jedem Schritt im Verlauf (⌘Z).
  //
  // **Er baut die Liste neu auf, nicht die Wahl ab**: die Felder stehen auf
  // `defaultValue` und brauchen frische Werte, aber wer nach einer Rücknahme
  // das Wort wieder suchen müsste, das er eben bearbeitet hat, wäre dafür
  // bestraft. Bis UXR-F3 räumte der `key` in `Lascheninhalt` die ganze Seite.
  verlaufsstand?: number
}

// Meldung ist ein Satz über ein gescheitertes Schreiben — und das Wort, zu dem
// er gehört. **Eine Antwort gehört zu dem Stand, aus dem sie stammt** (§7,
// Regel 2): kommt sie, während ein anderes Wort offen ist, nennt sie ihres.
interface Meldung {
  text: string
  bezug?: { wortID: string; name: string }
}

// Wunsch sagt, wohin der Fokus will, sobald das Ziel im Baum steht.
//
// **Als Wunsch und nicht als Aufruf**, weil das Ziel noch nicht da sein muss:
// der Editor des neuen Wortes entsteht erst, wenn die Liste nachgeladen ist.
type Wunsch =
  { art: 'name' } | { art: 'zeile'; wortID: string } | { art: 'anlegen' }

export function Schluesselwoerter({
  setID,
  verlaufsstand = 0,
}: SchluesselwoerterEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.schluesselwoerter

  const [fehler, setzeFehler] = useState<Meldung>()

  // stand steigt nach jedem geglückten Schreiben — daran hängen die
  // Parameterhilfen (#1342): wer den Parameter ins Muster einfügt, bekommt
  // danach andere Verweise angeboten.
  const [stand, setzeStand] = useState(0)

  // gewaehlt ist die Kennung des gewählten Wortes — **eine Kennung und kein
  // Wort**: das Wort kommt aus der Liste, und wo es dort nicht mehr steht, ist
  // nichts gewählt.
  const [gewaehlt, setzeGewaehlt] = useState<string>()

  // jetzt spiegelt die Wahl für Antworten, die später eintreffen: wer ein Wort
  // löscht und inzwischen ein anderes gewählt hat, behält das andere.
  const jetzt = useRef<string>(undefined)
  const wunsch = useRef<Wunsch>(undefined)
  const wurzel = useRef<HTMLElement>(null)
  const editorknoten = useRef<HTMLDivElement>(null)
  const namensfeld = useRef<HTMLInputElement>(null)
  const anlegenknopf = useRef<HTMLButtonElement>(null)

  const vorgang = useVorgang()

  const holen = useCallback(async () => {
    const antwort = await api().Schluesselwoerter(setID)

    return auspacken(antwort, () => ({
      woerter: antwort.woerter,
      modi: antwort.modi,
    }))
  }, [setID])

  const liste = useLaden(holen, [setID], false, verlaufsstand)

  // **Der Editor baut sich zu einem Verlaufsstand nicht mehr neu** (U06,
  // #2122). Bis dahin trug sein Schlüssel den Stand der angekommenen Antwort
  // (#2069), damit nach einer Rücknahme der zurückgeholte Wert in den Feldern
  // stand — und jede Änderung irgendwo nahm dem, der tippte, den Entwurf.
  // Jetzt folgt jedes Feld selbst dem Bestand (`useEntwurf`), und die Wahl
  // des Modus trägt ihn im Schlüssel.

  const waehle = (wortID: string | undefined) => {
    jetzt.current = wortID
    setzeGewaehlt(wortID)
  }

  // **Der Fokus folgt seinem Wunsch, sobald das Ziel steht** — nach jedem
  // Zeichnen, denn wann der Editor des neuen Wortes entsteht, entscheidet die
  // Antwort auf das Nachladen und nicht dieser Bauteil.
  useEffect(() => {
    const ziel = wunsch.current
    if (ziel === undefined) {
      return
    }

    const knoten =
      ziel.art === 'name'
        ? namensfeld.current
        : ziel.art === 'anlegen'
          ? anlegenknopf.current
          : wurzel.current?.querySelector<HTMLElement>(
              `[data-wort="${ziel.wortID}"]`,
            )

    if (knoten === null || knoten === undefined) {
      return
    }

    wunsch.current = undefined
    knoten.focus()
  })

  // handeln fährt einen Schreibvorgang und lädt danach nach.
  //
  // **Ein Vorgang je Datensatz** (§7, Regel 4) — der Schlüssel ist das Set:
  // zwei gleichzeitige Änderungen an derselben Liste gingen gegen einen Stand,
  // den das Backend noch nicht kennt. **Die Wahl ist kein Vorgang**: sie darf
  // mitten in einer Speicherung geschehen.
  //
  // **Nachgeladen wird nur nach einem geglückten Schreiben** (Regel 3): sonst
  // überschriebe ein Lesefehler die Meldung, die sagt, was zu tun ist.
  //
  // `bezug` ist das Wort, zu dem das Schreiben gehört: scheitert es, trägt die
  // Meldung seinen Namen, falls inzwischen ein anderes offen ist.
  const handeln = async (
    tun: () => Promise<void>,
    bezug?: Meldung['bezug'],
  ) => {
    const geglueckt = await vorgang.starte(setID, async () => {
      try {
        await tun()

        return true
      } catch (ursache: unknown) {
        setzeFehler({
          text: meldungAus(ursache, texte),
          ...(bezug === undefined ? {} : { bezug }),
        })

        return false
      }
    })

    if (geglueckt === true) {
      setzeFehler(undefined)
      liste.nachladen()
      setzeStand((bisher) => bisher + 1)
    }
  }

  // speichereOffenes schließt die Eingabe im Editor ab, bevor er abgelöst
  // wird (Regel 3: geschrieben ist geschrieben).
  //
  // **Das Feld schreibt beim Verlassen**, und ein Klick verlässt es nicht
  // immer: WebKit gibt einem Knopf beim Klick keinen Fokus, das Feld behält
  // ihn — und der Editor würde mit dem Text darin abgeräumt.
  const speichereOffenes = () => {
    editorknoten.current?.querySelector<HTMLElement>(':focus')?.blur()
  }

  const auswaehlen = (wortID: string) => {
    speichereOffenes()
    waehle(wortID)
  }

  const woerter = liste.daten?.woerter ?? []
  const modi = liste.daten?.modi ?? []

  const feste = woerter.filter((wort) => wort.fest)
  const eigeneWoerter = woerter.filter((wort) => !wort.fest)

  const aktuell = eigeneWoerter.find((wort) => wort.id === gewaehlt)
  const bezug = { wortID: aktuell?.id ?? '', name: aktuell?.wort ?? '' }

  // **Gewählt, aber noch nicht geladen**: das neue Wort steht erst nach dem
  // Nachladen in der Liste, und bis dahin soll der Editor nichts Falsches
  // sagen — weder den vorigen noch „wähle eines".
  const wartet = gewaehlt !== undefined && aktuell === undefined && liste.laedt

  const sichtbar =
    fehler === undefined
      ? undefined
      : fehler.bezug === undefined || fehler.bezug.wortID === gewaehlt
        ? fehler.text
        : `${fehler.bezug.name === '' ? eigene.ohneName : fehler.bezug.name} — ${fehler.text}`

  return (
    <section ref={wurzel} className="flex flex-col gap-4">
      <header className="flex flex-col gap-1">
        <h2 className="text-xl font-semibold">{eigene.ueberschrift}</h2>
        <Erklaerungsabsatz
          text={eigene.erklaerung}
          klasse="text-sm text-leise"
        />
      </header>

      <div>
        <Aktionsknopf
          aktion="anlegen"
          bezeichnung={eigene.anlegen}
          angehalten={vorgang.laeuft(setID)}
          griff={anlegenknopf}
          aufKlick={() => {
            speichereOffenes()
            void handeln(async () => {
              const antwort = await api().SchluesselwortAnlegen(setID)
              const neu = auspacken(antwort, () => antwort.wortID)

              // **Gewählt wird das neue Wort** (MSE), und der Fokus geht in
              // seinen Namen.
              waehle(neu)
              wunsch.current = { art: 'name' }
            })
          }}
        >
          {eigene.anlegen}
        </Aktionsknopf>
      </div>

      {sichtbar !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {sichtbar}
        </p>
      )}

      <Ladefeld
        laedt={liste.laedt && liste.daten === undefined}
        fehler={liste.fehler}
        leer={eigeneWoerter.length === 0}
      >
        <p className="text-sm text-leise">{eigene.keine}</p>
      </Ladefeld>

      {eigeneWoerter.length > 0 && (
        // **Nebeneinander, wenn der Platz da ist** — wie bei MSE; in einer
        // schmalen Schublade untereinander.
        <div className="grid gap-3 @3xl:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
          <Schluesselwortliste
            woerter={eigeneWoerter}
            gewaehlt={gewaehlt}
            waehlen={auswaehlen}
          />

          <div ref={editorknoten}>
            {aktuell !== undefined ? (
              <Schluesselworteditor
                // **Die Kennung als `key`**: ein anderes Wort baut den Editor
                // neu — Felder auf `defaultValue`, offene Hilfen, alles. Das
                // ist Regel 1 für das Bezugsobjekt Wort.
                key={aktuell.id}
                wort={aktuell}
                modi={modi}
                angehalten={vorgang.laeuft(setID)}
                namensfeld={namensfeld}
                setID={setID}
                stand={stand}
                bearbeiten={{
                  schreiben: (feld, wert, bekannt) => {
                    void handeln(async () => {
                      auspacken(
                        await api().SchluesselwortSetzen(
                          setID,
                          aktuell.id,
                          feld,
                          wert,
                          bekannt ?? null,
                        ),
                        () => undefined,
                      )
                    }, bezug)
                  },
                  loeschen: () => {
                    // **Das folgende, beim letzten das davor** — vor dem
                    // Löschen bestimmt, solange die Liste noch steht, und
                    // mit der Funktion der Kartenliste: dieselbe `ItemList`.
                    const naechstes = nachrueckendeKarte(
                      eigeneWoerter,
                      aktuell.id,
                    )

                    void handeln(async () => {
                      auspacken(
                        await api().SchluesselwortLoeschen(setID, aktuell.id),
                        () => undefined,
                      )

                      // **Nur, wenn es noch das gewählte ist**: wer
                      // inzwischen ein anderes gewählt hat, behält es.
                      if (jetzt.current === aktuell.id) {
                        waehle(naechstes)
                        wunsch.current =
                          naechstes === undefined
                            ? { art: 'anlegen' }
                            : { art: 'zeile', wortID: naechstes }
                      }
                    }, bezug)
                  },
                }}
              />
            ) : (
              !wartet && (
                <p className="text-sm text-leise">{eigene.waehleEines}</p>
              )
            )}
          </div>
        </div>
      )}

      <Schluesselwortfeste woerter={feste} />
    </section>
  )
}
