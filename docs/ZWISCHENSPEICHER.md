# Zwischenspeicher

**Was diese Anwendung aufhebt, um es nicht zweimal zu holen oder zu rechnen**
— OPS-08 (#1725).
Je Speicher: was er hält, woran sein Schlüssel hängt, wo er liegt, wie groß er
werden darf, wann er leer wird und was er bringt.

**Zu sehen und zu leeren** unter *Einstellungen › Diagnose › Zwischenspeicher*:
je Speicher Ort, Trefferquote, Einträge, Größe und Grenze, und ein Knopf, der
alle leert. Die Zahlen sind die **dieses Prozesses** und zählen ab seinem
Start.

**Nichts davon sind Daten.** Was hier steht, lässt sich jederzeit wegwerfen;
es füllt sich beim nächsten Gebrauch von selbst. Die Daten — Bestand, Bilder,
Pakete — liegen unter `~/Library/Application Support/MagicSetEditor4Mac`, die
Zwischenspeicher auf der Platte unter `~/Library/Caches/MagicSetEditor4Mac`.
**Bis OPS-08 lag der von Scryfall neben dem Bestand**; beim ersten Start zieht
er einmal um — **außer bei einem umgelenkten Bestand** (`MSE_BESTAND`): Läufe
und Proben sollen den Speicher einer installierten Anwendung nicht wegtragen.
**Der Umzug führt zusammen**: liegt am neuen Ort schon etwas — etwa von einem
Lauf mit umgelenktem Bestand, der den Speicher des Rechners mitbenutzt —,
kommt der alte Inhalt dazu, und was an beiden Orten liegt, behält den neuen
Stand. Bis zum Zweitblick über Phase 6.4 ging der Umzug nur in einen fehlenden
Ort und fand nach einem solchen Lauf nie mehr statt.

---

## Die acht

| Speicher | Ort | Hält | Schlüssel | Grenze | Leer wird er |
| --- | --- | --- | --- | --- | --- |
| `scryfall.antworten` | Platte, `Caches/…/scryfall` | Antworten von Scryfall als JSON | Anfrage (Weg und Frage), Prüfsumme als Dateiname | `zwischenspeicher.grenze`, Vorgabe 512 MB, ab dem Start | beim Leeren; ein Eintrag gilt 24 Stunden als frisch und wird danach neu geholt |
| `scryfall.bilder` | Platte, `…/scryfall/bilder` | Kartenbilder von Scryfall | Druck und Bildart | dieselbe Grenze | beim Leeren |
| `scryfall.sammeldaten` | Platte, `…/scryfall/sammeldaten` | die Sammeldatei `oracle_cards` | — **eine** Datei | keine: ein neuer Stand ersetzt sie | beim Leeren |
| `render.ebenen` | Arbeitsspeicher | abgetastete Bildebenen | Bildquelle (Zeiger), Größe, Güte | 64 MB, dann ganz geleert | beim Leeren, beim Neustart |
| `pakete.bilder` | Arbeitsspeicher | Bilder aus Vorlagenpaketen | Paketpfad | 512 Bilder **oder 128 MiB** entschlüsselter Bildpunkte, dann ganz geleert; ein Bild über 128 MiB wird nicht gemerkt. **512 seit PERF02** (#2352): unter 64 räumte die Prüfung eines ganzen Sets ständig — `Black` braucht 255 verschiedene | beim Leeren, beim Neustart |
| `skript.baeume` | Arbeitsspeicher, **ein** Speicher je Prozess | zerteilte MSE-Skripte | der Quelltext selbst | 4096 Bäume **oder 64 MiB** (geschätzt), dann ganz geleert; ein Baum über 64 MiB wird nicht gemerkt | beim Leeren, beim Neustart |
| `vorschau.bilder` | Arbeitsspeicher | Kartenvorschauen für das Fenster | Inhalt (Prüfsumme) | 8 Bilder, das älteste geht | beim Leeren, beim **Profilwechsel**, beim Neubeginn nach einem Zurückspielen |
| `tisch.texturen` | Arbeitsspeicher | Kartenbilder am Tisch | Zeichenplan, Bild- und Schriftabdrücke, Seite, Größe, Güte, **Fassung des Rasterers** | 128 MB, das älteste geht | beim Leeren, beim **Profilwechsel** |

**Und was hier bewusst nicht steht**: der Suchindex (`retrieval_…`) und die
Preisreihen liegen **im Bestand** und sind Daten mit eigener Pflege — ein
Neuaufbau des Index ist ein eigener Vorgang (A15). Der Bildvorrat der KI
(`Caches/…/bildvorschlaege`) hat seine Grenzen in
`internal/service/bildvorratschranken.go` und leert sich beim Start.

---

## Schlüssel, Profile und Zurückspielen (AK-02)

**Kein Speicher liefert ein fremdes Profil**, und das hat je Speicher einen
eigenen Grund:

- **Öffentliche Daten**: die drei von Scryfall halten, was jeder bekommt, der
  fragt. Ein Profil kommt im Schlüssel nicht vor, weil es in der Antwort nicht
  vorkommt.
- **Inhalt als Schlüssel**: Vorschauen, Spielbilder und Skriptbäume sind nach
  dem benannt, was sie sind. Dieselbe Karte in einem anderen Profil ist ein
  anderer Schlüssel. **Trotzdem leert ein Profilwechsel Vorschauen und
  Spielbilder**: eine Adresse, die im vorigen Profil entstand, soll im nächsten
  nichts mehr liefern.
- **Pakete**: Ebenen und Paketbilder hängen an Vorlagen, die allen Profilen
  gehören. **Eine geänderte Paketdatei gilt nach höchstens einer Sekunde**:
  der Stilspeicher prüft den Stempel der Pakete, aus denen ein Stil entstand
  (`internal/service/stilstempel.go`), und meldet er einen Wechsel, leert das
  Zeichenwerk Paketbilder und Ebenen mit
  (#2005).
  **Was der Stempel nicht deckt**, deckt auch das Leeren nicht: eine Schrift,
  die jemand in ein vorhandenes Paket legt, und ein Bild aus einem Paket, das
  kein Stil eines gezeichneten Sets las.
- **Die Flächen ohne Fenster behalten ihr Zeichenwerk** — `mse-mcp`,
  `mse-rest` und die Kommandozentrale, je Prozess und Bestand, höchstens vier
  (`internal/app/zeichenvorrat.go`, #2005). Ebenen, Paketbilder,
  Schriftsuche und Stile gelten damit dort wie im Fenster. **Neu gebaut**
  wird es, wenn sich Bilderverzeichnis, Schriftwurzel, die Paketwurzeln oder
  deren Änderungszeit ändern — ein Paket, das in eine Wurzel kommt oder aus
  ihr geht. **Den Stil eines Sets** liest der Stilspeicher je Ruf am Setkopf:
  ein Wechsel aus einem anderen Prozess gilt beim nächsten Bild, und ein Ruf
  aus einem Profil, das das Set nicht führt, hinterlässt nichts.
- **Zurückspielen**: es startet die Anwendung über `Beenden` und `Starten`
  neu. Das Zeichenwerk entsteht dabei neu — Ebenen, Paketbilder und
  Spielbilder mit ihm —, und die Vorschauen leert `Starten` eigens. Übrig
  bleiben die Skriptbäume, deren Schlüssel der Quelltext ist, und die
  öffentlichen Antworten.
- **Rechte**: kein Speicher hält eine Rechteentscheidung. Wer in einem Profil
  etwas nicht mehr darf, fragt den Riegel, und der fragt den Bestand.

---

## Grenzen, Ablauf, Leeren (AK-03)

**`zwischenspeicher.grenze`** (Megabyte, 1 bis 8192, Vorgabe 512, ab dem
nächsten Start): je Speicher auf der Platte, für Antworten und Bilder von
Scryfall. Beim Start gehen die ältesten Dateien — nach der Uhr der Platte —,
bis die Grenze eingehalten ist. Eine Datei, die nicht nach einem Eintrag
aussieht, bleibt liegen. **Beim Start jedes Prozesses, der lange läuft** —
Fenster, `mse-rest`, `mse-mcp`; bis zum Nachlauf über Phase 6.4 räumte nur das
Fenster, und ein Dienst, der Wochen lief, nie. **Dazwischen wächst er**: eine
Größenprüfung je Eintrag kostete einen Gang über den Ordner je Antwort. Die
Kommandozeile räumt nicht — sie lebt einen Befehl lang.

**Kein einstellbarer Ablauf, und das ist entschieden.** Scryfall bittet, seine
Daten **mindestens** 24 Stunden zu halten; kürzer wäre unhöflich, länger zeigte
die Anwendung ältere Preise, ohne dass es jemand bestellt hat. Ohne Netz
springt der letzte Stand ein und sagt, wie alt er ist (§16, Satz 5). Die
Speicher im Arbeitsspeicher verfallen nicht nach Zeit, sondern an ihrer
Grenze — ihr Inhalt veraltet nicht, weil ihr Schlüssel der Inhalt ist.

**Leeren** wirft alle acht weg — am Fenster, unter *Einstellungen › Diagnose*.
**Nur am Fenster**: die Speicher im Arbeitsspeicher gehören dem Prozess, der
sie füllt — ein Dienst behält seine bis zu seinem Ende —, und wer eine
Kommandozeile hat, leert die auf der Platte mit
`rm -r ~/Library/Caches/MagicSetEditor4Mac/scryfall`. Über Dienst und
Werkzeugserver gibt es kein Leeren: es gäbe jedem Token einen Weg, alle
Antworten neu holen zu lassen.

**Was gerade geschrieben wird, bleibt** — die Nachbardatei eines Bildes
(`.<name>.<zufall>`) und die `.teil` einer ladenden Sammeldatei zählen weder
zur Belegung noch fallen sie beim Leeren oder Beschneiden. Bis zum Zweitblick
über Phase 6.4 taten sie beides, und ein Leeren nahm einem laufenden Laden die
Datei unter den Füßen weg.

**Der Wiederaufbau ist der nächste Gebrauch.** Nichts wird auf Vorrat geholt:
ein geleerter Speicher füllt sich, wenn jemand eine Karte sucht, zeichnet oder
an den Tisch legt. Ohne Netz kommen die Antworten und Bilder von Scryfall erst
wieder, wenn es da ist — das steht vor dem Knopf.

---

## Was sie bringen (AK-04)

**Kalt gegen warm, gemessen** mit

```bash
go test -run '^$' -bench 'KaltUndWarm' -benchtime 2s ./internal/render ./internal/msescript ./internal/netz
```

**Gemessen am 2026-09-26** auf einem Apple M5 (Mac17,3), macOS 27.0, Go
1.27.1, mit `-benchtime 2s -benchmem`:

| Speicher | Kalt | Warm | Was ein Eintrag kostet |
| --- | ---: | ---: | --- |
| `render.ebenen` — eine Ebene 750 × 1050 auf 375 × 523, Güte hoch | 5,46 ms, 788 KB | 17 ns | 784 KB Arbeitsspeicher — **die 64 MB halten gut achtzig Ebenen** |
| `skript.baeume` — ein Skript von zwanzig Zeilen | 53,8 µs, 137 KB | 18 ns | ein Baum; 4096 davon halten sieben Sets mit eigenem Spiel und Stil (`internal/msescript`, `Arbeitsmenge`) |
| `scryfall.antworten` — 8 KB, das Netz **nachgestellt mit 50 ms** | 52,7 ms | 21 µs | 8 KB und die JSON-Hülle auf der Platte |

### Die Bytegrenzen und der Arbeitssatz (PERF03)

**Bis PERF03** (#2353,
Befund A-19) zählten `pakete.bilder` und `skript.baeume` nur Einträge, und
unter „64 Bilder" ließen sich Gigabyte halten — ausgepackt ist je Bild bis
1 GiB erlaubt ([0160](entscheidungen/0160-was-ein-bild-ausgepackt-kostet.md)).
Seitdem zählen sie Bytes, und die Diagnose zeigt Bytes und Grenze.

**Wie gezählt wird**: ein Paketbild mit den Puffern seiner Bildpunkte, je Art
so, wie `image` sie anlegt (`internal/msepaket/bildbelegung.go`); ein Baum
**geschätzt** aus Knoten, Texten, Kinder- und Namensfeldern und dem Schlüssel
— eher zu viel als zu wenig (`internal/msescript/baumbelegung.go`).

**Ob die Grenzen den Arbeitssatz tragen, gemessen am 2026-10-14** mit dem
Kanonlauf (`TestDieKanonischenSetsZeichnen`, Kennzahlen `speicher.*`) über
die zehn kanonischen Sets, Stichprobe 68 Karten je Set, 680 Karten, voller
Pack, in **einem** Prozess:

| Speicher | belegt | Einträge | Treffer | Fehlgriffe | Grenze |
| --- | ---: | ---: | ---: | ---: | ---: |
| `skript.baeume` | 12,6 MB | 2.113 | 1.214.190 | 2.113 | 64 MiB |
| `pakete.bilder` | 7,2 MB | 32 | 42.806 | 19.654 | 128 MiB |

Beide bleiben weit unter der Grenze, und keiner räumt. **Was die Zahl nicht
sagt**: die Fehlgriffe der Bilder sind zum großen Teil Bilder, die es nicht
gibt — ein Fehlschlag wird nicht gemerkt. Und es ist eine Stichprobe: wie
viele **verschiedene** Skripte alle 25.148 Karten brauchen, hat dieser Lauf
nicht gezählt. Der Kanonlauf wird rot, wenn ein Speicher öfter frisch holt
als trifft (`kanonspeicher_intern_test.go`).

**Was die Zahlen sagen, und was nicht**: *warm* ist ein Treffer, *kalt* ein
Fehlgriff. Ob ein Speicher im Betrieb trifft, sagt die **Trefferquote** unter
*Einstellungen › Diagnose* — eine Messung hier belegt den Gewinn je Treffer,
nicht wie oft er eintritt. Das Netz ist nachgestellt: gegen Scryfall selbst
wäre die Messung eine Last für einen fremden Dienst (§16, Satz 4).

**Warum keine Grenze eine langsame Stelle verdeckt**: jede Grenze steht fest
im Code oder in einer Einstellung mit Obergrenze, und keine wächst mit der
Last. Ein Pfad, der ohne Speicher langsam ist, bleibt es beim ersten Mal — die
Spalte *kalt* oben ist genau diese Zahl, und die Laufzeiten unter
*Metriken* (OPS-05) messen die Rufe, nicht den Speicher.
