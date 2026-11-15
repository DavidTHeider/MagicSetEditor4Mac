// **Die rechte Spalte der Kartenlasche** — die Handlungen der Karte und ihr
// Formular.
//
// # Warum eine eigene Datei
//
// Weil `Karten.tsx` mit ihr über die Zeilengrenze ging (§2), als die
// Handlungen aus der Kopfzeile hierher wanderten
// (#1240).
// Der Schnitt folgt der Frage: dort steht, **wie** die Lasche angeordnet ist,
// hier, **was in der Kartenspalte** steht.

import { useState } from 'react'

import type { KarteDTO, FeldbeschreibungDTO, KartenzeileDTO } from '../api'
import type { Bildstand } from '../haken/useKartenvorschau'
import type { Kartenablagegriff } from '../haken/useKartenablage'
import type { SuchtrefferDTO } from '../apiscryfall'
import type { Vorgang } from '../haken/useVorgang'
import { KIKastenraum } from '../bausteine/KIKasten'
import { Leerzustand } from '../bausteine/Leerzustand'
import { useTexte } from '../sprache/kontext'
import { Karteneditor, type Feldwahl } from './Karteneditor'
import { Kartenhandlungen } from './Kartenhandlungen'
import type { Ansichtswahl } from './vorschauansicht'

interface FlaechenEigenschaften {
  setID: string

  // inBearbeitung ist die aufgeschlagene Karte — oder keine.
  inBearbeitung: KartenzeileDTO | undefined

  karten: readonly KartenzeileDTO[]
  geoeffnete: KarteDTO | undefined
  formen: FeldbeschreibungDTO[]
  // symbolZeichnen führt zum Symboleditor — der Grund steht bei
  // [KartenEigenschaften].
  symbolZeichnen?: ((karteID: string, feld: string) => void) | undefined

  verlaufsstand: number

  vorgang: Vorgang
  ablage: Kartenablagegriff
  vorschlag: SuchtrefferDTO | undefined

  // feldwahl ist das gewählte Feld samt seinem Setzer — die Gestalt, die
  // `useState` gibt.
  feldwahl: Feldwahl

  // ansicht sind Maßstab und Überlagerungen der Vorschau — aus demselben
  // Grund über dem Editor wie `feldwahl` (UXR-06, `vorschauansicht.ts`).
  ansicht: Ansichtswahl

  aufFehler: (meldung: string | undefined) => void
  waehlen: (karteID: string) => void
  anlegen: () => void
  nachruecken: (geloescht: string) => void
  nachladen: () => void
  fehlerRaeumen: () => void
  bestandGeaendert: () => void
  geaendert: () => void

  // leereKarte macht aus einer Listenzeile eine Karte ohne Werte — der Stand,
  // solange die Werte noch unterwegs sind.
  leereKarte: (zeile: KartenzeileDTO) => KarteDTO

  // formularZiel ist der Inspector des Arbeitsplatzes (U5, #1416) — siehe
  // `Karteneditor`.
  formularZiel?: HTMLElement | null | undefined
}

export function Kartenarbeitsflaeche({
  setID,
  inBearbeitung,
  karten,
  geoeffnete,
  formen,
  symbolZeichnen,
  verlaufsstand,
  vorgang,
  ablage,
  vorschlag,
  feldwahl,
  ansicht,
  aufFehler,
  waehlen,
  anlegen,
  nachruecken,
  nachladen,
  fehlerRaeumen,
  bestandGeaendert,
  geaendert,
  leereKarte,
  formularZiel,
}: FlaechenEigenschaften) {
  const texte = useTexte()
  const [vorigesBild, setzeVorigesBild] = useState<Bildstand | undefined>(
    undefined,
  )

  return (
    <section
      aria-label={texte.seiten.karten.bearbeiten}
      className="flex min-w-0 flex-col gap-3"
    >
      {inBearbeitung ? (
        // **Der KI-Kasten bleibt offen, wenn die Karte wechselt** (UXR-21,
        // #1929) — sein Raum steht über dem Schlüssel des Editors.
        <KIKastenraum>
          {/* **Die Handlungen der Karte stehen bei der Karte** (#1240) —
              nicht in der Setzeile darüber. Seit UXR-06 (#1914) in **einer**
              Zeile mit dem Blättern und dem Speicherstand; was das Menü
              aufschlägt, steht darunter. */}
          <Kartenhandlungen
            setID={setID}
            karte={inBearbeitung}
            karten={karten}
            vorgang={vorgang}
            ablage={ablage}
            vorschlag={vorschlag}
            aufFehler={aufFehler}
            waehlen={waehlen}
            anlegen={anlegen}
            nachruecken={nachruecken}
            nachladen={nachladen}
            fehlerRaeumen={fehlerRaeumen}
            bestandGeaendert={bestandGeaendert}
          />
          {/* **Der Stand gehört nicht in den Schlüssel** (U01, #2119): er
              baute den Editor bei jeder Änderung irgendwo neu auf, und wer
              gerade tippte, verlor Entwurf und Fokus. Eine Rücknahme
              erreicht das Feld trotzdem — `Feldzeile` trägt den
              gespeicherten Wert im Schlüssel seiner Eingabe. */}
          <Karteneditor
            setID={setID}
            // **Das gewählte Feld liegt außerhalb des Schlüssels** (#1075):
            // der Neuaufbau ist gewollt, der Verlust der Wahl war es nicht.
            feldwahl={feldwahl}
            ansicht={ansicht}
            key={inBearbeitung.id}
            // **Das vorige Bild steht über dem Schlüssel** (AP-4, #2361),
            // wie `feldwahl`: so zeigt die nächste Karte es, bis ihr
            // eigenes da ist, statt leer zu blitzen.
            vorigesBild={vorigesBild}
            bildGezeigt={setzeVorigesBild}
            stand={verlaufsstand}
            karte={geoeffnete ?? leereKarte(inBearbeitung)}
            formen={formen}
            geaendert={geaendert}
            formularZiel={formularZiel}
            symbolZeichnen={
              symbolZeichnen === undefined
                ? undefined
                : (feld) => {
                    symbolZeichnen(inBearbeitung.id, feld)
                  }
            }
          />
        </KIKastenraum>
      ) : (
        <Leerzustand
          titel={texte.seiten.karten.keineKarteTitel}
          erklaerung={texte.seiten.karten.keineKarte}
        />
      )}
    </section>
  )
}
