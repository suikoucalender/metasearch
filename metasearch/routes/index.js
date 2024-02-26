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

    //日付を取得
    var dt = new Date();

    //ランダムな文字列を生成
    var chars = 'sdasdalASDKAJsdlaj298an2a2kd9';
    var rand_str = '';
    for (var i = 0; i < 8; i++) {
        rand_str += chars.charAt(Math.floor(Math.random() * chars.length));
    }

    //日付とEmailアドレスとランダム文字列からハッシュ値を計算
    var hash = crypto.createHash("md5").update(dt + email + rand_str).digest("hex");

    //ハッシュ値と同じ名前のディレクトリをtmp以下に作成
    execSync("mkdir tmp/" + hash);
    //HTMLファイル保存用のディレクトリを作成
    execSync("mkdir tmp/" + hash + "/" + hash);

    //拡張子に合わせて、新しいファイル名と移動先のパスを設定
    if (getExt(original_filename) == "gz") {
        newfilename = "tmp/" + hash + "/" + hash + ".gz";
    } else {
        newfilename = "tmp/" + hash + "/" + hash + ".fq";
    }

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
  db.all("SELECT * FROM data WHERE sp_name = ? ORDER BY percent DESC LIMIT 1000", [aValue], function(err, rows) {
    if (err) {
      res.status(500).send("データベースエラー");
      return console.error(err.message);
    }

    // IDに対応する情報を検索して配列に保存
    const srrNames = rows.map(row => row.srr_name);
    const infoFilePath = 'data/sra_info.txt';

    // infoファイルを読み込む
    fs.readFile(infoFilePath, 'utf8', (err, data) => {
      if (err) {
        return res.status(500).send("リストファイルを読み込めませんでした。");
      }
      const dataLines = data.trim().split('\n');
      const infoData = dataLines.map(line => line.split('\t')); //SRX, SRR, exp name, organism, study name
      const rowplus = rows.map(item => {
        const found = infoData.find(element => element[1] === item.srr_name);
        //console.log(found);
        if(found !== undefined){
          item.expname = found[2];
          item.organism = found[3];
          item.studyname = found[4];
        }
        return item;
      });
      db.get("SELECT COUNT(*) AS count FROM data WHERE sp_name = ? ORDER BY percent DESC LIMIT 1000", [aValue], (err, row) => {
        if (err) {
          return console.error(err.message);
        }
        res.render('species', { results: rowplus, key: req.query.name, count: row.count });
      });
    });
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
