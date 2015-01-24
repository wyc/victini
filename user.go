import (
	"time"

	"labix.org/v2/mgo/bson"
)

type User struct {
	Id           bson.ObjectId
	CreatedOn    time.Time
	LastActive   time.Time
	Email        string
	PasswordHash string
	Token        string
}

func (u User) Save() {

}
