// The texts for **database maintenance** — back up and restore
// (#940).
//
// Split out of `seiten.ts` the same way the card page was: one catalogue may
// live in several files as long as it arrives as **one** (§9).

export const bestandspflege = {
  titel: 'データベース',
  hinweis:
    'このアプリケーションが管理するすべてです：セット、カード、フィールドの値、履歴、配置した画像。バックアップには、データベースと画像の両方が1つのファイルにまとめて含まれます。',
  sichern: 'データベースをバックアップ…',
  waehlen: 'バックアップを復元…',
  rueckfrage: 'このバックアップを復元しますか？',
  datei: 'ファイル',
  angelegtAm: 'データベースの作成日',
  sets: 'セット',
  karten: 'カード',
  unbekannt: '—',
  warnung:
    '現在のデータベースは置き換えられます。失われることはありません。次に復元するまで、bestand-vor-rueckspielung.db という名前で横に残ります。',
  bilderLage: '配置した画像',
  bilderErsetzt:
    'このバックアップには画像が含まれています。現在の画像の代わりに使われ、これまでの画像は beiwerk-vorher という名前で横に残ります。',
  bilderGeleert:
    'このバックアップには画像が含まれていません。画像が配置されていなかった時点のものです。現在の画像は取り除かれ、beiwerk-vorher という名前で横に残ります。',
  bilderUnberuehrt:
    'このバックアップは 0.1.0 または 0.2.0 の単独のデータベースで、画像は含まれていません。現在の画像はそのまま残ります。別のコンピュータでは、画像はまったく存在しないことになります。',
  bestaetigen: '復元',
  abbrechen: 'キャンセル',

  // **The password of a backup** (OPS-06, #1723).
  kennwort: 'パスワード（任意）',
  kennwortWiederholen: 'パスワードを再入力',
  kennwortHinweis:
    'パスワードを設定するとバックアップが暗号化され、別のコンピュータでも開けます。忘れたパスワードは復旧できません。その場合、バックアップは失われます。',
  kennwortVerschieden: '2つのパスワードが一致しません。',
  klartext: 'パスワードなしでバックアップ — ファイルは平文になります',
  klartextHinweis:
    'コレクションは暗号化されています。パスワードなしでは、バックアップが保存先のメディア上で保護されないままになります。',
  gesperrt: 'このバックアップは暗号化されています。',
  gesperrtKennwort: 'バックアップのパスワード',
  entsperren: '開く',
  verschluesselt: '暗号化',
  ja: 'はい',
  nein: 'いいえ',
  paketeEnthalten: '含まれる独自パッケージ',
  paketeFehlen: '再取得が必要',
  keine: 'なし',
  zurueckgespielt:
    '復元しました。それまでのデータベースは {beiseite} として残っています。',
  rueckkehrUnveraendert: 'これまでのデータベースは変更されていません。',
  rueckkehrUnvollstaendig:
    '元に戻す処理も失敗しました。以前のデータベースの一部が bestand-vor-rueckspielung.db と beiwerk-vorher として隣に残っています。バックアップを復元するか、手動で元の場所に戻してください。',
  wiederherstellungTitel: 'データベースを開けませんでした',
  wiederherstellungHinweis:
    '使用できるデータベースファイルが用意されるまで、アプリケーションは動作しません。バックアップの復元がそのための方法です。現在うまく動作していないファイルを置き換えます。',
  wiederherstellungGrund: '報告された内容',
  wiederherstellungHinweisHeil:
    'データファイル自体はおそらく問題ありません。原因はその周りにあります。下の説明どおりに対処してから、もう一度試してください。置き換える必要はありません。',
  erneutVersuchen: 'もう一度試す',
  weiterGescheitert: 'もう一度試しましたが、データはまだ開けません。',

  wegZurueckTitel: '元に戻す方法',
  wegZurueck:
    '元に戻すには、ライブラリのフォルダにある「{datei}」を使います（{sets} セット、{karten} 枚のカード、スキーマ段階 {stand}）。下で選択してください。',

  chiffreTitel: 'ライブラリの暗号化',
  chiffreAnHinweis:
    'ライブラリはディスク上に暗号化されて保存されています。鍵がなければ、あなた自身でも開けません。',
  chiffreAusHinweis:
    'ライブラリは平文のまま保存されています。このファイルを持っている人は、氏名、生年月日、住所、写真、サインイン方法など、中のすべてを読むことができます。',
  chiffreAnschalten: '暗号化をオンにする…',
  chiffreAusschalten: '暗号化をオフにする…',
  chiffreFrageAn: 'ライブラリを今すぐ暗号化しますか？',
  chiffreFrageAus: '暗号化を再びオフにしますか？',
  chiffreWarnungAn:
    '鍵がここで作成され、この Mac のキーチェーンに保存されます。鍵は再計算できず、他のどこにも存在しません。鍵が失われると、データも失われます。先にバックアップを作成してください。',
  chiffreWarnungAus:
    'すべてが再び平文で保存され、鍵はキーチェーンから削除されます。',
  chiffreGrenzen:
    '暗号化は改ざんからは保護**しません**。同じファイルの複数のコピーを比較すると、変更された箇所がわかります。**バックアップには専用のパスワードを設定します**。バックアップ時に上で設定してください。パスワードなしでは平文になるため、その場合は「パスワードなしでバックアップ」を明示的に選ぶ必要があります。',
  chiffreLaeuft: 'ライブラリを書き換えています…',

  platzUeberschrift: 'ディスクの空き容量',
  platzErklaerung:
    'このアプリケーションが保存しているものと、その容量です。項目に属さないものは**その他**に数えられるため、フォルダ全体が対象です。**移行前のバックアップ**は、ライブラリを作り直す起動のたびに作成されます。残るのは**元に戻す方法**、つまりこのバージョンで開ける最新のバックアップです。**ログ**は、使用中のファイルが残ります。',
  platzDateien: '{anzahl} 個のファイル',
  platzFreigeben: '古いものを削除',
  platzFreigebenFuer: '{posten}：古いものを削除',
  platzFreigebenFrage: '{posten}：古いものを本当に削除しますか？',
  platzFreigebenErklaerung:
    '古いファイルは削除され、元に戻せません。元に戻す方法として必要なものは残ります。',
  platzLoeschenJa: '削除',
  platzNamen: {
    bestand: 'ライブラリ',
    bilder: '埋め込まれた画像',
    pakete: 'ダウンロードしたパッケージ',
    vorsicherungen: '移行前のバックアップ',
    protokoll: 'ログ',
    sonstiges: 'アプリケーションのフォルダ内のその他のファイル',
  },
  alter: {
    ueberschrift: '最後のバックアップ',
    unbekannt: '不明',
    unbekanntSatz:
      'このバージョンではまだ一度もバックアップしていません。不明は不明です——ファイルを読み戻さない限り、ここに「今日」と表示されることはありません。',
    heute: '今日',
    gestern: '1日前',
    vorTagen: '{tage}日前',
    alt: '{alter}・14日より前',
    unstimmig: '時計より後の時刻',
    satz: '最後のバックアップ：{zeitpunkt}、{flaeche}から{datei}へ。読み戻して検査済みです。',
    satzPlan:
      '最後のバックアップ：{zeitpunkt}、プランにより{datei}へ。読み戻して検査済みです。',
    unstimmigSatz:
      '最後のバックアップの日時は{zeitpunkt}で、このMacの時計より後です。経過は分かりません。念のためバックアップしてください。',
    flaechen: {
      fenster: 'ウインドウ',
      cli: 'コマンドライン',
      mcp: 'MCP',
      rest: 'サービス',
    },
    planSchalter: '毎週フォルダにバックアップし、7つ残す',
    planHilfe:
      'フォルダを選んだ後だけ動きます。削除するのは自分のバックアップだけで、他のファイルは削除しません。ウインドウが開いている間にバックアップします。',
    planOrdner: 'フォルダ：{ordner}',
    planZuletzt: 'プランによる最後のバックアップ：{zeitpunkt}。',
    planNie: 'プランによるバックアップはまだありません。',
    planFehler: 'プランは前回バックアップできませんでした：{grund}',
  },
} as const
