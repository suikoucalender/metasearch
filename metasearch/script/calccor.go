package main

import (
    "fmt"
    "encoding/csv"
    "io"
    "os"
    "math"
    "strconv"
    "io/ioutil"
    "sync"
    "sort"
    "strings"
    "path/filepath"
)

type ResMap struct {
    Name string
    Value float64
}

type Entry struct {
    name  string
    value float64
}
type List []Entry

func (l List) Len() int {
    return len(l)
}

func (l List) Swap(i, j int) {
    l[i], l[j] = l[j], l[i]
}

func (l List) Less(i, j int) bool {
    if l[i].value == l[j].value {
        return (l[i].name > l[j].name)
    } else {
        return (l[i].value > l[j].value)
    }
}

func main() {
    //新しいMutexオブジェクトを作成し、そのポインタをmutex変数に割り当てることで、以降このmutexを通じてロックとアンロックの操作を行うことができる
    mutex := &sync.Mutex{}
    if len(os.Args) < 3{
        fmt.Println(os.Args[0],"[-t num_threads(8)] [-n num_hits(10)] <input tsv> <db dir>")
        return
    }

    inputtsv:=""
    inputdir:=""
    num_threads:=8
    num_hits:=10
    tempflagt:=false
    tempflagn:=false
    tempflagcount:=0
      for i, arg := range os.Args {
        if i > 0 {
            if arg == "-t" {
                  tempflagt=true
            }else if arg == "-n" {
                tempflagn=true
            }else if tempflagt {
                tempflagt=false
                temp_num_threads, err := strconv.Atoi(arg) //string convert, ASCII to integer
                if err != nil {
                    fmt.Printf("%s\n", err.Error())
                    return
                }
                num_threads = temp_num_threads
            }else if tempflagn {
                tempflagn=false
                temp_num_hits, err := strconv.Atoi(arg)
                if err != nil {
                    fmt.Printf("%s\n", err.Error())
                    return
                }
                num_hits = temp_num_hits
            }else if tempflagcount == 0 {
                tempflagcount++
                inputtsv=arg
            }else if tempflagcount == 1 {
                tempflagcount++
                inputdir=arg
            } else {
                  fmt.Printf("Unknown args %d: %s\n", i, os.Args[i])
                return
            }
        }
      }
    if tempflagcount < 2{
        fmt.Println(os.Args[0],"[-t num_threads(8)] [-n num_hits(10)] <input tsv> <db dir>")
        return
    }

    _, a1, afp1 := userCsvNewReader(inputtsv)  //userCsvNewReaderは一行目のヘッダーからサンプルの名前と、二行目以降の種名：リード数をマップに詰めて返す
    delete(a1,"No Hit")
    delete(afp1,"No Hit")
    if len(a1)==0{
        fmt.Fprintf(os.Stderr, "No Hit\n")
        os.Exit(1)
    }
    //fmt.Println(a1)
    var wg sync.WaitGroup
    res := make(chan int, num_threads)
    list := List{}
    listlog := List{}
    list_unifrac := List{}
    list_weighted_jaccard := List{}
    list_weighted_jaccardlog := List{}
    listfp := List{}
    listfplog := List{}
    list_unifracfp := List{}
    list_weighted_jaccardfp := List{}
    list_weighted_jaccardlogfp := List{}

    paths:=dirwalk(inputdir)
    for _, path := range paths {
        //fmt.Println(path)
        wg.Add(1)
        go func(fname string){
            res<-1
            //deferキーワードを使用して無名関数を登録します。この無名関数は、外側の無名関数が終了する直前（リターンする直前）に実行されます。
            defer func(){
                <-res
                wg.Done()
            }()
            dbcont:=useIoutilReadFile(fname)  //ファイルを開いて文字列型で取得
            fn2s, a2s, a2fps:=splitCsv(dbcont) //1まとめの文字列をSRRごとに分けて連想配列として返す
            for i, a2 := range a2s {
                //fmt.Println(i)
                fn2 := fn2s[i]
                //va2:=values(a2)
                delete(a2,"No Hit")
                if len(a2)==0{
                    continue
                }
                va1:=[]int{}  //int配列の初期化
                va2:=[]int{}
                //入力ファイルa1側とDB側a2でそれぞれにしか値がない場合に反対側に0を加える処理
                for k1, v1 := range a1 {
                    va1 = append(va1, v1)
                    if v2, ok := a2[k1]; ok{
                        va2 = append(va2, v2)
                    }else{
                        va2 = append(va2, 0)
                    }
                }
                for k2, v2 := range a2 {
                    if _, ok := a1[k2]; !ok{
                        va1 = append(va1, 0)
                        va2 = append(va2, v2)
                    }
                }
                //fmt.Println(fn2,unifrac(a1,a2))

                af:=[]float64{}
                bf:=[]float64{}
                for i := range va1{
                    af = append(af, float64(va1[i]))
                    bf = append(bf, float64(va2[i]))
                }
                p:=Pearson(af, bf)
                //fmt.Println(fn2,p,len(va1),len(va2))

                aflog:=[]float64{}
                bflog:=[]float64{}
                for i := range va1{
                    //fmt.Println(va1[i], (math.Log(float64(va1[i])+0.25)-math.Log(0.25))/math.Log(2))
                    aflog = append(aflog, (math.Log(float64(va1[i])+0.25)-math.Log(0.25))/math.Log(2))
                    bflog = append(bflog, (math.Log(float64(va2[i])+0.25)-math.Log(0.25))/math.Log(2))
                }
                plog:=Pearson(aflog, bflog)

                wj:=weighted_jaccard(af, bf)
                wjlog:=weighted_jaccard(aflog, bflog)

                mutex.Lock()
                //list=append(list,Entry{fn2,p})
                list=add_to_list(fn2, p, list, num_hits) //num_hitsは出力ヒット数
                listlog=add_to_list(fn2, plog, listlog, num_hits)
                //list_unifrac=append(list_unifrac,Entry{fn2,unifrac(a1,a2)})
                //list_unifrac=add_to_list_unifrac(fn2, unifrac(a1,a2), list_unifrac, num_hits)
                list_unifrac=add_to_list_unifrac(fn2, jaccard(a1,a2), list_unifrac, num_hits)
                list_weighted_jaccard=add_to_list(fn2, wj, list_weighted_jaccard, num_hits)
                list_weighted_jaccardlog=add_to_list(fn2, wjlog, list_weighted_jaccardlog, num_hits)
                mutex.Unlock()
            }
            for i, afp2 := range a2fps {
                fn2 := fn2s[i]
                delete(afp2,"No Hit")
                if len(afp2)==0{
                    continue
                }
                va1:=[]int{}  //int配列の初期化
                va2:=[]int{}
                //入力ファイルa1側とDB側a2でそれぞれにしか値がない場合に反対側に0を加える処理
                for k1, v1 := range afp1 {
                    va1 = append(va1, v1)
                    if v2, ok := afp2[k1]; ok{
                        va2 = append(va2, v2)
                    }else{
                        va2 = append(va2, 0)
                    }
                }
                for k2, v2 := range afp2 {
                    if _, ok := afp1[k2]; !ok{
                        va1 = append(va1, 0)
                        va2 = append(va2, v2)
                    }
                }
                af:=[]float64{}
                bf:=[]float64{}
                for i := range va1{
                    af = append(af, float64(va1[i]))
                    bf = append(bf, float64(va2[i]))
                }
                p:=Pearson(af, bf)

                aflog:=[]float64{}
                bflog:=[]float64{}
                for i := range va1{
                    aflog = append(aflog, (math.Log(float64(va1[i])+0.25)-math.Log(0.25))/math.Log(2))
                    bflog = append(bflog, (math.Log(float64(va2[i])+0.25)-math.Log(0.25))/math.Log(2))
                }
                plog:=Pearson(aflog, bflog)

                wj:=weighted_jaccard(af, bf)
                wjlog:=weighted_jaccard(aflog, bflog)

                mutex.Lock()
                listfp=add_to_list(fn2, p, listfp, num_hits)
                listfplog=add_to_list(fn2, plog, listfplog, num_hits)
                list_unifracfp=add_to_list_unifrac(fn2, jaccard(afp1,afp2), list_unifracfp, num_hits)
                list_weighted_jaccardfp=add_to_list(fn2, wj, list_weighted_jaccardfp, num_hits)
                list_weighted_jaccardlogfp=add_to_list(fn2, wjlog, list_weighted_jaccardlogfp, num_hits)
                mutex.Unlock()
            }
        }(path)
    }
    wg.Wait()
    save_list_to_file(list, inputtsv+".result.correlation", num_hits)
    save_list_to_file(listlog, inputtsv+".result.correlation.log", num_hits)
    save_list_to_file(list_weighted_jaccard, inputtsv+".result.weighted_jaccard", num_hits)
    save_list_to_file(list_weighted_jaccardlog, inputtsv+".result.weighted_jaccard.log", num_hits)

    //fmt.Println("")
    sort.Slice(list_unifrac,func(i, j int) bool { return list_unifrac[i].value < list_unifrac[j].value })
    b2 := []byte{}
    for i, pk := range list_unifrac{
        if i >= num_hits {break}
        //fmt.Println(pk.name+"\t"+strconv.FormatFloat(pk.value, 'f', -1, 64))
        ll := []byte(pk.name+"\t"+strconv.FormatFloat(pk.value, 'f', -1, 64)+"\n")
        for _, l := range ll {b2 = append(b2, l)}
    }
    //errw2 := ioutil.WriteFile(inputtsv+".result.unifrac", b2, 0666)
    errw2 := ioutil.WriteFile(inputtsv+".result.jaccard", b2, 0666)
    if errw2 != nil {
        fmt.Println(os.Stderr, errw2)
           os.Exit(1)
    }

    save_list_to_file(listfp, inputtsv+".result.correlation.fullnodes", num_hits)
    save_list_to_file(listfplog, inputtsv+".result.correlation.fullnodes.log", num_hits)
    save_list_to_file(list_weighted_jaccardfp, inputtsv+".result.weighted_jaccard.fullnodes", num_hits)
    save_list_to_file(list_weighted_jaccardlogfp, inputtsv+".result.weighted_jaccard.fullnodes.log", num_hits)

    sort.Slice(list_unifracfp,func(i, j int) bool { return list_unifracfp[i].value < list_unifracfp[j].value })
    bfp2 := []byte{}
    for i, pk := range list_unifracfp{
        if i >= num_hits {break}
        ll := []byte(pk.name+"\t"+strconv.FormatFloat(pk.value, 'f', -1, 64)+"\n")
        for _, l := range ll {bfp2 = append(bfp2, l)}
    }
    errwfp2 := ioutil.WriteFile(inputtsv+".result.jaccard.fullnodes", bfp2, 0666)
    if errwfp2 != nil {
        fmt.Println(os.Stderr, errwfp2)
           os.Exit(1)
    }

}

