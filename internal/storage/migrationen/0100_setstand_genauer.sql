-- Der Setstand sieht Stiloptionen und den Ort der Bilder, und er hebt nur, was
-- sich ändert — Befunde 1 und 14 des Zweitblicks über Phase 3 von 0.9.0
-- (PERF02, #2352).
--
-- **Was 0099 nicht sah**: `styling_values` — die Optionen des Stils, die die
-- Vorschau frisch liest (`StiloptionSetzen` berührt `sets.stil` nicht) — und
-- `sets.herkunft`, den Ort, an dem die eigenen Bilder eines eingelesenen Sets
-- liegen. Eine Prüfung blieb über beide hinweg „aktuell".
--
-- **Was 0099 zu oft hob**: jedes `UPDATE`, auch eines, das denselben Wert
-- schrieb. Der Massenschreiber führt je Karte einen Upsert aus; eine
-- Massenänderung, die nichts änderte, machte eine Prüfung über 3.788 Karten
-- unnötig veraltet. Jetzt vergleicht jeder Auslöser `OLD` und `NEW` — mit
-- `IS NOT`, damit ein `NULL` nicht als gleich gilt.
--
-- **0099 bleibt, wie es ist** (seine Prüfsumme steht in jedem Bestand): diese
-- Wanderung ersetzt die Auslöser, die sich ändern, und legt drei dazu.

DROP TRIGGER set_aenderungen_set;
CREATE TRIGGER set_aenderungen_set AFTER UPDATE OF spiel, stil, name, herkunft, spielfassung, stilfassung ON sets
WHEN OLD.spiel IS NOT NEW.spiel OR OLD.stil IS NOT NEW.stil OR OLD.name IS NOT NEW.name
  OR OLD.herkunft IS NOT NEW.herkunft OR OLD.spielfassung IS NOT NEW.spielfassung
  OR OLD.stilfassung IS NOT NEW.stilfassung
BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id = NEW.id;
END;

DROP TRIGGER set_aenderungen_karte;
CREATE TRIGGER set_aenderungen_karte AFTER UPDATE OF set_id, reihenfolge ON cards
WHEN OLD.set_id IS NOT NEW.set_id OR OLD.reihenfolge IS NOT NEW.reihenfolge
BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id IN (OLD.set_id, NEW.set_id);
END;

DROP TRIGGER set_aenderungen_wert;
CREATE TRIGGER set_aenderungen_wert AFTER UPDATE ON card_values
WHEN OLD.card_id IS NOT NEW.card_id OR OLD.feld IS NOT NEW.feld OR OLD.wert IS NOT NEW.wert
  OR OLD.reihenfolge IS NOT NEW.reihenfolge
BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1
     WHERE set_id IN (SELECT set_id FROM cards WHERE id IN (OLD.card_id, NEW.card_id));
END;

DROP TRIGGER set_aenderungen_setwert;
CREATE TRIGGER set_aenderungen_setwert AFTER UPDATE ON set_values
WHEN OLD.set_id IS NOT NEW.set_id OR OLD.feld IS NOT NEW.feld OR OLD.wert IS NOT NEW.wert
  OR OLD.reihenfolge IS NOT NEW.reihenfolge
BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id IN (OLD.set_id, NEW.set_id);
END;

DROP TRIGGER set_aenderungen_wort;
CREATE TRIGGER set_aenderungen_wort AFTER UPDATE ON set_keywords
WHEN OLD.set_id IS NOT NEW.set_id OR OLD.reihenfolge IS NOT NEW.reihenfolge OR OLD.wort IS NOT NEW.wort
  OR OLD.muster IS NOT NEW.muster OR OLD.erinnerung IS NOT NEW.erinnerung OR OLD.regeln IS NOT NEW.regeln
  OR OLD.modus IS NOT NEW.modus
BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id IN (OLD.set_id, NEW.set_id);
END;

CREATE TRIGGER set_aenderungen_stil_neu AFTER INSERT ON styling_values BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id = NEW.set_id;
END;

CREATE TRIGGER set_aenderungen_stil AFTER UPDATE ON styling_values
WHEN OLD.set_id IS NOT NEW.set_id OR OLD.feld IS NOT NEW.feld OR OLD.wert IS NOT NEW.wert
  OR OLD.reihenfolge IS NOT NEW.reihenfolge
BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id IN (OLD.set_id, NEW.set_id);
END;

CREATE TRIGGER set_aenderungen_stil_weg AFTER DELETE ON styling_values BEGIN
    UPDATE set_aenderungen SET zaehler = zaehler + 1 WHERE set_id = OLD.set_id;
END;
