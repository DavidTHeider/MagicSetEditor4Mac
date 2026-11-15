// **Ein neues, leeres Set** —
// #891.
//
// # Was MSE2 dafür fragt
//
// Zwei Listen: **Spiel** und **Stil**, die zweite gefiltert auf `<spiel>-*`
// (`src/gui/new_window.cpp`, `NewSetWindow`, Fassung f9d9356 vom 2021-07-06).
// Dieselben zwei Angaben stehen hier — dazu der **Name**, den MSE2 erst beim
// Sichern verlangt und der hier von Anfang an dazugehört.
//
// # adapted: drei Schritte statt eines Fensters
//
// **Nach dem Artboard *Sets*** — „Ein neues Set, in drei Schritten": Art,
// Aussehen, Name (UX-05, S07.NAT-03, #1677).
// Wer nur einen Namen will, **überspringt**: Spiel und Stil stehen
// vorgewählt, wie MSE2 sie vorwählt (unten). **Zurück behält alles** — die
// Angaben stehen hier und nicht in den Schritten, die kommen und gehen.
// **Und ⌘N springt zum Namen** — wer es drückt, will anlegen, nicht wählen.
// Wo das Anlegen steht — offen oder in einer Schublade —, sagt `Setanlage`;
// der erste Schritt steht in `SetNeuArt.tsx`, der zweite ist die
// `Stilauswahl` wie auf der Stilseite eines offenen Sets (UXR-04, #1912).
import { useEffect, useId, useRef, useState } from 'react'

import { api, auspacken } from '../api'
import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Knopf } from '../bausteine/Knopf'
import { Ladefehler, Wirdgeladen } from '../bausteine/Ladehinweise'
import { Schrittfolge } from '../bausteine/Schrittfolge'
import { Stilauswahl } from '../bausteine/Stilauswahl'
import { useLaden } from '../haken/useLaden'
import { useVorgabe } from '../haken/useVorgabe'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { nachRang } from '../paketordnung'
import { Pflichtfeld } from '../bausteine/Pflichtfeld'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { SetNeuArt } from './SetNeuArt'

import type { PaketDTO } from '../api'

interface SetNeuEigenschaften {
  // angelegt bekommt die Kennung des neuen Sets — wer eines anlegt, will es
  // aufschlagen.
  angelegt: (setID: string) => void

  // anstoss springt zum Namen und setzt den Fokus hinein (⌘N).
  anstoss?: boolean
  anstossVerbraucht?: (() => void) | undefined

  // zuDenPaketen führt dorthin, wo ein fehlendes Spiel oder ein fehlender
  // Stil herkommt — UX-05 S07.NAT-04. **Zurück geht es über die
  // Seitengeschichte** (⌘[), und diese Seite bleibt dabei stehen.
  zuDenPaketen: () => void

  // ueberschriftVerborgen: in der Schublade nennt ihr Titel sie (AP-10).
  ueberschriftVerborgen?: boolean
}

type Schritt = 'art' | 'aussehen' | 'name'

