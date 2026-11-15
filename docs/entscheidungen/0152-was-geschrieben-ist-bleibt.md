# 0152 — Was geschrieben ist, bleibt: Protokoll, Belegschutz, bedingtes Fortschreiben

**Stand:** entschieden am 2026-09-30 · TRADE-03.AK-01, AK-03, AK-04 und NAT-01,
Schnitt 2 (#1717)

## Die Frage

Schnitt 1 ([0149](0149-was-eine-handelsnachricht-behaupten-darf.md)) hat
festgelegt, was eine empfangene Nachricht behaupten darf. Offen blieben vier
Dinge:

- Wie ein Vorgang **dorthin kam**: eine Revision hält in `handelsrevisionen`
  eine Zeile, und ein Schritt überschreibt ihren Zustand (0068). Der Weg von
  *angeboten* nach *angenommen* stand nirgends.
- Ob ein **abgeschlossener Beleg** sich still ändern lässt.
- Was geschieht, wenn **zwei Schreiber** dieselbe Revision fortschreiben,
  etwa das Fenster und ein Brief, der gerade ankommt.
- Was aus einem **Widerspruch in der Übergabe** wird.

## Was MSE2 tut

**Es handelt nicht** (siehe [0149](0149-was-eine-handelsnachricht-behaupten-darf.md)).
Marke **new**.

## Was gilt

1. **Das Protokoll schreibt der Bestand, nicht die Anwendung.** Zwei
   Auslöser an `handelsrevisionen` (Wanderung 0095) legen je neue Revision
   und je Zustandswechsel eine Zeile an: Revision, von, nach, Zeit. Damit
   landet jeder Weg darin, der eine Revision schreibt, also Leitung,
   Briefkasten, Abschluss, Kommandozeile, Dienst und MCP. Ein zweites
   Schreiben desselben Zustands legt **keine** Zeile an.
2. **Das Protokoll enthält nur Zustände.** Es nennt keine Karten, Namen,
   Unterschriften oder Abdrücke. Was es weiß, steht ohnehin in der Revision,
   und es ist auf Nachvollzug ausgelegt, nicht auf Inhalt (AK-03: sensible
   Daten begrenzt).
3. **Beleg, Journal, Fremdbeleg und Protokoll lassen sich nicht ändern
   und nicht löschen**, auch nicht mit SQL an der Anwendung vorbei. Das
   halten Auslöser, die mit einem Fehler abbrechen. Zwei Wege bleiben offen:
   - die **Gegenzeichnung** setzt `beidseitig` von 0 auf 1, und nur so;
   - ein **gelöschtes Profil** nimmt alles mit. Die Bedingung fragt, ob es
     das Profil noch gibt, und während der Löschkaskade gibt es das nicht
     mehr.

   Eine Korrektur ist eine **neue** Zeile, etwa über die Klärung, und kein
   Umschreiben.
4. **Fortschreiben ist bedingt** (AK-01). `Fortschreiben(vorher, danach)`
   schreibt dieselbe Revision nur, wenn Zustand **und** Zustimmende noch so
   sind, wie der Schreiber sie gelesen hat. Eine neue Revision wird nur
   angelegt, wenn es sie noch nicht gibt. Hat inzwischen jemand anderes
   geschrieben, lautet die Antwort `ErrUeberholt`, und nichts wird still
   überschrieben. **Das gilt auf allen drei Schreibwegen** — Leitung,
   Briefkasten und Abschluss —, **und eine neue Revision entsteht nur auf
   dem gelesenen Vorgänger**, wenn er mit Zustand und Stimmen noch dasteht
   (Befund N1 des Nachlaufs). **Eröffnen und Kopieren legen nur an, was es
   noch nicht gibt** (N2); ein unbedingtes Schreiben gibt es im Dienst nicht
   mehr. Die zwei letzten schrieben bis zum Zweitblick
   über Phase 6.8 unbedingt (Befund Z1): eine Stimme, gegen `angeboten`
   gelesen, setzte einen angenommenen Handel still zurück, und zwei
   Verbuchungen, die beide `uebergabeOffen` gelesen hatten, buchten beide.
   Der Briefkasten bekommt den gelesenen Stand mit; ist er überholt, fällt
   auch der Vermerk. **Die Nachricht bleibt im Ausgang des Absenders** und
   geht mit der nächsten Nachlieferung wieder hinaus — gegen den neuen
   Stand. **Das geschieht auf Aufruf** („Jetzt übertragen“,
   `mse handelszustellung`), nicht von selbst — **und dabei bleibt es**,
   entschieden mit NET-02.AK-04
   (#1706;
   Befund N4 des Nachlaufs über Phase 6.8): eine Wiederholung ohne Zutun
   hätte kein Ende, weil erst ein Mensch die Ankunft bestätigt, und sie wäre
   der erste Takt, der einen fremden Rechner anwählt. Der Grund steht in
   [`docs/NETZEIGENSCHAFTEN.md`](../NETZEIGENSCHAFTEN.md), Abschnitt *Was
   nach einem Abriss gilt*.
5. **Derselbe Fremdbeleg zweimal ist kein Fehler, ein anderer schon.** Die
   Gegenzeichnung ist wiederholbar. Kommt für denselben Vorgang ein
   **abweichender** Beleg der Gegenseite, antwortet sie mit einem Konflikt:
   der Handel gehört in die Klärung.
6. **Ein Konflikt in der Übergabe führt in die Klärung** (AK-04). Widerspricht
   eine empfangene Nachricht dem eigenen Stand, und zwar dort, wo die Matrix
   `klaeren` kennt (Übergabe offen, lokal verbucht, Abgleich), geht **diese
   Seite** den Schritt. Der Absender geht ihn nicht; wir haben den
   Widerspruch gesehen. In der Verhandlung bleibt ein abweichender Vorschlag,
   was er ist: ein Gegenangebot. **Auch nach einem verlorenen Wettlauf im
   Briefkasten** entscheidet derselbe Vergleich der Abdrücke über bekannt
   oder Konflikt; bis zum Zweitblick über Phase 6.8 hieß es dort immer
   „bekannt" (Befund Z2).
7. **Das Protokoll ist an allen vier Flächen lesbar**:
   - `HandelProtokoll`,
   - `lesen.handelsprotokoll`,
   - `mse handel protokoll`,
   - `GET /handel/{vorgang}/protokoll`.

   Die Änderung ist **erweiternd**. Im Fenster steht es als Kasten *Verlauf*
   in *Verbuchen*.

## Gemessen

- **Hüllenprobe `trade03nat`** (NAT-01) am nativen Fenster: ein Tausch in der
  Übergabe wird verbucht, gleich noch einmal verbucht, die Anwendung beendet
  und neu gestartet und ein drittes Mal verbucht. Das Ergebnis:
  - **eine** Bewegung im Journal (+1 × bolt-m10),
  - **ein** Wechsel nach *lokal verbucht* im Protokoll,
  - der Stand „Bei dir verbucht · Nur du hast unterschrieben", mit dem Satz,
    dass von der Gegenseite noch kein geprüfter Beleg vorliegt,
  - als Nächstes „Du bist dran: Abgleich melden".
- **Hüllenprobe `trade03`**: der Verlauf nennt die drei Wechsel des Tauschs
  mit lesbarer Zeit.
- **Läufe**:
  - jede Manipulation an Beleg, Journal, Fremdbeleg und Protokoll scheitert;
  - ein gelöschtes Profil hinterlässt nichts;
  - zwei Schreiber an derselben Revision ergeben einen Erfolg und ein
    `ErrUeberholt`;
  - ein Konflikt im Briefkasten führt in die Klärung;
  - dasselbe Protokoll kommt über Kommandozeile, Dienst und MCP.

## Was das nicht hält

**Ein Mensch mit `sqlite3` und der Datei in der Hand kann einen Auslöser
löschen.** Die Auslöser schützen gegen die Anwendung, gegen Werkzeuge und
gegen Versehen, nicht gegen den Besitzer der Datei. Gegenüber der Gegenseite
ist der **unterschriebene Beleg** der Nachweis ([0149](0149-was-eine-handelsnachricht-behaupten-darf.md)).
Ein umgeschriebener Beleg passt nicht mehr zu ihrer Unterschrift.

**Die Zeit vor Wanderung 0095 ist nachgetragen**, und zwar als eine Zeile je
Revision mit ihrem damaligen Zustand und dem Zeitpunkt ihrer letzten
Änderung. Die Zwischenschritte davor kennt niemand mehr.

## Verworfen

- **Das Protokoll in der Anwendung schreiben**: jeder neue Weg, der eine
  Revision schreibt, müsste daran denken. Der Briefkasten hätte es als
  Erster vergessen.
- **„Mit Spuren überschreibbar"** über Fassungen je Beleg: das wäre ein
  zweites Belegmodell mit eigener Gültigkeit. Ein Beleg, der sich nicht
  ändert, und eine Korrektur als neue Zeile sagen dasselbe einfacher.
- **Eine Hashkette über das Protokoll**: gegen den Besitzer der Datei hilft
  sie so wenig wie ein Auslöser, und gegenüber der Gegenseite gibt es die
  Unterschrift schon.