func save_list_to_file(list List, filename string, num_hits int){
    sort.Sort(list)
    //fmt.Println(list)
    b := []byte{}
    for i, pk := range list{
        if i >= num_hits {break}
        //fmt.Println(pk.name+"\t"+strconv.FormatFloat(pk.value, 'f', -1, 64))
        ll := []byte(pk.name+"\t"+strconv.FormatFloat(pk.value, 'f', -1, 64)+"\n")
        for _, l := range ll {b = append(b, l)}
    }
    errw := ioutil.WriteFile(filename, b, 0666)
    if errw != nil {
        fmt.Println(os.Stderr, errw)
        os.Exit(1)
    }
}

func add_to_list(name string, value float64, list List, num_hits int) List{
    n:= num_hits
    if len(list)>n{
        if value > list[n].value{
            list[n]=Entry{name, value}
            sort.Sort(list)
        }
    }else{
        list = append(list, Entry{name, value})
        sort.Sort(list)
    }
    return list
}

func add_to_list_unifrac(name string, value float64, list_unifrac List, num_hits int) List{
    n:= num_hits
    if len(list_unifrac)>n{
        if value < list_unifrac[n].value{
            list_unifrac[n]=Entry{name, value}
            sort.Slice(list_unifrac,func(i, j int) bool { return list_unifrac[i].value < list_unifrac[j].value })
        }
    }else{
        list_unifrac = append(list_unifrac, Entry{name, value})
        sort.Sort(list_unifrac)
    }
    return list_unifrac
}

