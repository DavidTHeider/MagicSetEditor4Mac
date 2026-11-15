-- Erzeugt aus den Wanderungen bis 0097_faktorhaushalt.sql — nicht von Hand ändern.
-- Eingefroren am 2026-10-12 (Vermerk 0189) — eine Änderung am Schema ist eine
-- neue Wanderung unter migrationen/, nie eine Änderung hier.

CREATE INDEX cards_nach_set    ON cards(set_id, reihenfolge);

CREATE INDEX werte_nach_karte  ON card_values(card_id, reihenfolge);

CREATE INDEX verlauf_nach_set  ON history(set_id, id);

CREATE INDEX verlauf_nach_gruppe ON history(set_id, gruppe);

CREATE INDEX karten_verweis_printing ON karten_verweis(printing);

CREATE INDEX karten_marken_marke ON karten_marken(marke);

CREATE UNIQUE INDEX turnierteilnehmer_stelle ON turnierteilnehmer (turnier_id, stelle);

CREATE UNIQUE INDEX turnierrunden_nummer ON turnierrunden (turnier_id, nummer);

CREATE UNIQUE INDEX turnierpaarungen_stelle ON turnierpaarungen (runde_id, stelle);

CREATE INDEX profile_nutzer ON profile (nutzer_id);

CREATE INDEX sets_profil     ON sets (profil_id);

CREATE INDEX decks_profil    ON decks (profil_id);

CREATE INDEX turniere_profil ON turniere (profil_id);

CREATE INDEX formate_profil  ON formate (profil_id);

CREATE TRIGGER sets_profil_pflicht BEFORE INSERT ON sets
WHEN NEW.profil_id = 0
BEGIN SELECT RAISE(ABORT, 'sets: profil_id fehlt'); END;

CREATE TRIGGER sets_profil_bleibt BEFORE UPDATE ON sets
WHEN NEW.profil_id = 0
BEGIN SELECT RAISE(ABORT, 'sets: profil_id fehlt'); END;

CREATE TRIGGER decks_profil_pflicht BEFORE INSERT ON decks
WHEN NEW.profil_id = 0
BEGIN SELECT RAISE(ABORT, 'decks: profil_id fehlt'); END;

CREATE TRIGGER decks_profil_bleibt BEFORE UPDATE ON decks
WHEN NEW.profil_id = 0
BEGIN SELECT RAISE(ABORT, 'decks: profil_id fehlt'); END;

CREATE TRIGGER turniere_profil_pflicht BEFORE INSERT ON turniere
WHEN NEW.profil_id = 0
BEGIN SELECT RAISE(ABORT, 'turniere: profil_id fehlt'); END;

CREATE TRIGGER turniere_profil_bleibt BEFORE UPDATE ON turniere
WHEN NEW.profil_id = 0
BEGIN SELECT RAISE(ABORT, 'turniere: profil_id fehlt'); END;

CREATE TRIGGER formate_profil_pflicht BEFORE INSERT ON formate
WHEN NEW.profil_id = 0
BEGIN SELECT RAISE(ABORT, 'formate: profil_id fehlt'); END;

CREATE TRIGGER formate_profil_bleibt BEFORE UPDATE ON formate
WHEN NEW.profil_id = 0
BEGIN SELECT RAISE(ABORT, 'formate: profil_id fehlt'); END;

CREATE INDEX anbieterkonten_nutzer ON anbieterkonten (nutzer_id);

CREATE UNIQUE INDEX nutzer_email   ON nutzer (email)   WHERE email <> '';

CREATE UNIQUE INDEX nutzer_telefon ON nutzer (telefon) WHERE telefon <> '';

CREATE UNIQUE INDEX nutzerfaktoren_art ON nutzerfaktoren (nutzer_id, art);

CREATE INDEX wiederherstellungscodes_nutzer ON wiederherstellungscodes (nutzer_id);

CREATE UNIQUE INDEX zugriffstoken_name ON zugriffstoken (profil_id, bezeichnung);

CREATE INDEX profilvorgaben_ordnung ON profilvorgaben (profil_id, art, ordnung);

CREATE INDEX profilfreigaben_nutzer ON profilfreigaben (nutzer_id, rolle);

CREATE UNIQUE INDEX decks_name_je_spiel ON decks (profil_id, spiel, name);

CREATE UNIQUE INDEX turniere_name_je_spiel ON turniere (profil_id, spiel, name);

CREATE UNIQUE INDEX formate_name_je_spiel ON formate (profil_id, spiel, name);

