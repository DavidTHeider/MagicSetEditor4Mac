-- **Die Saat eines leeren Bestands** — von Hand, mit denselben Ausdrücken wie
-- die Wanderungen, aus denen sie stammt: Zeit und Zufall lassen sich nicht
-- einfrieren (J02, #2266). Dass sie zu den Wanderungen passt, verglich bis zu
-- deren Wegfall `TestDieGrundlinieGleichtDenWanderungen` (Commit ebd92fa37).

-- 0001_grundlage.sql
INSERT INTO bestand (schluessel, wert)
VALUES ('angelegt_am', strftime('%Y-%m-%dT%H:%M:%fZ', 'now'));

-- 0034_nutzer_und_profile.sql
INSERT INTO nutzer (name) VALUES ('Ich');

INSERT INTO profile (nutzer_id, name)
SELECT id, 'Mein Profil' FROM nutzer WHERE name = 'Ich';

-- 0039_bestandskennung.sql
INSERT INTO bestandskennung (id, wert) VALUES (1, lower(hex(randomblob(16))));

-- 0047_freigabenachweise.sql
INSERT INTO freigabenrevision (id, revision) VALUES (1, 0);
