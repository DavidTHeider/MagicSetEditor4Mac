// Die Kartenvorschau auf der Stilseite
// (#643).
//
// # Warum sie hier steht
//
// **§6: eine Änderung ist in der Vorschau zu sehen, ohne dass jemand etwas
// bestätigt.** Für ein Kartenfeld hielt das seit je — der Editor steht neben
// dem Bild. Eine Stiloption lag auf einer anderen Seite als die Vorschau; wer
// die Rahmenfarbe umstellte, sah davon nichts.
//
// **MSE hat dafür keine zwei Seiten.** Sein Stilbereich hält einen eigenen
// `CardViewer` und zeichnet die gewählte Karte, sobald sich etwas ändert
// (`src/gui/set/style_panel.cpp`, Fassung f9d9356 vom 2021-07-06). Dass daraus
// hier eine Seite geworden ist, folgt §6; dass die Vorschau dabei fehlte, war
// ein Versehen.
//
// # Warum eine eigene Stelle und nicht [Kartenvorschau]
//
// **Die trägt mehr, als hier gebraucht wird**: anklickbare Feldflächen, das
// Zuschnittfenster, die Befunde des Rasterers. Sie hier zu benutzen hieße,
// ihre Eigenschaften zu Pflichtangaben zu machen — oder ihr die Fähigkeiten
// wegzunehmen, die der Editor braucht (§2).
import { useCallback } from 'react'

import { api, auspacken, type KarteDTO, type VorschauDTO } from '../api'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { Ladefeld } from './Ladefeld'

interface StilvorschauEigenschaften {
  setID: string

  // stand steigt bei jeder Änderung am Stil. **Ein Zähler und kein Rückruf**:
  // die Vorschau lädt selbst, und wer sie von außen nachlüde, hätte zwei
  // Ladewege auf dieselben Daten (§7, Regel 2).
  stand: number

  // probe ist eine Stiloption, die jemand gerade tippt — noch nicht
  // geschrieben (#590).
  //
  // **Solange sie dasteht, zeigt die Vorschau diesen Wert**; danach wieder den
  // des Bestands. Dieselbe Bauart wie `getippt` im Karteneditor seit #288.
  probe: { name: string; wert: string } | undefined
}

export function Stilvorschau({
  setID,
  stand,
  probe,
}: StilvorschauEigenschaften) {
  const texte = useTexte()

  const holenKarten = useCallback(async () => {
    const antwort = await api().Karten(setID)

    return auspacken(antwort, () => antwort.karten)
  }, [setID])

  const karten = useLaden<KarteDTO[]>(holenKarten, [setID])

  // **Die erste Karte des Sets**, und das ist eine Entscheidung: die Stilseite
  // kennt keine aufgeschlagene. Eine willkürliche wäre schlechter als eine
  // feste — beim nächsten Aufschlagen stünde eine andere da.
  const erste = karten.daten?.[0]

  const holenVorschau = useCallback(async () => {
    if (erste === undefined) {
      return undefined
    }

    // **Die getippte Option schlägt den Bestand** — und schreibt nichts. Ein
    // Schreiben je Tastendruck legte einen Schritt im Verlauf an, und ⌘Z nähme
    // Ziffern zurück statt Änderungen (`App.Stilprobevorschau`).
    const antwort =
      probe === undefined
        ? await api().Vorschau(erste.id, 1)
        : await api().Stilprobevorschau(erste.id, probe.name, probe.wert, 1)

    return auspacken(antwort, () => antwort.vorschau ?? undefined)
  }, [erste, probe])

  // **Kein Warten, sondern Abbrechen** — die Sonderregel dieses Projekts (§7).
  // Jede Änderung stößt einen Lauf an; der vorige wird im Backend über seinen
  // `context` abgebrochen, und `useLaden` verwirft eine überholte Antwort.
  const vorschau = useLaden<VorschauDTO | undefined>(
    holenVorschau,
    [erste, stand, probe],

    // **Das alte Bild bleibt stehen** — die Ausnahme von §7 Regel 1, und sie
    // ist ausdrücklich gewollt: die Abhängigkeiten ändern sich mit jedem
    // Tastendruck, und ein Bild, das dabei verschwindet und wiederkommt, ist
    // schlechter als ein einen Augenblick altes.
    //
    // Seit dem 2026-08-22 räumt `useLaden` von sich aus (Zweitblick über
    // Phase L); wer behalten will, sagt es.
    true,
  )

  return (
    <section aria-label={texte.seiten.stil.vorschauTitel} className="mt-4">
      <h3 className="font-semibold">{texte.seiten.stil.vorschauTitel}</h3>

      <Ladefeld
        laedt={karten.laedt}
        fehler={karten.fehler ?? vorschau.fehler}
        leer={erste === undefined}
      >
        <p className="mt-2 text-sm text-leise">
          {texte.seiten.stil.vorschauLeer}
        </p>
      </Ladefeld>

      {/* **Das alte Bild bleibt stehen, während das neue rechnet.** Wer es
          wegnähme, ließe die Karte bei jeder Änderung verschwinden und
          wiederkommen — dieselbe Regel wie an der Probevorschau des Editors. */}
      {vorschau.daten !== undefined && (
        <img
          src={vorschau.daten.bild}
          alt={texte.seiten.stil.vorschauTitel}
          width={vorschau.daten.breite}
          height={vorschau.daten.hoehe}
          className="mt-2 block max-w-full rounded-sm border border-kante"
        />
      )}
    </section>
  )
}
