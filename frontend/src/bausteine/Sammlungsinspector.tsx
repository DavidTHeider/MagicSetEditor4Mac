// **Der Inspector der Sammlung** — Werkstück U8 von Phase 5.1 (#1419): das
// gewählte Exemplar mit Menge, Zustand, Erwerb, Lagerort und Preis, und der
// Ort, an dem geändert und gelöscht wird.
//
// # Woher das Formular kommt
//
// Aus `Sammlungszeile.tsx`, wo es bis U8 in jeder Zeile aufklappte. Mit einer
// Fensterliste über zehntausend Zeilen gehört es an **eine** Stelle: die
// Zeile ist die Wahl, der Inspector die Bearbeitung (0076: Navigator,
// Zentrum, Inspector).
//
// # Der Entwurf hängt am Exemplar
//
// Ein Wechsel des gewählten Exemplars räumt ihn (§7, Regel 1) — über den
// `key` im Aufrufer, nicht über einen Effekt hier: `Sammlung.tsx` schlüsselt
// den Inspector über die Kennung. Ein Filterwechsel, der das Exemplar aus der
// Sicht nimmt, räumt **nichts**: die Änderung trifft weiter dieses Exemplar,
// und der Inspector sagt, dass es nicht mehr zu sehen ist.
import { useState, type ReactNode } from 'react'

import type { ExemplarDTO } from '../api'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { preisAuslegen, preisImFeld } from '../preiseingabe'
import { Anschlussknoepfe } from './Anschlussknoepfe'
import { Aufklappkasten } from './Aufklappkasten'
import { Beschriftetesfeld, eingabeklasse } from './Beschriftetesfeld'
import { Feld } from './Feld'
import { Entscheidung } from './Entscheidung'
import { Aktionsknopf } from './Aktionsknopf'
import { useExemplarhandlungen } from '../haken/useExemplarhandlungen'
import { Knopf } from './Knopf'
import { Exemplardecks } from './Exemplardecks'
import { Exemplarkopf } from './Exemplarkopf'
import { Preise } from './Preise'
import { Zustandswahl } from './Zustandswahl'
import { Zahlenfeld } from './Zahlenfeld'
import { Ausfuehrungswahl } from './Ausfuehrungswahl'

interface InspectorEigenschaften {
  exemplar: ExemplarDTO | undefined
  // inSicht sagt, ob das Exemplar in der gefilterten Liste steht.
  inSicht: boolean
  aufAendern: (exemplar: ExemplarDTO) => Promise<boolean>
  aufLoeschen: (exemplar: ExemplarDTO) => Promise<boolean>
  aufGeholt: () => void
  laeuft: boolean

  // marken sind die Marken des Exemplars — UX-12 (#1684), „+ Marke" im
  // Artboard *Sammlung*. Die Seite baut sie, weil sie den Filter hält.
  marken?: ReactNode
}

export function Sammlungsinspector(eigenschaften: InspectorEigenschaften) {
  const eigene = useTexte().seiten.sammlung
  const { exemplar } = eigenschaften

  if (exemplar === undefined) {
    return <p className="text-sm text-leise">{eigene.inspectorLeer}</p>
  }

  return (
    <Exemplarinspector
      key={exemplar.id}
      {...eigenschaften}
      exemplar={exemplar}
    />
  )
}

type Gewaehlt = Omit<InspectorEigenschaften, 'exemplar'> & {
  exemplar: ExemplarDTO
}

