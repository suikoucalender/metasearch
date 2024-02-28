#!/bin/bash

#tmp/hash値/hash値(.gz)
newfilename=$1 #tmp/cf5a956d1de00aaea36b87346b21b4e8/cf5a956d1de00aaea36b87346b21b4e8.fq.gz
hash=$2 #cf5a956d1de00aaea36b87346b21b4e8
original_filename=$3 #y2022-group3-fish-16S-Kabayakisantaro.fq
usr_email=$4 #g.ecc.u-tokyo.ac.jp

sdir=$(dirname `readlink -f $0`)
maindir=$(dirname "$sdir")
source ~/.bashrc

set -x

#DBディレクトリが格納されているディレクトリの絶対パス
#dbPath=/usr/local/yoshitake/
dbPath=$(dirname $(readlink -f $maindir/data/db))

script/run-silva-cor.sh $newfilename

#Singularityのイメージがなければ、githubのリリースから取ってくる。ファイルサイズが大きいのでソースコードには含められない。
if [ ! -e "${sdir}/python3_env_mako_installed.sif" ]; then
 wget -O "${sdir}/python3_env_mako_installed.sif" https://github.com/suikoucalender/metasearch/releases/download/0.1/python3_env_mako_installed.sif
fi
if [ ! -e "${sdir}/krona_v2.7.1_cv1.sif" ]; then
 wget -O "${sdir}/krona_v2.7.1_cv1.sif" https://github.com/suikoucalender/metasearch/releases/download/0.1/krona_v2.7.1_cv1.sif
fi

if [ ! -e $newfilename.tsv ]; then
 singularity run -B $maindir $sdir/python3_env_mako_installed.sif python $sdir/send_err_mail.py ${url}/${hash}/ ${usr_email} ${original_filename}
 exit
fi

echo "<html><body>" > $maindir/tmp/${hash}/result.html
for k in correlation correlation.log correlation.fullnodes correlation.fullnodes.log weighted_jaccard weighted_jaccard.log weighted_jaccard.fullnodes weighted_jaccard.fullnodes.log jaccard jaccard.fullnodes; do

echo "<a href='$k.output.html'>$k</a><br>" >> $maindir/tmp/${hash}/result.html
#類似度スコア＆サンプル名など
cat $sdir/table_header.html > tmp/$hash/$k.output.html
echo '<header><div class="head-content"><h1>MetaSearch Result</h1><h3>'"${original_filename}"'</h3></div></header><main><div class="main-content">' >> tmp/$hash/$k.output.html

cat $newfilename.tsv.result.$k |awk -F'\t' '
 FILENAME==ARGV[1]{a[NR]=$1; b[NR]=$2; a2[$1]=1}
 FILENAME==ARGV[2]&&$2 in a2{c[$2]=$3; d[$2]=$4; e[$2]=$5}
 END{for(i=1;i<=length(a); i++){print a[i]"\t"b[i]"\t"c[a[i]]"\t"d[a[i]]"\t"e[a[i]]}}' /dev/stdin data/sra_info.txt > tmp/$hash/$k.output.sampleinfo.txt
echo "<div class='info'><h3>Similarity and sample information</h3>" >> tmp/$hash/$k.output.html
awk -F'\t' 'BEGIN{print " <table id=\"info\"><thead><tr><th>ID</th><th>value</th><th>Experiment name</th><th>Organism</th><th>Study name</th></tr></thead><tbody>"}
 {print " <tr>"; print "  <td><a href=\"https://www.ncbi.nlm.nih.gov/sra/?term="$1"\">"$1"</a></td>"; for(i=2;i<=NF;i++){print "  <td>"$i"</td>"}; print " </tr>"} END{print "</tbody></table>"}' tmp/$hash/$k.output.sampleinfo.txt >> tmp/$hash/$k.output.html
echo "</div>" >> tmp/$hash/$k.output.html

