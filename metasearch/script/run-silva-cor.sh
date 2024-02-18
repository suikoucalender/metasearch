#!/bin/bash

if [ "$1" = "" ]; then echo $0 "<input.fastq(.gz)>"; exit 1; fi

echo "CMD: $0 $*"
n=40000 #使用するリード数x4

work=$PWD
pid=$$
date=`date '+%Y-%m-%d-%H-%M-%S-%N'`
tempdir=/tmp/metasearch_${date}-$pid
echo "tempdir: $tempdir"

input=`readlink -f "$1"` #$1=tmp/cf5a956d1de00aaea36b87346b21b4e8/cf5a956d1de00aaea36b87346b21b4e8.fq

sdir=$(dirname `readlink -f $0`)
source "$sdir"/config.sh

mkdir -p $tempdir/input
cd $tempdir

#入力ファイルがgz圧縮されているか調べて圧縮されていたらzcatそうでなければcatを使う
if [ `echo "$input"|grep [.]gz$|wc -l` = 1 ]; then
    zcat "$input"|head -n $n | awk 'NR%4==1{a=substr($1,2); if(a!~"/[1-4]$/"){a=a"/1"}; print ">"a} NR%4==2{print $0}' > input/input.fa
else
    cat "$input"|head -n $n | awk 'NR%4==1{a=substr($1,2); if(a!~"/[1-4]$/"){a=a"/1"}; print ">"a} NR%4==2{print $0}' > input/input.fa
fi

#BLAST->LCA解析を実行
#"$sdir"/metagenome~silva_SSU+LSU -c 8 -m 32 -d 50 -t 0.99 input

bitscore=100
top=0.99

set -eux
set -o pipefail

real_blastdb_path=`readlink -f "$blastdb_path"`
blastdb_dir=$(dirname "$real_blastdb_path")
$singularity_path exec -B ${tempdir} -B "$blastdb_dir" $sdir/ncbi_blast_2.13.0.sif blastn -num_threads 8 -db ${real_blastdb_path} -query ${tempdir}/input/input.fa -outfmt 6 -max_target_seqs 500 > $tempdir/blast.txt
cat $tempdir/blast.txt |
 awk -F'\t' '{split($1,arr,"/");
  if(arr[1]!=old){for(hit in data){temp[hit]=data[hit]["1"]+data[hit]["2"]}; PROCINFO["sorted_in"]="@val_num_desc"; for(hit in temp){print old"\t"hit"\t"temp[hit]}; old=arr[1]; delete data; delete temp};
  if(data[$2][arr[2]]==0){data[$2][arr[2]]=$12}}' |
 awk -F'\t' '$3>'$bitscore'{if(a[$1]==1){if($3>=topbit*'$top'){print $0}}else{a[$1]=1; topbit=$3; print $0}}' |
 awk -F'\t' 'FILENAME==ARGV[1]{name[$1]=$2} FILENAME==ARGV[2]{print name[$2]"\t"$0}' $blastdb_path.path /dev/stdin |
 awk -F'\t' '
function searchLCA(data,  i, j, res, res2, str, n, stopflag){
 for(i in data){
  if(n==0){n=split(i,res,";")}
  else{split(i,res2,";"); for(j in res){if(res[j]!=res2[j]){res[j]=""}}}
 }
 if(res[1]!=""){str=res[1]}
 else{
  #i: taxonomy path
  #葉緑体と植物の18Sは相同性が高いみたいなのでそれが混ざるときは葉緑体を優先させる
  chloroplast=0
  delete datachloro
  for(i in data){
   if(i~"^Bacteria;Cyanobacteria;Cyanobacteriia;Chloroplast;"){chloroplast++; datachloro[i]=1}
  }
  if(chloroplast>0){
   n2=0
   for(i in datachloro){
    if(n2==0){n2=split(i,res,";")}
    else{split(i,res2,";"); for(j in res){if(res[j]!=res2[j]){res[j]=""}}}
   }
  }
 }
 if(res[1]!=""){str=res[1]}
 else{
  str="unknown"; stopflag=1
 };
 for(i=2;i<=n;i++){if(stopflag==0 && res[i]!=""){str=str";"res[i]}else{stopflag=1}}
 return str;
}
{
 if($2!=old){if(old!=""){print searchLCA(data)"\t"oldstr}; delete data; data[$1]=1; old=$2; oldstr=$0}
 else{data[$1]=1}
}
END{if(length(data)>0){print searchLCA(data)"\t"oldstr}}
' > ${tempdir}/output.txt

awk -F'\t' 'BEGIN{print "id\tinput"} {cnt[$1]++} END{PROCINFO["sorted_in"]="@val_num_desc"; for(i in cnt){print i"\t"cnt[i]}}' ${tempdir}/output.txt > ${tempdir}/output.input
cp -rp ${tempdir}/output.input "$input".tsv

#mv input/*.ssu.blast.filtered.name.lca.cnt2.input "$input".tsv
#cd /tmp
#rm -rf $tempdir

if [ "`cat \"$input\".tsv |wc -l`" -lt 2 ]; then
    echo 'Error: No output!'
    exit 1
elif [ "`cat \"$input\".tsv |wc -l`" = 2 ] && [ "`tail -n 1 \"$input\".tsv|cut -f 1`" = "No Hit" ]; then
    echo 'Error: No hit in ribosomal database!'
    exit 1
fi

#awk -F'\t' '
# FILENAME==ARGV[1]{a["root;"$2]=$3}
# FILENAME==ARGV[2]{if($1=="id"){if(FNR>1){for(i in cnt){if(i!=""){print i"\t"cnt[i]}}}; print $0; delete cnt}else{cnt[a[$1]]+=$2}}
# END{for(i in cnt){if(i!=""){print i"\t"cnt[i]}}}
#' <(zcat "$sdir"/SILVA_132_SSU-LSU_Ref.fasta.name.species.gz) "$input".tsv > "$input".species.tsv

#awk -F'\t' '
# FILENAME==ARGV[1]{a["root;"$2]=$3}
# FILENAME==ARGV[2]{if($1=="id"){if(FNR>1){for(i in cnt){if(i!=""){print i"\t"cnt[i]}}}; print $0; delete cnt}else{cnt[a[$1]]+=$2}}
# END{for(i in cnt){if(i!=""){print i"\t"cnt[i]}}}
#' <(zcat "$sdir"/SILVA_132_SSU-LSU_Ref.fasta.name.genus.gz) "$input".tsv > "$input".genus.tsv

"$sdir"/calccor "$input".tsv "$sdir"/../data/db_merge
#"$sdir"/calccor "$input".genus.tsv "$maindir"/../data/db_genus_merge
#"$sdir"/calccor "$input".species.tsv "$maindir"/../data/db_species_merge
