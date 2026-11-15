// **Was du jetzt tun kannst** — TRADE-02.AK-02
// (#1716).
//
// # Warum hier nur die Schritte stehen, die gehen
//
// **Weil der Vertrag sie nennt.** Bis TRADE-02 standen alle neun Knöpfe da,
// mit der Begründung, dass ein Fenster, das die Matrix nachrechnete, eine
// zweite Fassung von ihr hätte. Das gilt weiter — und deshalb rechnet es
// nicht: `moegliche` kommt aus `domain.Handelsangebot.Moegliche`, und dieses
// Bauteil zeigt, was dort steht. Gemessen hat die Probe `trade02` vorher
// zwanzig Knöpfe über drei Zustände, die nur einen Fehler hätten geben
// können.
//
// **Der übliche nächste Schritt ist der Hauptknopf**, die Nebenwege stehen
// daneben. **Die drei des Abschlusses nicht**: sie bewegen Bestand und Beleg
// und stehen mit ihrer Freigabe in der Tafel „Verbuchen".
//
// # Der Hauptknopf ist ein Zug — P13
//
// **Er geht, was seine Entscheidung umfasst, und bringt es zum Gegenüber**
// (#2381,
// Vermerk 0200): „Anbieten und zustimmen" sagt es selbst, „Vorbereiten und übergeben"
// hält das eine passende Exemplar fest, und wer die Übergabe als Letzter
// bestätigt, verbucht im selben Zug. Sein Name sagt es. Die Nebenwege —
// Abbrechen, Ablehnen, Gegenangebot, Klären — bleiben einzelne Schritte.
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useState } from 'react'
import { useTexte } from '../sprache/kontext'
import { verlaesst } from '../handelszug'
import { Knopf } from './Knopf'

import { dieLeitungsschritte } from '../apihandelsleitung'

import type { Handelsvorgang as Vorgangsstand } from '../apihandelsleitung'
import type { Zugschritt } from '../handelszug'
import type { Handelszug } from '../haken/useHandelszug'
import type { Vorgang } from '../haken/useVorgang'

// istZugbeginn sagt, ob ein Zug mit diesem Leitungsschritt beginnt.
function istZugbeginn(schritt: string): schritt is Zugschritt {
  return ['anbieten', 'zustimmen', 'vorbereiten', 'uebergeben'].includes(
    schritt,
  )
}

interface Handelsschritteeigenschaften {
  kennung: string

  stand: Vorgangsstand
  geaendert: (neu: Vorgangsstand) => void

  // gebucht lädt nach einem Zug, der gebucht hat, alles nach — Journal und
  // Beleg eingeschlossen; seine Antwort ist dann kein ganzer Stand.
  gebucht: () => void

  // zug ist der des Raums — er zeigt auch, was unterwegs hängen blieb.
  zug: Handelszug

  // lauf ist der Riegel des Raums — derselbe wie im Handelsvorgang (§7,
  // Regel 4): ein Schritt und ein Vorschlag an demselben Handel gehen nicht
  // zugleich.
  lauf: Vorgang
}

type Leitungsschritt = (typeof dieLeitungsschritte)[number]

