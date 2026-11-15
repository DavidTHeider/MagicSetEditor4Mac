-- Erzeugt aus den Wanderungen bis 0097_faktorhaushalt.sql — nicht von Hand ändern.
-- Eingefroren am 2026-10-12 (Vermerk 0189) — eine Änderung am Schema ist eine
-- neue Wanderung unter migrationen/, nie eine Änderung hier.

CREATE TABLE profile (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  nutzer_id INTEGER NOT NULL REFERENCES nutzer (id) ON DELETE CASCADE,
  name      TEXT    NOT NULL,
  angelegt  TEXT    NOT NULL DEFAULT (datetime('now')), fassung INTEGER NOT NULL DEFAULT 1, subslogan          TEXT NOT NULL DEFAULT '', kurzebeschreibung  TEXT NOT NULL DEFAULT '', langebeschreibung  TEXT NOT NULL DEFAULT '', badge TEXT NOT NULL DEFAULT '', technisch INTEGER NOT NULL DEFAULT 0, ablauf    TEXT    NOT NULL DEFAULT '', sichtbarkeit TEXT NOT NULL DEFAULT 'privat',

  -- **Eindeutig je Nutzer, nicht insgesamt.** Zwei Personen dürfen beide ein
  -- Profil „Turnier" führen; dieselbe Person nicht zweimal.
  UNIQUE (nutzer_id, name),

  CHECK (trim(name) <> '' AND name = trim(name))
) STRICT;

CREATE TABLE "settings" (
  schluessel TEXT    NOT NULL,
  wert       TEXT    NOT NULL,
  profil_id  INTEGER NOT NULL,

  PRIMARY KEY (profil_id, schluessel)
) STRICT;

CREATE TABLE "export_options" (
  vorlage   TEXT    NOT NULL,
  feld      TEXT    NOT NULL,
  wert      TEXT    NOT NULL,
  profil_id INTEGER NOT NULL,

  PRIMARY KEY (profil_id, vorlage, feld)
) STRICT;

CREATE TABLE "spiel_listenspalten" (
  spiel     TEXT    NOT NULL,
  feld      TEXT    NOT NULL,
  sichtbar  INTEGER NOT NULL,
  stelle    INTEGER NOT NULL,
  profil_id INTEGER NOT NULL,

  PRIMARY KEY (profil_id, spiel, feld)
) STRICT;

CREATE TABLE "spiel_sortierung" (
  spiel       TEXT    NOT NULL,
  feld        TEXT    NOT NULL,
  aufsteigend INTEGER NOT NULL,
  profil_id   INTEGER NOT NULL,

  PRIMARY KEY (profil_id, spiel)
) STRICT;

CREATE TABLE "spiel_autoersetzungen" (
  spiel       TEXT    NOT NULL,
  muster      TEXT    NOT NULL,
  ersatz      TEXT    NOT NULL,
  ganzes_wort INTEGER NOT NULL DEFAULT 1,
  aktiv       INTEGER NOT NULL DEFAULT 1,
  eigen       INTEGER NOT NULL DEFAULT 1,
  profil_id   INTEGER NOT NULL,

  PRIMARY KEY (profil_id, spiel, muster)
) STRICT;

CREATE TABLE "spiel_autoersetzung_an" (
  spiel     TEXT    NOT NULL,
  an        INTEGER NOT NULL,
  profil_id INTEGER NOT NULL,

  PRIMARY KEY (profil_id, spiel)
) STRICT;

CREATE TABLE anbieterkonten (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  nutzer_id INTEGER NOT NULL REFERENCES nutzer (id) ON DELETE CASCADE,

  anbieter TEXT NOT NULL,
  subjekt  TEXT NOT NULL,

  -- **Nur zum Anzeigen.** Er ändert sich, wenn jemand ihn ändert.
  anzeigename TEXT NOT NULL DEFAULT '',

  angelegt TEXT NOT NULL DEFAULT (datetime('now')),

  UNIQUE (anbieter, subjekt),
  UNIQUE (nutzer_id, anbieter),

  -- **Die Liste steht auch hier**, nicht nur in `domain`. Der `CHECK` ist die
  -- zweite Reihe: er greift, wo niemand durch den Fachkern geht.
  CHECK (anbieter IN ('github', 'google', 'microsoft', 'x')),
  CHECK (trim(subjekt) <> '')
) STRICT;

