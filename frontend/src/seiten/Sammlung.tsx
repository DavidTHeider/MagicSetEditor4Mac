// **Die Sammlung als Arbeitsplatz** — Werkstück U8 von Phase 5.1 (#1419):
// Navigator mit Sicht, Marken, Massenvorgängen und Portfolio; Zentrum mit
// der Liste in drei Ansichten über eine Fensterliste (Vermerk 0085);
// Inspector mit dem gewählten Exemplar. Was die Seite tut, steht in
// `haken/useSammlung.ts` (§2).
//
// **Aufnehmen und Einlesen stehen im Kopf** (UXR-10, #1918, Artboard
// `Sammlung`) und öffnen je eine Schublade — vorher standen achtzehn Felder
// in 240 Punkten im Navigator, gemessen am 2026-10-02. **Darunter die
// Leiste** (AP-9, #2366), dann im Navigator zuerst die Sicht mit dem
// Suchfeld (`docs/UX-ARCHITEKTUR.md`).
import { useEffect, useState } from 'react'

import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Arbeitsplatz } from '../bausteine/Arbeitsplatz'
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Hinweiskasten } from '../bausteine/Hinweiskasten'
import { Knopf } from '../bausteine/Knopf'
import { Massenvorgaenge } from '../bausteine/Massenvorgaenge'
import { Mengenentscheidung } from '../bausteine/Mengenentscheidung'
import { mengenschritte, zurZeileSchliesst } from '../haken/mengenentscheidung'
import { Portfolioansicht } from '../bausteine/Portfolio'
import { Preisbulk } from '../bausteine/Preisbulk'
import { Sammlungsinspector } from '../bausteine/Sammlungsinspector'
import { Sammlungsordnung } from '../bausteine/Sammlungsordnung'
import { useAnsicht } from '../haken/useAnsicht'
import { useAnstoss } from '../haken/useAnstoss'
import { sammlungsansichtsschluessel, sammlungsformen } from './sammlungssicht'
import { Sammlungsleiste } from '../bausteine/Sammlungsleiste'
import { Sammlungssicht } from '../bausteine/Sammlungssicht'
import {
  Sammlungswerkzeuge,
  type Werkzeug,
} from '../bausteine/Sammlungswerkzeuge'
import { Seitenkopf } from '../bausteine/Seitenkopf'
import { Sammlungskennzahlen, useSammlungsleiste } from './Sammlungskennzahlen'
import { useArbeitsplatzgedaechtnis } from '../haken/useArbeitsplatzgedaechtnis'
import { exemplarschluessel, useSammlung } from '../haken/useSammlung'
import { useTexte } from '../sprache/kontext'
import { Verdeckung, useVerdeckt } from '../verdeckung'
import { Sammlungsliste } from './Sammlungsliste'
import { Sammlungszustaende } from './Sammlungszustaende'
import { Auswahlband } from '../bausteine/Auswahlband'
import { Markenfilterkasten } from '../bausteine/Markenfilterkasten'
import { Markenpflege, Markenvergabe } from '../bausteine/Markenpflege'
import { Objektkontextmenue } from '../bausteine/Objektkontextmenue'
import { Objektmarken } from '../bausteine/Objektmarken'
import { useMarkenfilter } from '../haken/useMarkenfilter'

