// Die Handlungen an der aufgeschlagenen Karte — eine Zeile über der Karte,
// ein Menü und, was das Menü aufschlägt (Werkstück U2 von Phase 5.1,
// #1413;
// UXR-06, #1914):
//
//   - **Eine Zeile**: vorige und nächste Karte samt Stelle, daneben das Menü
//     „Diese Karte". So zeichnet sie das Artboard *Karten*.
//   - **Ein Menü** für alles, was an dieser Karte geschieht — Duplizieren, die
//     Zwischenablage, Ordnen, den Bezug zu einem Druck, Teilen, Löschen. Eine
//     Tabstelle statt neun, die Pfeile wandern, Escape schließt, der Fokus
//     kehrt zurück (`Menue`).
//   - **Was mehr ist als ein Klick, steht darunter**: die Rückfrage vor dem
//     Löschen (`Entscheidung`), die vorbereitete Teilung und der Kasten mit
//     dem Bezug.
//
// # Warum die Zwischenablage, der Bezug und das Teilen im Menü stehen
//
// Bis UXR-06 standen „Teilen …" als Knopf neben dem Menü und „Ablage und
// Bezug" als Aufklapper darunter. Gemessen am nativen Fenster trug die
// Kartenseite damit 27 Bedienelemente im Bild, R2 aus `docs/UIREVIEW.md`
// erlaubt zwanzig — und alle drei sind seltener als das Blättern. **Keine
// Fähigkeit geht verloren**: jede liegt eine Ebene tiefer, ⌘C, ⌘X und ⌘V
// wirken in der Liste weiter, und „Kopieren" kündigt sein Kürzel im Menü an.
//
// # Die Regeln aus §7
//
// Die vier Regeln aus §7, an jedem Rückruf im Kommentar. **Regel 3 zuerst**:
// nach einem geglückten Schreiben wird die Wahl gesetzt, **bevor**
// nachgeladen wird; scheitert das Nachladen, steht die Wahl trotzdem richtig.
import { useRef, useState } from 'react'

import type { KartenzeileDTO } from '../api'
import type { SuchtrefferDTO } from '../apiscryfall'
import type { Kartenablagegriff } from '../haken/useKartenablage'
import { useKartenkopieren } from '../haken/useKartenkopieren'
import { useKartenordnen } from '../haken/useKartenordnen'
import type { Vorgang } from '../haken/useVorgang'
import { Menue, type Menueeintrag } from '../bausteine/Menue'
import { useHandlungsliste } from '../haken/useHandlungsliste'
import { Speicherstand } from '../bausteine/Speicherstand'
import { Teilen } from '../bausteine/Teilen'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Kartenblaettern } from './Kartenblaettern'
import { nachbarkarte } from './kartennachbar'
import { Kartenloeschen } from './Kartenloeschen'
import { Kartenverweis } from './Kartenverweis'
import { Werkzeugkasten } from './Werkzeugkasten'

interface HandlungsEigenschaften {
  setID: string
  karte: KartenzeileDTO
  karten: readonly KartenzeileDTO[]
  vorgang: Vorgang
  ablage: Kartenablagegriff
  vorschlag: SuchtrefferDTO | undefined
  aufFehler: (meldung: string | undefined) => void
  waehlen: (karteID: string) => void

  // anlegen legt eine neue Karte an — derselbe Weg wie „+ Karte anlegen“
  // über der Liste; im Menü mit ⌘↩ (AP-7, #2364).
  anlegen: () => void
  nachruecken: (geloescht: string) => void
  nachladen: () => void
  fehlerRaeumen: () => void
  bestandGeaendert: () => void
}

// eintrag baut eine Menüzeile — ruhend ohne Handlung, wenn es nichts zu tun
// gibt (am Rand der Liste), ruhend mit Handlung, solange der Vorgang läuft.
//
// **Die Zeile kommt als Verbund**, damit ihre Aktion neben dem Namen steht —
// dort, wo `aktionskatalog.test.ts` sie liest.
function eintrag(
  zeile: Pick<
    Menueeintrag,
    'schluessel' | 'name' | 'aktion' | 'mitKuerzel' | 'schreibt' | 'kuerzel'
  >,
  handlung: (() => void) | undefined,
  laeuft: boolean,
): Menueeintrag {
  return handlung === undefined
    ? { ...zeile, angehalten: true }
    : { ...zeile, aufWahl: handlung, angehalten: laeuft }
}

