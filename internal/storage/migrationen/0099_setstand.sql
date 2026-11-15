-- Ein Änderungszähler je Set — PERF02 (#2352), Befund Z-03 des Audits 0.9.0.
--
-- Eine Problemprüfung über ein ganzes Set gilt nur, solange das Set dasselbe
-- ist. Ihr Ergebnis trägt den Zähler, mit dem es begann; wer liest, vergleicht
-- ihn mit dem heutigen. **Ein Zähler in der Datenbank und nicht im Prozess**:
-- Fenster, Kommandozeile, Werkzeugserver und Dienst schreiben in denselben
-- Bestand, und jeder Weg muss ihn heben — auch die, die keinen Verlaufsschritt
-- schreiben (`kartenschreiber.go`, »Ohne Verlaufseintrag«).
--
-- Gehoben wird bei allem, was eine gezeichnete Karte verändern kann: Karten,
-- ihre Werte, die Werte und Schlüsselwörter des Sets, Spiel, Stil und deren
-- Fassung. **Nicht** bei `geoeffnet`: das Aufschlagen ändert keine Karte.
CREATE TABLE set_aenderungen (
    set_id  INTEGER PRIMARY KEY REFERENCES sets(id) ON DELETE CASCADE,
    zaehler INTEGER NOT NULL DEFAULT 0,

    CHECK (zaehler >= 0)
) STRICT;

INSERT INTO set_aenderungen (set_id) SELECT id FROM sets;

CREATE TRIGGER set_aenderungen_set_neu AFTER INSERT ON sets BEGIN
    INSERT INTO set_aenderungen (set_id) VALUES (NEW.id);
END;

CREATE TRIGGER set_aenderungen_set AFTER UPDATE OF spiel, stil, name, spielfassung, stilfassung ON sets BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id = NEW.id;
END;

CREATE TRIGGER set_aenderungen_karte_neu AFTER INSERT ON cards BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id = NEW.set_id;
END;

CREATE TRIGGER set_aenderungen_karte_weg AFTER DELETE ON cards BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id = OLD.set_id;
END;

CREATE TRIGGER set_aenderungen_karte AFTER UPDATE OF set_id, reihenfolge ON cards BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id IN (OLD.set_id, NEW.set_id);
END;

CREATE TRIGGER set_aenderungen_wert_neu AFTER INSERT ON card_values BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1
     WHERE set_id = (SELECT set_id FROM cards WHERE id = NEW.card_id);
END;

CREATE TRIGGER set_aenderungen_wert AFTER UPDATE ON card_values BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1
     WHERE set_id IN (SELECT set_id FROM cards WHERE id IN (OLD.card_id, NEW.card_id));
END;

CREATE TRIGGER set_aenderungen_wert_weg AFTER DELETE ON card_values BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1
     WHERE set_id = (SELECT set_id FROM cards WHERE id = OLD.card_id);
END;

CREATE TRIGGER set_aenderungen_setwert_neu AFTER INSERT ON set_values BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id = NEW.set_id;
END;

CREATE TRIGGER set_aenderungen_setwert AFTER UPDATE ON set_values BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id IN (OLD.set_id, NEW.set_id);
END;

CREATE TRIGGER set_aenderungen_setwert_weg AFTER DELETE ON set_values BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id = OLD.set_id;
END;

CREATE TRIGGER set_aenderungen_wort_neu AFTER INSERT ON set_keywords BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id = NEW.set_id;
END;

CREATE TRIGGER set_aenderungen_wort AFTER UPDATE ON set_keywords BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id IN (OLD.set_id, NEW.set_id);
END;

CREATE TRIGGER set_aenderungen_wort_weg AFTER DELETE ON set_keywords BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id = OLD.set_id;
END;
