// The **going back** texts — work item O7 of phase 5.2
// (#1431).

export const ruecksprung = {
  titel: '元に戻す',
  hinweis:
    '元に戻せるものは4つあり、必要な手間はそれぞれ異なります。そのうち2つは、このアプリケーション自身ではできません。その場合は、だれが行うかをここに示します。',

  arten: {
    appfassung: 'アプリケーション',
    schema: 'データの構造',
    daten: 'データ',
    pakete: 'テンプレートパッケージ',
  } as Record<string, string>,

  hindernisse: {
    nachProzessende:
      'このアプリケーションが終了した後、人がFinderでバンドルを入れ替えます。実行中に自分自身を置き換えるプログラムには、「アプリケーション」への書き込み権限を持つ2つ目のプロセスが必要になります。',
    nurVorwaerts:
      'マイグレーションを逆方向にたどる方法はありません。元に戻すにはバックアップを使います。当時の状態を復元してください。',
    unbekannteArt: 'このアプリケーションは、この種類を認識できません。',
  } as Record<string, string>,

  gehtHier: 'このアプリケーション自身でできます。',

  vergleichTitel: '復元にかかるコスト',
  datenstand: '復元後の状態',
  spalteGegenstand: '項目',
  spalteSicherung: 'バックアップ内',
  spalteJetzt: '現在',
  spalteVerlust: '失われるもの',

  gegenstaende: {
    sets: 'セット',
    karten: 'カード',
    decks: 'デッキ',
    exemplare: '現物',
  } as Record<string, string>,

  unbekannt: '—',

  vorherAusfahren:
    'ここで失われるものは、事前に書き出せます。セットは「別名で保存」から、コレクションは「書き出し」から行えます。',

  unklarTitel: '復元後に不明になるもの',
  unklarHinweis:
    'これらの現物は、復元後にコレクションに戻ります。まだ手元にあるかどうかはだれにもわかりません。その後に手放した可能性があります。確認するまでトレードの対象から外れます。何も差し引かれません。',
  unklarZeile: '{menge} × {name}',
  ohneNamen: '{menge} × {schluessel}',

  handpruefung:
    'コレクションの比較を取得できませんでした。このバックアップかお使いのデータが、その情報を提供していません。復元後は、コレクションを手動で確認してください。',

  rueckkehrenTitel: '記録された復元',
  keineRueckkehren:
    'まだありません。バックアップを復元すると、ここに表示されます。戻った先の状態と、不明のまま残ったものも記録されます。',
  rueckkehrZeile: '{zurueckgespielt}に{datenstand}の状態へ復元',
  ausDatei: '{quelle}から',
  offen: '未確認',
  quittiert: '{geklaertAm}に確認済み',
  betrifft: '{menge}枚が対象',
  quittieren: '確認済みにする',
  quittierenLaeuft: '記録中…',
  quittierenHinweis:
    '影響を受ける現物を手動で確認したことを記録します。何も差し引かれません。',
}
