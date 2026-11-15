// **Was das Set betrifft** — die Kopfzeile der Kartenlasche.
//
// Eigene Datei, weil `Karten.tsx` mit ihr über 400 Zeilen ginge (§2). Der
// Schnitt folgt einer Frage: dort steht, **wie** die Lasche angeordnet ist,
// hier, **was sich am Set** tun lässt.
//
// # Was hier nicht mehr steht
//
// Bis zum 2026-08-27 lag hier **alles**: der Hauptknopf, zwei Scryfall-Felder
// samt Knöpfen, der Listenschalter, drei Ablageknöpfe, der Verweiskasten,
// Duplizieren, Ordnen und Löschen — in **einem** Umbruchkasten. Bei 1280
// Punkten wurden daraus vier Reihen, und ein Suchfeld stand zwei Zeilen über
// seinem Knopf (#1240).
//
// **Die Handlungen einer Karte stehen jetzt bei der Karte**
// (`Kartenhandlungen.tsx`).
//
// # Eine Handlung und ein Menü
//
// **Seit UXR-05** (#1913)
// stehen hier zwei Bedienelemente: „Karte anlegen" und „Mehr". Davor waren
// es mit dem Spielpaket `magic` fünf gleichrangige — Scryfall, Skripte,
// Umschlag und Autoersetzung je mit eigenem Knopf neben dem Anlegen —, und
// das Artboard *Karten* zeigt keinen davon. Benutzt wird jeder einmal je Set
// oder seltener, das Anlegen bei jeder Karte.
//
// **Näher an MSE2 als vorher**: dort hängen die Skripte als Aufklappmenü an
// „Karte hinzufügen" (`src/gui/set/cards_panel.cpp`, Zeilen 160–176, Fassung
// f9d9356 vom 2021-07-06).
//
// **Warum über dem Arbeitsplatz und nicht unten im Navigator**, wo das
// Artboard „+ Karte anlegen" zeichnet: unter 1.072 Punkten klappt der
// Navigator von selbst ein (`engeGrenze` in `bausteine/arbeitsplatzmasse.ts`),
// und die führende Handlung der Fläche verschwände mit ihm.
import { useRef, useState, type ReactNode } from 'react'

import type { SuchtrefferDTO } from '../apiscryfall'
import type { Vorgang } from '../haken/useVorgang'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Menue, type Menueeintrag } from '../bausteine/Menue'
import { useHandlungsliste } from '../haken/useHandlungsliste'
import { useKartenumschlag } from '../haken/useKartenumschlag'
import type { Kartenskriptgriff } from '../haken/useKartenskripte'
import { useTexte } from '../sprache/kontext'
import { Autoersetzungen } from './Autoersetzungen'
import { Kartenskripte } from './Kartenskripte'
import { Scryfallanlage } from './Scryfallanlage'
import { Werkzeugkasten } from './Werkzeugkasten'

interface WerkzeugEigenschaften {
  // setID trägt die Autoersetzung (#1341): sie gilt je **Spiel**, und welches
  // das ist, sagt das Set.
  setID: string

  // kartenskripte sind die fertigen Züge des Spielpakets (#1337) — sie stehen
  // hier, weil sie dasselbe tun wie das Anlegen daneben: eine Karte entsteht.
  kartenskripte: Kartenskriptgriff

  vorgang: Vorgang

  anlegen: () => void

  // bestandGeaendert läuft, wenn eine Karte hereingekommen ist (#1639) —
  // die Liste daneben zeigte sonst den Stand von vorhin.
  bestandGeaendert: () => void

  // ausScryfall legt eine Karte aus einem Scryfall-Namen an (#1032).
  ausScryfall: (name: string, art: string) => void
  anlegefehler: string | undefined

  // gewaehlt ist der Name des Scryfall-Treffers — er steht im Anlegekasten.
  gewaehlt: string
  waehlenTreffer: (treffer: SuchtrefferDTO | undefined) => void

  // ansichtsort nimmt den Platz für die Schalter des Arbeitsplatzes (P03) —
  // ohne ihn, etwa in einem leeren Set, gibt es keinen.
  ansichtsort?: ((ort: HTMLDivElement | null) => void) | undefined
}

// Kasten ist, was das Menü aufschlägt — der Umschlag schlägt nichts auf, er
// läuft gleich.
type Kasten = 'scryfall' | 'skripte' | 'autoersetzung'

