# find *.h > /dev/null 2>&1
url=$1

# sleep 1 ;curl -Ss $url |grep "Submitted by:" |awk '{print substr($0, index($0, "Submitted by:"))}'|awk '{print substr($0, 1, index($0, "</span>")-1)}' | sed -e 's/Submitted by: <span>//'
sleep 1 ;curl -Ss $url |grep "Study:" |awk '{print substr($0, index($0, "Study:"))}'|awk '{print substr($0, 1, index($0, "<div")-1)}' | sed -e 's/Study: <span>//'
sleep 1 ;curl -Ss $url |grep "Sample:" |awk '{print substr($0, index($0, "Sample:"))}' |awk '{print substr($0, 1, index($0, "<div")-1)}'| sed -e 's/Sample: <span>//'
sleep 1 ;curl -Ss $url |grep "Organism:" |awk '{print substr($0, index($0, "Organism:"))}' |awk '{print substr($0, 1, index($0, "</a>")-1)}'| sed -e 's/Organism: <span><a href=".*">//'