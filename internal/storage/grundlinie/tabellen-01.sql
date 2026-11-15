-- Erzeugt aus den Wanderungen bis 0097_faktorhaushalt.sql — nicht von Hand ändern.
-- Eingefroren am 2026-10-12 (Vermerk 0189) — eine Änderung am Schema ist eine
-- neue Wanderung unter migrationen/, nie eine Änderung hier.

CREATE TABLE bestand (
  schluessel TEXT PRIMARY KEY,
  wert       TEXT NOT NULL,

  -- Zeitstempel tragen genau drei Nachkommastellen und enden auf Z.
  --
  -- Die Form ist hier festgeschrieben, weil sie sonst auseinanderläuft: SQLite
  -- schreibt mit strftime('%f') drei Stellen, Go mit RFC3339Nano bis zu neun.
  -- Eine Anwendung, die neun schreibt, und eine Migration, die drei einsetzt,
  -- laufen jahrelang nebeneinander her — bis eine Prüfung dazukommt, die eine
  -- der beiden Formen verlangt, und alles Vorhandene abweist.
  CHECK (
    schluessel <> 'angelegt_am'
    OR wert GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9].[0-9][0-9][0-9]Z'
  )
) STRICT;

CREATE TABLE packages (
  id       INTEGER PRIMARY KEY AUTOINCREMENT,
  art      TEXT NOT NULL,
  name     TEXT NOT NULL,
  version  TEXT NOT NULL DEFAULT '',
  pfad     TEXT NOT NULL,
  herkunft TEXT NOT NULL DEFAULT '', pruefsumme TEXT NOT NULL DEFAULT '', gruppe TEXT NOT NULL DEFAULT '', rang INTEGER NOT NULL DEFAULT 1000000, kurzname TEXT NOT NULL DEFAULT '', abdruck TEXT NOT NULL DEFAULT '',

  UNIQUE (art, name)
) STRICT;

CREATE TABLE package_deps (
  package_id  INTEGER NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
  braucht     TEXT NOT NULL,
  version     TEXT NOT NULL DEFAULT '',
  reihenfolge INTEGER NOT NULL,

  PRIMARY KEY (package_id, reihenfolge)
) STRICT;

CREATE TABLE sets (
  id       INTEGER PRIMARY KEY AUTOINCREMENT,
  name     TEXT NOT NULL,
  spiel    TEXT NOT NULL,
  stil     TEXT NOT NULL DEFAULT '',
  herkunft TEXT NOT NULL DEFAULT '',
  geoeffnet TEXT NOT NULL
, mse_fassung   TEXT NOT NULL DEFAULT '', spielfassung  TEXT NOT NULL DEFAULT '', stilfassung   TEXT NOT NULL DEFAULT '', versionskontrolle TEXT NOT NULL DEFAULT '', apprentice_code   TEXT NOT NULL DEFAULT '', profil_id INTEGER NOT NULL DEFAULT 0) STRICT;

CREATE TABLE cards (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  set_id      INTEGER NOT NULL REFERENCES sets(id) ON DELETE CASCADE,
  reihenfolge INTEGER NOT NULL,

  UNIQUE (set_id, reihenfolge)
) STRICT;

CREATE TABLE card_values (
  card_id     INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
  feld        TEXT NOT NULL,
  wert        TEXT NOT NULL,
  reihenfolge INTEGER NOT NULL,

  PRIMARY KEY (card_id, feld)
) STRICT;

CREATE TABLE set_values (
  set_id      INTEGER NOT NULL REFERENCES sets(id) ON DELETE CASCADE,
  feld        TEXT NOT NULL,
  wert        TEXT NOT NULL,
  reihenfolge INTEGER NOT NULL,

  PRIMARY KEY (set_id, feld)
) STRICT;

CREATE TABLE styling_values (
  set_id      INTEGER NOT NULL REFERENCES sets(id) ON DELETE CASCADE,
  feld        TEXT NOT NULL,
  wert        TEXT NOT NULL,
  reihenfolge INTEGER NOT NULL,

  PRIMARY KEY (set_id, feld)
) STRICT;

CREATE TABLE raw_blocks (
  set_id      INTEGER NOT NULL REFERENCES sets(id) ON DELETE CASCADE,
  datei       TEXT NOT NULL,
  zeile       INTEGER NOT NULL,
  inhalt      TEXT NOT NULL,
  reihenfolge INTEGER NOT NULL,

  PRIMARY KEY (set_id, reihenfolge)
) STRICT;

