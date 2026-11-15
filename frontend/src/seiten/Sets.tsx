// Die Setliste.
//
// **Was sie tut, steht hier; wie sie aussieht, in den Klassen** (AGENTS.md §2).
// Das Laden geht über `useLaden`, damit die Antwort eines überholten Laufs
// nicht in den neuen Stand gerät (§7, Regel 2).
//
// # Nach dem Artboard *Sets* — UXR-04
//
// (#1912)
// **Die Liste links, das Neue rechts daneben**, ab `lg`; im engen Fenster
// untereinander, die Liste zuerst. Im Kopf führt *Neues Set*, *Set öffnen …*
// steht daneben. Gemessen war vorher: drei Hauptknöpfe, *Set öffnen …* im
// Hauptgewicht, und Anlage und Einfuhr **unter** der Liste — mit dreißig Sets
// unter der Falz, wo nur ⌘N sie fand.
import { useCallback, useEffect, useState } from 'react'

import { api, auspacken, type SetDTO } from '../api'
import { useAnschluss, useAufgenommen } from '../anschlussraum'
import { Auswahlfilter } from '../bausteine/Auswahlfilter'
import { useLaden } from '../haken/useLaden'
import { type Speicherlauf } from '../haken/useSetSpeichern'
import { useSetloeschung } from '../haken/useSetloeschung'
import { useUmschlag } from '../haken/useUmschlag'
import { useSetoeffnung } from '../haken/useSetoeffnung'
import { useSetvorschau } from '../haken/useSetvorschau'
import { Setanlage, useSetanlage } from './Setanlage'
import { Verdeckung } from '../verdeckung'
import { Setvorschau } from './Setvorschau'

