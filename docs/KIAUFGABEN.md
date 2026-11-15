# KI an Nutzeraufgaben — was es gibt, woran es zu sehen ist, und was fehlt

**Stand:** 2026-10-04 · AI-03
(#1744).

AI-03 verlangt, dass die KI an **Karten-, Deck-, Vorlagen- und
Regelaufgaben** anschließt — mit nachvollziehbaren Quellen und kontrollierter
Übernahme. **Das meiste davon steht seit Phase 5.3**; diese Datei ordnet es den
Kriterien zu, mit der Stelle, an der es geschieht, und nennt, was fehlt.

## AK-01 — vier Aufgaben, jede mit einem Weg ohne KI

| Aufgabe | Mit KI | Ohne KI | Wo |
| --- | --- | --- | --- |
| **Semantische Suche** über eigene und freigegebene Inhalte | Recherche mit Einbettung (A16) | **Volltextsuche** — ohne Einbettungsmodell fällt die Recherche sichtbar auf sie zurück (`leere: ohne-modell`) | `internal/app/recherche.go`, `internal/service/recherchedienst.go`; Güte in [`RETRIEVALMESSUNG.md`](RETRIEVALMESSUNG.md) |
| **Dublettenkandidaten** | „Ähnliche Karten" über denselben Index | **Die Sammlung erkennt ein vorhandenes Exemplar** beim Anlegen (`domain.Dublette`) | `App.AehnlicheKarten` in `internal/app/recherche.go`; `internal/service/sammlungspflege.go` |
| **Strukturierte Feldvorschläge** | Vorschlag gegen das Schema der Paketfelder (A7, A11); **am Feld** für Name und Flavortext mit Vorrang fürs Gerätemodell (AI-05) | **Tippen**, und der Abgleich mit Scryfall | `internal/app/kivorschlag*.go`, `internal/app/kivorrangnaht.go`, `frontend/src/bausteine/KIFeldvorschlag.tsx` |
| **Kontextuelle Erläuterung** | die Einschätzung einer Karte mit Belegen (A13.2a) | **Die Regelauskunft** der Regelmaschine — sie ist die verbindliche | `internal/app/kieinschaetzungsbezug.go`, `frontend/src/bausteine/KIEinschaetzung.tsx` |

**Priorisiert** nach dem, was ein Mensch beim Bauen eines Sets am häufigsten
tut: Feldvorschläge vor Suche vor Erläuterung vor Dubletten.

## AK-02 — Rechte vor und nach dem Abruf

- **Vor dem Abruf**: private Korpora werden nur mit Freigabe durchsucht; ein
  nicht freigegebener nennt sich als Einschränkung
  (`internal/service/rechercheantwort.go`, „ist nicht freigegeben").
- **Nach dem Abruf**: Seite und Fortsetzung binden Anfrage, Profil, Rechte und
  Generation (`internal/service/recherchedienst.go`, Kopf); ein Kontextpaket
  wird vor der Verwendung erneut geprüft — Ziel, Profil und Ablauf
  (`internal/service/kontextpruefung.go`). **Gelöschte Quellen nicht**: der
  Index trägt im Betrieb nur öffentliche Quellen, und ein Korpus eigener
  Inhalte bringt diese Prüfung mit, wenn er kommt (#2436).
- **Der Index bindet das Modell**: ein Einbettungsvertrag nennt Modell und
  Dimension, und ein anderes Modell heißt Neubau
  (`internal/domain/kieinbettung.go`).

## AK-03 — eigene Daten, fremde Fakten, Vermutung

- **Belege stehen neben den Feldern, nicht als ihr Beweis**: was das Modell
  sah, steht unter `belege`; was es vorschlägt, ist eine Vermutung und wird
  als `ki` verbucht (`KIVorschlagDTO`, `internal/app/kivorschlag.go`).
- **Vorschau und Unterschied vor der Übernahme**: jeder vorgeschlagene Wert
  trägt `bisher`; geschrieben wird erst auf ausdrücklichen Druck, in einem
  Zug, mit Herkunft (A11, `KIVorschlagUebernehmen`).

## AK-04 — keine verbindliche Entscheidung, und die Güte am Testsatz

- **Die KI entscheidet keinen Handel und ist nicht die Regelmaschine**:
  `internal/service/handel*.go`, `tausch*.go`, `partie*.go` und
  `internal/domain/regel*.go` rufen keinen KI-Dienst (nachgezählt am
  2026-10-04 über die Importe).
- **Die Güte am fachlichen Testsatz** misst
  [`belege/fmguete.md`](belege/fmguete.md): acht selbst geschriebene Karten,
  Name und Flavortext, Deutsch und Englisch, am Gerätemodell (AI-05) — 32
  Fragen, keine leer, keine abgelehnt, im Mittel unter einer Sekunde. **Ob ein
  Vorschlag passt**, steht je Zeile, bewertet von Claude und nicht vom
  Auftraggeber. **Der erste Lauf fand Fehlvorschläge**: sieben von sechzehn
  Flavortexten wiederholten den Regeltext, fünf Namen waren Typ,
  Regelbeschreibung oder falsche Sprache — 9 passten, 11 teils, 12 nicht.
  Mit geschärften Sätzen am Formular (Eigenname; Stimmung statt Regel, in
  acht Sprachen) **20, 9 und 3**. Was bleibt: zwei Namen in der falschen
  Sprache, einer, der nur der Typ ist.
- **Und gegen zwei kleine Ollama-Modelle** (AI-02,
  [`belege/modellvergleich.md`](belege/modellvergleich.md)): dieselben 32
  Fragen in zwei Runden, über die Klienten der Anwendung. Formtreu sind alle
  drei vollständig; passend nach derselben Bewertung **Gerätemodell 18,
  gemma3:1b 9, qwen3:0.6b 0** von 32 — qwen3 gibt den Kartentyp als Namen
  und beschreibt im Flavortext den Auftrag, statt ihn zu erfüllen. Der Lauf
  fand dabei, dass qwen3 ohne `"think": false` vor jeder Antwort denkt — bei
  einer Frage über drei Minuten lang; behoben.

## Was fehlt

- **Eine Ansicht „Dubletten im Set"** — heute gibt es ähnliche Karten zu
  **einer** Karte, nicht die Liste der Paare. Ein eigenes Issue, sobald es
  jemand braucht; die Grundlage (Index, Ähnlichkeit) steht.

