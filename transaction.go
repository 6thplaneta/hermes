package hermes

import ()

func CreateTrans(token string, col Collectionist, obj interface{}) (interface{}, error) {
	db := col.GetDataSrc().DB
	trans, _ := db.Begin()
	result, err := col.Create(token, trans, obj)
	if err != nil {
		trans.Rollback()
		return result, err
	}
	trans.Commit()
	return result, nil
}
