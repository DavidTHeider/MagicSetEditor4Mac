// **Was in einer Zone liegt** — herausgelöst aus `Partiezonen.tsx` für
// PLAY-04.AK-01
// (#1712).
//
// # Warum eine eigene Datei
//
// **Weil es zwei Fragen sind** (§2): *welche Zonen gibt es und in welcher
// Reihenfolge* steht nebenan; hier steht, *wie eine Karte darin aussieht und
// was man an ihr tun kann*. Der Anlass war wie beim ersten Schnitt die
// Zeilengrenze — mit dem Kontextmenü ging die eine Datei auf 388 Zeilen —,
// der Schnitt folgt aber der Frage.
import type { KeyboardEvent } from 'react'

import { Kartenbild } from '../bausteine/Kartenbild'
import { Kontextmenue } from '../bausteine/Kontextmenue'
import { einzeltaste } from '../bausteine/Tischkartenknopf'
import { bildwunsch } from './tischbild'
import { einsetzen } from '../sprache/einsetzen'
import { handlungZu, type Klick } from '../spiel/kartengriff'
import { kartenbeschriftung, menueZurKarte } from '../spiel/kartenmenue'
import { useTexte } from '../sprache/kontext'
import { wirkungsbild } from '../tisch/tischwirkungen'

import type {
  Kartenwert,
  Objekthandlung,
  Sitz,
  Spielobjekt,
} from '../apipartie'
import type { Gestenziel } from '../bausteine/gestenvertrag'
import type { Geste } from '../haken/useGeste'
import type { Texturlauf } from '../haken/useSpieltexturen'
import type { Wirkungsart } from '../tisch/tischwirkungen'

// Zonenkarten zeigt, was in einer Zone liegt.
//
// **Jedes Stück ein Knopf.** Ein Klick löst die Primärhandlung aus, ⌘-Klick
// wählt (P4, #2372) — mit der Tastatur wie mit der Maus (§6).
export function Zonenkarten({
  objekte,
  gewaehlt,
  waehle,
  texturen,
  spiel,
  sitze,
  wirkungen,
  handlungen,
  greife,
  tastenFuer,
  taste,
  befehlen,
  geste,
  zieleFuer,
}: {
  objekte: Spielobjekt[]
  gewaehlt: string | undefined
  waehle: (kennung: string) => void
  texturen: Texturlauf
  spiel: string
  sitze: Sitz[]
  wirkungen: Map<string, Wirkungsart>
  // handlungen sind je Objekt, was an ihm geht; greife ist der Klick daran —
  // Primär-, Zweithandlung oder Auswahl (P4, #2372).
  handlungen: readonly Objekthandlung[]
  greife: (kennung: string, klick: Klick) => void
  // tastenFuer und taste sind die Einzeltasten einer Karte (C-09).
  tastenFuer: (kennung: string) => string
  taste: (kennung: string, taste: string) => boolean
  befehlen: (art: string, nutzlast: Record<string, unknown>) => void
  // geste ist der Griff der ganzen Liste (`Partiezonen`), zieleFuer die
  // Zonen, in die eine Karte darf — `undefined`, wenn der Kern kein Schieben
  // erlaubt (GESTE, #2358).
  geste: Geste
  zieleFuer: (objekt: Spielobjekt) => readonly Gestenziel[] | undefined
}) {
  const texte = useTexte()
  const eigene = texte.seiten.spielen

  if (objekte.length === 0) {
    return <span className="text-sm text-leise">{eigene.zoneLeer}</span>
  }

  return (
    <ul className="flex flex-wrap gap-1">
      {objekte.map((objekt) => (
        <Kontextmenue
          key={objekt.kennung}
          element="li"
          {...menueZurKarte(
            eigene,
            texte.seiten.handlungen,
            objekt,
            handlungZu(handlungen, objekt.kennung),
            befehlen,
            waehle,
          )}
        >
          <button
            type="button"
            aria-pressed={objekt.kennung === gewaehlt}
            onClick={(ereignis) => {
              greife(objekt.kennung, {
                detail: ereignis.detail,
                metaKey: ereignis.metaKey,
              })
            }}
            {...(tastenFuer(objekt.kennung) === ''
              ? {}
              : { 'aria-keyshortcuts': tastenFuer(objekt.kennung) })}
            onKeyDown={(ereignis) => {
              if (
                einzeltaste(ereignis, (wort) => taste(objekt.kennung, wort))
              ) {
                ereignis.preventDefault()

                return
              }

              tastenzug(ereignis, objekt, geste, zieleFuer)
            }}
            data-gegriffen={gegriffen(geste, objekt) || undefined}
            className={kartenklassen(
              objekt.kennung === gewaehlt,
              gegriffen(geste, objekt),
            )}
            {...wirkungsmarke(wirkungen.get(objekt.kennung))}
          >
            {/* **Was gerade geschah, steht an der Karte** — PLAY-04.AK-02
                (#1712). Das Zeichen ist die zweite Unterscheidung neben der
                Farbe (WCAG 1.4.1); **was** geschehen ist, steht als Satz in
                der Ereignisliste darunter, und deshalb ist es hier
                Schmuck. */}
            {zeichenVon(wirkungen.get(objekt.kennung))}
            <Kartenbild
              texturen={texturen}
              wunsch={bildwunsch(objekt, spiel, 'tisch')}
              ausweichguete="tisch-niedrig"
            />

            <span className="block max-w-20 truncate">
              {kartenbeschriftung(eigene, objekt)}
            </span>

            {/* **Was die Karte ist, steht an ihr** (M2, #1604): die Typzeile
                (CR 205.1) und die **gelebten** Werte (CR 122.1a). Gerechnet
                wird in der Domäne; hier steht, was ankommt. */}
            <span className="block max-w-20 truncate text-leise">
              {typtext(objekt)}
            </span>

            <span className="block max-w-20 truncate text-leise">
              {wertetext(objekt)}
            </span>

            {/* **Wer es beherrscht, steht daran** — CR 110.2a (M12,
                #1614).
                Die Zonenliste mischt die Karten aller Sitze; ohne diese
                Zeile ist ein Kontrollwechsel geschehen und **nirgends zu
                sehen**.

                **Nur auf Tisch und Stapel** — CR 109.4: außerhalb dieser
                beiden hat ein Objekt keinen Beherrscher, und eine Zeile,
                die dort einen nennte, behauptete etwas Falsches. */}
            {objekt.zone === 'tisch' || objekt.zone === 'stapel' ? (
              <span className="block max-w-20 truncate text-leise">
                {sitzname(sitze, objekt.beherrscher)}
              </span>
            ) : undefined}

            {/* **Ein Befund ist kein Fehler** (§8): die Karte steht da, nur
                dieses eine Wort kennt CR 205.2a nicht. */}
            <span className="block max-w-20 truncate text-leise">
              {befundtext(eigene, objekt)}
            </span>
          </button>
        </Kontextmenue>
      ))}
    </ul>
  )
}

