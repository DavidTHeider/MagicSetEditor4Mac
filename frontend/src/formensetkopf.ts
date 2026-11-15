// **Die Formatangaben eines Sets** — abgeteilt von `formen.ts`, damit beide
// unter der Zeilengrenze aus §2 bleiben. Der Schnitt folgt der Frage: dort
// stehen die Gestalten, die aus **Karten und Sets** kommen, hier die eine, die
// aus der **Datei** kommt.
// SetkopfDTO sind die fünf Formatangaben eines Sets
// (#1325).
//
// **Fünf benannte Felder und keine Liste**: jede hat eine eigene Bedeutung und
// eine eigene Aufschrift. Was sie bedeuten, steht in
// `internal/domain/setkopf.go` — nachgelesen in MSE2.
//
// **Leer heißt: stand nicht in der Datei.** Ein Set, das in dieser Anwendung
// entsteht, hat keine davon.
export interface SetkopfDTO {
  mseFassung: string
  spielfassung: string
  stilfassung: string
  versionskontrolle: string
  apprenticecode: string
}
