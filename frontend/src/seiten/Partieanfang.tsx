// **Eine Partie anfangen** — Werkstück G1 von Phase 5.5
// (#1458).
//
// # Warum das Spiel nicht gewählt wird
//
// **Weil es aus dem ersten Deck kommt.** Ein Deck gehört zu einem Spiel; ein
// Auswahlfeld daneben ließe sich auf etwas anderes stellen, und die Partie
// scheiterte an einer Angabe, die der Mensch gar nicht machen wollte. Was
// gilt, steht trotzdem da — sichtbar und nicht verstellbar.
//
// # Und warum die Sitzkennungen nicht dastehen
//
// **Weil sie Technik sind.** Am Tisch unterscheidet ein Name; die Kennung ist
// das, womit der Kern rechnet. Sie wird aus der Reihenfolge gebildet und ist
// damit innerhalb dieser Partie eindeutig — mehr verlangt [domain.Partie]
// nicht.
import { Ladefehler, Wirdgeladen } from '../bausteine/Ladehinweise'
import { useEffect, useRef, useState } from 'react'

import { api, auspacken } from '../api'
import { einsetzen } from '../sprache/einsetzen'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Knopf } from '../bausteine/Knopf'
import { Leerzustand } from '../bausteine/Leerzustand'
import { Regeltafel } from './Regeltafel'
import { Sitzliste, sitzkennung, type Sitzwunsch } from './Sitzliste'
import { useDecks } from '../haken/useDecks'
import { useEigeneProfile } from '../haken/useEigeneProfile'
import { useLaden } from '../haken/useLaden'
import { useRegelfreigabe } from '../haken/useRegelfreigabe'
import { spielbar } from '../spiel/leihwahl'
import { usePartiekontakte } from '../haken/usePartiekontakte'
import { useSitzwuensche } from '../haken/useSitzwuensche'
import { Anfaengermodus } from '../bausteine/Regelwahlfeld'
import { useTexte } from '../sprache/kontext'

import type { Partielauf } from '../haken/usePartie'
import type { Seitenschluessel } from '../navigation'
// wasFehlt nennt den ersten Schritt, der noch offen ist — oder `undefined`.
//
// **In der Reihenfolge, in der man sie tut**, und nicht als Liste: wer drei
// Sätze auf einmal liest, weiß danach nicht, wo er anfangen soll.
function wasFehlt(
  eigene: { fehltDeck: string; fehltName: string; fehltEigenesDeck: string },
  wuensche: readonly Sitzwunsch[],
  spiel: string,
): string | undefined {
  const besetzt = (wunsch: Sitzwunsch) =>
    wunsch.deckID !== '' ||
    wunsch.leiheID !== '' ||
    (wunsch.eingeladen ?? '') !== ''

  // **Leiht jeder Sitz, fehlt das Spiel** — eine Leihe bringt keines mit
  // (P1b, #2370), und „Wähle ein Deck“ wäre hier falsch.
  if (spiel === '' && wuensche.length > 0 && wuensche.every(besetzt)) {
    return eigene.fehltEigenesDeck
  }
  if (spiel === '') return eigene.fehltDeck
  if (wuensche.some((wunsch) => wunsch.name === '')) return eigene.fehltName
  if (!wuensche.every(besetzt)) return eigene.fehltDeck

  return undefined
}

