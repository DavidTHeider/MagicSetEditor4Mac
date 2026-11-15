-- Erzeugt aus den Wanderungen bis 0097_faktorhaushalt.sql — nicht von Hand ändern.
-- Eingefroren am 2026-10-12 (Vermerk 0189) — eine Änderung am Schema ist eine
-- neue Wanderung unter migrationen/, nie eine Änderung hier.

CREATE TABLE "kostenposten" (
  kennung       TEXT    NOT NULL,

  profil_id     INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,
  konto         TEXT    NOT NULL,

  -- aufgabe ist eine Marke — `karte.text`, `bild.entwurf` —, kein Prompt.
  aufgabe       TEXT    NOT NULL,

  hoechstbetrag INTEGER NOT NULL,
  betrag        INTEGER NOT NULL DEFAULT 0,
  waehrung      TEXT    NOT NULL,

  stand         TEXT    NOT NULL,

  -- grund sagt, warum ein Posten unklar ist — für Menschen.
  grund         TEXT    NOT NULL DEFAULT '',

  begonnen      TEXT    NOT NULL,
  beendet       TEXT    NOT NULL DEFAULT '',

  PRIMARY KEY (profil_id, kennung),

  CHECK (hoechstbetrag > 0),
  CHECK (betrag >= 0),
  CHECK (length(waehrung) = 3),
  CHECK (stand IN ('reserviert', 'abgerechnet', 'freigegeben', 'unklar'))
) STRICT;

CREATE TABLE kontakte (
  id          INTEGER PRIMARY KEY,
  profil_id   INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,
  name        TEXT    NOT NULL,
  wege        TEXT    NOT NULL DEFAULT '[]',
  bild        BLOB,
  bildart     TEXT    NOT NULL DEFAULT '',
  ursprung    TEXT    NOT NULL DEFAULT '',
  uebernommen TEXT    NOT NULL,
  fassung     INTEGER NOT NULL DEFAULT 1,

  CHECK (name <> '' AND length(name) <= 200),
  CHECK (json_valid(wege) AND json_type(wege) = 'array'),
  CHECK ((bild IS NULL) = (bildart = '')),
  CHECK (bild IS NULL OR length(bild) <= 524288)
) STRICT;

CREATE TABLE "sammlung" (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  printing        TEXT    NOT NULL,
  menge           INTEGER NOT NULL,
  zustand         TEXT    NOT NULL,
  sprache         TEXT    NOT NULL,
  finish          TEXT    NOT NULL DEFAULT 'nonfoil',
  erwerbsdatum    TEXT    NOT NULL DEFAULT '',
  erwerbspreis    TEXT    NOT NULL DEFAULT '',
  erwerbswaehrung TEXT    NOT NULL DEFAULT '',
  notiz           TEXT    NOT NULL DEFAULT '',
  fassung         INTEGER NOT NULL DEFAULT 1,
  name            TEXT    NOT NULL DEFAULT '',
  setcode         TEXT    NOT NULL DEFAULT '',
  nummer          TEXT    NOT NULL DEFAULT '',
  lagerort        TEXT    NOT NULL DEFAULT '',
  wunsch          INTEGER NOT NULL DEFAULT 0,
  profil_id       INTEGER NOT NULL,
  druck_id        INTEGER REFERENCES drucke(id),
  wunschrang      TEXT    NOT NULL DEFAULT '',
  wunschart       TEXT    NOT NULL DEFAULT '',

  UNIQUE (profil_id, printing, druck_id, zustand, sprache, finish, wunsch)
) STRICT;

CREATE TABLE turnierprotokoll (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  turnier_id INTEGER NOT NULL REFERENCES turniere(id) ON DELETE CASCADE,
  zeit       TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  vorgang    TEXT NOT NULL,
  angaben    TEXT NOT NULL DEFAULT '',
  paarung    INTEGER,

  CHECK (vorgang IN ('angelegt', 'eingelesen', 'angemeldet', 'ausgestiegen', 'eingestiegen',
                     'runde', 'ergebnis', 'zurueckgenommen', 'abgeschlossen',
                     'wiedereroeffnet'))
) STRICT;

