# coding: utf-8
import subprocess
from mako.template import Template
import os
import glob

#スクリプトの存在するディレクトリを取得
dirname = os.path.dirname(__file__)

def createGraph(uploaded_f_name, dbPath, rank):
    #Queryの組成ファイルを取得
    query_comp_f = open(uploaded_f_name + rank + ".tsv")

    
    #Queryの各組成の量を取得
    query_comp_top, query_comp_all = calc_comp(query_comp_f)

    #相関係数の結果ファイルとjaccard距離の結果ファイルを取得
    cor_result_f = open(uploaded_f_name + rank + ".tsv.result.correlation")
    jaccard_result_f = open(uploaded_f_name + rank + ".tsv.result.jaccard")

    #相関係数の結果に記載された各SRAのIDに対して、組成を取得
    cor_comp_object = get_comp(cor_result_f, dbPath)

    #jaccard距離の結果に記載された各SRAのIDに対して、組成を取得
    jaccard_comp_object = get_comp(jaccard_result_f, dbPath)

    #全てのオブジェクトに含まれる生物種の一覧を取得(積集合)
    cor_livings = get_livings(cor_comp_object) #相関係数の生物種
    jaccard_livings = get_livings(jaccard_comp_object) #jaccard係数の生物種
    
    #Queryと相関係数に含まれる生物種
    query_cor_livings = query_comp_top + cor_livings
    query_cor_livings = sorted(set(query_cor_livings), key=query_cor_livings.index)

    #Queryとjaccard係数に含まれる生物種
    query_jaccard_livings = query_comp_top + jaccard_livings
    query_jaccard_livings = sorted(set(query_jaccard_livings), key=query_jaccard_livings.index) 

    #Google Chart APIに渡すデータテーブルを作成
    cor_dataTable = createDataTable(query_cor_livings, query_comp_all, cor_comp_object)
    jaccard_dataTable = createDataTable(query_jaccard_livings, query_comp_all, jaccard_comp_object)


    #Makoテンプレートを取得
    t = Template(
        filename=dirname + "/templates/graphs.html",
        input_encoding="utf-8",
        output_encoding="utf-8",
        encoding_errors="replace"
    )
    
    #tmp/Hash値を取得
    uploaded_f_dirname = os.path.dirname(uploaded_f_name)
    #Hash値を取得
    hash_data = uploaded_f_dirname.replace("tmp/","")

    #保存先のHTMLを開く(tmp/Hash値/Hash値/graphs_cor.html)
    cor_html_f_out = open(uploaded_f_dirname + "/" + hash_data + "/graphs_cor" + rank + ".html", "w")
    jaccard_html_f_out = open(uploaded_f_dirname + "/" + hash_data + "/graphs_jaccard" + rank + ".html", "w")
    
    #HTMLテンプレートにRenderする
    renderHtml(t, cor_dataTable, cor_html_f_out)
    renderHtml(t, jaccard_dataTable, jaccard_html_f_out)

    query_comp_f.close()
    cor_result_f.close()
    jaccard_result_f.close()

#resultファイルから各SRAのIDに関して組成のオブジェクトを作成し、それらを格納したリストを返すための関数
def get_comp(result_f, dbPath):
    comp_object = {}

    rows = result_f.readlines()
    for row in rows:
        #SRAのIDを取得
        sra_id = row.split()[0].replace(".fastq","")
        
        #SRAのIDに対応する組成ファイルのパスを取得
        try:
            # comp_f_path = glob.glob("/home/yoshitake/yoshitake/db/SraId_*/" + sra_id + ".fastq.fasta.ssu.blast.filtered.name.lca.cnt2.input")[0]
            comp_f_path = glob.glob(dbPath + "SraId_*/" + sra_id + ".fastq.fasta.ssu.blast.filtered.name.lca.cnt2.input")[0]
            print(comp_f_path)
        except:
            continue
        
        #SRAのIDに対応する組成ファイルを取得
        comp_f = open(comp_f_path)

        #組成を取得
        comp_top, comp_all = calc_comp(comp_f)

        #リストに格納
        comp_object[sra_id] = [comp_top, comp_all]

        comp_f.close()
    
    return comp_object

#組成のオブジェクトを取得するための関数
def calc_comp(comp_f):
    row_comp = comp_f.readlines()

    comp_top = []
    comp_all = {}
    all = 0
    #各生物種、スコアに関してTop100を取得。101位以降は"other"としてカウント。"No Hit"はカウントしない
    for i, row in enumerate(row_comp):
        row = row.split("\t")
        if row[0] == "No Hit":
            continue
        if i == 0:
            continue
        elif i < 101:
            comp_top.append(row[0].replace("root;",""))
            comp_all[row[0].replace("root;","")] = int(row[1].replace("\n",""))
        else:
            comp_all[row[0].replace("root;","")] = int(row[1].replace("\n",""))
        all += int(row[1].replace("\n",""))
    comp_all["all"] = all
    comp_f.close()
    return comp_top, comp_all

#各SRAのIDの組成のオブジェクトから成るリストから生物種のみのリストを返す関数
def get_livings(comp_object):
    livings = []
    for comp in comp_object.values():
        livings.extend(comp[0])
    return livings

#Google Chart APIに渡すデータテーブルを作成するための関数
def createDataTable(livings,query_comp_all,comp_object):
    dataTable = [] #Google Chart APIに渡すデータテーブル
    dataTable.append([""] + livings + ["others"]) #データテーブルに生物種を追加
    
    #各生物種に関して、Queryの組成にどれだけ含まれているかを算出。(含まない場合は0)
    query_data = []
    query_others = query_comp_all["all"]
    for living in livings:
        query_data.append(query_comp_all.get(living,0))
        query_others = query_others - query_comp_all.get(living,0)
    dataTable.append(["Query"] + query_data + [query_others])

    #各生物種に関して、各組成オブジェクトにどれだけ含まれているかを算出
    for sra_id in comp_object.keys():
        data = []
        others = comp_object[sra_id][1]["all"]
        for living in livings:
            data.append(comp_object[sra_id][1].get(living,0))
            others = others - comp_object[sra_id][1].get(living,0)
        dataTable.append([sra_id] + data + [others])

    return dataTable

def renderHtml(t,dataTable,html_f_out):
    #データをTemplateのHTMLファイルにRender
    render_data = {'dataTable': dataTable}
    html_content = t.render(**render_data)

    #ファイルに書き込み
    html_f_out.write(html_content.decode())
    html_f_out.close()