func weighted_jaccard(x []float64, y []float64) (float64){
    sum_min := 0.0
    sum_max := 0.0
    for i, _ := range x{
        sum_min += math.Min(x[i], y[i])
        sum_max += math.Max(x[i], y[i])
    }
    if sum_max > 0{
        return sum_min / sum_max
    }else{
        return 0
    }
}

func jaccard(x map[string]int, y map[string]int) (float64){
    xsum := 0.0
    ysum := 0.0
    for _, xval := range x{
        xsum += float64(xval)
    }
    for _, yval := range y{
        ysum += float64(yval)
    }
    sx := 0.0
    sy := 0.0
    sxy := 0.0
    for xkey, xval := range x{
        if float64(xval) >= xsum * 0.01{
            sx++
            _, ok := y[xkey]
            if ok && float64(y[xkey]) >= ysum * 0.01{
                sxy++
            }
        }
    }
    for _, yval := range y{
        if float64(yval) >= ysum * 0.01{
            sy++
        }
    }
    return 1-(sxy/(sx+sy-sxy))
}

//https://mothur.org/wiki/weighted_unifrac_algorithm/を参考にしたけど、結局X, Y2つの微生物叢のツリーのリード相対量を引いたツリーを作って、エッジｘ2つのノードの値の足し算/2をノード総当りで計算
func unifrac(x map[string]int, y map[string]int) (float64){
    dist_dif := 0.0
    dist_total := 0.0
    xsum := 0
    ysum := 0
    for _, xval := range x{
        xsum += xval
    }
    for _, yval := range y{
        ysum += yval
    }
    for xkey, xval := range x{
        yval_xkey:=0
        _, ok := y[xkey]
        if ok{
            yval_xkey = y[xkey]
        }
        for ykey, yval := range y{
            xval_ykey:=0
            _, ok := x[ykey]
            if ok{
                xval_ykey = x[ykey]
            }
            total_val, different_val := unifracElement(xkey, ykey)
            dist_dif+=float64(different_val)*(math.Abs(float64(xval)/float64(xsum)-float64(yval_xkey)/float64(ysum))+math.Abs(float64(xval_ykey)/float64(xsum)-float64(yval)/float64(ysum)))/2
            dist_total+=float64(total_val)
        }
    }
    return dist_dif/dist_total
}

