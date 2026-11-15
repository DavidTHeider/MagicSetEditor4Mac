// **Die drei Bereiche des Deckbauplatzes** — Werkstück U10 von Phase 5.1
// (#1421).
//
// **Eine eigene Datei**, damit `Decks.tsx` unter der Zeilengrenze aus §2
// bleibt. Der Schnitt folgt der Frage: dort steht, **wie** der Arbeitsplatz
// aufgeteilt ist, hier, **was** in den Teilen steht.
//
// # Warum die Kartenwahl in der Mitte steht
//
// **Seit UXR-11** (#1919).
// Hier stand bis dahin: „im Navigator, weil sie eine Suche ist". Gemessen mit
// acht Decks begann sie dort **unter** der Deckliste, also unter der Falz —
// das Werkzeug, mit dem ein Deck entsteht, war das am schwersten erreichbare.
// Das Artboard *Decks* setzt die Suche in die Mitte, über die Plätze, die sie
// füllt; dort steht sie jetzt, mit **denselben** Handlungen wie die Plätze
// ([useDeckaktionen], keine zweite Abschrift). Der Navigator wählt, die Mitte
// baut, der Inspector sieht.
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Deckanalyse } from '../bausteine/Deckanalyse'
import { Deckanlagekasten } from './Deckanlagekasten'
import { Deckkartenwahl } from '../bausteine/Deckkartenwahl'
import { Deckliste } from '../bausteine/Deckliste'
import { Deckumschlag } from '../bausteine/Deckumschlag'
import { Deckuebernahme } from '../bausteine/Deckuebernahme'
import { Deckwerkbank } from '../bausteine/Deckwerkbank'
import { einsetzen } from '../sprache/einsetzen'
import { Knopf } from '../bausteine/Knopf'
import { useAnschluss } from '../anschlussraum'
import { useDeckaktionen } from '../haken/useDeckaktionen'
import { useDeckpruefung } from '../haken/useDeckpruefung'
import { useTexte } from '../sprache/kontext'

import type { Deckaktionen } from '../haken/useDeckaktionen'
import type { Deckhandlung } from '../haken/useDeckhandlung'
import type { Deckstand } from '../haken/useDecks'
import type { DeckDTO, HabenstandAntwort } from '../apidecks'
import type { Ladezustand } from '../haken/useLaden'
import type { Seitenschluessel } from '../navigation'
import { Deckstapelaufnahme } from '../bausteine/Deckstapelaufnahme'

