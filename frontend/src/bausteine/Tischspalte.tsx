// **Was neben dem Tisch steht** — Meldungen, der Zug, die Spieler, die
// Fragen des Spiels und die Befehle an der gewählten Karte. Herausgelöst aus
// `Partietisch.tsx`, als P7
// (#2375)
// ihn mit dem Vorhang über die Zeilengrenze aus §2 brachte.
//
// **Die Spalte gehört zur Sicht** — sie trägt, was nur der gezeigte Sitz
// sehen darf, und geht deshalb beim Wechsel und hinter dem Vorhang mit.
import { Knopf } from './Knopf'
import { Mischpflichten } from './Mischpflichten'
import { Partiebefehle } from '../seiten/Partiebefehle'
import { Spielerleiste } from './Spielerleiste'
import { Tischbefunde } from './Tischbefunde'
import { Tischfragen } from './Tischfragen'
import { Tischmeldungen } from './Tischmeldungen'
import { Zugleiste } from './Zugleiste'
import { partievorgang } from '../haken/partiewege'
import { Zuruecknehmen } from '../haken/useTischruecknahme'
import { useTexte } from '../sprache/kontext'

import type { Partie, Spielobjekt } from '../apipartie'
import type { Partielauf } from '../haken/usePartie'
import type { useSpieltexturen } from '../haken/useSpieltexturen'

export function Tischspalte({
  lauf,
  partie,
  objekt,
  texturen,
  entwaehle,
  kannZuruecknehmen,
}: {
  lauf: Partielauf
  partie: Partie

  // kannZuruecknehmen sagt, ob der Kern die Rücknahme anbietet — angemeldet
  // wird ⌘Z am Tisch (`useTischruecknahme`), nicht hier: die Spalte geht
  // beim Sichtwechsel und hinter dem Vorhang mit.
  kannZuruecknehmen: boolean

  // objekt ist die gewählte Karte — `undefined`: keine.
  objekt: Spielobjekt | undefined
  texturen: ReturnType<typeof useSpieltexturen>
  entwaehle: () => void
}) {
  const eigene = useTexte().seiten.spielen

  return (
    <div className="flex min-w-0 flex-col gap-4 @4xl:max-h-(--tischhoehe) @4xl:w-80 @4xl:shrink-0 @4xl:overflow-y-auto">
      <Tischmeldungen fehler={lauf.fehler} wartend={lauf.wartend} />

      {/* **Wer am Zug ist, steht neben dem Tisch** (G9, #1592) — oben in
        der Spalte, denn es ist die Auskunft, nach der man als Erstes
        sucht. Ohne laufenden Zug zeichnet die Leiste nichts; eine
        Partie aus G1 bis G8 ist ein Sandkasten und bleibt es. */}
      <Zugleiste
        partie={partie}
        erlaubt={lauf.erlaubt}
        laeuft={lauf.laeuft(partievorgang(partie.id))}
        passen={() => {
          lauf.befehlen('zug.passen', {})
        }}
        abkuerzen={() => {
          lauf.befehlen('zug.abkuerzen', {})
        }}
        zugende={() => {
          lauf.befehlen('zug.beenden', {})
        }}
      />

      {/* **Die Rücknahme steht unter dem Zug** — Entscheidung zu #2526: über
        der Zugleiste schob sie den Zug im schmalen Fenster unter die Falz,
        und der Kommentar darüber sagte schon, dass der Zug oben steht. */}
      {kannZuruecknehmen && (
        <Knopf
          gewicht="leise"
          kuerzel="Meta+Z"
          aufKlick={() => {
            lauf.befehlen(Zuruecknehmen, {})
          }}
        >
          {eigene.zuruecknehmenKnopf}
        </Knopf>
      )}

      {/*
      **Der Stand der Spieler steht unter dem Takt** (M3, #1605) —
      Leben, Vergiftungsmarken und Manavorrat. Sie liegen seit G9 am
      Modell und kamen im Fenster nicht vor; eine Regel, die greift,
      ohne dass jemand sieht, worauf, ist die erste Dimension aus §10.

      **Und sie bleibt, wo die Zugleiste verschwindet**: ein Sandkasten
      aus G1 bis G8 hat keinen Zug, aber Leben hat er.
    */}
      <Spielerleiste
        partie={partie}
        {...(lauf.erlaubt.includes('spieler.leben')
          ? {
              aendereLeben: (um: number) => {
                lauf.befehlen('spieler.leben', { um })
              },
            }
          : {})}
      />

      {/* **Was das Spiel anhält, steht über der Karte** — der Beginn,
        die Legendenwahl und die Reihung; warum an einer Stelle, steht
        in `Tischfragen`. */}
      <Tischfragen lauf={lauf} partie={partie} />

      <Mischpflichten partie={partie} />

      <Partiebefehle
        lauf={lauf}
        partie={partie}
        objekt={objekt}
        texturen={texturen}
        entwaehle={entwaehle}
      />

      <Tischbefunde
        sitzbefunde={lauf.befunde}
        texturfehler={texturen.fehler}
        texturbefunde={texturen.befunde}
      />
    </div>
  )
}