export function Handelsschritte({
  kennung,
  stand,
  geaendert,
  gebucht,
  zug,
  lauf,
}: Handelsschritteeigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.handeln
  const [fehler, setzeFehler] = useState<string>()

  const kennzeichen = `handelsvorgang/${kennung}`

  const saetze: Record<Leitungsschritt, string> = {
    anbieten: eigene.schrittAnbieten,
    gegenangebot: eigene.schrittGegenangebot,
    zustimmen: eigene.schrittZustimmen,
    vorbereiten: eigene.schrittVorbereiten,
    uebergeben: eigene.schrittUebergeben,
    ablehnen: eigene.schrittAblehnen,
    abbrechen: eigene.schrittAbbrechen,
    ablaufen: eigene.schrittAblaufen,
    klaeren: eigene.schrittKlaeren,
  }

  const gehbar = stand.moegliche.filter((schritt): schritt is Leitungsschritt =>
    (dieLeitungsschritte as readonly string[]).includes(schritt),
  )

  // **Der Name sagt, was der Zug geht** — und nicht nur, womit er beginnt.
  const saetzeFuer = (schritt: Leitungsschritt): string => {
    if (schritt === 'anbieten') {
      return eigene.zug.anbietenZustimmen
    }

    if (schritt === 'vorbereiten') {
      return eigene.zug.vorbereitenUebergeben
    }

    if (schritt === 'uebergeben' && verlaesst(schritt, stand, stand.ich)) {
      return eigene.zug.uebergebenVerbuchen
    }

    return saetze[schritt]
  }

  // ziehe geht den Zug des Hauptknopfs. **Was gebucht hat, lädt alles nach**
  // (Journal, Beleg); sonst trägt der Raum den Stand ein, den der Zug zuletzt
  // las. Ging gar nichts, bleibt alles, wie es war.
  // bisFuer sagt, wie weit der Knopf geht — **so weit, wie sein Name sagt**
  // (Befund 1 des Zweitblicks über Phase 8): „Übergabe bestätigen" verbucht
  // nicht, auch wenn die Stimme des Gegenübers inzwischen da ist.
  const bisFuer = (schritt: Zugschritt): Zugschritt => {
    if (schritt === 'anbieten' || schritt === 'zustimmen') {
      return 'zustimmen'
    }

    return schritt === 'uebergeben' && verlaesst(schritt, stand, stand.ich)
      ? 'bestaetigen'
      : 'uebergeben'
  }

  const ziehe = (schritt: Zugschritt) => {
    void zug.ziehen({ stand, schritt, bis: bisFuer(schritt) }, (neu) => {
      const gebuchtHat = neu.gegangen.some((einer) =>
        ['verbuchen', 'abgleichen', 'bestaetigen'].includes(einer),
      )

      // **Ein veralteter Stand wird gezeigt** — der Mensch entscheidet neu.
      if (gebuchtHat || neu.gescheitert?.teil === 'lesen') {
        gebucht()
      } else if (
        neu.gegangen.length > 0 ||
        neu.gescheitert?.teil === 'veraltet'
      ) {
        geaendert(neu.stand)
      }
    })
  }

  const schreite = (schritt: Leitungsschritt) => {
    void lauf.starte(kennzeichen, async () => {
      setzeFehler(undefined)

      try {
        // **An der Revision, die zu sehen war** (#2496).
        const antwort = await api().HandelSchritt(
          stand.vorgang,
          schritt,
          '',
          stand.abdruck,
        )
        geaendert(auspacken(antwort, () => antwort.vorgang))
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  return (
    <div className="flex flex-col gap-2">
      {gehbar.length === 0 ? (
        <p className="text-sm text-leise">{eigene.assistent.keinSchritt}</p>
      ) : (
        <div
          role="group"
          aria-label={eigene.assistent.schritte}
          className="flex flex-wrap gap-2"
        >
          {gehbar.map((schritt) =>
            schritt === stand.naechster && istZugbeginn(schritt) ? (
              <Knopf
                key={schritt}
                gewicht="haupt"
                angehalten={lauf.laeuft(kennzeichen)}
                aufKlick={() => {
                  ziehe(schritt)
                }}
              >
                {saetzeFuer(schritt)}
              </Knopf>
            ) : (
              <Knopf
                key={schritt}
                gewicht={schritt === stand.naechster ? 'haupt' : 'leise'}
                angehalten={lauf.laeuft(kennzeichen)}
                aufKlick={() => {
                  schreite(schritt)
                }}
              >
                {saetze[schritt]}
              </Knopf>
            ),
          )}
        </div>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}
    </div>
  )
}