CREATE TABLE history (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  set_id    INTEGER NOT NULL REFERENCES sets(id) ON DELETE CASCADE,
  zeitpunkt TEXT NOT NULL,
  handlung  TEXT NOT NULL,
  rueckweg  TEXT NOT NULL
, hinweg TEXT NOT NULL DEFAULT '', karte_id INTEGER NOT NULL DEFAULT 0, zurueckgenommen INTEGER NOT NULL DEFAULT 0, urheber TEXT NOT NULL DEFAULT 'fenster', gruppe TEXT NOT NULL DEFAULT '', herkunft TEXT NOT NULL DEFAULT '') STRICT;

CREATE VIRTUAL TABLE card_suche USING fts5(feld, wert, card_id UNINDEXED);

CREATE TABLE set_quellen (
  set_id INTEGER PRIMARY KEY REFERENCES sets(id) ON DELETE CASCADE,
  datei  TEXT NOT NULL,
  inhalt BLOB NOT NULL
) STRICT;

CREATE TABLE wert_herkunft (
  card_id  INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
  feld     TEXT NOT NULL,
  herkunft TEXT NOT NULL,

  PRIMARY KEY (card_id, feld)
) STRICT;

CREATE TABLE set_keywords (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  set_id      INTEGER NOT NULL REFERENCES sets(id) ON DELETE CASCADE,
  reihenfolge INTEGER NOT NULL,

  wort        TEXT NOT NULL DEFAULT '',
  muster      TEXT NOT NULL DEFAULT '',
  erinnerung  TEXT NOT NULL DEFAULT '',
  regeln      TEXT NOT NULL DEFAULT '',
  modus       TEXT NOT NULL DEFAULT '',

  UNIQUE (set_id, reihenfolge)
) STRICT;

CREATE TABLE set_pack_types (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  set_id      INTEGER NOT NULL REFERENCES sets(id) ON DELETE CASCADE,
  reihenfolge INTEGER NOT NULL,

  name        TEXT NOT NULL DEFAULT '',
  auswahl     TEXT NOT NULL DEFAULT '',
  filter      TEXT NOT NULL DEFAULT '',

  -- **Als INTEGER und nicht als BOOLEAN**: SQLite kennt keinen, und `STRICT`
  -- weist einen erfundenen Typ ab.
  waehlbar    INTEGER NOT NULL DEFAULT 1,

  UNIQUE (set_id, reihenfolge)
) STRICT;

CREATE TABLE set_pack_items (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  pack_id     INTEGER NOT NULL REFERENCES set_pack_types(id) ON DELETE CASCADE,
  reihenfolge INTEGER NOT NULL,

  name        TEXT NOT NULL DEFAULT '',
  anzahl      INTEGER NOT NULL DEFAULT 1,

  -- **Als REAL**: `weight` ist bei MSE ein `double` (`src/data/pack.hpp`,
  -- Zeile 75, Fassung f9d9356 vom 2021-07-06), und eine Gewichtung von 0,5
  -- ist zulässig.
  gewicht     REAL NOT NULL DEFAULT 1.0,

  UNIQUE (pack_id, reihenfolge)
) STRICT;

CREATE TABLE karten_verweis (
  card_id  INTEGER PRIMARY KEY REFERENCES cards(id) ON DELETE CASCADE,

  -- Die Scryfall-Kennung des Drucks, kanonisch klein — **derselbe Bezug wie
  -- `sammlung.printing`**, damit Entwurf und Besitz nebeneinanderliegen.
  printing TEXT NOT NULL,

  -- Der letzte gute Stand der Anzeige.
  name     TEXT NOT NULL,
  setcode  TEXT NOT NULL DEFAULT '',
  nummer   TEXT NOT NULL DEFAULT '',

  -- Die Kartenseite bei Scryfall — eine Adresse, keine Bezeichnung.
  adresse  TEXT NOT NULL DEFAULT '',

  -- Unix-Sekunden, wie `verlauf.zeitpunkt` und `preisstaende.erhoben`.
  gesetzt  INTEGER NOT NULL
, druck_id INTEGER REFERENCES drucke(id)) STRICT;

CREATE TABLE karten_marken (
  card_id INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,

  -- Die Marke in Normalform: getrimmt, innerer Leerraum gefaltet,
  -- kleingeschrieben.
  marke   TEXT NOT NULL,

  PRIMARY KEY (card_id, marke)
) STRICT;