export function Kartenhandlungen({
  setID,
  karte,
  karten,
  vorgang,
  ablage,
  vorschlag,
  aufFehler,
  waehlen,
  anlegen,
  nachruecken,
  nachladen,
  fehlerRaeumen,
  bestandGeaendert,
}: HandlungsEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten

  const [loeschenOffen, setzeLoeschenOffen] = useState(false)
  const [bezugOffen, setzeBezugOffen] = useState(false)

  // **Der Anlass zum Teilen** — eine Zahl, die das Menü hochzählt; der
  // Baustein bereitet bei jeder neuen vor (`Teilen`, `anlass`). Sie wird beim
  // Kartenwechsel **nicht** geräumt: der `key` baut den Baustein neu, und der
  // nimmt den Stand beim Aufbau als schon gesehen.
  const [teilanlass, setzeTeilanlass] = useState(0)

  const zeile = useRef<HTMLDivElement>(null)

  // **Ein Wechsel der Karte räumt die Frage und den Kasten mit** (§7, Regel
  // 1). Ohne das stünde die Rückfrage nach dem Kartenwechsel weiter da und
  // löschte beim nächsten Klick die **neue** Karte.
  const [zuletzt, setzeZuletzt] = useState(karte.id)
  if (zuletzt !== karte.id) {
    setzeZuletzt(karte.id)
    setzeLoeschenOffen(false)
    setzeBezugOffen(false)
  }

  const kopie = useKartenkopieren(setID, karte.id, vorgang, (neue) => {
    // **Regel 3**: geschrieben ist geschrieben. Die Kopie wird
    // aufgeschlagen, **bevor** nachgeladen wird; scheitert das Nachladen,
    // steht die Wahl trotzdem richtig.
    waehlen(neue)
    fehlerRaeumen()
    nachladen()
    bestandGeaendert()
  })

  const ordnung = useKartenordnen(setID, karte.id, karten, vorgang, () => {
    // **Regel 3**: die neue Ordnung steht; scheitert das Nachladen, bleibt
    // sie trotzdem.
    fehlerRaeumen()
    nachladen()
    bestandGeaendert()
  })

  // **Die Zwischenablage zählt, was sie trifft** (#1168): angehakt, sonst die
  // aufgeschlagene — eine Karte ist hier immer offen, die Zahl nie null.
  const zahl = String(ablage.anzahl)

  // **Was ins Set schreibt, trägt `schreibt`** (#2288): ein Leser sieht es
  // gesperrt und erklärt. Kopieren und Teilen lesen nur.
  //
  // **Und die Kürzel aus dem Entwurf `Kontext09`** (AP-7, #2364): ⌘↩, ⌘D,
  // ⌘⌫, ⇞ und ⇟. Die Leiste verdrahtet sie; was ein Textfeld mit ihnen tut,
  // misst `ap7kuerzel`.
  const vorige = nachbarkarte(karten, karte.id, 'zurueck')
  const naechste = nachbarkarte(karten, karte.id, 'vor')
  const eintraege: Menueeintrag[] = [
    eintrag(
      {
        schluessel: 'anlegen',
        name: eigene.anlegen,
        aktion: 'anlegen',
        schreibt: true,
        kuerzel: 'befehl+return',
      },
      anlegen,
      vorgang.laeuft('anlegen'),
    ),
    eintrag(
      {
        schluessel: 'duplizieren',
        name: eigene.duplizieren,
        aktion: 'duplizieren',
        schreibt: true,
        kuerzel: 'befehl+d',
      },
      kopie.kopieren,
      kopie.laeuft,
    ),
    eintrag(
      {
        schluessel: 'kopieren',
        name: einsetzen(eigene.ablage.kopieren, { zahl }),
        aktion: 'kopieren',
        mitKuerzel: true,
      },
      ablage.kopieren,
      ablage.laeuft,
    ),
    eintrag(
      {
        schluessel: 'ausschneiden',
        name: einsetzen(eigene.ablage.ausschneiden, { zahl }),
        schreibt: true,
      },
      ablage.ausschneiden,
      ablage.laeuft,
    ),
    // **Einfügen hängt an keiner Auswahl.** Was hineinkommt, steht in der
    // Zwischenablage; ob etwas darin ist, weiß erst der Aufruf — über die
    // Wails-Naht steht nur Text zur Verfügung, MSE2 fragt dafür das Format
    // ab (`adapted`, `CardsDataObject::format`).
    eintrag(
      {
        schluessel: 'einfuegen',
        name: eigene.ablage.einfuegen,
        schreibt: true,
      },
      ablage.einfuegen,
      ablage.laeuft,
    ),
    // **Blättern wie in MSE2**: *Select Previous/Next Card* auf ⇞ und ⇟, am
    // Rand abgeschaltet (`cards_panel.cpp`, Zeile 239, Fassung f9d9356 vom
    // 2021-07-06; dieselbe Regel wie `Kartenblaettern`).
    eintrag(
      {
        schluessel: 'vorige',
        name: eigene.blaettern.vorige,
        kuerzel: 'bild auf',
      },
      vorige === undefined
        ? undefined
        : () => {
            waehlen(vorige)
          },
      false,
    ),
    eintrag(
      {
        schluessel: 'naechste',
        name: eigene.blaettern.naechste,
        kuerzel: 'bild ab',
      },
      naechste === undefined
        ? undefined
        : () => {
            waehlen(naechste)
          },
      false,
    ),
    eintrag(
      { schluessel: 'nachOben', name: eigene.nachOben, schreibt: true },
      ordnung.nachOben,
      ordnung.laeuft,
    ),
    eintrag(
      { schluessel: 'nachUnten', name: eigene.nachUnten, schreibt: true },
      ordnung.nachUnten,
      ordnung.laeuft,
    ),
    eintrag(
      { schluessel: 'bezug', name: eigene.bezugAufklapp, schreibt: true },
      () => {
        setzeBezugOffen(true)
      },
      false,
    ),
    // **Das Bild dieser Karte teilen** — NET-04 (#1708).
    eintrag(
      { schluessel: 'teilen', name: texte.bausteine.teilen.knopf },
      () => {
        setzeTeilanlass((bisher) => bisher + 1)
      },
      false,
    ),
    eintrag(
      {
        schluessel: 'loeschen',
        name: eigene.loeschen,
        aktion: 'loeschen',
        schreibt: true,
        kuerzel: 'befehl+backspace',
      },
      () => {
        setzeLoeschenOffen(true)
      },
      false,
    ),
  ]

  // **Dieselbe Liste an drei Orten** (GESTE, #2358): das Menü „Diese
  // Karte" hier, der Rechtsklick auf das Kartenbild (`Kartenvorschau`) und
  // das Menü „Karte" in der Menüleiste.
  const handlungen = useHandlungsliste({
    titel: eigene.kartenmenue,
    art: 'karte',
    kennung: karte.id,
    eintraege,
  })

  return (
    <section aria-label={eigene.dieseKarte} className="flex flex-col gap-2">
      {/* **Vorige und nächste Karte stehen ganz vorn** (#1290): der Weg zur
          nächsten Karte ist in Serienarbeit der häufigste; er darf nicht
          hinter einem Aufklapp liegen und nicht unter der Falz. **Das Menü
          steht am rechten Rand** und klappt nach links auf: gleich wie breit
          „107 von 107" ist, ragt es nicht über die Spalte. */}
      <div ref={zeile} className="flex flex-wrap items-center gap-x-4 gap-y-2">
        <Kartenblaettern
          karten={karten}
          gewaehlt={karte.id}
          waehlen={waehlen}
        />

        <div className="ml-auto flex flex-wrap items-center gap-x-4 gap-y-2">
          {/* **Ob das Eingetippte im Bestand steht** (UX-05.AK-03, #1677):
              geschrieben wird beim Verlassen eines Feldes, und bis hierher
              sagte nur ein Fehler etwas darüber. **In der Zeile und nicht
              darunter** (UXR-06): der Satz kommt und geht mit jedem Feld,
              und eine eigene Zeile schob die Karte bei jedem Schreiben
              hinab. */}
          <Speicherstand setID={setID} />
          <Menue
            bezeichnung={eigene.dieseKarte}
            eintraege={handlungen}
            ausrichtung="rechts"
          />
        </div>
      </div>

      {kopie.fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {kopie.fehler}
        </p>
      )}

      {ordnung.fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {ordnung.fehler}
        </p>
      )}

      {loeschenOffen && (
        <Kartenloeschen
          setID={setID}
          karteID={karte.id}
          vorgang={vorgang}
          schliessen={() => {
            setzeLoeschenOffen(false)
          }}
          aufGeloescht={(geloescht) => {
            // **Regel 1 aus §7**: die gelöschte Karte war der offene
            // Kontext. Wer die Wahl stehen ließe, zeigte einen Editor auf
            // eine Karte, die es nicht mehr gibt — also rückt die nächste
            // nach, wie MSE2 es tut.
            //
            // **Regel 2 daneben**: wer während des Löschens eine andere
            // Karte aufschlägt, hat einen neuen offenen Kontext; über den
            // Rückgriff entscheidet [Karten] und nicht dieser Aufruf.
            nachruecken(geloescht)
            fehlerRaeumen()

            // **Regel 3**: das Nachladen ist Bestenfalls-Arbeit; scheitert
            // es, ist die Karte trotzdem weg und die Wahl geräumt.
            nachladen()
            bestandGeaendert()
          }}
        />
      )}

      {/* **Der `key` räumt die Teilung beim Kartenwechsel** — §7, Regeln 1
          und 2 in einer Gestalt (`docs/SCHREIBWEGE.md`): wer blättert, teilt
          nicht die Kopie der vorigen Karte. */}
      <Teilen
        key={karte.id}
        art="kartenbild"
        bezug={karte.id}
        anlass={teilanlass}
      />

      {bezugOffen && (
        <Werkzeugkasten
          titel={eigene.bezugAufklapp}
          offen
          schliessen={() => {
            setzeBezugOffen(false)
            zeile.current
              ?.querySelector<HTMLElement>('[aria-haspopup="menu"]')
              ?.focus()
          }}
        >
          {/* **Der Verweis auf einen echten Druck** (#1167). Ohne ihn holt
              jeder Abgleich irgendeinen Druck.

              **Der `key` räumt ihn beim Kartenwechsel** — §7, Regeln 1 und 2
              in einer Gestalt (`docs/SCHREIBWEGE.md`). */}
          <Kartenverweis
            key={karte.id}
            karteID={karte.id}
            vorschlag={vorschlag}
            vorgang={vorgang}
            aufFehler={aufFehler}
          />
        </Werkzeugkasten>
      )}
    </section>
  )
}
