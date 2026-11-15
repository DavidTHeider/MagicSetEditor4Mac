// **Was in einer Gruppe der Einstellungen steht** — UX-04.AK-01
// (#1676).
//
// **Eine eigene Datei**, damit `Einstellungen.tsx` unter der Zeilengrenze aus
// §2 bleibt. Der Schnitt folgt der Frage: dort steht, **welche** Gruppen es
// gibt und was in welche gehört; hier, **wie** eine aussieht.
import { Anbietergeheimnisse } from '../bausteine/Anbietergeheimnisse'
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { abschnittstitel } from './einstellungsabschnitte'
import { fachschluessel } from './einstellungsgruppen'
import { useTexte } from '../sprache/kontext'
import { Wahl } from './Wahl'

import type { ReactNode } from 'react'

import type { EinstellungDTO } from '../api'
import type { Abschnittsplan } from './einstellungsabschnitte'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

// Einstellungsabschnitt ist ein Abschnitt samt dem, was außer den
// Einstellungen darin steht — UXR-19
// (#1927).
export interface Einstellungsabschnitt extends Abschnittsplan {
  beiwerk?: ReactNode

  // aufschlagen ist ein Wunsch aus einer Übergabe, ihn zu öffnen (AP-10).
  aufschlagen?: string | undefined
}

// Einstellungsgruppe zeigt den Namen, die Einstellungen und das Beiwerk.
//
// # Warum der Name noch einmal dasteht
//
// **Weil der Entwurf ihn zeichnet** — im Kopf der Fläche, rechts neben der
// Liste. Er beantwortet dieselbe Frage wie ein Brotkrumenpfad: wo bin ich
// gerade? Auf einem schmalen Fenster steht die Liste **über** der Fläche,
// und dann ist er die einzige Antwort.
//
// **Der Untertitel steht seit UX-04.AK-01 nicht mehr hier**, sondern in der
// Zeile der [Gruppenliste](../bausteine/Gruppenliste.tsx): dort sieht ihn,
// wer noch sucht, und die Frage „bin ich hier richtig?" stellt sich vor dem
// Aufschlagen.
//
// # Und warum das Beiwerk unter den Einstellungen steht
//
// **Weil eine Einstellung eine Zeile ist und ein Kasten eine Fläche.** Wer
// zuerst die Kästen sähe, rollte an dem vorbei, wofür er die Seite
// aufgeschlagen hat.
export function Einstellungsgruppe({
  name,
  vorne,
  treffer,
  kopfmarke,
  eintraege,
  laeuft,
  setze,
  fach,
  abschnitte = [],
  children,
}: {
  name: string

  // vorne steht zwischen der Überschrift und den Einstellungen.
  //
  // **Über der Überschrift stünde es vor der Antwort auf „wo bin ich?"**,
  // und der Name der Gruppe ist auf einem schmalen Fenster die einzige.
  vorne?: ReactNode

  // treffer sind die Schlüssel, die zur laufenden Suchfrage passen —
  // UX-04.AK-03
  // (#1676).
  //
  // **Eine Menge und keine zweite Liste**: gezeichnet wird jede Einstellung
  // einmal, und ein Treffer bekommt eine Marke. Die Trefferliste daneben war
  // eine zweite Kopie desselben Bedienelements (#1768).
  treffer?: ReadonlySet<string> | undefined

  // kopfmarke steht rechts neben der Überschrift — „1 Treffer für „proxy"".
  kopfmarke?: ReactNode
  eintraege: readonly EinstellungDTO[]
  laeuft: (schluessel: string) => boolean
  setze: (schluessel: string, wert: string) => void

  // fach ist, was außer den Einstellungen nur Fachleute brauchen — es steht
  // mit den Fachschlüsseln im eingeklappten Kasten (UX-04.AK-04).
  fach?: ReactNode

  // abschnitte stehen eingeklappt unter dem Beiwerk — was selten gebraucht
  // wird und zusammengehört (UXR-19, #1927). Welche Schlüssel darin stehen,
  // sagt `einstellungsabschnitte.ts`.
  abschnitte?: readonly Einstellungsabschnitt[]

  children?: ReactNode
}) {
  const inAbschnitt = new Set(
    abschnitte.flatMap((abschnitt) => abschnitt.schluessel),
  )
  const allgemein = eintraege.filter(
    (eintrag) =>
      !fachschluessel.has(eintrag.schluessel) &&
      !inAbschnitt.has(eintrag.schluessel),
  )
  const fuerFachleute = eintraege.filter((eintrag) =>
    fachschluessel.has(eintrag.schluessel),
  )

  // **Der Namenskonflikt trägt die Namen der Bildseite** — dort steht er
  // auch, und sein Katalog ist ihrer („aus eigener Seite",
  // `einstellungsnamenswege_test.go`). Hier stand er roh da, bis die
  // Sichtprüfung für OPS-02 (#1719) ihn sah.
  const bilder = useTexte().seiten.bilder
  const eigenname = (schluessel: string) =>
    schluessel === 'bildkonflikt'
      ? { beschriftung: bilder.konflikt, werteNamen: bilder.konflikte }
      : {}

  const kasten = (abschnitt: Einstellungsabschnitt) => (
    <Abschnittskasten
      key={abschnitt.name}
      abschnitt={abschnitt}
      darin={eintraege.filter((eintrag) =>
        abschnitt.schluessel.includes(eintrag.schluessel),
      )}
      treffer={treffer}
      zeile={zeile}
    />
  )

  const zeile = (eintrag: EinstellungDTO) => (
    <Wahl
      key={eintrag.schluessel}
      eintrag={eintrag}
      {...eigenname(eintrag.schluessel)}
      gesucht={treffer?.has(eintrag.schluessel)}
      angehalten={laeuft(eintrag.schluessel)}
      aendern={(wert) => {
        setze(eintrag.schluessel, wert)
      }}
    />
  )

  return (
    <div className="flex flex-col gap-6">
      <h2 className="flex flex-wrap items-center gap-3 border-b border-kante pb-2 text-lg font-semibold">
        {name}

        {kopfmarke}
      </h2>

      {vorne}

      {/* **Ohne Einstellungen steht hier nichts** — eine leere Liste wäre
          ein Gitter ohne Inhalt, und das Beiwerk rückte um seine Höhe
          nach unten. */}
      {allgemein.length > 0 && (
        <div className="grid max-w-xl gap-5">{allgemein.map(zeile)}</div>
      )}

      {/* **Ein Abschnitt mit `oben` steht vor dem Beiwerk** — die Werkzeuge
          ohne Fenster unter *Verbindungen* (AP-10, B-19). */}
      {abschnitte.filter((abschnitt) => abschnitt.oben).map(kasten)}

      {children}

      {/* **Was selten gebraucht wird, steht eingeklappt** — R3 aus
          `docs/UIREVIEW.md`: eine Gruppe offen, der Rest dahinter. */}
      {abschnitte.filter((abschnitt) => !abschnitt.oben).map(kasten)}

      {/* **Eingeklappt, und die Suche öffnet ihn** — ein Treffer darin
          hält den Kasten offen, sonst führte „frist" zu einem Feld, das
          gar nicht gezeichnet ist (UX-04.AK-03). */}
      {(fuerFachleute.length > 0 || fach !== undefined) && (
        <FachleuteKasten
          offenGehalten={fuerFachleute.some((eintrag) =>
            treffer?.has(eintrag.schluessel),
          )}
        >
          {fuerFachleute.length > 0 && (
            <div className="grid max-w-xl gap-5">
              {fuerFachleute.map(zeile)}
            </div>
          )}
          {fach}
        </FachleuteKasten>
      )}
    </div>
  )
}