CREATE TABLE decks (
  id      INTEGER PRIMARY KEY AUTOINCREMENT,

  -- Die Fassung steigt mit jedem Schreiben — die Bedingung einer Änderung
  -- (§7, Regel 2), wie an `sammlung`.
  fassung INTEGER NOT NULL DEFAULT 1,

  name    TEXT    NOT NULL,

  -- Der Name des Spiels, wie `sets.spiel` ihn führt — **ohne** `.mse-game`.
  --
  -- **Ohne Spiel gäbe es kein Format**: ein Format ist immer das eines
  -- Spiels.
  spiel   TEXT    NOT NULL,

  notiz   TEXT    NOT NULL DEFAULT '',

  angelegt TEXT   NOT NULL DEFAULT (datetime('now')),
  geaendert TEXT  NOT NULL DEFAULT (datetime('now'))
, format_id INTEGER REFERENCES formate(id) ON DELETE SET NULL, profil_id INTEGER NOT NULL DEFAULT 0) STRICT;

CREATE TABLE formate (
  id      INTEGER PRIMARY KEY AUTOINCREMENT,

  -- Die Fassung steigt mit jedem Schreiben (§7, Regel 2), wie an `decks`.
  fassung INTEGER NOT NULL DEFAULT 1,

  name    TEXT    NOT NULL,

  -- Ein Format ist immer das **eines** Spiels.
  spiel   TEXT    NOT NULL,

  -- **Null heißt: keine Grenze.** Für `hoechstzahl_gleich` heißt es: beliebig
  -- oft; für `sideboardgroesse`: kein Sideboard.
  mindestgroesse     INTEGER NOT NULL DEFAULT 0,
  hoechstgroesse     INTEGER NOT NULL DEFAULT 0,
  sideboardgroesse   INTEGER NOT NULL DEFAULT 0,
  hoechstzahl_gleich INTEGER NOT NULL DEFAULT 0,

  -- Der Schlüssel in Scryfalls `legalities` — **leer heißt: unbekannt**, und
  -- dann bleibt jede Karte `ungeprüft`. Das ist die Brücke aus dem Vermerk.
  scryfallname TEXT NOT NULL DEFAULT '',

  angelegt  TEXT NOT NULL DEFAULT (datetime('now')),
  geaendert TEXT NOT NULL DEFAULT (datetime('now')), sideboard_offen INTEGER NOT NULL DEFAULT 0, profil_id INTEGER NOT NULL DEFAULT 0,

  CHECK (mindestgroesse >= 0),
  CHECK (hoechstgroesse >= 0),
  CHECK (sideboardgroesse >= 0),
  CHECK (hoechstzahl_gleich >= 0),

  -- **Null ist keine Grenze** und darf deshalb kein Widerspruch sein.
  CHECK (hoechstgroesse = 0 OR hoechstgroesse >= mindestgroesse)
) STRICT;

CREATE TABLE turniere (
  id      INTEGER PRIMARY KEY AUTOINCREMENT,
  fassung INTEGER NOT NULL DEFAULT 1,

  name  TEXT NOT NULL,
  spiel TEXT NOT NULL,

  -- **Ein Match aus einem Spiel** — dann sagen Spielquote und
  -- Gegnerspielquote nichts (MTR §3.1). Gerechnet werden sie trotzdem;
  -- dieses Feld entscheidet die Anzeige.
  einzelspiel INTEGER NOT NULL DEFAULT 0,

  notiz TEXT NOT NULL DEFAULT '',

  angelegt  TEXT NOT NULL DEFAULT (datetime('now')),
  geaendert TEXT NOT NULL DEFAULT (datetime('now')), profil_id INTEGER NOT NULL DEFAULT 0, format_id INTEGER REFERENCES formate(id) ON DELETE SET NULL, abgeschlossen TEXT,

  CHECK (einzelspiel IN (0, 1))
) STRICT;

CREATE TABLE turnierteilnehmer (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  turnier_id INTEGER NOT NULL REFERENCES turniere(id) ON DELETE CASCADE,

  name TEXT NOT NULL,

  -- **Die Herkunft der Deckliste, nicht ihre Quelle** (Vermerk 0062,
  -- Entscheidung 3). `ON DELETE SET NULL`: wer sein Deck löscht, verliert
  -- nicht seine eingereichte Liste.
  deck_id INTEGER REFERENCES decks(id) ON DELETE SET NULL,

  -- **Ausgestiegen heißt: wird nicht mehr gepaart, bleibt aber gezählt.**
  -- Appendix C — seine Matchquote geht weiter in die Gegnerquote jedes
  -- Gegners ein, den er hatte.
  ausgestiegen INTEGER NOT NULL DEFAULT 0,

  -- Die Anmeldereihenfolge — sie entscheidet bei völligem Gleichstand.
  stelle INTEGER NOT NULL,

  CHECK (ausgestiegen IN (0, 1)),
  CHECK (stelle > 0)
) STRICT;

