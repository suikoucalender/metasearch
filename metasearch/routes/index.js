'use strict';
var express = require('express');
var router = express.Router();
var crypto = require("crypto");
var fs = require("fs");
var multer = require("multer");
var upload = multer({ dest: "tmp/" });
var { execSync } = require('child_process');
require('date-utils');

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

    //吉武先生のスクリプトを実行,HTMLを作成
    //execSync("qsub -e ./qsub_log/e." + hash + " -o ./qsub_log/o." + hash + " -cwd -pe def_slot 4 -j y -N 'metasearch' script/metasearch_exec.sh " + newfilename + " " + hash + " " + original_filename + " " + email);
    //execSync("/opt/sge/bin/lx-amd64/qsub -e ./qsub_log/e." + hash + " -o ./qsub_log/o." + hash + " -j y -N metasearch script/qsubsh4 script/metasearch_exec.sh " + newfilename + " " + hash + " " + original_filename + " " + email);
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

module.exports = router;

function getExt(filename) {
    var pos = filename.lastIndexOf(".");
    if (pos === -1) return "";
    return filename.slice(pos + 1);
}
