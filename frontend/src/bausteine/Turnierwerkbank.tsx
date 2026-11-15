// **Das aufgeschlagene Turnier** — Werkstück 5 von Phase 4.7
// (#1294).
//
// **Sie hält die drei Teile zusammen** und baut keinen davon: Teilnehmer,
// Verlauf und Rangliste stehen je in einem eigenen Baustein (§2), und diese
// Datei sagt nur, in welcher Reihenfolge sie kommen.
import { useEffect, useRef } from 'react'

import { Entscheidung } from './Entscheidung'
import { Kontextmenue } from './Kontextmenue'
import { Laschen } from './Laschen'
import { Menue } from './Menue'
import { Turnierabschluss } from './Turnierabschluss'
import { Turnierdecklisten } from './Turnierdecklisten'
import { einsetzen } from '../sprache/einsetzen'
import { Turnierprotokoll } from './Turnierprotokoll'
import { Turnierrangliste } from './Turnierrangliste'
import { Turnierrunden } from './Turnierrunden'
import { Turnierteilnehmer } from './Turnierteilnehmer'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import type { TurnierDTO } from '../apiturniere'
import type { Turnierstand } from '../haken/useTurniere'
import { useTurnierhandlungen } from '../haken/useTurnierhandlungen'

export function Turnierwerkbank({
  stand,
  turnier,
  ansprung = false,
  angesprungen,
}: {
  stand: Turnierstand
  turnier: TurnierDTO

  // ansprung heißt: das Turnier wurde eben aus der Liste geöffnet, und der
  // Fokus soll zu seinem Namen — UXR-12 (#1920). Die Werkbank steht über
  // der Liste; ohne den Sprung stünde der Fokus auf einem Knopf darunter,
  // und das Fenster zeigte nicht, was aufging. **Einmal**: angesprungen
  // sagt der Seite, dass es geschehen ist — nach einem Schreiben baut sich
  // die Werkbank neu auf und soll nicht wieder springen.
  ansprung?: boolean
  angesprungen?: () => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.turniere

  const handlungen = useTurnierhandlungen(stand, turnier)
  const bezeichnung = einsetzen(eigene.handlungenAn, { name: turnier.name })

  const kopf = useRef<HTMLHeadingElement>(null)
  useEffect(() => {
    if (ansprung) {
      kopf.current?.focus()
      angesprungen?.()
    }
  }, [ansprung, angesprungen])

  return (
    <section
      aria-label={turnier.name}
      className="flex flex-col gap-4 rounded-md border border-kante p-4"
    >
      {/* **Der Kopf trägt das Menü des Turniers** — AP-6 (#2363), B-04: der
          Rechtsklick und ⇧F10 zeigen dieselbe Liste wie „…“ (`Kontext09`).
          Nur der Kopf und nicht die Laschen: deren Zeilen haben eigene. */}
      <Kontextmenue
        bezeichnung={bezeichnung}
        eintraege={handlungen.eintraege}
        objekt={`turnier/${turnier.id}`}
        klasse="flex flex-col gap-4"
      >
        <div>
          <h2
            ref={kopf}
            tabIndex={-1}
            className="text-lg font-semibold focus:outline-none"
          >
            {turnier.name}
          </h2>

          {/* **System, Runde und Zahl der Teilnehmenden** — wie im Artboard
              *Turniere* unter dem Namen (DOM-04.AK-02, #1699). „von" ist die
              empfohlene Rundenzahl der Turnierregeln, kein Ende. */}
          <p className="text-sm text-leise">{kopfzeile(turnier, eigene)}</p>
        </div>

        {/* **„Abschließen“ bleibt sichtbar, das übrige hinter „…“** — AP-6,
            B-04: Ausgeben, Bericht sichern und Löschen standen gleichrangig
            daneben. Abschließen ist, worauf ein Turnierabend zuläuft. */}
        <div className="flex flex-wrap items-start gap-2">
          <Turnierabschluss stand={stand} turnier={turnier} />

          <Menue bezeichnung={bezeichnung} eintraege={handlungen.eintraege}>
            {eigene.handlungenKurz}
          </Menue>

          {handlungen.meldung !== '' && (
            <p role="status" className="basis-full text-sm">
              {handlungen.meldung}
            </p>
          )}

          {handlungen.gefragt && (
            <div className="basis-full">
              <Entscheidung
                frage={eigene.loeschenFrage}
                bestaetigen={eigene.loeschenJa}
                aktion="loeschen"
                abbrechen={eigene.loeschenNein}
                laeuft={handlungen.laeuft}
                aufBestaetigen={() => {
                  void handlungen.loeschen()
                }}
                aufAbbrechen={handlungen.abbrechen}
              />
            </div>
          )}
        </div>
      </Kontextmenue>

      {/* **Gebiete statt Blöcke untereinander** — Werkstück I von Phase
          4.10
          (#1325).
          Wer ein Ergebnis einträgt, will die Runden sehen; wer den Stand
          sucht, die Rangliste. Beides zugleich zu zeigen heißt, beides zu
          suchen. **In der Ordnung des Artboards** (UXR-12, #1920): Runden
          und Rangliste vorn, nach ihnen fragt ein Abend; die Teilnehmenden,
          die Listen und das Protokoll danach. */}
      {/* **Wo das Turnier steht, dort setzt man an**: mit Runden bei den
          Runden, ohne bei den Teilnehmenden — dort fängt es an. Der `key` am
          Aufrufer hängt an der Rundenzahl (§7, Regel 1: ein Wechsel räumt
          den offenen Kontext), und ohne diese Zeile würfe jedes Legen einer
          Runde auf die erste Lasche zurück. */}
      <Laschen
        bezeichnung={eigene.gebiete}
        oeffne={turnier.runden.length === 0 ? 'teilnehmer' : 'runden'}
        gebiete={[
          {
            schluessel: 'runden',
            name: eigene.gebietRunden,
            inhalt: (
              <Turnierrunden
                stand={stand}
                turnier={turnier}
                nurLesen={turnier.abgeschlossen !== ''}
              />
            ),
          },
          {
            schluessel: 'rangliste',
            name: eigene.gebietRangliste,
            inhalt: (
              <div className="flex flex-col gap-3">
                <Turnierrangliste
                  stand={stand}
                  einzelspiel={turnier.einzelspiel}
                />

                {/* **Der Hinweis steht bei der Rangliste** und nicht unter
                    der ganzen Werkbank: er erklärt ihre vier Spalten. */}
                <Erklaerungsabsatz
                  text={eigene.quotenHinweis}
                  klasse="text-sm text-leise"
                />
              </div>
            ),
          },
          {
            schluessel: 'teilnehmer',
            name: eigene.gebietTeilnehmer,
            inhalt: (
              <Turnierteilnehmer
                stand={stand}
                turnier={turnier}
                nurLesen={turnier.abgeschlossen !== ''}
              />
            ),
          },
          {
            // **Die Listen in einer eigenen Lasche** — DOM-04.AK-02, wie im
            // Artboard.
            schluessel: 'decklisten',
            name: eigene.gebietDecklisten,
            inhalt: <Turnierdecklisten turnier={turnier} />,
          },
          {
            schluessel: 'protokoll',
            name: eigene.gebietProtokoll,
            inhalt: <Turnierprotokoll turnier={turnier} />,
          },
        ]}
      />
    </section>
  )
}

// kopfzeile sagt, wo das Turnier steht — **ohne „von", wo Appendix E der MTR
// (Fassung 2026-02-27) nichts empfiehlt**: „Runde 1 von 0" fand die Hüllenprobe bei zwei Teilnehmenden.
function kopfzeile(
  turnier: TurnierDTO,
  eigene: ReturnType<typeof useTexte>['seiten']['turniere'],
): string {
  const anzahl = String(turnier.teilnehmer.length)
  const runde = String(turnier.runden.length)

  if (turnier.runden.length === 0) {
    return einsetzen(eigene.kopfzeileOhneRunde, { anzahl })
  }

  if (turnier.empfohleneRunden === 0) {
    return einsetzen(eigene.kopfzeileOhneEmpfehlung, { runde, anzahl })
  }

  return einsetzen(eigene.kopfzeile, {
    runde,
    von: String(turnier.empfohleneRunden),
    anzahl,
  })
}