#Kronaのグラフ
cat $newfilename.tsv.result.$k|cut -f 1|
 while read i; do tail -n+2 data/db/*/$i.input|awk -F'\t' '{print $2"\t"$1}'|sed 's/;/\t/g' > tmp/$hash/$k.krona.$i.input; done
cat $newfilename.tsv|awk -F'\t' '{print $2"\t"$1}'|sed 's/;/\t/g' > tmp/$hash/$k.krona.input.input
in="tmp/$hash/$k.krona.input.input,input"
while read i; do
 in="$in tmp/$hash/$k.krona.$i.input,$i"
done < <(cat $newfilename.tsv.result.$k|cut -f 1)
singularity run -B $maindir $sdir/krona_v2.7.1_cv1.sif ktImportText -o $maindir/tmp/${hash}/$k.output.krona.html $in
echo "<div class='krona'><h3>Krona Compotion Graph</h3><iframe src='./$k.output.krona.html' width='1000' height='600' frameborder='0'></iframe></div>" >> tmp/$hash/$k.output.html

#割合表
cat $newfilename.tsv.result.$k|cut -f 1|
 while read i; do cat data/db/*/$i.input| awk -F'\t' 'NR==1{print $0} NR>1{a[$1]=$2; cnt+=$2} END{for(i in a){print i"\t"a[i]/cnt*100}}' > tmp/$hash/$k.merge.$i.input; done
cat $newfilename.tsv|awk -F'\t' 'NR==1{print $0} NR>1{a[$1]=$2; cnt+=$2} END{for(i in a){print i"\t"a[i]/cnt*100}}' > tmp/$hash/$k.merge.input
in=tmp/$hash/$k.merge.input
while read i; do
 in="$in tmp/$hash/$k.merge.$i.input"
done < <(cat $newfilename.tsv.result.$k|cut -f 1)
$sdir/merge_table.pl -k $in |sed 's/\t\t/\t0\t/g; s/\t\t/\t0\t/g; s/\t$/\t0/' > tmp/$hash/$k.output.merge.txt
(head -n 1 tmp/$hash/$k.output.merge.txt ; tail -n+2 tmp/$hash/$k.output.merge.txt |sort -k2,2nr -t$'\t') > tmp/$hash/$k.output.merge.sort.txt
echo "<div class='freq'><h3>Abundance</h3>" >> tmp/$hash/$k.output.html
cat tmp/$hash/$k.output.merge.sort.txt |awk -F'\t' '
 BEGIN{print "<table id=\"test\">"}
 NR==1{print " <thead><tr>"; for(i=1;i<=NF;i++){print "  <th>"$i"</th>"}; print " </tr></thead>"; print " <tbody>"}
 NR>1{
  ori=$1; n=split($1,arr,";");
  delete links
  taxpath[1]=arr[1]
  for(i=2;i<=length(arr);i++){
   taxpath[i]=taxpath[i-1]";"arr[i]
  }
  if(n>5){
   res="<a href='"'"'../species?name="taxpath[1]"'"'"'>"arr[1]"</a>"
   res=res";...;<a href='"'"'../species?name="taxpath[length(arr)-2]"'"'"'>"arr[length(arr)-2]"</a>"
   res=res";<a href='"'"'../species?name="taxpath[length(arr)-1]"'"'"'>"arr[length(arr)-1]"</a>"
   res=res";<a href='"'"'../species?name="taxpath[length(arr)]"'"'"'>"arr[length(arr)]"</a>"
  }else{
   res="<a href='"'"'../species?name="taxpath[1]"'"'"'>"arr[1]"</a>"
   for(i=2;i<=length(arr);i++){
    res=res";<a href='"'"'../species?name="taxpath[i]"'"'"'>"arr[i]"</a>"
   }
  }
  print "  <tr>";
  print "   <td title=\""ori"\">"res"</td>";
  for(i=2;i<=NF;i++){print "   <td>"$i"</td>"}; print "  </tr>"
 }
 END{print " </tbody>"; print "</table></body></html>"}' >> tmp/$hash/$k.output.html
echo "</div>" >> tmp/$hash/$k.output.html #class=freqの閉じ
echo "</div></main>" >> tmp/$hash/$k.output.html #mainの閉じ

done
echo "</body></html>" >> $maindir/tmp/${hash}/result.html

#singularity run -B $maindir -B $dbPath $sdir/python3_env_mako_installed.sif python $sdir/create_page.py $newfilename $original_filename $dbPath

cp -r tmp/${hash} public/
url=` cat config/config.json | grep "url" | sed -r 's/^[^:]*:(.*)$/\1/' | sed 's/\"//g' | sed "s/,//g" | sed 's/ //g'`

singularity run -B $maindir $sdir/python3_env_mako_installed.sif python $sdir/send_mail.py ${url}/${hash}/ ${usr_email} ${original_filename}
