-- Ein Gegenbeleg, der nicht passt, wird festgehalten — #2500.
--
-- Bis hierher endete die Zustellung eines nicht passenden Gegenbelegs mit
-- einem Fehler, und keine Seite hielt etwas fest: beim Absender stand weiter
-- „hinausgegangen", beim Empfänger nichts. Jetzt steht am eigenen Beleg der
-- jüngste Befund — von wem, welchen Abdruck der Beleg nannte, woran er
-- scheiterte (`domain.Gegenbelegmangel`) und der Satz der Prüfung.
--
-- **Einer je Vorgang** und nicht einer je Anlauf: ein Gegenüber, das immer
-- neue Belege schickt, füllt den Bestand nicht, und der jüngste ersetzt den
-- vorigen. **Der Fremdschlüssel auf den eigenen Beleg** sagt, wogegen geprüft
-- wurde — ohne ihn gibt es keinen Befund, und mit ihm geht er.
CREATE TABLE handelsgegenbelegbefunde (
    profil_id INTEGER NOT NULL,
    vorgang   TEXT    NOT NULL,

    -- von ist die Peerkennung aus der Leitung, wie bei den Fremdbelegen.
    von       TEXT    NOT NULL,

    -- abdruck ist der, den der abgewiesene Beleg nannte — eine Behauptung.
    abdruck   TEXT    NOT NULL,

    mangel    TEXT    NOT NULL,
    grund     TEXT    NOT NULL,
    empfangen TEXT    NOT NULL DEFAULT (datetime('now')),

    PRIMARY KEY (profil_id, vorgang),
    FOREIGN KEY (profil_id, vorgang)
        REFERENCES handelsbelege (profil_id, vorgang) ON DELETE CASCADE,

    CHECK (von <> ''),
    CHECK (grund <> ''),
    CHECK (mangel IN ('unterschrift', 'vorgang', 'revision', 'parteien', 'zeilen'))
) STRICT, WITHOUT ROWID;
