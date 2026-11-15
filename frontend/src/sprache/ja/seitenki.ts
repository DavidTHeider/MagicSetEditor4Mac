// **Die Texte der KI** — Werkstück A6 von Phase 5.3
// (#1445),
// erweitert für A7
// (#1446).
//
// **Ein Katalog darf sich auf mehrere Dateien verteilen, solange er als einer
// ankommt** (§9): `seiteneinstellungen.ts` stand mit ihnen an der
// Zeilengrenze aus §2.
export const ki = {
  kimodelleTitel: 'AIモデル',
  kimodelleKeine:
    'AIモデルが確定していません。プロバイダのモデル一覧は、まだ実証された機能ではありません。ここには、実際に確認されたモデルだけが表示されます。',
  kiNurLokal:
    '「ローカルのみ」が有効です。このMacの外にあるモデルには問い合わせません。',
  kiAuchCloud:
    '外部プロバイダのモデルが許可されています。送信した内容は、このMacの外に出ます。',
  kiBeleg: '{faehigkeit} — {datum}に確認: {beleg}',
  kiKontextfenster: 'コンテキストウィンドウ: {tokens} トークン — {beleg}',
  kiKontextfensterUnbekannt: 'コンテキストウィンドウ: 記載なし',
  kiHerkunftUnbekannt:
    '出所は不明です。ここで実際にどのファイルが動作しているかは確認されていません。',
  kiEinrichtungAus:
    'オンになっていません。何も問い合わせず、何もこのMacの外に出ません。',
  kiEinrichtungUnerreichbar:
    '設定されたアドレスから応答がありません。このアプリケーションは、そこで何も起動せず、何もインストールしません。',
  kiEinrichtungOhneModell:
    '応答はありますが、モデルがありません。このアプリケーションは、モデルをダウンロードしません。',
  kiEinrichtungOhneFaehigkeit:
    'モデルはありますが、ここで役立つ機能が実証されているものはありません。',
  kiEinrichtungBereit:
    '実証された機能を持つモデルが、少なくとも1つ利用できます。',
  kiLaufzeitOllama:
    'ここから応答がない場合、Ollama が起動していません。このアプリケーションは、Ollama の起動もインストールも行いません。それはOllama自身で行います。',
  kiLaufzeitTitel: 'ローカルAI',
  kiLaufzeitFassung: 'ランタイム: {fassung}',
  kiAdresse: '問い合わせ先: {adresse}。',
  kiAdresseVorgabe: 'このMac上の初期設定のアドレス',
  kiCloudAus: 'Ollamaのクラウドへのオフロードはオフになっています。',
  kiCloudAn: 'Ollamaのクラウドへのオフロードはオフになっていません。',
  kiCloudUngewiss:
    'Ollamaのクラウドへのオフロードがオフになっているかどうかは、ここからは確認できません。',
  kiCloudBeleg: '根拠: {beleg}',
  kiVorschlagTitel: 'モデルからの提案',
  kiVorschlagModell: 'モデル',
  kiVorschlagAuftrag: 'タスク',
  kiVorschlagAuftragHinweis: 'モデルに何をさせますか？',
  kiVorschlagen: '提案する',
  kiVorschlagLaeuft:
    'モデルが処理中です。初回はモデルの読み込みがあるため、時間がかかることがあります。',
  kiVorschlagAbbrechen: 'キャンセル',
  kiVorschlagUebernehmen: '適用',
  kiVorschlagNichts:
    'テンプレートが認識するフィールドは、モデルから提案されませんでした。',
  kiVorschlagAendertNichts:
    '提案してもカードは変更されません。「適用」を押したときだけ書き込まれます。',
  kiVorschlagUebernommen: '{felder}個のフィールドを書き込みました。',
  kiVorschlagAblauf: '{ablauf}まで適用できます。',
  // **The second provider** — A8, #1447.
  kiAnbieterZeile: 'プロバイダ {anbieter}: {stand}',
  kiVorschlagAnbieter: 'プロバイダとモデル',
  kiVorschlagModellWaehlen: '選択してください',
  kiVorschlagOhneModell:
    '利用できるモデルがありません。不足している内容は「AIモデル」に表示されます。',
  kiVorschlagAntwortVon: '{anbieter}/{modell} が回答しました。',
  // **Measuring what a model can do** — A9, #1448.
  kiMessbarTitel: '説明のないモデル',
  kiAdresseSocket:
    'このプログラム専用の Unix ソケット経由で使用します（ポートは開きません）。',
  kiGeraetSitzung: 'このセッションで使う',
  kiGeraetSitzungHinweis:
    '次の再起動まで有効です。その後は「Apple のデバイス上モデル」の設定が再び適用されます。',
  kiFeldVorschlagen: '提案',
  kiFeldVorschlagenFuer: '{feld} — 提案',
  kiFeldAuftragName:
    'このカードにふさわしい固有の名前を提案してください。カードの種類やルールの説明ではなく、日本語で名前だけを答えてください。',
  kiFeldAuftragFlavor:
    'このカードの短いフレーバーテキストを 1〜2 文で提案してください。ルール文を繰り返さず、雰囲気や物語を日本語で書いてください。',
  kiFeldKarte: 'これまでのカード：',
  kiFeldUebernehmen: '適用',
  kiFeldVon: '{anbieter} より · {weg}',
  kiFeldWegGewaehlt: '選択',
  kiFeldWegVorrang: '優先',
  kiFeldVorrangFragt: '優先時の問い合わせ先: {reihe}',
  kiFeldWegRueckfall: '代替',
  kiFeldAbgelehnt: 'モデルが拒否しました：{grund}',
  kiFeldNichts: 'このフィールドへの提案はありませんでした。',
  kiMessbarSatz:
    '{anbieter} はこれらのモデルを一覧に挙げていますが、内容については何も説明していません。測定では、固定のスキーマを持つ小さなタスクを送信します。そのタスク1件分の費用がかかります。',
  kiMessen: '測定',
  kiMessungVergessen: '測定結果を破棄',
  kiMessungLaeuft: '測定中…',
  kiMessungGeglueckt:
    '{alias} が応答しました。{faehigkeiten} が実証されました。',
  // **Selection, comparison and provenance** — A11, #1450.
  kiVorschlagAuswahl: '適用: {feld}',
  kiVorschlagBisher: '現在: {wert}',
  kiVorschlagBisherLeer: '現在は空です。',
  kiVorschlagNichtsGewaehlt:
    'チェックされたフィールドがありません。チェックしたものが適用されます。',
  kiHerkunftTitel: '値の出所',
  kiHerkunftZeile: '{feld}: {anbieter}/{modell}、{zeitpunkt}',
  kiHerkunftAufgabe: 'タスク: {aufgabe}',
  kiHerkunftVeraendert: 'その後、手動で編集されています。',
  kiHerkunftKeine: 'このカードには、提案に由来するフィールドがありません。',
  // **The image proposal** — A12, #1451.
  kiBildTitel: 'モデルによる画像',
  kiBildAendertNichts:
    '生成してもカードは変更されません。生成されるのは有効期限付きの提案です。内容を確認し、切り抜いてから、はじめて挿入します。',
  kiBildFeld: '画像フィールド',
  kiBildAufgabe: '何を描きますか？',
  kiBildAufgabeHinweis: '広い平原に落ちる稲妻、油絵',
  kiBildGroesse: 'サイズ',
  kiBildGuete: '品質',
  kiBildBearbeiten: '新しく生成せず、既存の画像を編集する',
  kiBildErzeugen: '画像を生成',
  kiBildLaeuft:
    '{anbieter}/{modell} で画像を作成しています。数秒から数分かかります。',
  kiBildAbbrechen: 'キャンセル',
  kiBildAbgebrochen:
    'キャンセルしました。何も取り込まれていません。それまでにかかった費用は、費用の記録に残っています。',
  kiBildVorschau: '{feld}の提案',
  kiBildMasse: '{breite} × {hoehe} ピクセル、{kilobytes} kB',
  kiBildAblauf: '{ablauf}まで有効です。',
  kiBildUmschrift: '描かれた内容: {umschrift}',
  kiBildNachweisTitel: 'この画像の出所',
  kiBildAngabeGroesse: 'サイズ',
  kiBildAngabeGuete: '品質',
  kiBildAngabeUmschrift: '修正後のプロンプト',
  kiBildAngabeMedientyp: 'ファイル形式',
  kiBildUebernehmen: '切り抜いて挿入',
  kiBildVerwerfen: '破棄',
  kiBildVerworfen: '提案は削除されました。',
  kiBildKosten:
    '画像の依頼は別に承認する必要があります。テキストの承認は、画像には適用されません。',
  // **Approving a paid request** —
  // #1539.
  kiFreigabeTitel: 'この依頼には費用がかかります',
  kiFreigabeSatz:
    '選択したモデルは {anbieter} で実行されます。上限額を承認してください。承認はこの1件の依頼にのみ有効で、5分後に失効します。',
  kiFreigabeBetrag: '上限額 (最小単位。EURの場合はセント)',
  kiFreigabeVerfuegbar: '本日の利用可能額: {betrag}',
  kiFreigabeOhneBudget:
    '{anbieter} には何も承認されていません。承認された予算がない限り、有料の依頼は送信されません。上限は「費用」で設定します。',
  kiFreigabeErteilen: '承認して問い合わせる',
  kiFreigabeVerzichten: '承認しない',
  kiFreigabeLaeuft: '承認を発行しています…',

  // **The evidence-backed assessment** — A13.2a
  // (#1452).
  kiEinschaetzungTitel: '評価を依頼',
  kiEinschaetzungErklaerung:
    'モデルが、ローカルの検索インデックスの該当箇所をもとにこのカードを評価します。カードもコレクションも、何も変更しません。',
  kiEinschaetzungArt: '評価の観点',
  kiEinschaetzungArtRegel: 'ルールに基づく',
  kiEinschaetzungArtBalance: '他との比較',
  kiEinschaetzungFrage: '評価してほしい内容',
  kiEinschaetzungFragen: '評価を取得',
  kiEinschaetzungLaeuft: 'モデルが根拠を読んでいます…',
  kiEinschaetzungLeer: 'まだ何も評価していません。',
  kiEinschaetzungBelegt: '言及された内容はすべて根拠の中にあります。',
  kiEinschaetzungUnbelegt:
    'この回答は、どの根拠にも含まれていない内容に言及しています。その内容は下に表示されます。',
  kiEinschaetzungBefunde: '根拠のない内容',
  kiEinschaetzungEinschraenkungen: '考慮されなかった内容',
  kiEinschaetzungBelege: '渡された該当箇所',
  kiEinschaetzungStaende: '情報源の状態',

  // **The balance analysis** — A13.2b
  // (#1452).
  kiEinschaetzungFormat: 'フォーマット',
  kiEinschaetzungOhneFormat: '指定なし',
  kiEinschaetzungVergleiche: '比較の対象',
  kiEinschaetzungStrategie: '見つけた方法',
  kiEinschaetzungOhneVergleich:
    'この回答には、比較対象のカードが挙げられていません。カードの位置づけに関する記述は、比較そのものです。',
  kiEinschaetzungSynergien: '対象外としたシナジー',

  // **The batch and the atomic group** — A13.1b
  // (#1452).
  kiBatchTitel: '複数のカードをまとめて',
  kiBatchSatz:
    '選択した各カードについて、1つのモデルが同じタスクに回答します。この段階では何も書き込まれません。書き込みは2番目のステップです。',
  kiBatchZiele: '{gesamt}枚中{gewaehlt}枚のカードを選択',
  kiBatchOhneKarten: 'このセットにはまだカードがありません。',
  kiBatchOhneName: '(名前なし)',
  kiBatchAuftrag: '各カードへのタスク',
  kiBatchFahren: '提案を取得',
  kiBatchLaeuft: '実行中…',
  kiBatchLagen: '結果',
  kiBatchGanzOderGar:
    'すべてが一度に取り込まれるか、何も取り込まれないかのどちらかです。1つでも失敗すると、すべてのカードは元のままになり、提案も開いたままになります。',
  kiBatchGruppe: 'グループとして適用 ({karten})',
  kiBatchGruppeLaeuft: '適用しています…',
  kiBatchGeschrieben: '{karten}枚のカードを書き込みました。',

  // **Card recognition from a photo** — A13.3
  // (#1452).
  kiErkennungTitel: '写真からカードを認識',
  kiErkennungSatz:
    'モデルが、写真に印刷されている内容を読み取ります。どのカードかは、入力したリストと同じ版の検索で決まります。コレクションには何も追加されません。',
  kiErkennungLesen: '写真を選択して読み取る',
  kiErkennungLaeuft: 'モデルが写真を読み取っています…',
  kiErkennungLeer: 'まだ写真を読み取っていません。',
  kiErkennungGelesen: '読み取った内容',
  kiErkennungNichtGelesen: '判断できなかった項目',
  kiErkennungUnlesbar: '判読不能',
  kiErkennungOhneWert: '読み取れませんでした',
  kiErkennungArten: {
    name: 'カード名',
    setcode: 'セット',
    nummer: 'コレクター番号',
    typzeile: 'タイプ行',
    layout: 'レイアウト',
    kunstwerk: 'イラスト',
  } as Record<string, string>,
  kiErkennungKandidaten: '一致する版',
  kiErkennungDrucke: '{drucke}個の版が一致します。',
  kiErkennungEinDruck: '一致する版は1つだけです。',
  kiErkennungWeitere:
    'この先にもまだあります。数えたのは最初のページだけです。',
  kiErkennungOhneKandidat:
    '読み取った内容に一致する版がありません。写真に原因があるか、このカードがここに掲載されていない可能性があります。',
  kiErkennungGestuetzt: '根拠: {arten}',
  kiErkennungOhneStuetze:
    'この版について、写真から再確認できる点はありません。検索で該当したため、表示されているだけです。',
  kiErkennungNichtsAngelegt:
    '何も追加されません。1枚を記録するには、上のフォームを使ってください。そうすれば、どの版を指すのかが確定します。',
  kiErkennungStand: '版の検索データの古さ: {alter}',
  kiErkennungModell: '{anbieter} の {modell} が読み取りました',

  // **The citation question** — A13.4
  // (#1452).
  kiVorschlagBelegfrage: 'ナレッジベースの該当箇所を含める (任意)',
  kiVorschlagBelegfrageHinweis: '検索する内容。空欄の場合は検索しません',
} as const
