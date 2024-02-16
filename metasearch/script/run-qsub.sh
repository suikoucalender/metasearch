#!/bin/bash

hash="$1" #cf5a956d1de00aaea36b87346b21b4e8
newfilename="$2" #tmp/cf5a956d1de00aaea36b87346b21b4e8/cf5a956d1de00aaea36b87346b21b4e8.fq
original_filename="$3" #y2022-group3-fish-16S-Kabayakisantaro.fq
email="$4" #akyoshita@g.ecc.u-tokyo.ac.jp

source ~/.bashrc
qsub -e ./qsub_log/e."$hash" -o ./qsub_log/o."$hash" -j y -N metasearch script/qsubsh8 script/metasearch_exec.sh "$newfilename" "$hash" "$original_filename" "$email"
#script/metasearch_exec.sh tmp/cf5a956d1de00aaea36b87346b21b4e8/cf5a956d1de00aaea36b87346b21b4e8.fq cf5a956d1de00aaea36b87346b21b4e8 y2022-group3-fish-16S-Kabayakisantaro.fq akyoshita@g.ecc.u-tokyo.ac.jp
