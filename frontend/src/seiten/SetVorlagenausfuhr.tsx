// Die Ausfuhr eines Sets über eine Exportvorlage.
//
// **Eine eigene Datei**, weil die Setliste sonst eine dritte Frage beantwortet
// (AGENTS.md §2): *welche Sets gibt es*, *wie sieht ein Sicherungsknopf aus*
// und *welche Vorlagen gibt es dafür*.
//
// **Kein Dialog über einem Dialog** (§6): die Wahlmöglichkeiten stehen im
// Seitenbereich unter der Tabelle, nicht in einem Modal.
//
// **Sie lädt, sobald sie dasteht** — im Ausgabeassistenten (UXR-09,
// #1917)
// und seit UXR-04 (#1912)
// auch an der Setkarte, deren Menü sie öffnet. Bis dahin stand sie dort je
// Zeile als Knopf „Über eine Exportvorlage …", der erst lud; mit dem Menü
// davor wäre er ein Schritt ohne Frage. Im Assistenten führt *Ausführen* —
// es ist die eine Handlung der Fläche (R1); an der Karte steht daneben
// *Schließen*, das sie wieder aushängt.
import { useCallback, useEffect, useRef, useState } from 'react'

import { api, auspacken, type SetDTO } from '../api'
import type {
  ExportoptionDTO,
  ExportvorlageDTO,
  VorlagenausfuhrDTO,
} from '../vorlagen'
import { useLauf } from '../haken/useLauf'
import { useVorgang } from '../haken/useVorgang'
import { istAbgebrochen, meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { Ausgabeauswahl } from '../bausteine/Ausgabeauswahl'
import { Ausgabefuss } from '../bausteine/Ausgabefuss'
import { Knopf } from '../bausteine/Knopf'
import { Vorlagenprobe } from './Vorlagenprobe'
import { Vorlagenergebnis } from './Vorlagenergebnis'
import { Vorlagenoptionen } from './Vorlagenoptionen'
import { Vorlagenwahl } from './Vorlagenwahl'

// Eigenschaften sagen, wo sie steht: als Ziel des Ausgabeassistenten — ohne
// „Schließen", mit *Ausführen* als Hauptknopf — oder an einer Setkarte, die
// sie über `schliessen` wieder aushängt.
type Eigenschaften = { satz: SetDTO } & (
  { imAssistenten: true } | { imAssistenten?: false; schliessen: () => void }
)

// **Der offene Zustand liegt hier und nicht in der Setliste.** Welche Zeile
// gerade eine Vorlage wählt, geht die Liste nichts an; sie beantwortet die
// Frage „welche Sets gibt es" (§2).
export function SetVorlagenausfuhr(eigenschaften: Eigenschaften) {
  const { satz } = eigenschaften
  const imAssistenten = eigenschaften.imAssistenten === true
  const texte = useTexte()
  const lauf = useLauf()
  const vorgang = useVorgang()

  const [vorlagen, setzeVorlagen] = useState<readonly ExportvorlageDTO[]>()
  // **Die gewählte Vorlage selbst, nicht ihr Schlüssel.**
  //
  // Der erste Anlauf merkte sich den Paketnamen und suchte die Vorlage bei
  // jedem Zugriff neu. Das kostete drei Zweige — `vorlage?.optionen`, `if
  // (vorlage)`, `gewaehlt ?? ''` —, und **keinen** davon kann ein Test
  // betreten: die Auswahl steht nur dort, wo eine Vorlage da ist. Ein Zweig
  // ohne Fall ist eine Behauptung darüber, was im Ernstfall geschieht (§3).
  const [vorlage, setzeVorlage] = useState<ExportvorlageDTO>()
  const [wahl, setzeWahl] = useState<Record<string, string>>({})
  const [ausfuhr, setzeAusfuhr] = useState<VorlagenausfuhrDTO>()
  const [fehler, setzeFehler] = useState<string>()
  const [befunde, setzeBefunde] = useState<readonly string[]>([])
  const [gescheitert, setzeGescheitert] = useState<readonly string[]>([])

  // **Welche Karten hinausgehen** (#1167) — leer heißt: das ganze Set.
  const [auswahl, setzeAuswahl] = useState<readonly string[]>([])

  // **Geladen wird beim Öffnen, nicht beim Bau der Seite.** Eine Setliste mit
  // dreißig Zeilen führte sonst dreißig Paketläufe über die Platte, von denen
  // niemand einen sieht — deshalb hängt die Setkarte diese Fläche erst ein,
  // wenn ihr Menü danach gefragt hat.
  const laden = useCallback(() => {
    void vorgang.starte(`vorlagen-${satz.id}`, async () => {
      setzeFehler(undefined)

      try {
        const antwort = await api().Exportvorlagen(satz.id)
        auspacken(antwort, () => undefined)

        setzeVorlagen(antwort.vorlagen)
        setzeBefunde(antwort.befunde)
        setzeVorlage(antwort.vorlagen[0])
        setzeWahl(vorgaben(antwort.vorlagen[0]))
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }, [satz.id, vorgang, texte])

  // **Einmal beim Aufschlagen** — und nicht nach jedem Sprachwechsel, der
  // `laden` neu baut: der lüde die Vorlagen neu und nähme die getroffene Wahl
  // weg. Die Referenz übersteht auch das Verbergen im Assistenten.
  const angestossen = useRef(false)
  useEffect(() => {
    if (!angestossen.current) {
      angestossen.current = true
      laden()
    }
  }, [laden])

  const auswaehlen = useCallback(
    (gewaehlt: ExportvorlageDTO) => {
      lauf.naechster()
      setzeVorlage(gewaehlt)
      setzeWahl(vorgaben(gewaehlt))
      setzeAusfuhr(undefined)
      setzeFehler(undefined)
      setzeGescheitert([])
    },
    [lauf],
  )

  // **Die Vorlage kommt als Argument, nicht aus einer Suche.** Der erste
  // Anlauf suchte sie hier über `gewaehlt` und brauchte dafür ein `if
  // (!vorlage) return` — einen Zweig, den kein Test betreten kann, weil der
  // Knopf nur dort steht, wo eine Vorlage gewählt ist. Ein Zweig ohne Fall ist
  // eine Behauptung darüber, was im Ernstfall geschieht (§3).
  const ausfuehren = useCallback(
    (vorlage: ExportvorlageDTO) => {
      void vorgang.starte(`ausfuehren-${satz.id}`, async () => {
        const meiner = lauf.naechster()

        setzeFehler(undefined)
        setzeAusfuhr(undefined)
        setzeGescheitert([])

        try {
          const antwort = await api().VorlageAusfuehren(
            satz.id,
            vorlage.paket,
            vorlage.dateityp,
            wahl,
            [...auswahl],
          )

          if (!lauf.istAktuell(meiner)) {
            return
          }

          // **Die Befunde vor dem Auspacken.** `auspacken` wirft bei einem
          // Fehler, und genau dann tragen sie den Grund.
          setzeGescheitert(antwort.befunde)

          auspacken(antwort, () => undefined)

          // **Ein abgebrochener Dialog ist keine Meldung wert** — wer selbst
          // abbricht, weiß, dass nichts geschrieben wurde.
          if (!antwort.abgebrochen) {
            setzeAusfuhr(antwort)
          }
        } catch (ursache: unknown) {
          // **Ein Anhalten ist keine Fehlermeldung** (FMT-03) — geschriebene
          // Bilder bleiben liegen, wie bei MSE2, und der Laufstand sagt es.
          if (lauf.istAktuell(meiner) && !istAbgebrochen(ursache)) {
            setzeFehler(meldungAus(ursache, texte))
          }
        }
      })
    },
    // **`auswahl` gehört dazu**
    // (#1170).
    //
    // Ohne sie hielte `useCallback` die **erste** — die leere. Leer heißt nach
    // Vertrag „das ganze Set": wer drei von dreihundert Karten wählt und
    // ausführt, bekäme alle dreihundert.
    [auswahl, lauf, satz.id, texte, vorgang, wahl],
  )

  // optionSetzen nimmt eine Wahl an einer Option an — und behält sie.
  //
  // **Die Vorlage kommt als Argument**, wie bei `ausfuehren`: ihr Paketname
  // gehört zu der Wahl, die eben getroffen wurde, nicht zu einer, die bis
  // zum Behalten vielleicht schon eine andere ist.
  const optionSetzen = (
    paket: string,
    option: ExportoptionDTO,
    wert: string,
  ) => {
    const meiner = lauf.naechster()

    setzeWahl((bisher) => ({ ...bisher, [option.name]: wert }))

    // **Die Wahl wird behalten — je Vorlage** (#492). MSE hält
    // sie ebenso (`exportOptionsFor`, `src/data/settings.hpp`,
    // Zeilen 180–186, Fassung f9d9356 vom 2021-07-06).
    //
    setzeAusfuhr(undefined)
    setzeFehler(undefined)
    setzeGescheitert([])

    // **Und der Lauf hängt nicht daran.** Ein gescheitertes
    // Behalten ist kein Grund, die Ausgabe anzuhalten: der Wert
    // steht im Zustand der Seite und geht mit. **Gemeldet wird er
    // trotzdem** — die Oberfläche behauptet sonst, die Option
    // werde behalten, und beim nächsten Öffnen steht der alte
    // Wert da.
    //
    // **Mit `auspacken` und nicht mit `.catch`.** Der Zweitblick
    // über Phase G hat es gefunden: die Bindung löst sich
    // **erfolgreich** auf und trägt den Fehler *im* Umschlag; ein
    // `.catch()` daran fängt nie etwas. Und die drei Zeilen
    // darüber räumten die Meldung gleich wieder weg — sie stehen
    // jetzt davor.
    void (async () => {
      try {
        auspacken(
          await api().ExportoptionSetzen(paket, option.name, wert),
          () => undefined,
        )
      } catch (ursache: unknown) {
        // **Und nur, wenn sie noch zum heutigen Stand gehört**
        // (§7, Regel 2). Der Nachlauf über Phase G hat gefunden,
        // dass die Nummer zwar hochgezählt, aber nicht behalten
        // wurde: eine überholte Meldung trug sich nach einer
        // neueren, geglückten Wahl ein — und behauptete, das
        // Behalten sei gescheitert, während der zuletzt gewählte
        // Wert längst im Bestand steht.
        if (lauf.istAktuell(meiner)) {
          setzeFehler(meldungAus(ursache, texte))
        }
      }
    })()
  }

  // **Ein Fehler beim Lesen steht beim Knopf, der es noch einmal versucht.**
  // Der erste Anlauf merkte ihn und zeigte weiter nur den Knopf — wer ihn
  // drückte, sah nichts geschehen. Eine Meldung, die niemand sieht, ist keine
  // (§6). **Der Knopf steht nur nach einem Fehler**; solange geladen wird,
  // sagt es ein Satz.
  if (vorlagen === undefined) {
    return fehler === undefined ? (
      <p role="status" className="text-leise">
        {texte.bausteine.zustaende.laedt}
      </p>
    ) : (
      <>
        <button
          type="button"
          disabled={vorgang.laeuft(`vorlagen-${satz.id}`)}
          aria-label={`${texte.seiten.sets.vorlageAusfuhr} — ${satz.name}`}
          onClick={laden}
          className={knopf}
        >
          {texte.seiten.sets.vorlageAusfuhr}
        </button>

        <p role="alert" className="mt-2 text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      </>
    )
  }

  return (
    <div
      className={
        imAssistenten ? undefined : 'mt-3 rounded-sm border border-kante p-3'
      }
    >
      {/* **Ohne gewählte Vorlage gibt es keine** — gewählt wird beim Laden
          die erste. So kennt der Zweig darunter seine Vorlage, statt sie an
          jeder Stelle noch einmal zu prüfen. */}
      {vorlage === undefined ? (
        <p className="max-w-prose text-leise">
          {texte.seiten.sets.vorlageKeine}
        </p>
      ) : (
        <>
          <Vorlagenwahl
            vorlagen={vorlagen}
            gewaehlt={vorlage}
            waehlen={auswaehlen}
          />

          <Vorlagenoptionen
            key={vorlage.paket}
            optionen={vorlage.optionen}
            wahl={wahl}
            setzen={(option, wert) => {
              optionSetzen(vorlage.paket, option, wert)
            }}
          />

          {/* **Welche Karten hinausgehen** (#1167). MSE2 fragt es vor der
              HTML-Ausgabe genauso (`SetWindow::onFileExportHTML`,
              `src/gui/set/window.cpp`, Zeile 677, Fassung f9d9356 vom
              2021-07-06). */}
          <div className="mt-4">
            <Ausgabeauswahl
              setID={satz.id}
              gewaehlt={auswahl}
              setzeGewaehlt={setzeAuswahl}
              angehalten={vorgang.laeuft(`ausfuehren-${satz.id}`)}
            />
          </div>

          {/* **„Wie es wird"** (UXR-F7, #2015): dieselbe Vorlage, dieselbe Wahl,
              dieselben Karten wie beim Knopf — ausgewertet, nicht geschrieben. */}
          <Vorlagenprobe
            setID={satz.id}
            paket={vorlage.paket}
            wahl={wahl}
            auswahl={auswahl}
          />

          {/* **In der Fußleiste der Ausgabe** (AP-10, B-20) — fest nur im
              Assistenten. */}
          <Ausgabefuss fest={imAssistenten}>
            <Knopf
              gewicht={imAssistenten ? 'haupt' : 'neben'}
              klein={!imAssistenten}
              angehalten={vorgang.laeuft(`ausfuehren-${satz.id}`)}
              aufKlick={() => {
                ausfuehren(vorlage)
              }}
            >
              {texte.seiten.sets.vorlageAusfuehren}
            </Knopf>

            {/* **Im Assistenten gibt es nichts zu schließen** — wer etwas
                anderes will, wählt unter „Wohin?" ein anderes Ziel. An der
                Setkarte hängt *Schließen* die Fläche aus, und mit ihr geht
                ihr Zustand; ein Lauf, der noch unterwegs ist, trägt sich in
                eine Fläche, die es nicht mehr gibt, nicht mehr ein. */}
            {eigenschaften.imAssistenten !== true && (
              <button
                type="button"
                onClick={eigenschaften.schliessen}
                className={knopf}
              >
                {texte.seiten.sets.vorlageSchliessen}
              </button>
            )}
          </Ausgabefuss>
        </>
      )}

      <Vorlagenergebnis
        fehler={fehler}
        gescheitert={gescheitert}
        befunde={befunde}
        ausfuhr={ausfuhr}
      />
    </div>
  )
}

const knopf =
  'whitespace-nowrap rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60'

// vorgaben nimmt die Vorgabewerte einer Vorlage.
function vorgaben(
  vorlage: ExportvorlageDTO | undefined,
): Record<string, string> {
  const gesetzt: Record<string, string> = {}

  for (const option of vorlage?.optionen ?? []) {
    // **Der behaltene Wert, nicht die Vorgabe** (#492). Ohne behaltenen steht
    // dort die Vorgabe — das entscheidet `internal/app`, damit die Antwort
    // nicht an zwei Stellen entsteht.
    gesetzt[option.name] = option.wert
  }

  return gesetzt
}