import type { Seitenschluessel } from '../navigation'
import { einsetzen } from '../sprache/einsetzen'
import { useSprache, useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { Laufstand } from '../bausteine/Laufstand'
import { Seitenkopf } from '../bausteine/Seitenkopf'
import { Ladefeld } from './Ladefeld'
import { Seteinfuhr } from './Seteinfuhr'
import { Setkarte } from './Setkarte'
import { Listenansicht } from '../bausteine/Listenansicht'
import {
  setansichtsschluessel,
  setformen,
  setsortierungen,
} from '../bausteine/listenordnung'
import { useAnsicht } from '../haken/useAnsicht'
import { Dateiverweis } from '../bausteine/Dateiverweis'

interface SetsEigenschaften {
  waehlen: (setID: string) => void

  // bestandsstand zählt fremde Schreibvorgänge — die Liste lädt dann nach
  // (Befund 3 des Gesamtreviews über 0.6.0).
  bestandsstand?: number

  // speichern kommt von **außen**, und das ist keine Bequemlichkeit.
  //
  // ⌘S gilt im ganzen Fenster und steht in `App`; die Knöpfe stehen hier. Zwei
  // Haken hießen zwei Vorgangslisten — und §7 Regel 4 („ein Vorgang je
  // Datensatz") gälte je Aufrufer statt je Set. Wer während eines Speicherns
  // über die Liste zusätzlich ⌘S drückt, stieße einen zweiten Schreibvorgang
  // auf dieselbe Datei an. Gefunden vom Zweitblick durch Codex über Phase 6.
  speichern: Speicherlauf

  // anstoss sagt, dass jemand von außen den Dateidialog will (#891).
  //
  // # Warum ein Anstoß und kein Aufruf
  //
  // ⌘O und der Ablageeintrag „Öffnen …" stehen eine Ebene höher, der Dialog
  // steht hier. Ein zweiter Aufruf dort wäre eine zweite Fassung derselben
  // Frage — und die, die niemand nachführt.
  //
  // # Warum er **verbraucht** wird
  //
  // Diese Seite entsteht neu, wenn jemand auf sie wechselt: ein Anstoß von
  // außen kommt an, **bevor** es die Seite gibt. Ein Vergleich mit dem
  // vorigen Wert liefe deshalb ins Leere — beim ersten Zeichnen ist der
  // vorige Wert der jetzige. Der Anstoß wird stattdessen gelesen und
  // zurückgesetzt.
  //
  // **Wahlfrei**, weil die Seite auch ohne Fenster steht: in der
  // Browservorschau gibt es keine Menüleiste.
  anstoss?: boolean
  anstossVerbraucht?: () => void

  // neuanstoss setzt den Fokus ins Namensfeld des neuen Sets — ⌘N (#891).
  // *Neues Set* im Kopf tut dasselbe (UXR-04).
  neuanstoss?: boolean
  neuAngestossen?: () => void

  // geloescht sagt der Ebene darüber, welches Set weg ist (#1218).
  //
  // **Weil sie es wissen muss und diese hier nicht.** War es das offene, gibt
  // es kein offenes mehr (§7, Regel 1) — welches offen ist, weiß `Anwendung`.
  //
  // **Wahlfrei**, weil die Seite auch ohne Fenster steht: in der
  // Browservorschau ist kein Set aufgeschlagen.
  geloescht?: (setID: string) => void

  // wechseln geht auf eine andere Seite — hier zu den Paketen, wenn einem
  // neuen Set Spiel oder Stil fehlt (UX-05 S07.NAT-04).
  wechseln: (seite: Seitenschluessel) => void
}

export function Sets({
  waehlen,
  speichern,
  anstoss = false,
  anstossVerbraucht,
  neuanstoss = false,
  neuAngestossen,
  geloescht,
  wechseln,
  bestandsstand = 0,
}: SetsEigenschaften) {
  const texte = useTexte()
  const sprache = useSprache()
  const holen = useCallback(async () => {
    const antwort = await api().Sets()

    return auspacken(antwort, () => antwort.sets)
  }, [])
  // **Der Stand liest nach und fragt nicht neu** (U11, #2125) — dieselbe
  // Liste, ein neuer Stand.
  const { daten, laedt, fehler, nachladen } = useLaden<SetDTO[]>(
    holen,
    [],
    false,
    bestandsstand,
  )
  const ansicht = useAnsicht(setansichtsschluessel, setformen)

  // **Das Löschen** (#1218) — ein Haken für die ganze Liste, nicht einer je
  // Zeile (§7, Regel 4).
  const loeschung = useSetloeschung(
    texte,
    nachladen,
    useCallback(
      (setID: string) => {
        geloescht?.(setID)
      },
      [geloescht],
    ),
  )

  // **Der portable Umschlag** (#1457) — ein Haken für beide Richtungen und
  // für die ganze Liste, wie `loeschung` darüber.
  const umschlag = useUmschlag()

  // **Das Öffnen aus jeder Quelle** — ein Haken (§2), seit FMT-03.
  const { oeffnen, laeuft, ausgang, oeffnungsfehler, teilerfolg } =
    useSetoeffnung(nachladen)

  // **Erst nachsehen, dann einlesen** — FMT-03 (#1692), `useSetvorschau`.
  const vorschau = useSetvorschau((pfad) => {
    void oeffnen(() => api().SetAusDatei(pfad))
  })

  // **Neues Set im Kopf geht denselben Weg wie ⌘N** — zum Namen, mit dem
  // Fokus darin. Ein eigener Anstoß neben dem von außen; verbraucht werden
  // beide an einer Stelle.
  const [neuGedrueckt, setzeNeuGedrueckt] = useState(false)

  // **Die Sets mit Problemen aus dem Überblick** — AP-8, Teil 2 (#2365): die
  // Liste zeigt genau sie; der Filter nennt sich und lässt sich lösen.
  const anschluss = useAnschluss()
  const [auswahl, setzeAuswahl] = useState<{
    name: string
    setIDs: readonly string[]
  }>()
  useAufgenommen(
    anschluss.offen?.art === 'setauswahl' ? anschluss.offen : undefined,
    ({ name, setIDs }) => {
      setzeAuswahl({ name, setIDs })
    },
  )
  const anlage = useSetanlage(
    daten !== undefined && daten.length > 0,
    neuanstoss || neuGedrueckt,
  )
  const neuVerbraucht = useCallback(() => {
    setzeNeuGedrueckt(false)
    neuAngestossen?.()
  }, [neuAngestossen])

  // **Der Anstoß von außen öffnet denselben Dialog wie der Knopf.**
  //
  // Nicht eine zweite Fassung derselben Frage: ⌘O und „Set öffnen …" landen in
  // derselben Funktion, und was der eine kann, kann der andere (#891).
  useEffect(() => {
    if (!anstoss) {
      return
    }

    anstossVerbraucht?.()

    void vorschau.waehlen()
    // **Nur am Anstoß.** `vorschau` entsteht bei jedem Zeichnen neu; hinge
    // der Effekt daran, öffnete er den Dialog immer wieder.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [anstoss, anstossVerbraucht])

  return (
    <Verdeckung value={anlage.offen}>
      <div className="contents" inert={anlage.offen}>
        <Seitenkopf
          titel={texte.seiten.sets.titel}
          erklaerung={texte.seiten.sets.erklaerung}
        >
          {/* **Die Handlungen stehen rechts oben** — dieselbe Stelle wie auf
            jeder anderen Seite (`Seitenkopf`). Beide im Nebengewicht, seit
            AP-3 (#2360): »Neues Set« springt in das Anlegen, das daneben
            offen steht, und dessen »Weiter« führt (`Wege09`) — je Blick
            einer (`Zustaende09`). */}
          <Knopf
            angehalten={laeuft || vorschau.laeuft}
            aufKlick={() => {
              void vorschau.waehlen()
            }}
          >
            {texte.seiten.sets.setOeffnen}
          </Knopf>

          <Aktionsknopf
            aktion="anlegen"
            mitKuerzel
            schreibt
            aufKlick={() => {
              setzeNeuGedrueckt(true)
            }}
          >
            {texte.seiten.sets.neuKopf}
          </Aktionsknopf>
        </Seitenkopf>

        <Setvorschau lauf={vorschau} angehalten={laeuft} />

        <Laufstand art="set" />
        <Laufstand art="mtgjson" />

        {teilerfolg !== undefined && (
          <p role="status" className="mt-3 max-w-prose">
            {einsetzen(texte.seiten.sets.teilerfolg, {
              anzahl: String(teilerfolg),
            })}
          </p>
        )}

        {oeffnungsfehler !== undefined && (
          <p role="alert" className="mt-3 text-fehler">
            {texte.bausteine.zustaende.fehler} {oeffnungsfehler}
          </p>
        )}

        {ausgang && (
          <div role="status" className="mt-3">
            <p className="text-leise">
              {ausgang.schonOffen
                ? texte.seiten.sets.schonOffen
                : `${zaehlwort(ausgang.karten, texte.allgemein.zaehlwoerter.karten)} ${texte.seiten.sets.geoeffnet}`}
            </p>

            {/* **Die Vorbehalte stehen beim Erfolg** (§8): das Set ist da, und
              diese Stellen sind es auch — nur ungedeutet. */}
            {ausgang.befunde.length > 0 && (
              <>
                <p className="mt-2 max-w-prose text-leise">
                  {texte.seiten.sets.beimOeffnenBefunde}
                </p>
                <ul className="mt-1 space-y-1 font-mono text-sm">
                  {ausgang.befunde.map((vorbehalt) => (
                    <li key={vorbehalt}>{vorbehalt}</li>
                  ))}
                </ul>
              </>
            )}
          </div>
        )}

        {speichern.fehler !== undefined && (
          <p role="alert" className="mt-3 text-fehler">
            {texte.bausteine.zustaende.fehler} {speichern.fehler}
          </p>
        )}

        {speichern.speicherung && (
          <div role="status" className="mt-3">
            <p className="text-leise">
              {`${texte.seiten.sets.gespeichert} ${speichern.speicherung.pfad}`}
            </p>
            <Dateiverweis pfad={speichern.speicherung.pfad} />

            {/* **Die Befunde stehen beim Erfolg** (§8): die Setdatei ist
              geschrieben, und was daneben fehlte, steht dabei. */}
            {speichern.speicherung.befunde.length > 0 && (
              <>
                <p className="mt-2 max-w-prose text-leise">
                  {texte.seiten.sets.beimSpeichernBefunde}
                </p>
                <ul className="mt-1 space-y-1 font-mono text-sm">
                  {speichern.speicherung.befunde.map((befund) => (
                    <li key={befund}>{befund}</li>
                  ))}
                </ul>
              </>
            )}
          </div>
        )}

        <div className="mt-4 grid items-start gap-6 lg:grid-cols-[minmax(0,1fr)_21rem]">
          <div className="min-w-0">
            <Ladefeld
              laedt={laedt}
              fehler={fehler}
              leer={daten?.length === 0}
              vorhanden={daten !== undefined}
            >
              {/* **Der leere Zustand erklärt den nächsten Schritt** (§6), statt
                „Keine Daten" zu zeigen. */}
              <p className="mt-3 max-w-prose text-leise">
                {texte.seiten.sets.leer}
              </p>
            </Ladefeld>

            {/* **Kacheln oder Liste, mit Suche und Sortierung** — UX-11
              (#1683). Leer heißt hier nichts: das sagt `Ladefeld` darüber. */}
            {auswahl !== undefined && (
              <Auswahlfilter
                name={auswahl.name}
                loesen={texte.seiten.sets.auswahlLoesen}
                aufLoesen={() => {
                  setzeAuswahl(undefined)
                }}
              />
            )}

            {daten && daten.length > 0 && (
              <Listenansicht
                bezeichnung={texte.seiten.sets.titel}
                eintraege={
                  auswahl === undefined
                    ? daten
                    : daten.filter((satz) => auswahl.setIDs.includes(satz.id))
                }
                schluessel={(satz) => satz.id}
                suchtext={(satz) => `${satz.name} ${satz.spiel} ${satz.stil}`}
                sortierungen={setsortierungen(texte, sprache)}
                form={ansicht.form}
                setzeForm={ansicht.setzeForm}
                leer={null}
                marken={{ art: 'set', name: (satz) => satz.name }}
                eintrag={(satz, gestalt) => (
                  <Setkarte
                    gestalt={gestalt}
                    satz={satz}
                    waehlen={waehlen}
                    speichern={speichern}
                    loeschung={loeschung}
                    umschlag={umschlag}
                  />
                )}
              />
            )}
          </div>

          {/* **Das Neue steht neben der Liste** (UXR-04) — ohne Set offen,
            sonst in einer Schublade (AP-10, `Setanlage`). */}
          <div className="flex min-w-0 flex-col gap-4">
            <Setanlage
              stand={anlage}
              gestalt={
                daten === undefined
                  ? 'keine'
                  : daten.length === 0
                    ? 'offen'
                    : 'schublade'
              }
              angelegt={waehlen}
              anstoss={neuanstoss || neuGedrueckt}
              anstossVerbraucht={neuVerbraucht}
              zuDenPaketen={() => {
                wechseln('pakete')
              }}
            />

            <Seteinfuhr
              laeuft={laeuft}
              oeffnen={(lesen) => {
                void oeffnen(lesen)
              }}
              umschlag={umschlag}
            />
          </div>
        </div>
      </div>
    </Verdeckung>
  )
}
