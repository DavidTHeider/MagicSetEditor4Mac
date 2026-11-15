// **Vorige und nächste Karte** —
// #1290.
//
// # ported, nicht erfunden (§15)
//
// MSE2 hat diese beiden Befehle im Kartenmenü, mit ⇞ und ⇟:
//
//	previous card:		Select &Previous Card	PgUp
//	next card:			Select &Next Card	PgDn
//
// `data/en.mse-locale/locale`, Zeilen 48–49, Fassung f9d9356 vom 2021-07-06;
// die Einträge selbst in `src/gui/set/cards_panel.cpp`, Zeilen 68–69, die
// Kennungen in `src/util/window_id.hpp`, Zeilen 98–99. Was sie tun, steht in
// [nachbarkarte].
//
// # Und warum es sie **sichtbar** gibt und nicht nur als Kürzel
//
// Weil MSE2 eine Menüleiste hat, in der ein Befehl mit seinem Kürzel steht —
// dieses Fenster hat für die Kartenlasche keine. Ein Kürzel, das nirgends
// steht, findet niemand (#245
// hat denselben Satz für ⌘1…⌘8 aufgeschrieben).
//
// # Die Stelle steht dabei
//
// **„12 von 107" ist keine Zierde.** Wer in Serie arbeitet, muss sehen, wo er
// ist; „nächste" allein sagt es nicht. MSE2 zeigt es in der Liste, die dort
// immer sichtbar ist — hier lässt sie sich wegklappen (#697), und dann bleibt
// nichts.
//
// # Am Rand abgeschaltet, nicht weg
//
// MSE2 lässt den Menüeintrag stehen und schaltet ihn ab
// (`ev.Enable(card_list->canSelectNext())`, `cards_panel.cpp`, Zeile 239). Ein
// Knopf, der **verschwindet**, verschiebt den daneben — und wer zweimal auf
// dieselbe Stelle klickt, trifft beim zweiten Mal etwas anderes.

import type { ReactNode } from 'react'

import type { KartenzeileDTO } from '../api'
import { Knopf } from '../bausteine/Knopf'
import { useTexte } from '../sprache/kontext'
import { nachbarkarte } from './kartennachbar'

interface BlaetterEigenschaften {
  karten: readonly KartenzeileDTO[]
  gewaehlt: string | undefined
  waehlen: (karteID: string) => void
}

export function Kartenblaettern({
  karten,
  gewaehlt,
  waehlen,
}: BlaetterEigenschaften) {
  const texte = useTexte()
  const worte = texte.seiten.karten.blaettern

  const vorige = nachbarkarte(karten, gewaehlt, 'zurueck')
  const naechste = nachbarkarte(karten, gewaehlt, 'vor')

  const stelle = karten.findIndex((karte) => karte.id === gewaehlt)

  return (
    <div className="flex items-center gap-2">
      <Blaetterknopf
        ziel={vorige}
        bezeichnung={`${worte.vorige} (${worte.kuerzelVorige})`}
        waehlen={waehlen}
      >
        ←
      </Blaetterknopf>

      {/* **`aria-live` steht hier nicht.** Die Stelle ändert sich mit jedem
          Blättern, und eine Vorlesestimme, die bei jedem Tastendruck „13 von
          107" sagt, überdeckt den Kartennamen, auf den es ankommt. */}
      <span className="text-sm tabular-nums text-leise">
        {stelle < 0
          ? ''
          : worte.stelle
              .replace('{stelle}', String(stelle + 1))
              .replace('{gesamt}', String(karten.length))}
      </span>

      <Blaetterknopf
        ziel={naechste}
        bezeichnung={`${worte.naechste} (${worte.kuerzelNaechste})`}
        waehlen={waehlen}
      >
        →
      </Blaetterknopf>
    </div>
  )
}

// Blaetterknopf ist einer der beiden Knöpfe — mit oder ohne Ziel.
//
// **Zwei Zweige und kein Wächter im Rumpf.** Ein `if (ziel !== undefined)`
// **im** Klickrumpf wäre ein Zweig, den kein Lauf betreten kann: der Knopf ist
// dann abgeschaltet, und ein abgeschalteter Knopf gibt kein Klickereignis. Hier
// sind beide Zweige das, was ein Lauf sieht — abgeschaltet oder nicht (§3).
function Blaetterknopf({
  ziel,
  bezeichnung,
  waehlen,
  children,
}: {
  ziel: string | undefined
  bezeichnung: string
  waehlen: (karteID: string) => void
  children: ReactNode
}) {
  if (ziel === undefined) {
    return (
      <Knopf bezeichnung={bezeichnung} angehalten>
        {children}
      </Knopf>
    )
  }

  return (
    <Knopf
      bezeichnung={bezeichnung}
      aufKlick={() => {
        waehlen(ziel)
      }}
    >
      {children}
    </Knopf>
  )
}