function Exemplarinspector({
  exemplar,
  inSicht,
  aufAendern,
  aufLoeschen,
  aufGeholt,
  laeuft,
  marken,
}: Gewaehlt) {
  const texte = useTexte()
  const eigene = texte.seiten.sammlung
  const [offen, setzeOffen] = useState(false)

  // **Die Rückfrage vor dem Löschen** — UX-10.AK-04.
  const [fragtLoeschen, setzeFragtLoeschen] = useState(false)
  const [entwurf, setzeEntwurf] = useState(exemplar)
  const darstellung = useDarstellung()

  // **Zwei Lesarten stehen am Feld** (LOC-03.AK-03) — und gehen, sobald
  // sich Preis oder Währung ändern.
  const [preisbefund, setzePreisbefund] = useState<string>()

  // **Ein Befund hält die weiteren Angaben offen**, bis jemand neu
  // bearbeitet — wie im Formular zum Aufnehmen (UXR-10).
  const [festgehalten, setzeFestgehalten] = useState(false)

  // **Die Menge wie beim Aufnehmen** (F03, #2245): `Number(wert)` machte aus
  // einem leeren Feld eine 0 und gab „2.5" an ein `int`.
  const [mengeGilt, setzeMengeGilt] = useState(true)

  const setze = (teil: Partial<ExemplarDTO>) => {
    setzeEntwurf((vorher) => ({ ...vorher, ...teil }))
  }

  // **Der Preis in der Schreibweise des Fensters** (Zweitblick über Phase
  // 6.3, S6): im Feld steht, was man selbst so tippte, und hinüber geht die
  // kanonische Form — gelesen wie beim Anlegen (`preiseingabe.ts`).
  const oeffnen = () => {
    setzePreisbefund(undefined)
    setzeFestgehalten(false)
    setzeMengeGilt(true)
    setzeEntwurf({
      ...exemplar,
      erwerbspreis: preisImFeld(
        exemplar.erwerbspreis,
        exemplar.erwerbswaehrung,
        darstellung,
      ),
    })
    setzeOffen(true)
  }

  // **Dieselben Handlungen im Menü „Exemplar“ und am Rechtsklick der Zeile**
  // (AP-6, Teil 2) — der Inspector meldet sie.
  useExemplarhandlungen({
    exemplar,
    laeuft,
    offen,
    fragt: fragtLoeschen,
    bearbeiten: oeffnen,
    loeschen: () => {
      setzeFragtLoeschen(true)
    },
  })

  const sichern = () => {
    const auslegung = preisAuslegen(
      entwurf.erwerbspreis,
      entwurf.erwerbswaehrung,
      darstellung,
      eigene.preisMehrdeutig,
    )

    if (auslegung.art === 'frage') {
      setzePreisbefund(auslegung.befund)
      setzeFestgehalten(true)

      return
    }

    void aufAendern({ ...entwurf, erwerbspreis: auslegung.kanonisch }).then(
      (geglueckt) => {
        if (geglueckt) {
          setzeOffen(false)
        }
      },
    )
  }

  return (
    <div className="flex flex-col gap-3">
      <Exemplarkopf exemplar={exemplar} />

      {!inSicht && (
        <p role="status" className="text-sm text-leise">
          {eigene.inspectorAusserhalb}
        </p>
      )}

      {/* **Der Rückweg** (U9, #1420): welche Decks dieses Stück benutzen. */}
      <Exemplardecks exemplarID={exemplar.id} />

      <div className="flex flex-wrap items-center gap-2">
        {!offen && (
          <Aktionsknopf
            aktion="bearbeiten"
            aufKlick={oeffnen}
            angehalten={laeuft}
            schreibt
          >
            {eigene.bearbeiten}
          </Aktionsknopf>
        )}
        {/* **Während die Frage steht, ruht der Knopf** — sonst stünde
            „Löschen" zweimal da, und ein zweiter Druck ginge gegen einen
            Stand, den der Wirt noch nicht kennt (§7, Regel 4).

            **Verborgen und nicht abgebaut** (QA01, #2382): die Frage gibt
            den Fokus ihrem Auslöser zurück, und ein abgebauter Knopf ist
            danach ein anderer Knoten — der Fokus fiel ins Nirgendwo. */}
        <span hidden={fragtLoeschen} className="contents">
          <Aktionsknopf
            aktion="loeschen"
            aufKlick={() => {
              setzeFragtLoeschen(true)
            }}
            angehalten={laeuft}
            schreibt
          >
            {eigene.loeschen}
          </Aktionsknopf>
        </span>
      </div>

      {/* **Die Rückfrage vor dem Löschen** — UX-10.AK-04
          (#1682).

          **Weil es keine Rücknahme gibt**: Menge, Zustand, Lagerort und
          Erwerbspreis sind von Hand eingetragen, und der Sammlungsbestand
          hat keinen Verlauf wie ein Set. */}
      {marken}

      {fragtLoeschen && (
        <Entscheidung
          frage={eigene.loeschenFrage}
          erklaerung={eigene.loeschenErklaerung}
          bestaetigen={eigene.loeschenJa}
          aktion="loeschen"
          abbrechen={texte.bausteine.aktionen.abbrechen}
          laeuft={laeuft}
          aufBestaetigen={() => {
            setzeFragtLoeschen(false)
            void aufLoeschen(exemplar)
          }}
          aufAbbrechen={() => {
            setzeFragtLoeschen(false)
          }}
        />
      )}

      {/* **Menge, Zustand, Lagerort zuerst** wie im Artboard (UXR-10): statt
          dreizehn Feldern sechs, das Übrige unter „Weitere Angaben". */}
      {offen && (
        <div className="flex flex-col gap-3">
          <div className="flex flex-wrap items-end gap-2">
            <Zahlenfeld
              beschriftung={eigene.menge}
              wert={entwurf.menge}
              kleinstens={1}
              ganzzahlig
              geaendert={(menge) => {
                setze({ menge })
              }}
              gilt={setzeMengeGilt}
            />
            <Zustandswahl
              wert={entwurf.zustand}
              setzeWert={(wert) => {
                setze({ zustand: wert })
              }}
            />
            <Beschriftetesfeld
              bezeichnung={eigene.lagerort}
              wert={entwurf.lagerort}
              setzeWert={(wert) => {
                setze({ lagerort: wert })
              }}
              breite={16}
            />
            <Beschriftetesfeld
              bezeichnung={eigene.druck}
              wert={entwurf.printing}
              setzeWert={(wert) => {
                setze({ printing: wert })
              }}
              breite={24}
            />
            <Beschriftetesfeld
              bezeichnung={eigene.sprache}
              wert={entwurf.sprache}
              setzeWert={(wert) => {
                setze({ sprache: wert })
              }}
              breite={4}
            />
            <Ausfuehrungswahl
              wert={entwurf.finish}
              setzeWert={(wert) => {
                setze({ finish: wert })
              }}
            />
          </div>
          <Aufklappkasten
            titel={eigene.weitereAngaben}
            offenGehalten={festgehalten}
          >
            <div className="flex flex-wrap items-end gap-2">
              <Beschriftetesfeld
                bezeichnung={eigene.name}
                wert={entwurf.name}
                setzeWert={(wert) => {
                  setze({ name: wert })
                }}
                breite={20}
              />
              <Beschriftetesfeld
                bezeichnung={eigene.setcode}
                wert={entwurf.setcode}
                setzeWert={(wert) => {
                  setze({ setcode: wert })
                }}
                breite={5}
              />
              <Beschriftetesfeld
                bezeichnung={eigene.nummer}
                wert={entwurf.nummer}
                setzeWert={(wert) => {
                  setze({ nummer: wert })
                }}
                breite={5}
              />
              {/* **Über `Feld`**, damit der Befund am Feld steht (§6). */}
              <Feld
                beschriftung={eigene.erwerbspreis}
                {...(preisbefund === undefined ? {} : { befund: preisbefund })}
              >
                {(verbindung) => (
                  <input
                    {...verbindung}
                    value={entwurf.erwerbspreis}
                    size={12}
                    onChange={(ereignis) => {
                      setzePreisbefund(undefined)
                      setze({ erwerbspreis: ereignis.target.value })
                    }}
                    className={eingabeklasse}
                  />
                )}
              </Feld>
              <Beschriftetesfeld
                bezeichnung={eigene.erwerbswaehrung}
                wert={entwurf.erwerbswaehrung}
                setzeWert={(wert) => {
                  setzePreisbefund(undefined)
                  setze({ erwerbswaehrung: wert })
                }}
                breite={5}
              />
              <Beschriftetesfeld
                bezeichnung={eigene.erwerbsdatum}
                wert={entwurf.erwerbsdatum}
                setzeWert={(wert) => {
                  setze({ erwerbsdatum: wert })
                }}
                art="date"
              />
              <Beschriftetesfeld
                bezeichnung={eigene.notiz}
                wert={entwurf.notiz}
                setzeWert={(wert) => {
                  setze({ notiz: wert })
                }}
                breite={24}
              />
            </div>
          </Aufklappkasten>
          <div className="flex flex-wrap gap-2">
            <Aktionsknopf
              aktion="sichern"
              bezeichnung={eigene.aenderungSichern}
              aufKlick={sichern}
              gewicht="haupt"
              angehalten={laeuft || !mengeGilt}
            >
              {eigene.aenderungSichern}
            </Aktionsknopf>
            <Knopf
              bezeichnung={eigene.abbrechen}
              aufKlick={() => {
                setzeOffen(false)
              }}
            >
              {eigene.abbrechen}
            </Knopf>
          </div>
        </div>
      )}

      <Preise printing={exemplar.printing} aufGeholt={aufGeholt} />

      {/* **Wohin das Exemplar weiterreist** — UX-03.AK-01
          (#1675).
          Ganz unten, wie im Entwurf: darüber steht, was man **an** ihm tut. */}
      <Anschlussknoepfe exemplar={exemplar} />
    </div>
  )
}