CREATE UNIQUE INDEX sets_herkunft_eindeutig
    ON sets (profil_id, herkunft) WHERE herkunft LIKE 'scryfall:%';

CREATE INDEX freigabenachweise_profil ON freigabenachweise (profil_id);

CREATE TRIGGER freigabenrevision_zeiger_neu
AFTER INSERT ON bestand
WHEN NEW.schluessel IN ('nutzer_zeiger', 'profil_zeiger')
BEGIN
  UPDATE freigabenrevision SET revision = revision + 1 WHERE id = 1;
END;

CREATE TRIGGER freigabenrevision_zeiger_um
AFTER UPDATE ON bestand
WHEN NEW.schluessel IN ('nutzer_zeiger', 'profil_zeiger')
BEGIN
  UPDATE freigabenrevision SET revision = revision + 1 WHERE id = 1;
END;

CREATE TRIGGER freigabenrevision_freigabe_neu
AFTER INSERT ON profilfreigaben
BEGIN
  UPDATE freigabenrevision SET revision = revision + 1 WHERE id = 1;
END;

CREATE TRIGGER freigabenrevision_freigabe_um
AFTER UPDATE ON profilfreigaben
BEGIN
  UPDATE freigabenrevision SET revision = revision + 1 WHERE id = 1;
END;

CREATE TRIGGER freigabenrevision_freigabe_weg
AFTER DELETE ON profilfreigaben
BEGIN
  UPDATE freigabenrevision SET revision = revision + 1 WHERE id = 1;
END;

CREATE TRIGGER card_values_eingefuegt AFTER INSERT ON card_values BEGIN
  INSERT INTO card_suche (rowid, feld, wert, card_id)
    VALUES (new.rowid, new.feld, new.wert, new.card_id);
END;

CREATE TRIGGER card_values_geaendert AFTER UPDATE ON card_values BEGIN
  DELETE FROM card_suche WHERE rowid = old.rowid;
  INSERT INTO card_suche (rowid, feld, wert, card_id)
    VALUES (new.rowid, new.feld, new.wert, new.card_id);
END;

CREATE TRIGGER card_values_geloescht AFTER DELETE ON card_values BEGIN
  DELETE FROM card_suche WHERE rowid = old.rowid;
END;

CREATE INDEX arbeitsverlauf_profil ON arbeitsverlauf (profil_id, zeitpunkt DESC);

CREATE INDEX drucke_rueckfall ON drucke (name, setcode, nummer, sprache);

CREATE INDEX karten_verweis_druck ON karten_verweis (druck_id);

CREATE INDEX turnierdeckplaetze_druck ON turnierdeckplaetze (druck_id);

CREATE INDEX deckplaetze_deck ON deckplaetze (deck_id, stelle);

CREATE UNIQUE INDEX deckplaetze_karte_je_bereich
  ON deckplaetze (deck_id, bereich, karte_id) WHERE karte_id IS NOT NULL;

CREATE UNIQUE INDEX deckplaetze_druck_je_bereich
  ON deckplaetze (deck_id, bereich, printing) WHERE printing IS NOT NULL;

CREATE UNIQUE INDEX deckplaetze_zeile_je_bereich
  ON deckplaetze (deck_id, bereich, druck_id) WHERE printing IS NULL AND druck_id IS NOT NULL;

CREATE INDEX deckplaetze_druck ON deckplaetze (druck_id);

CREATE INDEX deckzuordnungen_deck ON deckzuordnungen (deck_id);

CREATE INDEX deckzuordnungen_exemplar ON deckzuordnungen (exemplar_id);

CREATE INDEX preisstaende_neueste
  ON preisstaende(printing, quelle, markt, art, waehrung, finish,
                  zustand, sprache, erhoben DESC);

CREATE INDEX preisstaende_druck ON preisstaende (druck_id);

CREATE UNIQUE INDEX retrieval_eine_aktive
  ON retrieval_generationen (aktiv) WHERE aktiv = 1;

CREATE INDEX retrieval_dokumente_sicht
  ON retrieval_dokumente (generation, sichtbarkeit, profil_id);

CREATE INDEX partien_profil ON partien (profil_id, geaendert DESC);

CREATE INDEX matchbindungen_turnier ON matchbindungen (turnier_id);

CREATE UNIQUE INDEX handelsrevisionen_abdruck
    ON handelsrevisionen (profil_id, abdruck);

CREATE INDEX handelsrevisionen_zustand
    ON handelsrevisionen (profil_id, zustand, geaendert DESC);