// Deckbaunavigator ist der linke Bereich: Decks finden, anlegen, holen.
//
// # Warum Anlage und Übernahme in Kästen stehen
//
// **Weil sie Werkzeuge sind und keine Navigation.** Ein Deck legt man einmal
// an; gesucht wird bei jeder Karte. Ausgeklappt standen beide Formulare
// **unter** der Kartensuche und schoben sie unter die Falz: gemessen am
// 2026-09-09 in der Sichtprüfung war die Seite bei 1440 × 900 **1.847 Punkte**
// hoch, und die Felder der Kartensuche begannen bei 780.
//
// Das ist derselbe Befund, für den es [Aufklappkasten] gibt — sein eigener
// Kommentar beschreibt ihn (Werkstück C von Phase 4.10): Werkzeuge standen an
// der Stelle der Sache, die man immer braucht. **Die Kartensuche steht seit
// UXR-11 in der Mitte** (oben); der Befund gilt für die Deckliste genauso.
export function Deckbaunavigator({
  stand,
  handlung,
  wechseln,
}: {
  stand: Deckstand
  handlung: Deckhandlung
  wechseln: (seite: Seitenschluessel) => void
}) {
  const eigene = useTexte().seiten.decks

  // **Eine mitgereiste Karte wartet sichtbar** — UX-03.AK-01
  // (#1675).
  //
  // **Die Kartenwahl steht erst bei offenem Deck da.** Ohne diesen Satz käme
  // jemand aus der Sammlung hier an und sähe nichts von dem, was er gerade
  // geschickt hat — ein stiller Zustand, und §6 verlangt, dass ein leerer
  // sagt, was der nächste Schritt ist.
  const anschluss = useAnschluss()
  // **Und ein Stapel wartet genauso** (UX-03.AK-01): er braucht ein Deck,
  // in das er fällt.
  const wartend =
    stand.offen === null &&
    (anschluss.offen?.art === 'deck' || anschluss.offen?.art === 'deckstapel')
      ? anschluss.offen
      : undefined

  return (
    <div className="flex flex-col gap-6">
      {wartend !== undefined && (
        <p className="text-sm">
          {einsetzen(eigene.anschlussOhneDeck, { name: wartend.name })}
        </p>
      )}

      {/* **Das Anlegen steht über der Liste**, und das ist eine Messung, keine
          Vorliebe: darunter lag das Namensfeld auf der 16. Tabstelle, darüber
          auf der 13. — jedes Deck der Liste trägt drei Knöpfe, und die Liste
          wächst. Der Kernweg `deck` hält damit sein Ziel ≤ 14
          (`docs/UX-ARCHITEKTUR.md`).

          **Er beginnt zu, auch auf einem leeren Profil.** Ein Anlauf ließ ihn
          offen beginnen, solange es kein Deck gibt — und das hat sich gegen
          den Menschen gestellt: beim ersten Zeichnen ist **jede** Liste leer,
          also stand er offen und klappte zu, sobald die Antwort kam. Wer in
          dem Augenblick klickte, klappte ihn zu statt auf. Ein Zustand, der
          sich unter der Hand ändert, ist schlechter als einer, der stillsteht;
          welchen Schritt ein leeres Profil als Nächstes tut, sagt der
          Leerzustand der Liste. */}
      <Deckanlagekasten stand={stand} />

      <Deckliste stand={stand} handlung={handlung} />

      {/* **Die seltenen Wege in einem Kasten** — AP-6 (#2363), Entwurf
          `Bereiche09`: „Sekundäres im …“. Übernahme, Umschlag und „Karten
          holen aus“ standen einzeln am Fuß des Navigators und zählten vier
          zu den Bedienelementen im Bild (R2); das offene Deck lag mit 27
          über den zwanzig. Zugeklappt bleibt einer. */}
      <Aufklappkasten titel={eigene.wegeTitel}>
        <div className="flex flex-col gap-4">
          {/* **Das Holen neben dem Anlegen** — Werkstück 12 (#1351): beide
              bringen ein Deck in dieses Profil, und wer eines sucht, sucht
              sie an derselben Stelle. */}
          <Aufklappkasten titel={eigene.uebernahme}>
            <Deckuebernahme stand={stand} />
          </Aufklappkasten>

          {/* **Und der portable Umschlag** — Werkstück T7 (#1472): derselbe
              Weg, über den ein Set geht, trägt auch ein Deck. */}
          <Aufklappkasten titel={eigene.umschlagTitel}>
            <Deckumschlag deck={stand.offen} angenommen={stand.neuLaden} />
          </Aufklappkasten>

          {/* **Woher die Karten kommen** — UX-03.AK-01
              (#1675).
              **Zwei Wege ohne Last**: sie führen dorthin, wo etwas zu wählen
              ist, und von dort kommt es über den Anschluss zurück. Deshalb
              `wechseln` und nicht `uebergib`. */}
          <section className="flex flex-col gap-2 border-t border-kante pt-3">
            {/* **`h2` und nicht `h3`** — auf dem Deckbauplatz steht darüber
                die Seitenüberschrift, und ohne ein `h2` dazwischen wäre das
                ein Sprung von eins auf drei (`barrierefreiheit.test.tsx`). */}
            <h2 className="text-sm font-medium text-leise">
              {eigene.holenAus}
            </h2>

            <Knopf
              aufKlick={() => {
                wechseln('sammlung')
              }}
            >
              {eigene.holenSammlung}
            </Knopf>

            <Knopf
              aufKlick={() => {
                wechseln('sets')
              }}
            >
              {eigene.holenSet}
            </Knopf>
          </section>
        </div>
      </Aufklappkasten>
    </div>
  )
}

// Kartensuche legt eine gefundene Karte ins offene Deck.
function Kartensuche({
  aktionen,
  spiel,
}: {
  aktionen: Deckaktionen
  spiel: string
}) {
  const eigene = useTexte().seiten.decks

  return (
    <section className="flex flex-col gap-3">
      <h3 className="font-medium">{eigene.suche}</h3>

      <Deckkartenwahl
        spiel={spiel}
        angehalten={aktionen.laeuft}
        legen={(karteID, printing, menge, bereich) => {
          void aktionen.legen(karteID, printing, menge, bereich)
        }}
      />
    </section>
  )
}

// Deckbauzentrum ist der mittlere Bereich: das offene Deck.
export function Deckbauzentrum({
  stand,
  abgleich,
  zeigePruefung,
}: {
  stand: Deckstand
  abgleich: Ladezustand<HabenstandAntwort | null>
  zeigePruefung: (deckID: string) => void
}) {
  const eigene = useTexte().seiten.decks

  if (stand.offen === null) {
    return <p>{eigene.keinDeckOffen}</p>
  }

  // **Der `key` räumt beim Deckwechsel jedes Formular mit** (§7, Regel 1 und
  // 2 in einer Gestalt): ohne ihn stünde die halb ausgefüllte Zeile des
  // vorigen Decks über dem neuen, und der nächste Klick legte sie dort hinein.
  return (
    <Zentrum
      key={stand.offen.id}
      stand={stand}
      deck={stand.offen}
      abgleich={abgleich}
      zeigePruefung={zeigePruefung}
    />
  )
}

