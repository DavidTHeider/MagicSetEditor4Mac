// **Der Bereich Spielen** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Was hier steht und was nicht
//
// *Wie es aussieht* steht hier; *was die Seite tut*, in `usePartie` (§2).
//
// # Wie die Seite ohne Tisch steht
//
// **Nach dem Artboard *Eine Partie anfangen*** — UXR-13
// (#1921):
// der Anfang führt, und neben ihm steht in einer Spalte *Deine Partien*.
// **Neben und nicht darunter, sobald der Platz reicht** (`@4xl`, gefragt am
// Behälter und nicht am Fenster): bei 1.280 Punkten trägt die Spalte fünf
// Partien, im engen Fenster rückt sie unter den Anfang. Der fremde Tisch und
// die Leihe stehen hinter je einem Knopf im Kopf, in einer Schublade
// (`Spielschubladen`); wer an einem fremden Tisch sitzt, hat ihn oben auf der
// Seite (`Gastplatz`).
//
// # Und was das hier ausdrücklich nicht ist
//
// **Nicht der Tisch.** Der GPU-gerenderte Spieltisch mit Kartenflug,
// Mulligan-Profilen und abschaltbaren Animationen ist Werkstück G3
// (#1460).
// Was hier steht, ist der Weg, über den das Fenster den Spielkern **erreicht**
// — mit allem, was der Kern kann, und ohne eine Zeichenfläche.
//
// **Warum trotzdem jetzt**: eine Bindung ohne Aufrufer ist keine
// Anwendungsfunktion, sondern eine Zusage ohne Einlösung (§2, und der Wächter
// `internal/werkzeuge/bindungsaufrufer_test.go`). Die vier Flächen aus
// `FAEHIGKEITEN.md` sagen „an allen vier"; ohne diese Seite wäre das für die
// vierte unwahr.
import { useEffect, useRef, useState } from 'react'

import { useAnschluss } from '../anschlussraum'
import { Knopf } from '../bausteine/Knopf'
import { Partieanfang } from './Partieanfang'
import { useMitspielen } from '../haken/useMitspielen'
import { Partienliste } from '../bausteine/Partienliste'
import { Gastplatz, sitztAmFremdenTisch } from './Partiegast'
import { Partietisch } from './Partietisch'
import { Seitenkopf } from '../bausteine/Seitenkopf'
import { Spielschubladen, type Spielschublade } from './Spielschubladen'
import { usePartie } from '../haken/usePartie'
import { usePartiegast } from '../haken/usePartiegast'
import { usePartienliste } from '../haken/usePartienliste'
import { usePartieverwaltung } from '../haken/usePartieverwaltung'
import { partiename } from './partiewoerter'
import { useSeitenweg } from '../seitenweg'
import { useSprache, useTexte } from '../sprache/kontext'
import { Verdeckung, useVerdeckt } from '../verdeckung'

import type { Seitenschluessel } from '../navigation'