// gegriffen sagt, ob diese Karte gerade in der Hand der Tastatur ist.
function gegriffen(geste: Geste, objekt: Spielobjekt): boolean {
  return (
    geste.zustand.art === 'gegriffen' && geste.zustand.objekt === objekt.kennung
  )
}

// tastenzug deutet eine Taste an einer Karte — GESTE (#2358). **Eine
// gegriffene Karte nimmt Pfeile, Eingabe, Leertaste und Escape**; die übrigen
// gehen ihren Weg. Ungegriffen greift nur die Leertaste, und nur, wenn es ein
// Ziel gibt — sonst bleibt sie, was sie an einem Knopf ist: wählen.
function tastenzug(
  ereignis: KeyboardEvent<HTMLButtonElement>,
  objekt: Spielobjekt,
  geste: Geste,
  zieleFuer: (objekt: Spielobjekt) => readonly Gestenziel[] | undefined,
) {
  if (gegriffen(geste, objekt)) {
    if (geste.taste(ereignis.key)) {
      ereignis.preventDefault()
      ereignis.stopPropagation()
    }

    return
  }

  const ziele = ereignis.key === ' ' ? zieleFuer(objekt) : undefined
  if (ziele === undefined) {
    return
  }

  ereignis.preventDefault()
  geste.greifen({ objekt: objekt.kennung, quelle: objekt.zone, ziele })
}

// wirkungsmarke gibt das Merkmal, an dem eine Wirkung hängt — oder nichts.
//
// **Als Merkmal und nicht als Klasse**: ein Lauf kann es lesen, und die Farbe
// kommt aus derselben Tabelle wie die der Leinwand. Zwei Quellen für dasselbe
// Aussehen liefen auseinander (§10).
function wirkungsmarke(art: Wirkungsart | undefined) {
  if (art === undefined) return {}

  return {
    'data-wirkung': art,
    style: { outline: `2px solid #${wirkungsbild(art).farbe.toString(16)}` },
  }
}

