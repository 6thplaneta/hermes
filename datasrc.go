package hermes

import (
	"errors"
	"fmt"

	"github.com/coocood/freecache"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"gopkg.in/olivere/elastic.v3"
	"gopkg.in/redis.v4"
)

type CacheClient struct {
	Enabled   bool
	Engine    string
	Redis     *redis.Client
	FreeCache *freecache.Cache
}

type SearchClient struct {
	Engine    string
	IndexName string
	Elastic   *elastic.Client
	ToIndex   chan interface{}
}

func (s *SearchClient) LoopIndex() {
	for {
		obj := <-s.ToIndex
		err := DoIndexDocument(s, obj)
		if err != nil {
			fmt.Println("Error in Indexing Document: ", err)
		}

	}
}

type DataSrc struct {
	DB     *sqlx.DB
	Cache  *CacheClient
	Search *SearchClient
}

//init datasources if related configs exist
func (src *DataSrc) Init(conf *viper.Viper) error {
	//initial DataSrc.DB from the confing passed to it
	errDB := src.InitDB(conf)
	if errDB != nil {
		return errDB
	}
	//initial DataSrc.Cache from the confing passed to it

	errCache := src.InitCache(conf)
	if errCache != nil {
		return errCache
	}
	//initial DataSrc.Search from the confing passed to it
	errSearch := src.InitSearch(conf)
	if errSearch != nil {
		return errSearch
	}
	return nil
}

//hermes supports postgres,sqlite and mysql
func (src *DataSrc) InitDB(conf *viper.Viper) error {
	dbEngine := conf.GetString("db.engine")
	dbHost := conf.GetString("db.host")
	dbPort := conf.GetString("db.port")
	dbAddr := dbHost + ":" + dbPort
	dbName := conf.GetString("db.name")
	dbUser := conf.GetString("db.user")
	dbPassword := conf.GetString("db.password")

	var errdb error
	var db *sqlx.DB
	if dbEngine == "postgres" {
		db, errdb = sqlx.Connect("postgres", "host="+dbHost+"  port="+dbPort+"  user="+dbUser+" password="+dbPassword+" dbname="+dbName+" sslmode=disable")
		if errdb != nil {
			fmt.Println("Error connectiing to DB ", errdb)
			return errdb
		}
		src.DB = db

	} else if dbEngine == "mysql" {
		db, errdb = sqlx.Connect("mysql", dbUser+":"+dbPassword+"@"+dbAddr+"/"+dbName+"?charset=utf8&parseTime=True&loc=Local")
		if errdb != nil {
			fmt.Println("Error connectiing to DB ", errdb)
			return errdb
		}
		src.DB = db

	} else if dbEngine == "sqlite" {
		db, errdb = sqlx.Connect("sqlite3", dbName)
		if errdb != nil {
			fmt.Println("Error connecting to DB ", errdb)
			return errdb
		}
		src.DB = db
	}
	return nil

}

func (src *DataSrc) InitCache(conf *viper.Viper) error {
	src.Cache = &CacheClient{}
	useCache := conf.GetBool("cache.use_cache")
	if !useCache {
		return nil
	}
	engine := conf.GetString("cache.engine")
	src.Cache.Engine = engine
	if engine == "redis" {
		//
		client := redis.NewClient(&redis.Options{
			Addr:     conf.GetString("cache.redis.addr"),
			Password: conf.GetString("cache.redis.passwd"),
			DB:       0, // use default DB
		})
		_, err := client.Ping().Result()
		if err != nil {
			fmt.Println("Redis Ping Error: ", err)
			return err
		} else {
			src.Cache.Redis = client
			src.Cache.Enabled = true
		}
	} else if engine == "freecache" {
		cachesize := conf.GetInt("cache.free_cache.size") * 1024 * 1024

		src.Cache.FreeCache = freecache.NewCache(cachesize)
		src.Cache.Enabled = true
	} else {
		return errors.New("your cache scheme is not supported: " + engine)
	}
	return nil
}

func (src *DataSrc) InitSearch(conf *viper.Viper) error {
	searchEngine := conf.GetString("search.engine")
	// fmt.Println("search engine, (is elastic)?: ", searchEngine, searchEngine == "elastic")
	src.Search = &SearchClient{Engine: searchEngine}
	if searchEngine == "elastic" {
		addr := conf.GetString("search.elastic.addr")
		indexName := conf.GetString("search.elastic.index")
		fmt.Println("initing search client. Addr, Index: ", addr, indexName)
		client, err := elastic.NewClient(
			elastic.SetURL(addr),
			elastic.SetMaxRetries(10),
			// elastic.SetBasicAuth("user", "secret"))
		)
		if err != nil {
			fmt.Println("Error in connecting to ElasticSearch")
			return err
		}
		src.Search.Elastic = client
		src.Search.IndexName = indexName
		exists, err := client.IndexExists(indexName).Do()
		if err != nil {
			fmt.Println("Error in checking ElasticSearch Index")
			return err
		}
		if !exists {
			_, err = client.CreateIndex(indexName).Do()
			if err != nil {
				// Handle error
				fmt.Println("Error in creating ElasticSearch Index")
				return err
			}
		}
		src.Search.ToIndex = make(chan interface{}, 20000)
		go src.Search.LoopIndex()
	}
	return nil
}