function Zentrum({
  stand,
  deck,
  abgleich,
  zeigePruefung,
}: {
  stand: Deckstand
  deck: DeckDTO
  abgleich: Ladezustand<HabenstandAntwort | null>
  zeigePruefung: (deckID: string) => void
}) {
  const aktionen = useDeckaktionen(stand, deck)
  const pruefung = useDeckpruefung(deck)

  return (
    <div className="flex flex-col gap-4">
      {/* **Ein Stapel aus der Sammlung steht über der Werkbank** (UX-03.AK-01,
          #1675) — dort, wo auch eine einzelne Karte landet. */}
      <Deckstapelaufnahme
        legeAlle={aktionen.legeAlle}
        angehalten={aktionen.laeuft}
      />
      <Deckwerkbank
        stand={stand}
        deck={deck}
        aktionen={aktionen}
        pruefung={pruefung}
        abgleich={abgleich}
        zeigePruefung={() => {
          zeigePruefung(deck.id)
        }}
      >
        <Kartensuche aktionen={aktionen} spiel={deck.spiel} />
      </Deckwerkbank>
    </div>
  )
}

// Deckbauinspector ist der rechte Bereich: die Analyse, die nichts ändert.
//
// **Ohne Deck nicht derselbe Satz wie im Zentrum** (UX-15, #1771): zwei
// gleiche nebeneinander sagen nichts Zweites. Hier steht, was diese Spalte
// zeigen wird.
export function Deckbauinspector({
  stand,
  abgleich,
  oeffne,
}: {
  stand: Deckstand
  abgleich: Ladezustand<HabenstandAntwort | null>

  // oeffne ist eine Lasche, die von außen aufgehen soll (AP-8, #2365).
  oeffne?: string | undefined
}) {
  const eigene = useTexte().seiten.decks

  if (stand.offen === null) {
    return <p className="text-sm text-leise">{eigene.analyseLeer}</p>
  }

  return (
    <Analyse
      key={stand.offen.id}
      stand={stand}
      deck={stand.offen}
      abgleich={abgleich}
      oeffne={oeffne}
    />
  )
}

// Mitspielen nimmt das offene Deck an den Tisch.
//
// **Im Nebengewicht** — AP-3 (#2360), Entwurf `Bereiche09`: im offenen Deck
// führt „Ins Deck“.
//
// **Es wechselt die Fläche selbst** — über [useAnschluss]: wohin ein
// Anschluss gehört, steht an **einer** Stelle (`anschlussseite`) und nicht
// an jedem Knopf.
function Mitspielen({ deck }: { deck: DeckDTO }) {
  const eigene = useTexte().seiten.decks
  const anschluss = useAnschluss()

  return (
    <Knopf
      aufKlick={() => {
        anschluss.uebergib({
          art: 'partie',
          name: deck.name,
          deckID: deck.id,
        })
      }}
    >
      {eigene.mitspielen}
    </Knopf>
  )
}

function Analyse({
  stand,
  deck,
  abgleich,
  oeffne,
}: {
  stand: Deckstand
  deck: DeckDTO
  abgleich: Ladezustand<HabenstandAntwort | null>
  oeffne: string | undefined
}) {
  // **Zweimal derselbe Haken und trotzdem keine zweite Anfrage nach außen** —
  // Zentrum und Inspector stehen in getrennten Bäumen, und `useLaden` bündelt
  // nicht.
  //
  // **Nachgelesen am 2026-09-09**: `scryfall.Klient.Sammlung` geht über
  // `k.speicher.Hole(…)` (`internal/scryfall/sammlung.go`), und die Frist ist
  // `scryfall.Speicherfrist` = 24 Stunden. Der zweite Ruf kostet einen Weg
  // über die Naht und trifft denselben Stand; ins Netz geht er nicht.
  const pruefung = useDeckpruefung(deck)

  return (
    <div className="flex flex-col gap-4">
      <Deckanalyse
        stand={stand}
        deck={deck}
        pruefung={pruefung}
        abgleich={abgleich}
        oeffne={oeffne}
      />

      {/* **Der Weg an den Tisch** — UX-03.AK-01. Der Entwurf setzt ihn als
          Hauptknopf über die volle Breite an den Fuß des Inspectors: es ist
          die Handlung, auf die ein fertiges Deck hinausläuft. */}
      <Mitspielen deck={deck} />
    </div>
  )
}