CREATE TABLE turnierdeckplaetze (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  teilnehmer_id INTEGER NOT NULL REFERENCES turnierteilnehmer(id) ON DELETE CASCADE,

  -- **Mindestens eines von beiden ist gefüllt.** Anders als in `deckplaetze`
  -- schließen sie einander **nicht** aus: ein echter Druck hat beides — eine
  -- Kennung und einen Namen —, und der Name ist es, den ein Mensch liest.
  bezeichnung TEXT NOT NULL DEFAULT '',
  printing    TEXT NOT NULL DEFAULT '',

  menge   INTEGER NOT NULL,
  bereich TEXT    NOT NULL,
  stelle  INTEGER NOT NULL, druck_id INTEGER REFERENCES drucke(id),

  CHECK (menge > 0),
  CHECK (bereich IN ('hauptdeck', 'sideboard')),
  CHECK (bezeichnung <> '' OR printing <> '')
) STRICT;

CREATE TABLE turnierrunden (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  turnier_id INTEGER NOT NULL REFERENCES turniere(id) ON DELETE CASCADE,

  nummer INTEGER NOT NULL,

  CHECK (nummer > 0)
) STRICT;

CREATE TABLE turnierpaarungen (
  id       INTEGER PRIMARY KEY AUTOINCREMENT,
  runde_id INTEGER NOT NULL REFERENCES turnierrunden(id) ON DELETE CASCADE,

  a INTEGER NOT NULL REFERENCES turnierteilnehmer(id) ON DELETE CASCADE,

  -- **NULL heißt Bye** — eine Paarung ohne Gegner, nicht eine fehlende
  -- Paarung (Vermerk 0062, Entscheidung 8).
  b INTEGER REFERENCES turnierteilnehmer(id) ON DELETE CASCADE,

  -- **NULL heißt: noch nicht gespielt.** Eine offene Paarung zählt nicht.
  spiele_a      INTEGER,
  spiele_b      INTEGER,
  unentschieden INTEGER,

  stelle INTEGER NOT NULL,

  -- **Entweder ganz oder gar nicht** — ein halbes Ergebnis gibt es nicht.
  CHECK ((spiele_a IS NULL) = (spiele_b IS NULL)),
  CHECK ((spiele_a IS NULL) = (unentschieden IS NULL)),
  CHECK (spiele_a IS NULL OR (spiele_a >= 0 AND spiele_b >= 0 AND unentschieden >= 0)),

  -- **Ein Bye hat kein Ergebnis** — er *ist* eines (2–0, Appendix C).
  CHECK (b IS NOT NULL OR spiele_a IS NULL),

  -- **Niemand spielt gegen sich selbst.**
  CHECK (b IS NULL OR a <> b)
) STRICT;

CREATE TABLE nutzer (
  id       INTEGER PRIMARY KEY AUTOINCREMENT,
  name     TEXT    NOT NULL,

  -- bild ist der Dateiname im Bildverzeichnis — leer heißt: keines.
  --
  -- **Ein Name und kein Pfad**: ein Pfad zeigte nach der Weitergabe des
  -- Bestands auf Verzeichnisse eines fremden Rechners.
  bild     TEXT    NOT NULL DEFAULT '',

  angelegt TEXT    NOT NULL DEFAULT (datetime('now')), email      TEXT NOT NULL DEFAULT '', zweitemail TEXT NOT NULL DEFAULT '', telefon    TEXT NOT NULL DEFAULT '', kennwort TEXT NOT NULL DEFAULT '', anonym INTEGER NOT NULL DEFAULT 0, anrede        TEXT NOT NULL DEFAULT '', namenszusatz  TEXT NOT NULL DEFAULT '', vorname       TEXT NOT NULL DEFAULT '', weiterenamen  TEXT NOT NULL DEFAULT '', nachname      TEXT NOT NULL DEFAULT '', slogan   TEXT NOT NULL DEFAULT '', kurzebio TEXT NOT NULL DEFAULT '', langebio TEXT NOT NULL DEFAULT '', fassung INTEGER NOT NULL DEFAULT 1,

  -- **Der Name ist eindeutig**, denn er ist das Einzige, woran sich zwei
  -- Nutzer in der Profilwahl unterscheiden lassen.
  UNIQUE (name),

  -- **Und er ist nicht leer und trägt keine Ränder.**
  --
  -- Geprüft wird im Dienst (§2); dies ist die zweite Reihe. Sie steht hier,
  -- weil ein Bestand auch von Hand oder von einer früheren Fassung befüllt
  -- werden kann — und weil „David" und „David " in der Profilwahl zwei Zeilen
  -- mit demselben sichtbaren Namen wären.
  CHECK (trim(name) <> '' AND name = trim(name))
) STRICT;