export function Partieanfang({
  lauf,
  wechseln,
  leihstand = 0,
  gegenueber,
}: {
  lauf: Partielauf

  // wechseln führt ohne Deck zu den Decks — UXR-13.
  wechseln: (seite: Seitenschluessel) => void

  // leihstand zählt geglückte Schreiben an Leihen — dann werden sie neu
  // gelesen (P1b, #2370).
  leihstand?: number

  // gegenueber ist der Name am zweiten Sitz der jüngsten Partie — aus der
  // Liste, die die Seite schon hält (P6, #2374).
  gegenueber?: string | undefined
}) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  const decks = useDecks()

  // **Die Leihen, die sich spielen lassen** — P1b (#2370), die Regel steht in
  // `spiel/leihwahl.ts`. **Ohne Antwort keine** — ein Fehler hier nimmt
  // niemandem ein Deck.
  const geladeneLeihen = useLaden(async () => {
    const antwort = await api().Deckleihen()

    return auspacken(antwort, () => antwort.leihen)
  }, [leihstand])
  const leihen = (geladeneLeihen.daten ?? []).filter(spielbar)

  // **Die Profile des Menschen, der arbeitet** (G8,
  // #1465)
  // — und nicht die des gezeigten Profils: wer in ein fremdes wechselt,
  // bleibt, wer er ist. Ein Sitz gehört einem Profil **desselben** Nutzers.
  const { profile } = useEigeneProfile()
  const kontakte = usePartiekontakte()

  // **Die Sitze, vorbelegt und mit dem Deck, das mitgereist ist** — P6
  // (#2374); Zustand und Herkunft stehen in `useSitzwuensche`.
  const {
    wuensche,
    setzeWuensche,
    setzeWunsch,
    mitgereist,
    quellen,
    schnell,
    verbrauche,
  } = useSitzwuensche({
    decks: decks.geladen ? decks.decks : undefined,
    gegenueber,
    ersatzname: eigene.sitzZweiVorgabe,
  })

  // **Das Spiel kommt aus dem ersten eigenen Deck** — solange keines gewählt
  // ist, steht hier nichts, und der Knopf ruht. Eine Leihe bringt keines mit
  // (P1b, #2370); ein Sitz mit Leihe wird übersprungen.
  const gewaehlt = decks.decks.find(
    (deck) =>
      deck.id === wuensche.find((wunsch) => wunsch.deckID !== '')?.deckID,
  )
  const spiel = gewaehlt?.spiel ?? ''

  // **Die Strenge ist heute nicht wählbar** — der Sandkasten ist die einzige,
  // die jedes Regelwerk trägt, und was sonst zur Wahl stünde, sagt die
  // Regelschau darunter. Ein Auswahlfeld mit einer Wahl wäre eines zu viel;
  // es kommt, wenn ein Regelwerk mehr trägt (G4, #1461).
  //
  // **Der Bedienmodus dagegen ist wählbar** — PLAY-02.AK-01
  // (#1710):
  // wie viel das Programm übernimmt, entscheidet die Runde, und im freien
  // Modus je Bereich.
  const [modus, setzeModus] = useState(Anfaengermodus)
  const [abgewaehlt, setzeAbgewaehlt] = useState<string[]>([])

  const regeln = useRegelfreigabe(
    spiel,
    'sandkasten',
    wuensche.length,
    modus,
    abgewaehlt,
  )

  // **Was fehlt, steht da — nicht nur, dass etwas fehlt.**
  //
  // #1460 verlangt „verständliche Sperren": ein Knopf, der grau ist und
  // schweigt, lässt raten. Die Liste nennt den ersten fehlenden Schritt in
  // der Reihenfolge, in der man sie tut.
  const fehlt = wasFehlt(eigene, wuensche, spiel)
  const vollstaendig = fehlt === undefined

  // **Die Spielart folgt aus der Sitzzahl und wird nicht gewählt.**
  //
  // Ein Auswahlfeld daneben ließe sich auf „Goldfish" stellen, während zwei
  // Sitze dastehen — dieselbe Überlegung wie beim Spiel eine Zeile weiter
  // oben. **Ein Sitz ist ein Aufbau und kein Spiel** (CR 100.1, Fassung
  // 2026-08-19), und genau das steht dann da.
  const spielart =
    wuensche.length === 1 ? eigene.spielartGoldfish : eigene.spielartSelfplay

  // **Das Format kommt aus dem ersten Deck**, wie das Spiel — es steht am
  // Deck und wird nicht daneben gewählt.
  const format = gewaehlt?.formatID ?? ''

  // **Eines genügt**: dasselbe Deck darf an zwei Sitzen liegen, und ein Sitz
  // allein ist eine Spielart (Goldfish).
  // **Kein Deck heißt: angekommen und keines** (UX-08.NAT-01) — solange die
  // Antwort aussteht, steht das da, und ein Fehlschlag bietet den nächsten
  // Versuch an statt „lege eines an".
  // **Ein Anfang zur Zeit** — während er läuft, steht der Knopf still und
  // sagt es; ein zweiter Klick legt keine zweite Partie an (`PartieStart09`).
  const bereit = vollstaendig && !lauf.laeuft('partie-anfangen')
  const anfangen = () => {
    // **Der Anfang verbraucht, was mitgereist ist** (P6, #2374).
    if (mitgereist !== undefined) verbrauche()
    lauf.anfangen({
      spiel,
      // **Die Vorgaben des Kerns bleiben Vorgaben.** Ein leeres
      // Regelwerk heißt Sandkasten, ein leerer Wirt der erste Sitz —
      // beides entscheidet `internal/app` an **einer** Stelle für
      // alle vier Flächen (§16).
      //
      // **Die Abwahl aber ist eine Wahl und keine Vorgabe** — P1a
      // (#2335).
      // Hier stand bis zum 2026-10-13 `[]`: die Tafel zeigte die
      // Zugstruktur als manuell, und die Partie fror sie trotzdem
      // als automatisch ein. Der Modus geht nicht mit — er erlaubt
      // die Abwahl nur, und der Anfängermodus räumt sie
      // (`Regelwahlfeld`); eingefroren wird, was abgewählt ist.
      regelwerk: {
        kennung: '',
        fassung: 0,
        strenge: '',
        abgewaehlt: [...abgewaehlt],
      },
      wirt: '',
      sitze: wuensche.map((wunsch, stelle) => ({
        kennung: sitzkennung(stelle),
        name: wunsch.name,
        deckID: wunsch.deckID,

        // **Eine angenommene Leihe statt eines Decks** — P1b
        // (#2370); bis dahin stand hier immer eine leere. Ob sie gilt,
        // prüft der Kern (G6, #1463).
        leiheID: wunsch.leiheID,

        profilID: wunsch.profilID,

        // **Ein eingeladener Gast bringt sein Deck mit** — P9c (#2377).
        eingeladen: wunsch.eingeladen ?? '',
      })),
    })
  }

  // **Zwei Tafeln und der Anfang darunter** — das Artboard *Eine Partie
  // anfangen* (PLAY-01, PLAY-02, S11.NAT-01): wer sitzt, nach welchen Regeln,
  // und dann der Knopf.
  // **Schnell spielen fängt von selbst an** (P6, #2374) — einmal je Wunsch,
  // sobald die Vorbelegung dasteht, und nur, was vollständig ist; sonst
  // steht das Formular da und sagt, was fehlt.
  const erledigt = useRef(0)
  useEffect(() => {
    if (schnell > erledigt.current) {
      erledigt.current = schnell
      if (bereit) anfangen()
    }
  })

  // **Und die Vorbelegung wird abgewartet** — sonst stünden die Felder einen
  // Augenblick leer und füllten sich dann (P6).
  if (!decks.geladen || quellen === undefined) {
    return decks.ladefehler === undefined ? (
      <Wirdgeladen />
    ) : (
      <Ladefehler fehler={decks.ladefehler} wiederholen={decks.neuLaden} />
    )
  }

  // **Und der nächste Schritt ist ein Knopf** — UXR-13
  // (#1921).
  // Am Probebestand stand die Seite ohne Hauptknopf da (R1 in
  // `docs/UIREVIEW.md`), und „Lege eines unter Decks an" verlangte einen Weg,
  // den man sich merken muss (`Leerzustand.tsx`).
  if (decks.decks.length === 0) {
    return (
      <Leerzustand titel={eigene.leerDecksTitel} erklaerung={eigene.leerDecks}>
        <Knopf
          gewicht="haupt"
          aufKlick={() => {
            wechseln('decks')
          }}
        >
          {eigene.leerDecksZiel}
        </Knopf>
      </Leerzustand>
    )
  }

  // **Die Zeile über Spiel, Spielart und Format** — sie steht bei den
  // Sitzen, denn aus dem ersten Deck kommt das Spiel.
  const auskunft =
    spiel === ''
      ? ''
      : [
          `${eigene.spiel}: ${spiel}`,
          `${eigene.spielart}: ${spielart}`,
          ...(format === '' ? [] : [`${eigene.format}: ${format}`]),
        ].join(' · ')

  return (
    // **⌘↩ fängt an** — von jedem Feld des Formulars aus (`PartieStart09`).
    // eslint-disable-next-line jsx-a11y/no-static-element-interactions -- ein Kürzel für das ganze Formular, der Knopf bleibt das Bedienelement
    <section
      className="flex max-w-4xl flex-col gap-4"
      onKeyDown={(ereignis) => {
        if (ereignis.key === 'Enter' && ereignis.metaKey && bereit) {
          ereignis.preventDefault()
          anfangen()
        }
      }}
    >
      {mitgereist !== undefined && (
        <p className="max-w-prose text-sm">
          {einsetzen(eigene.anschlussWartet, { name: mitgereist.name })}
        </p>
      )}

      <Sitzliste
        wuensche={wuensche}
        setzeWunsch={setzeWunsch}
        hinzufuegen={() => {
          setzeWuensche((bisher) => [
            ...bisher,
            { name: '', deckID: '', profilID: '', leiheID: '' },
          ])
        }}
        wegnehmen={(stelle) => {
          setzeWuensche((bisher) => bisher.filter((_, lage) => lage !== stelle))
        }}
        decks={decks.decks}
        leihen={leihen}
        kontakte={kontakte}
        profile={profile}
        auskunft={auskunft}
      />

      {/* **Was durchgesetzt wird, steht vor dem Anfangen** (G4, #1461).
          Erst wenn ein Deck gewählt ist, gibt es ein Spiel — und erst dann
          eine Antwort, die zu etwas gehört. */}
      {regeln.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {regeln.fehler}
        </p>
      )}

      {regeln.freigabe !== undefined && (
        <Regeltafel
          freigabe={regeln.freigabe}
          modus={modus}
          setzeModus={setzeModus}
          abgewaehlt={abgewaehlt}
          setzeAbgewaehlt={setzeAbgewaehlt}
        />
      )}

      <div>
        {/*
          **Was schiefging, steht beim Knopf** — CLAUDE.md §6: am Seitenkopf
          stünde die Meldung fern vom Klick, und für den Klickenden geschähe
          sichtbar nichts.

          **Nur die eigene Meldung**: `fehlerVorgang` sagt, wozu sie gehört.
          Eine, die vom Einsteigen in eine fremde Partie stammt, gehört
          nicht an diesen Knopf.
        */}
        {lauf.fehler !== undefined &&
          lauf.fehlerVorgang === 'partie-anfangen' && (
            <p role="alert" className="mb-2 max-w-prose text-sm text-fehler">
              {lauf.fehler}
            </p>
          )}

        {fehlt === undefined ? undefined : (
          <p className="mb-2 max-w-prose text-sm text-leise">{fehlt}</p>
        )}

        {/* **Dass die Decks einfrieren, steht beim Knopf**, wie im Artboard
            — es ist das Letzte, was man vor dem Klick wissen muss. */}
        <Erklaerungsabsatz
          text={eigene.anfangErklaerung}
          klasse="mb-2 max-w-prose text-sm text-leise"
        />

        <Knopf
          gewicht="haupt"
          schreibt
          kuerzel="Meta+Enter"
          aufKlick={anfangen}
          angehalten={!bereit}
        >
          {lauf.laeuft('partie-anfangen')
            ? eigene.anfangenLaeuft
            : eigene.anfangen}
        </Knopf>
      </div>
    </section>
  )
}
