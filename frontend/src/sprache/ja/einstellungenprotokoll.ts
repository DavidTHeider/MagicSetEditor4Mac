// The log settings — OPS-04 (#1721).
import { messhinweise, messnamen, messwerte } from './einstellungenmessung'
import { netzwerte } from './einstellungennetz'

export const protokollnamen = {
  'protokoll.stufe': 'ログに記録する量',
  'protokoll.datei': 'ログファイル',
  'protokoll.system': 'macOS の統合ログ',
  'protokoll.ausgabe': '標準エラー出力',
  'protokoll.format': '行の形式',
  'protokoll.grenze': 'ログファイル1つの大きさ（MB）',
  'protokoll.aufbewahrung': '残しておく古いファイルの数',
  'protokoll.komponenten': '詳細ログの対象',
  'protokoll.ort': 'ログファイルのフォルダ',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messnamen,
} as const

export const protokollwerte = {
  'protokoll.stufe.fehler': 'エラーのみ',
  'protokoll.stufe.warnung': '警告とエラー',
  'protokoll.stufe.info': 'イベント（初期設定）',
  'protokoll.stufe.ausfuehrlich': '詳細 — すべてのステップ',
  'protokoll.datei.an': 'オン',
  'protokoll.datei.aus': 'オフ',
  'protokoll.system.an': 'オン',
  'protokoll.system.aus': 'オフ',
  'protokoll.ausgabe.aus': 'オフ',
  'protokoll.ausgabe.an': 'オン',
  'protokoll.format.text': '読みやすいテキスト',
  'protokoll.format.json': 'ツール向けの JSON',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messwerte,
  // **Und die Kontingente der Dienste** — QA-08.AK-03 (#1739): ihre Namen
  // stehen beim Netz, die Werte kommen hier dazu, weil `seiteneinstellungen.ts`
  // an der Zeilengrenze aus §2 steht.
  ...netzwerte,
} as const

export const protokollhinweise = {
  'protokoll.stufe':
    'ログに記録する内容です。**詳細**は問題を見つけるのに役立ちます。この場合、保持するのは1ファイルあたり最大1 MBと古いファイル1つだけになり、診断情報には詳細ログがオンであることが表示されます。次回の起動から有効になります。',
  'protokoll.datei':
    '不具合の報告に添付する `mse.log` ファイルです。作成できない場合でも他の出力先への書き込みは続き、診断情報に理由が表示されます。',
  'protokoll.system':
    'システムのログです。macOS のコンソールで「MagicSetEditor4Mac」を指定して確認できます。',
  'protokoll.ausgabe':
    'ターミナルから起動する場合向けです。すべての行をターミナルの表示先にも出力します。',
  'protokoll.format':
    '**JSON** は、イベントごとに1行で、名前付きのフィールドを持ちます。ログを収集するツール向けです。',
  'protokoll.grenze':
    'このサイズを超えると、次回の起動時に新しいファイルが始まります。空欄の場合は4 MBです。指定できるのは1〜64です。',
  'protokoll.aufbewahrung':
    '現在のファイルのほかに、`mse.log.1` から最も古いものまで、この数だけ古いファイルを残します。空欄の場合は1つです。指定できるのは1〜9です。',
  'protokoll.komponenten':
    '**詳細**の場合のみ有効です。どの部分のステップを記録するかを指定します。たとえば、ウインドウの呼び出しには `app` を指定します。空欄の場合はすべてです。イベントとエラーは常に記録されます。',
  'protokoll.ort':
    '/ または ~/ で始まるフォルダです。空欄の場合は、データベースの隣になります。',
  // **Und die Messung** — OPS-05 (#1722), aus ihrer eigenen Datei.
  ...messhinweise,
} as const