// zeichenVon gibt das Zeichen einer Wirkung — oder nichts.
//
// **`aria-hidden`, und das ist eine Entscheidung**: ein Vorleseprogramm
// bekäme sonst an jeder bewegten Karte ein Dreieck vorgelesen. Was geschehen
// ist, steht als **Satz** in der Ereignisliste (`Partiebefehle`), und die ist
// die Fläche, auf der es angesagt gehört.
function zeichenVon(art: Wirkungsart | undefined) {
  if (art === undefined) return undefined

  return (
    <span aria-hidden className="block text-xs">
      {wirkungsbild(art).zeichen}
    </span>
  )
}

// kartenklassen gibt die Klassen eines Kartenknopfs.
//
// **Außerhalb des JSX**, weil `keintextimjsx.test.ts` jede Zeichenkette mit

// Platzhalter im Baum abweist — und zwar zu Recht: ein Textknoten mit
// `${…}` ist genau die Gestalt, in der ein Oberflächentext durchrutscht.
// Klassennamen sind keine Oberflächentexte, also stehen sie hier.
function kartenklassen(gewaehlt: boolean, inDerHand = false): string {
  const grund =
    'rounded-sm border px-2 py-1 text-xs focus-visible:outline ' +
    'focus-visible:outline-2 focus-visible:outline-offset-2 ' +
    'focus-visible:outline-ring '

  return (
    // **`border-marke` und nicht `border-ring`.** Es ist dieselbe Farbe,
    // aber `ring` ist der **Fokus**: ein Rand, der so aussieht wie er,
    // sagte einem Blick, hier stehe die Schreibmarke. `marke` ist das
    // Wort für „dieses hier ist gemeint" — und es steht in
    // `docs/KONTRASTE.md` mit gemessenen Werten für beide
    // Erscheinungsbilder.
    grund +
    (gewaehlt ? 'border-marke bg-flaeche font-semibold' : 'border-kante') +
    // **Gegriffen ist gestrichelt** — dieselbe Farbe wie gewählt, eine
    // zweite Unterscheidung neben ihr (WCAG 1.4.1), und keine neue Farbe.
    (inDerHand ? ' border-dashed border-marke' : '')
  )
}

// sitzname gibt den Namen zu einer Sitzkennung.
//
// **Die Kennung, wo keiner steht** — ein Bestand von Hand geändert, ein Sitz
// ausgetragen: der Tisch bleibt lesbar, statt eine leere Stelle zu zeigen
// (§8).
function sitzname(sitze: Sitz[], kennung: string): string {
  return sitze.find((sitz) => sitz.kennung === kennung)?.name ?? kennung
}

// typtext gibt die Typzeile einer Karte — leer, wo es keine gibt.
//
// **Mit dem Gedankenstrich der CR** (U+2014, CR 205.3a), und ohne ihn, wo
// keine Untertypen stehen.
//
// **Außerhalb des JSX**, aus demselben Grund wie [kartenklassen]: eine
// Zeichenkette mit Platzhalter im Baum ist die Gestalt, in der ein
// Oberflächentext durchrutscht. **Ein Kartentyp ist keiner** — er kommt aus
// dem Spielpaket und wird nicht übersetzt (§9).
function typtext(objekt: Spielobjekt): string {
  const eigen = objekt.eigenschaften
  if (!eigen) {
    return ''
  }

  const oben = [...eigen.supertypen, ...eigen.typen].join(' ')
  if (eigen.untertypen.length === 0) {
    return oben
  }

  return `${oben} — ${eigen.untertypen.join(' ')}`.trim()
}

// wertetext gibt Stärke und Widerstandskraft als `4/5`.
//
// **Leer, wo keine stehen** — CR 208.3: eine Nichtkreatur hat **keine**
// Stärke, und eine Null an ihrer Stelle wäre eine Aussage, die niemand
// gemacht hat.
function wertetext(objekt: Spielobjekt): string {
  const eigen = objekt.eigenschaften
  if (!eigen?.staerke.steht && !eigen?.widerstandskraft.steht) {
    return ''
  }

  return `${zahltext(eigen.staerke)}/${zahltext(eigen.widerstandskraft)}`
}

// zahltext gibt einen Wert als Text — **der Stern bleibt ein Stern**
// (CR 208.2).
function zahltext(wert: Kartenwert): string {
  return wert.stern ? '*' : String(wert.zahl)
}

// befundtext nennt die Wörter der Typzeile, die das Regelwerk nicht kennt.
//
// **Leer, wo alles aufgeht** — und das ist der Normalfall.
function befundtext(
  eigene: { unbekannterTyp: string },
  objekt: Spielobjekt,
): string {
  const unbekannt = objekt.eigenschaften?.unbekannteTypen ?? []
  if (unbekannt.length === 0) {
    return ''
  }

  return einsetzen(eigene.unbekannterTyp, { woerter: unbekannt.join(' ') })
}
