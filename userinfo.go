/*
Userinfo library for web applications to store userdata in a BoltDB.

*/

package userinfo

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

func init() {
	ran = rand.New(rand.NewSource(time.Now().UnixNano())) // new random source
}

var users = map[string]Person{}
var objectboxes = map[string]ObjectBox{}
var objects = map[string]UserObject{}
var db *bolt.DB
var ran *rand.Rand
var blankUser Person
var err error
var dbname = ".db"

type Person struct {
	ID, FirstName, LastName, NickName, Email, Data string      `json:",omitempty"`
	Gender, BodyType, Age, LookingFor              uint8       `json:",string,omitempty"`
	Height, ZipCode                                uint32      `json:",string,omitempty"`
	More                                           interface{} `json:",string,omitempty"`
}

type ObjectBox struct {
	OwnerID      string    // Doesn't need ObjectBox ID because each user only has one
	TimeCreated  time.Time `json:",string"`
	TimeModified time.Time `json:",string,omitempty"`
	Bucket       string    `json:",omitempty"`        // big
	Objects      []string  `json:",string,omitempty"` // big

}

func (ob UserObject) Base64() string {
	file, err := os.Open(ob.ObjectID)
	if err != nil {
		log.Println(err)
		return ""
	}
	var b = make([]byte, ob.Size)
	n, err := file.Read(b)
	if err != nil {
		log.Println(err)
		return ""
	}
	return string(b[:n])
}

func (ob UserObject) RawBytes() []byte {
	file, err := os.Open(ob.ObjectID)
	if err != nil {
		log.Println(err)
		return []byte("")
	}
	var b = make([]byte, ob.Size)
	n, err := file.Read(b)
	if err != nil {
		log.Println(err)
		return []byte("")
	}
	dst := make([]byte, 1024*1000)
	i, err := base64.StdEncoding.Decode(dst, b[:n])
	return dst[:i]
}

type UserObject struct {
	OwnerID      string
	ObjectID     string
	Title        string
	Filename     string
	Extension    string
	Data         string    `json:",omitempty"` // base64 encoded inside json encoded
	TimeCreated  time.Time `json:",string"`
	TimeModified time.Time `json:",string,omitempty"`
	Size         int       `json:",string"`
	IsMessage    bool      `json:",string,omitempty"`
	IsMedia      bool      `json:",string,omitempty"`
	IsFile       bool      `json:",string,omitempty"`
	IsPublic     bool      `json:",string,omitempty"`
	Permissions  []string  // List of users able to view/read the databytes
	ImgSrc       string    `json:",omitempty"`
	ImgBase64    string    `json:",omitempty"`
	LinkHref     string    `json:",omitempty"`
	LinkTitle    string    `json:",omitempty"`
	Ref          string    `json:",omitempty"`
	Class        string    `json:",omitempty"` // for css
}

//var userdatas = map[string]ObjectBox{}

func Delete(bucket, key string) error {
	// Delete the key in a different write transaction.
	if err := db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(bucket)).Delete([]byte(key))
	}); err != nil {
		return err
	}
	return nil
}

// Write: Insert data into a bucket.
func Write(bucket, key string, value []byte) error {
	if err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}

		if key != "" { // blank key just creates the bucket. nil value still gets stored if key is named.
			if err := b.Put([]byte(key), value); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}
	return nil
}

// Read: retrieve []byte(value) of bucket[key]
func Read(bucket, key string) []byte {
	if bucket == "" || key == "" {
		return nil
	}

	var v []byte
	err = db.View(func(tx *bolt.Tx) error {
		if tx.Bucket([]byte(bucket)) == nil {
			return nil
		}
		v = tx.Bucket([]byte(bucket)).Get([]byte(key))
		return nil // no error
	})
	if err != nil {
		log.Println(err)
		return nil

	}
	return v
}

// Scan updates the (memory) user map. Must be called after writes to the db
func Scan() map[string]Person {
	var i = 0
	if err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("user"))
		if err != nil {
			return err
		}

		// Iterate over items in sorted key order.
		if err := b.ForEach(func(k, v []byte) error {
			i++
			var user Person
			err = json.Unmarshal(v, &user)
			if err != nil {
				return err
			}
			// Insert into user map[id string]user
			users[string(k)] = user
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	return users
}

// Scan updates the (memory) user map. Must be called after writes to the userdata db to get the true map
func ScanObjectBoxes() map[string]ObjectBox {
	var i = 0
	if err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("objectbox"))
		if err != nil {
			return err
		}

		// Iterate over items in sorted key order.
		if err := b.ForEach(func(k, v []byte) error {
			i++
			var ud ObjectBox
			err = json.Unmarshal(v, &ud)
			if err != nil {
				return err
			}
			// Insert into user map[id string]user
			objectboxes[string(k)] = ud
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	return objectboxes
}

// Scan updates the (memory) user map. Must be called after writes to the userdata db to get the true map
func ScanObjects() map[string]UserObject {
	var i = 0
	if err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("object"))
		if err != nil {
			return err
		}

		// Iterate over items in sorted key order.
		if err := b.ForEach(func(k, v []byte) error {
			i++
			var ud UserObject
			err = json.Unmarshal(v, &ud)
			if err != nil {
				return err
			}
			// Insert into user map[id string]user
			objects[string(k)] = ud
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}
	return objects
}

// Init 1
// userinfo.Init("the.db", []string{"user","userdata","objects"})
func Init(location string, buckets []string) {
	db, err = bolt.Open(location, 0666, nil)
	if err != nil {
		log.Fatal(err)
	}

	// go func() {
	// 	log.Println("Starting stats loop")
	// 	for {
	// 		fmt.Println("stats:", db.Stats().TxStats.WriteTime)
	// 		time.Sleep(10 * time.Second)
	// 		// create buckets if they dont exist
	// 	}
	// }()
	for _, bucket := range buckets {
		err = Write(bucket, "", nil)
		if err != nil {
			log.Println(err)
		}
	}

}

// Close safely close BoltDB
func Close() {
	if err := db.Close(); err != nil {
		log.Fatal(err)
	}
}
