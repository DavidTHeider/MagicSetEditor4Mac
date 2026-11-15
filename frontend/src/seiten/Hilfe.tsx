// **Die Hilfe** — #679.
//
// # Warum es sie gibt
//
// Die Anwendung arbeitet mit Begriffen, die aus MSE stammen und die niemand
// mitbringt: Stilbogen, Stiloption, Symbolschrift, Exportvorlage, Druckbogen,
// Vorbehalt. Wer sie nicht kennt, fand in der Anwendung nichts, was sie
// erklärt — der Hilfekatalog hatte neun Zeilen und keinen Leser.
//
// # Was sie nicht ist
//
// **Kein Handbuch neben der Anwendung**, und keine Übersetzung der
// MSE-Dokumentation. Was aus einem Paket kommt, bleibt unübersetzt (§9); hier
// stehen die Begriffe der **Anwendung**, je in einem Satz.
import { useEffect, useId, useState } from 'react'

import { Flaeche } from '../bausteine/Flaeche'
import { Gruppenliste } from '../bausteine/Gruppenliste'
import { seitenschluessel, type Seitenschluessel } from '../navigation'
import { Hilfedatenschutz } from './Hilfedatenschutz'
import { Hilfebetrieb } from './Hilfebetrieb'
import { Hilfeschnittstellen } from './Hilfeschnittstellen'
import { Hilfebereiche, Hilfefuehrungen } from './Hilfefuehrungen'
import { Hilfeherkunft, Hilfesuchfeld, Hilfetreffer } from './Hilfesuche'
import { Hilfeueber } from './Hilfeueber'
import { Hilfeweiterlesen } from './Hilfeweiterlesen'
import { Kuerzeltafel } from './Kuerzeltafel'
import { Satz } from '../bausteine/Satz'
import { Seitenkopf } from '../bausteine/Seitenkopf'
import { useTexte } from '../sprache/kontext'

interface HilfeEigenschaften {
  // fassung ist die laufende Fassung — sie steht im Abschnitt „Über" (#1168).
  fassung: string

  // zuUeber sagt, dass der Abschnitt „Über" gemeint ist: der Menüeintrag
  // führt hierher, und ohne das Rollen läge er unter zehn Bildschirmhöhen.
  zuUeber: boolean

  // einfuehrungZeigen holt die ersten Schritte zurück — aus den Führungen (#1422).
  einfuehrungZeigen: () => void

  // herkunft ist die Seite, von der man kam (UXR-22) — ihre Hilfe steht
  // obenan, wie „Wo du gerade warst" im Artboard.
  herkunft?: Seitenschluessel | undefined

  // angekommen meldet, dass gerollt wurde — der Zustand oben wird dabei
  // zurückgesetzt, sonst rollte jeder weitere Besuch der Hilfe wieder dorthin.
  angekommen: () => void
}

