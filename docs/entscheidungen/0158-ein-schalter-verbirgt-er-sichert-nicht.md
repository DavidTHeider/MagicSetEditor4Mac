# 0158 — Ein Schalter verbirgt, er sichert nicht

**Stand:** angenommen am 2026-10-01 · OPS-10
(#1908).
Zuordnung **bereichs- und fähigkeitsweise**, so entschieden am selben Tag.

## Die Frage

OPS-10 verlangt Schalter für Bereiche und Funktionen, vorgabemäßig alle an.
An allen vier Flächen soll **eine** Abweisung (`bereichAbgeschaltet`) gelten.
Zu klären war: Wo greift ein Schalter? Woher weiß jeder Weg, an welchem er
hängt? Und was darf ein Schalter ausdrücklich nicht werden?

## Die Entscheidung

1. **Zehn Schalter in `internal/domain/schalter.go`.** Vier sind
   Arbeitsbereiche: Sammeln, Spielen, Handeln, Ressourcen. Sechs sind
   Funktionen: KI, Netz und Gegenüber, Marktplätze und Preise, Turniere,
   Dienst für Skripte, Werkzeugserver über HTTP. Jeder Schalter hat eine
   Kennung, eine Art, eine Vorgabe und eine Reife (*stabil*, *Vorschau*,
   *auslaufend*). **Gestalten und das Programm sind Kern** und haben keinen
   Schalter. Sonst gäbe es kein Fenster, in dem man ihn wieder einschaltet.
2. **Die Zuordnung steht in `docs/SCHALTER.md`, je Fähigkeit.** Jede Zeile
   nennt den Bereich (`kern` oder eine Kennung) und höchstens eine Funktion.
   Ein Abschnitt allein trägt nicht: Freigaben an ein Gegenüber hängen am
   Handeln **und** am Netz, und der Proxy ist Kern, obwohl er im Netzabschnitt
   steht.
3. **Die Wege kommen aus `FAEHIGKEITEN.md`.** Die Tafel
   `internal/domain/schalterwege.json` ist erzeugt, nicht von Hand
   geschrieben. `internal/werkzeuge/schalterwege_test.go` schreibt sie mit
   `SCHALTERWEGE_SCHREIBEN=1` und wird rot, sobald eine Fähigkeit ohne Zeile
   dasteht oder die Tafel nicht mehr passt.
4. **Wann ein Weg abweist.**
   - Eine Fähigkeit ist aus, sobald **einer** ihrer Schalter aus ist.
   - Ein Weg weist erst ab, wenn **jede** Fähigkeit, die ihn nennt, aus ist.
   - Ein Weg, den auch der Kern nimmt, steht gar nicht in der Tafel.
     Sonst schaltete „Spielen aus“ einen Weg ab, den Gestalten braucht.
5. **Ein Tor je Fläche, und überall dieselbe Antwort.** Sie lautet
   `domain.Abschaltung` → Code `abgeschaltet`, Schlüssel
   `bereichAbgeschaltet`, Wert `schalter`. REST antwortet mit 403.
   - **Fenster:** in `Bindungen.app()`, neben dem Tor der Sitzung. Der Weg
     ist die äußerste exportierte Methode im Aufrufrahmen. Wails gibt eine
     Panik nur als Satz weiter; deshalb trägt der Satz den Fehler als JSON
     hinter `mse4mac-fehler:`, und `anbindung.ts` macht daraus einen
     `Bindungsfehler`.
   - **MCP:** vor dem Werkzeug, hinter der Wache des Zugriffstokens.
   - **REST:** vor dem Endpunkt, hinter der Marke. Die Platzhalternamen sind
     weggelassen, weil Katalog und `net/http` sie verschieden nennen dürfen.
   - **Kommandozeile:** vor der Verteilung. Ein Platzhalter wie `<id>` steht
     dort für jedes Wort.
6. **Gespeichert als Einstellung `schalter.<kennung>`**, Werte `an` und
   `aus`, Träger **Rechner**, sofort wirksam. Ein Wert, den es nicht gibt,
   gilt mit der Vorgabe.
7. **Dienst und Werkzeugserver enden mit ihrem eigenen Schalter**
   (OPS-10.AK-03).
   - Schaltet das Fenster einen ab, beendet es den Dienst, den es gestartet
     hat.
   - Ein eigenständiger `mse-rest` und ein `mse-mcp` über HTTP fragen ihn bei
     jeder Anfrage. Ist er aus, geben sie die Abweisung zurück und beenden
     ihren Prozess.
   - Eingeschaltet startet keiner von selbst.

   **Und mit ihnen alles andere, was im Fenster läuft** (OPS-10b,
   #1978,
   nachgetragen am 2026-10-01). Der erste Anlauf beendete nur die zwei
   Dienste; Horchposten, Fernhorchen, Ankündigung, Partiefreigabe,
   Partiebeitritt, Indexaufbau und Bildaufträge liefen weiter — und ihr
   Beenden hing am selben Schalter, war also abgewiesen. Ein offener Port,
   den das Fenster nicht mehr schließen konnte.
   - **Die Regel: ein Lauf endet, sobald sein Beenden gesperrt ist.** Die
     Domäne führt je Lauf seine Beenden-Bindung
     (`internal/domain/schalterlaeufe.go`) und fragt dieselbe Tafel wie das
     Tor — keine zweite Zuordnung, die beim nächsten Schalter
     auseinanderliefe. Die zwei Dienste sind zwei Läufe dieser Liste.
   - **Zwei Anlässe.** Legt das Fenster den Schalter um, beendet es nach
     dem Schreiben. Legt ihn `mse`, MCP oder REST um, erfährt es das
     Fenster beim nächsten Weg, der den Stand liest: das Tor beendet
     nebenbei, was er sperrt.
   - **Ein Lesefehler hält keinen Port offen**: der eben geschriebene
     Schalter gilt dann allein.
   - **Eine Partie ist kein Lauf.** `PartieBeenden` hängt an `spielen` und
     ist ein Datensatz: wer wieder einschaltet, findet sie, wie er sie
     verließ. Ein Wächter verlangt für jedes gesperrte *Beenden* und
     *Abbrechen* der Tafel einen Lauf oder einen Grund.
8. **Das Fenster verbirgt Seiten, nicht Knöpfe.** Ein Schalter, der wie ein
   Bereich heißt, verbirgt dessen Seiten; einer, der wie eine Seite heißt
   (`turniere`), verbirgt sie.
   - Sie verschwinden aus Seitenleiste und Palette.
   - Wer per Kürzel oder Verlauf trotzdem hinkommt, sieht einen leeren
     Zustand mit dem Weg zu den Schaltern.
   - **Die Dienstzone bleibt immer.**
   - Ein Knopf auf einer Seite des Kerns, etwa der KI-Vorschlag an der Karte,
     bleibt stehen und bekommt beim Drücken die Abweisung.
   - **Die Menüleiste führt keine Bereichseinträge** (nachgelesen in
     `internal/app/menue.go`); dort ist nichts zu verbergen.
9. **Die Reife ist sichtbar**: am Schalter auf der Einstellungsseite und als
   Marke am Bereich in der Leiste, solange sie nicht *stabil* ist. Eine
   Funktion mit Vorgabe „aus“ auszuliefern heißt, im Verzeichnis
   `Vorgabe: false` zu setzen. Kein Test und kein Wächter muss dafür
   umgebaut werden; `TestEineVorschauKommtAbgeschaltet` belegt es.

## Was ein Schalter nicht ist

**Keine Autorisierung.** Ein abgeschalteter Bereich ist verborgen, nicht
gesichert.

- Die Rollen aus `AUTORISIERUNG.md` gelten unverändert.
- Ein Bestand, der nicht antwortet, lässt jeden Weg durch. Der Weg stößt dann
  selbst auf den Fehler und nennt ihn.
- **Ein Sicherheitsriegel wird nie ein Schalter.** Ein Mensch, der einen
  Riegel abschalten könnte, hätte einen Knopf, der ihn öffnet. Wer etwas
  schützen will, nimmt Rollen, Freigaben und Richtlinien (`richtlinie.*`),
  denn die prüft jede Fläche.
- **Daten bleiben unberührt.** Abschalten löscht nichts, und Einschalten
  stellt alles wieder her.

## Verworfen

- **Je Abschnitt von `FAEHIGKEITEN.md` ein Schalter.** Das trägt weit, aber
  nicht ganz (siehe Punkt 2), und die Ausnahmen wären eine zweite Liste
  gewesen.
- **Eine Liste der Schalter im Fenster.** Das Fenster liest das Verzeichnis
  über die Einstellungen und die Reife über `Schalter`. Eine zweite Liste
  liefe beim elften Schalter auseinander.
- **Eine Weiterleitung von der abgeschalteten Seite zum Start.** Sie ließe
  rätseln, warum die Taste nichts tut (§6: leere Zustände erklären den
  nächsten Schritt).

## Was es kostet

- **Ein Lesen der Einstellungen je Ruf** auf einem Weg aus der Tafel. Am
  Dienst und am Werkzeugserver über HTTP ist es jede Anfrage, weil ihr eigener
  Schalter gefragt wird. Ein Weg des Kerns kostet am Fenster nur den Blick in
  den Aufrufrahmen.
- **Ein Lauf, der startet, während sein Schalter ausgeht, endet selbst**
  (Zweitblick über Phase 6.12, Z2, nachgetragen am 2026-10-01).
  - Das Tor fragt vor dem Start, und das Abschalten beendet, was
    eingetragen ist. Dazwischen liegt der Start selbst.
  - Deshalb fragt jeder Start nach dem Eintragen noch einmal: erst
    eintragen, dann fragen, gegen erst schreiben, dann beenden.
  - Die Antwort ist der Stand danach, und der sagt „läuft nicht“.
  - **Nachlauf, N2: der Stand allein reichte nicht.** Ein Start, der aus und
    wieder an erlebte, sah danach nur „an“ und lief weiter. Deshalb zählt
    das Abschalten je Lauf mit. Ein Start nimmt vorher seine Marke und
    vergleicht danach mit ihr.
  - Beendet wird **nur der eigene Lauf**. Eine verspätete Nachfrage hätte
    sonst einen neuen Bildauftrag mitgenommen.
- **Die Gäste am eigenen Tisch gehen mit `spielen` und `netz`** (Z3). Das
  Ende der Freigabe allein schickt weiterhin niemanden vom Tisch.
  - **Nachlauf, N1:** ein Gast, der sich einträgt, nachdem das Abschalten
    die Runde entlassen hat, fragt mit derselben Marke nach und geht.
- **Dienst und Werkzeugserver fragen ihren eigenen Schalter vor dem Start
  und vor jeder Anfrage** (Z7).
  - **Abgeschaltet beim Start heißt: Rückgabe 0** (Nachlauf, N4). Der
    LaunchAgent startet einen Dienst neu, der mit einem Fehler endet
    (`KeepAlive` mit `SuccessfulExit: false`, `man launchd.plist`). Mit
    Rückgabe 1 liefe ein abgeschalteter Dienst immer wieder an — und nach
    dem Einschalten von selbst. Der Grund steht auf der Fehlerausgabe.
  - Vorher öffneten beide ihren Port auch mit Schalter aus.
  - Der Werkzeugserver über HTTP fragte nur vor einem Werkzeug, und `ping`
    ging durch.
- **Ein Lesefehler nach dem Schreiben nimmt im Fenster nichts zurück** (Z9).
  Der eben geschriebene Schalter gilt, und der Fehler steht da.
- **Ein Lauf, den ein anderer Prozess abschaltet, endet erst beim nächsten
  Weg des Fensters, der den Stand liest.** Bis dahin hält er seinen Port —
  das Fenster beobachtet die Einstellungen nicht. Eine Seite, die schon
  offen ist, zeigt ihn weiter als laufend, bis sie fragt.
- **Die Tafel enthält auch Bezeichner, die keine Wege sind**, zum Beispiel
  Hilfsmethoden aus der Fensterspalte. Sie werden nie gefragt und schaden
  nicht. Weggefiltert werden sie nicht, damit die Tafel ist, was der Katalog
  sagt.