export function SetNeu({
  angelegt,
  anstoss = false,
  anstossVerbraucht,
  zuDenPaketen,
  ueberschriftVerborgen = false,
}: SetNeuEigenschaften) {
  const texte = useTexte()
  const worte = texte.seiten.sets
  const vorgang = useVorgang()

  const titelID = useId()
  const namensfeld = useRef<HTMLInputElement>(null)
  const schrittkopf = useRef<HTMLHeadingElement>(null)

  const [name, setzeName] = useState('')
  const [spiel, setzeSpiel] = useState('')
  const [stil, setzeStil] = useState('')
  const [gesetzt, setzeSchritt] = useState<Schritt>()
  const [fehler, setzeFehler] = useState<string>()

  // **Die Pakete sind die Quelle** — nicht eine Liste hier.
  const pakete = useLaden(holePakete, [])
  const alle = pakete.daten ?? []

  // **In MSE2s Ordnung** — `position hint:`, dann der Name. Der erste ist
  // der vorgewählte; nach dem Namen allein war das für Magic „DayNight".
  const spiele = alle.filter((paket) => paket.art === 'mse-game').sort(nachRang)

  // **Der Stil hängt am Spiel** — `<spiel>-*`, wie in MSE2.
  const stileVon = (eines: string) =>
    alle
      .filter(
        (paket) =>
          paket.art === 'mse-style' && paket.name.startsWith(eines + '-'),
      )
      .sort(nachRang)

  // **Gewählt wird nur, was nutzbar ist** — UX-05.AK-04: nutzbar heißt, es
  // gibt einen Stil dazu (`Set::validate`, `src/data/set.cpp`, Fassung
  // f9d9356 vom 2021-07-06). Hat keines einen, bleibt die Wahl über allen.
  const nutzbare = spiele.filter((eines) => stileVon(eines.name).length > 0)
  const waehlbar = nutzbare.length > 0 ? nutzbare : spiele
  const namen = waehlbar.map((eines) => eines.name)

  // **Vorgewählt wie in MSE2** — das Spiel aus `settings.default_game`
  // (`src/gui/new_window.cpp`, Zeile 53), hier die Vorgabe des Profils; der
  // Stil aus `default_stylesheet` (Zeile 66). **Abgeleitet und nicht
  // gesetzt**: ein Spielwechsel zieht den Stil mit, ohne dass ein Effekt eine
  // Zeichnung zu spät nachführt.
  const vorgabespiel = useVorgabe('spiel')[0] ?? ''
  const gewaehltesSpiel = namen.includes(spiel)
    ? spiel
    : namen.includes(vorgabespiel)
      ? vorgabespiel
      : (namen[0] ?? '')

  const stile = stileVon(gewaehltesSpiel)
  const stilnamen = stile.map((einer) => einer.name)
  const vorgabevorlage = useVorgabe('vorlage')[0] ?? ''
  const gewaehlterStil = stilnamen.includes(stil)
    ? stil
    : stilnamen.includes(vorgabevorlage)
      ? vorgabevorlage
      : (stilnamen[0] ?? '')

  // **Ein Spiel ist kein Schritt** — UX-05.AK-04: gibt es nur eines, entfällt
  // seine Wahl, und aus drei Schritten werden zwei.
  const erster: Schritt = waehlbar.length > 1 ? 'art' : 'aussehen'
  const schritte: readonly Schritt[] =
    erster === 'art' ? ['art', 'aussehen', 'name'] : ['aussehen', 'name']
  const schritt =
    gesetzt !== undefined && schritte.includes(gesetzt) ? gesetzt : erster
  const stelle = schritte.indexOf(schritt)

  // **Was davor und danach kommt** — ohne einen Vorgänger gibt es kein
  // Zurück, ohne einen Nachfolger kein Weiter, sondern das Anlegen.
  const vorher = schritte[stelle - 1]
  const nachher = schritte[stelle + 1]

  // **Wohin der Fokus soll, sobald das Ziel dasteht** — der Kopf eines
  // Schritts nach „Weiter" und „Zurück", das Namensfeld nach „Überspringen"
  // und nach ⌘N. **Ein Ziel und kein sofortiger Ruf**: ⌘N kommt oft, bevor die
  // Pakete da sind, und dann gibt es das Feld noch nicht. **Beim ersten
  // Zeichnen gibt es keines** — wer die Seite aufschlägt, hat nichts gedrückt.
  const [fokusziel, setzeFokusziel] = useState<'kopf' | 'name'>()

  // **⌘N springt zum Namen** — beim Zeichnen und nicht in einem Effekt, damit
  // der Schritt schon steht, wenn der Effekt darunter den Fokus setzt. **Auch,
  // wenn die Seite mit dem Anstoß erst entsteht**: der Vergleich beginnt bei
  // „kein Anstoß".
  const [letzterAnstoss, setzeLetzterAnstoss] = useState(false)
  if (anstoss !== letzterAnstoss) {
    setzeLetzterAnstoss(anstoss)
    if (anstoss) {
      setzeSchritt('name')
      setzeFokusziel('name')
    }
  }

  useEffect(() => {
    if (anstoss) {
      anstossVerbraucht?.()
    }
  }, [anstoss, anstossVerbraucht])

  const geladen = pakete.daten !== undefined
  const erfuellt = useRef('')
  // **Nur auf einen neuen Wunsch**, nicht beim Einblenden (G11, `useRueckkehr`)
  // — und jeder Anstoß ist einer: ein zweiter *Neues Set* findet das Feld.
  useEffect(() => {
    const wunsch = `${String(fokusziel)}/${schritt}/${String(geladen)}/${String(anstoss)}`
    if (erfuellt.current === wunsch) return
    erfuellt.current = wunsch
    if (fokusziel === 'name') {
      namensfeld.current?.focus()
    } else if (fokusziel === 'kopf') {
      schrittkopf.current?.focus()
    }
  }, [fokusziel, schritt, geladen, anstoss])

  const gehe = (ziel: Schritt) => {
    setzeSchritt(ziel)
    setzeFokusziel(ziel === 'name' ? 'name' : 'kopf')
  }

  const anlegen = () => {
    void vorgang.starte('anlegen', async () => {
      setzeFehler(undefined)

      try {
        const antwort = await api().SetAnlegen(
          name,
          gewaehltesSpiel,
          gewaehlterStil,
        )
        const kennung = auspacken(antwort, () => antwort.id)

        setzeName('')
        setzeSchritt(undefined)
        setzeFokusziel(undefined)
        angelegt(kennung)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  const schrittnamen = {
    art: worte.neuSchrittArt,
    aussehen: worte.neuSchrittAussehen,
    name: worte.neuSchrittName,
  }

  const spielpaket = waehlbar.find((eines) => eines.name === gewaehltesSpiel)
  const stilpaket = stile.find((einer) => einer.name === gewaehlterStil)

  return (
    <section className="flex flex-col gap-3" aria-labelledby={titelID}>
      <h2
        id={titelID}
        className={ueberschriftVerborgen ? 'sr-only' : 'text-lg font-medium'}
      >
        {worte.neuTitel}
      </h2>

      <Erklaerungsabsatz
        text={worte.neuHinweis}
        klasse="max-w-prose text-sm text-leise"
      />

      {/* **Eine ausstehende Liste ist kein Stand** (UX-08.NAT-01): solange die
          Pakete unterwegs sind, stünden sonst zwei Schritte da, wo es drei
          gibt — und nach einem Fehlschlag gar keine Wahl. */}
      {pakete.daten === undefined ? (
        pakete.fehler === undefined ? (
          <Wirdgeladen />
        ) : (
          <Ladefehler fehler={pakete.fehler} wiederholen={pakete.nachladen} />
        )
      ) : spiele.length === 0 ? (
        <Paketweg satz={worte.neuOhneSpiele} zuDenPaketen={zuDenPaketen} />
      ) : (
        <>
          <Schrittfolge
            bezeichnung={worte.neuSchritte}
            schritte={schritte.map((einer) => schrittnamen[einer])}
            aktuell={stelle}
          />

          <h3
            ref={schrittkopf}
            tabIndex={-1}
            className="text-base font-medium focus:outline-none"
          >
            {schrittnamen[schritt]}
          </h3>

          {schritt === 'art' && (
            <SetNeuArt
              spiele={waehlbar}
              stileJeSpiel={(eines) => stileVon(eines).length}
              gewaehlt={gewaehltesSpiel}
              waehlen={(eines) => {
                setzeSpiel(eines)
                // **Ein Stil gehört zu einem Spiel** (§7, Regel 1).
                setzeStil('')
              }}
              zuDenPaketen={zuDenPaketen}
            />
          )}

          {schritt === 'aussehen' &&
            (gewaehltesSpiel !== '' && stile.length === 0 ? (
              <Paketweg
                satz={einsetzen(worte.neuOhneStile, {
                  spiel: gewaehltesSpiel,
                })}
                zuDenPaketen={zuDenPaketen}
              />
            ) : (
              <Stilauswahl
                beschriftung={worte.neuStil}
                eintraege={stile.map((einer) => ({
                  kennung: einer.name,
                  anzeigename: einer.anzeigename,
                }))}
                gewaehlt={gewaehlterStil}
                waehlen={setzeStil}
              />
            ))}

          {schritt === 'name' && (
            <>
              <Pflichtfeld
                beschriftung={worte.neuName}
                wert={name}
                griff={namensfeld}
                autofokus
                geaendert={setzeName}
              />

              {spielpaket !== undefined && stilpaket !== undefined && (
                <p className="text-sm text-leise">
                  {einsetzen(worte.neuWahl, {
                    spiel: spielpaket.anzeigename,
                    stil: stilpaket.anzeigename,
                  })}
                </p>
              )}
            </>
          )}

          <div className="flex flex-wrap items-center gap-2">
            {vorher !== undefined && (
              <Knopf
                aufKlick={() => {
                  gehe(vorher)
                }}
              >
                {worte.neuSchrittZurueck}
              </Knopf>
            )}

            {nachher === undefined ? (
              <Aktionsknopf
                aktion="anlegen"
                gewicht="haupt"
                angehalten={
                  name.trim() === '' ||
                  gewaehlterStil === '' ||
                  vorgang.laeuft('anlegen')
                }
                aufKlick={anlegen}
              >
                {worte.neuAnlegen}
              </Aktionsknopf>
            ) : (
              <>
                <Knopf
                  gewicht="leise"
                  aufKlick={() => {
                    gehe('name')
                  }}
                >
                  {worte.neuUeberspringen}
                </Knopf>
                <Knopf
                  gewicht="haupt"
                  aufKlick={() => {
                    gehe(nachher)
                  }}
                >
                  {worte.neuWeiter}
                </Knopf>
              </>
            )}
          </div>
        </>
      )}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}
    </section>
  )
}

// holePakete liest den Paketkatalog.
async function holePakete(): Promise<PaketDTO[]> {
  const antwort = await api().Pakete()

  return auspacken(antwort, () => antwort.pakete)
}

// Paketweg sagt, was fehlt, und führt dorthin, wo es herkommt — UX-05
// S07.NAT-04. **Mit dem Satz über den Rückweg**, denn ein Knopf, der
// wegführt, ohne zu sagen, wie man zurückkommt, lässt ein halb ausgefülltes
// Formular zurück, das niemand wiederfindet.
function Paketweg({
  satz,
  zuDenPaketen,
}: {
  satz: string
  zuDenPaketen: () => void
}) {
  const worte = useTexte().seiten.sets

  return (
    <div className="flex max-w-prose flex-col items-start gap-2 text-sm">
      <p className="text-leise">{satz}</p>
      <Knopf aufKlick={zuDenPaketen}>{worte.neuZuDenPaketen}</Knopf>
      <p className="text-leise">{worte.neuZurueck}</p>
    </div>
  )
}
