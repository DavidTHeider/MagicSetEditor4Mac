-- Erzeugt aus den Wanderungen bis 0097_faktorhaushalt.sql — nicht von Hand ändern.
-- Eingefroren am 2026-10-12 (Vermerk 0189) — eine Änderung am Schema ist eine
-- neue Wanderung unter migrationen/, nie eine Änderung hier.

CREATE TABLE "preisstaende" (
  id        INTEGER PRIMARY KEY AUTOINCREMENT,

  printing  TEXT NOT NULL,
  druck_id  INTEGER REFERENCES drucke(id),

  -- Wer den Preis genannt hat — `scryfall`, `mtgjson`, ein Händler.
  quelle    TEXT NOT NULL,

  -- Wo er gilt: `paper` oder `mtgo`. MTGJSONs Vokabular (§9).
  markt     TEXT NOT NULL,

  -- In welche Richtung: `retail` (was der Anbieter verlangt) oder `buylist`
  -- (was er zahlt). Wer beides in einer Reihe führte, verglich einen
  -- Einkaufs- mit einem Verkaufspreis.
  art       TEXT NOT NULL,

  waehrung  TEXT NOT NULL,
  finish    TEXT NOT NULL DEFAULT 'nonfoil',

  -- Erhaltung und Sprachfassung — leer heißt unbekannt. Heute füllt sie keine
  -- Quelle; sie stehen im Schlüssel, weil A4 (#1443) Angebote bringt, die sich
  -- **genau darin** unterscheiden.
  zustand   TEXT NOT NULL DEFAULT '',
  sprache   TEXT NOT NULL DEFAULT '',

  -- `genannt`, `weggefallen` oder `ungeprueft`.
  lage      TEXT NOT NULL DEFAULT 'genannt',

  -- In Hundertsteln. Null, wo die Lage nicht `genannt` ist — und dort
  -- bedeutet sie nichts.
  wert      INTEGER NOT NULL,

  -- Der Tag, den die **Quelle** nennt, als Unix-Sekunden — null heißt: keiner.
  beobachtet INTEGER NOT NULL DEFAULT 0,

  -- Der Zeitpunkt des Abrufs, als Unix-Sekunden.
  erhoben   INTEGER NOT NULL,

  -- **Ein Stand je Reihe und Zeitpunkt.** Die Reihe ist alles, was die Frage
  -- benennt; der Zeitpunkt ist, wann gefragt wurde.
  UNIQUE (printing, quelle, markt, art, waehrung, finish, zustand, sprache, erhoben)
) STRICT;

CREATE TABLE ki_modellproben (
  profil_id  INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,

  -- anbieter ist sein **Name** — `openai`, `anthropic`, `ollama`.
  anbieter   TEXT    NOT NULL,

  -- alias ist der Name, unter dem gefragt wurde.
  alias      TEXT    NOT NULL,

  -- faehigkeit ist eine der sieben aus `domain.KIFaehigkeit`.
  --
  -- **Ohne `CHECK`, und das ist eine Entscheidung.** A5 hat gezeigt, was
  -- eine zweite Aufzählung kostet (Wanderung 0056); die Liste steht in Go,
  -- und `domain.KIModellprobe.Validate` hält sie. Ein `CHECK` daneben liefe
  -- beim achten Wort auseinander, und der Fehler käme aus SQLite statt aus
  -- der Fachlogik.
  faehigkeit TEXT    NOT NULL,

  -- datum ist der Tag der Messung, JJJJ-MM-TT.
  --
  -- **Eine Prüfung altert** ([domain.KIGeprueft]): sie sagt, wie alt die
  -- Auskunft ist, nicht dass sie noch stimmt.
  datum      TEXT    NOT NULL,

  -- beleg sagt, **wie** gemessen wurde.
  beleg      TEXT    NOT NULL,

  -- laufzeit ist, womit geantwortet wurde — samt Fassung, wo es eine gibt.
  laufzeit   TEXT    NOT NULL,

  -- modell ist, was der Anbieter als antwortendes Modell genannt hat.
  --
  -- **Nicht dasselbe wie der Alias.** Ein Alias zeigt auf eine Fassung, und
  -- wenn sie getauscht wird, ist die Messung eine über etwas anderes. Wer
  -- beides nebeneinander hat, sieht es.
  modell     TEXT    NOT NULL,

  -- **Eine Messung je Profil, Anbieter, Alias und Fähigkeit.** Eine zweite
  -- ersetzt die erste: was gestern galt, sagt über heute nichts, und zwei
  -- Daten nebeneinander ließen offen, welches gilt.
  PRIMARY KEY (profil_id, anbieter, alias, faehigkeit)
) STRICT, WITHOUT ROWID;

CREATE TABLE ki_feldherkunft (
  profil_id INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,

  -- karte_id ist die Karte, an der das Feld steht.
  --
  -- **Mit `ON DELETE CASCADE`**: eine gelöschte Karte lässt keine Herkunft
  -- zurück, die auf nichts zeigt.
  karte_id  INTEGER NOT NULL REFERENCES cards(id) ON DELETE CASCADE,

  -- feld ist der Feldname, wie ihn das Paket nennt.
  feld      TEXT    NOT NULL,

  -- anbieter ist, wen man gefragt hat — `ollama`, `anthropic`, `openai`, `xai`.
  anbieter  TEXT    NOT NULL,

  -- modell ist, was **geantwortet** hat.
  --
  -- **Nicht der Alias, unter dem gefragt wurde**: ein Anbieter darf tauschen,
  -- was dahinterliegt, und dann wäre die Auskunft eine über etwas anderes.
  modell    TEXT    NOT NULL,

  -- aufgabe ist, was ein Mensch getippt hat.
  aufgabe   TEXT    NOT NULL,

  -- auftrag ist die Kennung des Auftrags — sie bindet die Felder **einer**
  -- Übernahme zusammen.
  auftrag   TEXT    NOT NULL,

  -- zeitpunkt ist, wann übernommen wurde — RFC 3339 in UTC.
  zeitpunkt TEXT    NOT NULL,

  -- bezug ist der Abdruck des übernommenen Werts (`sha256:…`).
  --
  -- **Ein Abdruck und keine Kopie.** Der Wert steht im Feld; ihn hier
  -- zweitzuschreiben hieße, ihn an zwei Stellen zu pflegen. Wozu er da ist:
  -- später lässt sich sagen, ob das Feld **noch** das trägt, was das Modell
  -- vorgeschlagen hat — wer danach von Hand ändert, darf das, und dann soll
  -- die Herkunft es nicht verschweigen.
  bezug     TEXT    NOT NULL,

  -- **Eine Herkunft je Profil, Karte und Feld.** Eine zweite Übernahme
  -- ersetzt die erste: was am Feld steht, kommt aus **einer** Antwort, und
  -- zwei Herkünfte nebeneinander ließen offen, welche gilt.
  PRIMARY KEY (profil_id, karte_id, feld)
) STRICT, WITHOUT ROWID;

CREATE TABLE ki_bildherkunft (
  profil_id INTEGER NOT NULL,

  karte_id  INTEGER NOT NULL,

  -- feld ist das Bildfeld der Karte.
  feld      TEXT    NOT NULL,

  -- name ist die Angabe — `groesse`, `guete`, `umschrift`, `seed`.
  name      TEXT    NOT NULL,

  -- wert ist, was der Dienst dazu gesagt hat.
  wert      TEXT    NOT NULL,

  -- **Eine Angabe je Name.** Zweimal `groesse` mit verschiedenen Werten wäre
  -- eine Auskunft, die sich selbst widerspricht.
  PRIMARY KEY (profil_id, karte_id, feld, name),

  -- **Mit der Feldherkunft zusammen, nicht neben ihr.** Wird die Zeile in
  -- `ki_feldherkunft` ersetzt oder fällt sie mit ihrer Karte weg, gehen die
  -- Angaben mit — sonst stünde hier die Güte eines Bildes, das es nicht mehr
  -- gibt.
  FOREIGN KEY (profil_id, karte_id, feld)
    REFERENCES ki_feldherkunft(profil_id, karte_id, feld) ON DELETE CASCADE
) STRICT, WITHOUT ROWID;

CREATE TABLE retrieval_generationen (
  -- generation zählt hoch; die höchste ist nicht zwingend die aktive.
  generation    INTEGER PRIMARY KEY,

  -- manifest ist das vollständige Manifest als JSON.
  --
  -- **Als JSON und nicht als zwanzig Spalten**: es ist ein Beleg, den
  -- niemand abfragt — gelesen wird es ganz oder gar nicht, und seine Felder
  -- stehen in `domain.Indexmanifest`.
  manifest      TEXT    NOT NULL,

  -- fingerabdruck ist der Abdruck des Manifests — für den Vergleich, ohne
  -- das JSON zu lesen.
  fingerabdruck TEXT    NOT NULL,

  -- raumabdruck ist der Abdruck des **Vektorraums** allein.
  --
  -- **Getrennt vom Fingerabdruck**, und das ist der Kern: ein Index mit
  -- anderer Suchkonfiguration ist ein anderer Index, aber derselbe Raum.
  raumabdruck   TEXT    NOT NULL,

  gebaut        TEXT    NOT NULL,

  -- zustand ist `baut`, `fertig` oder `verworfen`.
  --
  -- **Ein Neubau ist sichtbar, solange er läuft** — sonst sähe ein Abbruch
  -- aus wie ein Index, der nie begonnen wurde, und die Zeilen blieben liegen.
  zustand       TEXT    NOT NULL
                  CHECK (zustand IN ('baut', 'fertig', 'verworfen')),

  -- aktiv ist 1 bei genau einer Generation — oder bei keiner.
  aktiv         INTEGER NOT NULL DEFAULT 0 CHECK (aktiv IN (0, 1))
) STRICT;

CREATE TABLE retrieval_dokumente (
  id             INTEGER PRIMARY KEY,

  generation     INTEGER NOT NULL
                   REFERENCES retrieval_generationen(generation) ON DELETE CASCADE,

  -- kennung benennt den Abschnitt innerhalb seiner Quelle.
  kennung        TEXT    NOT NULL,

  quelle         TEXT    NOT NULL,
  bezugsadresse  TEXT    NOT NULL,
  spiel          TEXT    NOT NULL,

  -- kennungssystem, fachkennung und semantik sind die Identität nach A2 —
  -- leer bei einer Regelstelle, die keine Karte nennt.
  kennungssystem TEXT    NOT NULL,
  fachkennung    TEXT    NOT NULL,
  semantik       TEXT    NOT NULL,

  regelnummer    TEXT    NOT NULL,
  urheber        TEXT    NOT NULL,

  sprache        TEXT    NOT NULL,
  titel          TEXT    NOT NULL,
  text           TEXT    NOT NULL,

  quellrevision  TEXT    NOT NULL,

  -- quelldatum ist das **Gültigkeitsdatum**, abgerufen die Abrufzeit.
  quelldatum     TEXT    NOT NULL,
  abgerufen      TEXT    NOT NULL,

  inhaltshash    TEXT    NOT NULL,

  -- merkmale sind die strukturierten Filterangaben als JSON.
  --
  -- **Als JSON und nicht als Spalten**, weil sie mehrwertig sind: Farben,
  -- Farbidentität, Typzeilenwörter und Formatlagen sind Listen und eine
  -- Abbildung. Fünf Nebentabellen für einen Filter, der ohnehin über die
  -- Kandidaten läuft, wären fünf Verknüpfungen je Suche.
  --
  -- **Leer ist `{}` und nicht `NULL`** — ein Dokument ohne Merkmale hat
  -- keine, und das ist eine Aussage.
  merkmale       TEXT    NOT NULL DEFAULT '{}',

  abschnitt      INTEGER NOT NULL,
  von            INTEGER NOT NULL,
  bis            INTEGER NOT NULL,

  -- profil_id ist 0 bei öffentlichen Dokumenten.
  --
  -- **Ohne Fremdschlüssel auf `profile`**: ein gelöschtes Profil nimmt seine
  -- Zeilen über den Weg in `service` mit, und ein `ON DELETE CASCADE` hier
  -- räumte sie hinter dem Rücken der Tombstones weg.
  profil_id      INTEGER NOT NULL,
  sichtbarkeit   TEXT    NOT NULL
                   CHECK (sichtbarkeit IN ('oeffentlich', 'profil')),

  -- **Ein Abschnitt je Generation.** Zweimal derselbe wäre ein Treffer, der
  -- doppelt zählt.
  UNIQUE (generation, kennung)
) STRICT;

CREATE VIRTUAL TABLE retrieval_volltext USING fts5(
  titel, text, dokument_id UNINDEXED
);

CREATE TABLE retrieval_vektoren (
  dokument_id INTEGER PRIMARY KEY
                REFERENCES retrieval_dokumente(id) ON DELETE CASCADE,

  -- vektor ist little-endian float32, so lang wie die Dimension des
  -- Manifests.
  --
  -- **Ein BLOB und keine Tabelle mit einer Zeile je Zahl**: 384 Zeilen je
  -- Abschnitt wären bei dreißigtausend Abschnitten elf Millionen Zeilen für
  -- eine Rechnung, die ohnehin über alle geht.
  vektor      BLOB    NOT NULL
) STRICT, WITHOUT ROWID;

CREATE TABLE retrieval_gesperrt (
  kennung   TEXT    NOT NULL,

  -- profil_id ist 0 für eine Sperre, die für alle gilt.
  profil_id INTEGER NOT NULL,

  gesperrt  TEXT    NOT NULL,

  PRIMARY KEY (kennung, profil_id)
) STRICT, WITHOUT ROWID;

CREATE TABLE peeridentitaeten (
  profil_id INTEGER NOT NULL,

  -- kennung ist die öffentliche Peerkennung in ihrer Textform.
  kennung   TEXT    NOT NULL,

  -- seit sagt, wann diese Kennung entstanden ist.
  seit      TEXT    NOT NULL,

  -- vorher ist die Kennung, die vor einer Rotation galt — leer, wenn es
  -- keine gab.
  --
  -- **Sie steht hier, weil ein Kontakt sie kennt.** Wer rotiert, verlangt von
  -- jedem bekannten Gegenüber eine neue Prüfung; ohne die alte Kennung wüsste
  -- niemand mehr, welche das war.
  vorher    TEXT    NOT NULL DEFAULT '',

  PRIMARY KEY (profil_id),
  FOREIGN KEY (profil_id) REFERENCES profile (id) ON DELETE CASCADE
) STRICT, WITHOUT ROWID;

CREATE TABLE peerkontakte (
  id           INTEGER PRIMARY KEY AUTOINCREMENT,
  profil_id    INTEGER NOT NULL,

  -- kennung ist die Peerkennung des Gegenübers — **der Nachweis**.
  kennung      TEXT    NOT NULL,

  -- fingerabdruck ist die Lesehilfe für den Vergleich durch Menschen.
  fingerabdruck TEXT   NOT NULL DEFAULT '',

  -- name ist, wie das Gegenüber sich nennt. **Kein Nachweis.**
  name         TEXT    NOT NULL DEFAULT '',

  -- eigener_name ist, wie dieses Profil ihn nennt.
  --
  -- **Er steht hier, damit der selbstgewählte nicht gebraucht wird.** Zwei
  -- gleichnamige Peers sind sonst nicht zu unterscheiden.
  eigener_name TEXT    NOT NULL DEFAULT '',

  -- stand ist 'unbestaetigt', 'bestaetigt' oder 'gesperrt'.
  stand        TEXT    NOT NULL,

  -- gesehen ist der Zeitpunkt der letzten Begegnung.
  gesehen      TEXT    NOT NULL,

  -- fassung zählt die Änderungen — für das bedingte Schreiben.
  fassung      INTEGER NOT NULL DEFAULT 1, adresse TEXT NOT NULL DEFAULT '', notiz TEXT NOT NULL DEFAULT '', erstmals TEXT NOT NULL DEFAULT '', vorgaenger TEXT NOT NULL DEFAULT '',

  UNIQUE (profil_id, kennung),
  FOREIGN KEY (profil_id) REFERENCES profile (id) ON DELETE CASCADE
) STRICT;

CREATE TABLE peerrechte (
  kontakt_id INTEGER NOT NULL,

  -- zweck ist einer aus `domain.DieZwecke`.
  zweck      TEXT    NOT NULL,

  -- seit sagt, wann die Freigabe gegeben wurde.
  seit       TEXT    NOT NULL,

  PRIMARY KEY (kontakt_id, zweck),
  FOREIGN KEY (kontakt_id) REFERENCES peerkontakte (id) ON DELETE CASCADE
) STRICT, WITHOUT ROWID;

CREATE TABLE partien (
    id          TEXT    PRIMARY KEY,
    profil_id   INTEGER NOT NULL REFERENCES profile(id) ON DELETE CASCADE,

    spiel       TEXT    NOT NULL,
    stand       TEXT    NOT NULL,
    fassung     INTEGER NOT NULL DEFAULT 1,

    -- zustand ist die ganze Partie als JSON — der Grund steht oben.
    zustand     TEXT    NOT NULL,

    angelegt    TEXT    NOT NULL DEFAULT (datetime('now')),
    geaendert   TEXT    NOT NULL DEFAULT (datetime('now'))
) STRICT;
