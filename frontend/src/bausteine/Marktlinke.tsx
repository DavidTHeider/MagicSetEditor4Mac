// **Wohin eine Karte außerhalb dieses Programms führt** — Werkstück T8 von
// Phase 5.6
// (#1473).
//
// # Warum an jeder Zeile steht, wie genau sie trifft
//
// **Weil ein Suchlink nicht hält, was ein Produktlink hält.** „Bei CardTrader
// ansehen" führt bei dem einen auf **diese** Karte und bei dem anderen auf
// eine Liste, in der sie vermutlich steht. Wer beides gleich beschriftete,
// verspräche eine Genauigkeit, die er nicht hat — und das Issue verlangt
// ausdrücklich eine „ehrliche Beschriftung".
//
// # Und warum hier Knöpfe stehen und keine Verweise
//
// **Weil dieses Programm den Browser öffnet und nicht das Fenster.** Ein
// `<a href>` im Fenster einer Wails-Anwendung führte in dieses Fenster; was
// hinaus soll, geht über den Plattformöffner — und der prüft die Adresse ein
// zweites Mal.
import { useId, useState } from 'react'

import { Knopf } from './Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'
import { dieGenaueArt, dieStartseite } from '../apimarktlinke'

import type { Marktlink } from '../apimarktlinke'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

interface Marktlinkeeigenschaften {
  // printing ist die Druckkennung — leer für eine eigene Karte.
  printing: string

  // name trägt die Suche.
  name: string
}

export function Marktlinke({ printing, name }: Marktlinkeeigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten
  const lauf = useVorgang()
  const ueberschrift = useId()

  // **Der Fehler des Öffnens steht neben dem des Ladens und nicht darin.**
  // `useLaden` hält den Stand einer **Frage**; ein Klick ist keine.
  const [oeffnungsfehler, setzeOeffnungsfehler] = useState<string>()

  const links = useLaden(async () => {
    const antwort = await api().Marktlinks(printing, name)

    return auspacken(antwort, () => antwort.links)
  }, [printing, name])

  const oeffnen = (link: Marktlink): Promise<void> =>
    lauf.starte(`marktlink/${link.anbieter}`, async () => {
      setzeOeffnungsfehler(undefined)

      try {
        const antwort = await api().MarktlinkOeffnen(
          link.anbieter,
          link.adresse,
        )

        auspacken(antwort, () => undefined)
      } catch (ursache) {
        setzeOeffnungsfehler(meldungAus(ursache, texte))
      }
    })

  return (
    <section aria-labelledby={ueberschrift} className="flex flex-col gap-2">
      <h3 id={ueberschrift} className="text-sm font-medium">
        {eigene.marktTitel}
      </h3>

      <Erklaerungsabsatz
        text={eigene.marktErklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {/* **Ein gescheitertes Öffnen ist eine Meldung, ein gescheitertes
          Lesen eine Notiz.**

          Der Unterschied ist, wer gehandelt hat: wer eben geklickt hat,
          wartet auf eine Antwort und bekommt sie angesagt; wer nur eine
          Karte aufgeschlagen hat, hat nichts getan, was misslingen könnte —
          und ein `alert` neben den Preisen daneben riefe zweimal dasselbe
          aus. */}
      {oeffnungsfehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {oeffnungsfehler}
        </p>
      )}

      {links.fehler !== undefined && (
        <p role="status" className="text-sm text-leise">
          {links.fehler}
        </p>
      )}

      {links.daten !== undefined && (
        <ul
          aria-labelledby={ueberschrift}
          className="flex flex-col gap-1 text-sm"
        >
          {links.daten.map((link) => (
            <li
              key={link.anbieter}
              className="flex flex-wrap items-center gap-2"
            >
              <Knopf
                gewicht="leise"
                angehalten={lauf.laeuft(`marktlink/${link.anbieter}`)}
                aufKlick={() => {
                  void oeffnen(link)
                }}
              >
                {link.anbieter}
              </Knopf>

              {/* **Die Genauigkeit steht dran** — und die Vorgabe sagt
                  nichts: `produkt` ist, was man erwartet, und steht
                  deshalb nicht da. */}
              <span className="text-xs text-leise">
                {link.art === dieGenaueArt
                  ? eigene.marktGenau
                  : link.art === dieStartseite
                    ? eigene.marktNurAnbieter
                    : eigene.marktSuche}
              </span>
            </li>
          ))}
        </ul>
      )}
    </section>
  )
}