CREATE INDEX reservierungen_exemplar
    ON reservierungen (profil_id, exemplar_id);

CREATE INDEX reservierungen_vorgang
    ON reservierungen (profil_id, vorgang);

CREATE INDEX sammlungsfreigaben_empfaenger
    ON sammlungsfreigaben (profil_id, empfaenger);

CREATE INDEX handelsbriefkasten_vorgang
    ON handelsbriefkasten (profil_id, vorgang);

CREATE INDEX handelsbriefkasten_behandelt
    ON handelsbriefkasten (behandelt);

CREATE INDEX handelsausgang_vorgang
    ON handelsausgang (profil_id, vorgang);

CREATE INDEX handelsjournal_profil ON handelsjournal (profil_id, gebucht);

CREATE INDEX handelsjournal_vorgang ON handelsjournal (profil_id, vorgang);

CREATE UNIQUE INDEX druckkennungen_eindeutig
  ON druckkennungen (system, wert) WHERE semantik = 'druck';

CREATE INDEX druckkennungen_wert ON druckkennungen (system, wert);

CREATE TRIGGER freigabenachweise_grenze BEFORE INSERT ON freigabenachweise
BEGIN
  SELECT RAISE(ABORT, 'freigabenachweise: die Grenze ist negativ')
   WHERE NEW.grenze_betrag < 0;
  SELECT RAISE(ABORT, 'freigabenachweise: eine Grenze ohne Währung')
   WHERE NEW.grenze_betrag > 0 AND trim(NEW.grenze_waehrung) = '';
END;

CREATE INDEX deckplaetze_karte             ON deckplaetze (karte_id);

CREATE INDEX decks_format                  ON decks (format_id);

CREATE INDEX ki_feldherkunft_karte         ON ki_feldherkunft (karte_id);

CREATE INDEX reservierungen_nach_exemplar  ON reservierungen (exemplar_id);

CREATE INDEX turnierdeckplaetze_teilnehmer ON turnierdeckplaetze (teilnehmer_id);

CREATE INDEX turnierpaarungen_a            ON turnierpaarungen (a);

CREATE INDEX turnierpaarungen_b            ON turnierpaarungen (b);

CREATE INDEX turnierteilnehmer_deck        ON turnierteilnehmer (deck_id);

CREATE INDEX deckleihen_profil ON deckleihen (profil_id, angelegt DESC);

CREATE INDEX kostenposten_fenster
  ON kostenposten (profil_id, konto, begonnen);

CREATE UNIQUE INDEX kontakte_ursprung ON kontakte (profil_id, ursprung) WHERE ursprung <> '';

CREATE INDEX kontakte_profil ON kontakte (profil_id, name);

CREATE INDEX sammlung_printing ON sammlung (printing);

CREATE INDEX sammlung_lagerort ON sammlung (lagerort);

CREATE INDEX sammlung_druck ON sammlung (druck_id);

CREATE TRIGGER sammlung_regeln_neu BEFORE INSERT ON sammlung
BEGIN
  SELECT RAISE(ABORT, 'sammlung: die Menge ist kleiner als eins')
   WHERE NEW.menge < 1;
  SELECT RAISE(ABORT, 'sammlung: diesen Zustand gibt es nicht')
   WHERE NEW.zustand NOT IN ('mint', 'nearMint', 'excellent', 'good',
                             'lightPlayed', 'played', 'poor');
  SELECT RAISE(ABORT, 'sammlung: diese Ausführung gibt es nicht')
   WHERE NEW.finish NOT IN ('nonfoil', 'foil', 'etched');
  SELECT RAISE(ABORT, 'sammlung: ohne Sprache')
   WHERE trim(NEW.sprache) = '';
END;

CREATE TRIGGER sammlung_regeln_um
BEFORE UPDATE OF menge, zustand, finish, sprache ON sammlung
BEGIN
  SELECT RAISE(ABORT, 'sammlung: die Menge ist kleiner als eins')
   WHERE NEW.menge IS NOT OLD.menge AND NEW.menge < 1;
  SELECT RAISE(ABORT, 'sammlung: diesen Zustand gibt es nicht')
   WHERE NEW.zustand IS NOT OLD.zustand
     AND NEW.zustand NOT IN ('mint', 'nearMint', 'excellent', 'good',
                             'lightPlayed', 'played', 'poor');
  SELECT RAISE(ABORT, 'sammlung: diese Ausführung gibt es nicht')
   WHERE NEW.finish IS NOT OLD.finish
     AND NEW.finish NOT IN ('nonfoil', 'foil', 'etched');
  SELECT RAISE(ABORT, 'sammlung: ohne Sprache')
   WHERE NEW.sprache IS NOT OLD.sprache AND trim(NEW.sprache) = '';