func unifracElement(xpath string, ypath string) (int, int){
    xitems := strings.Split(xpath, ";")
    yitems := strings.Split(ypath, ";")
    xlen := len(xitems)
    ylen := len(yitems)
    xfin := 0
    for i, _ := range xitems {
        xfin=i
        if ylen <= i{
            break
        }else{
            if xitems[i] != yitems[i]{
                break
            }
        }
        xfin=i+1
    }
    //fmt.Println(xlen+ylen, (xlen+ylen-2*xfin),xpath,ypath)
    return xlen+ylen, (xlen+ylen-2*xfin)
}

func dirwalk(dir string) []string {
   files, err := ioutil.ReadDir(dir)
   if err != nil {
       fmt.Println("dirwalk error!")
       panic(err)
   }
   var paths []string
   for _, file := range files {
       if file.IsDir() {
           paths = append(paths, dirwalk(filepath.Join(dir, file.Name()))...)
           continue
       }
       paths = append(paths, filepath.Join(dir, file.Name()))
   }
   return paths
}

//ファイルの内容を文字列で返す
func useIoutilReadFile(fileName string) string{
    bytes, err := ioutil.ReadFile(fileName)
    if err != nil {
        fmt.Println("useIoutilReadFile error!")
        panic(err)
    }
    return string(bytes)
}

//連結されたDBファイルの1つのバイト情報を受け取り、SRRファイルごとに分割された連想配列として返す
func splitCsv(dbcont string) ([]string, []map[string]int, []map[string]int){
    myslicemap := []map[string]int{}
    mymap := map[string]int{}
    myslicemapfp := []map[string]int{}
    mymapfp := map[string]int{}
    name := ""
    nameslice := []string{}

    slice := strings.Split(dbcont, "\n")
    //fmt.Println(len(slice))
    for _, str := range slice {
        //fmt.Println("item: ",name,i,str)
        item:= strings.Split(str, "\t")
        if len(item) > 1{
            if item[0]=="id"{
                if name!=""{
                    //ループで2データ目以降で前のデータが残っている場合の処理
                    //temp_mapを作ってmymapをコピーしておかないとmymapをappendするだけだと、この後mymapを初期化するのでダメ？
                    temp_map := map[string]int{}
                    for key, value := range mymap {
                        temp_map[key] = value
                    }
                    myslicemap = append(myslicemap, temp_map)

                    temp_mapfp := map[string]int{}
                    for key, value := range mymapfp {
                        temp_mapfp[key] = value
                    }
                    myslicemapfp = append(myslicemapfp, temp_mapfp)
                }
                mymap = map[string]int{}
                mymapfp = map[string]int{}
                name=item[1]
                nameslice = append(nameslice, name)
            }else{
                //id行以外の場合
                f, err := strconv.Atoi(item[1])
                if err != nil {
                    fmt.Println("splitCsv error!")
                    fmt.Printf("%s\n", err.Error())
                    panic(err)
                }
                mymap[item[0]]=f

                //taxonomy pathの親ノードにも加算した値を作っていく
                tpnodes:=strings.Split(item[0],";")
                var builder strings.Builder //GOでは文字列は変更不可能だそうなのでBuilderを作って追加していく
                for i, tpnode := range tpnodes {
                    if i != 0 {
                        builder.WriteString(";")
                    }
                    builder.WriteString(tpnode)
                    mymapfp[builder.String()]+=f //nodeがまだ追加されていなくても、されていても大丈夫
                }
            }
        }
    }
    //fmt.Println("test")
    myslicemap = append(myslicemap, mymap)
    myslicemapfp = append(myslicemapfp, mymapfp)
    return nameslice, myslicemap, myslicemapfp
}