export function Kartenwerkzeuge({
  setID,
  kartenskripte,
  vorgang,
  anlegen,
  bestandGeaendert,
  ausScryfall,
  anlegefehler,
  gewaehlt,
  waehlenTreffer,
  ansichtsort,
}: WerkzeugEigenschaften) {
  const eigene = useTexte().seiten.karten

  const [offen, setzeOffen] = useState<Kasten>()

  // **Zu heißt verborgen, nicht fort** — die Regel des Aufklappens aus
  // `docs/UX-ARCHITEKTUR.md` (S11.NAT-03). Ein Kasten, der einmal offen war,
  // bleibt gezeichnet; wer zwischen Scryfall und Autoersetzung wechselt,
  // findet den getippten Namen wieder.
  const [besucht, setzeBesucht] = useState<readonly Kasten[]>([])

  const oeffne = (kasten: Kasten) => {
    setzeOffen(kasten)
    setzeBesucht((bisher) =>
      bisher.includes(kasten) ? bisher : [...bisher, kasten],
    )
  }

  const umschlag = useKartenumschlag(setID, bestandGeaendert)

  // **Anlegen ist eine Handlung am Set** — Befund 2 des Zweitblicks über
  // Phase 6 von 0.9.0 (AP-7, #2364): ⌘↩ stand nur an den Handlungen einer
  // offenen Karte, und ein leeres Set hat keine. Der Kopf meldet „Karte
  // anlegen“ deshalb selbst; ist eine Karte offen, meldet sie sich danach
  // an und hat die Leiste (`objekthandlungsraum.tsx`: das zuletzt
  // angemeldete Objekt gilt).
  useHandlungsliste({
    titel: eigene.kartenmenue,
    art: 'karte',
    kennung: '',
    eintraege: [
      {
        schluessel: 'anlegen',
        name: eigene.anlegen,
        aktion: 'anlegen',
        schreibt: true,
        kuerzel: 'befehl+return',
        angehalten: vorgang.laeuft('anlegen'),
        aufWahl: anlegen,
      },
    ],
  })

  // **Wer einen Kasten schließt, landet wieder am Menü** — der Knopf, der
  // ihn schloss, steht danach nicht mehr im Bild, und der Fokus fiele sonst
  // an den Anfang des Fensters.
  const zeile = useRef<HTMLDivElement>(null)
  const zumMenue = () => {
    zeile.current?.querySelector<HTMLElement>('[aria-haspopup="menu"]')?.focus()
  }

  const titel: Record<Kasten, string> = {
    scryfall: eigene.ausScryfallAufklapp,
    skripte: eigene.kartenskripteTitel,
    autoersetzung: eigene.autoersetzung.ueberschrift,
  }

  // **Ohne Skripte im Spielpaket kein Eintrag** (#1337) — und ein Kasten, der
  // offen war, steht dann auch nicht mehr da.
  const mitSkripten = kartenskripte.skripte.length > 0

  const eintraege: Menueeintrag[] = [
    {
      schluessel: 'scryfall',
      schreibt: true,
      name: titel.scryfall,
      aufWahl: () => {
        oeffne('scryfall')
      },
    },
    ...(mitSkripten
      ? [
          {
            schluessel: 'skripte',
            name: titel.skripte,
            aufWahl: () => {
              oeffne('skripte')
            },
          },
        ]
      : []),
    {
      schluessel: 'umschlag',
      name: eigene.umschlagAnnehmen,
      angehalten: umschlag.laeuft,
      aufWahl: () => {
        void umschlag.annehmen()
      },
    },
    {
      schluessel: 'autoersetzung',
      name: titel.autoersetzung,
      aufWahl: () => {
        oeffne('autoersetzung')
      },
    },
  ]

  const inhalt: Record<Kasten, ReactNode> = {
    scryfall: (
      <Scryfallanlage
        vorgang={vorgang}
        ausScryfall={ausScryfall}
        gewaehlt={gewaehlt}
        waehlen={waehlenTreffer}
      />
    ),
    skripte: (
      <Kartenskripte
        skripte={kartenskripte.skripte}
        laeuft={kartenskripte.laeuft}
        ausfuehren={kartenskripte.ausfuehren}
        bericht={kartenskripte.bericht}
        befunde={kartenskripte.befunde}
      />
    ),
    autoersetzung: <Autoersetzungen setID={setID} />,
  }

  return (
    <div className="flex flex-col gap-2">
      {/* **Zwei Gruppen in einer Zeile** — P03 (#2174): was eine Karte
          entstehen lässt, und wie der Arbeitsplatz steht. Wird es eng, bricht
          die zweite als Ganzes um. */}
      <div ref={zeile} className="flex flex-wrap items-center gap-3">
        {/* **Die Haupthandlung steht vorn** — dieselbe Stelle wie auf jeder
            Seite, und die **einzige** in dieser Zeile. */}
        <div
          role="group"
          aria-label={eigene.gruppeAnlage}
          className="flex items-center gap-3"
        >
          <Aktionsknopf
            aktion="anlegen"
            gewicht="haupt"
            schreibt
            angehalten={vorgang.laeuft('anlegen')}
            aufKlick={anlegen}
          >
            {eigene.anlegen}
          </Aktionsknopf>

          <Menue
            bezeichnung={eigene.mehr}
            eintraege={eintraege}
            gewicht="leise"
          />
        </div>

        {/* **Der Platz für die Schalter des Arbeitsplatzes** — sie bleiben
            seine (`Arbeitsplatzschalter`), stehen aber hier. */}
        {ansichtsort !== undefined && (
          <div
            ref={ansichtsort}
            role="group"
            aria-label={eigene.gruppeAnsicht}
            className="ml-auto flex items-center gap-2"
          />
        )}
      </div>

      {anlegefehler !== undefined && (
        <p role="alert" className="text-fehler">
          {anlegefehler}
        </p>
      )}

      {/* **Was der Umschlag meldet, steht an der Zeile, aus der er kam** (§6)
          — der Menüeintrag selbst ist dann schon wieder zu. */}
      {umschlag.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {umschlag.fehler}
        </p>
      )}

      {umschlag.satz !== undefined && (
        <p role="status" className="text-sm text-leise">
          {umschlag.satz}
        </p>
      )}

      {besucht.map((kasten) => (
        <Werkzeugkasten
          key={kasten}
          titel={titel[kasten]}
          offen={offen === kasten && (kasten !== 'skripte' || mitSkripten)}
          schliessen={() => {
            setzeOffen(undefined)
            zumMenue()
          }}
        >
          {inhalt[kasten]}
        </Werkzeugkasten>
      ))}
    </div>
  )
}
