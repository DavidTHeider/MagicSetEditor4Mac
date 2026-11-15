# 0097 — Was über ein KI-Modell feststeht

**Stand:** angenommen · **Datum:** 2026-09-11 · **Werkstück:** A6 von Phase 5.3
(#1445)

## Worum es geht

Mit A6 bekommt dieses Programm zum ersten Mal einen Begriff von „KI-Modell".
Was daran schwierig ist, ist nicht die Technik, sondern die **Auskunft**: ein
Anbieter nennt Namen, und aus einem Namen folgt nichts. Er sagt nicht, was das
Modell kann, woher es kommt, was wirklich geladen wurde und wo gerechnet wird.

Vier Entscheidungen halten das auseinander.

## 1 Sieben Fähigkeiten und kein „KI-Antwort"-Blob

Das Issue sagt es wörtlich: „Embeddings liefern Vektoren, Reranker
Kandidatenbewertungen, Generierung Vorschläge — **kein generischer
»KI-Antwort«-Blob**."

Also sieben: `textgenerierung`, `strukturiert`, `vision`, `bildgenerierung`,
`bildbearbeitung`, `einbettung`, `neuordnung`.

**Ein gemeinsamer Typ darüber wäre die Stelle, an der ein Vektor als Vorschlag
in ein Kartenfeld gerät.** Irgendwo müsste dann ein `switch` entscheiden, was
gerade zurückkam — und ein `switch` über einen Inhalt, den ein fremder Dienst
bestimmt, ist keine Prüfung.

## 2 Eine Fähigkeit ohne Beleg gilt als ungeprüft

„Unbekannt bleibt unbekannt; **Modelllisten garantieren keine Capability**."

Deshalb trägt jede Fähigkeit ein **Datum** und einen **Beleg** — einen Satz
darüber, *wie* geprüft wurde. Ohne beides zählt sie nicht.

**Und das Datum sagt nicht, dass es noch stimmt.** Ein Anbieter tauscht ein
Modell hinter seinem Namen aus, eine Laufzeit ändert ihre Vorgaben; was am
2026-09-11 ging, geht deshalb nicht auf ewig. Das Datum sagt, wie alt die
Auskunft ist — mehr kann es nicht.

## 3 Die Herkunft ist eine Kette, und eine halbe gibt es nicht

Zwischen dem, was jemand veröffentlicht hat, und dem, was hier läuft, liegen
Schritte: Basis → Adapter → Umwandlung → **ausgeführtes Artefakt**. Jedes Glied
nennt Quelle, **unveränderliche** Revision, Dateihashes und Lizenz.

**Ohne das letzte Glied beschreibt die Kette, was es gibt, und nicht, was
gelaufen ist.** Deshalb ist genau eines davon Pflicht.

**Eine leere Kette ist erlaubt und heißt „unbekannt".** Das Issue verbietet
ausdrücklich den „erratenen HF-Commit aus einem Ollama-Alias": ein Alias nennt
einen Namen, keine Revision. Wer daraus eine ableitet, hat eine
Herkunftsangabe **erfunden** — und die ist schlimmer als keine, weil niemand
ihr ansieht, dass sie geraten war.

**Was eine Kette nicht belegt**, sagt das Issue selbst: „Lizenz/Hash allein
belegt weder sichere Ausführung noch lokale Inferenz." Ein Hash sagt, welche
Bytes geladen wurden. Was sie tun, sagt er nicht.

## 4 Der Ort ist erklärt und nicht gemessen

**„Loopback allein beweist keine lokale Inferenz."** Ein Dienst auf
`127.0.0.1` kann weiterleiten; die Adresse sagt, wohin gefragt wird, nicht, wo
gerechnet wird.

Der Ort steht deshalb als **Angabe** am Modell. Wer ihn falsch setzt, hat
gelogen und sich nicht geirrt — und das ist eine ehrlichere Lage als eine
Messung, die etwas beweist, das sie nicht beweisen kann.

**Die Regel dazu kommt aus 5.2 und ist keine zweite.** `richtlinie.nurlokal`
gibt es längst und sagt genau das. Beim Bauen dieses Werkstücks stand hier
fast eine eigene Einstellung `ki.ort`; sie wäre die dritte zweite Aufzählung
dieses Tages gewesen. Das Issue verlangt ausdrücklich, die Ebenen von 5.2 zu
**konsumieren** — und die Faltung über Gerät, Nutzer und Profil macht `domain`
ohnehin, sodass „restriktivste Regel gilt" beantwortet ist, bevor `service`
läuft.

## Und der Satz, um den herum der Auftrag gebaut ist

> **Generieren verändert die Zielkarte nicht.**

Ein Auftrag friert einen Stand ein — Profil, Zielkarte, Zielrevision,
Schemafassung, Frist, Budgetreservierung — und erzeugt einen **Vorschlag** mit
einer Kennung. Was daraus wird, entscheidet ein **zweiter** Schritt, und der
prüft noch einmal alles, was sich inzwischen geändert haben kann.

**Warum zweimal geprüft wird**: zwischen Vorschlag und Übernahme liegt Zeit.
Die Karte kann sich geändert haben, die Frist kann abgelaufen sein, das Paket
kann ein anderes sein. Wer nur beim Vorschlagen prüfte, schriebe später gegen
einen Stand, den niemand angesehen hat — dieselbe Überlegung wie §7, Regel 2.

**Und die Kennung ist konsumierbar**: zweimal übernehmen hieße, denselben
Vorschlag auf einen Stand zu legen, der ihn schon trägt.

## Ein leeres Schema ist kein Freibrief

Der schärfste Satz des Issues zum Schema: „Fehlendes/unlesbares Paket erzeugt
für KI **keinen beliebigen Feldfreibrief**."

Ein Schema, das nichts weiß, weiß **nicht**, dass alles erlaubt ist — es weiß
nichts. Wer daraus „dann halt alles" machte, hätte die Prüfung genau dort
abgeschaltet, wo sie am nötigsten ist. `KISchema.Pruefe` weist deshalb jedes
Feld ab, das nicht ausdrücklich dasteht.

**Und ein leerer Vorschlag ist kein Befund**: ein Modell darf sich verweigern.
Das Issue nennt „Refusal" neben den Fehlern — eine Antwort ohne Vorschlag ist
eine Antwort, und sie führt zu keiner Kartenänderung.

## Was heute noch fehlt

**Ein Anbieter.** Der Katalog ist leer, und die Oberfläche sagt es: „Es steht
kein KI-Modell fest." A7
(#1446)
bringt Ollama als lokale Referenz; dann bekommt auch
`service.KIAuftragswerk` seinen ersten Aufrufer.

**Das ist bewusst so geschnitten** und kein Zustand, den jemand übersehen
hat: das Issue verlangt „den **separat abnehmbaren** Modellvertrag zuerst".
Was dieses Werkstück liefert, ist die Antwort auf „was weiß dieses Programm
über KI-Modelle" — und die lautet heute ehrlich: nichts, und hier ist die
Form, in der es etwas wissen könnte.
