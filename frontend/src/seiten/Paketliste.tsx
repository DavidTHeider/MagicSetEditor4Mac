// Der Katalog der installierten Pakete (#54), als Verwaltung (#244).
//
// **Was sie tut, steht hier; wie sie aussieht, in den Klassen** (§2).
//
// # Was sich mit #244 geändert hat
//
// Vorher: eine Tabelle mit 329 Zeilen, ungefiltert, ungruppiert, ohne Suche.
// Das ist keine Verwaltung, das ist eine Liste — wer ein bestimmtes Paket
// suchte, scrollte.
//
// **Seit DOM-01 ein Arbeitsplatz** (#1696), gezeichnet nach dem Artboard
// `Pakete`: der Navigator sucht und filtert — auch nach der Herkunft —, das
// Zentrum trägt die Tabelle, der Inspector das gewählte Paket samt Woher,
// Abhängigen und lokaler Änderung. **Jede Änderung beginnt mit einem Plan**
// (`usePaketplan`), und er steht im Inspector, wo vorher nichts stand.
//
// Jetzt: gesucht, gefiltert, gruppiert und geordnet. **Die Gruppierung ist
// eine Portierung und keine Erfindung** — sie steht in den Paketdateien
// (`installer group:`), und MSE2 baut daraus denselben Baum
// (`src/gui/package_update_list.cpp`, Zeilen 18–75 und 200–212, Fassung
// f9d9356 vom 2021-07-06). Der Baum selbst steht in [paketbaum].
import { useMemo, useState } from 'react'

import type { AbhaengigkeitDTO, PaketDTO } from '../api'
import { useAnschluss, useAufgenommen } from '../anschlussraum'
import { Arbeitsplatz } from '../bausteine/Arbeitsplatz'
import { Aufklappkasten } from '../bausteine/Aufklappkasten'
import { Knopf } from '../bausteine/Knopf'
import { useArbeitsplatzgedaechtnis } from '../haken/useArbeitsplatzgedaechtnis'
import type { Ladezustand } from '../haken/useLaden'
import type { PaketeAntwort } from '../pakete'
import { usePaketplan } from '../haken/usePaketplan'
import { usePaketweitergabe } from '../haken/usePaketweitergabe'
import { einsetzen } from '../sprache/einsetzen'
import { useSprache, useTexte } from '../sprache/kontext'
import { Paketanlage } from './Paketanlage'
import { Katalogleer, Katalogohnetreffer } from './Paketleer'
import { Paketfilter } from './Paketfilter'
import { Paketinspector } from './Paketinspector'
import { Paketplanansicht } from './Paketplanansicht'
import { Pakettabelle } from './Pakettabelle'
import { sichtbareZeilen } from './paketbaum'
import { arten, leereWahl, sicht, type Sichtwahl } from './paketsicht'
import { Markenfilterknopf } from '../bausteine/Markenfilterkasten'
import { nachMarken, useMarkenfilter } from '../haken/useMarkenfilter'

interface Eigenschaften {
  // pakete sind, was die Seite gelesen hat (`usePakete`) — die Leiste
  // darüber zählt dieselben (AP-9). **Nachgelesen wird über eine Zahl und
  // keinen Rückruf**: die Übernahme (#55) legt Pakete ins Verzeichnis, und
  // wer stattdessen das Fenster neu lüde, verlöre jeden anderen Stand mit.
  pakete: Ladezustand<PaketeAntwort>

  // befunde sind die Abhängigkeitsbefunde, die die Seite gelesen hat — für
  // die Zahl an jeder Zeile (U14, #2159).
  befunde: AbhaengigkeitDTO[] | undefined
}

