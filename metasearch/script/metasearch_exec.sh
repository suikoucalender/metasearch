#!/bin/bash

#tmp/hash値/hash値(.gz)
newfilename=$1 #tmp/cf5a956d1de00aaea36b87346b21b4e8/cf5a956d1de00aaea36b87346b21b4e8.fq
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

#k=correlation
for k in correlation correlation.log weighted_jaccard weighted_jaccard.log jaccard correlation.fullnodes correlation.fullnodes.log weighted_jaccard.fullnodes weighted_jaccard.fullnodes.log jaccard.fullnodes; do

tempname=`basename $newfilename.tsv.result.$k`
echo "<a href='$tempname'>$tempname</a>" >> $maindir/tmp/${hash}/result.html

cat $newfilename.tsv.result.$k|cut -f 1|
 while read i; do tail -n+2 data/db/*/$i.input|awk -F'\t' '{print $2"\t"$1}'|sed 's/;/\t/g' > tmp/$hash/$k.krona.$i.input; done
cat $newfilename.tsv|awk -F'\t' '{print $2"\t"$1}'|sed 's/;/\t/g' > tmp/$hash/$k.krona..input.input
in=""; for i in $maindir/tmp/$hash/$k.krona.*.input; do in="$in "$i,`echo $i|sed 's%.*/'"$k"'[.]krona[.]\+%%; s/[.]input$//'`; done
singularity run -B $maindir $sdir/krona_v2.7.1_cv1.sif ktImportText -o $maindir/tmp/${hash}/$k.output.krona.html $in
echo "<a href='$k.output.krona.html'>$k.output.krona.html</a>" >> $maindir/tmp/${hash}/result.html

cat $newfilename.tsv.result.$k|cut -f 1|
 while read i; do cat data/db/*/$i.input| awk -F'\t' 'NR==1{print $0} NR>1{a[$1]=$2; cnt+=$2} END{for(i in a){print i"\t"a[i]/cnt*100}}' > tmp/$hash/$k.merge.$i.input; done
cat $newfilename.tsv|awk -F'\t' 'NR==1{print $0} NR>1{a[$1]=$2; cnt+=$2} END{for(i in a){print i"\t"a[i]/cnt*100}}' > tmp/$hash/$k.merge.input
$sdir/merge_table.pl -k tmp/$hash/$k.merge.input tmp/$hash/$k.merge.*.input|sed 's/\t\t/\t0\t/g; s/\t\t/\t0\t/g; s/\t$/\t0/' > tmp/$hash/$k.output.merge.txt
(head -n 1 tmp/$hash/$k.output.merge.txt ; tail -n+2 tmp/$hash/$k.output.merge.txt |sort -k2,2nr -t$'\t') > tmp/$hash/$k.output.merge.sort.txt
cat $sdir/table_header.html > tmp/$hash/$k.output.merge.sort.html
cat tmp/$hash/$k.output.merge.sort.txt |awk -F'\t' '
 BEGIN{print "<table id='test'>"}
 NR==1{print " <thead><tr>"; for(i=1;i<=NF;i++){print "  <th>"$i"</th>"}; print " </tr></thead>"; print " <tbody>"}
 NR>1{n=split($1,arr,";"); if(n>5){$1=arr[1]";...;"arr[length(arr)-2]";"arr[length(arr)-1]";"arr[length(arr)]}; print "  <tr>"; for(i=1;i<=NF;i++){print "   <td>"$i"</td>"}; print "  </tr>"}
 END{print " </tbody>"; print "</table>"}' >> tmp/$hash/$k.output.merge.sort.html
echo "</body></html>" >> tmp/$hash/$k.output.merge.sort.html
echo "<a href='$k.output.merge.sort.html'>$k.output.merge.sort.html</a>" >> $maindir/tmp/${hash}/result.html

cat $newfilename.tsv.result.$k |awk -F'\t' '
 FILENAME==ARGV[1]{a[NR]=$1; b[NR]=$2; a2[$1]=1}
 FILENAME==ARGV[2]&&$2 in a2{c[$2]=$3; d[$2]=$4; e[$2]=$5}
 END{for(i=1;i<=length(a); i++){print a[i]"\t"b[i]"\t"c[a[i]]"\t"d[a[i]]"\t"e[a[i]]}}' /dev/stdin data/sra_info.txt > tmp/$hash/$k.output.sampleinfo.txt
echo "<a href='$k.output.sampleinfo.txt'>$k.output.sampleinfo.txt</a>" >> $maindir/tmp/${hash}/result.html

done

#singularity run -B $maindir -B $dbPath $sdir/python3_env_mako_installed.sif python $sdir/create_page.py $newfilename $original_filename $dbPath

#for class in "" .genus .species
#do
#	singularity run -B $maindir $sdir/krona_v2.7.1_cv1.sif ktImportText $maindir/tmp/${hash}/result${class}.kraken -o $maindir/tmp/${hash}/${hash}/krona_out${class}.html
#done
#singularity run -B $maindir $sdir/krona_v2.7.1_cv1.sif ktImportText $maindir/tmp/${hash}/result.kraken -o $maindir/tmp/${hash}/krona_out.html

cp -r tmp/${hash} public/
url=` cat config/config.json | grep "url" | sed -r 's/^[^:]*:(.*)$/\1/' | sed 's/\"//g' | sed "s/,//g" | sed 's/ //g'`

singularity run -B $maindir $sdir/python3_env_mako_installed.sif python $sdir/send_mail.py ${url}/${hash}/ ${usr_email} ${original_filename}
