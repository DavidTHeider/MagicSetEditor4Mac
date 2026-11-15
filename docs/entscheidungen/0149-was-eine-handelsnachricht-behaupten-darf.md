# 0149 — Was eine Handelsnachricht behaupten darf, und ein Ausweg nach der Einigung

**Stand:** entschieden am 2026-09-30 · TRADE-03.AK-01 bis AK-04, NAT-01
(#1717)

## Die Frage

TRADE-03 verlangt, dass Abschluss, Übertragung und Verbuchung auch bei
Unterbrechungen eindeutig bleiben. Eine Kartierung des Bestands aus Phase
5.6 — nachgelesen, nicht übernommen — fand fünf Stellen, an denen das nicht
galt. Dieser Vermerk hält fest, was daraus als Regel folgt.

## Was MSE2 tut

**Es handelt nicht** (gesucht nach `trade`, `offer`, `exchange` in
`/Users/david/mse2-quelle/src`, Fassung f9d9356 vom 2021-07-06). Marke
**new**; die Zustandsmatrix steht in `internal/domain/handelsmatrix.go`.

## Was gilt

1. **Eine empfangene Revision behauptet keinen Zustand und keine fremde
   Stimme.** Zustand und Stimmen stehen nicht im Abdruck; `Nimmt` prüfte sie
   deshalb nicht, und wer eine erste Revision als `lokalVerbucht` oder einen
   Vorschlag mit beiden Stimmen schickte, stand hier dort. Jetzt: eine erste
   Revision kommt vom Initiator, steht auf `angeboten` und trägt höchstens
   seine Stimme; ein Vorschlag ist, was `Fortgeschrieben` baut — `verhandlung`
   und höchstens die Stimme des Absenders (`PruefeErsteRevision`,
   `NimmtVon`). **Und eine erste Revision nennt dieses Profil als Partei**
   — sonst legte ein berechtigter Kontakt einen Vorgang zwischen sich und
   einem Dritten in einem fremden Bestand an (Befund Z3 des Zweitblicks über
   Phase 6.8, `Handelspost.MitAusweis`).
2. **Eine Stimme gehört zu ihrem Beschluss.** Eine Zustimmung zum Angebot,
   die erst ankam, als hier schon vorbereitet war, zählte als Stimme zur
   Übergabe. Stimmen zählen nur, wenn drüben derselbe Zustand steht; sonst
   zählt allein der Schritt dorthin.
3. **Eine Urkunde wird bei jedem Anlauf gegengezeichnet.** Kam sie vor der
   eigenen Verbuchung, stand der Vermerk der Zustellung schon, und „bekannt"
   hieß danach, dass `beidseitig` nie wahr wurde. Die Gegenzeichnung ist
   wiederholbar und läuft auch bei einer bekannten Nachricht.
4. **Nur ein Tausch bewegt Bestand** — der Kopf des Vertrags sagte es,
   durchgesetzt hat es niemand. `Verbuchen` weist Teilen, Deckleihe und
   Markt ab.
5. **Eine ungeteilte Bewertung bleibt hier.** Sie steht dann nicht im
   Abdruck und ging trotzdem hinaus.
6. **Nach der Einigung gibt es einen Ausweg für jede Partei**: `abbrechen`
   aus `angenommen` und `vorbereitet` (die Zusagen werden frei), `klaeren`
   aus `uebergabeOffen` — dort können die Karten schon unterwegs sein, und
   ein Abbruch behauptete, es sei nichts geschehen. Die Invariante dazu:
   **aus jedem Zustand außer dem Entwurf führt ein Ausweg, den jede Partei
   allein gehen kann**. `TestJederZustandStehtInDerMatrix` hielt sie nicht —
   ihm genügte ein Übergang, der beide braucht.
7. **In der Oberfläche** steht die Art des Handels, ein Teilen oder eine
   Deckleihe sagt, dass sie nichts bewegt, eine Einigung heißt
   **Vereinbarung**, und „Zugestimmt haben" steht nur neben dem Beschluss,
   dem die Stimmen gelten (`beschluss` am Vorgang, gerechnet).

## Verworfen

- **Empfangene Revisionen still normalisieren** (Zustand und Stimmen
  überschreiben): der Abdruck bliebe gleich, aber der Absender hätte eine
  andere Lesart als der Empfänger. Abweisen sagt, dass etwas nicht stimmt.
- **Einen Abbruch aus `uebergabeOffen`**: nach begonnener Übergabe ist das
  kein Rückzug, sondern eine Frage an Menschen.

## Was offen bleibt (Schnitt 2)

- **Ein Protokoll der Schritte**: wer wann zugestimmt hat, steht nur bis zum
  nächsten Wechsel in der Revision; eine eigene, nur anhängende Tafel fehlt.
- **Belege gegen Überschreiben schützen** (Auslöser), und ein abweichender
  zweiter Gegenbeleg als Konflikt statt still verworfen.
- **Zwei Schreiber an derselben Revision** (Fenster und Leitung) ohne
  erwarteten Stand.