export function Hilfe({
  fassung,
  zuUeber,
  angekommen,
  einfuehrungZeigen,
  herkunft,
}: HilfeEigenschaften) {
  const texte = useTexte()
  const { hilfe } = texte

  // **Der Anker steht hier und nicht im Aufrufer** — welchen Abschnitt es
  // gibt, weiß diese Seite.
  const ueberID = useId()

  // **Der Wunsch, das Gebiet »Über« aufzuschlagen** — Werkstück I von Phase
  // 4.10 (#1325).
  //
  // Seit die Hilfe Gebiete führt (Laschen, seit AP-11 ein
  // Inhaltsverzeichnis), steht der Abschnitt nicht mehr im Baum, solange sein
  // Gebiet zu ist: `scrollIntoView` fand nichts, und der Menüeintrag „Über
  // MagicSetEditor4Mac" führte ins Leere. Die Zahl macht
  // jeden Ruf unterscheidbar, damit er **jedes Mal** wirkt und nicht nur beim
  // ersten.
  // **Eine Zählung und keine Uhrzeit.** `Date.now()` beim Zeichnen ist
  // unrein: derselbe Zustand ergäbe zwei verschiedene Ergebnisse. Die Zahl
  // muss nur **anders** sein als die vorige, damit das Inhaltsverzeichnis den
  // Wunsch als neuen erkennt.
  // **Und seit U11 ist es nicht nur „Über"**: der KI-Bereich führt zu den
  // Schnittstellen — derselbe Wunsch, ein anderes Gebiet.
  const [wunsch, setzeWunsch] = useState<{
    gebiet: 'ueber' | 'schnittstellen' | 'seiten'
    zahl: number
  }>({ gebiet: 'ueber', zahl: 0 })
  const oeffne =
    wunsch.zahl === 0 ? undefined : `${wunsch.gebiet}#${String(wunsch.zahl)}`

  // **Beim Zeichnen und nicht in einem Effekt.** React rät von einem
  // `setState` im Effekt ab — es löst ein zweites Zeichnen aus, das sich
  // vermeiden lässt: der Wunsch hängt allein an `zuUeber`, und der steht beim
  // Zeichnen schon fest.
  // **Beginnt bei `false`, nicht bei `zuUeber`.** Der Menüeintrag wird
  // gedrückt, **bevor** die Hilfeseite steht: sie kommt mit `zuUeber === true`
  // zur Welt. Ein Anfangswert von `zuUeber` sähe darin keine Änderung, und der
  // Weg führte ins Leere — gemessen am 2026-09-03, `menueleiste.test.tsx` fand
  // die Überschrift nicht.
  const [gesehen, setzeGesehen] = useState(false)

  if (zuUeber !== gesehen) {
    setzeGesehen(zuUeber)

    if (zuUeber) {
      setzeWunsch((bisher) => ({ gebiet: 'ueber', zahl: bisher.zahl + 1 }))
    }
  }

  // **Gerollt wird erst, wenn das Gebiet steht** — in einem zweiten Effekt,
  // der am Wunsch hängt: React zeichnet zwischen den beiden, und vorher gäbe
  // es den Abschnitt nicht.
  useEffect(() => {
    if (oeffne === undefined || wunsch.gebiet !== 'ueber') {
      return
    }

    document.getElementById(ueberID)?.scrollIntoView({ block: 'start' })
    angekommen()
  }, [oeffne, wunsch.gebiet, ueberID, angekommen])

  // **Die Reihenfolge ist die des Weges**, nicht das Alphabet: ein Set trägt
  // Karten, eine Karte hängt an einem Spiel und einem Stil, und ganz am Ende
  // steht, was beim Hinausgeben herauskommt.
  const begriffe = [
    hilfe.begriffe.set,
    hilfe.begriffe.karte,
    hilfe.begriffe.spiel,
    hilfe.begriffe.stil,
    hilfe.begriffe.stiloption,
    hilfe.begriffe.symbolschrift,
    hilfe.begriffe.paket,
    hilfe.begriffe.exportvorlage,
    hilfe.begriffe.druckbogen,
    hilfe.begriffe.vorbehalt,
    hilfe.begriffe.befund,
  ]

  // **Die Reihenfolge kommt aus `navigation.ts`**, nicht aus dem Katalog.
  //
  // Ein `Object.entries` über den Hilfekatalog gäbe die Reihenfolge, in der
  // jemand die Einträge geschrieben hat — und die verschiebt sich beim nächsten
  // Nachtragen. Hier steht dieselbe Folge wie in der Seitenleiste: erst die
  // fünf Seiten der Anwendung, dann die acht Laschen des Sets.
  const seitenhilfen = seitenschluessel.map(
    (schluessel) => hilfe.seiten[schluessel],
  )

  const sprachpunkte = [
    { name: hilfe.sprache.schluesselName, satz: hilfe.sprache.schluesselSatz },
    { name: hilfe.sprache.uebersetztName, satz: hilfe.sprache.uebersetztSatz },
    { name: hilfe.sprache.fensterName, satz: hilfe.sprache.fensterSatz },
    { name: hilfe.sprache.leisteName, satz: hilfe.sprache.leisteSatz },
  ]

  // **Die Frage und was sie durchsucht** — UXR-22. Die Gebiete aus Namen
  // und Sätzen; was eine Tabelle ist oder einen Knopf trägt, findet man
  // über seine Lasche.
  const [suche, setzeSuche] = useState('')
  const sucht = suche.trim() !== ''
  const { fuehrungen } = hilfe
  const durchsucht = [
    { name: texte.seiten.hilfe.glossar, punkte: begriffe },
    {
      name: texte.seiten.hilfe.seitenuebersicht,
      punkte: seitenhilfen.map((seite) => ({
        name: seite.ueberschrift,
        satz: seite.text,
      })),
    },
    { name: hilfe.sprache.ueberschrift, punkte: sprachpunkte },
    {
      name: fuehrungen.ueberschrift,
      punkte: [
        { name: fuehrungen.startName, satz: fuehrungen.startSatz },
        {
          name: fuehrungen.arbeitsplatzName,
          satz: fuehrungen.arbeitsplatzSatz,
        },
        { name: fuehrungen.paletteName, satz: fuehrungen.paletteSatz },
        { name: fuehrungen.leisteName, satz: fuehrungen.leisteSatz },
      ],
    },
  ]

  return (
    <>
      <Seitenkopf
        titel={texte.seiten.hilfe.titel}
        erklaerung={texte.seiten.hilfe.erklaerung}
      />

      {/* **Gebiete statt Flächen untereinander** — Werkstück I von Phase
          4.10
          (#1325).

          Sprungmarken hülfen gegen die Höhe, ließen aber alles zugleich
          stehen. Wer die Hilfe aufschlägt, sucht **eine** Antwort — und liest
          die Hilfe nicht von vorn nach hinten.

          **Die Begriffe stehen zuerst** (UXR-22, #1930): die Übersicht aller
          Seiten stand hier bis dahin vorn (#806) und machte die Rolltiefe von
          4,1. Die Seite, um die es geht, steht seither darüber.

          **Die Ordnung steht an einer Stelle** — in dieser Liste. */}
      <Hilfesuchfeld suche={suche} setzeSuche={setzeSuche} />

      {sucht && (
        <Hilfetreffer
          suche={suche}
          gebiete={durchsucht}
          leeren={() => {
            setzeSuche('')
          }}
        />
      )}

      {!sucht && herkunft !== undefined && (
        <div className="mb-4">
          <Hilfeherkunft
            herkunft={herkunft}
            zurUebersicht={() => {
              setzeWunsch((bisher) => ({
                gebiet: 'seiten',
                zahl: bisher.zahl + 1,
              }))
            }}
          />
        </div>
      )}

      {/* **Verborgen und nicht abgeräumt**, solange gesucht wird: das offene
          Gebiet bleibt, wenn das Feld wieder leer ist. */}
      <div hidden={sucht}>
        {/* **Ein Inhaltsverzeichnis links statt elf Laschen in zwei Reihen**
            — AP-11 (#2368), B-29, `Ruhe09`; ↑ und ↓ gehen darin weiter. Die
            Suche bleibt darüber. */}
        <Gruppenliste
          bezeichnung={texte.seiten.hilfe.aufDieserSeite}
          oeffne={oeffne}
          gebiete={[
            {
              schluessel: 'glossar',
              name: texte.seiten.hilfe.glossar,
              inhalt: (
                <Flaeche
                  ueberschrift={texte.seiten.hilfe.glossar}
                  sprungmarke="hilfe-glossar"
                >
                  {/* **Eine Beschreibungsliste und keine Aufzählung**: Begriff und Satz
              gehören zusammen, und ein Vorleseprogramm sagt das dann auch. */}
                  <dl className="grid gap-3 sm:grid-cols-[10rem_minmax(0,1fr)]">
                    {begriffe.map((begriff) => (
                      <div key={begriff.name} className="contents">
                        <dt className="font-semibold">{begriff.name}</dt>
                        <dd className="max-w-prose text-leise">
                          <Satz text={begriff.satz} />
                        </dd>
                      </div>
                    ))}
                  </dl>
                </Flaeche>
              ),
            },
            {
              schluessel: 'seiten',
              name: texte.seiten.hilfe.seitenuebersicht,
              inhalt: (
                <Flaeche
                  ueberschrift={texte.seiten.hilfe.seitenuebersicht}
                  sprungmarke="hilfe-seiten"
                >
                  <dl className="grid gap-3 sm:grid-cols-[10rem_minmax(0,1fr)]">
                    {seitenhilfen.map((seite) => (
                      <div key={seite.ueberschrift} className="contents">
                        <dt className="font-semibold">{seite.ueberschrift}</dt>
                        <dd className="max-w-prose text-leise">
                          <Satz text={seite.text} />
                        </dd>
                      </div>
                    ))}
                  </dl>
                  <Hilfebereiche
                    zuSchnittstellen={() => {
                      setzeWunsch((bisher) => ({
                        gebiet: 'schnittstellen',
                        zahl: bisher.zahl + 1,
                      }))
                    }}
                  />
                </Flaeche>
              ),
            },
            {
              schluessel: 'fuehrungen',
              name: hilfe.fuehrungen.ueberschrift,
              inhalt: <Hilfefuehrungen einfuehrungZeigen={einfuehrungZeigen} />,
            },
            {
              schluessel: 'kuerzel',
              name: texte.seiten.einstellungen.kuerzelUeberschrift,
              inhalt: <Kuerzeltafel />,
            },
            {
              schluessel: 'sprache',
              name: hilfe.sprache.ueberschrift,
              inhalt: (
                <Flaeche
                  ueberschrift={hilfe.sprache.ueberschrift}
                  sprungmarke="hilfe-sprache"
                >
                  <p className="mb-4 max-w-prose text-leise">
                    <Satz text={hilfe.sprache.einleitung} />
                  </p>

                  <dl className="grid gap-3 sm:grid-cols-[10rem_minmax(0,1fr)]">
                    {sprachpunkte.map((punkt) => (
                      <div key={punkt.name} className="contents">
                        <dt className="font-semibold">{punkt.name}</dt>
                        <dd className="max-w-prose text-leise">
                          <Satz text={punkt.satz} />
                        </dd>
                      </div>
                    ))}
                  </dl>
                </Flaeche>
              ),
            },
            {
              schluessel: 'schnittstellen',
              name: hilfe.schnittstellen.ueberschrift,
              inhalt: <Hilfeschnittstellen />,
            },
            // **Und was davon über dich ist** — Werkstück 13 von Phase 4.12
            // (#1351). **Neben den Schnittstellen und nicht darin**: jene
            // sagen, welche Wege es gibt, dieser, was auf ihnen von **dir**
            // hinausgeht. Zwei Fragen, zwei Gebiete.
            {
              schluessel: 'datenschutz',
              name: hilfe.datenschutz.ueberschrift,
              inhalt: <Hilfedatenschutz />,
            },
            {
              // **Betrieb und Vertrauen** (OPS-07, #1724) — was installiert
              // ist, wo es liegt und was geprüft wurde.
              schluessel: 'betrieb',
              name: hilfe.betrieb.ueberschrift,
              inhalt: <Hilfebetrieb />,
            },
            {
              schluessel: 'erststart',
              name: hilfe.erststart.ueberschrift,
              inhalt: (
                <Flaeche
                  ueberschrift={hilfe.erststart.ueberschrift}
                  sprungmarke="hilfe-erststart"
                >
                  <p className="max-w-prose text-leise">
                    {hilfe.erststart.text}
                  </p>
                </Flaeche>
              ),
            },
            {
              schluessel: 'weiterlesen',
              name: hilfe.weiterlesen.ueberschrift,
              inhalt: <Hilfeweiterlesen />,
            },
            {
              schluessel: 'ueber',
              name: hilfe.ueber.ueberschrift,
              inhalt: (
                <div id={ueberID}>
                  <Hilfeueber fassung={fassung} />
                </div>
              ),
            },
          ]}
        />
      </div>
    </>
  )
}
