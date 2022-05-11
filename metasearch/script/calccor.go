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
                temp_num_threads, err := strconv.Atoi(arg)
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

    _, a1 := userCsvNewReader(inputtsv)  //userCsvNewReaderは一行目のヘッダーからサンプルの名前と、二行目以降の種名：リード数をマップに詰めて返す
    delete(a1,"No Hit")
    if len(a1)==0{
        fmt.Fprintf(os.Stderr, "No Hit\n")
        os.Exit(1)
    }
    //fmt.Println(a1)
    var wg sync.WaitGroup
    res := make(chan int, num_threads)
    list := List{}
    list_unifrac := List{}

    paths:=dirwalk(inputdir)
    for _, path := range paths {
        //fmt.Println(path)
        wg.Add(1)
        go func(fname string){
            res<-1
            defer func(){
                <-res
                wg.Done()
            }()
            //fn2, a2:=userCsvNewReader(fname) //読み込むファイルが大きい場合はこちら
            dbcont:=useIoutilReadFile(fname)  //ファイルが小さい場合
            //fn2, a2:=splitCsv(dbcont)                        //ファイルが小さい場合(old)
            fn2s, a2s:=splitCsv(dbcont)
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
                p:=Pearson(va1, va2)
                //fmt.Println(fn2,p,len(va1),len(va2))
                mutex.Lock()
                //list=append(list,Entry{fn2,p})
                list=add_to_list(fn2, p, list, num_hits)
                //list_unifrac=append(list_unifrac,Entry{fn2,unifrac(a1,a2)})
                //list_unifrac=add_to_list_unifrac(fn2, unifrac(a1,a2), list_unifrac, num_hits)
                list_unifrac=add_to_list_unifrac(fn2, jaccard(a1,a2), list_unifrac, num_hits)
                mutex.Unlock()
            }
        }(path)
    }
    wg.Wait()
    sort.Sort(list)
    //fmt.Println(list)
    b := []byte{}
    for i, pk := range list{
        if i > num_hits {break}
        //fmt.Println(pk.name+"\t"+strconv.FormatFloat(pk.value, 'f', -1, 64))
        ll := []byte(pk.name+"\t"+strconv.FormatFloat(pk.value, 'f', -1, 64)+"\n")
        for _, l := range ll {b = append(b, l)}
    }
    errw := ioutil.WriteFile(inputtsv+".result.correlation", b, 0666)
    if errw != nil {
        fmt.Println(os.Stderr, errw)
        os.Exit(1)
    }

    //fmt.Println("")
    sort.Slice(list_unifrac,func(i, j int) bool { return list_unifrac[i].value < list_unifrac[j].value })
    b2 := []byte{}
    for i, pk := range list_unifrac{
        if i > num_hits {break}
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

func useIoutilReadFile(fileName string) string{
    bytes, err := ioutil.ReadFile(fileName)
    if err != nil {
        fmt.Println("useIoutilReadFile error!")
        panic(err)
    }
    return string(bytes)
}

func splitCsv(dbcont string) ([]string, []map[string]int){
    myslicemap := []map[string]int{}
    mymap := map[string]int{}
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
                temp_map := map[string]int{}
                for key, value := range mymap {
                    temp_map[key] = value
                }
                myslicemap = append(myslicemap, temp_map)
            }
            mymap = map[string]int{}
            name=item[1]
            nameslice = append(nameslice, name)
        }else{
            f, err := strconv.Atoi(item[1])
            if err != nil {
                fmt.Println("splitCsv error!")
                fmt.Printf("%s\n", err.Error())
                panic(err)
            }
            mymap[item[0]]=f
        }
        }
    }
    //fmt.Println("test")
    myslicemap = append(myslicemap, mymap)
    return nameslice, myslicemap
}

func userCsvNewReader(fileName string) (string, map[string]int){
    fp, err := os.Open(fileName)
    if err != nil {
        fmt.Println("userCsvNewReader error!")
        panic(err)
    }
    defer fp.Close()

    array := map[string]int{}
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
                return name, array
            }
            array[record[0]]=f
            //fmt.Println(f)
        }
    }
    return name, array
}
func Pearson(a, b []int) float64 {

    if len(a) != len(b) {
        return 0
        panic("len(a) != len(b)")
    }
    af:=[]float64{} 
    bf:=[]float64{}
    for i := range a{
        af = append(af, float64(a[i]))
        bf = append(bf, float64(b[i]))
    }
    var abar, bbar float64
    var n int
    for i := range a {
        //if !math.IsNaN(a[i]) && !math.IsNaN(b[i]) {
            abar += af[i]
            bbar += bf[i]
            n++
        //}
    }
    nf := float64(n)
    abar, bbar = abar/nf, bbar/nf

    var numerator float64
    var sumAA, sumBB float64

    for i := range a {
        //if !math.IsNaN(a[i]) && !math.IsNaN(b[i]) {
            numerator += (af[i] - abar) * (bf[i] - bbar)
            sumAA += (af[i] - abar) * (af[i] - abar)
            sumBB += (bf[i] - bbar) * (bf[i] - bbar)
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
