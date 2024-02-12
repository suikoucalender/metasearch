#!/bin/bash

#tmp/hash値/hash値(.gz)
newfilename=$1
hash=$2
original_filename=$3
usr_email=$4

sdir=$(dirname `readlink -f $0`)
#source "$maindir"/settings.sh
source ~/.bashrc

#DBディレクトリが格納されているディレクトリの絶対パス(必ず'/'を入れる)
#dbPath=/usr/local/yoshitake/
dbPath=$(dirname $(readlink -f $sdir/../data/db))

script/run-silva-cor.sh $newfilename

#Singularityのイメージがなければ、githubのリリースから取ってくる。ファイルサイズが大きいのでソースコードには含められない。
if [ ! -e "${sdir}/python3_env_mako_installed.sif" ]; then
 wget -O "${sdir}/python3_env_mako_installed.sif" https://github.com/suikoucalender/metasearch/releases/download/0.1/python3_env_mako_installed.sif
fi
if [ ! -e "${sdir}/krona_v2.7.1_cv1.sif" ]; then
 wget -O "${sdir}/krona_v2.7.1_cv1.sif" https://github.com/suikoucalender/metasearch/releases/download/0.1/krona_v2.7.1_cv1.sif
fi

singularity run -B $dbPath $sdir/python3_env_mako_installed.sif python script/create_page.py $newfilename $original_filename $dbPath

for class in "" .genus .species
do
	singularity run $sdir/krona_v2.7.1_cv1.sif ktImportText tmp/${hash}/result${class}.kraken -o tmp/${hash}/${hash}/krona_out${class}.html
done

cp -r tmp/${hash}/${hash} public/
url=` cat config/config.json | grep "url" | sed -r 's/^[^:]*:(.*)$/\1/' | sed 's/\"//g' | sed "s/,//g" | sed 's/ //g'`

singularity run $sdir/python3_env_mako_installed.sif python script/send_mail.py ${url}/${hash}/ ${usr_email} ${original_filename}
