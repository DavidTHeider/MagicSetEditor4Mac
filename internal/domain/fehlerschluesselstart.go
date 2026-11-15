package domain

// **Die Lagen des Starts** — aus `fehlerschluessel.go` herausgelöst, als F05
// (#2344)
// drei Lagen dazubrachte und die Datei über die Zeilengrenze aus §2 hob. Der
// Schnitt folgt der Frage: hier steht, **warum ein Start scheiterte**.

const (
	// SchluesselBestandUnbrauchbar und SchluesselBestandNeuereFassung sagen,
	// **warum der Start scheiterte** — Werkstück 3 von Phase 4.13
	// (#1352).
	//
	// **Sie gehören nicht zu einem `error`, sondern zu einem Befund.** Der
	// Start gibt keine Bindungsantwort zurück; er hinterlässt einen Zustand,
	// den [app.Startbefund] herausgibt. Gemeinsam ist ihnen der Zweck: das
	// Fenster soll einen Satz zeigen, der sagt, was zu tun ist — und nicht
	// „journal_mode: sqlite3: file is not a database".
	//
	// **Und sie sind nicht dieselben wie die der Sicherung.** Eine Sicherung
	// wählt man aus und kann eine andere nehmen; der **Bestand** ist der eine,
	// mit dem die Anwendung arbeitet. Zwei Lagen, zwei Ratschläge.
	SchluesselBestandUnbrauchbar   Schluessel = "bestandUnbrauchbar"
	SchluesselBestandNeuereFassung Schluessel = "bestandNeuereFassung"

	// SchluesselBestandAeltereFassung ist das Gegenstück für den Bestand:
	// **die Datei ist in Ordnung**, nur kommt diese Fassung nicht mehr an sie
	// heran (Vermerk 0189).
	SchluesselBestandAeltereFassung Schluessel = "bestandAeltereFassung"

	// SchluesselBestandNurLesbar, SchluesselBestandEA und
	// SchluesselStartUngeklaert sind drei weitere Lagen des Starts — F05
	// (#2344).
	// **Keine davon rät zum Ersetzen**: die Datei lässt sich nur lesen, der
	// Datenträger meldete einen Fehler, oder es ist nicht geklärt, was
	// geschah. Belegt, voll und gesperrt nehmen die Schlüssel, die es für
	// diese Lagen schon gibt.
	SchluesselBestandNurLesbar Schluessel = "bestandNurLesbar"
	SchluesselBestandEA        Schluessel = "bestandEA"
	SchluesselStartUngeklaert  Schluessel = "startUngeklaert"

	// SchluesselKaumPlatz ist ein Fehler des Datenträgers bei zugleich kaum
	// Platz — #2407: **beides wird genannt**, denn ohne Beleg in der Kette
	// ist der Fehler kein Beweis für eine volle Platte. Für den Start wie für
	// eine Sicherung.
	SchluesselKaumPlatz Schluessel = "kaumPlatz"
)
