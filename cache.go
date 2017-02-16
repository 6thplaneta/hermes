package hermes

import (
	// "bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"github.com/coocood/freecache"
	"gopkg.in/redis.v4"
	"gopkg.in/vmihailenco/msgpack.v2"
	"reflect"
	"strconv"
	"time"
)

func CacheList(dsrc *DataSrc, coll Collectionist, token string, params *Params, pg *Paging, populate, project string) (interface{}, error) {
	client := dsrc.Cache
	if !client.Enabled || coll.Conf().CacheExpire == 0 {
		return coll.List(token, params, pg, populate, project)
	}
	if client.Engine == "redis" {
		return cacheListRedis(client, coll, token, params, pg, populate, project)
	}
	if client.Engine == "freecache" {
		return cacheListFreeCache(client, coll, token, params, pg, populate, project)
	}
	return coll.List(token, params, pg, populate, project)
}

func cacheListRedis(client *CacheClient, coll Collectionist, token string, params *Params, pg *Paging, populate, project string) (interface{}, error) {
	keybin, err := binKey(params, pg, populate, project)
	key := keyHash(keybin)
	if err != nil {
		return nil, err
	}
	keyExists := true
	val, errRG := client.Redis.Get(key).Result()
	if errRG == redis.Nil {
		keyExists = false
	} else if errRG != nil {
		return nil, errRG
	}
	instType := coll.GetInstanceType()
	if keyExists {
		slice := reflect.MakeSlice(reflect.SliceOf(instType), 0, 0)
		x := reflect.New(slice.Type())
		x.Elem().Set(slice)
		result := x.Interface()
		errUnm := msgpack.Unmarshal([]byte(val), result)
		if errUnm != nil {
			return nil, errUnm
		}
		return result, nil
	} else {
		// get result and add to redis
		res, errList := coll.List(token, params, pg, populate, project)
		if errList != nil {
			return nil, errList
		}
		bin, errMar := msgpack.Marshal(res)
		if errMar != nil {
			return nil, errMar
		}
		redSetErr := client.Redis.Set(key, bin, time.Duration(coll.Conf().CacheExpire)*time.Second).Err()
		if redSetErr != nil {
			return nil, redSetErr
		}
		return res, nil
	}

}

func cacheListFreeCache(client *CacheClient, coll Collectionist, token string, params *Params, pg *Paging, populate, project string) (interface{}, error) {
	key, err := binKey(params, pg, populate, project)
	if err != nil {
		return nil, err
	}
	keyExists := true
	val, errRG := client.FreeCache.Get(key)

	if errRG == freecache.ErrNotFound {
		keyExists = false
	} else if errRG != nil {
		return nil, errRG
	}
	instType := coll.GetInstanceType()
	if keyExists {
		slice := reflect.MakeSlice(reflect.SliceOf(instType), 0, 0)
		x := reflect.New(slice.Type())
		x.Elem().Set(slice)
		result := x.Interface()
		errUnm := msgpack.Unmarshal([]byte(val), result)
		if errUnm != nil {
			return nil, errUnm
		}
		return result, nil
	} else {
		// get result and add to redis
		res, errList := coll.List(token, params, pg, populate, project)
		if errList != nil {
			return nil, errList
		}
		bin, errMar := msgpack.Marshal(res)
		if errMar != nil {
			return nil, errMar
		}
		frSetErr := client.FreeCache.Set(key, bin, coll.Conf().CacheExpire)
		if frSetErr != nil {
			return nil, frSetErr
		}
		return res, nil
	}

}

func keyHash(key []byte) string {
	h := sha1.New()
	h.Write(key)
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs)
}

func binKey(params *Params, pg *Paging, populate, project string) ([]byte, error) {
	// fmt.Printf("params %+v \n", params)
	// var prmBin, pgBin bytes.Buffer
	// binary.Write(&prmBin, binary.BigEndian, params)
	// binary.Write(&pgBin, binary.BigEndian, pg)
	prmBin, err := msgpack.Marshal(params)
	pgBin, err2 := msgpack.Marshal(pg)
	if err != nil {
		return nil, err
	}
	if err2 != nil {
		return nil, err2
	}
	key := prmBin
	key = append(key, pgBin...)
	key = append(key, []byte(populate)...)
	key = append(key, []byte(project)...)
	return key, nil
}

func CacheGet(dsrc *DataSrc, coll Collectionist, token string, id int, populate string) (interface{}, error) {
	client := dsrc.Cache
	if !client.Enabled || coll.Conf().CacheExpire == 0 {
		return coll.Get(token, id, populate)
	}
	if client.Engine == "redis" {
		return cacheGetRedis(client, coll, token, id, populate)
	}
	if client.Engine == "freecache" {
		return cacheGetFreeCache(client, coll, token, id, populate)
	}
	return coll.Get(token, id, populate)
}

func cacheGetRedis(client *CacheClient, coll Collectionist, token string, id int, populate string) (interface{}, error) {
	key := strconv.Itoa(id) + populate
	keyExists := true
	val, errRG := client.Redis.Get(key).Result()
	if errRG == redis.Nil {
		keyExists = false
	} else if errRG != nil {
		return nil, errRG
	}
	instType := coll.GetInstanceType()
	if keyExists {
		result := reflect.New(instType).Interface()
		errUnm := msgpack.Unmarshal([]byte(val), result)
		if errUnm != nil {
			return nil, errUnm
		}
		return result, nil
	} else {
		// get result and add to redis
		res, errGet := coll.Get(token, id, populate)
		if errGet != nil {
			return nil, errGet
		}
		bin, errMar := msgpack.Marshal(res)
		if errMar != nil {
			return nil, errMar
		}
		redSetErr := client.Redis.Set(key, bin, time.Duration(coll.Conf().CacheExpire)*time.Second).Err()
		if redSetErr != nil {
			return nil, redSetErr
		}
		return res, nil
	}

}

func cacheGetFreeCache(client *CacheClient, coll Collectionist, token string, id int, populate string) (interface{}, error) {
	// key, err := binKey(params, pg, populate, project)
	// if err != nil {
	// 	return nil, err
	// }
	idbin := make([]byte, 4)
	binary.LittleEndian.PutUint32(idbin, uint32(id))
	key := append(idbin, []byte(populate)...)
	keyExists := true
	val, errRG := client.FreeCache.Get(key)

	if errRG == freecache.ErrNotFound {
		keyExists = false
	} else if errRG != nil {
		return nil, errRG
	}
	instType := coll.GetInstanceType()
	if keyExists {
		result := reflect.New(instType).Interface()
		errUnm := msgpack.Unmarshal([]byte(val), result)
		if errUnm != nil {
			return nil, errUnm
		}
		return result, nil
	} else {
		// get result and add to redis
		res, errGet := coll.Get(token, id, populate)
		if errGet != nil {
			return nil, errGet
		}
		bin, errMar := msgpack.Marshal(res)
		if errMar != nil {
			return nil, errMar
		}
		frSetErr := client.FreeCache.Set(key, bin, coll.Conf().CacheExpire)
		if frSetErr != nil {
			return nil, frSetErr
		}
		return res, nil
	}

}
