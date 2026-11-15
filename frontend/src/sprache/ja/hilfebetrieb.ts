// **Operation and trust** — OPS-07
// (#1724).
export const betrieb = {
  ueberschrift: '運用と信頼性',
  einleitung:
    '何がインストールされ、どこにあり、何が検証され、何が検証されていないか。このアプリケーション自身から読み取っています。',
  fassungName: 'バージョン',
  fassungSatz: '{fassung}、{stand}からビルド（{gebaut}時点の状態）。',
  fassungVeraendert: 'ビルドには、コミットされていない変更が含まれていました。',
  datenName: 'データ',
  datenSatz: '適用済みのマイグレーションは{zahl}件で、最新は{letzte}です。',
  datenOhne: 'データが開かれていません。',
  signaturName: '署名',
  signaturSatz:
    'アドホック署名済みで、**公証されていません**。初回起動時、macOSはアプリケーションを拒否します。システム設定 › プライバシーとセキュリティ › **このまま開く**で、一度だけ許可してください。',
  aktualisierungName: 'アップデート',
  aktualisierung: {
    aus: 'オフ — このアプリケーションは新しいバージョンを確認しません。',
    melden: 'GitHubの新しいバージョンを通知しますが、何も取得しません。',
    holen: 'GitHubから新しいバージョンを取得し、入れ替える前に確認します。',
  },
  orteName: 'すべての保存場所',
  orte: {
    bestand: 'データ',
    pakete: 'テンプレートパッケージ',
    bilder: '挿入した画像',
    zwischenspeicher: 'キャッシュ',
    protokoll: 'ログ',
    leistungsprofile: 'パフォーマンスプロファイル',
  },
  lizenzenName: 'ライセンス',
  lizenzenSatz:
    'ソースコードはGNU General Public License 2.0以降のもとにあります。第三者の部品とそのライセンスは、ソースコードのそばにある`DRITTANBIETER.md`に記載されています。テンプレートパッケージには独自の条件があり、AIモデルにはその提供元の条件があります。',
  empfaengerName: 'データの送信先',
  empfaengerSatz:
    'オンにしたサービスにのみ送信されます。どのサービスで何が送信されるかは、**ヘルプ › このアプリケーションがあなたについて知っていること**に記載されています。利用状況データは送信されません。',
  grenzenName: 'セキュリティ上の境界',
  grenzenSatz:
    'サービスとツールサーバは、ほかに設定しない限り`127.0.0.1`のみで待ち受け、トークンを要求します。**設定 › 診断**でネットワーク上のアドレスを指定すると、TLSでのみ動作し、このコンピュータの外に出ることがページに表示されます。stdioでは、ツールサーバは専用のストリームで通信します。データは**設定 › データとバックアップ**で暗号化できます。',
  kontaktName: 'うまく動かないとき',
  kontaktSatz:
    '**設定 › 診断**に、診断結果、診断バンドル、このプロセスの数値があります。不具合の報告は github.com/DavidTHeider/MagicSetEditor4Mac へお寄せください。自動で外部に送信されるものはありません。',
  pruefstandName: '検証済みの内容',
  pruefstandSatz:
    '検証済みとは、実行によって確認されたことを意味します。コミットを明記した証拠がソースコードとともにあるプローブか、すべてのパイプラインが実行するテストです。合格したものだけがリリースされます。実装のみとは、ビルド済みですが、オリジナルやプラットフォームに対して一度も実行されていないことを意味します。',
  spalteBereich: '領域',
  spalteGegenstand: '対象',
  spalteStand: '状態',
  spalteBeleg: '根拠',
  belegIssue: 'イシュー #{nummer}',
  bereiche: {
    plattform: 'プラットフォーム',
    sprache: '言語',
    barrierefreiheit: 'アクセシビリティ',
    regeln: 'ルール',
    integration: '外部サービス',
  },
  staende: {
    geprueft: '検証済み',
    vorhanden: '実装のみ',
    beschrieben: '記述のみ',
    blockiert: '人の対応待ち',
  },
  gegenstaende: {
    'macOS 15.3 bis 26': 'macOS 15.3～26',
    'macOS 15.2 auf Apple Silicon': 'Apple シリコン搭載の macOS 15.2',
    'macOS 27 auf Apple Silicon': 'Apple シリコン搭載の macOS 27',
    'Tastatur und Vorleser im WebKit':
      'WebKitでのキーボードとスクリーンリーダー',
  },
}
