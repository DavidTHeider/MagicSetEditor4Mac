// Die **Vorlagenwerkstatt**
// (#1164;
// als Arbeitsplatz seit Werkstück U7 von Phase 5.1,
// #1418).
//
// # Wofür
//
// Was eine Karte kann und wie sie aussieht, steht in Textdateien im
// Vorlagenpaket. Bis Phase 3.3 waren sie nur mit einem Texteditor daneben zu
// ändern — die Anwendung konnte sie lesen und ausführen, nicht bearbeiten.
//
// # Drei Regionen
//
// **Navigator**: der Baum aus Paketen und Dateien mit einem Filter darüber.
// **Zentrum**: der Editor, Speichern, der Diff vor dem Speichern — und die
// Rückfrage, wenn ein Wechsel Ungespeichertes verwürfe. **Inspector**: die
// Testkarte, die Vorschau mit der Marke „veraltet", die Diagnose mit dem
// Sprung zur Stelle und die Herkunft des Namens unter dem Zeiger.
//
// Bis U7 waren es ein Auswahlfeld, eine Dateiliste und eine Textfläche
// untereinander — und der Kernweg `paket` kostete sechs Handlungen
// (UX-ARCHITEKTUR). Mit dem Baum sind es fünf: ⌘5, Datei, in den Text,
// tippen, Speichern.
//
// # Diese Datei zeigt, `useVorlagenwerkstatt` entscheidet
//
// Was die Werkstatt tut, steht im Haken; hier steht, wie sie aussieht (§2).
import { Ladefehler } from '../bausteine/Ladehinweise'
import { Arbeitsplatz } from '../bausteine/Arbeitsplatz'
import { Knopf } from '../bausteine/Knopf'
import { Seitenkopf } from '../bausteine/Seitenkopf'
import { Beiwerktafel } from './Beiwerktafel'
import { Dateikopf } from './Dateikopf'
import { Vorlagenbaum } from './Vorlagenbaum'
import { Vorlagenbefund } from './Vorlagenbefund'
import { Vorlagendatei } from './Vorlagendatei'
import { Vorlageninspector } from './Vorlageninspector'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import type { Seitenschluessel } from '../navigation'
import { useArbeitsplatzgedaechtnis } from '../haken/useArbeitsplatzgedaechtnis'
import { useDateiaktionen } from '../haken/useDateiaktionen'
import { useDateikontext } from './dateikontext'
import { Kontextmenue } from '../bausteine/Kontextmenue'
import { usePaketkopie } from '../haken/usePaketkopie'
import { useVorlagenwerkstatt } from '../haken/useVorlagenwerkstatt'

