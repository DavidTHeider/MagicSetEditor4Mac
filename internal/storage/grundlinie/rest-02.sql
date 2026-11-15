-- Erzeugt aus den Wanderungen bis 0097_faktorhaushalt.sql — nicht von Hand ändern.
-- Eingefroren am 2026-10-12 (Vermerk 0189) — eine Änderung am Schema ist eine
-- neue Wanderung unter migrationen/, nie eine Änderung hier.

CREATE TRIGGER handelsbeleg_bleibt
BEFORE UPDATE ON handelsbelege
WHEN NEW.profil_id IS NOT OLD.profil_id OR NEW.vorgang IS NOT OLD.vorgang
  OR NEW.abdruck IS NOT OLD.abdruck OR NEW.form IS NOT OLD.form
  OR NEW.unterschrift IS NOT OLD.unterschrift OR NEW.gebucht IS NOT OLD.gebucht
  OR NEW.beidseitig < OLD.beidseitig
BEGIN
    SELECT RAISE(ABORT, 'ein Beleg wird nicht umgeschrieben — nur gegengezeichnet');
END;

CREATE TRIGGER handelsbeleg_geht_mit_dem_profil
BEFORE DELETE ON handelsbelege
WHEN EXISTS (SELECT 1 FROM profile WHERE id = OLD.profil_id)
BEGIN
    SELECT RAISE(ABORT, 'ein Beleg wird nicht gelöscht');
END;

CREATE TRIGGER handelsjournal_bleibt
BEFORE UPDATE ON handelsjournal
BEGIN
    SELECT RAISE(ABORT, 'eine Journalzeile wird nicht geändert');
END;

CREATE TRIGGER handelsjournal_geht_mit_dem_profil
BEFORE DELETE ON handelsjournal
WHEN EXISTS (SELECT 1 FROM profile WHERE id = OLD.profil_id)
BEGIN
    SELECT RAISE(ABORT, 'eine Journalzeile wird nicht gelöscht');
END;

CREATE TRIGGER handelsfremdbeleg_bleibt
BEFORE UPDATE ON handelsfremdbelege
BEGIN
    SELECT RAISE(ABORT, 'ein Fremdbeleg wird nicht geändert');
END;

CREATE TRIGGER handelsfremdbeleg_geht_mit_dem_profil
BEFORE DELETE ON handelsfremdbelege
WHEN EXISTS (SELECT 1 FROM profile WHERE id = OLD.profil_id)
BEGIN
    SELECT RAISE(ABORT, 'ein Fremdbeleg wird nicht gelöscht');
END;

DELETE FROM sqlite_sequence WHERE name <> 'migrationen';

INSERT INTO sqlite_sequence (name, seq) VALUES ('arbeitsverlauf', 0);

INSERT INTO sqlite_sequence (name, seq) VALUES ('deckplaetze', 0);

INSERT INTO sqlite_sequence (name, seq) VALUES ('deckzuordnungen', 0);

INSERT INTO sqlite_sequence (name, seq) VALUES ('drucke', 0);

INSERT INTO sqlite_sequence (name, seq) VALUES ('handelsjournal', 0);

INSERT INTO sqlite_sequence (name, seq) VALUES ('handelsprotokoll', 0);

INSERT INTO sqlite_sequence (name, seq) VALUES ('nutzer', 1);

INSERT INTO sqlite_sequence (name, seq) VALUES ('preisstaende', 0);

INSERT INTO sqlite_sequence (name, seq) VALUES ('profile', 1);

INSERT INTO sqlite_sequence (name, seq) VALUES ('reservierungen', 0);

INSERT INTO sqlite_sequence (name, seq) VALUES ('sammlung', 0);

INSERT INTO sqlite_sequence (name, seq) VALUES ('wiederherstellungsmengen', 0);
