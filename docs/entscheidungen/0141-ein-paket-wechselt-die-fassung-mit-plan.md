# 0141 — Ein Paket wechselt die Fassung mit Plan, voriger Fassung und ohne Kaskade

**Stand:** entschieden am 2026-09-27 · DOM-01.AK-01 bis AK-04
(#1696)

## Die Frage

DOM-01 verlangt, dass Installation, Aktualisierung, Deinstallation und die
Rückkehr zur vorigen kompatiblen Fassung **vorher** zeigen, welche Sets und
Stile betroffen sind und welche Abhängigkeiten brechen (AK-02), dass eigene
Kopien und geänderte Pakete nicht still überschrieben werden (AK-03) — und
dass Herkunft, Version, Prüfsumme, Lizenz, Abhängigkeiten und lokale
Änderungen am Objekt sichtbar und filterbar sind (AK-01).

Bis hierher band der Verwalter ein und entfernte, sonst nichts: ein
vorhandenes Paket war ein Konflikt, eine neue Fassung ließ sich nur über
Entfernen und Einbinden legen — und dabei ging die alte verloren.

## Was MSE2 tut

Nachgelesen in `src/data/installer.cpp`, `installer.hpp` und
`src/util/io/package_manager.cpp`, Fassung f9d9356 vom 2021-07-06:

- **`PackageStatus`** kennt „installiert", „Update verfügbar", „lokal
  geändert" (`PackageVersion::STATUS_MODIFIED` in
  `src/util/io/package_manager.hpp`). Geändert heißt dort: die
  **Änderungszeit einer Datei** weicht von der Liste ab, die beim Segnen
  (`bless`) festgehalten wurde, oder eine Datei kam hinzu oder fehlt
  (`PackageVersion::check_status`, `src/util/io/package_manager.cpp`,
  Zeilen 446–484). **Gezeigt wird es nicht**: `determineStatus` liest es nur
  hinter `USE_MODIFIED_CHECK`, und der steht auf 0 — Pakete ließen sich
  nicht segnen (`src/data/installer.cpp`, Zeilen 25 und 306).
- **`set_package_action`** zieht eine Handlung durch die Abhängigkeiten nach:
  wer ein Paket entfernt, entfernt die, die es brauchen, mit
  (`remove_package_dependency`); wer eines installiert, installiert seinen
  Bedarf mit (`add_package_dependency`).
- Ein Rückweg zur vorigen Fassung **fehlt**: `PackageDirectory::install`
  legt die neue als `<name>.new` an, entfernt die alte, benennt um und
  segnet (`src/util/io/package_manager.cpp`, Zeilen 358–374).

## Die Entscheidung

1. **Der Plan nennt, er nimmt nichts mit** — Portierungsmarke **adapted**.
   MSE2s Kaskade zieht still Pakete mit, die niemand genannt hat. Hier
   nennt der Plan die Abhängigen als **Konflikte** und die Sets als
   **Betroffene**; ausgeführt wird genau der geplante Vorgang. Wer den Stil
   mitnehmen will, plant ihn als eigenen Vorgang. Der Grund ist AK-02 selbst:
   „erkennen" heißt benennen, und eine Kaskade handelt, bevor jemand
   erkannt hat.