// bestandsstand zählt fremde Schreibvorgänge (#156) — Befund 3 des Nachlaufs
// zum Gesamtreview über 0.6.0.
export function Vorlagen({
  bestandsstand = 0,
  wechseln,
}: {
  bestandsstand?: number

  // wechseln führt aus dem leeren Baum zu den Paketen (G02, #2249).
  wechseln?: (seite: Seitenschluessel) => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.vorlagen
  const werkstatt = useVorlagenwerkstatt(bestandsstand)
  const kopie = usePaketkopie(werkstatt)
  const aktionen = useDateiaktionen(werkstatt)
  const gedaechtnis = useArbeitsplatzgedaechtnis()

  const { wahl, puffer, rueckfrage } = werkstatt
  const beiwerk = wahl?.beiwerk === true ? wahl : null
  const offen =
    wahl !== null &&
    beiwerk === null &&
    !puffer.laedt &&
    puffer.lesefehler === undefined

  const datei = useDateikontext(werkstatt, aktionen, offen)

  return (
    <section className="flex flex-col gap-4">
      <Seitenkopf titel={eigene.titel} erklaerung={eigene.erklaerung} />

      {werkstatt.paketeFehler !== undefined && (
        <Ladefehler
          fehler={werkstatt.paketeFehler}
          wiederholen={werkstatt.paketeNeuLaden}
        />
      )}

      <Arbeitsplatz
        bezeichnung={eigene.arbeitsplatz}
        gedaechtnis={gedaechtnis}
        // **Der Inspector weicht, nicht der Baum** — der Baum ist der einzige
        // Weg zu einer Datei; bei 1.280 × 800 stand die Werkstatt sonst ohne
        // ihn da (gefunden im Messlauf `arbeitswegmass.mjs`).
        beiEnge="inspector"
        navigator={{
          bezeichnung: eigene.baumBezeichnung,
          inhalt: (
            <Kontextmenue
              bezeichnung={eigene.datei}
              eintraege={datei.eintraege}
              objekt={datei.objekt}
              anstoss={datei.anstoss}
              anstossVerbraucht={datei.verbraucht}
              nurAngestossen
            >
              <Vorlagenbaum
                dateianstoss={datei.anstossen}
                {...(wechseln === undefined
                  ? {}
                  : {
                      zuDenPaketen: () => {
                        wechseln('pakete')
                      },
                    })}
                pakete={werkstatt.pakete}
                geladen={werkstatt.paketeGeladen}
                fehlschlag={werkstatt.paketeFehler !== undefined}
                gewaehlt={wahl}
                waehlen={werkstatt.waehlen}
                geaendert={puffer.geaendert}
                kopie={kopie}
                fassung={werkstatt.fassung}
              />
            </Kontextmenue>
          ),
        }}
        zentrum={
          <div className="flex flex-col gap-3">
            {/* **Die Rückfrage steht im Zentrum, nicht in einem Dialog** (§6) —
                dort, wo der Text steht, um den es geht. */}
            {rueckfrage !== null && (
              <div
                role="group"
                aria-label={eigene.verwerfen}
                className="flex flex-col gap-2 rounded-sm border border-kante p-3"
              >
                <p className="text-sm">
                  {einsetzen(eigene.verwerfenFrage, {
                    datei: rueckfrage.datei,
                  })}
                </p>
                <div className="flex gap-2">
                  <Knopf
                    gewicht="haupt"
                    aufKlick={() => {
                      werkstatt.verwerfen(rueckfrage)
                    }}
                  >
                    {eigene.verwerfen}
                  </Knopf>
                  <Knopf gewicht="neben" aufKlick={werkstatt.bleiben}>
                    {eigene.hierBleiben}
                  </Knopf>
                </div>
              </div>
            )}

            {/* **Der leere Zustand nennt den nächsten Schritt** (§6). */}
            {wahl === null && (
              <p className="max-w-prose text-sm text-leise">
                {eigene.waehleImBaum}
              </p>
            )}

            {wahl !== null && puffer.lesefehler !== undefined && (
              <p role="alert" className="text-sm text-fehler">
                {puffer.lesefehler}
              </p>
            )}

            {beiwerk !== null && (
              <Beiwerktafel datei={beiwerk.datei} aktionen={aktionen} />
            )}

            {offen && (
              <Dateikopf
                aktionen={aktionen}
                kopie={kopie}
                paket={wahl.paket}
                datei={wahl.datei}
                hauptdatei={wahl.datei === werkstatt.hauptdatei}
                gesperrt={puffer.geaendert}
              />
            )}

            {offen && (
              <Vorlagendatei
                puffer={puffer}
                stelle={werkstatt.pruefung.befund?.stelle ?? null}
                namen={werkstatt.namen}
                sprungziel={werkstatt.sprungziel}
                aufZeiger={werkstatt.setzeZeiger}
                befund={
                  <Vorlagenbefund
                    befund={werkstatt.pruefung.befund}
                    zurStelle={werkstatt.zurStelle}
                  />
                }
              />
            )}
          </div>
        }
        inspector={{
          bezeichnung: eigene.inspectorBezeichnung,
          inhalt: offen ? (
            <Vorlageninspector
              paket={wahl.paket}
              datei={wahl.datei}
              gueltig={werkstatt.pruefung.gueltig}
              veraltet={werkstatt.veraltet}
              karteID={werkstatt.karteID}
              setzeKarteID={werkstatt.setzeKarteID}
              karten={werkstatt.karten}
              kartenFehler={werkstatt.kartenFehler}
              symbol={werkstatt.symbol}
              verwendungen={werkstatt.verwendungen}
              verwendungenLaden={werkstatt.verwendungenLaden}
              oeffnen={werkstatt.oeffnen}
            />
          ) : (
            // **Nicht derselbe Satz wie im Zentrum** (UX-15, #1771): zwei
            // gleiche nebeneinander sagen nichts Zweites. Hier steht, was
            // diese Spalte zeigen wird.
            <p className="text-sm text-leise">{eigene.inspectorLeer}</p>
          ),
        }}
      />
    </section>
  )
}
