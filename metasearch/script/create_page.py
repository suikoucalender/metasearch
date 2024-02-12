# coding: utf-8
import sys
import os
from create_result import createResult 
from create_graph import createGraph
from create_krona_input import createKronaInput

argv = sys.argv

#tmp/hash値/hash値(.gz)で与えられる
uploaded_f_name = argv[1]

original_filename = argv[2]

#DBディレクトリの絶対パス
dbPath = argv[3]

rankList = [""] #, ".genus" , ".species"]

for rank in rankList:
	rank_with_underbar = rank.replace(".", "_")
	dbPath_new = dbPath + "db" + rank_with_underbar + "/"

	createResult(uploaded_f_name, original_filename, rank)
	createGraph(uploaded_f_name, dbPath_new, rank)

	createKronaInput(uploaded_f_name, rank)