2. **Eine vorige Fassung, eine Stufe tief** — Portierungsmarke **new**
   (gesucht in `installer.cpp`, `package_manager.cpp`,
   `packages_window.cpp`; MSE2 hat keinen Rückweg). Ein Ersetzen hebt die
   bisherige Fassung **neben** dem verwalteten Verzeichnis auf
   (`<verwaltet>-vorige/<name>.<art>`), samt einem Vermerk mit Herkunft,
   Prüfsumme und Abdruck. **Zurück tauscht beide** und ist damit selbst
   umkehrbar. Eine Kette von Fassungen wäre ein Archiv, und das verlangt
   AK-02 nicht („zur vorigen kompatiblen Version").

   **Ein gescheiterter Wechsel hinterlässt den ganzen Stand von vorher** —
   die aktuelle Fassung, die vorige samt Vermerk **und** den Katalog, der
   erst abgeglichen wird, wenn das gewechselte Paket lesbar ist (Zweitblick
   über Phase 6.5, Befunde Z1 und Z2; Nachlauf N1 und N2). Scheitert der
   Rückweg selbst, sagt der Fehler es und nennt den Ort (N4). **Ein Rest
   eines abgebrochenen Wechsels** — `.beiseite`, `.tausch` neben der
   vorigen Fassung — wird zurückgelegt statt geräumt, denn er ist die
   einzige erreichbare vorige Fassung; liegt am Ort schon eine, hält der
   nächste Wechsel an und nennt beide (N3).
3. **Lokal geändert heißt: der Abdruck weicht ab** — Portierungsmarke
   **adapted** von `check_status`, mit demselben Maß. Der Abdruck geht über
   Pfad, Größe und Änderungszeit jeder Datei und entsteht beim Einbinden —
   das Einbinden **segnet**, und damit fällt der Grund weg, aus dem MSE2 die
   Prüfung abgeschaltet hat. **Eine berührte
   Datei zählt als geändert**, auch wenn ihr Inhalt gleich blieb: eine
   Prüfsumme über den Inhalt kostete bei 526 Paketen jedes Öffnen der
   Liste, und die falsche Richtung dieses Fehlers ist die harmlose — es
   wird einmal mehr gefragt.
4. **Ein geändertes Paket wird nur mit Bestätigung ersetzt, und auch dann
   geht die Änderung nicht verloren** (AK-03): sie liegt danach als vorige
   Fassung da. Ein automatisches Update gibt es nicht; kein Weg ersetzt
   ungefragt.
5. **Die Lizenz steht als „nicht angegeben"** da. Ein MSE-Paket hat kein
   Lizenzfeld (`src/data/installer.hpp`, dieselbe Fassung); was dort nicht
   steht, wird hier nicht erfunden (§10, „auch eine Ablehnung braucht
   Deckung").
6. **Nutzung nur bei Spielen und Stilen.** Gezählt wird über die Setköpfe
   dieses Profils; welches Set eine Symbolschrift braucht, steht in keinem
   Kopf, und eine Null hieße dort „keines", wo es „unbekannt" ist.

7. **Was einem Set fehlt, steht am Set** (AK-04) — Portierungsmarke
   **adapted**. MSE2 öffnet ein Set mit fehlendem Paket gar nicht
   (`PackageNotFoundError`, `src/util/io/package.cpp` Zeile 80); hier öffnet
   es mit dem Behelfsplan (§8), und der **Bedarf** nennt den Grund: vom
   Setkopf aus — Spiel und Stil — und Stufe für Stufe durch `depends on:`,
   jedes Paket einmal. Behoben wird **aus dem Set heraus**: *Paket
   einbinden …* zeigt dort denselben Plan wie auf der Paketseite, *Bei den
   Paketen suchen* springt zur Paketseite mit gefüllter Suche. **Für einen
   fehlenden Stil wird ein Name genannt**, und zwar der, unter dem jemand
   sucht: der Wert, wie er dasteht, wenn er schon mit dem Spiel beginnt,
   sonst die Kurzform mit dem Spiel davor. Gesucht wird weiter unter beiden.
8. **Keine stillen Rechte** (AK-04). Ein Paket bringt Vorlagen und
   MSE-Script mit; das Skript hat keinen Zugriff auf Dateien, Netz oder
   Programme (`internal/msescript` importiert nichts davon außerhalb der
   Tests), ein Bezug packt keine ausführbaren Dateien aus, und **was ein
   Paket braucht, wird nicht mitgeholt** — Entscheidung 1. Der Plan nennt
   vor dem Einbinden die **Quelle** und diesen Satz.

## Die Flächen

Alle vier, **erweiternd** nach §16:

| Fähigkeit | Fenster | MCP | CLI | REST |
| --- | --- | --- | --- | --- |
| Plan | `PaketPlan`, `PaketPlanWaehlen` | `lesen.paket.plan` | `mse paket-plan` | `GET /pakete/plan` |
| Ersetzen | `PaketErsetzen` | `schreiben.paket.ersetzen` | `mse paket-ersetzen` | `PUT /pakete/ersetzt` |
| Zurück | `PaketZurueck` | `schreiben.paket.zurueck` | `mse paket-zurueck` | `POST /pakete/{art}/{name}/zurueck` |
| Bedarf eines Sets | `Paketbedarf` | `lesen.set.paketbedarf` | `mse paketbedarf` | `GET /sets/{setID}/paketbedarf` |

**Am Fenster ist das Einbinden jetzt zweischrittig**: `PaketPlanWaehlen`
öffnet den Verzeichnisdialog und gibt einen Plan; eingebunden wird mit
`PaketEinbindenVon` und der Quelle aus dem Plan. `PaketInstallieren` ist als
Bindung weggefallen — die Wails-Naht ist nach §16 ausgenommen.

## Verworfen

- **Die Kaskade aus MSE2 übernehmen.** Sie ist das Gegenteil von
  „erkennen": sie löst einen Konflikt, den niemand gesehen hat.
- **Eine Prüfsumme über den Inhalt als Maß für „geändert".** Teuer bei
  jedem Blick auf die Liste, und der Gewinn ist eine Frage weniger.
- **Mehrere vorige Fassungen.** Ein Archiv ohne Anlass; wer eine ältere
  braucht, bezieht sie an ihrem Stand.

## Was offen bleibt

- **Nach einem Zurück behält der Katalog eine Herkunft**, wenn die
  zurückgeholte Fassung keine hatte: der Abgleich lässt eine leere Angabe
  die vorhandene nicht überschreiben. Die Herkunftsart kann dadurch
  „übernommen" statt „eigen" lauten.
