package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/clientv3"
)

var (
	dialTimeout    = 5 * time.Second
	requestTimeout = 2 * time.Second
	endpoints      = []string{"http://127.0.0.1:2379"}
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	log.Println("存储值")
	if _, err := cli.Put(context.TODO(), "sensors", `{sensor01:{topic:"w_sensor01"}}`); err != nil {
		log.Fatal(err)
	}
	log.Println("获取值")
	if resp, err := cli.Get(context.TODO(), "sensors"); err != nil {
		log.Fatal(err)
	} else {
		log.Println("resp: ", resp)
	}
	// see https://github.com/coreos/etcd/blob/master/clientv3/example_kv_test.go#L220
	log.Println("事务&超时")
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	_, err = cli.Txn(ctx).
		If(clientv3.Compare(clientv3.Value("key"), ">", "abc")). // txn value comparisons are lexical
		Then(clientv3.OpPut("key", "XYZ")).                      // this runs, since 'xyz' > 'abc'
		Else(clientv3.OpPut("key", "ABC")).
		Commit()
	cancel()
	if err != nil {
		log.Fatal(err)
	}
	// see https://github.com/coreos/etcd/blob/master/clientv3/example_watch_test.go
	log.Println("监视")
	rch := cli.Watch(context.Background(), "", clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
		}
	}
}

//export ETCDCTL_API=3
// ./etcdctl get  --prefix=true "/etcd3_naming/hello_service"
