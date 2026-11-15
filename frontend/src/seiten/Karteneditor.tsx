// Der Karteneditor mit Live-Vorschau.
//
// # Die vier Regeln aus AGENTS.md §7, alle vier an einer Stelle
//
//  1. **Ein Wechsel des Bezugsobjekts räumt den offenen Kontext.** `karteID`
//     steckt in den Abhängigkeiten jedes Ladelaufs; wer die Karte wechselt,
//     bekommt neue Werte und eine neue Vorschau, und der Feldfehler der alten
//     verschwindet.
//  2. **Eine Antwort gehört zu dem Stand, aus dem sie stammt.** Das erledigt
//     `useLaden` über `useLauf` — für Erfolg und Fehler.
//  3. **Geschrieben ist geschrieben.** Nach einem geglückten Schreiben wird die
//     Vorschau nachgeladen; nach einem gescheiterten **nicht** — sonst
//     überschriebe ein Lesefehler die Meldung, die sagt, was zu tun ist.
//  4. **Ein Vorgang je Datensatz.** `useVorgang` hält den zweiten Aufruf an,
//     solange der erste läuft.
import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'

import { KIKasten } from '../bausteine/KIKasten'
import { KIBereich } from '../bausteine/KIBereich'
import { useBildfeld } from '../haken/useBildfeld'
import { useScryfallabgleich } from '../haken/useScryfallabgleich'
import { useFeldschreiben } from '../haken/useFeldschreiben'
import { useWortwahl } from '../haken/useWortwahl'
import { useKartenvorschau } from '../haken/useKartenvorschau'
import { useAutoersetzung } from '../haken/useAutoersetzung'
import { useVorgang } from '../haken/useVorgang'
import { useTexte } from '../sprache/kontext'
import {
  angeboteneFelder,
  bildfelderAus,
  kartennameAus,
} from './angebotenefelder'
import { Feldformular } from './Feldformular'
import { useFeldtiefe } from '../haken/useFeldtiefe'
import { Feldzeile } from './Feldzeile'
import { Stilfelder } from './Stilfelder'
import type { Feldfehler } from './feldkennung'
import { feldvorschlag } from './feldvorschlag'
import { Kartenvorschau } from './Kartenvorschau'
import { useSymbolpalette, useWortlisten } from './feldbeiwerk'
import { feldFokussieren, symbolEinfuegenIn } from './karteneditorgriffe'
import { Problemliste, problemeAus } from './Problemliste'
import { vergroesserungFuer } from './zoom'
import { useVorschauansicht } from './vorschauansicht'
import { useFeldordnung } from './formularruhe'
import type { KarteneditorEigenschaften } from './karteneditoreigenschaften'

export type { Feldwahl } from './karteneditoreigenschaften'

