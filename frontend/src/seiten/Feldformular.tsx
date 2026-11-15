// **Wie die Felder einer Karte angeordnet sind** —
// #1271,
// seit dem 2026-09-21 in drei Tiefen (UX-05.AK-01,
// #1677).
//
// **Eine eigene Datei**, weil `Karteneditor.tsx` mit ihr über 400 Zeilen ginge
// (§2). Der Schnitt folgt der Frage: dort steht, **was** das Fenster tut —
// laden, schreiben, abbrechen, den Stand halten —, hier, **wie** die Felder
// dabei stehen.
//
// # Der Befund, der zu den zwei Gruppen führte
//
// Gemessen am 2026-08-28 im laufenden Fenster, an einer Karte von
// `Vanguard.mse-set`: **16 Felder, 1.683 px hoch in einem 900-px-Fenster** —
// zwei Bildschirme für eine Karte. Und die letzten vier hießen `stylesheet`,
// `stylesheet version`, `has styling` und `styling data`.
//
// # Was diese vier sind — nachgelesen, nicht erschlossen
//
//	StyleSheetP stylesheet;
//	IndexMap<FieldP,ValueP> styling_data;
//	/// Is the styling_data set?
//
// `src/data/card.hpp`, Zeilen 44–50, Fassung f9d9356 vom 2021-07-06: sie sind
// **Mitglieder von `Card`** und keine `card field:` des Spiels —
// `magic.mse-game/card_fields` erklärt keines von ihnen. Sie stehen im
// Formular, weil dieses Projekt sie als Feldwerte ablegt.
//
// **Das ist das Kriterium**, und es ist kein Namensmuster: ein Feld, das das
// Spielpaket nicht erklärt, ist keine Karteneigenschaft, die jemand beim
// Gestalten ausfüllt. Es bleibt erreichbar — ein Set aus MTG JSON trägt solche
// Felder, und sie zu verstecken hieße, sie stillschweigend zu verlieren
// (`angebotenefelder.ts`).
//
// # Und warum daraus drei wurden
//
// **Weil zwischen „was jemand ausfüllt" und „was die Anwendung dazu weiß"
// eine dritte Gruppe liegt**: Felder, die das Spielpaket erklärt und ein
// Skript füllt. Sie sagen, **warum** etwas auf der Karte steht, und
// standen bisher zwischen Kartenname und Regeltext. Die Zuordnung und ihre
// Quellen stehen in `feldtiefe.ts`.
import type { ReactNode } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { Knopf } from '../bausteine/Knopf'
import {
  feldtiefen,
  stehtIn,
  tiefeVon,
  verdeckterBefund,
  wirksameTiefe,
} from './feldtiefe'
import { useTexte } from '../sprache/kontext'

import type { Befundstelle, Feldtiefe } from './feldtiefe'
import type { FeldbeschreibungDTO } from '../api'
import type { Angebotenesfeld } from './angebotenefelder'

interface FeldformularEigenschaften {
  felder: readonly Angebotenesfeld[]

  // nachName sagt, ob das Spielpaket ein Feld erklärt — und damit, ob es
  // Karteninhalt ist oder Buchführung.
  nachName: ReadonlyMap<string, FeldbeschreibungDTO>

  // tiefe ist der gewählte Bedienmodus.
  tiefe: Feldtiefe

  // probleme sind die Befunde an dieser Karte — für den Hinweis auf einen,
  // der gerade verdeckt liegt (AK-02).
  probleme: readonly Befundstelle[]

  // aufTiefe springt in eine andere Tiefe.
  aufTiefe: (tiefe: Feldtiefe) => void

  // zeile zeichnet **ein** Feld. **Ein Rückruf und keine zwanzig
  // Eigenschaften**: was eine Feldzeile braucht, weiß der Editor; diese Datei
  // entscheidet nur, wo sie steht (§2).
  zeile: (feld: Angebotenesfeld) => ReactNode

  // unruhig verbirgt das Formular, bis seine Felder stehen (#2301,
  // `formularruhe.ts`) — unsichtbar, aber im Baum und erreichbar: ein
  // Vorleser hört `aria-busy`, und nichts verliert seinen Fokus.
  unruhig?: boolean

  // formular ist die Referenz auf das `<form>` — der Editor sucht darin das
  // Eingabefeld eines angeklickten Kartenbereichs.
  formular: React.RefObject<HTMLFormElement | null>
}

// spalten ist das Gitter, in dem Felder nebeneinander stehen.
//
// **Und `spaltenMitAbstand` daneben statt einer Zeichenkette mit Platzhalter
// im JSX**: `keintextimjsx.test.ts` kann einem Ausdruck nicht ansehen, dass
// er kein Oberflächentext ist — und eine Ausnahme dafür machte den Wächter
// an der Stelle blind, an der er gebraucht wird.
//
// **Ab 32 rem dieser Spalte zwei Spalten**: darunter ist eine breite besser
// als zwei schmale — in ein Feld von 176 Punkten passt kein Regeltext.
const spalten =
  'flex flex-col gap-3 @lg:grid @lg:grid-cols-2 @lg:items-start @lg:gap-x-6'

