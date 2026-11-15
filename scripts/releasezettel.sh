#!/usr/bin/env bash
# Der Zettel, der ins .dmg gehört — ausgelagert aus scripts/release.sh aus
# demselben Grund wie releaseversion.sh und die drei anderen: mit ihm ginge der
# Freigabeweg über 400 Zeilen (§2).
#
# **Und der Schnitt folgt der Frage** (§2): hier steht, *was auf dem Zettel
# steht*; in release.sh, *wie eine Freigabe entsteht*. Zwei Fragen, und die
# eine ändert sich, ohne dass die andere davon weiß.
#
# Nicht ausführbar — dieses Skript wird eingebunden.

# erststartnotiz schreibt den Zettel, der ins .dmg gehört.
#
# Er steht hier und nicht als Datei im Repository, damit er den Namen und die
# Fassung mitbekommt — und weil ein Zettel, den niemand erzeugt, irgendwann eine
# Anleitung für eine Fassung ist, die es nicht mehr gibt.
#
# **Der Wortlaut ist abgelesen, nicht erinnert** (#62): nachgestellt am
# 2026-08-16 auf macOS 26.6.1, Bildschirmfotos in docs/bilder/. Ein Wächter hält
# ihn gegen README.md und docs/RELEASE.md
# (internal/werkzeuge/erststartweg_test.go).
#
# **Und drei Wächter halten seinen Schluss** (#894). Bis zum 2026-08-23 hielt
# ihn keiner, und er trug zwei falsche Angaben in jedes ausgelieferte .dmg:
# „Lizenz: MIT" — seit dem 2026-08-16 gilt GPL-2.0-or-later — und die
# Behauptung, Inhalte Dritter würden nicht mitgeliefert, während dieses Skript
# dreißig Zeilen weiter unten mitliefern.sh ruft und 167 MB davon einlegt.
#
#   · internal/werkzeuge/lizenz_test.go hält die Lizenz gegen sechs andere Orte
#   · internal/werkzeuge/beigabenzettel_test.go hält die Beigabe gegen
#     mitliefern.sh — solange der Freigabeweg sie einlegt
#   · internal/werkzeuge/erststartweg_test.go hält den Weg wie bisher

erststartnotiz() {
  cat <<'NOTIZ'
MagicSetEditor4Mac — beim ersten Start
======================================

Diese Anwendung ist NICHT notarisiert. Es gibt keinen Apple-Developer-Account
und es soll keinen geben. macOS weist sie deshalb beim ersten Start ab.

Sie wird EINMALIG freigegeben:

  1. MagicSetEditor4Mac in den Ordner "Programme" ziehen und starten.
  2. Es erscheint ein Fenster mit dem Titel

       "MagicSetEditor4Mac" nicht geöffnet

     und dem Satz "Apple konnte nicht überprüfen, ob "MagicSetEditor4Mac"
     frei von Schadsoftware ist, die deinen Mac beschädigen oder deine
     Privatsphäre verletzen kann."

     ACHTUNG: Der obere, breitere Knopf heißt "In den Papierkorb legen".
     Der richtige ist der untere: "Fertig".
  3. Systemeinstellungen öffnen -> Datenschutz & Sicherheit.
  4. Nach unten zum Abschnitt "Sicherheit" scrollen. Dort steht

       "MagicSetEditor4Mac" wurde blockiert, um deinen Mac zu schützen.

  5. "Dennoch öffnen" anklicken und mit Kennwort oder Touch ID bestätigen.
  6. Die App erneut starten und die Rückfrage mit "Öffnen" bestätigen.

Danach startet sie ohne weitere Nachfrage.

Ab macOS 15 ist das der einzige Weg. Das frühere Öffnen über Rechtsklick ->
"Öffnen" gibt es nicht mehr; eine Anleitung, die es nennt, ist veraltet.

Lizenz
------

Der Quelltext dieses Programms steht unter GPL-2.0-or-later. Der Lizenztext
liegt als LICENSE bei.

Diese Lizenz gilt NICHT für Inhalte Dritter: Kartenrahmen, Mana-Symbole,
Schriften, Regeltexte, Namen und Marken von Wizards of the Coast oder anderen
Rechteinhabern. Die gehören ihren Rechteinhabern.

Und dieses Bündel bringt solche Inhalte MIT. Damit nach dem ersten Start
sofort eine Karte dasteht, liegen unter MagicSetEditor4Mac.app/Contents/
Resources/ Vorlagenpakete und ein Beispielset — Rahmen, Manasymbole,
Stilbögen und die Schriften der Vorlagen für M15-Karten. Für die Schriften
gelten je eigene Lizenzen; wer eine gleichnamige selbst installiert hat,
behält sie.

Kommandozeile, MCP-Server und Dienst
------------------------------------

Neben dem Fenster liegen drei Programme im Bündel, unter
MagicSetEditor4Mac.app/Contents/MacOS/:

  mse        die Kommandozeile
  mse-mcp    der MCP-Server für ein Sprachmodell
  mse-rest   der Dienst auf der Rückschleife

Sie liegen NICHT im PATH — diese Anwendung installiert nichts außerhalb ihres
eigenen Bündels. Wer sie von überall aufrufen will, legt den Verweis selbst:

  ln -s /Applications/MagicSetEditor4Mac.app/Contents/MacOS/mse /usr/local/bin/mse

Wem was gehört und unter welchen Bedingungen, steht in DRITTANBIETER.md.
NOTIZ
}
