# coding: utf-8
import os

def createKronaInput(uploaded_f_name, rank):
    #Queryの組成ファイルを取得
    query_comp_f = open(uploaded_f_name + rank + ".tsv")
    #tmp/Hash値を取得
    uploaded_f_dirname = os.path.dirname(uploaded_f_name)

    #保存先のファイルのパスを指定(tmp/Hash値/result.kraken)
    krona_input_f = open(uploaded_f_dirname + "/result" + rank + ".kraken", "w")

    rows = query_comp_f.readlines()
    for i,row in enumerate(rows):
        #Headerの行では処理をしない
        if i == 0:
            continue
        row = row.split("\t")
        #No Hitの行では処理をしない
        if row[0] == "No Hit":
            continue
        
        text = row[1].replace("\n","\t") #スコアを書き込み
        taxanomy = row[0].split(";")
        taxanomy_str = "\t".join(taxanomy)
        text += taxanomy_str + "\n"

        krona_input_f.write(text)
    
    krona_input_f.close()