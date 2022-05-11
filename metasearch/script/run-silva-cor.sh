#!/bin/bash

if [ "$1" = "" ]; then echo $0 "<input.fastq(.gz)>"; exit 1; fi

echo "CMD: $0 $*"
n=40000 #使用するリード数

work=$PWD
pid=$$
date=`date '+%Y-%m-%d-%H-%M-%S-%N'`
tempdir=/tmp/${date}-$pid
echo "tempdir: $tempdir"

input=`readlink -f "$1"`

maindir=$(dirname `readlink -f $0`)
source "$maindir"/settings.sh

mkdir -p $tempdir/input
cd $tempdir

#入力ファイルがgz圧縮されているか調べて圧縮されていたらzcatそうでなければcatを使う
if [ `echo "$input"|grep [.]gz$|wc -l` = 1 ]; then
    zcat "$input"|head -n $n > input/input.fq
else
    cat "$input"|head -n $n > input/input.fq
fi

#BLAST->LCA解析を実行
"$maindir"/metagenome~silva_SSU+LSU -c 8 -m 32 -d 50 -t 0.99 input

mv input/*.ssu.blast.filtered.name.lca.cnt2.input "$input".tsv
cd /tmp
rm -rf $tempdir

if [ "`cat \"$input\".tsv |wc -l`" -lt 2 ]; then
    echo 'Error: No output!'
    exit 1
elif [ "`cat \"$input\".tsv |wc -l`" = 2 ] && [ "`tail -n 1 \"$input\".tsv|cut -f 1`" = "No Hit" ]; then
    echo 'Error: No hit in ribosomal database!'
    exit 1
fi

awk -F'\t' '
 FILENAME==ARGV[1]{a["root;"$2]=$3}
 FILENAME==ARGV[2]{if($1=="id"){if(FNR>1){for(i in cnt){if(i!=""){print i"\t"cnt[i]}}}; print $0; delete cnt}else{cnt[a[$1]]+=$2}}
 END{for(i in cnt){if(i!=""){print i"\t"cnt[i]}}}
' <(zcat "$maindir"/SILVA_132_SSU-LSU_Ref.fasta.name.species.gz) "$input".tsv > "$input".species.tsv

awk -F'\t' '
 FILENAME==ARGV[1]{a["root;"$2]=$3}
 FILENAME==ARGV[2]{if($1=="id"){if(FNR>1){for(i in cnt){if(i!=""){print i"\t"cnt[i]}}}; print $0; delete cnt}else{cnt[a[$1]]+=$2}}
 END{for(i in cnt){if(i!=""){print i"\t"cnt[i]}}}
' <(zcat "$maindir"/SILVA_132_SSU-LSU_Ref.fasta.name.genus.gz) "$input".tsv > "$input".genus.tsv

"$maindir"/calccor "$input".tsv "$maindir"/db_merge
"$maindir"/calccor "$input".genus.tsv "$maindir"/db_genus_merge
"$maindir"/calccor "$input".species.tsv "$maindir"/db_species_merge