CREATE TABLE nutzerfaktoren (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  nutzer_id   INTEGER NOT NULL REFERENCES nutzer(id) ON DELETE CASCADE,
  art         TEXT    NOT NULL CHECK (art IN ('totp', 'biometrie')),
  bezeichnung TEXT    NOT NULL DEFAULT '',
  angelegt_am TEXT    NOT NULL DEFAULT (datetime('now'))
, letzter_schritt INTEGER NOT NULL DEFAULT 0
  CHECK (letzter_schritt >= 0)) STRICT;

CREATE TABLE wiederherstellungscodes (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  nutzer_id    INTEGER NOT NULL REFERENCES nutzer(id) ON DELETE CASCADE,
  gehuetet     TEXT    NOT NULL,

  -- **Gebraucht statt gelöscht.** Wer seine Codes zählt, soll sehen, wie
  -- viele er noch hat — und ein gelöschter Code sähe aus wie einer, den es
  -- nie gab. Leer heißt: noch frei.
  gebraucht_am TEXT    NOT NULL DEFAULT ''
) STRICT;

CREATE TABLE bestandskennung (
  -- **Genau eine Zeile**, und der `CHECK` sagt es. Eine zweite wäre eine
  -- zweite Wahrheit darüber, welcher Bestand dies ist.
  id   INTEGER PRIMARY KEY CHECK (id = 1),
  wert TEXT    NOT NULL
) STRICT;

CREATE TABLE nutzereinstellungen (
  nutzer_id  INTEGER NOT NULL REFERENCES nutzer(id) ON DELETE CASCADE,
  schluessel TEXT    NOT NULL,
  wert       TEXT    NOT NULL,

  PRIMARY KEY (nutzer_id, schluessel)
) STRICT;

CREATE TABLE zugriffstoken (
  id          INTEGER PRIMARY KEY,
  profil_id   INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,

  -- bezeichnung sagt, wofür es da ist — „Sicherungsskript", „Editor".
  --
  -- **Ohne sie ist eine Liste von Token unbrauchbar**: sie sehen einander
  -- gleich, und wer eines abschalten will, weiß nicht welches.
  bezeichnung TEXT    NOT NULL,

  -- gehuetet ist SHA-256 über dem Geheimnis — nie der Klartext.
  gehuetet    TEXT    NOT NULL,

  -- scope ist `lesen` oder `schreiben`.
  scope       TEXT    NOT NULL,

  -- ablauf ist RFC 3339 — leer heißt: läuft nicht ab.
  ablauf      TEXT    NOT NULL DEFAULT '',

  angelegt    TEXT    NOT NULL,
  fassung     INTEGER NOT NULL DEFAULT 1
, principal TEXT NOT NULL DEFAULT '', letzte_nutzung TEXT NOT NULL DEFAULT '', vorgaenger_gehuetet TEXT NOT NULL DEFAULT '', vorgaenger_bis TEXT NOT NULL DEFAULT '', sitz TEXT NOT NULL DEFAULT '', sitz_partie TEXT NOT NULL DEFAULT ''
  CHECK ((sitz_partie = '') = (sitz = ''))) STRICT;

CREATE TABLE profilvorgaben (
  profil_id INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,

  -- art ist `format`, `spiel`, `vorlage`, `deck` oder `lagerort`.
  art       TEXT    NOT NULL,

  -- wert ist die Kennung einer Zeile oder der Name eines Pakets.
  --
  -- **Als Text, auch wo es eine Zahl ist.** Fünf Arten teilen sich diese
  -- Spalte; eine davon nennt einen Paketnamen, und ein Verbund aus Zahl und
  -- Text in einer Spalte ginge in einer `STRICT`-Tabelle gar nicht.
  wert      TEXT    NOT NULL,

  ordnung   INTEGER NOT NULL DEFAULT 0,

  PRIMARY KEY (profil_id, art, wert)
) STRICT;