func userCsvNewReader(fileName string) (string, map[string]int, map[string]int){
    fp, err := os.Open(fileName)
    if err != nil {
        fmt.Println("userCsvNewReader error!")
        panic(err)
    }
    defer fp.Close()

    array := map[string]int{} //taxonomy pathのノードそのまま
    arrayfp := map[string]int{} //taxonomy pathの親のノードにも加算した結果
    name := ""
    reader := csv.NewReader(fp)
    reader.Comma = '\t'
    reader.FieldsPerRecord = 2 // 各行のフィールド数。多くても少なくてもエラーになる
    reader.LazyQuotes = true
    //reader.ReuseRecord = true // true の場合は、Read で戻ってくるスライスを次回再利用する。パフォーマンスが上がる
    cnti:=0
    for {
        record, err := reader.Read()
        if err == io.EOF {
            break
        } else if err != nil {
            fmt.Println("userCsvNewReader2 error!")
            panic(err)
        }
        //fmt.Println(record[0])
        cnti++
        if cnti == 1 {
            name=record[1]
        }else {
            f, err := strconv.Atoi(record[1])
            if err != nil {
                fmt.Printf("%s\n", err.Error())
                return name, array, arrayfp
            }
            array[record[0]]=f
            //fmt.Println(f)
            //taxonomy pathの親ノードにも加算した値を作っていく
            tpnodes:=strings.Split(record[0],";")
            var builder strings.Builder //GOでは文字列は変更不可能だそうなのでBuilderを作って追加していく
            for i, tpnode := range tpnodes {
                if i != 0 {
                    builder.WriteString(";")
                }
                builder.WriteString(tpnode)
                arrayfp[builder.String()]+=f //nodeがまだ追加されていなくても、されていても大丈夫
            }
        }
    }
    return name, array, arrayfp
}
func Pearson(a, b []float64) float64 {

    if len(a) != len(b) {
        return 0
        panic("len(a) != len(b)")
    }
    var abar, bbar float64
    var n int
    for i := range a {
        //if !math.IsNaN(a[i]) && !math.IsNaN(b[i]) {
            abar += a[i]
            bbar += b[i]
            n++
        //}
    }
    nf := float64(n)
    abar, bbar = abar/nf, bbar/nf

    var numerator float64
    var sumAA, sumBB float64

    for i := range a {
        //if !math.IsNaN(a[i]) && !math.IsNaN(b[i]) {
            numerator += (a[i] - abar) * (b[i] - bbar)
            sumAA += (a[i] - abar) * (a[i] - abar)
            sumBB += (b[i] - bbar) * (b[i] - bbar)
        //}
    }

    return numerator / (math.Sqrt(sumAA) * math.Sqrt(sumBB))
}

func keys(m map[string]int) []string {
    ks := []string{}
    for k, _ := range m {
        ks = append(ks, k)
    }
    return ks
}

func values(m map[string]int) []int {
    vs := []int{}
    for _, v := range m {
        vs = append(vs, v)
    }
    return vs
}

func to_a(m map[string]int) []interface{} {
    a := []interface{}{}
    for k, v := range m {
        a = append(a, []interface{}{k, v})
    }
    return a
}

func indexes(m map[string]int, keys []string) []int {
    vs := []int{}
    for _, k := range keys {
        vs = append(vs, m[k])
    }
    return vs
}
