// **The network settings** — work item N5 of phase 5.4
// (#1571).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge — dieselbe Gestalt wie `einstellungenzugaenge.ts`
// daneben. **Ein Katalog bleibt es trotzdem**: die drei Teile werden dort
// hineingestreut, und der Vertragstest gegen die Go-Seite sieht einen.

// netznamen sind die Beschriftungen der sechs Einstellungen.
export const netznamen = {
  'netz.proxy.http': 'httpのプロキシ',
  'netz.proxy.https': 'httpsのプロキシ',
  'netz.proxy.aus': 'プロキシを使わずに接続',
  'netz.frist.verbindung': 'タイムアウト: 接続 (秒)',
  'netz.frist.kopf': 'タイムアウト: 応答ヘッダ (秒)',
  'netz.frist.stillstand': 'タイムアウト: 停滞 (秒)',
  'netz.schnittstelle': 'ローカルネットワーク検索のインターフェース',
  'netz.lauschen': '自分のサービスが待ち受ける場所',
  'dienst.tls.zertifikat': 'サービス独自の証明書',
  'dienst.tls.schluessel': '独自の証明書の鍵',
  'dienst.tls.kette': '中間証明書',
  'dienst.tls.anker': '証明書のトラストアンカー',
  'grenzen.kontingente': 'サービスの利用枠',
}

// netzhilfe ist der erklärende Satz je Einstellung.
export const netzhilfe = {
  'netz.proxy.http':
    '`http://` のアドレスが経由するプロキシです。**空欄の場合は、環境の `HTTP_PROXY` が適用されます。** ここに入力すると、そちらより優先されます。スキーム付きで入力してください (例: `http://proxy.company:3128`)。**パスワードは含めないでください**。暗号化されないままディスクに保存されてしまうためです。認証が必要なプロキシは、環境の `HTTPS_PROXY` で指定します。',
  'netz.proxy.https':
    '`https://` のアドレスについて同様の設定で、`HTTPS_PROXY` に相当します。多くのネットワークではどちらか一方にしかプロキシがないため、**別々に**設定します。',
  'netz.proxy.aus':
    'プロキシを**経由しない**接続先です。`NO_PROXY` と同様に、ホスト名とアドレス範囲をカンマで区切って入力します。ループバックアドレスは常に対象外です。同じネットワーク内の**別の**マシンで動作している Ollama は、ここに入力してください。',
  'netz.frist.verbindung':
    'TLS を含む接続の確立に許される時間です。空欄の場合は15秒です。',
  'netz.frist.kopf':
    'サーバが応答ヘッダを送信するまでに許される時間です。空欄の場合は30秒です。',
  'netz.frist.stillstand':
    '進行中の応答が停滞したまま許される時間です。低速な回線で大きなダウンロードが途中で切れる場合は、上の2つではなく**この値を大きくします**。空欄の場合は60秒です。',
  'netz.schnittstelle':
    'ローカルネットワークの検索がどのネットワークインターフェースで問い合わせ、アナウンスがどれを通って出ていくかを指定します（例: `en0`）。**空欄の場合は自動**で、IPv4とマルチキャストに対応する適切なインターフェースがすべて使われます。現在存在しない名前は空欄と同じ扱いになり、検出事項としてログに記録されます。',
  'netz.lauschen':
    'REST と MCP が待ち受けるアドレス（カンマ区切り）。**空はこの Mac のみ**（`127.0.0.1`）。`0.0.0.0` はすべてのインターフェースを意味し、単独で指定します。**アドレスがこのコンピュータの外に出る場合は TLS とトークンが必要です**。「ウインドウのないツール」 で選ぶ方が簡単です。次にサービスを起動したときに有効になります。',
  'dienst.tls.zertifikat':
    'RESTサービスとMCPサービスがループバックアドレス以外に対して提示する証明書です（PEM形式）。**空欄の場合は、このプロファイルの身元に基づく証明書が使われます。**独自の認証局を使うと、クライアントはフィンガープリントではなく、名前と証明書チェーンで検証します。サービスの次回起動時から有効になります。',
  'dienst.tls.schluessel':
    '証明書の秘密鍵です（PEM形式）。**自分だけが読み取れる状態（0600）**にしてください。そうでない場合、サービスは起動しません。両方のファイルを入れ替えた場合は、再起動しなくても次回のハンドシェイクから有効になります。',
  'dienst.tls.kette':
    '証明書自体に含まれていない場合に指定する中間証明書です。空欄の場合は、なしとなります。',
  'dienst.tls.anker':
    '起動前に証明書チェーンを検証する際の基準となる、独自の認証局のルート証明書です。空欄の場合は、システムのトラストアンカーが使われます。',
  'grenzen.kontingente':
    'REST と MCP を通じて、機械からのアクセスが1分あたりにできる量です。アクセスごとと全体の両方に適用されます。**オフは管理された環境向けです**。負荷テストのように、誰も他の誰かの取り分を奪わない場合に使います。アクセス制御、TLS、作業ごとの同時実行枠、リクエスト本文と応答の上限、タイムアウト、推測に対するロックは残ります。次にサービスを起動したときから有効になります。',
}

// netzbereich ist die Überschrift des Bereichs samt Hinweis.
export const netzbereich = {
  netz: 'このプログラムの外部への接続',
  netzHinweis:
    'この設定は、コマンドライン、MCPサーバ、サービスを含む**すべて**のサービスに同時に適用されます。空欄の場合は、環境の設定が適用されます。現在何が適用されていて、それがどこから来ているかは、「ネットワークと到達性」に表示されます。',
}

// netzwerte sind die Namen der Werte — QA-08.AK-03 (#1739).
export const netzwerte = {
  'grenzen.kontingente.an': 'オン',
  'grenzen.kontingente.aus': 'オフ',
}
