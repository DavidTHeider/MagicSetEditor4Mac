package domain

// Die gemeinsame Sprache der Suche.
//
// **Hier und nicht in `storage`**, weil `internal/app` sie braucht und
// `internal/app` nichts aus `storage` importieren darf (§2). Derselbe Befund
// hat schon `SetHead` und `CardRow` hierher gebracht: sobald zwei Schichten
// über dieselbe Sache reden, gehört das Wort dafür nach innen.

// SearchQuery ist eine Anfrage an den Bestand.
//
// **Die Filter sind kombinierbar und alle freiwillig.** Wer nur `Text` setzt,
// sucht überall; wer `Field` dazu setzt, sucht in einem Feld; wer `SetID`
// setzt, sucht in einem Set. Jede Kombination ist zulässig, weil jede sich
// jemand wünscht. `SetID` 0 heißt: über alle Sets.
type SearchQuery struct {
	Text  string
	Field string
	SetID int64
}

// SearchHit ist eine gefundene Stelle: eine Karte, ein Feld, sein Wert.
type SearchHit struct {
	CardID int64
	SetID  int64
	Number int
	Field  string
	Value  string
}

// SearchLimit begrenzt, wie viele Treffer eine Suche zurückgibt.
//
// **Eine Schranke, die man sieht.** Wer „a" sucht, trifft bei 3.000 Karten
// alles; eine Liste, die dann zehn Sekunden baut, wirkt kaputt. Dass gekürzt
// wurde, steht in `SearchResult.Truncated` — eine stille Kürzung sähe aus wie
// ein vollständiges Ergebnis.
const SearchLimit = 200

// SearchResult ist, was eine Suche zurückgibt.
//
// **Mit der Kürzung im Ergebnis.** Der erste Anlauf gab nur die Treffer und
// ließ den Aufrufer über eine Schranke im Speicher nachrechnen, ob gekürzt
// wurde — eine Zusage, die man vergessen kann, und dann sieht eine gekürzte
// Liste wie ein vollständiges Ergebnis aus.
type SearchResult struct {
	Hits      []SearchHit
	Truncated bool
}