CREATE TABLE profilfreigaben (
  profil_id INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,
  nutzer_id INTEGER NOT NULL REFERENCES nutzer(id) ON DELETE CASCADE,

  -- rolle ist `leser`, `mitarbeiter` oder `verwalter` — aufsteigend, und jede
  -- schließt die darunter ein.
  rolle     TEXT    NOT NULL,

  angelegt  TEXT    NOT NULL,

  PRIMARY KEY (profil_id, nutzer_id)
) STRICT;

CREATE TABLE deck_herkunft (
  deck_id      INTEGER PRIMARY KEY REFERENCES decks(id) ON DELETE CASCADE,

  -- `profil` oder `einfuhr` — die Prüfung steht in `domain.Herkunftsart`.
  art          TEXT    NOT NULL,

  -- Die lesbare Bezeichnung: ein Profilname, ein Dateiname.
  quelle       TEXT    NOT NULL,

  -- Die Kennung des Quellprofils — 0 bei einer Einfuhr. **Ohne Fremdschlüssel**,
  -- siehe oben.
  quell_profil INTEGER NOT NULL DEFAULT 0,

  -- Das Deck, aus dem es übernommen wurde — 0, wenn es keines gab. Ebenfalls
  -- ohne Fremdschlüssel: es kann im Quellprofil längst gelöscht sein.
  quell_deck   INTEGER NOT NULL DEFAULT 0,

  uebernommen  TEXT    NOT NULL
) STRICT;

CREATE TABLE sammlung_herkunft (
  exemplar_id  INTEGER PRIMARY KEY REFERENCES sammlung(id) ON DELETE CASCADE,

  art          TEXT    NOT NULL,
  quelle       TEXT    NOT NULL,
  quell_profil INTEGER NOT NULL DEFAULT 0,
  uebernommen  TEXT    NOT NULL
) STRICT;

CREATE TABLE freigabenachweise (
  kennung     TEXT    PRIMARY KEY,
  nutzer_id   INTEGER NOT NULL,
  profil_id   INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,
  zweck       TEXT    NOT NULL,
  ziel        TEXT    NOT NULL,
  revision    INTEGER NOT NULL,

  -- grenze_betrag ist die Kostengrenze in der kleinsten Einheit; null heißt
  -- keine Grenze, und dann ist die Währung leer.
  grenze_betrag   INTEGER NOT NULL DEFAULT 0,
  grenze_waehrung TEXT    NOT NULL DEFAULT '',

  -- Zeiten als RFC 3339 in UTC — wie überall im Bestand.
  ausgestellt TEXT    NOT NULL,
  ablauf      TEXT    NOT NULL,

  eingeloest  INTEGER NOT NULL DEFAULT 0,

  CHECK (trim(zweck) <> '' AND trim(ziel) <> ''),
  CHECK (eingeloest IN (0, 1))
) STRICT;

CREATE TABLE freigabenrevision (
  id       INTEGER PRIMARY KEY CHECK (id = 1),
  revision INTEGER NOT NULL DEFAULT 0
) STRICT;

CREATE TABLE arbeitsverlauf (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  profil_id INTEGER NOT NULL,
  art       TEXT    NOT NULL,
  kennung   TEXT    NOT NULL,
  name      TEXT    NOT NULL DEFAULT '',
  vorgang   TEXT    NOT NULL,
  zeitpunkt TEXT    NOT NULL,

  CHECK (art IN ('set', 'deck', 'paket')),
  CHECK (vorgang IN ('geoeffnet', 'eingefuehrt', 'ausgegeben')),
  CHECK (kennung <> ''),

  -- **Ein Schritt je Objekt und Vorgang** — der jüngste Zeitpunkt gewinnt.
  UNIQUE (profil_id, art, kennung, vorgang)
) STRICT;

