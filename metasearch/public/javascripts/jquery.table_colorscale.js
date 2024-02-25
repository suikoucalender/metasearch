/*!
 * jQuery Table ColorScale v0.9.0
 * (c) 2013 Hidehisa YASUDA.
 * Released under the MIT license
 */

(function ( $ ) {
  $.fn.tableColorScale = function(opt){

    opt = opt || {};
    var options = {}

    options.type = opt.type || 'percentile';
    /*
      none
      percentile
      sigma
      topn
      bar
    */
    
    options.typeOpt = opt.typeOpt || false;
    options.funcs = {};

    options.css = $.extend({
      '75%' : {backgroundColor: 'rgba(0,255,0,0.4)'},
      '50%' : {backgroundColor: 'rgba(208,255,0,0.4)'},
      '25%' : {backgroundColor: 'rgba(255,128,0,0.4)'},
      '0%' : {backgroundColor: 'rgba(255,0,0,0.4)'},

      '2sigma+' : {backgroundColor: 'rgba(0,255,0,0.4)'},
      '1sigmaTo2sigma' : {backgroundColor: 'rgba(128,255,0,0.4)'},
      'avgTo1sigma' : {backgroundColor: 'rgba(204,255,0,0.4)'},
      'avgTo-1sigma' : {backgroundColor: 'rgba(255,204,0,0.4)'},
      '-1sigmaTo-2sigma' : {backgroundColor: 'rgba(255,128,0,0.4)'},
      '-2sigma-' : {backgroundColor: 'rgba(255,0,0,0.4)'},

      'topnTop' : {backgroundColor: 'rgba(0,255,0,0.4)'},
      'topnBottom' : {backgroundColor: 'rgba(255,0,0,0.4)'},
      
      'databar' : {backgroundColor: 'rgba(128,201,255,0.4)'}
    }, opt.css);

    options.numTextAlign = opt.numTextAlign || 'right';

    var
      elements = [],  // 要素を（数値も含めて全部）放り込んでる
      numbers = [], // 要素のなかの数値データだけ放り込んでる
      func,
      i;

    // 数値にする関数
    // /[^\d.]/でやると文字列に数字だけ入ってるのが数値扱いされるのでやらない
    // /\\/がヒットせずよくわからなかったので強引に「\」を削除
    var getNumber = function(s) {
      var ss = s.replace(/[,%$\\:]/g, '');
      var sss = '';
      var len = ss.length;
      for (var i = 0; i < len; i++) {
        if (ss.charCodeAt(i) != 165) {
          sss += ss[i];
        }
      }
      return (sss == parseFloat(sss) ? parseFloat(sss) : false);
    }


    // 標準の処理
    // 対象から数値だけを抽出
    numbers = [];
    this.each(function(){
      var n = getNumber(this.innerHTML);
      if (n !== false) {
        numbers.push(n);
      }
      elements.push(n);
    });

    // 数値のデータがなければ終わり
    if (!numbers.length) {
      return;
    }



    // スタイルごとの処理

    /**
     * none
     * 何も処理をしない（オプションのtext-alignだけ利用するため）
     */
    options.funcs.none = function(targets, options) {
    }


    /**
     * percentile
     * グループ内で大小の順に並べたときの全体におけるパーセンタイルに応じて
     */
    options.funcs.percentile = function(targets, options) {

      // 最後のスタイル設定のときの基準値
      var metrics ={};
      metrics['100%'] = Math.max.apply(null, numbers),
      metrics['0%'] = Math.min.apply(null, numbers),
      metrics['75%'] = metrics['0%'] + (metrics['100%'] - metrics['0%']) * 0.75,
      metrics['50%'] = metrics['0%'] + (metrics['100%'] - metrics['0%']) * 0.5,
      metrics['25%'] = metrics['0%'] + (metrics['100%'] - metrics['0%']) * 0.25

      targets.each(function(i){
        if (elements[i] === false) {
          return;
        }

        if (elements[i] >= metrics['75%']) {
          $(this).css(options.css['75%']);
        }
        else if (elements[i] >= metrics['50%']) {
          $(this).css(options.css['50%']);
        }
        else if (elements[i] >= metrics['25%']) {
          $(this).css(options.css['25%']);
        }
        else {
          $(this).css(options.css['0%']);
        }
      });
    };

    /** 
     * sigma
     * 平均値からの標準偏差ベースの差によって判定
     */
    options.funcs.sigma = function(targets, options) {
      var getsigma = function(a){
        var sum = 0;
        var avg = 0;
        var mean = 0;
        var sigma = 0;
        var leastNum = 2;
        // sumとavgは個数が2未満のときでも処理する必要アリ
        if (a.length && !(a.length > leastNum)) {
          $.each(a, function(){
            sum += this;
          });
          avg = sum / a.length;
        }
        if (a.length > leastNum) {
          sum = 0;
          avg = 0;
          mean = 0;
          sigma = 0;
          // ソートはデフォルト文字コード順。。。
          a.sort(function(a,b){return (Number(b)-Number(a));});
          //中央値
          if (a.length % 2) {
            mean = a[Math.round(a.length / 2) - 1];
          }
          else {
            mean = (a[a.length / 2 - 1] + a[a.length / 2]) / 2;
          }
          // 標準偏差
          sum = 0;
          $.each(a, function(){
            sum += this;
          });
          avg = sum / a.length;
          sigma = 0;
          $.each(a, function(){
            sigma += Math.pow(this - avg, 2);
          });
          sigma = Math.sqrt(sigma / a.length);
        }
        return {
          'sum' : sum,
          'avg' : avg,
          'mean' : mean,
          'sigma' : sigma
        };
      }
      a = getsigma(numbers);
      var metrics = {};
      metrics['2sigma'] = a.avg + a.sigma * 2;
      metrics['1sigma'] = a.avg + a.sigma;
      metrics['avg'] = a.avg;
      metrics['-1sigma'] = a.avg - a.sigma;
      metrics['-2sigma'] = a.avg - a.sigma * 2;

      targets.each(function(i){
        if (elements[i] === false) {
          return;
        }

        if (elements[i] >= metrics['2sigma']) {
          $(this).css(options.css['2sigma+']);
        }
        else if (elements[i] >= metrics['1sigma']) {
          $(this).css(options.css['1sigmaTo2sigma']);
        }
        else if (elements[i] >= metrics['avg']) {
          $(this).css(options.css['avgTo1sigma']);
        }
        else if (elements[i] >= metrics['-1sigma']) {
          $(this).css(options.css['avgTo-1sigma']);
        }
        else if (elements[i] >= metrics['-2sigma']) {
          $(this).css(options.css['-1sigmaTo-2sigma']);
        }
        else {
          $(this).css(options.css['-2sigma-']);
        }
      });
    }


    /**
     * topn
     * 上位・下位のn個を指定
     * オプションのtypeOptの指定で範囲を制御
     * typeOpt値の応じた動作は次のとおり
     *  10 : 上位10個
     *  -10 : 下位10個
     *  10-5 : 上位10個と下位5個
     */
    options.funcs.topn = function(targets, options) {

      if (!options.typeOpt) {
        options.typeOpt = 10;
      }
      // +n-n形式の場合は、それぞれに分解して実行
      // +n-n、+n/-n、+n -n、+n&-n などでも指定可能で、先頭の+はオプション
      var matches;
      if (matches = String(options.typeOpt).match(/^\+*(\d+)[\s\/&]*(-\d+)/, matches)) {
        options.funcs.topn(targets, $.extend(options, {typeOpt : matches[1]}));
        options.funcs.topn(targets, $.extend(options, {typeOpt : matches[2]}));
        return;
      }

      // numbersをソート
      if (options.typeOpt > 0) {
        numbers.sort(function(a,b){return (Number(b)-Number(a));});
      }
      else {
        numbers.sort(function(a,b){return (Number(a)-Number(b));});
      }

      var topn = [];
      $(numbers).each(function(){
        if (topn.length == 0 || this != topn[topn.length - 1]) {
          topn.push(this);
        }
        if (topn.length >= Math.abs(options.typeOpt)) {
          return false;
        }
      });

      targets.each(function(i){
        if (elements[i] === false) {
          return;
        }

        if (options.typeOpt > 0 && elements[i] >= topn[topn.length - 1]) {
          $(this).css(options.css['topnTop']);
        }
        else if(options.typeOpt < 0 && elements[i] <= topn[topn.length - 1]) {
          $(this).css(options.css['topnBottom']);
        }
      });
    }



    /**
     * databar
     * データバーを描画。グループ内最大値を100%とする
     */
    options.funcs.databar = function(targets, options) {

      var metrics ={};
      metrics['100%'] = Math.max.apply(null, numbers),

      targets.each(function(i){
        if (elements[i] === false) {
          return;
        }
        var bar = $(document.createElement('span'))
          .css($.extend({
            'position': 'absolute',
            'top': 0,
            'left': 0,
            'zIndex': 0,
            'display': 'block',
            'height': '100%',
            'width' : ((elements[i]/metrics['100%'])*100)+'%'
          }, options.css['databar']))
        $(this).prepend(bar);
        $(this).wrapInner($(document.createElement('div')).css({'position':'relative'}));
      });
    }


    // 処理を実行
    if (typeof options.type === 'string' && options.funcs[options.type]) {
      func = options.funcs[options.type];
    }
    else if (typeof options.type === 'function') {
      func = options.type;
    }
    if (func) {
      func(this, options);
      // 数値データのtext-alignを調整
      if (options.numTextAlign) {
        this.each(function(i){
          if (elements[i] === false) {
            return;
          }
          else {
            $(this).css({'textAlign': options.numTextAlign});
          }
        });
      }
    }
    else {
      console.debug('options.type not function');
    }

    return this;

  }
})( jQuery );