// FachleuteKasten ist der Kasten aus dem Artboard *Einstellungen*: „Was
// Fachleute hier sonst noch einstellen", mit dem Satz, dass Aufklappen
// nichts ändert (S11.NAT-03).
function FachleuteKasten({
  offenGehalten,
  children,
}: {
  offenGehalten: boolean
  children: ReactNode
}) {
  const eigene = useTexte().seiten.einstellungen

  return (
    <Aufklappkasten titel={eigene.fachleute} offenGehalten={offenGehalten}>
      <div className="flex flex-col gap-5">
        <p className="max-w-prose text-sm text-leise">{eigene.fachleuteSatz}</p>
        {children}
      </div>
    </Aufklappkasten>
  )
}

// Abschnittskasten ist ein Abschnitt einer Gruppe, eingeklappt — UXR-19
// (#1927).
//
// **Der Name steht am Knopf und nirgends sonst**: er ist zugleich die
// Überschrift des Kastens (`Aufklappkasten`). Ein Beiwerk darin zeichnet
// deshalb seine eigene nicht (`eingebettet`).
//
// **Ein Abschnitt ohne Inhalt steht nicht da** — ein Knopf, der nichts
// aufklappt, verspräche etwas, was nicht kommt. Das geschieht, wenn eine
// ältere Go-Seite die Einstellungen eines Abschnitts noch nicht kennt.
function Abschnittskasten({
  abschnitt,
  darin,
  treffer,
  zeile,
}: {
  abschnitt: Einstellungsabschnitt
  darin: readonly EinstellungDTO[]
  treffer: ReadonlySet<string> | undefined
  zeile: (eintrag: EinstellungDTO) => ReactNode
}) {
  const texte = useTexte()

  if (darin.length === 0 && abschnitt.beiwerk === undefined) {
    return null
  }

  return (
    <Aufklappkasten
      titel={abschnittstitel(texte, abschnitt.name)}
      aufschlagen={abschnitt.aufschlagen}
      offenGehalten={abschnitt.schluessel.some(
        (schluessel) => treffer?.has(schluessel) === true,
      )}
    >
      <div className="flex flex-col gap-5">
        {darin.length > 0 && (
          <div className="grid max-w-xl gap-5">{darin.map(zeile)}</div>
        )}
        {abschnitt.beiwerk}
      </div>
    </Aufklappkasten>
  )
}

// Anmeldegebiet erklärt, wo eine Anmeldung herkommt — Nacharbeit zu
// Werkstück 4 von Phase 4.12
// (#1351).
//
// # Warum es das gibt
//
// **Weil die Nutzerseite hierher verweist und hier nichts stand.** „Wer sich
// über einen Anbieter anmelden will, registriert dort eine eigene Anwendung
// und trägt ihre Kennung in den Einstellungen ein" — und die Einstellungen
// hatten kein Feld dafür; die Domäne wies den Schlüssel sogar ab. Der ganze
// Anmeldeweg war damit unerreichbar (§10, erste Dimension).
export function Anmeldegebiet() {
  const eigene = useTexte().seiten.einstellungen

  return (
    <div className="flex max-w-prose flex-col gap-4">
      <Erklaerungsabsatz text={eigene.anmeldungHinweis} klasse="text-leise" />

      {/* **Ein zweiter Absatz und kein längerer erster** — Befund 3 des
          Zweitblicks über Phase 4.12 (#1351). Wohin das Geheimnis gehört,
          ist eine eigene Frage, und sie betrifft nur einen der vier
          Anbieter. */}
      <Erklaerungsabsatz text={eigene.anmeldungGeheimnis} klasse="text-leise" />

      {/* **Und das Feld dafür** — #1401. */}
      <Anbietergeheimnisse />
    </div>
  )
}
