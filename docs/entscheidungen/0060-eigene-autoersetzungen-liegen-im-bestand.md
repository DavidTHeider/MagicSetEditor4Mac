# 0060 — Eigene Autoersetzungen liegen im Bestand, je Spiel

**Datum:** 2026-08-31
**Anlass:** #1341,
Werkstück E von Phase 4.11
**Status:** entschieden

---

## Die Frage

Die `auto replace:`-Regeln eines Spielpakets greifen seit
#897 beim
Tippen. **Eigene** Regeln gibt es nicht, und wer sie einführt, muss zuerst
sagen, **wo sie liegen**.

Das Issue nannte zwei Möglichkeiten:

- **neben dem Paket**, wie die Marken
  ([0053](0053-marken-liegen-neben-dem-set.md)) — dann überleben sie einen
  Paketwechsel und gelten über Spiele hinweg;
- **im Paket**, als Paketdatei — dann reisen sie mit dem Paket und lassen sich
  weitergeben.

## Was MSE2 tut — nachgelesen, nicht erschlossen

**Keines von beiden.** MSE hält sie in den **Einstellungen**, und zwar je
Spiel:

```cpp
class GameSettings {
  …
  bool                 use_auto_replace;
  vector<AutoReplaceP> auto_replaces;  ///< Things to autoreplace in textboxes
```

`src/data/settings.hpp`, Zeilen 83–84, Fassung `f9d9356` vom 2021-07-06. Das
Fenster, das sie bearbeitet, schreibt genau dorthin zurück
(`AutoReplaceWindow::store`, `src/gui/auto_replace_window.cpp`, Zeilen
292–295).

**Und die Regeln des Pakets wandern hinein**, statt daneben zu stehen:

```cpp
FOR_EACH_CONST(ar, game.auto_replaces) {
  bool already_have = false;
  FOR_EACH(ar2, auto_replaces) {
    if (ar->match == ar2->match) { ar2->custom = false; already_have = true; break; }
  }
  if (!already_have) { ar->custom = false; auto_replaces.push_back(ar); }
}
```

`GameSettings::initDefaults`, `src/data/settings.cpp`, Zeilen 88–104. Drei
Dinge stehen darin:

1. **Verglichen wird über `match`.** Zwei Regeln mit demselben Muster sind
   dieselbe Regel.
2. **Eine Regel des Pakets ist nicht `custom`.** Ihr Muster lässt sich nicht
   ändern und sie lässt sich nicht löschen — nur ihr Ersatz, ihr `enabled` und
   ihr `whole word` (`AutoReplaceWindow::refreshItem`, Zeilen 253–264).
3. **Was der Nutzende angelegt hat, bleibt `custom`** und überlebt jede
   Paketaktualisierung.

## Die Entscheidung

**Im Bestand, je Spiel** — MSEs Ort, in der Ablage dieses Projekts.

Das ist dieselbe Abweichung, die `docs/PORTIERUNGSKARTE.md` für die
Einstellungen schon führt: *`adapted` — im Bestand statt in einer Datei.* Und
dieselbe wie bei den Spalten der Kartenliste
(`0021_listenspalten.sql`, seit [0189](0189-nur-die-grundlinie.md) in der Grundlinie), die aus demselben
`GameSettings` kommen und aus demselben Grund je Spiel liegen: **welche Regeln
es gibt, sagt das Spiel.**

## Was gegen die beiden Vorschläge des Issues spricht

**Neben dem Paket, spielübergreifend.** Es wäre die stärkere Zusage — und die
falsche. Eine Regel wie `(W)` → `<sym>W</sym>` ergibt nur dort einen Sinn, wo
`<sym>` etwas bedeutet; in einem anderen Spiel schreibt sie Unsinn in den Text.
MSE hat sie nicht ohne Grund je Spiel abgelegt.

**Im Paket.** Es hieße, in ein **fremdes** Paket zu schreiben. §11 erlaubt das
Lesen; ein Programm, das eine bezogene Vorlage verändert, macht aus einem Stand
mit Prüfsumme einen ohne. Die Paketwerkstatt tut das ausdrücklich und auf
Verlangen — die Autoersetzung ist aber keine Änderung an der Vorlage, sondern
eine Gewohnheit der schreibenden Person.

## Was daraus folgt

- Eine Tabelle `spiel_autoersetzungen` je Spiel, mit `eigen` als Gegenstück zu
  MSEs `custom`.
- Der **globale Schalter** je Spiel kommt mit
  (`use_auto_replace`) — `internal/msepaket/autoersetzungen.go` nennt sein
  Fehlen seit dem 2026-08-23 als benannte Lücke.
- Die Regeln des Pakets werden beim Lesen **hineingemischt**, nicht kopiert:
  eine Paketaktualisierung bringt neue Regeln mit, und eine entfernte
  verschwindet. Was der Bestand dazu führt, ist die **Abweichung** — Ersatz,
  Schalter, Wortgrenze —, und die bleibt.

## Was verworfen wurde

**Eine Kopie der Paketregeln in den Bestand schreiben**, wie MSE es tut
(`auto_replaces.push_back(ar)`). MSE hat dafür einen Grund, den es selbst als
`TODO: when we start saving games, clone here` notiert: seine Liste hält
Zeiger auf die Regeln des Spiels. Eine Kopie im Bestand hieße hier, dass eine
Paketaktualisierung **nichts** ändert — und dass ein Muster, das der Urheber
berichtigt hat, in jedem Bestand falsch bleibt.