CREATE TABLE angeheftet (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,
  profil_id INTEGER NOT NULL,
  art       TEXT    NOT NULL,
  kennung   TEXT    NOT NULL,
  name      TEXT    NOT NULL DEFAULT '',
  zeitpunkt TEXT    NOT NULL,

  CHECK (art IN ('set', 'deck', 'paket')),
  CHECK (kennung <> ''),

  UNIQUE (profil_id, art, kennung)
) STRICT;

CREATE TABLE drucke (
  id      INTEGER PRIMARY KEY AUTOINCREMENT,
  name    TEXT    NOT NULL DEFAULT '',
  setcode TEXT    NOT NULL DEFAULT '',
  nummer  TEXT    NOT NULL DEFAULT '',
  sprache TEXT    NOT NULL DEFAULT ''
) STRICT;

CREATE TABLE "deckplaetze" (
  id       INTEGER PRIMARY KEY AUTOINCREMENT,
  deck_id  INTEGER NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
  karte_id INTEGER REFERENCES cards(id) ON DELETE CASCADE,
  printing TEXT,
  druck_id INTEGER REFERENCES drucke(id),
  menge    INTEGER NOT NULL,
  bereich  TEXT    NOT NULL,
  stelle   INTEGER NOT NULL DEFAULT 0,

  -- **Genau ein Bezug**: die eigene Karte — oder ein Druck, ob mit Alias
  -- oder nur als Zeile.
  CHECK ((karte_id IS NOT NULL) <> (printing IS NOT NULL OR druck_id IS NOT NULL)),

  CHECK (menge >= 1),
  CHECK (bereich IN ('hauptdeck', 'sideboard'))
) STRICT;

CREATE TABLE deckzuordnungen (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  deck_id     INTEGER NOT NULL REFERENCES decks(id) ON DELETE CASCADE,
  platz_id    INTEGER NOT NULL REFERENCES deckplaetze(id) ON DELETE CASCADE,
  exemplar_id INTEGER NOT NULL REFERENCES sammlung(id) ON DELETE CASCADE,
  menge       INTEGER NOT NULL,

  CHECK (menge >= 1),

  -- **Dieselbe Zeile zweimal ist eine neue Menge**, keine zweite Zuordnung.
  UNIQUE (platz_id, exemplar_id)
) STRICT;

CREATE TABLE wiederherstellungen (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,

  -- zurueckgespielt ist der Zeitpunkt der Rückkehr, datenstand der Zeitpunkt,
  -- **auf** den zurückgegangen wurde. Zwei Zahlen, die man leicht verwechselt
  -- — und die Abnahme verlangt ausdrücklich die zweite.
  zurueckgespielt TEXT    NOT NULL,
  datenstand      TEXT    NOT NULL,

  -- quelle ist der Name der Sicherung, für den Menschen.
  quelle          TEXT    NOT NULL,

  klaerung        TEXT    NOT NULL,
  geklaert_am     TEXT    NOT NULL DEFAULT '',

  CHECK (klaerung IN ('offen', 'quittiert'))
) STRICT;

CREATE TABLE kostenbudgets (
  profil_id    INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,

  -- konto ist das Providerkonto — sein **Name**, nicht sein Schlüssel.
  konto        TEXT    NOT NULL,

  zeitraum     TEXT    NOT NULL,

  -- betrag in der kleinsten Einheit der Währung; **ganzzahlig**, damit
  -- keine Rundung zwischen Grenze und Summe steht.
  betrag       INTEGER NOT NULL,
  waehrung     TEXT    NOT NULL,

  geaendert_am TEXT    NOT NULL,

  CHECK (betrag >= 0),
  CHECK (zeitraum IN ('tag', 'monat')),
  CHECK (length(waehrung) = 3),

  PRIMARY KEY (profil_id, konto, zeitraum)
) STRICT;
