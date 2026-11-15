-- Erzeugt aus den Wanderungen bis 0097_faktorhaushalt.sql — nicht von Hand ändern.
-- Eingefroren am 2026-10-12 (Vermerk 0189) — eine Änderung am Schema ist eine
-- neue Wanderung unter migrationen/, nie eine Änderung hier.

CREATE TABLE partieereignisse (
    partie_id   TEXT    NOT NULL REFERENCES partien(id) ON DELETE CASCADE,

    -- folge ist maßgeblich; die Uhr ist ein Metadatum.
    folge       INTEGER NOT NULL,

    art         TEXT    NOT NULL,

    -- oeffentlich und privat liegen getrennt, weil sie getrennte
    -- Integritätsnachweise haben (#1458).
    oeffentlich TEXT    NOT NULL,
    privat      TEXT    NOT NULL DEFAULT '[]',

    kette       TEXT    NOT NULL,
    geschehen   TEXT    NOT NULL DEFAULT (datetime('now')),

    PRIMARY KEY (partie_id, folge)
) STRICT;

CREATE TABLE partiebefehle (
    partie_id   TEXT    NOT NULL REFERENCES partien(id) ON DELETE CASCADE,

    -- kennung kommt vom Absender — der Grund steht bei domain.Spielbefehl.
    kennung     TEXT    NOT NULL,

    absender    TEXT    NOT NULL,
    art         TEXT    NOT NULL,

    -- vonFolge und bisFolge zeigen auf den Ausschnitt des Stroms, den
    -- dieser Befehl erzeugt hat. **Ein Ausschnitt und keine Kopie**: die
    -- Ereignisse stehen einmal da, und eine zweite Fassung liefe davon.
    von_folge   INTEGER NOT NULL,
    bis_folge   INTEGER NOT NULL,

    angewendet  TEXT    NOT NULL DEFAULT (datetime('now')),

    PRIMARY KEY (partie_id, kennung)
) STRICT;

CREATE TABLE partiefreigaben (
    partie_id TEXT NOT NULL REFERENCES partien(id) ON DELETE CASCADE,

    -- von ist der Sitz, dessen private Ergänzungen es sind.
    von       TEXT NOT NULL,

    -- an ist, wer mitlesen darf — oder '*'.
    an        TEXT NOT NULL,

    gegeben   TEXT NOT NULL DEFAULT (datetime('now')),

    PRIMARY KEY (partie_id, von, an)
) STRICT;

CREATE TABLE matchbindungen (
    paarung_id INTEGER PRIMARY KEY
        REFERENCES turnierpaarungen (id) ON DELETE CASCADE,

    turnier_id INTEGER NOT NULL REFERENCES turniere (id) ON DELETE CASCADE,

    -- fassung ist die des Turniers beim Binden.
    fassung    INTEGER NOT NULL,

    partie_id  TEXT    NOT NULL,

    -- sitze ist die Zuordnung Teilnehmer → Sitz → Peer → Deckabdruck.
    sitze      TEXT    NOT NULL,

    -- hoechstens ist die Zahl der Spiele; 0 heißt: keine Grenze.
    hoechstens INTEGER NOT NULL DEFAULT 0,

    gebunden   TEXT    NOT NULL DEFAULT (datetime('now'))
) STRICT;

CREATE TABLE matchmeldungen (
    paarung_id    INTEGER NOT NULL
        REFERENCES matchbindungen (paarung_id) ON DELETE CASCADE,

    -- von ist der Sitz, der gemeldet hat.
    --
    -- **Er ist der Schlüssel, nicht ein Zeitpunkt.** Wer zweimal meldet, hat
    -- sich beim ersten Mal vertippt; zwei Zeilen desselben Sitzes wären zwei
    -- Aussagen, und die Übernahme müsste raten, welche gilt.
    von           TEXT    NOT NULL,

    fassung       INTEGER NOT NULL,
    partie_id     TEXT    NOT NULL,

    spiele_a      INTEGER NOT NULL,
    spiele_b      INTEGER NOT NULL,
    unentschieden INTEGER NOT NULL,

    -- kette ist der letzte Kettenabdruck des Ereignisstroms.
    kette         TEXT    NOT NULL,

    gemeldet      TEXT    NOT NULL DEFAULT (datetime('now')),

    PRIMARY KEY (paarung_id, von)
) STRICT;

CREATE TABLE handelsrevisionen (
    profil_id INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,
    vorgang   TEXT    NOT NULL,
    revision  INTEGER NOT NULL,

    abdruck   TEXT    NOT NULL,
    eltern    TEXT    NOT NULL DEFAULT '',
    fassung   INTEGER NOT NULL,
    zustand   TEXT    NOT NULL,

    -- angebot ist die Revision als JSON — der Grund steht oben.
    angebot   TEXT    NOT NULL,

    angelegt  TEXT    NOT NULL DEFAULT (datetime('now')),
    geaendert TEXT    NOT NULL DEFAULT (datetime('now')),

    PRIMARY KEY (profil_id, vorgang, revision),

    CHECK (revision >= 1),
    CHECK ((revision = 1 AND eltern = '') OR (revision > 1 AND eltern <> ''))
) STRICT;

CREATE TABLE reservierungen (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    profil_id   INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,

    vorgang     TEXT    NOT NULL,
    revision    INTEGER NOT NULL,

    exemplar_id INTEGER NOT NULL REFERENCES sammlung(id) ON DELETE CASCADE,
    menge       INTEGER NOT NULL,
    fassung     INTEGER NOT NULL,

    -- frist ist ein Zeitpunkt in RFC 3339 — leer heißt: ohne Frist.
    frist       TEXT    NOT NULL DEFAULT '',
    uebergeben  INTEGER NOT NULL DEFAULT 0,

    angelegt    TEXT    NOT NULL DEFAULT (datetime('now')),

    CHECK (menge >= 1),
    CHECK (revision >= 1),
    CHECK (uebergeben IN (0, 1))
) STRICT;

CREATE TABLE "wiederherstellungsmengen" (
  id                   INTEGER PRIMARY KEY AUTOINCREMENT,
  wiederherstellung_id INTEGER NOT NULL
                         REFERENCES wiederherstellungen(id) ON DELETE CASCADE,
  profil_id            INTEGER NOT NULL,

  schluessel           TEXT    NOT NULL,
  zustand              TEXT    NOT NULL,
  sprache              TEXT    NOT NULL,
  finish               TEXT    NOT NULL DEFAULT 'nonfoil',
  menge                INTEGER NOT NULL,

  name                 TEXT    NOT NULL DEFAULT '',
  setcode              TEXT    NOT NULL DEFAULT '',
  nummer               TEXT    NOT NULL DEFAULT '',

  CHECK (menge >= 1),

  UNIQUE (wiederherstellung_id, profil_id, schluessel, zustand, sprache, finish)
) STRICT;

CREATE TABLE sammlungsfreigaben (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    profil_id  INTEGER NOT NULL REFERENCES profile (id) ON DELETE CASCADE,

    -- empfaenger ist die Peerkennung des Gegenübers.
    empfaenger TEXT    NOT NULL,

    -- umfang ist einer aus `domain.DieFreigabeumfaenge`.
    umfang     TEXT    NOT NULL,

    -- zweck ist einer aus `domain.DieZwecke` — **nie `paarung`**.
    zweck      TEXT    NOT NULL,

    -- behalten ist die Behaltemenge bei `dubletten`, sonst 0.
    behalten   INTEGER NOT NULL DEFAULT 0,

    -- ablauf ist ein Zeitpunkt in RFC 3339 — **leer heißt unbefristet**.
    ablauf     TEXT    NOT NULL DEFAULT '',

    erteilt    TEXT    NOT NULL,

    CHECK (behalten >= 0),
    CHECK (zweck <> 'paarung'),

    UNIQUE (profil_id, empfaenger, zweck, umfang)
) STRICT;

CREATE TABLE sammlungsfreigabeauswahl (
    freigabe_id INTEGER NOT NULL
                  REFERENCES sammlungsfreigaben (id) ON DELETE CASCADE,
    printing    TEXT    NOT NULL,

    PRIMARY KEY (freigabe_id, printing)
) STRICT, WITHOUT ROWID;

CREATE TABLE handelsbriefkasten (
    profil_id  INTEGER NOT NULL REFERENCES profile (id) ON DELETE CASCADE,

    -- schluessel ist der Idempotenzschlüssel aus der Domäne — Profil,
    -- Vorgang, Operation und Revision, mit Nullbytes getrennt.
    schluessel TEXT    NOT NULL,

    -- abdruck ist der Hash über Umschlag **und** Nutzlast. Weicht er bei
    -- gleichem Schlüssel ab, ist das ein Konflikt und keine Wiederholung.
    abdruck    TEXT    NOT NULL,

    -- ergebnis ist die Antwort, die schon gegeben wurde — als JSON.
    ergebnis   TEXT    NOT NULL,

    -- vorgang steht daneben, damit sich ein Handel als Ganzes aufräumen
    -- lässt, ohne den Schlüssel zu zerlegen.
    vorgang    TEXT    NOT NULL,

    -- behandelt ist die **eigene** Uhr und nicht die des Absenders: dessen
    -- steht in keiner Unterschrift.
    behandelt  TEXT    NOT NULL DEFAULT (datetime('now')),

    PRIMARY KEY (profil_id, schluessel)
) STRICT, WITHOUT ROWID;

CREATE TABLE handelsausgang (
    profil_id   INTEGER NOT NULL REFERENCES profile (id) ON DELETE CASCADE,

    nachricht_id TEXT   NOT NULL,
    vorgang      TEXT   NOT NULL,

    -- empfaenger ist die Peerkennung des Gegenübers.
    empfaenger   TEXT   NOT NULL,

    -- nachricht ist Umschlag und Nutzlast als JSON.
    nachricht    TEXT   NOT NULL,

    -- versuche zählt, wie oft sie schon hinausging — für eine Drossel, die
    -- ein totes Gegenüber nicht endlos anruft.
    versuche     INTEGER NOT NULL DEFAULT 0,

    angelegt     TEXT   NOT NULL DEFAULT (datetime('now')),
    zuletzt      TEXT   NOT NULL DEFAULT '',

    PRIMARY KEY (profil_id, nachricht_id)
) STRICT, WITHOUT ROWID;

CREATE TABLE handelsjournal (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    profil_id INTEGER NOT NULL REFERENCES profile (id) ON DELETE CASCADE,

    -- art ist die Richtung: 'zugang' oder 'abgang'.
    art       TEXT    NOT NULL CHECK (art IN ('zugang', 'abgang')),

    -- anlass ist, warum gebucht wurde — die sechs aus der Domäne.
    anlass    TEXT    NOT NULL CHECK (anlass IN (
                  'tausch', 'kauf', 'verkauf',
                  'geschenk', 'einfuhr', 'korrektur')),

    -- vorgang ist der Handel, zu dem die Zeile gehört — leer bei allem
    -- anderen.
    vorgang   TEXT    NOT NULL DEFAULT '',

    -- bezug ist die Zeile, auf die sich eine Korrektur bezieht — 0 sonst.
    bezug     INTEGER NOT NULL DEFAULT 0,

    -- grund steht an einer Korrektur und sonst nirgends.
    grund     TEXT    NOT NULL DEFAULT '',

    printing  TEXT    NOT NULL,
    menge     INTEGER NOT NULL CHECK (menge > 0),
    zustand   TEXT    NOT NULL,
    sprache   TEXT    NOT NULL,
    finish    TEXT    NOT NULL DEFAULT 'nonfoil',

    -- exemplar_id, lagerort und notiz sind **lokal**: sie stehen hier und
    -- nicht im Beleg, der hinausgeht.
    exemplar_id INTEGER NOT NULL DEFAULT 0,
    lagerort    TEXT   NOT NULL DEFAULT '',
    notiz       TEXT   NOT NULL DEFAULT '',

    gebucht   TEXT    NOT NULL DEFAULT (datetime('now')),

    -- **Eine Korrektur nennt ihren Grund** — die Regel steht in der Domäne
    -- und hier noch einmal, weil eine Zeile ohne Grund im Journal schlimmer
    -- ist als keine.
    CHECK ((anlass <> 'korrektur') OR (grund <> '' AND bezug <> 0))
) STRICT;

CREATE TABLE handelsbelege (
    profil_id    INTEGER NOT NULL REFERENCES profile (id) ON DELETE CASCADE,
    vorgang      TEXT    NOT NULL,

    -- abdruck ist der Abdruck der Revision, die abgeschlossen wurde.
    abdruck      TEXT    NOT NULL,

    -- form ist die Unterschriftsform als Text — das, was unterschrieben
    -- wurde, und nicht das, was jemand daraus rechnet.
    form         TEXT    NOT NULL,

    unterschrift TEXT    NOT NULL,
    beidseitig   INTEGER NOT NULL DEFAULT 0,

    gebucht      TEXT    NOT NULL DEFAULT (datetime('now')),

    PRIMARY KEY (profil_id, vorgang)
) STRICT, WITHOUT ROWID;

CREATE TABLE "druckkennungen" (
  druck_id INTEGER NOT NULL REFERENCES drucke(id) ON DELETE CASCADE,
  system   TEXT    NOT NULL,
  semantik TEXT    NOT NULL,
  wert     TEXT    NOT NULL,

  CHECK (system IN ('scryfall', 'mtgjson', 'cardmarket', 'cardtrader',
                    'tcgplayer', 'scrydex', 'gatherer')),
  CHECK (semantik IN ('druck', 'konzept', 'produkt', 'kunstwerk')),
  CHECK (wert <> ''),

  UNIQUE (druck_id, system, semantik, wert)
) STRICT;

CREATE TABLE handelsfolge (
    profil_id INTEGER NOT NULL REFERENCES profile (id) ON DELETE CASCADE,
    vorgang   TEXT    NOT NULL,

    -- folge ist die zuletzt vergebene Nummer — **ab eins**, und sie geht nie
    -- zurück.
    folge     INTEGER NOT NULL DEFAULT 0,

    PRIMARY KEY (profil_id, vorgang),

    CHECK (folge >= 0)
) STRICT, WITHOUT ROWID;

CREATE TABLE handelsfremdbelege (
    profil_id    INTEGER NOT NULL REFERENCES profile (id) ON DELETE CASCADE,
    vorgang      TEXT    NOT NULL,

    -- von ist die Peerkennung dessen, der unterschrieben hat.
    --
    -- **Sie kommt aus der Leitung und nicht aus der Nutzlast**: wer spricht,
    -- sagt die Sitzung.
    von          TEXT    NOT NULL,

    -- abdruck ist der Abdruck der Revision, die drüben abgeschlossen wurde.
    abdruck      TEXT    NOT NULL,

    unterschrift TEXT    NOT NULL,

    empfangen    TEXT    NOT NULL DEFAULT (datetime('now')),

    PRIMARY KEY (profil_id, vorgang)
) STRICT, WITHOUT ROWID;

CREATE TABLE "deckleihen" (
    kennung    TEXT    NOT NULL,
    profil_id  INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,

    -- dokument ist die unterschriebene Leihe als JSON.
    dokument   TEXT    NOT NULL,

    -- abzug ist der eingefrorene Abzug als JSON.
    abzug      TEXT    NOT NULL,

    verliehen  INTEGER NOT NULL,
    verbraucht TEXT    NOT NULL DEFAULT '',
    widerrufen INTEGER NOT NULL DEFAULT 0,

    angelegt   TEXT    NOT NULL DEFAULT (datetime('now')),
    geaendert  TEXT    NOT NULL DEFAULT (datetime('now')),

    PRIMARY KEY (profil_id, kennung),

    CHECK (verliehen IN (0, 1) AND widerrufen IN (0, 1))
) STRICT;