export function Karteneditor({
  setID,
  karte,
  formen,
  geaendert,
  symbolZeichnen,
  feldwahl,
  ansicht: ansichtwahl,
  formularZiel,
  stand = 0,
  vorigesBild,
  bildGezeigt,
}: KarteneditorEigenschaften) {
  const texte = useTexte()
  const vorgang = useVorgang()
  const [fehler, setzeFehler] = useState<Feldfehler>()

  // **Die `auto replace:`-Regeln des Spiels** (#897): `->` wird zu `→`, `(C)`
  // zu `©` — beim Tippen und nur dann.
  const autoErsetzen = useAutoersetzung(setID)

  // **Das Formular hält seine Felder**, damit ein Klick in die Vorschau eines
  // davon fokussieren kann.
  const formular = useRef<HTMLFormElement>(null)

  // **Was gerade getippt wird** — Feld und Wert, noch nicht geschrieben.
  //
  // Solange etwas darin steht, zeigt die Vorschau **diesen** Wert. Beim
  // Verlassen des Feldes wird geschrieben, der Eintrag geräumt und die
  // gewöhnliche Vorschau nachgeladen: ab da steht wieder im Bild, was im
  // Bestand steht.
  const [getippt, setzeGetippt] = useState<{ feld: string; wert: string }>()

  // **Die Größe der Vorschau** (U5) sagt, wie groß gerastert wird — und das
  // Rastern hängt an diesem Editor; gehalten wird sie darüber (UXR-06).
  const ansicht = useVorschauansicht(ansichtwahl)
  const [bildstand, setzeBildstand] = useState(0) // A12, #1451

  // **Die Palette gehört zu einem Feld, nicht zur Karte.** Welche Symbolschrift
  // gilt, steht am Feldstil; sie für alle 24 Felder zugleich zu holen wären 24
  // Läufe durch die Skriptumgebung, von denen 22 leer zurückkämen.
  //
  // **Gesetzt beim Fokus und nicht beim Verlassen gelöscht.** Ein Klick auf
  // eine Palettenschaltfläche nimmt dem Feld den Fokus; wer beim Verlassen
  // löschte, ließe die Palette unter der Hand verschwinden.
  //
  // **Und es kann eine Ebene höher liegen** (#1075): dieser Editor wird bei
  // jedem Kartenwechsel neu gebaut, und eine Wahl, die das überstehen soll,
  // gehört über ihn. Warum, steht in `Karten.tsx` an `feldwahl`.
  const eigene = useState<string>()
  const [gewaehlt, setzeGewaehlt] = feldwahl ?? eigene

  // **Was der Editor zeigt, steht in einem Haken** (§2): die geladene
  // Vorschau, die Probe beim Tippen und die Regel, welche gilt.
  const vorschau = useKartenvorschau(
    karte.id,
    getippt,
    vergroesserungFuer(ansicht[0].zoom),
    stand,
    vorigesBild,
  )
  const { gezeigt, vorgaben } = vorschau

  // **Das gezeigte Bild geht nach oben**, über den Schlüssel dieses Editors —
  // die nächste Karte zeigt es, bis ihr eigenes da ist (AP-4, #2361).
  useEffect(() => {
    if (gezeigt !== undefined) {
      bildGezeigt?.({ karteID: karte.id, bild: gezeigt })
    }
  }, [gezeigt, karte.id, bildGezeigt])

  const palette = useSymbolpalette(karte.id, gewaehlt)

  // **Die Auswahllisten hängen am geschriebenen Wert** (#1074): welcher
  // Eintrag angekreuzt ist, steht im Feld — nicht im Entwurf.
  const feldwert = karte.werte.find((wert) => wert.feld === gewaehlt)?.wert
  const wortlisten = useWortlisten(karte.id, gewaehlt, feldwert ?? '')

  // nachName findet die Beschreibung eines Feldes.
  //
  // **`undefined` ist ein Fall, kein Mangel**: ein Set kann Felder führen, die
  // sein Spiel nicht erklärt. [Feldeingabe] macht daraus ein Textfeld.
  const nachName = new Map(formen.map((form) => [form.name, form]))

  // **Der eigene Stil der Karte** steht als Wert `stylesheet` — leer heißt
  // „der des Sets" (`service.Kartenstilfeld`).
  const eigenerStil =
    karte.werte.find((wert) => wert.feld === 'stylesheet')?.wert.trim() ?? ''
  const ordnung = useFeldordnung(ansicht, gezeigt, vorschau.laedt, eigenerStil)
  const felder = angeboteneFelder(
    karte.werte,
    formen,
    ordnung.flaechen,
    ordnung.reihenfolge,
  )
  const kartenname = kartennameAus(karte.werte)

  // istBildfeld sagt es über die Feldbeschreibung und nicht über den Namen —
  // [bildfelderAus] beantwortet dieselbe Frage für die ganze Vorlage.
  const istBildfeld = (feld: string) => nachName.get(feld)?.art === 'image'

  const waehleFeld = (feld: string) => {
    feldFokussieren(formular.current, feld)
  }

  // einfuegen setzt ein Symbol an der Schreibmarke ein; den neuen Wert
  // schreibt der Entwurf des Feldes (`Feldzeile`), im Konflikt gar nicht.
  const einfuegen = (
    feld: string,
    wertRoh: string,
    code: string,
    roh: boolean,
  ) => symbolEinfuegenIn(formular.current, feld, wertRoh, code, roh)

  const {
    bildSetzen,
    vonScryfall,
    vonKI,
    merkeFeld,
    zuschnitt,
    einsetzen: bildEinsetzen,
    abbrechen: zuschnittAbbrechen,
  } = useBildfeld({
    setID,
    karteID: karte.id,
    vorgang,
    melden: setzeFehler,
    geglueckt: () => {
      setzeGetippt(undefined)
      setzeBildstand((bisher) => bisher + 1)
      vorschau.nachladen()
      geaendert()
    },
  })

  // **Der Feldabgleich** (#1013). Er schreibt über dieselbe Naht wie das
  // Tippen; nach einer geglückten Übernahme lädt der Editor nach, damit die
  // neuen Werte im Formular stehen.
  const abgleich = useScryfallabgleich({
    setID,
    karteID: karte.id,
    vorgang,
    texte,
    geglueckt: () => {
      setzeGetippt(undefined)
      vorschau.nachladen()
      geaendert()
    },
  })

  // **Dieselben Wege für beide Schreibarten** (#1074): ein getippter Wert und
  // eine Wahl gehen in dasselbe Feld, über denselben Vorgangsschlüssel — sonst
  // liefen sie gegen einen Stand, den das Backend noch nicht kennt (§7).
  const wege = {
    setID,
    karteID: karte.id,
    vorgang,
    texte,
    melden: setzeFehler,
    geglueckt: () => {
      setzeFehler(undefined)
      setzeGetippt(undefined)
      vorschau.nachladen()
      // **Und die Karte nachgeladen** (#1074): eine Wortwahl ändert das Feld,
      // ohne dass jemand hineingetippt hätte.
      geaendert()
    },
  }

  const tiefe = useFeldtiefe()
  const schreiben = useFeldschreiben(wege)
  const wortWaehlen = useWortwahl(wege)

  const probleme = problemeAus(fehler, gezeigt?.befunde ?? [])

  const formularteil = (
    <>
      <Problemliste probleme={probleme} zumFeld={waehleFeld} />
      {/* **Der Behälter misst sich selbst** (#697). Eine Bildschirmbreite wie
          `xl:grid-cols-2` misst das Fenster und nicht diese Spalte —
          gemessen an der laufenden Anwendung bekam das Formular bei 1.440
          Punkten Fenster 375 und teilte sie in **zwei Spalten à 176**. Eine
          Container-Anfrage misst, was da ist. */}
      <div className="@container flex min-w-64 flex-1">
        {/* **Wie die Felder stehen, entscheidet `Feldformular`** (#1271,
            UX-05). Hier steht nur, **was** eine Zeile braucht. */}
        <Feldformular
          felder={felder}
          unruhig={ordnung.unruhig}
          nachName={nachName}
          {...tiefe}
          probleme={probleme}
          formular={formular}
          zeile={(wert) => (
            <Feldzeile
              key={wert.feld}
              feld={wert.feld}
              wert={wert.wert}
              form={nachName.get(wert.feld)}
              fehler={fehler?.feld === wert.feld ? fehler.meldung : undefined}
              laeuft={vorgang.laeuft(`${karte.id}/${wert.feld}`)}
              palette={gewaehlt === wert.feld ? palette.daten : undefined}
              wortlisten={
                gewaehlt === wert.feld ? (wortlisten.daten ?? []) : []
              }
              wortWaehlen={(bereich, weg, bekannt) => {
                void wortWaehlen(wert.feld, bereich, weg, bekannt)
              }}
              hergeleitet={vorgaben.get(wert.feld)}
              vorschlag={feldvorschlag(karte, wert.feld, nachName, geaendert)}
              autoErsetzen={autoErsetzen}
              gewaehlt={() => {
                setzeGewaehlt(wert.feld)
              }}
              getippt={(neuerWert) => {
                setzeGetippt(
                  neuerWert === undefined
                    ? undefined
                    : { feld: wert.feld, wert: neuerWert },
                )
              }}
              schreiben={(neuerWert, bekannt) =>
                schreiben(wert.feld, neuerWert, bekannt)
              }
              einfuegen={(code, roh) =>
                einfuegen(wert.feld, wert.wert, code, roh)
              }
              bildWaehlen={() => {
                void bildSetzen(wert.feld)
              }}
              symbolZeichnen={
                symbolZeichnen === undefined
                  ? undefined
                  : () => {
                      symbolZeichnen(wert.feld)
                    }
              }
            />
          )}
        />
      </div>

      <Stilfelder key={karte.id} wege={wege} stand={stand} fehler={fehler} />

      {/* **Vorschlag, Bild und Herkunft stehen unter dem Formular** (A7,
          #1446; A11, #1450; A12, #1451), in einem Kasten (UXR-21, #1929) —
          und der `key` räumt sie beim Kartenwechsel mit ab: §7, Regel 1 und
          2 in einer Zeile. Wer ihn entfernt, nimmt die unsichtbare mit weg. */}
      <KIKasten>
        <KIBereich
          key={karte.id}
          karteID={karte.id}
          setID={setID}
          bildfelder={bildfelderAus(formen)}
          zuschneiden={vonKI}
          // **Beide Stände steigen nur**, also steigt ihre Summe mit jedem.
          bildstand={bildstand + stand}
          geaendert={geaendert}
        />
      </KIKasten>
    </>
  )

  const vorschauteil = (
    <Kartenvorschau
      karteID={karte.id}
      vorschau={vorschau}
      ansicht={ansicht}
      imZentrum={formularZiel !== undefined}
      gezeigt={gezeigt}
      zuschnitt={zuschnitt}
      laeuft={
        zuschnitt !== undefined &&
        vorgang.laeuft(`${karte.id}/${zuschnitt.feld}`)
      }
      istBildfeld={istBildfeld}
      bildWaehlen={(feld) => {
        void bildSetzen(feld)
      }}
      gewaehltesBildfeld={
        gewaehlt !== undefined && istBildfeld(gewaehlt) ? gewaehlt : undefined
      }
      abgleich={abgleich}
      kartenname={kartenname}
      manakosten={
        karte.werte.find((wert) => wert.feld === 'casting cost')?.wert ?? ''
      }
      vonScryfall={(feld, name, art) => {
        void vonScryfall(feld, name, art)
      }}
      feldWaehlen={waehleFeld}
      merkeFeld={merkeFeld}
      uebernehmen={(auswahl) => {
        void bildEinsetzen(auswahl)
      }}
      abbrechen={zuschnittAbbrechen}
    />
  )

  // **Ohne Arbeitsplatz nebeneinander, mit ihm verteilt**: das Formular geht in
  // den Inspector, die Vorschau bleibt hier — im Zentrum. **`gap-6` und nicht
  // `gap-8`** (#1271): 32 Punkte Abstand waren bei 1280 der Unterschied
  // zwischen nebeneinander und untereinander; der Grund für die Rechnung steht
  // bei [Kartenvorschau].
  if (formularZiel === undefined) {
    return (
      <div className="flex flex-wrap gap-6">
        {formularteil}
        {vorschauteil}
      </div>
    )
  }

  return (
    <>
      {formularZiel !== null && createPortal(formularteil, formularZiel)}
      {vorschauteil}
    </>
  )
}
