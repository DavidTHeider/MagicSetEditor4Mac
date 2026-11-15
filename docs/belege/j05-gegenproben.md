# J05 — die nachgeholten Gegenproben

**Gelaufen am 2026-10-11** — J05
(#2269),
Apple M5, macOS 27.0.1, go1.27.2, Vitest 5.0.3, in einem eigenen,
losgelösten Arbeitsbaum.

**Die Frage**: Für Commits mit Produktcode in E01, E07, F02, der Folge
#2297, G08, G10, G11 und G12 war kein roter Lauf festgehalten
(`audits/0.8.0/UMSETZUNG.md` (Entwicklungsarchiv), Abschnitt 3).
Greifen ihre Tests? **Das Verfahren**: je Commit seine eigenen Testdateien
gegen den Stand **davor** (`<commit>^`) — dort müssen sie rot sein — und
gegen den Stand **danach** — dort grün. Gefahren wird je Testdatei nur, was
sie definiert (`-run '^(Test…|Fuzz…)$'`), im Fenster die Datei allein.

**Was das belegt und was nicht**: dass der Test die Umsetzung erkennt. Dass
er **vor** ihr da war, belegt es nicht — Test und Umsetzung stehen in
denselben Commits. Ein „rot" kann auch ein Übersetzungsfehler sein: ein Test,
der eine neue Funktion ruft, übersetzt ohne sie nicht. Das ist die Gestalt,
in der ein testgetriebener Anfang in Go rot ist.

## Das Skript

```bash
#!/usr/bin/env bash
# gegenprobe.sh <commit> — die Tests eines Commits gegen den Stand davor und den Stand danach.
set -uo pipefail
c="$1"; cd /Users/david/MagicSetEditor4Mac-gp
source scripts/activate.sh >/dev/null 2>&1
tests=$(git show --format= --name-only "$c" | grep -E '_test\.go$|\.test\.tsx?$')
lauf() { # lauf <etikett>
  local gesamt=0
  for t in $tests; do
    [ -f "$t" ] || { echo "  $1 $t: fehlt"; continue; }
    case "$t" in
      *.go)
        namen=$(grep -oE '^func (Test|Fuzz)[A-Za-z0-9_]+' "$t" | sed 's/^func //' | paste -sd'|' -)
        [ -n "$namen" ] || continue
        aus=$(nice -n 10 go test -count=1 -run "^(${namen})\$" "./$(dirname "$t")" 2>&1); rc=$?
        echo "  $1 $t: rc=$rc $(printf '%s\n' "$aus" | grep -E '^(--- FAIL|FAIL|ok)|undefined|cannot|not enough|too many' | head -3 | tr '\n' ' ' | cut -c1-220)" ;;
      *)
        aus=$(cd frontend && nice -n 10 ./node_modules/.bin/vitest run "${t#frontend/}" 2>&1); rc=$?
        echo "  $1 $t: rc=$rc $(printf '%s\n' "$aus" | grep -E 'Tests +[0-9]|FAIL|✗|×' | head -3 | tr '\n' ' ' | cut -c1-220)" ;;
    esac
  done
}
echo "== $c $(git log -1 --format=%s "$c" | cut -c1-90)"
git checkout -q -f --detach "$c^" && git checkout -q "$c" -- $tests && lauf vorher
git checkout -q -f --detach "$c" && lauf nachher
git checkout -q -f --detach main
```

Zwei Commits, die nur die Zustandsprobe des Fensters ändern (`bca5b2a7e`,
`5ff419845`), deckt das Skript nicht ab — ihr Test ist eine `.test.mjs`. Sie
liefen von Hand mit `node --test scripts/zustaende.test.mjs`: davor 0 von 1
und 6 von 8 Fällen grün, danach 7 von 7 und 8 von 8.

## Ergebnis

| Commit | Betreff | Testdateien rot davor | grün danach |
| --- | --- | --- | --- |
| `931c7aa0b` | fix(2238): eine gescheiterte Vorschau beim Tippen steht an der Vorschau | 1 von 1 | 1 von 1 |
| `6e09f6b5e` | fix(2279): Bildchen der Kartenliste erst, wenn das Rollen innehält | 1 von 1 | 1 von 1 |
| `38522f271` | feat(2244): Text wirft seinen Schatten wie bei MSE2 | 4 von 4 | 4 von 4 |
| `2eab47f21` | fix(2244): ein Bildfeld ohne Bild zeigt das Schachbrett wie MSE2 | 2 von 2 | 2 von 2 |
| `ac89f51f7` | fix(2244): to_int, to_real und to_number rechnen eine Farbe wie MSE2 | 1 von 2 | 2 von 2 |
| `d4ca75c8d` | fix(2244): die Schriftsuche findet den Schnitt Bold einer Familie | 2 von 2 | 2 von 2 |
| `53a806608` | fix(2244): weight und style aus dem font-Block gelten fürs ganze Feld | 3 von 3 | 3 von 3 |
| `9bcd4246f` | fix(2244): kursiv ohne italic name nimmt den Schnitt Italic der Familie | 3 von 3 | 3 von 3 |
| `e28c47fa4` | fix(2244): ein zu langer Text wird gesucht wie bei MSE2 | 2 von 2 | 2 von 2 |
| `4668ae89c` | feat(2297): eine Mehrfachwahl als Reihe ihrer Wahlbilder zeichnen | 3 von 3 | 3 von 3 |
| `c4f161959` | feat(g08): die Lauschwahl der eigenen Dienste als Einstellung | 5 von 5 | 5 von 5 |
| `cce1294b1` | feat(g08): die Schnittstellen dieses Rechners zum Wählen | 1 von 2 | 2 von 2 |
| `7edd2ebf9` | feat(g08): ein Dienst lauscht auf jeder gewählten Adresse | 1 von 1 | 1 von 1 |
| `e29917dcd` | feat(g08): die Ansage nennt jede Adresse und jede fehlende | 1 von 1 | 1 von 1 |
| `e2d3ac44e` | feat(g08): mse-rest und mse-mcp lauschen auf jeder gewählten Adresse | 5 von 5 | 5 von 5 |
| `e8ac6c6c6` | feat(g08): der Dienst aus dem Fenster startet auf der ganzen Lauschwahl | 6 von 9 | 9 von 9 |
| `b055a0ea2` | feat(g08): alle Ziele mit einem Ruf einrichten und zurücknehmen | 1 von 1 | 1 von 1 |
| `200684fb6` | feat(g08): ein Knopf richtet alles ein, und die Schnittstellen sind wählbar | 4 von 5 | 5 von 5 |
| `b8cc3f23b` | feat(g08): der Dienst im Fenster folgt der Lauschwahl und nennt jeden Lauscher | 2 von 2 | 2 von 2 |
| `6f19ce4f5` | feat(g10): Schreibknöpfe für einen Leser gesperrt und erklärt | 6 von 6 | 6 von 6 |
| `c90b06be6` | fix(g10): gesperrte Markenknöpfe sehen aus wie ein gesperrter Knopf | 1 von 1 | 1 von 1 |
| `f187b00c1` | fix(g10): Aufnehmen im Leerzustand und die Freigabe für einen Leser gesperrt | 4 von 4 | 4 von 4 |
| `a431ebad3` | fix(g10): Einfügen im leeren Set ist für einen Leser gesperrt | 2 von 2 | 2 von 2 |
| `9f1b7540c` | test(g10): der Lesefehlerfall wartet auf die Absage statt auf den Ruf | 0 von 1 | 1 von 1 |
| `a67f474bd` | feat(g11): Startseite, Sammlung und Ausgabeauswahl sagen, dass sie laden | 3 von 3 | 3 von 3 |
| `bca5b2a7e` | test(g11): die Zustandsprobe misst, ob eine Seite ihre Bindungen zeigt | 0 von 0 | 0 von 0 |
| `5ff419845` | fix(g11): die Zustandsprobe vergleicht auch den Bedienzustand | 0 von 0 | 0 von 0 |
| `e728df085` | test(g): drei Zweige der Oberfläche abgedeckt, die G02 und G11 brachten | 0 von 3 | 3 von 3 |
| `d95c75875` | fix(g12): die Freigabe für den Neuanfang lässt sich am Fenster ausstellen | 3 von 4 | 4 von 4 |
| `1331f3df7` | test(g12): die Freigabetests in Prettier-Form | 0 von 1 | 1 von 1 |
| `0a26e577b` | fix(g12): der Neuanfang zählt jedes Profil des Bestands | 4 von 4 | 4 von 4 |

**Jeder Commit mit Produktcode hat mindestens eine Testdatei, die davor rot
und danach grün ist.** Grün davor sind nur Dateien, die der Commit
mitverändert, ohne dass sie das Neue prüfen — in jedem solchen Commit ist
eine andere Datei rot. Ohne roten Lauf sind allein reine Testcommits: das
Abdecken von Zweigen (`e728df085`), die Form (`1331f3df7`) und das Warten
eines Tests (`9f1b7540c`).

## Die Rohausgabe

Je Zeile gekürzt auf 220 Zeichen.

```
== 931c7aa0b fix(2238): eine gescheiterte Vorschau beim Tippen steht an der Vorschau
  vorher frontend/src/seiten/tippvorschaufehler.test.tsx: rc=1    × meldet eine gescheiterte Probevorschau und erholt sich mit der nächsten Eingabe 15103ms ⎯⎯⎯⎯⎯⎯⎯ Failed Tests 1 ⎯⎯⎯⎯⎯⎯⎯  FAIL  src/seiten/tippvorschaufehler.test.tsx > meldet eine gesche
  nachher frontend/src/seiten/tippvorschaufehler.test.tsx: rc=0       Tests  1 passed (1) 
== 6e09f6b5e fix(2279): Bildchen der Kartenliste erst, wenn das Rollen innehält
  vorher frontend/src/haken/listenbilderruhe.test.tsx: rc=1    × fragt beim schnellen Rollen nur das Fenster, bei dem die Liste innehält 16ms ⎯⎯⎯⎯⎯⎯⎯ Failed Tests 1 ⎯⎯⎯⎯⎯⎯⎯  FAIL  src/haken/listenbilderruhe.test.tsx > fragt beim schnellen Rollen nur 
  nachher frontend/src/haken/listenbilderruhe.test.tsx: rc=0       Tests  1 passed (1) 
== 38522f271 feat(2244): Text wirft seinen Schatten wie bei MSE2
  vorher internal/msepaket/schriftschatten_test.go: rc=1 internal/msepaket/schriftschatten_test.go:58:9: s.SchattenFarbe undefined (type msepaket.Schrift has no field or method SchattenFarbe) internal/msepaket/schriftschatten_test.go:58:42: s.SchattenX undefined (type msepaket
  vorher internal/render/textschatten_intern_test.go: rc=1 internal/render/textschatten_intern_test.go:31:45: undefined: Textschatten internal/render/textschatten_intern_test.go:32:43: undefined: Textschatten internal/render/textschatten_intern_test.go:40:45: undefined: Textscha
  vorher internal/service/schriftschatten_test.go: rc=1 internal/service/schriftschatten_test.go:25:19: undefined: render.Textschatten internal/service/schriftschatten_test.go:27:46: undefined: render.Textschatten internal/service/schriftschatten_test.go:31:11: undefined: ren
  vorher internal/svg/textschatten_test.go: rc=1 internal/svg/textschatten_test.go:20:22: plan.Ebenen[2].Satz.Schatten undefined (type render.Satz has no field or method Schatten) internal/svg/textschatten_test.go:20:40: undefined: render.Textschatten internal/svg/text
  nachher internal/msepaket/schriftschatten_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket	0.255s 
  nachher internal/render/textschatten_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/render	0.216s 
  nachher internal/service/schriftschatten_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/service	0.326s 
  nachher internal/svg/textschatten_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/svg	0.174s 
== 2eab47f21 fix(2244): ein Bildfeld ohne Bild zeigt das Schachbrett wie MSE2
  vorher internal/render/platzhalter_intern_test.go: rc=1 internal/render/platzhalter_intern_test.go:40:34: undefined: Platzhalterknoten FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/internal/render [build failed] FAIL 
  vorher internal/service/bildplatzhalter_test.go: rc=1 --- FAIL: TestEinLeeresBildfeldZeigtDasSchachbrett (0.00s) FAIL FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/internal/service	0.335s 
  nachher internal/render/platzhalter_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/render	0.209s 
  nachher internal/service/bildplatzhalter_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/service	0.321s 
== ac89f51f7 fix(2244): to_int, to_real und to_number rechnen eine Farbe wie MSE2
  vorher internal/msescript/bibliothek_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript	0.178s 
  vorher internal/msescript/umwandlungfarbe_test.go: rc=1 --- FAIL: TestEineFarbeWirdZurZahlIhrerKanaele (0.00s) FAIL FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript	0.082s 
  nachher internal/msescript/bibliothek_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript	0.192s 
  nachher internal/msescript/umwandlungfarbe_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/msescript	0.092s 
== d4ca75c8d fix(2244): die Schriftsuche findet den Schnitt Bold einer Familie
  vorher internal/schriften/familienschnitt_test.go: rc=1 FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften [build failed] FAIL 
  vorher internal/schriften/schriftlesen_test.go: rc=1 FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften [build failed] FAIL 
  nachher internal/schriften/familienschnitt_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften	0.168s 
  nachher internal/schriften/schriftlesen_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/schriften	0.068s 
== 53a806608 fix(2244): weight und style aus dem font-Block gelten fürs ganze Feld
  vorher internal/msepaket/schriftgewicht_test.go: rc=1 internal/msepaket/schriftgewicht_test.go:49:20: feld.Schrift.Gewicht undefined (type msepaket.Schrift has no field or method Gewicht) internal/msepaket/schriftgewicht_test.go:49:60: feld.Schrift.Lage undefined (type msep
  vorher internal/service/schriftgewicht_test.go: rc=1 FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/internal/service [build failed] FAIL 
  vorher internal/service/textbreitegewicht_test.go: rc=1 FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/internal/service [build failed] FAIL 
  nachher internal/msepaket/schriftgewicht_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket	0.212s 
  nachher internal/service/schriftgewicht_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/service	0.311s 
  nachher internal/service/textbreitegewicht_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/service	0.232s 
== 9bcd4246f fix(2244): kursiv ohne italic name nimmt den Schnitt Italic der Familie
  vorher internal/render/kursivschnitte_test.go: rc=1 internal/render/kursivschnitte_test.go:54:33: undefined: render.Kursivschnitte FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/internal/render [build failed] FAIL 
  vorher internal/service/textbreitegewicht_test.go: rc=1 --- FAIL: TestDieInhaltsbreiteMisstDenKursivenSchnittDerFamilie (0.00s) FAIL FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/internal/service	0.336s 
  vorher internal/werkzeuge/rasterfassung_test.go: rc=1 --- FAIL: TestDieRasterfassungPasstZumRasterer (0.01s) FAIL FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/internal/werkzeuge	0.534s 
  nachher internal/render/kursivschnitte_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/render	0.201s 
  nachher internal/service/textbreitegewicht_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/service	0.331s 
  nachher internal/werkzeuge/rasterfassung_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/werkzeuge	0.521s 
== e28c47fa4 fix(2244): ein zu langer Text wird gesucht wie bei MSE2
  vorher internal/render/massstabssuche_test.go: rc=1 --- FAIL: TestDerMassstabFolgtMSE2sSuche (0.00s) --- FAIL: TestEineLeereLetzteZeileZaehltBeimPassen (0.00s) --- FAIL: TestMitAbsatzbaendernZaehltDieHoechsteAbsatzhoehe (0.00s) 
  vorher internal/werkzeuge/rasterfassung_test.go: rc=1 --- FAIL: TestDieRasterfassungPasstZumRasterer (0.01s) FAIL FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/internal/werkzeuge	0.522s 
  nachher internal/render/massstabssuche_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/render	0.189s 
  nachher internal/werkzeuge/rasterfassung_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/werkzeuge	0.526s 
== 4668ae89c feat(2297): eine Mehrfachwahl als Reihe ihrer Wahlbilder zeichnen
  vorher internal/msepaket/feldstilreihe_test.go: rc=1 internal/msepaket/feldstilreihe_test.go:38:12: feld.Abstand undefined (type msepaket.Feldstil has no field or method Abstand) internal/msepaket/feldstilreihe_test.go:38:44: feld.Richtung undefined (type msepaket.Feldstil
  vorher internal/service/bildreihe_intern_test.go: rc=1 internal/service/bildreihe_intern_test.go:24:14: undefined: reihenanfang internal/service/bildreihe_intern_test.go:44:14: undefined: inRichtung internal/service/bildreihe_intern_test.go:53:18: undefined: wahlbildebene 
  vorher internal/service/bildreihe_test.go: rc=1 internal/service/bildreihe_intern_test.go:24:14: undefined: reihenanfang internal/service/bildreihe_intern_test.go:44:14: undefined: inRichtung internal/service/bildreihe_intern_test.go:53:18: undefined: wahlbildebene 
  nachher internal/msepaket/feldstilreihe_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/msepaket	0.205s 
  nachher internal/service/bildreihe_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/service	0.350s 
  nachher internal/service/bildreihe_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/service	0.241s 
== c4f161959 feat(g08): die Lauschwahl der eigenen Dienste als Einstellung
  vorher internal/app/bindungeneinstellungen_test.go: rc=1 internal/app/einstellungen_test.go:300:10: undefined: domain.NetzLauschen FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app [build failed] FAIL 
  vorher internal/app/einstellungen_test.go: rc=1 internal/app/einstellungen_test.go:300:10: undefined: domain.NetzLauschen FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app [build failed] FAIL 
  vorher internal/domain/einstellungsreihe_test.go: rc=1 internal/domain/freieeinstellungen_test.go:72:9: undefined: domain.NetzLauschen internal/domain/einstellungsreihe_test.go:113:10: undefined: domain.NetzLauschen internal/domain/lauschwahl_test.go:36:27: undefined: domain
  vorher internal/domain/lauschwahl_test.go: rc=1 internal/domain/freieeinstellungen_test.go:72:9: undefined: domain.NetzLauschen internal/domain/einstellungsreihe_test.go:113:10: undefined: domain.NetzLauschen internal/domain/lauschwahl_test.go:36:27: undefined: domain
  vorher internal/domain/wirkzeitpunkt_test.go: rc=1 internal/domain/freieeinstellungen_test.go:72:9: undefined: domain.NetzLauschen internal/domain/einstellungsreihe_test.go:113:10: undefined: domain.NetzLauschen internal/domain/lauschwahl_test.go:36:27: undefined: domain
  nachher internal/app/bindungeneinstellungen_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.549s 
  nachher internal/app/einstellungen_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.457s 
  nachher internal/domain/einstellungsreihe_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain	0.218s 
  nachher internal/domain/lauschwahl_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain	0.126s 
  nachher internal/domain/wirkzeitpunkt_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain	0.128s 
== cce1294b1 feat(g08): die Schnittstellen dieses Rechners zum Wählen
  vorher internal/domain/wirkzeitpunkt_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain	0.239s 
  vorher internal/entdeckung/schnittstellenliste_intern_test.go: rc=1 internal/entdeckung/schnittstellenliste_intern_test.go:56:19: undefined: schnittstellenlisteAus internal/entdeckung/schnittstellenliste_intern_test.go:69:16: undefined: Schnittstelleneintrag internal/entdeckung/schnittst
  nachher internal/domain/wirkzeitpunkt_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/domain	0.127s 
  nachher internal/entdeckung/schnittstellenliste_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/entdeckung	0.218s 
== 7edd2ebf9 feat(g08): ein Dienst lauscht auf jeder gewählten Adresse
  vorher internal/rest/lauschenaufallen_test.go: rc=1 internal/rest/lauschenaufallen_test.go:34:35: undefined: rest.LauschenAufAllen internal/rest/lauschenaufallen_test.go:41:19: undefined: rest.Adressen internal/rest/lauschenaufallen_test.go:101:35: undefined: rest.Lausche
  nachher internal/rest/lauschenaufallen_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest	0.273s 
== e29917dcd feat(g08): die Ansage nennt jede Adresse und jede fehlende
  vorher internal/app/dienstansagemehr_intern_test.go: rc=1 internal/app/dienstansagemehr_intern_test.go:26:15: undefined: Fehlendeadresse internal/app/dienstansagemehr_intern_test.go:48:32: gelesen.adressen undefined (type dienstlauf has no field or method adressen) internal/app
  nachher internal/app/dienstansagemehr_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.552s 
== e2d3ac44e feat(g08): mse-rest und mse-mcp lauschen auf jeder gewählten Adresse
  vorher cmd/mse-mcp/lauschwahl_test.go: rc=1 --- FAIL: TestDerWerkzeugserverLauschtAufJederGenanntenAdresse (0.15s) FAIL FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/cmd/mse-mcp	0.673s 
  vorher cmd/mse-rest/lauschwahl_test.go: rc=1 --- FAIL: TestDerDienstLauschtAufJederGenanntenAdresse (0.14s) --- FAIL: TestOhneAdresseGiltDieEinstellung (0.13s) FAIL 
  vorher internal/app/dienstansagemehr_intern_test.go: rc=1 internal/app/dienstansagemehr_intern_test.go:28:22: undefined: domain.Lauschbefund internal/app/dienstzeugnis_intern_test.go:136:62: cannot use nil as string value in argument to anwendung.Dienstzeugnis internal/app/dien
  vorher internal/app/dienstzeugnis_intern_test.go: rc=1 internal/app/dienstansagemehr_intern_test.go:28:22: undefined: domain.Lauschbefund internal/app/dienstzeugnis_intern_test.go:136:62: cannot use nil as string value in argument to anwendung.Dienstzeugnis internal/app/dien
  vorher internal/app/dienstzeugnismehr_intern_test.go: rc=1 internal/app/dienstansagemehr_intern_test.go:28:22: undefined: domain.Lauschbefund internal/app/dienstzeugnis_intern_test.go:136:62: cannot use nil as string value in argument to anwendung.Dienstzeugnis internal/app/dien
  nachher cmd/mse-mcp/lauschwahl_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/cmd/mse-mcp	0.646s 
  nachher cmd/mse-rest/lauschwahl_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/cmd/mse-rest	0.918s 
  nachher internal/app/dienstansagemehr_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.533s 
  nachher internal/app/dienstzeugnis_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.876s 
  nachher internal/app/dienstzeugnismehr_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.553s 
== e8ac6c6c6 feat(g08): der Dienst aus dem Fenster startet auf der ganzen Lauschwahl
  vorher frontend/src/bausteine/dienstposten.test.tsx: rc=0       Tests  15 passed (15) 
  vorher frontend/src/bausteine/dienstzeugnis.test.tsx: rc=0       Tests  2 passed (2) 
  vorher frontend/src/seiten/einstellungsdoppelung.test.tsx: rc=0       Tests  4 passed (4) 
  vorher internal/app/dienstgleichzeitig_intern_test.go: rc=1 internal/app/dienstgleichzeitig_intern_test.go:45:29: cannot use func(context.Context, domain.Dienstart, []domain.Bindung) (dienstlauf, error) {…} (value of type func("context".Context, domain.Dienstart, []domain.Bindu
  vorher internal/app/dienstmehr_intern_test.go: rc=1 internal/app/dienstgleichzeitig_intern_test.go:45:29: cannot use func(context.Context, domain.Dienstart, []domain.Bindung) (dienstlauf, error) {…} (value of type func("context".Context, domain.Dienstart, []domain.Bindu
  vorher internal/app/dienstposten_intern_test.go: rc=1 internal/app/dienstgleichzeitig_intern_test.go:45:29: cannot use func(context.Context, domain.Dienstart, []domain.Bindung) (dienstlauf, error) {…} (value of type func("context".Context, domain.Dienstart, []domain.Bindu
  vorher internal/app/dienstwerk_intern_test.go: rc=1 internal/app/dienstgleichzeitig_intern_test.go:45:29: cannot use func(context.Context, domain.Dienstart, []domain.Bindung) (dienstlauf, error) {…} (value of type func("context".Context, domain.Dienstart, []domain.Bindu
  vorher internal/app/dienstzeugnisstand_intern_test.go: rc=1 internal/app/dienstgleichzeitig_intern_test.go:45:29: cannot use func(context.Context, domain.Dienstart, []domain.Bindung) (dienstlauf, error) {…} (value of type func("context".Context, domain.Dienstart, []domain.Bindu
  vorher internal/app/schalternachlauf_intern_test.go: rc=1 internal/app/dienstgleichzeitig_intern_test.go:45:29: cannot use func(context.Context, domain.Dienstart, []domain.Bindung) (dienstlauf, error) {…} (value of type func("context".Context, domain.Dienstart, []domain.Bindu
  nachher frontend/src/bausteine/dienstposten.test.tsx: rc=0       Tests  15 passed (15) 
  nachher frontend/src/bausteine/dienstzeugnis.test.tsx: rc=0       Tests  2 passed (2) 
  nachher frontend/src/seiten/einstellungsdoppelung.test.tsx: rc=0       Tests  4 passed (4) 
  nachher internal/app/dienstgleichzeitig_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.692s 
  nachher internal/app/dienstmehr_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.660s 
  nachher internal/app/dienstposten_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.405s 
  nachher internal/app/dienstwerk_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.693s 
  nachher internal/app/dienstzeugnisstand_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.575s 
  nachher internal/app/schalternachlauf_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.515s 
== b055a0ea2 feat(g08): alle Ziele mit einem Ruf einrichten und zurücknehmen
  vorher internal/app/schnittstellenalle_intern_test.go: rc=1 internal/app/schnittstellenalle_intern_test.go:38:25: undefined: SchnittstellenAlleAntwort internal/app/schnittstellenalle_intern_test.go:38:65: undefined: SchnittstellenausgangDTO internal/app/schnittstellenalle_intern_
  nachher internal/app/schnittstellenalle_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.551s 
== 200684fb6 feat(g08): ein Knopf richtet alles ein, und die Schnittstellen sind wählbar
  vorher frontend/src/bausteine/lauschwahl.test.tsx: rc=1  FAIL  src/bausteine/lauschwahl.test.tsx [ src/bausteine/lauschwahl.test.tsx ] 
  vorher frontend/src/bausteine/schnittstellenalle.test.tsx: rc=1      × zeigt vorher jede Datei und Zeile, die hinzukommt, und richtet alles ein 21ms      × nimmt alles zurück und meldet einen Fehler der Antwort 3ms      × zeigt ohne Ziele keinen Block und sperrt das Zurücknehmen
  vorher frontend/src/seiten/technik.test.tsx: rc=1      × bietet keinen Weg an, etwas auszuführen 64ms ⎯⎯⎯⎯⎯⎯⎯ Failed Tests 1 ⎯⎯⎯⎯⎯⎯⎯  FAIL  src/seiten/technik.test.tsx > Der technische Bereich > bietet keinen Weg an, etwas auszuführen 
  vorher internal/app/netzschnittstellen_intern_test.go: rc=1 internal/app/netzschnittstellen_intern_test.go:20:27: NeueApp("").MitNetzschnittstellen undefined (type *App has no field or method MitNetzschnittstellen) internal/app/netzschnittstellen_intern_test.go:28:38: NeueBindung
  vorher internal/rest/lauschenaufallen_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest	0.282s 
  nachher frontend/src/bausteine/lauschwahl.test.tsx: rc=0       Tests  5 passed (5) 
  nachher frontend/src/bausteine/schnittstellenalle.test.tsx: rc=0       Tests  4 passed (4) 
  nachher frontend/src/seiten/technik.test.tsx: rc=0       Tests  7 passed (7) 
  nachher internal/app/netzschnittstellen_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.586s 
  nachher internal/rest/lauschenaufallen_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/rest	0.179s 
== b8cc3f23b feat(g08): der Dienst im Fenster folgt der Lauschwahl und nennt jeden Lauscher
  vorher frontend/src/bausteine/dienstposten.test.tsx: rc=1  FAIL  src/bausteine/dienstposten.test.tsx [ src/bausteine/dienstposten.test.tsx ] 
  vorher frontend/src/bausteine/dienstpostenlauscher.test.tsx: rc=1  FAIL  src/bausteine/dienstpostenlauscher.test.tsx [ src/bausteine/dienstpostenlauscher.test.tsx ] 
  nachher frontend/src/bausteine/dienstposten.test.tsx: rc=0       Tests  15 passed (15) 
  nachher frontend/src/bausteine/dienstpostenlauscher.test.tsx: rc=0       Tests  1 passed (1) 
== 6f19ce4f5 feat(g10): Schreibknöpfe für einen Leser gesperrt und erklärt
  vorher frontend/src/bausteine/Aufklappkasten.test.tsx: rc=1  FAIL  src/bausteine/Aufklappkasten.test.tsx [ src/bausteine/Aufklappkasten.test.tsx ] 
  vorher frontend/src/bausteine/aktionsknopf.test.tsx: rc=1  FAIL  src/bausteine/aktionsknopf.test.tsx [ src/bausteine/aktionsknopf.test.tsx ] 
  vorher frontend/src/bausteine/markennurlesend.test.tsx: rc=1  FAIL  src/bausteine/markennurlesend.test.tsx [ src/bausteine/markennurlesend.test.tsx ] 
  vorher frontend/src/bausteine/nurlesehinweis.test.tsx: rc=1  FAIL  src/bausteine/nurlesehinweis.test.tsx [ src/bausteine/nurlesehinweis.test.tsx ] 
  vorher frontend/src/haken/schreibrecht.test.tsx: rc=1  FAIL  src/haken/schreibrecht.test.tsx [ src/haken/schreibrecht.test.tsx ] 
  vorher frontend/src/schreibknoepfe.test.ts: rc=1      × seiten/Sets.tsx: texte.seiten.sets.neuKopf ist gesperrt (Ein leeres Set anlegen) 3ms      × seiten/Seteinfuhr.tsx: texte.seiten.sets.einlesenTitel ist gesperrt (Ein Set öffnen — es landet im Bestand des Profi
  nachher frontend/src/bausteine/Aufklappkasten.test.tsx: rc=0       Tests  5 passed (5) 
  nachher frontend/src/bausteine/aktionsknopf.test.tsx: rc=0       Tests  14 passed (14) 
  nachher frontend/src/bausteine/markennurlesend.test.tsx: rc=0       Tests  2 passed (2) 
  nachher frontend/src/bausteine/nurlesehinweis.test.tsx: rc=0       Tests  2 passed (2) 
  nachher frontend/src/haken/schreibrecht.test.tsx: rc=0       Tests  5 passed (5) 
  nachher frontend/src/schreibknoepfe.test.ts: rc=0       Tests  17 passed (17) 
== c90b06be6 fix(g10): gesperrte Markenknöpfe sehen aus wie ein gesperrter Knopf
  vorher frontend/src/bausteine/markennurlesend.test.tsx: rc=1      × lässt das Feld lesen und sperrt Setzen und Entfernen 134ms      × öffnet einem Leser die Zeile nicht und einem Mitarbeiter schon 6ms ⎯⎯⎯⎯⎯⎯⎯ Failed Tests 2 ⎯⎯⎯⎯⎯⎯⎯ 
  nachher frontend/src/bausteine/markennurlesend.test.tsx: rc=0       Tests  2 passed (2) 
== f187b00c1 fix(g10): Aufnehmen im Leerzustand und die Freigabe für einen Leser gesperrt
  vorher frontend/src/schreibknoepfe.test.ts: rc=1      × seiten/Sammlungszustaende.tsx: eigene.neuesTitel ist gesperrt (Ein Exemplar aufnehmen — derselbe Weg wie im Kopf) 2ms      × seiten/Ruecksetzung.tsx: worte.freigeben ist gesperrt (Den Neuanfang freigeben — v
  vorher frontend/src/seiten/ruecksetzungsfreigabe.test.tsx: rc=1      × sperrt die Freigabe für einen Leser 18ms ⎯⎯⎯⎯⎯⎯⎯ Failed Tests 1 ⎯⎯⎯⎯⎯⎯⎯  FAIL  src/seiten/ruecksetzungsfreigabe.test.tsx > die Freigabe für den Neuanfang > sperrt die Freigabe für e
  vorher frontend/src/seiten/sammlungladen.test.tsx: rc=1      × behauptet nach einem ersten Lesefehler keine leere Sammlung 21ms ⎯⎯⎯⎯⎯⎯⎯ Failed Tests 1 ⎯⎯⎯⎯⎯⎯⎯  FAIL  src/seiten/sammlungladen.test.tsx > Die Sammlung beim Öffnen > behauptet nach ei
  vorher frontend/src/seiten/sammlungleser.test.tsx: rc=1      × sperrt das Aufnehmen und lässt das Einlesen offen 43ms ⎯⎯⎯⎯⎯⎯⎯ Failed Tests 1 ⎯⎯⎯⎯⎯⎯⎯  FAIL  src/seiten/sammlungleser.test.tsx > Die leere Sammlung für einen Leser > sperrt das Aufne
  nachher frontend/src/schreibknoepfe.test.ts: rc=0       Tests  20 passed (20) 
  nachher frontend/src/seiten/ruecksetzungsfreigabe.test.tsx: rc=0       Tests  3 passed (3) 
  nachher frontend/src/seiten/sammlungladen.test.tsx: rc=0       Tests  4 passed (4) 
  nachher frontend/src/seiten/sammlungleser.test.tsx: rc=0       Tests  1 passed (1) 
== a431ebad3 fix(g10): Einfügen im leeren Set ist für einen Leser gesperrt
  vorher frontend/src/schreibknoepfe.test.ts: rc=1      × seiten/Kartenleer.tsx: texte.seiten.karten.ablage.einfuegen ist gesperrt (Karten in ein leeres Set einfügen) 2ms ⎯⎯⎯⎯⎯⎯⎯ Failed Tests 1 ⎯⎯⎯⎯⎯⎯⎯  FAIL  src/schreibknoepfe.test.ts > Die
  vorher frontend/src/seiten/kartenleerleser.test.tsx: rc=1      × sperrt das Einfügen und erklärt es 125ms ⎯⎯⎯⎯⎯⎯⎯ Failed Tests 1 ⎯⎯⎯⎯⎯⎯⎯  FAIL  src/seiten/kartenleerleser.test.tsx > Das leere Set für einen Leser > sperrt das Einfügen und erklär
  nachher frontend/src/schreibknoepfe.test.ts: rc=0       Tests  21 passed (21) 
  nachher frontend/src/seiten/kartenleerleser.test.tsx: rc=0       Tests  1 passed (1) 
== 9f1b7540c test(g10): der Lesefehlerfall wartet auf die Absage statt auf den Ruf
  vorher frontend/src/haken/schreibrecht.test.tsx: rc=0       Tests  5 passed (5) 
  nachher frontend/src/haken/schreibrecht.test.tsx: rc=0       Tests  5 passed (5) 
== a67f474bd feat(g11): Startseite, Sammlung und Ausgabeauswahl sagen, dass sie laden
  vorher frontend/src/bausteine/ausgabeauswahlladen.test.tsx: rc=1      × sagt, dass sie lädt, und nennt solange keine Kartenzahl 15031ms      × meldet ein Scheitern und lädt erneut 55ms ⎯⎯⎯⎯⎯⎯⎯ Failed Tests 2 ⎯⎯⎯⎯⎯⎯⎯ 
  vorher frontend/src/seiten/sammlungladen.test.tsx: rc=1      × sagt, dass sie lädt, und behauptet solange keine leere Sammlung 15034ms      × bietet nach einem Lesefehler den nächsten Versuch an 68ms      × lädt nach einem Lesefehler der Marken die Marken erneut 28ms 
  vorher frontend/src/seiten/startladen.test.tsx: rc=1      × sagt, dass er lädt, solange die Antwort aussteht 15019ms      × nennt ein Scheitern samt Grund und lädt erneut 39ms      × sagt, dass er lädt, solange die Antwort aussteht 15003ms 
  nachher frontend/src/bausteine/ausgabeauswahlladen.test.tsx: rc=0       Tests  2 passed (2) 
  nachher frontend/src/seiten/sammlungladen.test.tsx: rc=0       Tests  3 passed (3) 
  nachher frontend/src/seiten/startladen.test.tsx: rc=0       Tests  4 passed (4) 
== bca5b2a7e test(g11): die Zustandsprobe misst, ob eine Seite ihre Bindungen zeigt
== 5ff419845 fix(g11): die Zustandsprobe vergleicht auch den Bedienzustand
== e728df085 test(g): drei Zweige der Oberfläche abgedeckt, die G02 und G11 brachten
  vorher frontend/src/seiten/aufmerksamkeit.test.tsx: rc=0       Tests  7 passed (7) 
  vorher frontend/src/seiten/paketleer.test.tsx: rc=0       Tests  2 passed (2) 
  vorher frontend/src/seiten/weiterarbeiten.test.tsx: rc=0       Tests  11 passed (11) 
  nachher frontend/src/seiten/aufmerksamkeit.test.tsx: rc=0       Tests  7 passed (7) 
  nachher frontend/src/seiten/paketleer.test.tsx: rc=0       Tests  2 passed (2) 
  nachher frontend/src/seiten/weiterarbeiten.test.tsx: rc=0       Tests  11 passed (11) 
== d95c75875 fix(g12): die Freigabe für den Neuanfang lässt sich am Fenster ausstellen
  vorher frontend/src/seiten/ruecksetzung.test.tsx: rc=0       Tests  11 passed (11) 
  vorher frontend/src/seiten/ruecksetzungsfreigabe.test.tsx: rc=1      × verlangt für die ganze Anwendung eine Freigabe statt eines Namens 71ms      × meldet eine verweigerte Freigabe am Ort und gibt den Knopf nicht frei 15019ms ⎯⎯⎯⎯⎯⎯⎯ Failed Tests 2 ⎯⎯⎯⎯⎯
  vorher internal/app/ruecksetzungsfreigabe_test.go: rc=1 internal/app/ruecksetzungsfreigabe_test.go:49:24: anwendung.RuecksetzungFreigabeErteilen undefined (type *app.App has no field or method RuecksetzungFreigabeErteilen) internal/app/ruecksetzungsfreigabe_test.go:82:24: anw
  vorher internal/flaechen/neuanfangsfreigabe_test.go: rc=1 internal/flaechen/neuanfangsfreigabe_test.go:79:54: app.RuecksetzungsAnwendungUeber(db, ort).RuecksetzungFreigabeErteilen undefined (type *app.App has no field or method RuecksetzungFreigabeErteilen) FAIL	github.com/Davi
  nachher frontend/src/seiten/ruecksetzung.test.tsx: rc=0       Tests  11 passed (11) 
  nachher frontend/src/seiten/ruecksetzungsfreigabe.test.tsx: rc=0       Tests  2 passed (2) 
  nachher internal/app/ruecksetzungsfreigabe_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.877s 
  nachher internal/flaechen/neuanfangsfreigabe_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/flaechen	15.894s 
== 1331f3df7 test(g12): die Freigabetests in Prettier-Form
  vorher frontend/src/seiten/ruecksetzungsfreigabe.test.tsx: rc=0       Tests  2 passed (2) 
  nachher frontend/src/seiten/ruecksetzungsfreigabe.test.tsx: rc=0       Tests  2 passed (2) 
== 0a26e577b fix(g12): der Neuanfang zählt jedes Profil des Bestands
  vorher internal/app/neuanfangfremde_test.go: rc=1 --- FAIL: TestDerNeuanfangZaehltFremdeProfileMit (0.09s) --- FAIL: TestEineFremdeOffeneRueckkehrSperrtDenNeuanfang (0.08s) FAIL 
  vorher internal/service/ruecksetzungfremde_test.go: rc=1 --- FAIL: TestDerNeuanfangNenntNurDasEigeneProfil (0.00s) FAIL FAIL	github.com/DavidTHeider/MagicSetEditor4Mac/internal/service	0.383s 
  vorher internal/storage/nutzerfehler_intern_test.go: rc=1 internal/storage/nutzerfehler_intern_test.go:70:14: undefined: AlleProfile internal/storage/nutzerfehler_intern_test.go:75:14: undefined: AlleProfile internal/storage/nutzerfehler_intern_test.go:80:14: undefined: AllePro
  vorher internal/storage/ruecksetzung_test.go: rc=1 internal/storage/nutzerfehler_intern_test.go:70:14: undefined: AlleProfile internal/storage/nutzerfehler_intern_test.go:75:14: undefined: AlleProfile internal/storage/nutzerfehler_intern_test.go:80:14: undefined: AllePro
  nachher internal/app/neuanfangfremde_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/app	0.741s 
  nachher internal/service/ruecksetzungfremde_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/service	0.362s 
  nachher internal/storage/nutzerfehler_intern_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage	0.295s 
  nachher internal/storage/ruecksetzung_test.go: rc=0 ok  	github.com/DavidTHeider/MagicSetEditor4Mac/internal/storage	1.161s 
fertig
```
