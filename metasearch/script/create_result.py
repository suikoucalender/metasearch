# coding: utf-8
import subprocess
from mako.template import Template
import os

def createResult(uploaded_f_name, original_filename, rank):
    #スクリプトの存在するディレクトリを取得
    dirname = os.path.dirname(__file__)

    #相関係数の結果ファイルとjaccard距離の結果ファイルを取得
    cor_result_f = open(uploaded_f_name + rank + ".tsv.result.correlation")
    jaccard_result_f = open(uploaded_f_name + rank + ".tsv.result.jaccard")

    cor_list = sra_parse(cor_result_f, dirname)
    jaccard_list = sra_parse(jaccard_result_f, dirname)

    #Makoテンプレートを取得
    t = Template(
        filename=dirname + "/templates/result.html",
        input_encoding="utf-8",
        output_encoding="utf-8",
        encoding_errors="replace"
    )

    #データをTemplateのHTMLファイルにRender
    render_data = {'cor_list': cor_list, 'jaccard_list': jaccard_list, 'original_filename': original_filename, 'rank': rank}
    html_content = t.render(**render_data)

    #tmp/Hash値を取得
    uploaded_f_dirname = os.path.dirname(uploaded_f_name)
    #Hash値を取得
    hash_data = uploaded_f_dirname.replace("tmp/","")

    #保存先のHTMLを開く(tmp/Hash値/Hash値/result.html)
    #print(uploaded_f_dirname + "/result" + rank + ".html")
    html_f_out = open(uploaded_f_dirname + "/result" + rank + ".html", "w")
    html_f_out.write(html_content.decode())

    html_f_out.close()
    cor_result_f.close()
    jaccard_result_f.close()


def sra_parse(result_f, dirname):
    #NCBIのSRAのURL
    url_template = "https://www.ncbi.nlm.nih.gov/sra/?term="

    #TemplateのHTMLにRenderするためのリストを作成
    out_list = []

    #相関関係の結果ファイルからSRAの情報を取得
    row_list = result_f.readlines()
    for row in row_list:
        try:
            row = row.split()
            #SRAのIDを取得
            sra_id = row[0].replace(".fastq","")

            #相関係数を取得
            cor = row[1]

            #SRAにアクセスするためのURLを生成
            url = url_template + sra_id

            #SRAのデータを格納するためのオブジェクトを作成
            sra_data = {}

            #HTMLをParse、サンプル情報などを取得
            cmd = ["bash", dirname + "/html_parse.sh", url]
            res = subprocess.check_output(cmd).decode().split("\n")
            study = res[0]
            sample = res[1]
            organism = res[2]

            #col_listにSRAの情報を格納
            sra_data["sra_id"] = sra_id
            sra_data["cor"] = cor
            sra_data["study"] = study
            sra_data["sample"] = sample
            sra_data["organism"] = organism

            out_list.append(sra_data)
        except:
            pass
    
    return out_list