CREATE TABLE symbolbibliothek (
  id        INTEGER PRIMARY KEY,
  profil_id INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,
  name      TEXT    NOT NULL COLLATE NOCASE,
  inhalt    TEXT    NOT NULL,
  abdruck   TEXT    NOT NULL,
  vorher    TEXT    NOT NULL DEFAULT '',
  angelegt  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  geaendert TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),

  -- **Ein Name einmal je Profil**, ohne Rücksicht auf Groß und Klein — wer
  -- „Krone" und „krone" nebeneinander hat, weiß nicht, welches er einsetzt.
  UNIQUE (profil_id, name),

  -- Die zweite Reihe hinter `domain.PruefeSymbolname` und
  -- `msepaket.GrenzeSymboldatei` (docs/INVARIANTEN.md).
  CHECK (name <> '' AND length(name) <= 80),
  CHECK (length(inhalt) <= 1048576 AND length(vorher) <= 1048576),
  CHECK (abdruck <> '')
) STRICT;

CREATE TABLE "objektmarken" (
  profil_id INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,
  art       TEXT    NOT NULL,
  kennung   TEXT    NOT NULL,
  marke     TEXT    NOT NULL,

  PRIMARY KEY (profil_id, art, kennung, marke),

  CHECK (art IN ('set', 'exemplar', 'deck', 'paket', 'turnier', 'partie', 'handel', 'symbol')),
  CHECK (kennung <> ''),
  CHECK (marke <> '' AND length(marke) <= 40)
) STRICT;

CREATE TABLE sprachpakete (
  kennung   TEXT NOT NULL PRIMARY KEY,
  name      TEXT NOT NULL,
  basis     TEXT NOT NULL,
  stand     TEXT NOT NULL,
  lizenz    TEXT NOT NULL DEFAULT '',
  herkunft  TEXT NOT NULL DEFAULT '',
  texte     TEXT NOT NULL DEFAULT '{}',
  angelegt  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),
  geaendert TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),

  -- Die zweite Reihe hinter `domain.PruefeSprachkennung` und
  -- `domain.Sprachpaket.Pruefe` (docs/INVARIANTEN.md).
  CHECK (kennung <> '' AND length(kennung) <= 35),
  CHECK (json_valid(texte) AND json_type(texte) = 'object')
) STRICT;

CREATE TABLE handelsarchiv (
    profil_id  INTEGER NOT NULL REFERENCES profile (id) ON DELETE CASCADE,
    vorgang    TEXT    NOT NULL,
    archiviert TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now')),

    PRIMARY KEY (profil_id, vorgang),

    CHECK (vorgang <> '')
) STRICT, WITHOUT ROWID;

CREATE TABLE kontakteinladungen (
    profil_id  INTEGER NOT NULL REFERENCES profile (id) ON DELETE CASCADE,
    nonce      TEXT    NOT NULL,
    kennung    TEXT    NOT NULL,
    angenommen TEXT    NOT NULL,

    PRIMARY KEY (profil_id, nonce),

    CHECK (nonce <> '' AND kennung <> '')
) STRICT, WITHOUT ROWID;

CREATE TABLE handelsprotokoll (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    profil_id INTEGER NOT NULL REFERENCES profile (id) ON DELETE CASCADE,
    vorgang   TEXT    NOT NULL,
    revision  INTEGER NOT NULL,

    -- von ist leer, wenn die Revision neu ist.
    von       TEXT    NOT NULL,
    nach      TEXT    NOT NULL,
    zeit      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
) STRICT;

CREATE TABLE ausgestellte_einladungen (
    profil_id  INTEGER NOT NULL REFERENCES profile (id) ON DELETE CASCADE,
    nonce      TEXT    NOT NULL,
    zweck      TEXT    NOT NULL,
    ablauf     TEXT    NOT NULL,
    eingeloest TEXT    NOT NULL DEFAULT '',
    kennung    TEXT    NOT NULL DEFAULT '',

    PRIMARY KEY (profil_id, nonce),

    CHECK (nonce <> '' AND zweck <> '' AND ablauf <> ''),
    CHECK ((eingeloest = '') = (kennung = ''))
) STRICT, WITHOUT ROWID;

CREATE TABLE faktorhaushalt (
  nutzer_id    INTEGER PRIMARY KEY REFERENCES nutzer(id) ON DELETE CASCADE,
  fehlversuche INTEGER NOT NULL CHECK (fehlversuche > 0),
  gesperrt_bis INTEGER NOT NULL DEFAULT 0 CHECK (gesperrt_bis >= 0)
) STRICT;