END;

CREATE UNIQUE INDEX sammlung_ohne_druckzeile
  ON sammlung (profil_id, printing, zustand, sprache, finish, wunsch)
  WHERE druck_id IS NULL;

CREATE INDEX turniere_format ON turniere (format_id);

CREATE INDEX turnierprotokoll_turnier ON turnierprotokoll (turnier_id);

CREATE INDEX objektmarken_marke ON objektmarken (profil_id, art, marke);

CREATE TRIGGER objektmarken_set_weg AFTER DELETE ON sets
BEGIN
  DELETE FROM objektmarken WHERE art = 'set' AND kennung = CAST(OLD.id AS TEXT);
END;

CREATE TRIGGER objektmarken_exemplar_weg AFTER DELETE ON sammlung
BEGIN
  DELETE FROM objektmarken WHERE art = 'exemplar' AND kennung = CAST(OLD.id AS TEXT);
END;

CREATE TRIGGER objektmarken_deck_weg AFTER DELETE ON decks
BEGIN
  DELETE FROM objektmarken WHERE art = 'deck' AND kennung = CAST(OLD.id AS TEXT);
END;

CREATE TRIGGER objektmarken_paket_weg AFTER DELETE ON packages
BEGIN
  DELETE FROM objektmarken WHERE art = 'paket' AND kennung = OLD.art || '/' || OLD.name;
END;

CREATE TRIGGER objektmarken_turnier_weg AFTER DELETE ON turniere
BEGIN
  DELETE FROM objektmarken WHERE art = 'turnier' AND kennung = CAST(OLD.id AS TEXT);
END;

CREATE TRIGGER objektmarken_partie_weg AFTER DELETE ON partien
BEGIN
  DELETE FROM objektmarken WHERE art = 'partie' AND kennung = OLD.id;
END;

CREATE TRIGGER objektmarken_handel_weg AFTER DELETE ON handelsfolge
BEGIN
  DELETE FROM objektmarken
   WHERE art = 'handel' AND profil_id = OLD.profil_id AND kennung = OLD.vorgang;
END;

CREATE TRIGGER objektmarken_symbol_weg AFTER DELETE ON symbolbibliothek
BEGIN
  DELETE FROM objektmarken WHERE art = 'symbol' AND kennung = CAST(OLD.id AS TEXT);
END;

CREATE TRIGGER handelsrevisionen_letzte_weg AFTER DELETE ON handelsrevisionen
WHEN NOT EXISTS (
    SELECT 1 FROM handelsrevisionen
     WHERE profil_id = OLD.profil_id AND vorgang = OLD.vorgang
)
BEGIN
    DELETE FROM objektmarken
     WHERE art = 'handel' AND profil_id = OLD.profil_id AND kennung = OLD.vorgang;
    DELETE FROM handelsarchiv
     WHERE profil_id = OLD.profil_id AND vorgang = OLD.vorgang;
END;

CREATE INDEX handelsprotokoll_vorgang ON handelsprotokoll (profil_id, vorgang, id);

CREATE TRIGGER handelsprotokoll_neu
AFTER INSERT ON handelsrevisionen
BEGIN
    INSERT INTO handelsprotokoll (profil_id, vorgang, revision, von, nach)
    VALUES (NEW.profil_id, NEW.vorgang, NEW.revision, '', NEW.zustand);
END;

CREATE TRIGGER handelsprotokoll_wechsel
AFTER UPDATE OF zustand ON handelsrevisionen
WHEN OLD.zustand <> NEW.zustand
BEGIN
    INSERT INTO handelsprotokoll (profil_id, vorgang, revision, von, nach)
    VALUES (NEW.profil_id, NEW.vorgang, NEW.revision, OLD.zustand, NEW.zustand);
END;

CREATE TRIGGER handelsprotokoll_bleibt
BEFORE UPDATE ON handelsprotokoll
BEGIN
    SELECT RAISE(ABORT, 'ein Protokolleintrag wird nicht geändert');
END;

CREATE TRIGGER handelsprotokoll_geht_mit_dem_profil
BEFORE DELETE ON handelsprotokoll
WHEN EXISTS (SELECT 1 FROM profile WHERE id = OLD.profil_id)
BEGIN
    SELECT RAISE(ABORT, 'ein Protokolleintrag wird nicht gelöscht');
END;
