package model

import (
	"git.ifengidc.com/likuo/go-check-certs/config"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"time"
)

type Status int

const (
	Online  Status = 0
	Offline Status = 1
)

var (
	certC = config.MongoSession.DB(config.MongoDatabase).C("cert")
)

type CertModel struct {
	ID         bson.ObjectId `bson:"_id" json:"id"`
	User       []string      `bson:"user" json:"user"`
	Status     Status        `bson:"status" json:"status"`
	Host       string        `bson:"host" json:"host"`
	Port       string        `bson:"port" json:"port"`
	AddTime    time.Time     `bson:"add_time" json:"add_time"`
	UpdateTime time.Time     `bson:"update_time" json:"update_time"`
	Cert       []CertInfo    `bson:"cert" json:"cert"`
}

type CertInfo struct {
	CommonName  string    `bson:"common_name" json:"common_name"`
	ExpireHours int64     `bson:"expire_hours" json:"expire_hours"`
	IsCA        bool      `bson:"is_ca" json:"is_ca"`
	NotBefore   time.Time `bson:"not_before" json:"not_before"`
	NotAfter    time.Time `bson:"not_after" json:"not_after"`
}

func init() {
	certCIndex := []mgo.Index{
		{
			Key:        []string{"host"},
			Unique:     true,
			Background: true,
			Sparse:     true,
		},
		{
			Key:        []string{"user"},
			Background: true,
			Sparse:     true,
		},
	}

	for _, v := range certCIndex {
		err := certC.EnsureIndex(v)
		if err != nil {
			config.Logger.Error("EnsureIndex error", zap.Error(err))
		}
	}
}

func CreateCertInfo(c CertModel) (bool, error) {
	// cc := CertModel{}
	cc, exists, err := GetCertInfoByHost(c.Host)
	if err != nil {
		return false, err
	}
	// Insert new cert if host not exists
	if !exists {
		ok, err := InsertCertInfo(c)
		return ok, err
	}

	// 判断user是否已经在userlist中
	for _, user := range cc.User {
		if user == c.User[0] {
			return true, nil
		}
	}

	// Update cert user if host is exists
	cc.User = append(cc.User, c.User[0])
	ok, err := UpdateCertInfo(cc)
	return ok, err
}

func InsertCertInfo(c CertModel) (bool, error) {
	c.ID = bson.NewObjectId()
	c.AddTime = time.Now()
	c.UpdateTime = time.Now()
	c.Status = Online

	err := certC.Insert(c)
	if err != nil {
		if strings.Contains(err.Error(), "E11000 duplicate key error collection") {
			return false, nil // key 重复要特殊处理
		}
		return false, err
	}
	return true, nil

}

func UpdateCertInfo(c CertModel) (bool, error) {
	// user 去重
	c.User = RemoveDuplicateElement(c.User)

	// 更新配置
	err := certC.Update(bson.M{"host": c.Host, "status": Online}, bson.M{
		"$set": bson.M{
			"user":        c.User,
			"status":      c.Status,
			"port":        c.Port,
			"update_time": time.Now(),
			"cert":        c.Cert,
		},
	})
	if err != nil {
		if err == mgo.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func DeleteCertInfo(c CertModel) (bool, error) {
	//err := certC.Update(bson.M{"_id": c.ID, "status": Online}, bson.M{
	//	"$set": bson.M{
	//		"status": Offline,
	//	},
	//})

	err := certC.RemoveId(c.ID)

	if err != nil {
		if err == mgo.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func DeleteUserFromCertInfo(c CertModel, delUser string) (bool, error) {
	userList := []string{}
	for _, user := range c.User {
		if user == delUser {
			continue
		}
		userList = append(userList, user)
	}
	err := certC.Update(bson.M{"_id": c.ID, "status": Online}, bson.M{
		"$set": bson.M{
			"user": userList,
		},
	})
	if err != nil {
		if err == mgo.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func GetCertInfoByHost(host string) (CertModel, bool, error) {
	c := CertModel{}
	err := certC.Find(bson.M{"host": host, "status": Online}).One(&c)
	if err != nil {
		if err == mgo.ErrNotFound {
			return c, false, nil
		}
		return c, false, err
	}
	return c, true, nil
}

func GetCertInfoByUser(user, host string) (CertModel, bool, error) {
	c := CertModel{}
	err := certC.Find(bson.M{"user": user, "host": host, "status": Online}).One(&c)
	if err != nil {
		if err == mgo.ErrNotFound {
			return c, false, nil
		}
		return c, false, err
	}
	return c, true, nil
}

func GetCertInfoListByUser(user string) ([]CertModel, bool, error) {
	var certModelList []CertModel
	err := certC.Find(bson.M{"user": user, "status": Online}).All(&certModelList)
	if err != nil {
		return certModelList, false, err
	}
	return certModelList, true, nil
}

func GetCertInfoListAll() ([]CertModel, bool, error) {
	var certModelList []CertModel
	err := certC.Find(bson.M{"status": Online}).All(&certModelList)
	if err != nil {
		if err == mgo.ErrNotFound {
			return certModelList, false, nil
		}
		return certModelList, false, err
	}
	return certModelList, true, nil
}

func RemoveDuplicateElement(list []string) []string {
	result := make([]string, 0, len(list))
	temp := map[string]struct{}{}
	for _, e := range list {
		if _, ok := temp[e]; !ok {
			temp[e] = struct{}{}
			result = append(result, e)
		}
	}
	return result
}
