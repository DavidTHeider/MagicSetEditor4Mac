// **Alle Karten anhaken** —
// #1242.
//
//	void CardsPanel::doSelectAll() { … }   // leitet je nach Fokus weiter
//	void ItemList::doSelectAll()   { … }   // wählt dort ALLE Karten
//
// `src/gui/set/cards_panel.cpp`, Zeilen 404–410; `src/gui/control/item_list.cpp`,
// Zeilen 43–64; Fassung `f9d9356` vom 2021-07-06.
//
// # Der Befund
//
// In `menue.go`, `HANDBUCH.md` und `PARITAET.md` stand, *Alles auswählen* sei
// bei MSE2 nur Sache der Textmarke. **Das stimmte nicht** — angestrichen vom
// Nachlauf über Phase 3.7.
//
// # Das fachliche Gegenstück
//
// Die **Anhakhaken** der Kartenliste. An ihnen hängen Massenänderung und
// Zwischenablage; „alle anhaken" ist damit nicht nur portierbar, sondern
// nützlich.
//
// # Eine eigene Datei
//
// **Weil `Karten.tsx` mit ihr über 400 Zeilen ginge** (§2). Der Schnitt folgt
// der Frage: dort steht, **was** die Seite zeigt; hier, **was ein Menüeintrag
// mit ihrer Auswahl tut**.
import { useCallback } from 'react'

import { ereignisAllesAuswaehlen, useMenueereignis } from '../fenster'

// Anhakbar ist alles, was eine Kennung hat.
interface Anhakbar {
  id: string
}

// **Und warum die Anhakung beim Setwechsel geräumt wird** (§7, Regel 1).
//
// Sie stand in `Karten.tsx` schon einmal und wurde wieder entfernt, weil die
// Gegenprobe grün blieb — mit der Begründung, eine Kartenkennung gebe es genau
// einmal, und die gefilterte Liste erledige den Rest. Das übersieht den
// **Rückwechsel auf dasselbe Set**: dann sind die alten Haken wieder da, und
// eine Massenänderung geht gegen eine Auswahl, die niemand mehr getroffen hat.
//
// Der Test wechselte nur hin, nie zurück. Der Zweitblick über Phase 5 hat
// beides gefunden — die Lücke und den Test, der sie nicht sah.

// useAllesAnhaken hakt auf den Menüeintrag hin jede Karte der Liste an.
//
// **Die gefilterte Liste und nicht der ganze Bestand.** Wer gesucht hat, sieht
// eine Auswahl — und „alle" meint, was dasteht. Genau so wählt MSE: `ItemList`
// führt die **sortierte, gefilterte** Liste.
export function useAllesAnhaken(
  liste: readonly Anhakbar[],
  setzeAngehakt: (kennungen: readonly string[]) => void,
): void {
  useMenueereignis(
    ereignisAllesAuswaehlen,
    useCallback(() => {
      setzeAngehakt(liste.map((karte) => karte.id))
    }, [liste, setzeAngehakt]),
  )
}
