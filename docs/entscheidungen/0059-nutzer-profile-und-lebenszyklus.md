# 0059 — Zwei weitere Phasen: Nutzer und Profile, und der Lebenszyklus

**Datum:** 2026-08-31
**Zustand:** angenommen

## Der Anlass

Zwei Anforderungen, die keine der bestehenden Phasen trägt, und ein Fehlerbild,
das eine dritte Lücke sichtbar gemacht hat.

## Phase 4.12 — Nutzer und Profile (#1351)

**Ein oder mehrere Nutzer, jedem ein oder mehrere Profile**, und jeder
bearbeitet nur seine eigenen Daten. Dazu eine Anmeldung zwischen Profilen mit
zweitem Faktor, eine verschlüsselbare Ablage und eine Beobachtbarkeit je
Nutzer.

**Warum eine eigene Phase und nicht ein Werkstück von 4.10.** Weil sie das
Datenmodell anfasst — jede Tabelle bekommt eine Zugehörigkeit, und jede
Abfrage eine Bedingung. Das ist keine Oberflächenarbeit, die man in einer
Bedienphase nebenbei erledigt; §13 sagt dazu: „eine Phase wird nicht
stillschweigend größer."

**Drei Dinge, die dabei nicht behauptet werden dürfen.** Sie stehen im Issue
und hier, weil sie leicht zu übersehen und teuer zu korrigieren sind:

- **RLS gibt es in SQLite nicht.** PostgreSQL kennt `ROW LEVEL SECURITY`,
  SQLite nicht. Was durchsetzbar ist, ist eine Zugehörigkeitsspalte, Sichten,
  die sie klemmen, und ein Wächter über die Abfragen. Das ist der ehrliche
  Ersatz — ihn „RLS" zu nennen wäre eine Aussage über eine Fähigkeit, die die
  Ablage nicht hat.
- **Ein zweiter Faktor schützt hier gegen jemanden am selben Rechner**, nicht
  gegen einen Angreifer im Netz: die Anwendung läuft lokal und ohne Server.
  Wer das nicht sagt, verspricht mit der Anmeldemaske mehr, als sie hält.
- **Verschlüsselung und Trennung sind zwei Fragen.** Die Ablage soll sich
  verschlüsseln lassen, auch wenn es nur einen Nutzer gibt. Und die Wahl des
  Mittels ist offen: SQLCipher brächte eine C-Abhängigkeit in ein Projekt, das
  mit `modernc.org/sqlite` bisher **reines Go** ist.

**Newsletter und Support kommen nicht.** Die Stammdaten werden so geführt, dass
es später möglich wäre — mehr nicht. §10: auch eine Ablehnung braucht Deckung,
und eine Zusage erst recht.

## Phase 4.13 — Der Lebenszyklus (#1352)

**Der Anlass war ein Dialog**, der einen rohen Go-Fehler samt einem Pfad aus
dem Arbeitsbaum zeigte. Der Fall selbst war ein Entwicklungsartefakt — zwei
`wails dev`-Instanzen stritten um dasselbe Bindungsverzeichnis, und ein
ausgeliefertes `.app` fasst diesen Pfad nie an.

**Die Klasse bleibt.** Eine Meldung aus dem Inneren in einem nativen Modal ist
keine Fehlerbehandlung, und §6 sagt es seit Anfang: „Fehler erscheinen am
verursachenden Feld, nicht als anonymer Hinweis." Die Phase fragt danach über
den ganzen Lebenszyklus: Start, Platz, Aktualisierung, Protokoll, Abbruch.

**Warum am Ende und nicht sofort.** Weil sie über den **fertigen** Stand fragen
muss. Eine Lebenszyklusphase vor 4.12 zählte Meldungswege, die 4.12 danach
verändert.

## Und ein Zusatz zu Phase 4.11

**Die Anti-Korruptionsschicht gilt auch dort.** §16 verlangt sie für fremde
**Dienste** — Transport, fremdes DTO, Zuordnung an einer Stelle. Für Phase 4.11
heißt das dasselbe eine Ebene weiter: was aus einem **MSE2-Paket** kommt, wird
in `msepaket` gelesen und dort auf die Domäne abgebildet; eine Struktur aus
einer fremden Datei darf nicht in `service` durchschlagen. Dieselbe Regel, die
`internal/mtgjson` seit Phase 9 einhält.

## Die Reihenfolge danach

4.11 → 4.6 → 4.7 → 4.8 → 4.9 → 4.10 → 4.12 → 4.13.

**4.12 nach 4.10**, weil die Profilwahl eine Oberfläche braucht, die es dann
gibt. **4.13 zuletzt**, aus dem Grund darüber.

## Was verworfen wurde

**Die Nutzerverwaltung in 4.10 hineinnehmen.** Der Name der Phase sagt dann
nicht mehr, was in ihr geschieht — und der Umfang wäre doppelt so groß wie das,
wofür sie geschnitten wurde.

**Den Lebenszyklus als Sammelposten führen.** Eine Arbeit ohne Phase hat keinen
Abschluss: keinen Zweitblick, keinen Nachlauf, keine Fortschreibung. Genau
dieser Zustand hat die sechs Paritätsissues von Phase 4.11 monatelang ohne Ort
liegen lassen ([0058](0058-eine-ausserplanmaessige-paritaetsphase.md)).
