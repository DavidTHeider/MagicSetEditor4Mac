// **Der Deckbauplatz** — Werkstück U10 von Phase 5.1
// (#1421).
//
// # Drei Bereiche statt einer langen Seite
//
// **Links suchen, in der Mitte bauen, rechts ansehen.** Bis U10 stand alles
// untereinander: Anlage, Holen, Deckliste, und darunter das offene Deck mit
// vier Laschen — die Seite war 1.318 Punkte hoch (gemessen bei 1440 × 900),
// und wer die Auswertung sehen wollte, scrollte an allem vorbei.
//
// **Bei engem Fenster weicht der Inspector**, nicht der Navigator: ohne die
// Deckliste käme man an kein zweites Deck; ohne die Analyse baut man weiter.
// Dieselbe Wahl wie am Sammlungsarbeitsplatz (U8).
import { Ladefehler } from '../bausteine/Ladehinweise'
import { useEffect, useState } from 'react'

import { useAnschluss, useAufgenommen } from '../anschlussraum'
import { Arbeitsplatz } from '../bausteine/Arbeitsplatz'
import { Deckeinfuhr } from '../bausteine/Deckeinfuhr'
import { Handlungsanmeldung } from '../bausteine/Handlungsanmeldung'
import { Formatverwaltung } from '../bausteine/Formatverwaltung'
import { Hinweiskasten } from '../bausteine/Hinweiskasten'
import {
  Deckbauinspector,
  Deckbaunavigator,
  Deckbauzentrum,
} from './deckbaubereiche'
import { Knopf } from '../bausteine/Knopf'
import { Schublade } from '../bausteine/Schublade'
import { Seitenkopf } from '../bausteine/Seitenkopf'
import { Netzwege } from './Netzwege'
import { useArbeitsplatzgedaechtnis } from '../haken/useArbeitsplatzgedaechtnis'
import { useDeckabgleich } from '../haken/useDeckabgleich'
import { useDeckhandlung } from '../haken/useDeckhandlung'
import { useDecks } from '../haken/useDecks'
import { useTexte } from '../sprache/kontext'

import type { Seitenschluessel } from '../navigation'

// **Die Anschlüsse im Kopf** — UX-02
// (#1674).
//
// Seit die Seitenleiste **Bereiche** führt, hat *Spielen* einen Eintrag und
// drei Seiten. Der Weg zu den anderen beiden steht hier, so wie der Entwurf
// ihn zeichnet: `<a class="knopf" href="Turniere.dc.html">` und
// `<a class="knopf haupt" href="PartieStart.dc.html">Partie beginnen</a>` im
// Kopf des Artboards *Decks*. Ohne ihn verlören zwei Seiten ihren sichtbaren
// Zugang.
//
// **Und die Partie heißt hier „Partie anfangen" und nicht „Spielen".** Der
// Eintrag der Leiste trägt denselben Namen wie die Seite; zwei Knöpfe
// „Spielen" auf einem Bildschirm sind für ein Vorleseprogramm nicht zu
// unterscheiden — der Lauf daneben hat es sofort gezeigt.
interface Deckseigenschaften {
  wechseln: (seite: Seitenschluessel) => void

  // bestandsstand zählt fremde Schreibvorgänge — die Liste lädt dann nach
  // (Befund 3 des Gesamtreviews über 0.6.0).
  bestandsstand?: number
}

