#!/bin/bash

#tmp/hash値/hash値(.gz)
newfilename=$1
hash=$2
original_filename=$3
usr_email=$4

#DBディレクトリが格納されているディレクトリの絶対パス(必ず'/'を入れる)
dbPath=/usr/local/yoshitake/

script/run-silva-cor.sh $newfilename 

/home/yoshitake/tool/singularity-3.5.2/bin/singularity run -B $dbPath script/python3_env_mako_installed.sif python script/create_page.py $newfilename $original_filename $dbPath

for class in "" .genus .species
do
	/home/yoshitake/tool/singularity-3.5.2/bin/singularity run script/krona_v2.7.1_cv1.sif ktImportText tmp/${hash}/result${class}.kraken -o tmp/${hash}/${hash}/krona_out${class}.html
done

/usr/bin/cp -r tmp/${hash}/${hash} public/
url=` cat config/config.json | grep "url" | sed -r 's/^[^:]*:(.*)$/\1/' | sed 's/\"//g' | sed "s/,//g" | sed 's/ //g'`

/home/yoshitake/tool/singularity-3.5.2/bin/singularity run script/python3_env_mako_installed.sif python script/send_mail.py ${url}/${hash}/ ${usr_email} ${original_filename}