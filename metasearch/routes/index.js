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
    //���N�G�X�g����Email�A�h���X���擾
    var email = req.body.email;

    //Email�A�h���X�ɕs�v�ȕ�������܂܂Ȃ������`�F�b�N
    var regex = /^[a-zA-Z0-9_.+-]+@([a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9]*\.)+[a-zA-Z]{2,}$/;
    if (!(regex.test(email))) {
        console.log("incorrect email address");
        return
    }

    //���̃t�@�C�������擾
    var original_filename = req.body.filename;

    //�t�@�C�����Ɋ܂܂��R�}���h�C���W�F�N�V�����𔭐���������ꕶ������������
    original_filename = original_filename.replace(/\$/g, "_").replace(/\;/g, "_").replace(/\|/g, "_").replace(/\&/g, "_").replace(/\'/g, "_").replace(/\(/g, "_").replace(/\)/g, "_").replace(/\</g, "_").replace(/\>/g, "_").replace(/\*/g, "_").replace(/\?/g, "_").replace(/\{/g, "_").replace(/\}/g, "_").replace(/\[/g, "_").replace(/\]/g, "_").replace(/\!/g, "_").replace(/\`/g, "_").replace(/\"/g, "_");

    //���N�G�X�g�ő����ė��镪����̃t�@�C�������擾(Hash�l�ɂȂ��Ă���)
    var filename = req.file.filename;

    //�ύX��̃t�@�C����
    var newfilename;

    //���t���擾
    var dt = new Date();

    //�����_���ȕ�����𐶐�
    var chars = 'sdasdalASDKAJsdlaj298an2a2kd9';
    var rand_str = '';
    for (var i = 0; i < 8; i++) {
        rand_str += chars.charAt(Math.floor(Math.random() * chars.length));
    }

    //���t��Email�A�h���X�ƃ����_�������񂩂�n�b�V���l���v�Z
    var hash = crypto.createHash("md5").update(dt + email + rand_str).digest("hex");

    //�n�b�V���l�Ɠ������O�̃f�B���N�g����tmp�ȉ��ɍ쐬
    execSync("mkdir tmp/" + hash);
    //HTML�t�@�C���ۑ��p�̃f�B���N�g�����쐬
    execSync("mkdir tmp/" + hash + "/" + hash);

    //�g���q�ɍ��킹�āA�V�����t�@�C�����ƈړ���̃p�X��ݒ�
    if (getExt(original_filename) == "gz") {
        newfilename = "tmp/" + hash + "/" + hash + ".gz";
    } else {
        newfilename = "tmp/" + hash + "/" + hash + ".fq";
    }

    //�t�@�C�����̕ύX�ƃt�@�C���̈ړ�
    fs.rename("tmp/" + filename, newfilename, (err) => {
        if (err) throw err;
    });

    //�g���搶�̃X�N���v�g�����s,HTML���쐬
    execSync("qsub -e ./qsub_log/e." + hash + " -o ./qsub_log/o." + hash + " -cwd -pe def_slot 4 -j y -N 'metasearch' script/metasearch_exec.sh " + newfilename + " " + hash + " " + original_filename + " " + email);

    //���X�|���X��Ԃ�(���ꂪ�Ȃ���POST����肭�����Ȃ�)
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
