package mongo

import (
	"strconv"
	"strings"

	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/appconfig"
	"gopkg.in/mgo.v2"
)

var registryAuthDBSession *mgo.Session
var arrowcloudDBSession *mgo.Session

//InitDatabase initializes connection to mongo db
func InitDatabase() {

	dbOpts := map[string]string{
		"hostname": appconfig.MongoHosts(),
		"port":     appconfig.MongoPort(),
		"rsname":   appconfig.MongoRSName(),
		"username": appconfig.MongoUsername(),
		"password": appconfig.MongoPassword(),
		"dbname":   "registry_auth",
	}
	var err error
	registryAuthDBSession, err = connectDB(dbOpts)
	if err != nil {
		panic(err)
	}
	dbOpts["dbname"] = "arrowcloud"
	arrowcloudDBSession, err = connectDB(dbOpts)
	if err != nil {
		panic(err)
	}
}

func connectDB(options map[string]string) (session *mgo.Session, err error) {

	hostname := options["hostname"]
	port := options["port"]
	rsname := options["rsname"]
	dbname := options["dbname"]
	numConn := options["poolsize"]
	username := options["username"]
	password := options["password"]

	url := "mongodb://"
	if username != "" {
		url += (username + ":" + password + "@")
	}

	var urlServers = ""
	if strings.Contains(hostname, ",") {
		// This is a mongodb replica
		hosts := strings.Split(hostname, ",")
		for i, host := range hosts {
			if i > 0 {
				urlServers += ","
			}
			urlServers += (host + ":" + port)
		}
		url += urlServers
		url += ("/" + dbname)
		if rsname != "" {
			url += ("?replicaSet=" + rsname)
		}

	} else {
		// This is normal single mongodb
		urlServers = hostname + ":" + port
		url += urlServers
		// url += ("/" + dbname)
	}
	// filter out the MongoDB secrets
	result := strings.Split(url, "@")
	if len(result) > 1 {
		log.Debugf("Mongo URL: %s", "mongodb://??:??@" + result[1])
	} else {
		log.Debugf("Mongo URL: %s", result[0])
	}

	session, err = mgo.Dial(url)

	if err != nil {
		log.Errorf("Failed to create mongo session. %v", err)
		return nil, err
	}
	// defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	if numConn == "" {
		numConn = "5"
	}
	intNumConn, err := strconv.ParseInt(numConn, 10, 0)
	if err != nil {
		log.Warningf("Failed to parse numConn %s as int. Will use 5. %v", numConn, err)
		intNumConn = 5
	}
	session.SetPoolLimit(int(intNumConn))

	return session, nil
}

//GetSession returns a registry_auth db session copy
func GetSession() *mgo.Session {
	return registryAuthDBSession.Copy()
}

//GetArrowCloudDBSession returns a arrowcloud db session copy
func GetArrowCloudDBSession() *mgo.Session {
	return arrowcloudDBSession.Copy()
}
