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

func (src *DataSrc) Init(conf *viper.Viper) error {
	errDB := src.InitDB(conf)
	if errDB != nil {
		return errDB
	}
	errCache := src.InitCache(conf)
	if errCache != nil {
		return errCache
	}
	errSearch := src.InitSearch(conf)
	if errSearch != nil {
		return errSearch
	}
	return nil
}

func (src *DataSrc) InitDB(conf *viper.Viper) error {
	dbEngine := conf.GetString("DB.Engine")
	dbHost := conf.GetString("DB.Host")
	dbPort := conf.GetString("DB.Port")
	dbAddr := dbHost + ":" + dbPort
	dbName := conf.GetString("DB.Name")
	dbUser := conf.GetString("DB.User")
	dbPassword := conf.GetString("DB.Password")

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
	useCache := conf.GetBool("Cache.UseCache")
	if !useCache {
		return nil
	}
	engine := conf.GetString("Cache.Engine")
	src.Cache.Engine = engine
	if engine == "redis" {
		//
		client := redis.NewClient(&redis.Options{
			Addr:     conf.GetString("Cache.Redis.Addr"),
			Password: conf.GetString("Cache.Redis.Passwd"),
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
		cachesize := conf.GetInt("Cache.FreeCache.Size") * 1024 * 1024
		src.Cache.FreeCache = freecache.NewCache(cachesize)
		src.Cache.Enabled = true
	} else {
		return errors.New("your cache scheme is not supported: " + engine)
	}
	return nil
}

func (src *DataSrc) InitSearch(conf *viper.Viper) error {
	searchEngine := conf.GetString("Search.Engine")
	fmt.Println("search engine, (is elastic)?: ", searchEngine, searchEngine == "elastic")
	src.Search = &SearchClient{Engine: searchEngine}
	if searchEngine == "elastic" {
		addr := conf.GetString("Search.Elastic.Addr")
		indexName := conf.GetString("Search.Elastic.Index")
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
