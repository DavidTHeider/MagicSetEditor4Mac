package domain

// **Der Schlüssel einer Absage über die Leitung** — P11
// (#2379).
// Eine eigene Datei, weil `fehlerschluessel.go` an der Zeilengrenze aus §2
// steht; angemeldet wird er mit allen übrigen in `fehlerschluesselwege.go`.

// SchluesselNichtFreigegeben heißt: das Gegenüber gibt diesem Profil zum
// Handeln nichts frei. **Nicht der Satz zu [ErrErlaubnis]**: der schickt zu
// einem Verwalter dieses Profils, und hier fehlt keine Rolle — gefragt ist
// das Gegenüber.
const SchluesselNichtFreigegeben Schluessel = "nichtFreigegeben"

// SchluesselZeileNichtVerbuchbar heißt: eine Zeile des Tauschs ließe sich auf
// der Seite, die sie bekommt, nicht verbuchen — es fehlt etwa die Sprache
// (#2495).
// `{zeile}` nennt sie.
const SchluesselZeileNichtVerbuchbar Schluessel = "zeileNichtVerbuchbar"

// SchluesselZusageUnpassend heißt: dieses Exemplar erfüllt keine Zeile, die
// diese Seite weggibt — Zustand, Sprache oder Ausführung weichen ab
// (#2492).
const SchluesselZusageUnpassend Schluessel = "zusageUnpassend"