export function Paketliste({ pakete: stand, befunde }: Eigenschaften) {
  const texte = useTexte()
  const sprache = useSprache()
  const eigene = texte.seiten.pakete
  const gedaechtnis = useArbeitsplatzgedaechtnis()

  const { daten, laedt, fehler, nachladen } = stand

  // **Nach einer Änderung wird neu gelesen** — auch nach einer
  // gescheiterten; der Grund steht in `usePaketplan`.
  const planung = usePaketplan(nachladen)
  const weitergabe = usePaketweitergabe()

  // gewaehlt ist das Paket im Inspector, als `art/name`.
  const [gewaehlt, setzeGewaehlt] = useState<string>()

  const [wahl, setzeWahl] = useState<Sichtwahl>(leereWahl)

  // **Jede Wahl, jeder Plan und jede Absage ist ein Anlass für den
  // Inspector** — UXR-17 (#1925). Bei 1280 Punkten weicht er (`beiEnge`), und
  // ohne Anlass stand all das in einer Spalte, die niemand sieht. **Der Plan
  // zählt, wenn er da ist, nicht wenn er gefragt wird**: ein abgebrochener
  // Dialog öffnete sonst eine leere Spalte. Nachgezogen beim Zeichnen, wie
  // im `Aufklappkasten`.
  const [anlass, setzeAnlass] = useState(0)
  const zeigen = () => {
    setzeAnlass((zahl) => zahl + 1)
  }
  const [gemeldet, setzeGemeldet] = useState({
    plan: planung.plan,
    fehler: planung.fehler,
  })
  if (gemeldet.plan !== planung.plan || gemeldet.fehler !== planung.fehler) {
    setzeGemeldet({ plan: planung.plan, fehler: planung.fehler })

    if (planung.plan !== undefined || planung.fehler !== undefined) {
      zeigen()
    }
  }

  // **Ein Paket, das einem Set fehlt, wird gesucht** — DOM-01.AK-04. Ohne
  // Endung, denn gesucht wird im Namen; und die übrigen Filter weichen, damit
  // kein Artfilter den Treffer versteckt.
  const anschluss = useAnschluss()
  useAufgenommen(
    anschluss.offen?.art === 'paketsuche' ? anschluss.offen : undefined,
    (gesucht) => {
      // **Die Ordnung bleibt** — sie ist kein Filter, und wer sie gewählt
      // hat, will sie nach einer Suche nicht neu wählen (AP-9: die Leiste
      // nimmt über diesen Weg mit leerer Suche alle Filter zurück).
      setzeWahl({
        ...leereWahl,
        ordnung: wahl.ordnung,
        suche: gesucht.paket.replace(/\.mse-[a-z-]+$/u, ''),
      })
    },
  )
  // **Und die Pakete mit Befund aus der Aufmerksamkeit** — AP-8 (#2365):
  // genau sie, die übrigen Filter weichen wie oben.
  useAufgenommen(
    anschluss.offen?.art === 'paketauswahl' ? anschluss.offen : undefined,
    ({ name, namen }) => {
      setzeWahl({ ...leereWahl, auswahl: { name, namen } })
    },
  )
  // **Eingeklappt und nicht ausgeklappt — die Menge zählt die Ausnahmen.**
  //
  //	package_list->expandAll();
  //
  // `src/gui/packages_window.cpp`, Zeile 222, Fassung f9d9356 vom 2021-07-06:
  // MSE2 klappt beim Öffnen **alles** auf. Der erste Anlauf hier machte es
  // umgekehrt und begründete es mit der Länge — über die 324 Pakete des
  // installierten `M15-Magic-Pack` wären es rund 480 Zeilen. Das ist wahr und
  // war trotzdem der falsche Schluss: eine gruppierte, eingerückte,
  // geordnete Liste ist **nicht** dieselbe wie die flache von vorher, und wer
  // sie zuklappt, verlangt von jedem Suchenden erst einen Klick.
  //
  // Was bleibt, ist das Zuklappen als **Werkzeug**: wer `magic` mit seinen 208
  // Paketen wegräumen will, kann es.
  const [eingeklappt, setzeEingeklappt] = useState<ReadonlySet<string>>(
    new Set(),
  )

  const pakete = useMemo(() => daten?.pakete ?? [], [daten])

  // **Die Aufschrift des Sammelknotens kommt von hier**, weil hier die
  // Sprachkataloge erreichbar sind (§9). Was hineingehört, entscheidet
  // [paketbaum] — die Regel steht bei MSE (`package_update_list.cpp`,
  // Zeilen 204–205, Fassung f9d9356 vom 2021-07-06).
  const zeilen = useMemo(
    () => sicht(pakete, wahl, texte.seiten.pakete.katalogOhneGruppe, sprache),
    [pakete, wahl, texte, sprache],
  )

  // **Eine Suche hebt jedes Zuklappen auf.** Wer sucht, will den Treffer
  // sehen und nicht die Gruppe, in der er stecken könnte — und die gefilterte
  // Menge ist klein genug, dass es nichts kostet.
  const suchend = wahl.suche.trim() !== ''

  // **Marken an Paketen** — UX-12 (#1684): ein Filter nimmt Pakete weg,
  // die keine der gewählten tragen; Gruppenzeilen bleiben stehen.
  const marken = useMarkenfilter('paket')
  const { getroffen } = marken
  const gezeigt = useMemo(
    () =>
      nachMarken(
        suchend ? zeilen : sichtbareZeilen(zeilen, (k) => !eingeklappt.has(k)),
        (zeile) => zeile.paket && `${zeile.paket.art}/${zeile.paket.name}`,
        getroffen,
      ),
    [zeilen, eingeklappt, suchend, getroffen],
  )

  // **Die Vorbehalte je Paket** (#57) — als Zahl an der Zeile statt nur im
  // Abschnitt darunter. #244 nennt es ausdrücklich: der Zustand einer Zeile
  // gehört an die Zeile.
  //
  // **Gelesen hat die Seite** (U14, #2159): der Abschnitt darunter zeigt
  // dieselben Befunde, und zwei Rufe für dieselbe Frage gaben zwei Stände.
  // Hier stand ein eigener Ruf, damit die Liste nicht von der Reihenfolge
  // zweier Geschwister abhing — ihr gemeinsamer Elternteil löst beides.

  // **Der Schlüssel ist `name.art`**, weil
  // `service.Paketverwalter.Abhaengigkeiten` ihn so bildet — `magic-m15.mse-style`.
  const vorbehalte = useMemo(() => {
    const gezaehlt = new Map<string, number>()

    for (const befund of befunde ?? []) {
      gezaehlt.set(befund.paket, (gezaehlt.get(befund.paket) ?? 0) + 1)
    }

    return gezaehlt
  }, [befunde])

  // **Ein Wechsel des Pakets räumt den offenen Plan** (§7, Regel 1): ein
  // Plan für das eine unter dem Namen des anderen wäre die schlimmste
  // Verwechslung, die diese Seite machen kann.
  const waehlen = (paket: PaketDTO) => {
    const kennung = `${paket.art}/${paket.name}`

    if (kennung !== gewaehlt) {
      planung.verwerfen()
    }

    setzeGewaehlt(kennung)
    zeigen()
  }

  const planen = (art: 'entfernen' | 'zurueck', paket: PaketDTO) => {
    setzeGewaehlt(`${paket.art}/${paket.name}`)
    void planung.planen(art, paket)
  }

  const imInspector = pakete.find(
    (paket) => `${paket.art}/${paket.name}` === gewaehlt,
  )

  const umschalten = (schluessel: string) => {
    setzeEingeklappt((zu) => {
      const naechste = new Set(zu)

      if (naechste.has(schluessel)) {
        naechste.delete(schluessel)
      } else {
        naechste.add(schluessel)
      }

      return naechste
    })
  }

  // **Gezählt wird das Gefundene, nicht das Sichtbare.** Zugeklappt ist eine
  // Anzeigefrage; „0 von 4" zu melden, weil gerade nichts aufgeklappt ist,
  // wäre eine Antwort auf eine Frage, die niemand gestellt hat — und der
  // leere Zustand darunter träte ein, obwohl vier Pakete dastehen. Beim ersten
  // Anlauf tat er genau das.
  const gefunden = zeilen.filter((zeile) => zeile.paket !== undefined).length

  return (
    <section className="flex flex-col gap-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-lg font-semibold">{eigene.katalogUeberschrift}</h2>

        {/* **Einbinden beginnt mit einem Plan** (DOM-01.AK-02): der Dialog
            wählt ein Verzeichnis, der Inspector sagt, ob es ein neues Paket
            ist oder ein vorhandenes ersetzt — und erst dann wird gebunden. */}
        <Knopf
          angehalten={planung.laeuft}
          aufKlick={() => {
            setzeGewaehlt(undefined)
            void planung.waehlen()
          }}
        >
          {eigene.plan.einbinden}
        </Knopf>
      </div>

      {/* **Nur, solange nichts dasteht** (#2069): ein Nachlesen lässt die
          Tabelle stehen, und ein Satz darüber schöbe sie trotzdem. */}
      {laedt && daten === undefined && <p>{texte.bausteine.zustaende.laedt}</p>}

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      {/* **Mit der Handlung daneben** (G02, #2249): wer auf eine leere
          Liste schaut, schaut nicht zum Knopf oben. */}
      {daten?.pakete.length === 0 && (
        <Katalogleer
          laeuft={planung.laeuft}
          einbinden={() => {
            setzeGewaehlt(undefined)
            void planung.waehlen()
          }}
        />
      )}

      <Arbeitsplatz
        bezeichnung={eigene.plan.arbeitsplatz}
        gedaechtnis={gedaechtnis}
        beiEnge="inspector"
        inspectorAnlass={anlass}
        navigator={{
          bezeichnung: eigene.plan.navigator,
          inhalt: (
            <div className="flex flex-col gap-4">
              <Paketfilter
                wahl={wahl}
                arten={arten(pakete, sprache)}
                aendern={setzeWahl}
              />

              {/* **Die Zahl steht als `status` da**, nicht als stiller Text:
                  wer tippt, hört „12 von 329", ohne die Liste durchzugehen. */}
              <p role="status" className="text-sm text-leise">
                {einsetzen(eigene.katalogGefunden, {
                  gefunden: String(gefunden),
                  gesamt: String(pakete.length),
                })}
              </p>

              <Markenfilterknopf art="paket" stand={marken} />
            </div>
          ),
        }}
        zentrum={
          gefunden === 0 ? (
            pakete.length > 0 && (
              <Katalogohnetreffer
                zuruecksetzen={() => {
                  // **Die Ordnung bleibt** — sie ist keine Auswahl.
                  setzeWahl({ ...leereWahl, ordnung: wahl.ordnung })
                }}
              />
            )
          ) : (
            /* **Die Liste rollt für sich — und das ist eine Anpassung, keine
                Portierung** (§15, Kennzeichnung `adapted`). MSE2 klappt alles
                auf (`packages_window.cpp`, Zeile 222, Fassung f9d9356 vom
                2021-07-06) und kann sich das leisten, weil seine Liste ein
                **eigenes Fenster** ist. Hier rollt das Zentrum des
                Arbeitsplatzes für sich, und die Abschnitte darunter bleiben
                erreichbar (§15, Regel 9: nach dem **Zweck** ersetzen). */
            <Pakettabelle
              zeilen={gezeigt}
              eingeklappt={eingeklappt}
              suchend={suchend}
              umschalten={umschalten}
              vorbehalte={vorbehalte}
              gewaehlt={gewaehlt}
              waehlen={waehlen}
              planen={planen}
              laeuft={planung.laeuft}
              marken={marken}
            />
          )
        }
        inspector={{
          bezeichnung: eigene.plan.inspector,
          inhalt: (
            <div className="flex flex-col gap-4">
              {/* **Ein offener Plan verdrängt die Angaben**: er ist die
                  Rückfrage, und neben ihr stünde eine zweite Stelle mit
                  denselben Knöpfen. */}
              <Paketplanansicht planung={planung} />

              {planung.plan === undefined && (
                <Paketinspector
                  paket={imInspector}
                  laeuft={planung.laeuft}
                  aufLoeschen={(paket) => {
                    planen('entfernen', paket)
                  }}
                  aufZurueck={(paket) => {
                    planen('zurueck', paket)
                  }}
                  weitergabe={weitergabe}
                />
              )}
            </div>
          ),
        }}
      />

      {daten !== undefined && daten.befunde.length > 0 && (
        <>
          <p className="max-w-prose text-leise">{eigene.katalogBefunde}</p>
          <ul className="space-y-1 font-mono text-sm">
            {daten.befunde.map((befund) => (
              <li key={befund}>{befund}</li>
            ))}
          </ul>
        </>
      )}

      {/* **Anlegen steht unter dem Arbeitsplatz** (#1350): es bringt ein
          Paket aus dem Nichts in den Katalog, und danach steht es oben.
          **Eingeklappt seit UXR-17** (#1925) — vier Felder, die man selten
          braucht, standen offen unter der Liste. */}
      <Aufklappkasten titel={eigene.anlegen} schreibt>
        <Paketanlage
          spiele={pakete.filter((eines) => eines.art === 'mse-game')}
          angelegt={nachladen}
        />
      </Aufklappkasten>
    </section>
  )
}