export function Spielen({
  wechseln,
  bestandsstand = 0,
}: {
  // wechseln führt auf eine andere Seite — ohne Deck zu den Decks (UXR-13).
  wechseln: (seite: Seitenschluessel) => void

  // bestandsstand zählt fremde Schreibvorgänge — die Liste lädt dann nach
  // (Befund 3 des Gesamtreviews über 0.6.0).
  bestandsstand?: number
}) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const lauf = usePartie(bestandsstand)
  const liste = usePartienliste(bestandsstand)

  // **Der Lauf am fremden Tisch gehört der Seite** — UXR-13: der Beitritt
  // steht in der Schublade, der Platz auf der Seite, und beide zeigen
  // denselben Stand.
  const gast = usePartiegast()

  // **„Bei Ada mitspielen“ von der Deckseite** — P9c (#2377).
  useMitspielen(gast)

  // **Aus einem Turnier: eine gebundene Partie** — DOM-04.AK-02
  // (#1699).
  // **In einem Effekt**, wie bei den Decks: Aufschlagen lädt die Partie, und
  // ein Ruf nach draußen gehört nicht ins Zeichnen.
  const anschluss = useAnschluss()
  const aufzuschlagen =
    anschluss.offen?.art === 'partieansicht' ? anschluss.offen : undefined
  const { ansehen } = lauf
  useEffect(() => {
    if (aufzuschlagen !== undefined) {
      ansehen(aufzuschlagen.partieID, aufzuschlagen.sitz)
    }
  }, [aufzuschlagen, ansehen])

  // **Ein im Gegenüberkasten aufgenommenes Gegenüber steht gleich zur Wahl**
  // — beim Beitritt und bei der Leihe (UX-08.AK-04, #1680). Derselbe Zähler
  // wie beim Handeln (UX-03.AK-03).
  const [kontaktstand, setzeKontaktstand] = useState(0)
  // leihstand zählt geglückte Schreiben an Leihen — die Deckwahl am Sitz
  // lädt dann neu (P1b, #2370).
  const [leihstand, setzeLeihstand] = useState(0)

  // **Welche Schublade steht** — darunter ist die Seite gesperrt (`inert`)
  // und verdeckt (`verdeckung.ts`): eine Taste am fremden Tisch gehörte sonst
  // einem Tisch, den niemand sieht.
  const [schublade, setzeSchublade] = useState<Spielschublade>()
  const verdeckt = useVerdeckt() || schublade !== undefined

  // **Wer sich gesetzt hat, sieht seinen Tisch** — die Schublade des
  // Beitritts geht zu, sobald der Stand sagt, dass er sitzt. **Beim Zeichnen
  // und nicht in einem Effekt**: der Wechsel des Stands ist der Anlass, und
  // ein Effekt ließe die Schublade einen Zug lang über dem Tisch stehen.
  const sitzt = sitztAmFremdenTisch(gast)
  const [sassZuletzt, setzeSassZuletzt] = useState(sitzt)
  if (sitzt !== sassZuletzt) {
    setzeSassZuletzt(sitzt)

    if (sitzt && schublade === 'gast') {
      setzeSchublade(undefined)
    }
  }

  // **Wer den Tisch verlässt, sieht die Liste, wie sie jetzt ist** — UXR-F16
  // (#2031).
  //
  // **Weil „dran" am Tisch umspringt**: wer am Tisch gepasst hat, kommt zu
  // einer Liste zurück, die noch vom Laden der Seite stammt — und die sagte
  // „dran" über eine Partie, auf die nicht mehr er wartet. Zug, Objektzahl und
  // Stand waren schon bisher so alt, nur fiel es weniger auf.
  //
  // **Beim Verlassen und nicht bei jedem Zug**: am Tisch ist die Liste nicht
  // zu sehen, und ein Ruf je Befehl holte sie für niemanden. **Bestenfalls**
  // (§7, Regel 3): scheitert das Laden, zeigt die Liste die Meldung, die sie
  // immer zeigt, und die Partie bleibt, wie sie war.
  const amTisch = lauf.partie !== undefined
  const warAmTisch = useRef(amTisch)
  const { nachladen: listeNachladen } = liste
  useEffect(() => {
    if (warAmTisch.current && !amTisch) {
      listeNachladen()
    }

    warAmTisch.current = amTisch
  }, [amTisch, listeNachladen])

  // **Die Verwaltung lädt die Liste nach** (§7, Regel 3): wer eine Partie
  // kopiert oder einliest, soll sie sehen — und ein Lesefehler danach macht
  // die Änderung nicht rückgängig.
  const verwaltung = usePartieverwaltung(liste.nachladen)

  // **Über dem Tisch steht die Partie im Weg** — UX-03.AK-02, *Spielen ›
  // Partie gegen Bo › Am Tisch* wie im Artboard *Tisch*. Der Rahmen kennt nur
  // das offene Set, und ob ein Tisch offensteht, weiß nur diese Seite. **Der
  // Tisch eines Gastes sagt nichts**: er steht neben dem eigenen Anfang, nicht
  // an seiner Stelle.
  //
  // **Und ohne Tisch sagt er, was die Seite tut** — *Spielen › Partie
  // anfangen*, wie im Artboard *Eine Partie anfangen*.
  const sprache = useSprache()
  useSeitenweg(
    lauf.partie === undefined
      ? [{ beschriftung: eigene.anfangen }]
      : [
          { beschriftung: partiename(eigene, lauf.partie, sprache) },
          { beschriftung: eigene.tischTitel },
        ],
  )

  // **Am Tisch ist der Tisch die Seite** — Befund F09 der nativen Prüfung
  // von 0.5.0 (UX-03, UX-08): kein Seitenkopf über ihm. Wo er ist, sagt die
  // Brotkrume, und „Am Tisch" ist dann die Überschrift der Seite.
  //
  // **Der `key` trägt die Partiekennung** — §7, Regel 1: wer eine zweite
  // Partie anfängt, bekommt einen leeren Tisch und nicht die halb
  // ausgefüllten Felder der ersten.
  if (lauf.partie !== undefined) {
    return (
      <div className="flex flex-col gap-6">
        <Partietisch
          key={lauf.partie.id}
          lauf={lauf}
          partie={lauf.partie}
          ebene={1}
        />
      </div>
    )
  }

  return (
    <>
      <Verdeckung value={verdeckt}>
        <div className="flex flex-col gap-6" inert={schublade !== undefined}>
          {/* **Ohne Tisch heißt die Seite, was sie tut** — „Eine Partie
              anfangen", wie im Artboard; „Spielen" ist der Bereich, und der
              steht in der Brotkrume darüber. **Daneben das Seltene**, je ein
              Knopf zu seiner Schublade (UXR-13); wer schon an einem fremden
              Tisch sitzt, setzt sich nicht noch einmal. */}
          <Seitenkopf titel={eigene.anfangTitel} erklaerung={eigene.erklaerung}>
            {!sitzt && (
              <Knopf
                aufKlick={() => {
                  setzeSchublade('gast')
                }}
              >
                {eigene.gastOeffnen}
              </Knopf>
            )}
            <Knopf
              aufKlick={() => {
                setzeSchublade('leihe')
              }}
            >
              {eigene.leiheOeffnen}
            </Knopf>
          </Seitenkopf>

          {/*
            **Am Kopf steht nur, was nirgends sonst hingehört** — CLAUDE.md
            §6. Die Meldung des Anfangens steht bei ihrem Knopf, unten in
            `Partieanfang`; sie hier **zusätzlich** zu zeigen hieße, denselben
            Satz zweimal hinzustellen, und wer ihn zweimal liest, sucht zwei
            Fehler. Was hierher gehört, ist das, was **ohne** Tisch
            schiefgeht — ein Einstieg in eine Partie, die es nicht mehr gibt.
          */}
          {lauf.fehler !== undefined &&
            lauf.fehlerVorgang !== 'partie-anfangen' && (
              <p role="alert" className="text-fehler">
                {lauf.fehler}
              </p>
            )}

          <div className="@container">
            <div className="flex flex-col gap-6 @4xl:flex-row @4xl:items-start">
              <div className="flex min-w-0 flex-1 flex-col gap-6">
                {/* **Wer an einem fremden Tisch sitzt, spielt dort** — der
                    Platz steht über dem eigenen Anfang. */}
                <Gastplatz lauf={gast} />

                <Partieanfang
                  lauf={lauf}
                  wechseln={wechseln}
                  leihstand={leihstand}
                  gegenueber={liste.partien[0]?.sitze[1]?.name}
                />
              </div>

              {/*
                **Der Weg zurück in die eigene Partie** (#1600), als Spalte
                neben dem Anfang wie im Artboard: wer hier ist, fängt eine an
                oder nimmt eine auf, und beides ist die eigene Seite. Sobald
                ein Tisch offensteht, ist die Liste weg — dann gibt es nichts
                aufzunehmen.
              */}
              <div className="flex min-w-0 flex-col @4xl:w-80 @4xl:shrink-0">
                <Partienliste
                  partien={liste.partien}
                  geladen={liste.geladen}
                  fehler={liste.fehler}
                  wiederholen={liste.nachladen}
                  einsteigen={lauf.ansehen}
                  verwaltung={verwaltung}
                />
              </div>
            </div>
          </div>
        </div>
      </Verdeckung>

      <Spielschubladen
        offen={schublade}
        schliessen={() => {
          setzeSchublade(undefined)
        }}
        gast={gast}
        kontaktstand={kontaktstand}
        leiheGeaendert={() => {
          setzeLeihstand((bisher) => bisher + 1)
        }}
        kontaktGeaendert={() => {
          setzeKontaktstand((bisher) => bisher + 1)
        }}
      />
    </>
  )
}