const spaltenMitAbstand = `mt-2 ${spalten}`

// ruhigesFormular und unruhigesFormular sind das Formular sichtbar und
// verborgen (#2301) — als Konstanten aus demselben Grund wie die Spalten.
const ruhigesFormular = 'flex w-full flex-col gap-3'
const unruhigesFormular = `${ruhigesFormular} opacity-0`

export function Feldformular({
  felder,
  nachName,
  tiefe: gewaehlt,
  probleme,
  aufTiefe,
  zeile,
  formular,
  unruhig = false,
}: FeldformularEigenschaften) {
  const eigene = useTexte().seiten.karten

  // **Was gilt, ist nicht immer, was gewählt ist** — der Grund steht bei
  // [wirksameTiefe].
  const tiefe = wirksameTiefe(gewaehlt, felder, nachName)

  const inTiefe = (gesucht: Feldtiefe) =>
    felder.filter((feld) => tiefeVon(nachName.get(feld.feld)) === gesucht)

  const verdeckt = verdeckterBefund(tiefe, probleme, nachName)

  return (
    <form
      ref={formular}
      aria-busy={unruhig || undefined}
      className={unruhig ? unruhigesFormular : ruhigesFormular}
      onSubmit={(ereignis) => {
        ereignis.preventDefault()
      }}
    >
      {/* **Der Wechsler steht über dem Formular** — UX-05.AK-01. Er gehört
          dorthin, wo die Felder stehen, und nicht auf die Einstellungsseite:
          dort steht, womit man **anfängt**; hier wechselt man beim Arbeiten.

          **Drei Knöpfe und keine Auswahlliste**: drei Werte sind mit einem
          Blick zu sehen, und `aria-pressed` sagt einem Vorleseprogramm,
          welcher gilt. */}
      <fieldset className="flex flex-wrap items-baseline gap-2">
        <legend className="text-sm font-medium">{eigene.tiefe}</legend>

        {feldtiefen.map((eine) => (
          <Knopf
            key={eine}
            gewicht="leise"
            gedrueckt={eine === tiefe}
            aufKlick={() => {
              aufTiefe(eine)
            }}
          >
            {eigene[tiefenwort(eine)]}
          </Knopf>
        ))}
      </fieldset>

      {/* **Ein Befund unter einer verdeckten Tiefe steht oben** — AK-02.
          Er sagt, **wo** er liegt, und bietet den Weg dorthin an; ein
          Hinweis ohne Handlung ließe jemanden suchen. */}
      {verdeckt !== undefined && (
        <p
          role="status"
          className="flex flex-wrap items-center gap-2 rounded-sm border border-kante px-3 py-2 text-sm"
        >
          {einsetzen(eigene.befundVerdeckt, {
            tiefe: eigene[tiefenwort(verdeckt)],
          })}

          <Knopf
            aufKlick={() => {
              aufTiefe(verdeckt)
            }}
          >
            {eigene.befundSpringen}
          </Knopf>
        </p>
      )}

      {/* **Der Karteninhalt zuerst, in der Leseordnung der Karte** — die
          Reihenfolge kommt aus `angebotenefelder.ts` und ist dort begründet. */}
      <div className={spalten}>{inTiefe('normal').map(zeile)}</div>

      <Gruppe
        titel={eigene.tiefeErweitertTitel}
        hinweis={eigene.tiefeErweitertHinweis}
        steht={stehtIn(tiefe, 'erweitert')}
        felder={inTiefe('erweitert')}
        zeile={zeile}
      />

      <Gruppe
        titel={eigene.buchfuehrung}
        hinweis={eigene.buchfuehrungHinweis}
        steht={stehtIn(tiefe, 'experte')}
        felder={inTiefe('experte')}
        zeile={zeile}
      />
    </form>
  )
}

// Gruppe ist ein Block tieferer Felder mit seiner Überschrift.
//
// **Ohne solche Felder steht gar nichts da** — eine leere Überschrift wäre
// eine Zeile, die erklärt werden müsste.
function Gruppe({
  titel,
  hinweis,
  steht,
  felder,
  zeile,
}: {
  titel: string
  hinweis: string
  steht: boolean
  felder: readonly Angebotenesfeld[]
  zeile: (feld: Angebotenesfeld) => ReactNode
}) {
  if (!steht || felder.length === 0) {
    return null
  }

  return (
    <section
      aria-label={titel}
      className="flex flex-col gap-1 rounded-sm border border-kante px-3 py-2"
    >
      <h3 className="text-sm font-medium">{titel}</h3>

      <p className="max-w-prose text-sm text-leise">{hinweis}</p>

      <div className={spaltenMitAbstand}>{felder.map(zeile)}</div>
    </section>
  )
}

// tiefenwort gibt den Katalogschlüssel zu einer Tiefe.
function tiefenwort(
  tiefe: Feldtiefe,
): 'tiefeNormal' | 'tiefeErweitert' | 'tiefeExperte' {
  if (tiefe === 'normal') {
    return 'tiefeNormal'
  }

  if (tiefe === 'erweitert') {
    return 'tiefeErweitert'
  }

  return 'tiefeExperte'
}