// bestandsstand zählt fremde Schreibvorgänge — die Sammlung lädt dann nach
// (Befund 3 des Gesamtreviews über 0.6.0). **`bestandGeaendert` meldet die
// eigenen** dem Fenster (R3 des Nachlaufs über Phase 6 von 0.9.0): erst dann
// liest die Deckseite ihren Abgleich neu.
export function Sammlung({
  bestandsstand = 0,
  bestandGeaendert,
}: {
  bestandsstand?: number
  bestandGeaendert: () => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.sammlung
  // **Marken und gespeicherte Filter** — UX-12 (#1684), Artboard *Sammlung*.
  const marken = useMarkenfilter('exemplar')
  const sammlung = useSammlung(
    marken.getroffen,
    bestandsstand,
    bestandGeaendert,
  )

  // **Ein Rechtsklick oder ⇧F10 auf einer anderen Zeile schlägt sie zuerst
  // auf** — AP-6, Teil 2 (#2363); wie, steht am Haken.
  const { anstoss, verbraucht, anstossen } = useAnstoss<number>(
    sammlung.gehalten === undefined ? undefined : String(sammlung.gehalten),
    sammlung.waehlen,
  )
  const gedaechtnis = useArbeitsplatzgedaechtnis()
  const leiste = useSammlungsleiste(sammlung.runde, gedaechtnis)

  // **Wie oft eine Dublette zusammengeführt wurde** (DOM-03) — dann ist die
  // Aufnahme geschehen, und das Formular räumt sich wie nach einem Erfolg.
  const [geleert, setzeGeleert] = useState(0)
  const { dublette } = sammlung

  // **Welche Schublade steht** — UXR-10. Darunter ist die Seite gesperrt
  // (`inert`) und verdeckt (`verdeckung.ts`): ⌥⌘[ gehörte sonst einem
  // Navigator, den niemand sieht.
  const [schublade, setzeSchublade] = useState<Werkzeug>()
  const verdeckt = useVerdeckt() || schublade !== undefined

  // **Die Ansicht wird gemerkt** — UX-11.AK-04 (#1683): je Profil, geerbt
  // vom Menschen. Was gilt, kommt einmal in die Sicht; wer umschaltet,
  // schreibt es fort.
  const ansicht = useAnsicht(sammlungsansichtsschluessel, sammlungsformen)
  const { setzeSicht } = sammlung
  useEffect(() => {
    setzeSicht({ ansicht: ansicht.form })
  }, [ansicht.form, setzeSicht])

  const {
    sichtbare,
    gesamt,
    insgesamt,
    sicht,
    laedt,
    ladefehler,
    fehler,
    gehaltenes,
    vorgang,
  } = sammlung

  // setzeTeil schreibt eine Sicht fort — und eine gewählte Ansicht in die
  // Einstellung, für die Sicht im Navigator wie für die Leiste darüber.
  const setzeTeil = (teil: Partial<typeof sicht>) => {
    if (teil.ansicht !== undefined) {
      ansicht.setzeForm(teil.ansicht)
    }
    setzeSicht(teil)
  }

  // zuruecksetzen nimmt Suche, Ausführung und Marken zurück — **ein Weg
  // für beide Knöpfe**, den im leeren Zustand und den in der Sicht.
  const zuruecksetzen = () => {
    setzeSicht({ suche: '', ausfuehrung: 'alle' })
    marken.setzeFilter({ ...marken.filter, marken: [] })
  }

  // **Der Schreibfehler geht vor** — er ist das neueste Wort; ein Lesefehler
  // steht nur, solange keiner dasteht.
  const seitenfehler = fehler ?? ladefehler ?? marken.fehler

  // **Ein Konflikt bietet den heutigen Stand an** (UX-08.AK-02, #1680) —
  // „Bitte neu laden" ohne einen Knopf dazu schickte suchen. **Er steht in
  // der Schublade, solange eine steht** (`Sammlungswerkzeuge`).
  const konfliktweg = sammlung.konflikt
    ? {
        beschriftung: texte.bausteine.zustaende.neuLaden,
        aufKlick: sammlung.nachKonflikt,
      }
    : undefined
  // **Ein Lesefehler bietet den nächsten Versuch an** — G11 (#2277): ohne
  // Knopf blieb nur, die Seite zu verlassen.
  const wiederholung = {
    beschriftung: texte.bausteine.zustaende.wiederholen,
    aufKlick: ladefehler !== undefined ? sammlung.nachladen : marken.nachladen,
  }
  const meldung = seitenfehler !== undefined && (
    <Hinweiskasten
      art="fehler"
      meldung={seitenfehler}
      handlung={fehler === undefined ? wiederholung : konfliktweg}
    />
  )

  return (
    <>
      <Verdeckung value={verdeckt}>
        <div className="flex flex-col gap-4" inert={schublade !== undefined}>
          {/* **Eine führende Handlung** (R1): *+ Aufnehmen*, wie im
              Artboard; das Einlesen daneben ist eine Nebenhandlung. */}
          <Seitenkopf titel={eigene.titel} erklaerung={eigene.erklaerung}>
            <Knopf
              aufKlick={() => {
                setzeSchublade('einfuhr')
              }}
            >
              {eigene.einfuhrOeffnen}
            </Knopf>
            <Aktionsknopf
              aktion="anlegen"
              gewicht="haupt"
              schreibt
              aufKlick={() => {
                setzeSchublade('aufnehmen')
              }}
            >
              {eigene.aufnehmenOeffnen}
            </Aktionsknopf>
          </Seitenkopf>

          {schublade === undefined && meldung}

          {/* **Die Leiste über der Sammlung** — AP-9 (#2366). */}
          <Sammlungskennzahlen
            stueck={sammlung.stueck}
            insgesamt={insgesamt}
            leiste={leiste}
            zeigeAlle={zuruecksetzen}
          />

          <Arbeitsplatz
            bezeichnung={eigene.arbeitsplatz}
            gedaechtnis={gedaechtnis}
            beiEnge="inspector"
            navigator={{
              bezeichnung: eigene.navigator,
              inhalt: (
                <div className="flex flex-col gap-4">
                  {/* **Erst, wenn es etwas zu sehen gibt**: eine Suche über
                      eine leere Sammlung ist eine Zeile, die erklärt werden
                      müsste. */}
                  {/* **Die Zahlen kommen aus der Antwort** (#1500) — nicht
                      aus der Länge der Liste: die ist seit dem Ausschnitt
                      die des Fensters und nicht die der Sammlung. */}
                  {/* **„3 Exemplare → Deck"** — UX-03.AK-01 (#1675), das
                      Band aus dem Artboard: oben im Navigator, solange etwas
                      gewählt ist. */}
                  <Auswahlband
                    exemplare={sammlung.exemplare.filter((exemplar) =>
                      sammlung.gewaehlt.has(exemplar.id),
                    )}
                  />
                  {insgesamt > 0 && (
                    <Sammlungssicht
                      sicht={sicht}
                      setze={setzeTeil}
                      gezeigt={
                        marken.getroffen === undefined
                          ? gesamt
                          : sichtbare.length
                      }
                      gesamt={insgesamt}
                      zuruecksetzen={zuruecksetzen}
                    />
                  )}
                  {insgesamt > 0 && (
                    <Markenfilterkasten art="exemplar" stand={marken} />
                  )}
                  {/* **Was man mit einer Auswahl tut** — die Marke für alle
                      gewählten und die Massenänderung, beides erst, wenn
                      etwas angehakt ist (UXR-10). */}
                  {sammlung.gewaehlt.size > 0 && (
                    <Markenvergabe
                      art="exemplar"
                      kennungen={[...sammlung.gewaehlt].map(String)}
                      geaendert={marken.nachladen}
                    />
                  )}
                  {insgesamt > 0 && (
                    <Massenvorgaenge
                      gewaehlt={sammlung.gewaehlteZeilen}
                      gesamt={insgesamt}
                      aufGeaendert={sammlung.nachMasse}
                    />
                  )}
                  {/* **Das Seltene eingeklappt** — UXR-10 (#1918): die drei
                      Fragen an eine geordnete Sammlung (#1292, Vermerk
                      0057), die Pflege der Marken und der Bulkbezug (A3,
                      #1442). Man braucht sie, wenn man aufräumt, nicht bei
                      jedem Blick; zu fragen beginnen sie erst offen. */}
                  {insgesamt > 0 && (
                    <Aufklappkasten titel={eigene.ordnungTitel}>
                      <Sammlungsordnung />
                    </Aufklappkasten>
                  )}
                  {insgesamt > 0 && (
                    <Aufklappkasten
                      titel={texte.bausteine.objektmarken.pflegeUeberschrift}
                    >
                      <Markenpflege
                        runde={marken.runde}
                        geaendert={marken.nachladen}
                      />
                    </Aufklappkasten>
                  )}
                  {/* **Der Bulkbezug steht vor dem Portfolio** (A3, #1442):
                      er füllt, was das Portfolio rechnet. */}
                  <Aufklappkasten titel={eigene.bulkTitel}>
                    <Preisbulk aufGeholt={sammlung.nachMasse} />
                  </Aufklappkasten>
                  <Portfolioansicht
                    stand={leiste.portfolio}
                    anlass={leiste.anlass}
                  />
                  {/* **Der Vorbehalt steht beim Portfolio**: die Quelle sagt
                      ausdrücklich, dass sie für keinen Preis einsteht; wer
                      eine Summe liest, soll das mitlesen. */}
                  <p className="text-xs text-leise">{eigene.vorbehalt}</p>
                </div>
              ),
            }}
            zentrum={
              <div className="flex h-full min-h-0 flex-col gap-2">
                {/* **Wie die Liste aussieht, steht über ihr** — UXR-10, nach
                    dem Artboard: Ansicht als Segmente, Sortierung und
                    Gruppierung daneben (`Sammlungsleiste`). */}
                {insgesamt > 0 && (
                  <Sammlungsleiste sicht={sicht} setze={setzeTeil} />
                )}
                <Sammlungszustaende
                  insgesamt={insgesamt}
                  sichtbar={sichtbare.length}
                  laedt={laedt}
                  geladen={sammlung.geladen}
                  aufnehmen={() => {
                    setzeSchublade('aufnehmen')
                  }}
                  einlesen={() => {
                    setzeSchublade('einfuhr')
                  }}
                  zuruecksetzen={zuruecksetzen}
                />
                {/* **Eine bestimmte Höhe für das Fenster der Liste.**
                    `flex-1` löste sie gegen einen unbestimmten Rahmen auf,
                    und `clientHeight` war die Höhe aller zehntausend Zeilen
                    — die Fensterliste zeichnete alles (Sichtprüfung
                    2026-09-09: 10.001 Zeilen im DOM, 1,3 s). **Seit UX-15
                    (#1771) ist der Rahmen bestimmt**: das Zentrum hat die
                    Höhe bis an den Fensterrand, und die Liste nimmt, was
                    darin bleibt, statt einer zweiten Rechnung über `100vh`,
                    die zwei Rollbalken übereinander gab. */}
                {sichtbare.length > 0 && (
                  <Objektkontextmenue
                    art="exemplar"
                    kennung={String(sammlung.gehalten ?? '')}
                    klasse="min-h-0 flex-1"
                    anstoss={anstoss}
                    anstossVerbraucht={verbraucht}
                    nurAngestossen
                  >
                    <Sammlungsliste
                      exemplare={sichtbare}
                      sicht={sicht}
                      gewaehlt={sammlung.gewaehlt}
                      anhaken={sammlung.anhaken}
                      gehalten={sammlung.gehalten}
                      waehlen={sammlung.waehlen}
                      aufFenster={sammlung.aufFenster}
                      anstossen={(ereignis, id) => {
                        anstossen(ereignis, String(id), id)
                      }}
                    />
                  </Objektkontextmenue>
                )}
              </div>
            }
            inspector={{
              bezeichnung: eigene.inspector,
              inhalt: (
                <Sammlungsinspector
                  exemplar={gehaltenes}
                  inSicht={sammlung.inSicht}
                  aufAendern={sammlung.aendern}
                  aufLoeschen={sammlung.loeschen}
                  aufGeholt={sammlung.nachSchreiben}
                  laeuft={
                    gehaltenes !== undefined &&
                    vorgang.laeuft(exemplarschluessel(gehaltenes.id))
                  }
                  marken={
                    gehaltenes === undefined ? undefined : (
                      <Objektmarken
                        art="exemplar"
                        kennung={String(gehaltenes.id)}
                        marken={marken.je[String(gehaltenes.id)] ?? []}
                        geaendert={marken.nachladen}
                      />
                    )
                  }
                />
              ),
            }}
          />
        </div>
      </Verdeckung>

      <Sammlungswerkzeuge
        offen={schublade}
        schliessen={() => {
          setzeSchublade(undefined)
        }}
        meldung={meldung}
        anlegen={sammlung.anlegen}
        laeuft={vorgang.laeuft('neu')}
        nachladen={sammlung.nachSchreiben}
        geleert={geleert}
        entscheidung={
          dublette !== undefined && (
            <Mengenentscheidung
              dublette={dublette}
              schritte={zurZeileSchliesst(
                mengenschritte(dublette, sammlung, () => {
                  setzeGeleert((bisher) => bisher + 1)
                }),
                () => {
                  setzeSchublade(undefined)
                },
              )}
            />
          )
        }
      />
    </>
  )
}
