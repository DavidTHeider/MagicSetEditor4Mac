// Die Texte der Technikseite über **Dienste und Wege** — welches Zeugnis
// ein Dienst vorzeigt (NET-03, #1707) und wie ein Gegenüber über fremde
// Netze angebunden ist (NET-02, #1706). Aus `seitentechnik.ts` herausgelöst,
// weil die Datei an der Zeilengrenze aus §2 stand.

export const dienstzeugnis = {
  eigenes: '{aussteller}発行の独自の証明書（{bis}まで有効）',
  identitaet:
    'このプロファイルの身元に基づく証明書です。接続する側はフィンガープリントを照合します。',
}

export const fernwege = {
  titel: '相手の接続方法',
  keine: 'この招待を使った相手はまだいません。',
  direkt: '直接接続、{adresse} 経由',
  relais: 'リレー {adresse} 経由 — まだ直接の経路はありません',
  aufbau: '既知ですが、まだどの経路でも接続できません',
  unbestimmt: '接続中: {anzahl} — 経路はまだ分かりません。',
}

// **Command line and tool server with one click** (OPS-09, #1907).
export const schnittstellen = {
  titel: 'コマンドラインとツールサーバを設定する',
  satz: 'ボタンを1つ押すと、この Mac 上のプログラムがコマンドラインやツールサーバを見つけるために必要な内容を書き込みます。もう1つのボタンで、書き込んだ内容だけを取り除けます。事前に、どのファイルのどの行が対象かが表示されます。他者による記述はそのまま残ります。',
  namen: {
    befehle: '~/.local/bin にある mse と mse-mcp へのリンク',
    befehlszeile: 'ログインシェルの PATH にあるコマンドライン',
    codex: 'Codex 用のツールサーバ',
    claudedesktop: 'Claude Desktop',
    claudecode: 'Claude Code',
    dienst: 'LaunchAgent としてのスクリプト用サービス',
  },
  staende: {
    offen: '未設定',
    eingerichtet: '設定済み',
    veraltet: '古い',
    fremd: '他者の記述',
    nichtGefunden: 'この Mac には見つかりません',
    unnoetig: '不要',
  },
  staendeSaetze: {
    veraltet:
      'アプリケーションの場所が変わりました。「設定する」で修復できます。',
    fremd: '他者による記述がすでにあります — そのまま残ります。',
    unnoetig: 'すでに PATH に含まれています。',
  },
  zeilen: '次の行が追加されます',
  vorschau: {
    json: 'このエントリがファイルに追加されます',
    befehl: 'このコマンドが実行されます',
    agent: 'このファイルが作成され、このコマンドで読み込まれます',
  },
  nurEintrag:
    '{name} はその後変更されていました。取り除いたのはこのアプリのエントリだけで、それ以外は残したままです。',
  einrichten: '設定する',
  widerrufen: '取り消す',
  alle: {
    titel: 'すべてまとめて',
    satz: 'この Mac にあり、まだ設定されていない対象をすべて設定します。または、設定済みのものをすべて元に戻します。他者の項目はそのまま残ります。',
    vorschau: '書き込まれる内容（{anzahl} 件）',
    vorschauLeer: '不足はありません。ここにあるものは設定済みです。',
    einrichten: 'すべて設定',
    widerrufen: 'すべて元に戻す',
    ergebnis: '結果',
    ausgaenge: {
      eingerichtet: '設定済み',
      widerrufen: '元に戻しました',
      fremd: '他者の項目（変更なし）',
      unveraendert: '変更なし',
      gescheitert: '失敗',
      abgebrochen: '中止',
    },
  },
  lauschwahl: {
    titel: 'サービスの待ち受け先',
    satz: 'スクリプト用サービス（ウインドウから起動したものと LaunchAgent）と HTTP 経由のツールサーバに、次回の起動から適用されます。',
    wahl: '到達経路',
    schleife: 'この Mac のみ（127.0.0.1）',
    auswahl: '選択したアドレス',
    alle: 'すべてのインターフェース（0.0.0.0）',
    alleSatz: '後から追加されるものも含め、すべてのインターフェース。',
    aussen:
      '選択したアドレスはこの Mac の外に出ます。その場合、サービスは TLS とアクセス（トークン）を要求し、macOS がローカルネットワークの許可を求めることがあります。',
    inaktiv: 'オフ',
    ohneAdresse: 'アドレスなし',
    liste: 'この Mac のインターフェース',
    arten: {
      schleife: 'ループバック',
      netz: 'ネットワーク',
      tunnel: 'トンネル（VPN）',
      bruecke: 'ブリッジ',
      direkt: 'ダイレクト無線',
    },
    speichern: '選択を適用',
    gespeichert: '適用しました。次にサービスを起動したときに有効になります。',
    keineGewaehlt: 'アドレスを 1 つ以上選んでください。',
    platzhalter: '「サービスの待ち受け先」の設定どおり',
    weitere: 'ほかの待ち受け先',
    ohne: '{adresse} なし：{grund}',
  },
}
