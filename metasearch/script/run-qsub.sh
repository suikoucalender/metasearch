#!/bin/bash

hash="$1"
newfilename="$2"
original_filename="$3"
email="$4"

source ~/.bashrc
qsub -e ./qsub_log/e." + hash + " -o ./qsub_log/o." + hash + " -j y -N metasearch script/qsubsh8 script/metasearch_exec.sh "$newfilename" "$hash" "$original_filename" "$email"
