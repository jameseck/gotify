echo "starting ceph-mon"
ceph-mon --id 0
echo "starting ceph-osd"
ceph-osd --id 0
echo "starting ceph-mds"
ceph-mds -i a

/go/gotify