export function Decks({ wechseln, bestandsstand = 0 }: Deckseigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.decks

  const stand = useDecks(bestandsstand)
  const gedaechtnis = useArbeitsplatzgedaechtnis()
  const handlung = useDeckhandlung(stand)

  // **Die Menüleiste klappt den Navigator auf** (#2469) — die Zeile, an der
  // eine Handlung steht, lebt dort. Gesagt wird, ob er zu war.
  const navigatorZeigen = () => {
    if (gedaechtnis.navigator) {
      return false
    }

    gedaechtnis.merke('navigator', true)

    return true
  }
  const [formateOffen, setzeFormateOffen] = useState(false)

  // **Aus der Befehlspalette: ein Deck, beim Namen gefunden** — es steht
  // aufgeschlagen da (UX-02.AK-02, #1674). **In einem Effekt**: Aufschlagen
  // lädt das Deck, und ein Ruf nach draußen gehört nicht ins Zeichnen.
  const anschluss = useAnschluss()
  const aufzuschlagen =
    anschluss.offen?.art === 'deckansicht' ? anschluss.offen : undefined
  const { waehle } = stand
  useEffect(() => {
    if (aufzuschlagen !== undefined) {
      void waehle(aufzuschlagen.deckID)
    }
  }, [aufzuschlagen, waehle])

  // **Und mit seiner Lasche** — AP-8 (#2365), B-14: die Aufmerksamkeit führt
  // wegen fehlender Karten her, und die stehen in der Prüfung. **Der Wunsch
  // gehört zu diesem Deck** (§7, Regel 1): ein danach aufgeschlagenes
  // anderes geht auf, wie es aufgeht. Die Nummer lässt denselben Weg ein
  // zweites Mal wirken (`Laschen.oeffne`).
  const [laschenwunsch, setzeLaschenwunsch] = useState<{
    deckID: string
    oeffne: string
  }>()
  const [wuensche, setzeWuensche] = useState(0)
  useAufgenommen(aufzuschlagen, ({ deckID, lasche }) => {
    if (lasche !== undefined) {
      setzeWuensche(wuensche + 1)
      setzeLaschenwunsch({
        deckID,
        oeffne: `${lasche}#${String(wuensche + 1)}`,
      })
    }
  })
  const oeffne =
    laschenwunsch?.deckID === stand.offen?.id
      ? laschenwunsch?.oeffne
      : undefined

  // **Was vom offenen Deck in der Sammlung liegt** — einmal gefragt, von der
  // Leiste und der Lasche „Prüfung“ gelesen (AP-9, #2366).
  const abgleich = useDeckabgleich(stand.offen, bestandsstand)

  // **Die Leiste führt in die Prüfung** — auch, wenn der Inspector
  // eingeklappt ist: der Klick ist der Wunsch, ihn zu sehen, und wird
  // gemerkt wie ein Klick auf „Inspector einblenden“. Bei Enge geht er über
  // dem Zentrum auf (`inspectorAnlass`).
  const [inspectorAnlass, setzeInspectorAnlass] = useState(0)
  const zeigePruefung = (deckID: string) => {
    if (!gedaechtnis.inspector) {
      gedaechtnis.merke('inspector', true)
    }
    setzeWuensche(wuensche + 1)
    setzeLaschenwunsch({ deckID, oeffne: `pruefung#${String(wuensche + 1)}` })
    setzeInspectorAnlass(inspectorAnlass + 1)
  }

  return (
    <div className="flex flex-col gap-4">
      <Seitenkopf titel={eigene.titel} erklaerung={eigene.erklaerung}>
        <Knopf
          aufKlick={() => {
            wechseln('turniere')
          }}
        >
          {texte.seiten.turniere.titel}
        </Knopf>

        {/* **Die Formate in einer Schublade** — AP-6 (#2363). Sie lagen
            hinter einer Lasche neben den Decks, und das Laschenpaar zählte
            zweimal zu den Bedienelementen im Bild (R2); `Bereiche09` hat
            keines. **Im Kopf und nicht an der Formatwahl**: die steht nur im
            offenen Deck, und ohne Deck kämen niemand mehr an die Formate. */}
        <Knopf
          aufKlick={() => {
            setzeFormateOffen(true)
          }}
        >
          {eigene.formateKnopf}
        </Knopf>

        {/* **Im Nebengewicht** — AP-3 (#2360): er führt auf eine andere
            Seite, und die Decks haben ihre eigene führende Handlung — „Deck
            anlegen“, im offenen Deck „Ins Deck“ (`Bereiche09`). */}
        <Knopf
          aufKlick={() => {
            wechseln('spielen')
          }}
        >
          {texte.seiten.spielen.anfangen}
        </Knopf>

        {/* **Und die zwei kurzen Wege der Netzpartie** — P9c (#2377). */}
        <Netzwege decks={stand.decks} />
      </Seitenkopf>

      {stand.ladefehler !== undefined && (
        <Ladefehler fehler={stand.ladefehler} wiederholen={stand.neuLaden} />
      )}

      {/* **Ein Konflikt bietet den heutigen Stand an** (UX-08.AK-02,
          #1680) — „Bitte neu laden" ohne einen Knopf dazu schickte suchen. */}
      {stand.fehler !== undefined && (
        <Hinweiskasten
          art="fehler"
          meldung={stand.fehler}
          handlung={
            stand.konflikt
              ? {
                  beschriftung: texte.bausteine.zustaende.neuLaden,
                  aufKlick: stand.nachKonflikt,
                }
              : undefined
          }
        />
      )}

      <Arbeitsplatz
        bezeichnung={eigene.deckplatz}
        gedaechtnis={gedaechtnis}
        beiEnge="inspector"
        inspectorAnlass={inspectorAnlass}
        navigator={{
          // **Er heißt, was er zeigt** — seit die Kartensuche in der
          // Mitte steht (UXR-11, #1919), stehen hier die Decks.
          bezeichnung: eigene.titel,
          inhalt: (
            <Deckbaunavigator
              stand={stand}
              handlung={handlung}
              wechseln={wechseln}
            />
          ),
        }}
        zentrum={
          <Deckbauzentrum
            stand={stand}
            abgleich={abgleich}
            zeigePruefung={zeigePruefung}
          />
        }
        inspector={{
          bezeichnung: eigene.analyse,
          inhalt: (
            <Deckbauinspector
              stand={stand}
              abgleich={abgleich}
              oeffne={oeffne}
            />
          ),
        }}
      />

      {/* **Das Menü „Deck“ in der Menüleiste, solange ein Deck offen ist**
          (#2469) — hier und nicht an der Zeile: die fehlt, wenn die Suche
          sie nicht trifft oder der Navigator zu ist. Dieselbe Liste wie
          Überlauf und Rechtsklick (Zweitblick über Phase 6, Befund 6). */}
      {stand.offen !== null && (
        <Handlungsanmeldung
          titel={eigene.deckmenue}
          art="deck"
          kennung={stand.offen.id}
          eintraege={handlung.eintraege(stand.offen, navigatorZeigen)}
        />
      )}

      {stand.offen !== null && handlung.einfuhrOffen && (
        <Schublade
          titel={eigene.einfuhrTitel}
          schliessenBezeichnung={
            texte.allgemein.arbeitsplatz.schubladen.schliessen
          }
          schliessen={handlung.schliesseEinfuhr}
        >
          <div className="flex flex-col gap-4 px-4 pb-5 pt-3">
            <Deckeinfuhr deck={stand.offen} gelegt={stand.neuLaden} />
          </div>
        </Schublade>
      )}

      {formateOffen && (
        <Schublade
          titel={eigene.gebietFormate}
          schliessenBezeichnung={
            texte.allgemein.arbeitsplatz.schubladen.schliessen
          }
          schliessen={() => {
            setzeFormateOffen(false)
          }}
        >
          <div className="flex flex-col gap-4 px-4 pb-5 pt-3">
            <Formatverwaltung stand={stand} />
          </div>
        </Schublade>
      )}
    </div>
  )
}
