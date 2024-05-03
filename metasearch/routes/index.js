'use strict';
var express = require('express');
var router = express.Router();
var crypto = require("crypto");
var fs = require("fs");
var multer = require("multer");
var upload = multer({ dest: "tmp/" });
var { execSync } = require('child_process');
require('date-utils');
var sqlite3 = require('sqlite3').verbose();
const readline = require('readline');

/* GET home page. */
router.get('/', function (req, res, next) {
    res.render('index');
});

router.post('/upload', upload.single('file'), function (req, res, next) {
    //リクエストからEmailアドレスを取得
    var email = req.body.email;

    //Emailアドレスに不要な文字列を含まないかをチェック
    var regex = /^[a-zA-Z0-9_.+-]+@([a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]*\.)+[a-zA-Z]{2,}$/;
    if (!(regex.test(email))) {
        console.log("incorrect email address");
        return
    }

    //元のファイル名を取得
    var original_filename = req.body.filename;

    //ファイル名に含まれるコマンドインジェクションを発生しうる特殊文字を除去する
    original_filename = original_filename.replace(/\$/g, "_").replace(/\;/g, "_").replace(/\|/g, "_").replace(/\&/g, "_").replace(/\'/g, "_").replace(/\(/g, "_").replace(/\)/g, "_").replace(/\</g, "_").replace(/\>/g, "_").replace(/\*/g, "_").replace(/\?/g, "_").replace(/\{/g, "_").replace(/\}/g, "_").replace(/\[/g, "_").replace(/\]/g, "_").replace(/\!/g, "_").replace(/\`/g, "_").replace(/\"/g, "_");

    //リクエストで送られて来る分割後のファイル名を取得(Hash値になっている)
    var filename = req.file.filename;

    //変更後のファイル名
    var newfilename;

    ////日付を取得
    //var dt = new Date();
    //
    ////ランダムな文字列を生成
    //var chars = 'sdasdalASDKAJsdlaj298an2a2kd9';
    //var rand_str = '';
    //for (var i = 0; i < 8; i++) {
    //    rand_str += chars.charAt(Math.floor(Math.random() * chars.length)); //ランダムな1文字を取得
    //}

    // 現在の日時を取得
    const now = new Date();

    // 各部分を個別に取得して、必要に応じてゼロ埋めする
    const year = now.getFullYear();
    const month = String(now.getMonth() + 1).padStart(2, '0'); // 月は0から始まるため、1を加える
    const day = String(now.getDate()).padStart(2, '0');
    const hour = String(now.getHours()).padStart(2, '0');
    const minute = String(now.getMinutes()).padStart(2, '0');
    const second = String(now.getSeconds()).padStart(2, '0');
    const millisecond = String(now.getMilliseconds()).padStart(3, '0');

    // フォーマットに従って文字列を組み立てる
    let dateString = `${year}${month}${day}${hour}${minute}${second}${millisecond}`;

    //ランダムな文字列を生成
    var chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
    for (var i = 0; i < 4; i++) {
        dateString += chars.charAt(Math.floor(Math.random() * chars.length)); //ランダムな1文字を取得
    }

    console.log("new data: "+dateString);

    ////日付とEmailアドレスとランダム文字列からハッシュ値を計算
    //var hash = crypto.createHash("md5").update(dt + email + rand_str).digest("hex");
    var hash = dateString;

    //ハッシュ値と同じ名前のディレクトリをtmp以下に作成
    fs.mkdirSync('tmp/'+hash, { recursive: true });
    //execSync("mkdir tmp/" + hash);
    ////HTMLファイル保存用のディレクトリを作成
    //execSync("mkdir tmp/" + hash + "/" + hash);

    ////拡張子に合わせて、新しいファイル名と移動先のパスを設定
    //if (getExt(original_filename) == "gz") {
    //    newfilename = "tmp/" + hash + "/" + hash + ".gz";
    //} else {
    //    newfilename = "tmp/" + hash + "/" + hash + ".fq";
    //}
    newfilename = "tmp/" + hash + "/" + hash + ".fq.gz";

    //ファイル名の変更とファイルの移動
    fs.rename("tmp/" + filename, newfilename, (err) => {
        if (err) throw err;
    });

    //解析スクリプトを実行,HTMLを作成
    execSync("script/run-qsub.sh "+hash+" "+newfilename+" "+original_filename+" "+email);
    //レスポンスを返す(これがないとPOSTが上手くいかない)
    res.send("uploaded");
});

router.get('/uploaded', function (req, res, next) {
    res.render('uploaded');
});

router.get('/about', function (req, res, next) {
    res.render('about');
});

router.get('/help', function (req, res, next) {
    res.render('help');
});

router.get('/contact', function (req, res, next) {
    res.render('contact');
});

router.get('/species', function (req, res, next) {
  // GETパラメータから`name`の値を取得
  var aValue = req.query.name;
  var db = new sqlite3.Database('data/species.db');
  // SQLiteを使って`name`の値に基づいてデータベースを検索
  //db.all("select * from data join srainfo on data.srr_name = srainfo.srr_id where data.sp_name = ? order by data.percent desc limit 1000", [aValue], function(err, rows) {
  db.all("SELECT * FROM data WHERE sp_name = ? ORDER BY percent DESC LIMIT 1000", [aValue], function(err, rows) {
    if (err) {
      res.status(500).send("データベースエラー");
      return console.error(err.message);
    }
    const srrids = rows.map(item => {
      //console.log(item)
      return item.srr_name
    });
    const placeholders = srrids.map((_, index) => '?').join(','); // SQLインジェクションを防ぐ
    const orderClause = srrids.map((id, index) => `WHEN srr_id = '${id}' THEN ${index}`).join(' ');
    //console.log(placeholders)
    //console.log(orderClause)
    // SQLクエリを実行
    //db.all(`SELECT * FROM srainfo WHERE srr_id IN (${placeholders}) ORDER BY CASE ${orderClause} END`, srrids, (err, rows2) => { //出来るだけ順番をクエリー順に保存するならこっちだけど、結果がないIDは飛ばされるので後で自分で連結することにした
    db.all(`SELECT * FROM srainfo WHERE srr_id IN (${placeholders})`, srrids, (err, rows2) => {
      if (err) {
        console.error('クエリ実行エラー: ' + err.message);
        return;
      }
      // 結果の表示
      const rows2r = {}
      for(let i of rows2){rows2r[i.srr_id]=i}
      const rowplus = rows.map(item => {
        if(item.srr_name in rows2r){
          item.expname = rows2r[item.srr_name].srx_name
          item.organism = rows2r[item.srr_name].srs_org
          item.studyname = rows2r[item.srr_name].srp_name
          item.reads = rows2r[item.srr_name].srr_reads
          item.geo = rows2r[item.srr_name].srs_geo
        }else{
           item.expname = ""
           item.organism = ""
           item.studyname = ""
           item.reads = ""
           item.geo = ""
        }
        return item
      });

      db.get("SELECT COUNT(*) AS count FROM data WHERE sp_name = ?", [aValue], (err, row) => {
        if (err) {
          return console.error(err.message);
        }
        res.render('species', { results: rowplus, key: req.query.name, count: row.count });
      });
    });

  });

  // データベース接続を閉じる
  db.close((err) => {
    if (err) {
      console.error('データベースの切断エラー: ' + err.message);
      return;
    }
    console.log('データベースの接続を閉じました');
  });
});

router.get('/srr', function (req, res, next) {
  const requestedFile = req.query.id; // GETパラメータからファイル名を取得
  const listFilePath = 'data/input.list'; // ファイルリストのパス

  // ファイルリストを読み込む
  fs.readFile(listFilePath, 'utf8', (err, data) => {
    if (err) {
      return res.status(500).send("リストファイルを読み込めませんでした。");
    }

    // ファイルリストからファイルのパスを検索
    const lines = data.split('\n');
    const fileEntry = lines.find(line => line.startsWith(requestedFile + '\t'));
    if (!fileEntry) {
      return res.status(404).send("ファイルがリストに見つかりません。");
    }

    // ファイルのパスを抽出して応答する
    const filePath = 'data/'+fileEntry.split('\t')[1];
    //console.log(filePath);

    fs.readFile(filePath, 'utf8', (err, data) => {
      if (err) {
        return res.status(500).send("ファイルを読み込めませんでした。");
      }
      // テキストファイルの内容をPugテンプレートに渡してレンダリング

      // データを行に分割し、最初の行（ヘッダー）を除外
      const rows = data.trim().split('\n').slice(1);

      // 各行をタブで分割してオブジェクトの配列に変換
      const tableData = rows.map(row => {
        const columns = row.split('\t');
        return {
          name: columns[0],
          abundance: columns[1],
          // 必要に応じて他の列も追加
        };
      });

      res.render('srr', { tableData: tableData, key: req.query.id});
    });
  });
});

module.exports = router;

function getExt(filename) {
    var pos = filename.lastIndexOf(".");
    if (pos === -1) return "";
    return filename.slice(pos + 1);
}
