-- Wann der Bestand gesichert wurde — M01 (#2345), Befunde A-06 und A-12 des
-- Audits 0.9.0. Die erste Wanderung nach der Grundlinie (Vermerk 0189, Satz 6).
--
-- Eine Zeile je Sicherung, die geschrieben und geprüft ist. Gelesen wird die
-- jüngste; wer vermerkt, entscheidet der Aufrufer. Die Bedingungen sind die
-- zweite Reihe hinter `domain.Sicherungsvermerk.Pruefe` (docs/INVARIANTEN.md).
--
-- `datei` ist ein Name und kein Pfad: der Bestand wird weitergegeben, und ein
-- Pfad trüge den Namen des Benutzerkontos mit.
CREATE TABLE sicherungsvermerke (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    gesichert      TEXT    NOT NULL,
    anlass         TEXT    NOT NULL,
    flaeche        TEXT    NOT NULL,
    datei          TEXT    NOT NULL,
    groesse        INTEGER NOT NULL,
    pruefsumme     TEXT    NOT NULL,
    verschluesselt INTEGER NOT NULL DEFAULT 0,

    CHECK (gesichert GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9].[0-9][0-9][0-9]Z'),
    CHECK (anlass IN ('hand', 'plan', 'wanderung')),
    CHECK (flaeche IN ('fenster', 'cli', 'mcp', 'rest')),
    CHECK (datei <> '' AND instr(datei, '/') = 0 AND instr(datei, '\') = 0),
    CHECK (groesse > 0),
    CHECK (length(pruefsumme) = 64 AND pruefsumme NOT GLOB '*[^0-9a-f]*'),
    CHECK (verschluesselt IN (0, 1))
) STRICT;

CREATE INDEX sicherungsvermerke_nach_zeit ON sicherungsvermerke (gesichert);
